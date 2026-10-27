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

package notify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Make sure a deployment can replace just the layout and keep the built-in
// emails, so a re-style does not need a copy of every file.
func TestLoadTemplates_OverridesOneFileAndKeepsTheRest(t *testing.T) {
	dir := t.TempDir()
	layout := `{{define "layout"}}<p>Example University</p>{{template "body" .}}{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "layout.html.tmpl"), []byte(layout), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = LoadTemplates("") })

	if err := LoadTemplates(dir); err != nil {
		t.Fatal(err)
	}
	m, err := Render(AccountReady, Data{Site: Site{SiteName: "Example HPC"}, Username: "jdoe"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(m.HTML, "Example University") {
		t.Errorf("layout override not used: %q", m.HTML)
	}
	if !strings.Contains(m.HTML, "Your account is ready") || !strings.Contains(m.Text, "jdoe") {
		t.Errorf("built-in email body lost after the layout override")
	}
}

// Make sure a broken override fails at startup instead of sending a broken email.
func TestLoadTemplates_RejectsABrokenOverride(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "account-ready.txt.tmpl"), []byte("{{define"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = LoadTemplates("") })

	if err := LoadTemplates(dir); err == nil {
		t.Fatal("expected an error for a template that does not parse")
	}
}
