//go:build integration

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

package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/internal/db/dbtest"
	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Make sure a repeat returns the first user, ignores a posted id and status, and never returns a regular user.
func TestCreateTempAccount_RepeatReturnsTheSameUser(t *testing.T) {
	ctx := tracingtest.Context()
	database := dbtest.Open(t, dbtest.CoreDSN(t))
	bus, err := events.New(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	svc := service.New(database, bus)
	create := func(body string) (int, models.User) {
		rec := httptest.NewRecorder()
		NewHandlers(svc).createTempAccount(rec, httptest.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(body)))
		var u models.User
		_ = json.NewDecoder(rec.Body).Decode(&u)
		return rec.Code, u
	}
	org, _ := svc.CreateOrganization(ctx, &models.Organization{Name: "Example University"})
	body := `{"id": "chosen", "status": "ACTIVE", "type": "VIRTUAL", "organization_id": "` + org.ID + `", "email": "` + uuid.NewString() + `@example.edu"}`

	code, first := create(body)
	if code != http.StatusCreated || first.ID == "chosen" || first.Status != models.UserPending {
		t.Fatalf("first create: status %d, user %+v", code, first)
	}
	if code, again := create(body); code != http.StatusOK || again.ID != first.ID {
		t.Fatalf("repeat create: status %d, user %+v, want 200 and %s", code, again, first.ID)
	}
	regular, _ := svc.CreateUser(ctx, &models.User{OrganizationID: org.ID, Email: uuid.NewString() + "@example.edu"})
	if code, _ := create(strings.Replace(body, first.Email, regular.Email, 1)); code != http.StatusConflict {
		t.Fatalf("create over a regular user: status %d, want 409", code)
	}
}
