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

// Package airavata registers each Custos user in Airavata when their CILogon identity is created, and adds VIRTUAL
// users to one group. Calls run as an Airavata service account.
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
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Type is the connector type in the config file and the subscriber name on the bus.
const Type = "airavata"

// LoadConnector registers the identity subscriber.
func LoadConnector(_ context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup, _ *identity.Router, cfg *config.ConnectorConfig) error {
	settings := make(map[string]string, 3)
	for _, key := range []string{"base_url", "service_token", "virtual_users_group_id"} {
		value, _ := cfg.GetStringField(key)
		if value == "" {
			slog.Info("airavata: required setting not set; skipping", "setting", key)
			health.Record(Type, "config", fmt.Errorf("required setting %s not set", key))
			return nil
		}
		settings[key] = value
	}
	r := &registrar{core: coreService, baseURL: strings.TrimSuffix(settings["base_url"], "/"), token: settings["service_token"], virtualUsersGroupID: settings["virtual_users_group_id"]}
	eventBus.SubscribeUserIdentityCreated(Type, r.register)
	return nil
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

type registrar struct {
	core                                *service.Service
	baseURL, token, virtualUsersGroupID string
}

func (r *registrar) register(ctx context.Context, created models.UserIdentity) error {
	// Only CILogon registers, since registering overwrites the user and would undo a deactivation.
	// ".../users/12345" becomes "cilogon:12345".
	sub := created.OIDCSub
	i := strings.LastIndexByte(sub, '/')
	if !strings.HasPrefix(sub, "http://cilogon.org/") || i == len(sub)-1 {
		return nil
	}
	airavataUserID := "cilogon:" + sub[i+1:]
	user, err := r.core.GetUser(ctx, created.UserID)
	if err != nil {
		return err
	}
	// Airavata refuses a registration with no name, and a Custos user may have none.
	if err := r.post(ctx, "/api/v1/users", map[string]string{
		"userId": airavataUserID, "email": user.Email,
		"firstName": cmp.Or(user.FirstName, user.Email), "lastName": cmp.Or(user.LastName, user.Email),
	}); err != nil {
		return err
	}
	if user.Type == models.UserTypeVirtual {
		if err := r.post(ctx, "/api/v1/groups/"+r.virtualUsersGroupID+"/members", map[string]string{"userId": airavataUserID, "groupRole": "MEMBER"}); err != nil {
			return err
		}
	}
	slog.InfoContext(ctx, "airavata: user registered", "user_id", user.ID, "airavata_user_id", airavataUserID)
	return nil
}

// post sends one request as the service account; 409 means already done.
func (r *registrar) post(ctx context.Context, path string, body map[string]string) error {
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
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
