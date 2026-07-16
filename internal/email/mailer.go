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

// Package email renders and sends product notification mail.
package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"github.com/wneessen/go-mail"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// AccountReadyData fills the account-ready templates. FirstName and
// ExpiresOn are optional; empty values omit their sections.
type AccountReadyData struct {
	Username    string
	PortalURL   string
	ClusterHost string
	SiteName    string
	FirstName   string
	ExpiresOn   string
}

// Mailer sends product mail. Implementations must be safe for concurrent use.
type Mailer interface {
	SendAccountReady(ctx context.Context, to string, data AccountReadyData) error
}

// Noop drops every message; the default when no relay is configured.
type Noop struct{}

func (Noop) SendAccountReady(context.Context, string, AccountReadyData) error { return nil }

// Config carries the relay settings plus the deployment identity the
// templates render.
type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	From        string
	SiteName    string
	PortalURL   string
	ClusterHost string
}

// Configured reports whether the relay settings are usable. Values still
// carrying an unexpanded "${...}" placeholder count as unset.
func (c Config) Configured() bool {
	for _, v := range []string{c.Host, c.Username, c.Password, c.SiteName, c.PortalURL, c.ClusterHost} {
		if v == "" || strings.Contains(v, "${") {
			return false
		}
	}
	return true
}

type smtpMailer struct {
	cfg      Config
	htmlTmpl *htmltemplate.Template
	textTmpl *texttemplate.Template
}

// New builds an SMTP-backed mailer, or Noop when the config is incomplete.
func New(cfg Config) Mailer {
	if !cfg.Configured() {
		return Noop{}
	}
	if cfg.From == "" || strings.Contains(cfg.From, "${") {
		cfg.From = cfg.Username
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	return &smtpMailer{
		cfg:      cfg,
		htmlTmpl: htmltemplate.Must(htmltemplate.ParseFS(templateFS, "templates/account-ready.html.tmpl")),
		textTmpl: texttemplate.Must(texttemplate.ParseFS(templateFS, "templates/account-ready.txt.tmpl")),
	}
}

// RenderAccountReady returns the text and HTML parts; exported for tests.
func RenderAccountReady(data AccountReadyData) (text, html string, err error) {
	textTmpl, err := texttemplate.ParseFS(templateFS, "templates/account-ready.txt.tmpl")
	if err != nil {
		return "", "", err
	}
	htmlTmpl, err := htmltemplate.ParseFS(templateFS, "templates/account-ready.html.tmpl")
	if err != nil {
		return "", "", err
	}
	var tb, hb bytes.Buffer
	if err := textTmpl.Execute(&tb, data); err != nil {
		return "", "", err
	}
	if err := htmlTmpl.Execute(&hb, data); err != nil {
		return "", "", err
	}
	return tb.String(), hb.String(), nil
}

func (m *smtpMailer) SendAccountReady(ctx context.Context, to string, data AccountReadyData) error {
	if data.SiteName == "" {
		data.SiteName = m.cfg.SiteName
	}
	if data.PortalURL == "" {
		data.PortalURL = m.cfg.PortalURL
	}
	if data.ClusterHost == "" {
		data.ClusterHost = m.cfg.ClusterHost
	}

	var tb, hb bytes.Buffer
	if err := m.textTmpl.Execute(&tb, data); err != nil {
		return fmt.Errorf("render text part: %w", err)
	}
	if err := m.htmlTmpl.Execute(&hb, data); err != nil {
		return fmt.Errorf("render html part: %w", err)
	}

	msg := mail.NewMsg()
	if err := msg.From(m.cfg.From); err != nil {
		return fmt.Errorf("set from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("set recipient: %w", err)
	}
	msg.Subject(fmt.Sprintf("Your %s account is ready", data.SiteName))
	msg.SetBodyString(mail.TypeTextPlain, tb.String())
	msg.AddAlternativeString(mail.TypeTextHTML, hb.String())

	client, err := mail.NewClient(m.cfg.Host,
		mail.WithPort(m.cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(m.cfg.Username),
		mail.WithPassword(m.cfg.Password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return fmt.Errorf("build smtp client: %w", err)
	}
	return client.DialAndSendWithContext(ctx, msg)
}
