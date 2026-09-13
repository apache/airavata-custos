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

package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	adminGroupName = "cluster-admins"
	adminGroupID   = 76
)

// fakeCore is an in-memory CoreService stub that records audit-event writes
// and returns the user and identities it was given.
type fakeCore struct {
	user              *models.User
	identities        []models.UserIdentity
	auditEvents       []models.AuditEvent
	createdIdentity   *models.UserIdentity
	markedProvisioned []string
}

func (f *fakeCore) GetUser(_ context.Context, _ string) (*models.User, error) {
	return f.user, nil
}

func (f *fakeCore) ListUserIdentitiesForUser(_ context.Context, _ string) ([]models.UserIdentity, error) {
	out := make([]models.UserIdentity, len(f.identities))
	copy(out, f.identities)
	return out, nil
}

func (f *fakeCore) CreateUserIdentity(_ context.Context, ui *models.UserIdentity) (*models.UserIdentity, error) {
	f.createdIdentity = ui
	return ui, nil
}

func (f *fakeCore) MarkComputeClusterUserProvisioned(_ context.Context, id string) error {
	f.markedProvisioned = append(f.markedProvisioned, id)
	return nil
}

func (f *fakeCore) CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
	tracing.PopulateAuditIDs(ctx, &e.TraceID, &e.SpanID, &e.ParentSpanID)
	if e.ID == "" {
		e.ID = fmt.Sprintf("audit-%d", len(f.auditEvents)+1)
	}
	if e.EventTime.IsZero() {
		e.EventTime = time.Now().UTC()
	}
	f.auditEvents = append(f.auditEvents, *e)
	return e, nil
}

// mockRegistry configures mockComanageServer
type mockRegistry struct {
	adminGroup  int  // CoGroup id returned for adminGroupName, 0 = not in the registry
	adminMember bool // person 42 is already in the admin group
	posted      []client.CoGroupMemberCreateOne
}

func mockComanageServer(t *testing.T, reg *mockRegistry) *httptest.Server {
	t.Helper()
	// composite served on every /people/<id> GET.
	composite := `{
        "CoPerson":{"meta":{"id":42},"co_id":2,"status":"A"},
        "Name":[{"given":"E2E","family":"Test","type":"official","primary_name":true}],
        "EmailAddress":[{"mail":"e2e@example.edu","type":"official","verified":false}],
        "Identifier":[
            {"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"},
            {"identifier":"2000099","type":"uidnumber","login":false,"status":"A"}
        ]
    }`

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/people"):
			_, _ = io.WriteString(w, `[{"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"}]`)
		case r.Method == http.MethodGet && strings.Contains(path, "/people/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, composite)
		case r.Method == http.MethodPut && strings.Contains(path, "/people/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, composite)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_people.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoPeople","Version":"1.0","CoPeople":[]}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_groups.json"):
			groups := "[]"
			if reg.adminGroup != 0 {
				groups = `[{"Version":"1.0","Id":` + strconv.Itoa(reg.adminGroup) + `,"CoId":2,"Name":"` + adminGroupName + `","Status":"Active"}]`
			}
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroups","Version":"1.0","CoGroups":`+groups+`}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/co_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"CoGroup","Id":"55"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/identifiers.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"Identifiers","Version":"1.0","Identifiers":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/identifiers.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"Identifier","Id":"7"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_group_members.json"):
			members := "[]"
			if reg.adminMember {
				members = `[{"Version":"1.0","Id":9,"CoGroupId":` + strconv.Itoa(reg.adminGroup) + `,"Person":{"Type":"CO","Id":42},"Member":true,"Owner":false}]`
			}
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroupMembers","Version":"1.0","CoGroupMembers":`+members+`}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/co_group_members.json"):
			var body client.CoGroupMemberCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode membership body: %v", err)
			}
			reg.posted = append(reg.posted, body.CoGroupMembers...)
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"CoGroupMember","Id":"11"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/unix_cluster/unix_cluster_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"UnixClusterGroups","Version":"1.0","UnixClusterGroups":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/unix_cluster/unix_cluster_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"UnixClusterGroup","Id":"3"}`)
		default:
			t.Logf("unexpected mock request: %s %s", r.Method, path)
			http.NotFound(w, r)
		}
	}))
}

func newOrchestratorForTest(t *testing.T, srv *httptest.Server, core CoreService, adminGroup string) *Orchestrator {
	t.Helper()
	c := client.New(client.Config{
		RegistryURL:       srv.URL,
		COID:              2,
		APIUser:           "co_2.test",
		APIKey:            "k",
		PersonIDType:      "comanage_id",
		UnixClusterID:     1,
		ClusterAdminGroup: adminGroup,
		DefaultShell:      "/bin/bash",
		HomedirPrefix:     "/home/",
		HTTPTimeout:       5 * time.Second,
	})
	return &Orchestrator{c: c, core: core}
}
