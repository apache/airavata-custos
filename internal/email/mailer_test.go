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

package email

import (
	"strings"
	"testing"
)

func TestRenderAccountReadyFullData(t *testing.T) {
	text, html, err := RenderAccountReady(AccountReadyData{
		Username:    "trial-jdoe",
		PortalURL:   "https://portal.example.org",
		ClusterHost: "login.example.org",
		SiteName:    "Example HPC",
		FirstName:   "Jane",
		ExpiresOn:   "August 12, 2026",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{"Hi Jane", "trial-jdoe", "ssh trial-jdoe@login.example.org", "August 12, 2026", "https://portal.example.org"} {
		if !strings.Contains(text, want) {
			t.Errorf("text part missing %q", want)
		}
		if !strings.Contains(html, want) && !strings.Contains(html, strings.ReplaceAll(want, "@", "@")) {
			t.Errorf("html part missing %q", want)
		}
	}
}

func TestRenderAccountReadyOptionalFieldsOmitted(t *testing.T) {
	text, html, err := RenderAccountReady(AccountReadyData{
		Username: "trial-x", PortalURL: "https://p", ClusterHost: "h", SiteName: "S",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(text, "Hi,") || strings.Contains(text, "Access ends") {
		t.Errorf("optional handling wrong in text: %q", text[:80])
	}
	if strings.Contains(html, "Access ends") {
		t.Error("expiry note rendered without ExpiresOn")
	}
}

func TestEscapesHTMLInValues(t *testing.T) {
	_, html, err := RenderAccountReady(AccountReadyData{
		Username: `x<script>alert(1)</script>`, PortalURL: "https://p", ClusterHost: "h", SiteName: "S",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(html, "<script>") {
		t.Error("html part did not escape the username")
	}
}

func TestNoopWhenUnconfigured(t *testing.T) {
	if _, ok := New(Config{Host: "smtp.example.org", Username: "${EMAIL_HOST_USER}"}).(Noop); !ok {
		t.Error("unexpanded placeholders must yield the noop mailer")
	}
	if _, ok := New(Config{}).(Noop); !ok {
		t.Error("empty config must yield the noop mailer")
	}
}
