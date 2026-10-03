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
	"github.com/apache/airavata-custos/internal/store"
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
	writtenIdentity   *models.UserIdentity
	markedProvisioned []string
	members           []store.MembershipWithUser
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
	f.writtenIdentity = ui
	return ui, nil
}

func (f *fakeCore) UpdateUserIdentity(_ context.Context, ui *models.UserIdentity) error {
	f.writtenIdentity = ui
	return nil
}

func (f *fakeCore) MarkComputeClusterUserProvisioned(_ context.Context, id string) error {
	f.markedProvisioned = append(f.markedProvisioned, id)
	return nil
}

func (f *fakeCore) ListMembersForAllocation(_ context.Context, _ string) ([]store.MembershipWithUser, error) {
	return f.members, nil
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
	adminGroup  int    // CoGroup id returned for adminGroupName, 0 = not in the registry
	adminMember bool   // person 42 is already in the admin group
	account     string // username of the cluster account person 42 has, "" = none
	userGroup   int    // CoGroup id of that account's primary group, 0 = not in the registry
	posted      []client.CoGroupMemberCreateOne
	put         [][]byte // bodies of the person PUTs
	deleted     []string // paths of the DELETE calls

	members []client.CoGroupMemberListOne // membership rows served for any group, overriding adminMember
}

func mockComanageServer(t *testing.T, reg *mockRegistry) *httptest.Server {
	t.Helper()
	// composite served on every /people/<id> GET.
	account, login := "", ""
	if reg.account != "" {
		account = `"UnixClusterAccount":[{"meta":{"id":8},"unix_cluster_id":1,"username":"` + reg.account + `","uid":2000099,"status":"A"}],`
		login = `,{"identifier":"` + reg.account + `","type":"uid","login":false,"status":"A"}`
	}
	composite := `{
        "CoPerson":{"meta":{"id":42},"co_id":2,"status":"A"},
        "Name":[{"given":"E2E","family":"Test","type":"official","primary_name":true}],
        "EmailAddress":[{"mail":"e2e@example.edu","type":"official","verified":false}],
        ` + account + `
        "Identifier":[
            {"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"},
            {"identifier":"2000099","type":"uidnumber","login":false,"status":"A"}` + login + `
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
			body, _ := io.ReadAll(r.Body)
			reg.put = append(reg.put, body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, composite)
		case r.Method == http.MethodDelete:
			reg.deleted = append(reg.deleted, path)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_people.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoPeople","Version":"1.0","CoPeople":[]}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_groups.json"):
			var groups []string
			if reg.adminGroup != 0 {
				groups = append(groups, `{"Version":"1.0","Id":`+strconv.Itoa(reg.adminGroup)+`,"CoId":2,"Name":"`+adminGroupName+`","Status":"Active"}`)
			}
			if reg.userGroup != 0 {
				groups = append(groups, `{"Version":"1.0","Id":`+strconv.Itoa(reg.userGroup)+`,"CoId":2,"Name":"`+reg.account+`","Status":"Active"}`)
			}
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroups","Version":"1.0","CoGroups":[`+strings.Join(groups, ",")+`]}`)
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
			if reg.members != nil {
				b, _ := json.Marshal(reg.members)
				members = string(b)
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
			bindings := "[]"
			if reg.userGroup != 0 {
				bindings = `[{"Version":"1.0","Id":3,"UnixClusterId":1,"CoGroupId":` + strconv.Itoa(reg.userGroup) + `}]`
			}
			_, _ = io.WriteString(w, `{"ResponseType":"UnixClusterGroups","Version":"1.0","UnixClusterGroups":`+bindings+`}`)
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
