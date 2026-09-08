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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/models"
)

const adminGroupName = "cluster-admins"

// adminGroupServer serves the person composite and the admin group lookup.
// existingMember decides whether the person is already in the group, and every
// POSTed membership body is appended to posted.
func adminGroupServer(t *testing.T, groupFound, existingMember bool, posted *[]client.CoGroupMemberCreateOne) *httptest.Server {
	t.Helper()
	composite := `{
        "CoPerson":{"meta":{"id":42},"co_id":2,"status":"A"},
        "Identifier":[{"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"}]
    }`

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case r.Method == http.MethodGet && strings.Contains(path, "/people/"):
			_, _ = io.WriteString(w, composite)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_groups.json"):
			if !groupFound {
				_, _ = io.WriteString(w, `{"ResponseType":"CoGroups","Version":"1.0","CoGroups":[]}`)
				return
			}
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroups","Version":"1.0","CoGroups":[
                {"Version":"1.0","Id":76,"CoId":2,"Name":"`+adminGroupName+`","Status":"Active"}]}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_group_members.json"):
			if !existingMember {
				_, _ = io.WriteString(w, `{"ResponseType":"CoGroupMembers","Version":"1.0","CoGroupMembers":[]}`)
				return
			}
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroupMembers","Version":"1.0","CoGroupMembers":[
                {"Version":"1.0","Id":9,"CoGroupId":76,"Person":{"Type":"CO","Id":42},"Member":true,"Owner":false}]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/co_group_members.json"):
			var body client.CoGroupMemberCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode membership body: %v", err)
			}
			*posted = append(*posted, body.CoGroupMembers...)
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"CoGroupMember","Id":"11"}`)
		default:
			t.Logf("unexpected mock request: %s %s", r.Method, path)
			http.NotFound(w, r)
		}
	}))
}

func adminOrchestrator(srv *httptest.Server, groupName string, core CoreService) *Orchestrator {
	c := client.New(client.Config{
		RegistryURL:       srv.URL,
		COID:              2,
		APIUser:           "co_2.test",
		APIKey:            "k",
		PersonIDType:      "comanage_id",
		UnixClusterID:     1,
		ClusterAdminGroup: groupName,
		HTTPTimeout:       5 * time.Second,
	})
	return &Orchestrator{c: c, core: core}
}

func adminCore() *fakeCore {
	return &fakeCore{
		user: &models.User{ID: "user-1", Email: "admin@example.invalid"},
		identities: []models.UserIdentity{
			{UserID: "user-1", Source: "comanage", ExternalID: "Person100099"},
		},
	}
}

func TestEnsureClusterAdminGroup_JoinsAsPlainMember(t *testing.T) {
	var posted []client.CoGroupMemberCreateOne
	srv := adminGroupServer(t, true, false, &posted)
	defer srv.Close()

	core := adminCore()
	orch := adminOrchestrator(srv, adminGroupName, core)

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", LocalUsername: "admin1", AccessLevel: models.ClusterAccessAdmin}
	if err := orch.EnsureClusterAdminGroup(context.Background(), cu); err != nil {
		t.Fatalf("EnsureClusterAdminGroup: %v", err)
	}

	if len(posted) != 1 {
		t.Fatalf("want 1 membership POST, got %d", len(posted))
	}
	got := posted[0]
	if got.CoGroupId != 76 {
		t.Errorf("co_group_id = %d, want 76", got.CoGroupId)
	}
	if got.Person.Id != 42 {
		t.Errorf("person id = %d, want 42", got.Person.Id)
	}
	if !got.Member {
		t.Error("member = false, want true")
	}
	if got.Owner {
		t.Error("owner = true, want false: an owner can change the group in the registry")
	}

	var joined bool
	for _, e := range core.auditEvents {
		if e.EventType == "ComanageClusterAdminGroupJoined" {
			joined = true
		}
	}
	if !joined {
		t.Error("no ComanageClusterAdminGroupJoined audit event")
	}
}

func TestEnsureClusterAdminGroup_AlreadyMemberDoesNotPost(t *testing.T) {
	var posted []client.CoGroupMemberCreateOne
	srv := adminGroupServer(t, true, true, &posted)
	defer srv.Close()

	orch := adminOrchestrator(srv, adminGroupName, adminCore())

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", LocalUsername: "admin1", AccessLevel: models.ClusterAccessAdmin}
	if err := orch.EnsureClusterAdminGroup(context.Background(), cu); err != nil {
		t.Fatalf("EnsureClusterAdminGroup: %v", err)
	}
	if len(posted) != 0 {
		t.Fatalf("want no membership POST for an existing member, got %d", len(posted))
	}
}

func TestEnsureClusterAdminGroup_MissingGroupFails(t *testing.T) {
	tests := []struct {
		name       string
		groupName  string
		groupFound bool
		wantErr    string
	}{
		{"not configured", "", true, "not configured"},
		{"not in registry", adminGroupName, false, "not in the registry"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var posted []client.CoGroupMemberCreateOne
			srv := adminGroupServer(t, tt.groupFound, false, &posted)
			defer srv.Close()

			core := adminCore()
			orch := adminOrchestrator(srv, tt.groupName, core)

			cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", AccessLevel: models.ClusterAccessAdmin}
			err := orch.EnsureClusterAdminGroup(context.Background(), cu)
			if err == nil {
				t.Fatal("want an error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
			if len(posted) != 0 {
				t.Errorf("want no membership POST, got %d", len(posted))
			}

			var failed bool
			for _, e := range core.auditEvents {
				if e.EventType == "ComanageProvisioningFailed" {
					failed = true
				}
			}
			if !failed {
				t.Error("no ComanageProvisioningFailed audit event")
			}
		})
	}
}
