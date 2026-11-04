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

import (
	"sort"

	"golang.org/x/tools/go/ssa"
)

// Transition is To published after From on one core code path.
type Transition struct {
	From        string
	To          string
	SiteFrom    string
	SiteTo      string
	Func        string
	Chain       []string
	Conditional bool
	fromFn      *ssa.Function
}

// laterInstrs returns the instructions CFG-reachable after s in its function.
func laterInstrs(s ssa.Instruction) []ssa.Instruction {
	b := s.Block()
	var out []ssa.Instruction
	after := false
	for _, in := range b.Instrs {
		if after {
			out = append(out, in)
		}
		if in == s {
			after = true
		}
	}
	seen := map[*ssa.BasicBlock]bool{}
	stack := append([]*ssa.BasicBlock{}, b.Succs...)
	for len(stack) > 0 {
		x := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[x] {
			continue
		}
		seen[x] = true
		out = append(out, x.Instrs...)
		stack = append(stack, x.Succs...)
	}
	return out
}

var baseReach = map[*ssa.Function]map[*ssa.Function]pathEl{}

// coreTransitions climbs from each core publish through its callers.
// In each caller it records the publishes reachable after the call.
func (st *state) coreTransitions() []Transition {
	type key struct{ from, to, sf, stt string }
	best := map[key]Transition{}
	put := func(t Transition) {
		k := key{t.From, t.To, t.SiteFrom, t.SiteTo}
		if old, ok := best[k]; ok && len(old.Chain) <= len(t.Chain) {
			return
		}
		best[k] = t
	}
	// byEntry keys each origin by the call in its own function that leads to Publish.
	byEntry := map[ssa.Instruction][]origin{}
	for _, o := range st.origins {
		byEntry[o.sites[0]] = append(byEntry[o.sites[0]], o)
	}
	// callers inverts succs over core code. A callback's caller is the function that passes it.
	callers := map[*ssa.Function][]step{}
	for _, h := range st.modFns {
		if inConnectors(h) {
			continue
		}
		for _, stp := range succs(h) {
			if !inConnectors(stp.callee) {
				callers[stp.callee] = append(callers[stp.callee], step{callee: h, site: stp.site, kind: stp.kind})
			}
		}
	}
	// scan records each publish reachable after at in G. prefix is the chain from a.fn up to G.
	scan := func(a origin, G *ssa.Function, at ssa.Instruction, prefix []string) {
		for _, in := range laterInstrs(at) {
			for _, b := range byEntry[in] {
				c := condOf(append([]ssa.Instruction{in}, b.sites[1:]...))
				put(Transition{From: a.topic, To: b.topic, SiteFrom: a.pubPos, SiteTo: b.pubPos, Func: fname(G),
					Chain: append(append([]string{}, prefix...), b.chain[1:]...), Conditional: c, fromFn: a.fn})
			}
			if _, ok := in.(ssa.CallInstruction); !ok {
				continue
			}
			for _, stp := range succs(G) {
				if stp.site != in || inConnectors(stp.callee) {
					continue
				}
				par := baseReach[stp.callee]
				if par == nil {
					par = bfs(stp.callee, true)
					baseReach[stp.callee] = par
				}
				for _, g := range sortedFns(par) {
					for _, b := range st.originsBy[g] {
						names, sites := pathTo(par, stp.callee, g)
						chain := append(append(append([]string{}, prefix...), names...), b.chain[1:]...)
						c := condOf(append(append([]ssa.Instruction{in}, sites...), b.sites...))
						put(Transition{From: a.topic, To: b.topic, SiteFrom: a.pubPos, SiteTo: b.pubPos, Func: fname(G),
							Chain: chain, Conditional: c, fromFn: a.fn})
					}
				}
			}
		}
	}
	type up struct {
		G      *ssa.Function
		at     ssa.Instruction
		prefix []string
	}
	for _, a := range st.origins {
		if inConnectors(a.fn) {
			continue
		}
		seen := map[ssa.Instruction]bool{a.sites[0]: true}
		q := []up{{a.fn, a.sites[0], []string{fname(a.fn)}}}
		for len(q) > 0 {
			u := q[0]
			q = q[1:]
			scan(a, u.G, u.at, u.prefix)
			for _, cl := range callers[u.G] {
				if seen[cl.site] {
					continue
				}
				seen[cl.site] = true
				q = append(q, up{cl.callee, cl.site, append(append([]string{}, u.prefix...), "^ "+fname(cl.callee))})
			}
		}
	}
	out := make([]Transition, 0, len(best))
	for _, t := range best {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		if a.SiteFrom != b.SiteFrom {
			return a.SiteFrom < b.SiteFrom
		}
		return a.SiteTo < b.SiteTo
	})
	return out
}

// handlerPrimary lists the topics a handler publishes directly.
// A topic is indirect when a transition in the handler's reach leads to it from another of its topics.
func (st *state) handlerPrimary(trans []Transition, handler string, fn *ssa.Function) []string {
	emitted := map[string]bool{}
	for _, e := range st.emits {
		if e.RootKind == "handler" && e.Handler == handler {
			emitted[e.Topic] = true
		}
	}
	par := st.reach[fn]
	for _, t := range trans {
		if _, ok := par[t.fromFn]; ok && t.From != t.To && emitted[t.From] {
			delete(emitted, t.To)
		}
	}
	return sortedKeys(emitted)
}
