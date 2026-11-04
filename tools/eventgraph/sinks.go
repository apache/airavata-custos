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
	"go/constant"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/ssa"
)

// Sink is a function that leaves the process or writes the database.
type Sink struct {
	Name  string
	Class string // external transport, external API method or db write
	Scope string // core or connector dir
}

var sqlWrite = regexp.MustCompile(`(?is)\b(insert\s+into|delete\s+from)\s+([a-z_][a-z0-9_."]*)|\b(update)\s+([a-z_][a-z0-9_."]*)(?:\s+(?:as\s+)?[a-z_][a-z0-9_]*)?\s+set\b`)

var extCallees = map[string]bool{
	"(*net/http.Client).Do": true, "(*net/http.Client).Get": true, "(*net/http.Client).Post": true,
	"(*net/http.Client).PostForm": true, "(*net/http.Client).Head": true,
	"net/http.Get": true, "net/http.Post": true, "net/http.PostForm": true, "net/http.Head": true,
	"net/smtp.SendMail": true, "os/exec.Command": true, "os/exec.CommandContext": true,
}

func extIO(c *ssa.Function) bool {
	if c == nil || inModule(c) {
		return false
	}
	s := c.String()
	if extCallees[s] {
		return true
	}
	n := c.Name()
	return strings.Contains(pkgOf(c), "mail") && (strings.HasPrefix(n, "DialAndSend") || n == "Send")
}

// classifySinks marks module functions that call an external I/O function
// (HTTP client, SMTP, exec) or carry an SQL INSERT/UPDATE/DELETE literal.
func (st *state) classifySinks() {
	st.sinks = map[*ssa.Function]*Sink{}
	for _, f := range st.modFns {
		file := pos(f.Pos())
		scope := "core"
		if strings.HasPrefix(file, "connectors/") {
			scope = connectorOf(file)
		}
		ext := false
		dbWrite := false
		node := cg.Nodes[f]
		for _, b := range f.Blocks {
			for _, in := range b.Instrs {
				for _, op := range in.Operands(nil) {
					if k, ok := (*op).(*ssa.Const); ok && k.Value != nil && k.Value.Kind() == constant.String {
						dbWrite = dbWrite || sqlWrite.MatchString(constant.StringVal(k.Value))
					}
				}
				c, ok := in.(ssa.CallInstruction)
				if !ok {
					continue
				}
				callee := c.Common().StaticCallee()
				if extIO(callee) {
					ext = true
				}
				if node != nil {
					for _, e := range node.Out {
						if e.Site == c && extIO(e.Callee.Func) {
							ext = true
						}
					}
				}
			}
		}
		switch {
		case ext:
			st.sinks[f] = &Sink{Name: fname(f), Class: "external", Scope: scope}
		case dbWrite:
			st.sinks[f] = &Sink{Name: fname(f), Class: "db write", Scope: scope}
		}
	}
	// Exported functions of a client package that reach its transport are the external API methods.
	transport := map[string][]*ssa.Function{}
	for f, s := range st.sinks {
		if s.Class == "external" {
			s.Class = "external transport"
			transport[pkgOf(f)] = append(transport[pkgOf(f)], f)
		}
	}
	for k := range transport {
		sort.Slice(transport[k], func(i, j int) bool { return fname(transport[k][i]) < fname(transport[k][j]) })
	}
	for _, f := range st.modFns {
		ts := transport[pkgOf(f)]
		if len(ts) == 0 || st.sinks[f] != nil || f.Parent() != nil || f.Object() == nil || !f.Object().Exported() {
			continue
		}
		par := bfs(f, false)
		for _, t := range ts {
			if _, ok := par[t]; ok {
				st.sinks[f] = &Sink{Name: fname(f), Class: "external API method", Scope: st.sinks[t].Scope}
				break
			}
		}
	}
}
