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
	"io/fs"
	"os"
	"path/filepath"
	texttemplate "text/template"
)

// Template names and also used in the audit rows.
const (
	RequestReceived = "request-received"
	AccountReady    = "account-ready"
	RequestDenied   = "request-denied"
	RequestApproved = "request-approved"
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
	if err := LoadTemplates(""); err != nil {
		panic(err)
	}
}

// LoadTemplates parses the emails from the built-in files. A file of the same
// name in the template dir replaces the built-in one, so a deployment can restyle the layout.
func LoadTemplates(dir string) error {
	read := func(name string) (string, error) {
		if dir != "" {
			b, err := os.ReadFile(filepath.Join(dir, name))
			if err == nil {
				return string(b), nil
			}
			if !os.IsNotExist(err) {
				return "", err
			}
		}
		b, err := fs.ReadFile(templateFS, "templates/"+name)
		return string(b), err
	}
	layout, err := read("layout.html.tmpl")
	if err != nil {
		return err
	}
	loaded := map[string]emailTemplate{}
	for _, name := range []string{RequestReceived, AccountReady, RequestDenied, RequestApproved} {
		text, err := read(name + ".txt.tmpl")
		if err != nil {
			return err
		}
		html, err := read(name + ".html.tmpl")
		if err != nil {
			return err
		}
		var t emailTemplate
		if t.text, err = texttemplate.New(name + ".txt.tmpl").Parse(text); err != nil {
			return err
		}
		if t.html, err = htmltemplate.New("layout.html.tmpl").Parse(layout); err != nil {
			return err
		}
		if t.html, err = t.html.New(name + ".html.tmpl").Parse(html); err != nil {
			return err
		}
		loaded[name] = t
	}
	templates = loaded
	return nil
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
