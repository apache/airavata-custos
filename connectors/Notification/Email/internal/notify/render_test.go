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
	"strings"
	"testing"
)

// Make sure an account with no project yet, like one an admin adds, renders
// without empty "Requested through" and "Project" lines.
func TestRender_LeavesOutMissingProject(t *testing.T) {
	for _, name := range []string{RequestReceived, RequestDenied} {
		m, err := Render(name, Data{Site: Site{SiteName: "Example HPC"}, Username: "jdoe", Reason: "No allocation."})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, label := range []string{"Requested through", "Project"} {
			if strings.Contains(m.Text, label) || strings.Contains(m.HTML, label) {
				t.Errorf("%s shows %q with no project", name, label)
			}
		}
	}
}

// Make sure the admin's free-text reason cannot inject markup into the email.
func TestRender_EscapesTheReason(t *testing.T) {
	m, err := Render(RequestDenied, Data{Site: Site{SiteName: "Example HPC"}, Reason: `<a href="https://example.com">click</a>`})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(m.HTML, `<a href="https://example.com">`) {
		t.Fatal("the reason was rendered as markup")
	}
}
