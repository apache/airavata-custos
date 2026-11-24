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

package main

// analyse loads ./... and collects the facts that graph reduces.
// An origin is a function that publishes a constant topic, with its call chain to Publish.
// A root is where a flow starts: a subscribed handler, an HTTP route, a goroutine or a loop-named func.
// An emit is a topic a root reaches, by its shortest chain.
// Connector roots walk anywhere. Core roots never enter connectors/.

import (
	"go/constant"
	"go/types"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type Subscription struct {
	Subscriber  string
	Connector   string
	Topic       string
	Handler     string
	File        string
	Conditional bool
	hfn         *ssa.Function
}
type Emit struct {
	Handler     string
	RootKind    string
	Topic       string
	Chain       []string
	Conditional bool
}

type origin struct {
	topic  string
	fn     *ssa.Function
	chain  []string
	sites  []ssa.Instruction
	pubPos string
}

// walkRoot is a function the analysis walks from: an event handler, HTTP route, goroutine or loop.
type walkRoot struct {
	fn    *ssa.Function
	kind  string
	layer string // base or the connector dir
	key   string
}

// connectorOf names the component that owns a file.
// Under connectors/ it is connectors/<name> or connectors/<group>/<name>.
// Elsewhere it is the top two path elements.
func connectorOf(file string) string {
	parts := strings.Split(file, "/")
	if len(parts) >= 3 && parts[0] == "connectors" {
		switch parts[2] {
		case "internal", "pkg", "api", "cmd":
			return "connectors/" + parts[1]
		}
		return "connectors/" + parts[1] + "/" + parts[2]
	}
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return file
}

func layerOf(f *ssa.Function) string {
	if inConnectors(f) {
		return connectorOf(pos(f.Pos()))
	}
	return "base"
}

var routeRe = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE|HEAD) /`)

type state struct {
	pkgs       []*packages.Package
	all        map[*ssa.Function]bool
	modFns     []*ssa.Function
	origins    []origin
	originsBy  map[*ssa.Function][]origin
	subs       []Subscription
	roots      []walkRoot
	reach      map[*ssa.Function]map[*ssa.Function]pathEl
	emits      []Emit
	routeOf    map[*ssa.Function]string
	connType   map[string]string
	sinks      map[*ssa.Function]*Sink
	unresolved []map[string]any
	loadErrs   []string
}

func analyse() *state {
	st := &state{originsBy: map[*ssa.Function][]origin{},
		reach: map[*ssa.Function]map[*ssa.Function]pathEl{}, routeOf: map[*ssa.Function]string{},
		connType: map[string]string{}}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedModule,
		Dir:   repoRoot,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(err)
	}
	st.pkgs = pkgs
	modPath = pkgs[0].Module.Path
	eventsPkg = modPath + "/pkg/events"
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		for _, e := range p.Errors {
			st.loadErrs = append(st.loadErrs, e.Error())
		}
	})
	fset = pkgs[0].Fset
	var ssapkgs []*ssa.Package
	prog, ssapkgs = ssautil.AllPackages(pkgs, ssa.InstantiateGenerics)
	prog.Build()
	st.all = ssautil.AllFunctions(prog)
	cg = vta.CallGraph(st.all, cha.CallGraph(prog))

	var evp *ssa.Package
	for _, p := range ssapkgs {
		if p != nil && p.Pkg.Path() == eventsPkg {
			evp = p
		}
	}
	if evp == nil {
		panic("events package not found")
	}
	for f := range st.all {
		if inModule(f) && !inEvents(f) && len(f.Blocks) > 0 {
			st.modFns = append(st.modFns, f)
		}
	}
	sort.Slice(st.modFns, func(i, j int) bool { return fname(st.modFns[i]) < fname(st.modFns[j]) })

	busPtr := types.NewPointer(evp.Pkg.Scope().Lookup("Bus").Type())
	mset := prog.MethodSets.MethodSet(busPtr)
	method := func(n string) *ssa.Function {
		if s := mset.Lookup(evp.Pkg, n); s != nil {
			return prog.MethodValue(s)
		}
		return nil
	}
	st.findPublishes(method("Publish"))
	st.findSubscriptions(method("Subscribe"))
	st.findRoutes()
	st.findConnectorTypes()
	st.findRoots()
	st.classifySinks()

	for _, r := range st.roots {
		par := bfs(r.fn, r.layer == "base")
		st.reach[r.fn] = par
		best := map[string]Emit{}
		for _, f := range sortedFns(par) {
			for _, o := range st.originsBy[f] {
				names, sites := pathTo(par, r.fn, f)
				chain := append(names, o.chain[1:]...)
				if old, ok := best[o.topic]; ok && len(old.Chain) <= len(chain) {
					continue
				}
				best[o.topic] = Emit{Handler: fname(r.fn), RootKind: r.kind, Topic: o.topic, Chain: chain, Conditional: condOf(append(sites, o.sites...))}
			}
		}
		for _, t := range sortedKeys(best) {
			st.emits = append(st.emits, best[t])
		}
	}
	return st
}

func (st *state) findPublishes(publish *ssa.Function) {
	for _, e := range sortedIn(cg.Nodes[publish]) {
		ts := lift(e.Caller.Func, []aval{{v: argAt(e.Site, publish, 3)}}, []hop{{e.Caller.Func, e.Site}}, map[*ssa.Function]bool{})
		for _, t := range ts {
			if t.Vals[0].Kind != "str" {
				st.unresolved = append(st.unresolved, map[string]any{"kind": "publish topic", "site": instrPos(e.Site), "why": t.Vals[0].S})
				continue
			}
			o := origin{topic: t.Vals[0].S, fn: t.Hops[len(t.Hops)-1].Fn, pubPos: instrPos(e.Site)}
			for i := len(t.Hops) - 1; i >= 0; i-- {
				o.chain = append(o.chain, fname(t.Hops[i].Fn))
				o.sites = append(o.sites, t.Hops[i].Site)
			}
			st.origins = append(st.origins, o)
			st.originsBy[o.fn] = append(st.originsBy[o.fn], o)
		}
	}
}

func (st *state) findSubscriptions(subscribe *ssa.Function) {
	for _, e := range sortedIn(cg.Nodes[subscribe]) {
		vals := []aval{{v: argAt(e.Site, subscribe, 1)}, {v: argAt(e.Site, subscribe, 2)}, {v: argAt(e.Site, subscribe, 3)}}
		for _, t := range lift(e.Caller.Func, vals, []hop{{e.Caller.Func, e.Site}}, map[*ssa.Function]bool{}) {
			if allInEvents(t.Hops) {
				continue
			}
			s := Subscription{}
			var outer []ssa.Instruction
			for _, h := range t.Hops {
				if !inEvents(h.Fn) {
					if s.File == "" {
						s.File = instrPos(h.Site)
					}
					outer = append(outer, h.Site)
				}
			}
			s.Subscriber, s.Topic = t.Vals[0].S, t.Vals[1].S
			if t.Vals[0].Kind != "str" {
				s.Subscriber = "UNRESOLVED(" + t.Vals[0].S + ")"
				st.unresolved = append(st.unresolved, map[string]any{"kind": "subscriber name", "site": s.File, "why": t.Vals[0].S})
			}
			if t.Vals[1].Kind != "str" {
				s.Topic = "UNRESOLVED(" + t.Vals[1].S + ")"
				st.unresolved = append(st.unresolved, map[string]any{"kind": "subscription topic", "site": s.File, "why": t.Vals[1].S})
			}
			if t.Vals[2].Kind == "func" {
				s.hfn = t.Vals[2].Fn
				s.Handler = fname(s.hfn)
			} else {
				s.Handler = "UNRESOLVED(" + t.Vals[2].S + ")"
				st.unresolved = append(st.unresolved, map[string]any{"kind": "handler", "site": s.File, "why": t.Vals[2].S})
			}
			s.Connector = connectorOf(s.File)
			if !strings.HasPrefix(s.File, "connectors/") {
				s.Connector = "core"
			}
			s.Conditional = condOf(outer)
			st.subs = append(st.subs, s)
		}
	}
	sort.Slice(st.subs, func(i, j int) bool {
		a, b := st.subs[i], st.subs[j]
		if a.Subscriber != b.Subscriber {
			return a.Subscriber < b.Subscriber
		}
		if a.Topic != b.Topic {
			return a.Topic < b.Topic
		}
		return a.Handler < b.Handler
	})
}

// findRoutes maps HTTP handler funcs to the "METHOD /path" constant passed with them.
func (st *state) findRoutes() {
	for _, f := range st.modFns {
		for _, b := range f.Blocks {
			for _, in := range b.Instrs {
				c, ok := in.(ssa.CallInstruction)
				if !ok {
					continue
				}
				route := ""
				for _, a := range c.Common().Args {
					if k, ok := a.(*ssa.Const); ok && k.Value != nil && k.Value.Kind() == constant.String && routeRe.MatchString(constant.StringVal(k.Value)) {
						route = constant.StringVal(k.Value)
					}
				}
				if route == "" {
					continue
				}
				for _, a := range c.Common().Args {
					if _, ok := a.Type().Underlying().(*types.Signature); !ok {
						continue
					}
					for _, o := range local(a, 0) {
						if o.l != nil && o.l.Kind == "func" {
							if old, ok := st.routeOf[o.l.Fn]; !ok || old > route {
								st.routeOf[o.l.Fn] = route
							}
						}
					}
				}
			}
		}
	}
}

// findConnectorTypes records the package-level "Type" constant of each connector.
func (st *state) findConnectorTypes() {
	packages.Visit(st.pkgs, nil, func(p *packages.Package) {
		if !strings.HasPrefix(p.PkgPath, modPath+"/connectors/") || p.Types == nil {
			return
		}
		if c, ok := p.Types.Scope().Lookup("Type").(*types.Const); ok && c.Val().Kind() == constant.String {
			st.connType[connectorOf(pos(c.Pos()))] = constant.StringVal(c.Val())
		}
	})
}

func (st *state) findRoots() {
	seen := map[*ssa.Function]bool{}
	add := func(f *ssa.Function, kind string) {
		if f == nil || seen[f] {
			return
		}
		seen[f] = true
		r := walkRoot{fn: f, kind: kind, layer: layerOf(f)}
		r.key = fname(f)
		if rt, ok := st.routeOf[f]; ok {
			r.key = "route " + rt
		}
		st.roots = append(st.roots, r)
	}
	for _, s := range st.subs {
		add(s.hfn, "handler")
	}
	for _, f := range st.modFns {
		if _, ok := st.routeOf[f]; ok {
			add(f, "http route")
		}
	}
	loopName := regexp.MustCompile(`(?i)^(run|start|reconcile|poll|sweep|process|loop|tick|drain|scan|watch|monitor|expire)`)
	for _, f := range st.modFns {
		for _, b := range f.Blocks {
			for _, in := range b.Instrs {
				if g, ok := in.(*ssa.Go); ok {
					if c := g.Call.StaticCallee(); c != nil && inModule(c) {
						add(c, "goroutine (go stmt in "+fname(f)+")")
					}
				}
			}
		}
	}
	for _, f := range st.modFns {
		file := pos(f.Pos())
		if (strings.HasPrefix(file, "connectors/") || strings.HasPrefix(file, "internal/") || strings.HasPrefix(file, "pkg/")) &&
			f.Parent() == nil && f.Synthetic == "" && loopName.MatchString(f.Name()) && f.Object() != nil {
			add(f, "loop-named func")
		}
	}
}

func appendU(s []string, v string) []string {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}

func sortedKeys[V any](m map[string]V) []string {
	var k []string
	for x := range m {
		k = append(k, x)
	}
	sort.Strings(k)
	return k
}

func allInEvents(hops []hop) bool {
	for _, h := range hops {
		if !inEvents(h.Fn) {
			return false
		}
	}
	return true
}

func sortedFns(par map[*ssa.Function]pathEl) []*ssa.Function {
	out := make([]*ssa.Function, 0, len(par))
	for f := range par {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return fname(out[i]) < fname(out[j]) })
	return out
}
