//go:build livemail

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
	"context"
	"os"
	"testing"
)

// Sends all three emails to the defined EMAIL_TEST_TO address.
func TestLiveSend(t *testing.T) {
	to := os.Getenv("EMAIL_TEST_TO")
	if to == "" {
		t.Skip("EMAIL_TEST_TO not set")
	}
	sender, err := NewSMTP(os.Getenv("EMAIL_SMTP_HOST"), 587, os.Getenv("EMAIL_SMTP_USERNAME"), os.Getenv("EMAIL_SMTP_PASSWORD"), os.Getenv("EMAIL_FROM"))
	if err != nil {
		t.Fatal(err)
	}
	d := Data{
		Site:          Site{SiteName: os.Getenv("EMAIL_SITE_NAME"), PortalURL: os.Getenv("EMAIL_PORTAL_URL"), ClusterHost: os.Getenv("EMAIL_CLUSTER_HOST"), SupportEmail: os.Getenv("EMAIL_SUPPORT"), LogoURL: os.Getenv("EMAIL_LOGO_URL")},
		FirstName:     "John",
		Username:      "jdoe",
		Source:        "ACCESS",
		ProjectNumber: "CIS250123",
		Reason:        "We could not confirm an active allocation for this project.",
	}
	for _, name := range []string{RequestReceived, AccountReady, RequestDenied} {
		msg, err := Render(name, d)
		if err != nil {
			t.Fatal(err)
		}
		msg.To = to
		if err := sender.Send(context.Background(), msg); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}
