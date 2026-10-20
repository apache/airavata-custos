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

package subscribers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

var layout = []Mount{
	{Path: "/home/{user}", Owner: "{user}", Group: "{user}", Mode: 0o700, HardLimit: 1, HardLimitInodes: 1},
	{Path: "/scratch/{user}", Owner: "{user}", Group: "{user}", Mode: 0o700, HardLimit: 1, HardLimitInodes: 1, View: true},
	{Path: "/project/{allocation}", Owner: "0", Group: "{allocation}", Mode: 0o2750, HardLimit: 1, HardLimitInodes: 1, View: true},
	{Path: "/project/{allocation}/{user}", Owner: "{user}", Group: "{user}", Mode: 0o700, View: true},
}

var (
	groupMounts  = []string{"folder /project/proj-a", "view /project/proj-a", "quota /project/proj-a"}
	jdoeOwn      = []string{"folder /home/custos-jdoe", "quota /home/custos-jdoe", "folder /scratch/custos-jdoe", "view /scratch/custos-jdoe", "quota /scratch/custos-jdoe"}
	jdoeMember   = []string{"folder /project/proj-a/custos-jdoe", "view /project/proj-a/custos-jdoe"}
	asmithMember = []string{"folder /project/proj-a/custos-asmith", "view /project/proj-a/custos-asmith"}
)

// allocationWithMembers serves one active allocation on cluster c1 with two
// provisioned members, jdoe (u1) and asmith (u2), and one member whose account is not provisioned yet (u3).
func allocationWithMembers() *service.CoreServiceMock {
	group, provisioned := "proj-a", time.Now()
	member := func(userID, username string, at *time.Time) store.MembershipWithUser {
		return store.MembershipWithUser{
			ComputeAllocationMembership: models.ComputeAllocationMembership{ComputeAllocationID: "a1", UserID: userID, MembershipStatus: models.ACTIVE},
			LocalUsername:               username, ProvisionedAt: at,
		}
	}
	return &service.CoreServiceMock{
		ListAllocationsForUserFunc: func(context.Context, string) ([]models.ComputeAllocationMembership, error) {
			return []models.ComputeAllocationMembership{{ComputeAllocationID: "a1", MembershipStatus: models.ACTIVE}}, nil
		},
		GetComputeAllocationFunc: func(context.Context, string) (*models.ComputeAllocation, error) {
			return &models.ComputeAllocation{ID: "a1", ComputeClusterID: "c1", Status: models.ACTIVE, PosixGroup: &group, GroupProvisionedAt: &provisioned}, nil
		},
		ListMembersForAllocationFunc: func(context.Context, string) ([]store.MembershipWithUser, error) {
			return []store.MembershipWithUser{member("u1", "custos-jdoe", &provisioned), member("u2", "custos-asmith", &provisioned), member("u3", "custos-pending", nil)}, nil
		},
	}
}

// recordingVMS stands in for VMS and records each create as "<kind> <path>", starting from held.
// A deleted view becomes "removed".
func recordingVMS(t *testing.T, held ...string) (Cluster, *[]string) {
	t.Helper()
	log := slices.Clone(held)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Path string }
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch r.URL.Path {
		case "/api/token/":
			_, _ = w.Write([]byte(`{"access":"t"}`))
		case "/api/folders/create_folder/":
			log = append(log, "folder "+body.Path)
		case "/api/quotas/":
			log = append(log, "quota "+body.Path)
		case "/api/views/":
			if r.Method != http.MethodGet {
				log = append(log, "view "+body.Path)
				return
			}
			var views []map[string]any
			for i, e := range log {
				if p, ok := strings.CutPrefix(e, "view "); ok {
					views = append(views, map[string]any{"id": i, "path": p})
				}
			}
			_ = json.NewEncoder(w).Encode(views)
		default:
			if id, err := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/views/"), "/")); err == nil && r.Method == http.MethodDelete {
				log[id] = "removed"
			}
		}
	}))
	t.Cleanup(srv.Close)
	return Cluster{VMS: client.New(srv.URL, "u", "p", 1), Mounts: layout}, &log
}

func join(parts ...[]string) []string {
	var all []string
	for _, p := range parts {
		all = append(all, p...)
	}
	return all
}

// Make sure a provision event creates the account's own directories and its
// directory in each of its projects, and nothing for the other members.
func TestEnsureUser(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		cluster       string
		provisionedAt *time.Time
		want          []string
	}{
		{name: "own directories, then its project directory", cluster: "c1", provisionedAt: &now, want: join(jdoeOwn, groupMounts, jdoeMember)},
		{name: "skips unprovisioned", cluster: "c1"},
		{name: "skips unlisted cluster", cluster: "c2", provisionedAt: &now},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, log := recordingVMS(t)
			s := NewStorageSubscriber(nil, allocationWithMembers(), map[string]Cluster{"c1": c})
			cu := models.ComputeClusterUser{ComputeClusterID: tt.cluster, UserID: "u1", LocalUsername: "custos-jdoe", ProvisionedAt: tt.provisionedAt}
			if err := s.ensureUser(context.Background(), cu); err != nil || !slices.Equal(*log, tt.want) {
				t.Fatalf("err=%v calls=%v, want %v", err, *log, tt.want)
			}
		})
	}
}

// Make sure a membership event creates only that member's project directory and nothing until the member's account is provisioned.
func TestEnsureMember(t *testing.T) {
	tests := []struct {
		name, userID string
		want         []string
	}{
		{name: "the member's directory only", userID: "u1", want: join(groupMounts, jdoeMember)},
		{name: "nothing for an unprovisioned member", userID: "u3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, log := recordingVMS(t)
			s := NewStorageSubscriber(nil, allocationWithMembers(), map[string]Cluster{"c1": c})
			m := models.ComputeAllocationMembership{ComputeAllocationID: "a1", UserID: tt.userID, MembershipStatus: models.ACTIVE}
			if err := s.ensureMember(context.Background(), m); err != nil || !slices.Equal(*log, tt.want) {
				t.Fatalf("err=%v calls=%v, want %v", err, *log, tt.want)
			}
		})
	}
}

// Make sure an allocation creates nothing until its group is on the cluster, then the project's and every provisioned member's directories.
func TestEnsureAllocation(t *testing.T) {
	group, now := "proj-a", time.Now()
	tests := []struct {
		name               string
		groupProvisionedAt *time.Time
		want               []string
	}{
		{name: "nothing before the group is provisioned"},
		{name: "project and members once provisioned", groupProvisionedAt: &now, want: join(groupMounts, jdoeMember, asmithMember)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, log := recordingVMS(t)
			s := NewStorageSubscriber(nil, allocationWithMembers(), map[string]Cluster{"c1": c})
			a := models.ComputeAllocation{ID: "a1", ComputeClusterID: "c1", Status: models.ACTIVE, PosixGroup: &group, GroupProvisionedAt: tt.groupProvisionedAt}
			if err := s.ensureAllocation(context.Background(), a); err != nil || !slices.Equal(*log, tt.want) {
				t.Fatalf("err=%v calls=%v, want %v", err, *log, tt.want)
			}
		})
	}
}

// Make sure a deletion removes only its own views, keeps folders and quotas, and repeats harmlessly.
func TestRemove(t *testing.T) {
	group := "proj-a"
	provisioned := join(jdoeOwn, groupMounts, jdoeMember, asmithMember)
	withoutViews := func(of string) []string {
		return slices.DeleteFunc(slices.Clone(provisioned), func(e string) bool { return strings.HasPrefix(e, "view "+of) })
	}
	tests := []struct {
		name   string
		remove func(*StorageSubscriber) error
		want   []string
	}{
		{name: "account deleted, allocation kept", remove: func(s *StorageSubscriber) error {
			return s.removeUser(context.Background(), models.ComputeClusterUser{ComputeClusterID: "c1", UserID: "u1", LocalUsername: "custos-jdoe"})
		}, want: slices.DeleteFunc(withoutViews("/scratch/custos-jdoe"), func(e string) bool { return e == "view /project/proj-a/custos-jdoe" })},
		{name: "allocation deleted, accounts kept", remove: func(s *StorageSubscriber) error {
			return s.removeAllocation(context.Background(), models.ComputeAllocation{ID: "a1", ComputeClusterID: "c1", PosixGroup: &group})
		}, want: withoutViews("/project/")},
		{name: "nothing for an empty name", remove: func(s *StorageSubscriber) error {
			return s.removeUser(context.Background(), models.ComputeClusterUser{ComputeClusterID: "c1", UserID: "u1", LocalUsername: ""})
		}, want: provisioned},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, log := recordingVMS(t, provisioned...)
			s := NewStorageSubscriber(nil, allocationWithMembers(), map[string]Cluster{"c1": c})
			for run := range 2 {
				if err := tt.remove(s); err != nil {
					t.Fatalf("run %d: %v", run+1, err)
				}
				left := slices.DeleteFunc(slices.Clone(*log), func(e string) bool { return e == "removed" })
				if !slices.Equal(left, tt.want) {
					t.Fatalf("run %d: left %v, want %v", run+1, left, tt.want)
				}
			}
		})
	}
}
