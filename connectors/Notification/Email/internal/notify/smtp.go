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

	"github.com/wneessen/go-mail"
)

type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

type Sender interface {
	Send(ctx context.Context, m Message) error
}

type SMTP struct {
	client *mail.Client
	from   string
}

// NewSMTP TLS is required only with a login, so a password never goes out unencrypted.
func NewSMTP(host string, port int, username, password, from string) (*SMTP, error) {
	opts := []mail.Option{mail.WithPort(port), mail.WithTLSPolicy(mail.TLSOpportunistic)}
	if username != "" {
		opts = []mail.Option{
			mail.WithPort(port),
			mail.WithTLSPolicy(mail.TLSMandatory),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(username),
			mail.WithPassword(password),
		}
	}
	client, err := mail.NewClient(host, opts...)
	if err != nil {
		return nil, err
	}
	return &SMTP{client: client, from: from}, nil
}

func (s *SMTP) Send(ctx context.Context, m Message) error {
	msg := mail.NewMsg()
	if err := msg.From(s.from); err != nil {
		return err
	}
	if err := msg.To(m.To); err != nil {
		return err
	}
	msg.Subject(m.Subject)
	msg.SetBodyString(mail.TypeTextPlain, m.Text)
	msg.AddAlternativeString(mail.TypeTextHTML, m.HTML)
	return s.client.DialAndSendWithContext(ctx, msg)
}
