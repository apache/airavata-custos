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

// graph shapes the JSON that GET /events/graph serves.
// Each Event is a bus topic. Its Entry list names what publishes it first:
// API, timer, startup or a connector type. Subs counts connector subscribers.
// An Edge links two events. A Term ends a flow.

import (
	"cmp"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/ssa"
)

type Entry struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type Event struct {
	ID    string  `json:"id"`
	Entry []Entry `json:"entry"`
	Subs  int     `json:"subs"`
}

// Edge is a transition from one event to the next.
// Kind core: a core function publishes To after From.
// Kind conn: a connector handler on From publishes To directly, not through another of its topics.
type Edge struct {
	From  string   `json:"from"`
	To    string   `json:"to"`
	Kind  string   `json:"kind"`
	Conn  string   `json:"conn"`
	Actor string   `json:"actor"`
	Sites []string `json:"sites"`
	Chain []string `json:"chain"`
	Cond  bool     `json:"cond"`
	// Order marks a core pair published in a helper and then its caller: sequence, not cause.
	Order bool `json:"order,omitempty"`
}

// Term is a connector handler on From that publishes no event.
type Term struct {
	From    string   `json:"from"`
	Conn    string   `json:"conn"`
	Actor   string   `json:"actor"`
	Site    string   `json:"site"`
	Effects []string `json:"effects"`
	Cond    bool     `json:"cond"`
}

type Graph struct {
	Events []*Event `json:"events"`
	Edges  []*Edge  `json:"edges"`
	Terms  []*Term  `json:"terms"`
}

var (
	tagRe    = regexp.MustCompile(` \[[^]]*\]$`)
	closRe   = regexp.MustCompile(`\$\d+$`)
	recvRe   = regexp.MustCompile(`^\(\*?`)
	tickerRe = regexp.MustCompile(`loop|goroutine`)
)

func bare(f string) string { return tagRe.ReplaceAllString(f, "") }

func short(f string) string {
	s := strings.ReplaceAll(recvRe.ReplaceAllString(bare(f), ""), ")", "")
	return s[strings.LastIndex(s, "/")+1:]
}

func actor(f string) string {
	s := short(f)
	if i := strings.Index(s, "."); i >= 0 {
		s = s[i+1:]
	}
	return closRe.ReplaceAllString(s, "")
}

func shorts(xs []string) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = short(x)
	}
	return out
}

func rootLabel(key string) string {
	if l, ok := strings.CutPrefix(key, "route "); ok {
		return l
	}
	return short(key)
}

// entryKind names what starts a root: a connector by its type, core by trigger.
func (st *state) entryKind(r rootT) string {
	switch {
	case r.layer != "base":
		return st.connName(r.layer)
	case r.kind == "http route":
		return "API"
	case r.kind == "main":
		return "startup"
	case tickerRe.MatchString(r.kind):
		return "timer"
	}
	return r.kind
}

func (st *state) connName(dir string) string { return cmp.Or(st.connType[dir], dir) }

// graph reduces the analysis to a sorted Graph.
func (st *state) graph() Graph {
	events, edges, terms := map[string]*Event{}, map[[3]string]*Edge{}, map[[2]string]*Term{}
	node := func(t string) *Event {
		if events[t] == nil {
			events[t] = &Event{ID: t, Entry: []Entry{}}
		}
		return events[t]
	}
	edge := func(e *Edge) {
		node(e.From)
		node(e.To)
		k := [3]string{e.From, e.To, e.Actor}
		if o := edges[k]; o != nil {
			o.Cond, o.Order = o.Cond && e.Cond, o.Order && e.Order
			for _, s := range e.Sites {
				o.Sites = appendU(o.Sites, s)
			}
			return
		}
		edges[k] = e
	}

	pubSite := map[[2]string]string{}
	for _, o := range st.origins {
		node(o.topic)
		pubSite[[2]string{o.chain[0], o.topic}] = o.pubPos
	}
	trans := map[[2]string]bool{}
	for _, x := range st.trans {
		if x.From == x.To {
			continue
		}
		trans[[2]string{x.From, x.To}] = true
		edge(&Edge{From: x.From, To: x.To, Kind: "core", Conn: "core", Actor: actor(x.Func), Sites: []string{x.SiteFrom + " → " + x.SiteTo},
			Chain: shorts(x.Chain), Cond: x.Conditional, Order: actor(x.Chain[0]) != actor(x.Func)})
	}

	// An entry starts the topics its root publishes that no transition from another of them explains.
	for _, r := range st.roots {
		if r.kind == "handler" {
			continue
		}
		var ts []string
		for _, e := range st.emits {
			if e.Handler == fname(r.fn) && e.RootKind == r.kind {
				ts = append(ts, e.Topic)
			}
		}
		for _, t := range ts {
			if !slices.ContainsFunc(ts, func(a string) bool { return a != t && trans[[2]string{a, t}] }) {
				node(t).Entry = append(node(t).Entry, Entry{st.entryKind(r), rootLabel(r.key)})
			}
		}
	}

	for _, s := range st.subs {
		if s.Connector == "core" {
			continue
		}
		c := st.connName(s.Connector)
		name := c + "." + actor(s.Handler)
		node(s.Topic).Subs++
		var primary []string
		if s.hfn != nil {
			primary = st.handlerPrimary(st.trans, s.Handler, s.hfn)
		}
		emitted := false
		for _, e := range st.emits {
			if e.RootKind != "handler" || e.Handler != s.Handler || s.hfn != nil && !slices.Contains(primary, e.Topic) {
				continue
			}
			emitted = true
			site := cmp.Or(pubSite[[2]string{bare(e.Chain[len(e.Chain)-1]), e.Topic}], "?")
			edge(&Edge{From: s.Topic, To: e.Topic, Kind: "conn", Conn: c, Actor: name, Sites: []string{s.File + " → " + site},
				Chain: shorts(e.Chain), Cond: e.Conditional})
		}
		if !emitted {
			terms[[2]string{s.Topic, name}] = &Term{From: s.Topic, Conn: c, Actor: name, Site: s.File, Effects: st.effects(s.hfn), Cond: s.Conditional}
		}
	}

	var g Graph
	for _, e := range events {
		slices.SortFunc(e.Entry, func(a, b Entry) int { return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.Label, b.Label)) })
		e.Entry = slices.Compact(e.Entry)
		g.Events = append(g.Events, e)
	}
	for _, e := range edges {
		g.Edges = append(g.Edges, e)
	}
	for _, t := range terms {
		g.Terms = append(g.Terms, t)
	}
	slices.SortFunc(g.Events, func(a, b *Event) int { return cmp.Compare(a.ID, b.ID) })
	slices.SortFunc(g.Edges, func(a, b *Edge) int {
		return cmp.Or(cmp.Compare(a.From, b.From), cmp.Compare(a.To, b.To), cmp.Compare(a.Actor, b.Actor))
	})
	slices.SortFunc(g.Terms, func(a, b *Term) int { return cmp.Or(cmp.Compare(a.From, b.From), cmp.Compare(a.Actor, b.Actor)) })
	return g
}

// effects lists the external calls and connector database writes a handler reaches.
func (st *state) effects(h *ssa.Function) []string {
	out := []string{}
	for f := range st.reach[h] {
		if s := st.sinks[f]; s != nil && (s.Class != "db write" || s.Scope != "core") {
			out = appendU(out, short(s.Name))
		}
	}
	slices.Sort(out)
	return out
}
