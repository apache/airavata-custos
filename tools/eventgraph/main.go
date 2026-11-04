// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Command eventgraph maps the Custos event bus of one revision into the event
// graph served at GET /events/graph. It builds SSA and a VTA call graph over ./...,
// resolves every bus Publish and Subscribe call back to its constant topic, and
// follows each handler, loop and HTTP root to the topics it publishes.
// Callbacks passed as arguments count as called at the passing site; the bus internals
// are not followed. Any argument it cannot resolve fails the run, so the graph never
// silently drops an event. It runs from go generate, never at request time.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

var (
	repoRoot, modPath, eventsPkg string
	prog                         *ssa.Program
	cg                           *callgraph.Graph
	fset                         *token.FileSet
)

func main() {
	flag.StringVar(&repoRoot, "root", ".", "repository root")
	out := flag.String("o", "", "output json, stdout if empty")
	flag.Parse()
	repoRoot, _ = filepath.Abs(repoRoot)
	st := analyse()
	// A load error or an unresolved bus call would drop events from the graph silently.
	if len(st.loadErrs)+len(st.unresolved) > 0 {
		fmt.Fprintln(os.Stderr, strings.Join(st.loadErrs, "\n"))
		for _, u := range st.unresolved {
			fmt.Fprintf(os.Stderr, "unresolved %s at %s: %s\n", u["kind"], u["site"], u["why"])
		}
		os.Exit(1)
	}
	b, err := json.MarshalIndent(st.graph(), "", " ")
	if err != nil {
		panic(err)
	}
	b = append(b, '\n')
	if *out == "" {
		os.Stdout.Write(b)
	} else if err := os.WriteFile(*out, b, 0o644); err != nil {
		panic(err)
	}
}

// ---------- naming helpers ----------

func inModule(f *ssa.Function) bool {
	if f == nil {
		return false
	}
	return strings.HasPrefix(pkgOf(f), modPath)
}

func pkgOf(f *ssa.Function) string {
	if f.Pkg != nil {
		return f.Pkg.Pkg.Path()
	}
	if o := f.Origin(); o != nil && o.Pkg != nil {
		return o.Pkg.Pkg.Path()
	}
	return ""
}

func inEvents(f *ssa.Function) bool { return pkgOf(f) == eventsPkg }

func fname(f *ssa.Function) string {
	s := f.String()
	s = strings.ReplaceAll(s, modPath+"/", "")
	return s
}

func pos(p token.Pos) string {
	if !p.IsValid() {
		return ""
	}
	ps := fset.Position(p)
	r, err := filepath.Rel(repoRoot, ps.Filename)
	if err != nil {
		r = ps.Filename
	}
	return fmt.Sprintf("%s:%d", r, ps.Line)
}

func instrPos(i ssa.Instruction) string {
	if i == nil {
		return ""
	}
	if p := i.Pos(); p.IsValid() {
		return pos(p)
	}
	if c, ok := i.(ssa.CallInstruction); ok {
		if v := c.Common().Value; v != nil && v.Pos().IsValid() {
			return pos(v.Pos())
		}
	}
	return pos(i.Parent().Pos())
}

// boundTarget returns the method a bound-method wrapper calls.
func boundTarget(f *ssa.Function) *ssa.Function {
	if !strings.HasPrefix(f.Synthetic, "bound method wrapper") {
		return f
	}
	for _, b := range f.Blocks {
		for _, in := range b.Instrs {
			if c, ok := in.(ssa.CallInstruction); ok {
				if s := c.Common().StaticCallee(); s != nil {
					return s
				}
				if c.Common().IsInvoke() {
					// Interface method value; keep the wrapper.
					return f
				}
			}
		}
	}
	return f
}

// ---------- value resolution ----------

type leaf struct {
	Kind string // str, func, unres
	S    string
	Fn   *ssa.Function
}

// opt is a resolved leaf or, when l is nil, the caller's argument for parameter d.
type opt struct {
	l *leaf
	d int
}

func unres(why string) opt { return opt{l: &leaf{Kind: "unres", S: why}} }

func local(v ssa.Value, depth int) []opt {
	if depth > 12 {
		return []opt{unres("depth")}
	}
	switch x := v.(type) {
	case *ssa.Const:
		if x.Value != nil && x.Value.Kind() == constant.String {
			return []opt{{l: &leaf{Kind: "str", S: constant.StringVal(x.Value)}}}
		}
		return []opt{unres("non-string const " + x.String())}
	case *ssa.ChangeType:
		return local(x.X, depth+1)
	case *ssa.Parameter:
		for i, p := range x.Parent().Params {
			if p == x {
				return []opt{{d: i}}
			}
		}
	case *ssa.Function:
		return []opt{{l: &leaf{Kind: "func", Fn: x}}}
	case *ssa.MakeClosure:
		fn := x.Fn.(*ssa.Function)
		if strings.HasPrefix(fn.Synthetic, "bound method wrapper") {
			return []opt{{l: &leaf{Kind: "func", Fn: boundTarget(fn)}}}
		}
		if inEvents(fn) && fn.Parent() != nil {
			// The bus's typed adapter closure: the user handler is its func-typed binding.
			for _, b := range x.Bindings {
				if isFuncish(b.Type()) {
					return local(b, depth+1)
				}
			}
		}
		return []opt{{l: &leaf{Kind: "func", Fn: fn}}}
	case *ssa.Alloc:
		return stored(x, depth)
	}
	return []opt{unres(fmt.Sprintf("%T at %s", v, pos(v.Pos())))}
}

// argAt returns the argument a call site passes for the callee's parameter p.
func argAt(site ssa.CallInstruction, callee *ssa.Function, p int) ssa.Value {
	c := site.Common()
	if c.IsInvoke() {
		if p == 0 {
			return c.Value
		}
		p--
	}
	if p < len(c.Args) {
		return c.Args[p]
	}
	return nil
}

type hop struct {
	Fn   *ssa.Function
	Site ssa.Instruction
}

type tuple struct {
	Vals []leaf
	Hops []hop // from the sink's caller outward
}

type aval struct {
	v     ssa.Value
	fixed *leaf
}

func product(opts [][]opt) [][]opt {
	res := [][]opt{{}}
	for _, os := range opts {
		var nx [][]opt
		for _, r := range res {
			for _, o := range os {
				c := append(append([]opt{}, r...), o)
				nx = append(nx, c)
			}
		}
		res = nx
	}
	return res
}

func lift(F *ssa.Function, vals []aval, hops []hop, seen map[*ssa.Function]bool) []tuple {
	opts := make([][]opt, len(vals))
	for i, a := range vals {
		if a.fixed != nil {
			opts[i] = []opt{{l: a.fixed}}
		} else if a.v == nil {
			opts[i] = []opt{unres("missing arg")}
		} else {
			opts[i] = local(a.v, 0)
		}
	}
	var out []tuple
	for _, combo := range product(opts) {
		if !slices.ContainsFunc(combo, func(o opt) bool { return o.l == nil }) {
			t := tuple{Hops: hops}
			for _, o := range combo {
				t.Vals = append(t.Vals, *o.l)
			}
			out = append(out, t)
			continue
		}
		if seen[F] || len(hops) > 16 {
			out = append(out, fixUnres(combo, "recursion", hops))
			continue
		}
		seen[F] = true
		found := false
		if node := cg.Nodes[F]; node != nil {
			for _, e := range sortedIn(node) {
				found = true
				nv := make([]aval, len(combo))
				for i, o := range combo {
					if o.l == nil {
						nv[i] = aval{v: argAt(e.Site, F, o.d)}
					} else {
						nv[i] = aval{fixed: o.l}
					}
				}
				out = append(out, lift(e.Caller.Func, nv, append(append([]hop{}, hops...), hop{e.Caller.Func, e.Site}), seen)...)
			}
		}
		delete(seen, F)
		if !found {
			out = append(out, fixUnres(combo, "no caller of "+fname(F), hops))
		}
	}
	return out
}

func fixUnres(combo []opt, why string, hops []hop) tuple {
	t := tuple{Hops: hops}
	for _, o := range combo {
		if o.l != nil {
			t.Vals = append(t.Vals, *o.l)
		} else {
			t.Vals = append(t.Vals, leaf{Kind: "unres", S: why})
		}
	}
	return t
}

// ---------- control dependence ----------

func exitBlock(b *ssa.BasicBlock) bool { return len(b.Succs) == 0 }

// postdominates reports whether every path from F's entry to an exit passes b.
func postdominates(b *ssa.BasicBlock) bool {
	F := b.Parent()
	entry := F.Blocks[0]
	if b == entry {
		return true
	}
	seen := map[*ssa.BasicBlock]bool{b: true}
	stack := []*ssa.BasicBlock{entry}
	seen[entry] = true
	for len(stack) > 0 {
		x := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if exitBlock(x) {
			return false
		}
		for _, s := range x.Succs {
			if !seen[s] {
				seen[s] = true
				stack = append(stack, s)
			}
		}
	}
	return true
}

// condOf reports whether any site does not post-dominate its function's entry.
func condOf(sites []ssa.Instruction) bool {
	for _, s := range sites {
		if s != nil && s.Block() != nil && !postdominates(s.Block()) {
			return true
		}
	}
	return false
}

// ---------- graph traversal ----------

type step struct {
	callee *ssa.Function
	site   ssa.Instruction
	kind   string
}

var succCache = map[*ssa.Function][]step{}

func succs(F *ssa.Function) []step {
	if s, ok := succCache[F]; ok {
		return s
	}
	var out []step
	seen := map[[2]any]bool{}
	add := func(c *ssa.Function, site ssa.Instruction, kind string) {
		if c == nil || inEvents(c) {
			return
		}
		if !inModule(c) && c.Synthetic == "" {
			return
		}
		k := [2]any{c, site}
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, step{c, site, kind})
	}
	if n := cg.Nodes[F]; n != nil {
		for _, e := range n.Out {
			c := e.Site.Common()
			kind := "static"
			if c.IsInvoke() {
				kind = "iface"
			} else if c.StaticCallee() == nil {
				if paramDerived(c.Value) {
					continue
				}
				kind = "dyn"
			}
			if _, ok := e.Site.(*ssa.Go); ok {
				kind = "go:" + kind
			}
			if _, ok := e.Site.(*ssa.Defer); ok {
				kind = "defer:" + kind
			}
			add(e.Callee.Func, e.Site, kind)
		}
	}
	for _, b := range F.Blocks {
		for _, in := range b.Instrs {
			call, ok := in.(ssa.CallInstruction)
			if !ok {
				continue
			}
			for j, a := range call.Common().Args {
				var cb *ssa.Function
				switch x := a.(type) {
				case *ssa.MakeClosure:
					cb = boundTarget(x.Fn.(*ssa.Function))
				case *ssa.Function:
					cb = x
				}
				if cb != nil && callInvokesArg(call, j) {
					add(cb, in, "callback")
				}
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := instrPos(out[i].site), instrPos(out[j].site)
		if a != b {
			return a < b
		}
		return fname(out[i].callee) < fname(out[j].callee)
	})
	succCache[F] = out
	return out
}

type pathEl struct {
	prev *ssa.Function
	st   step
}

// inConnectors reports whether f is declared under connectors/.
func inConnectors(f *ssa.Function) bool { return strings.HasPrefix(pos(f.Pos()), "connectors/") }

// bfs walks the pruned graph from r; base keeps the walk out of connector code.
func bfs(r *ssa.Function, base bool) map[*ssa.Function]pathEl {
	par := map[*ssa.Function]pathEl{r: {}}
	q := []*ssa.Function{r}
	for len(q) > 0 {
		f := q[0]
		q = q[1:]
		for _, s := range succs(f) {
			if base && inConnectors(s.callee) {
				continue
			}
			if _, ok := par[s.callee]; !ok {
				par[s.callee] = pathEl{f, s}
				q = append(q, s.callee)
			}
		}
	}
	return par
}

// pathTo returns the chain r..f, each non-static call tagged with its kind, and its call sites.
func pathTo(par map[*ssa.Function]pathEl, r, f *ssa.Function) ([]string, []ssa.Instruction) {
	var names []string
	var sites []ssa.Instruction
	for x := f; x != r; {
		pe := par[x]
		n := fname(x)
		if pe.st.kind != "static" {
			n += " [" + pe.st.kind + "]"
		}
		names = append([]string{n}, names...)
		sites = append([]ssa.Instruction{pe.st.site}, sites...)
		x = pe.prev
	}
	return append([]string{fname(r)}, names...), sites
}

func isFuncish(t types.Type) bool {
	if p, ok := t.Underlying().(*types.Pointer); ok {
		t = p.Elem()
	}
	_, ok := t.Underlying().(*types.Signature)
	return ok
}

// stored resolves a local variable to the values stored into it.
func stored(a *ssa.Alloc, depth int) []opt {
	var r []opt
	for _, ref := range *a.Referrers() {
		if s, ok := ref.(*ssa.Store); ok && s.Addr == a {
			r = append(r, local(s.Val, depth+1)...)
		}
	}
	if len(r) == 0 {
		return []opt{unres("alloc without store")}
	}
	return r
}

// paramDerived reports whether a called func value comes from the caller's
// parameters or captures; those calls are modelled at the passing site.
func paramDerived(v ssa.Value) bool {
	switch x := v.(type) {
	case *ssa.Parameter, *ssa.FreeVar:
		return true
	case *ssa.UnOp:
		if x.Op != token.MUL {
			return false
		}
		switch y := x.X.(type) {
		case *ssa.FreeVar:
			return true
		case *ssa.Alloc:
			n := 0
			for _, r := range *y.Referrers() {
				if s, ok := r.(*ssa.Store); ok && s.Addr == y {
					n++
					if !paramDerived(s.Val) {
						return false
					}
				}
			}
			return n > 0
		}
	}
	return false
}

// callInvokesArg reports whether some callee of the call runs its j-th
// argument, rather than only storing or registering it.
func callInvokesArg(call ssa.CallInstruction, j int) bool {
	c := call.Common()
	var callees []*ssa.Function
	if s := c.StaticCallee(); s != nil {
		callees = []*ssa.Function{s}
	} else if n := cg.Nodes[call.Parent()]; n != nil {
		for _, e := range n.Out {
			if e.Site == call {
				callees = append(callees, e.Callee.Func)
			}
		}
	}
	for _, g := range callees {
		p := j
		if c.IsInvoke() {
			p = j + 1
		}
		if invokesParam(g, p, map[*ssa.Function]bool{}) {
			return true
		}
	}
	return false
}

var invokeMemo = map[[2]any]bool{}

// invokesParam reports whether g calls its parameter p, directly or by passing
// it to a callee that does.
// Outside the module every func is assumed to run its func arguments.
func invokesParam(g *ssa.Function, p int, seen map[*ssa.Function]bool) bool {
	if !inModule(g) && g.Synthetic == "" {
		return true
	}
	if inEvents(g) || p >= len(g.Params) || len(g.Blocks) == 0 {
		return false
	}
	k := [2]any{g, p}
	if v, ok := invokeMemo[k]; ok {
		return v
	}
	if seen[g] {
		return false
	}
	seen[g] = true
	r := valueInvoked(g.Params[p], seen)
	invokeMemo[k] = r
	return r
}

func valueInvoked(v ssa.Value, seen map[*ssa.Function]bool) bool {
	refs := v.Referrers()
	if refs == nil {
		return false
	}
	for _, r := range *refs {
		switch x := r.(type) {
		case ssa.CallInstruction:
			c := x.Common()
			if c.Value == v && !c.IsInvoke() {
				return true
			}
			for j, a := range c.Args {
				if a == v && callInvokesArgSeen(x, j, seen) {
					return true
				}
			}
		}
	}
	return false
}

func callInvokesArgSeen(call ssa.CallInstruction, j int, seen map[*ssa.Function]bool) bool {
	c := call.Common()
	if s := c.StaticCallee(); s != nil {
		return invokesParam(s, j, seen)
	}
	return callInvokesArg(call, j)
}

// sortedIn orders a node's in-edges by caller and site so output is stable.
func sortedIn(n *callgraph.Node) []*callgraph.Edge {
	in := append([]*callgraph.Edge{}, n.In...)
	sort.SliceStable(in, func(i, j int) bool {
		a, b := fname(in[i].Caller.Func)+instrPos(in[i].Site), fname(in[j].Caller.Func)+instrPos(in[j].Site)
		return a < b
	})
	return in
}
