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
	"bytes"
	"embed"
	htmltemplate "html/template"
	texttemplate "text/template"
)

// Template names and also used in the audit rows.
const (
	RequestReceived = "request-received"
	AccountReady    = "account-ready"
	RequestDenied   = "request-denied"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Site is the deployment identity every email renders.
type Site struct {
	SiteName     string
	PortalURL    string
	ClusterHost  string
	SupportEmail string
	LogoURL      string
}

// Data fills a template. Empty optional fields drop their lines.
type Data struct {
	Site
	Subject       string
	FirstName     string
	Username      string
	Source        string
	ProjectNumber string
	Reason        string
}

type emailTemplate struct {
	text *texttemplate.Template
	html *htmltemplate.Template
}

var templates = map[string]emailTemplate{}

func init() {
	for _, name := range []string{RequestReceived, AccountReady, RequestDenied} {
		templates[name] = emailTemplate{
			text: texttemplate.Must(texttemplate.ParseFS(templateFS, "templates/"+name+".txt.tmpl")),
			html: htmltemplate.Must(htmltemplate.ParseFS(templateFS, "templates/layout.html.tmpl", "templates/"+name+".html.tmpl")),
		}
	}
}

// Render builds the subject and both bodies of the named email.
func Render(name string, d Data) (Message, error) {
	t := templates[name]
	var subject, text, html bytes.Buffer
	if err := t.text.ExecuteTemplate(&subject, "subject", d); err != nil {
		return Message{}, err
	}
	d.Subject = subject.String()
	if err := t.text.ExecuteTemplate(&text, "text", d); err != nil {
		return Message{}, err
	}
	if err := t.html.ExecuteTemplate(&html, "layout", d); err != nil {
		return Message{}, err
	}
	return Message{Subject: d.Subject, Text: text.String(), HTML: html.String()}, nil
}
