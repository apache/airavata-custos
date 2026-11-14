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

// Package airavata registers each Custos user in Airavata and adds VIRTUAL users to one
// Airavata group, so they have that group's access when they sign in to Airavata. Airavata names
// a user by their CILogon subject, so a user is registered when an identity carrying that
// subject is created, which is their first sign-in to Custos. Every call is made as an
// Airavata service account; a retry registers the user again, which makes them active.
package airavata

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Type is the connector type in the config file and the subscriber name on the bus.
const Type = "airavata"

// LoadConnector registers the identity subscriber.
func LoadConnector(_ context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup, _ *identity.Router, cfg *config.ConnectorConfig) error {
	url, _ := cfg.GetStringField("url")
	token, _ := cfg.GetStringField("service_token")
	group, _ := cfg.GetStringField("virtual_user_gid")
	if url == "" || token == "" || group == "" {
		slog.Info("airavata: url, service_token or virtual_user_gid not set; skipping")
		return nil
	}
	p := &provisioner{core: coreService, url: strings.TrimSuffix(url, "/"), token: token, group: group}
	eventBus.SubscribeUserIdentityCreated(Type, p.provision)
	return nil
}

type provisioner struct {
	core              *service.Service
	url, token, group string
}

func (p *provisioner) provision(ctx context.Context, created models.UserIdentity) error {
	// Registering overwrites the Airavata user, so only the identity that carries the
	// subject may do it: one from another source would undo a suspension there.
	id, ok := airavataUserID(created.OIDCSub)
	if !ok {
		return nil
	}
	user, err := p.core.GetUser(ctx, created.UserID)
	if err != nil {
		return err
	}
	// Airavata refuses a registration with no name, and a Custos user may have none.
	if err := p.post(ctx, "/api/v1/users", map[string]string{
		"userId": id, "email": user.Email,
		"firstName": cmp.Or(user.FirstName, user.Email), "lastName": cmp.Or(user.LastName, user.Email),
	}); err != nil {
		return err
	}
	if user.Type == models.UserTypeVirtual {
		if err := p.post(ctx, "/api/v1/groups/"+p.group+"/members", map[string]string{"userId": id, "groupRole": "MEMBER"}); err != nil {
			return err
		}
	}
	slog.InfoContext(ctx, "airavata: user registered", "user_id", user.ID, "airavata_user_id", id)
	return nil
}

// airavataUserID names a CILogon subject the way Airavata does when that user signs in:
// "http://cilogon.org/serverE/users/12345" is the user "cilogon:12345".
func airavataUserID(sub string) (string, bool) {
	i := strings.LastIndexByte(sub, '/')
	if !strings.HasPrefix(sub, "http://cilogon.org/") || i == len(sub)-1 {
		return "", false
	}
	return "cilogon:" + sub[i+1:], true
}

// post sends one request as the service account. 409 means it was already done.
func (p *provisioner) post(ctx context.Context, path string, body map[string]string) error {
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 300 || resp.StatusCode == http.StatusConflict {
		return nil
	}
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("airavata POST %s: %s: %s", path, resp.Status, msg)
}
