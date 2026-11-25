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

// cli/internal/client/tres.go
package client

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseTRES parses a comma-separated TRES spec:
//
//	"cpu=100", "cpu=100,mem=8000", "gres/gpu=2", "cpu=10,gres/gpu=4"
func ParseTRES(s string) ([]TRES, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]TRES, 0, len(parts))
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("malformed TRES entry %q (want key=value)", p)
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("TRES count %q not a number", kv[1])
		}
		out = append(out, TRESFromKey(strings.TrimSpace(kv[0]), n))
	}
	return out, nil
}

// Key joins the type and name as SLURM writes them, e.g. "gres/gpu".
func (t TRES) Key() string {
	if t.Name == "" {
		return t.Type
	}
	return t.Type + "/" + t.Name
}

// TRESFromKey splits a "type[/name]" key, e.g. "gres/gpu", into a TRES.
func TRESFromKey(key string, count float64) TRES {
	typ, name, _ := strings.Cut(key, "/")
	return TRES{Type: typ, Name: name, Count: count}
}

// FormatTRES writes a TRES list in the form ParseTRES reads.
func FormatTRES(list []TRES) string {
	parts := make([]string, 0, len(list))
	for _, t := range list {
		parts = append(parts, t.Key()+"="+strconv.FormatFloat(t.Count, 'f', -1, 64))
	}
	return strings.Join(parts, ",")
}
