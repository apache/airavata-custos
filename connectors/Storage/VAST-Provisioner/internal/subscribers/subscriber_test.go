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
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

var layout = []Mount{
	{Path: "/home/{user}", Owner: "{user}", Group: "{user}", Mode: 0o700, HardLimit: 1, HardLimitInodes: 1},
	{Path: "/scratch/{user}", Owner: "{user}", Group: "{user}", Mode: 0o700, HardLimit: 1, HardLimitInodes: 1},
	{Path: "/project/{allocation}", Owner: "0", Group: "{allocation}", Mode: 0o2750, HardLimit: 1, HardLimitInodes: 1},
	{Path: "/project/{allocation}/{user}", Owner: "{user}", Group: "{user}", Mode: 0o700},
}

// oneAllocation serves one active allocation whose only member is the account.
func oneAllocation() *service.CoreServiceMock {
	group, provisioned := "proj-a", time.Now()
	return &service.CoreServiceMock{
		ListAllocationsForUserFunc: func(context.Context, string) ([]models.ComputeAllocationMembership, error) {
			return []models.ComputeAllocationMembership{{ComputeAllocationID: "a1", MembershipStatus: models.ACTIVE}}, nil
		},
		GetComputeAllocationFunc: func(context.Context, string) (*models.ComputeAllocation, error) {
			return &models.ComputeAllocation{ID: "a1", ComputeClusterID: "c1", Status: models.ACTIVE, PosixGroup: &group}, nil
		},
		ListMembersForAllocationFunc: func(context.Context, string) ([]store.MembershipWithUser, error) {
			return []store.MembershipWithUser{{
				ComputeAllocationMembership: models.ComputeAllocationMembership{MembershipStatus: models.ACTIVE},
				LocalUsername:               "custos-jdoe", ProvisionedAt: &provisioned,
			}}, nil
		},
	}
}

func TestEnsureUser(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		cluster       string
		provisionedAt *time.Time
		want          []string
	}{
		{name: "creates in order", cluster: "c1", provisionedAt: &now, want: []string{"folder /home/custos-jdoe", "quota /home/custos-jdoe", "folder /scratch/custos-jdoe", "quota /scratch/custos-jdoe", "folder /project/proj-a", "quota /project/proj-a", "folder /project/proj-a/custos-jdoe"}},
		{name: "skips unprovisioned", cluster: "c1"},
		{name: "skips unlisted cluster", cluster: "c2", provisionedAt: &now},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log []string
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
				}
			}))
			defer srv.Close()
			s := NewStorageSubscriber(nil, oneAllocation(), map[string]Cluster{"c1": {VMS: client.New(srv.URL, "u", "p", 1), Mounts: layout}})
			cu := models.ComputeClusterUser{ComputeClusterID: tt.cluster, LocalUsername: "custos-jdoe", ProvisionedAt: tt.provisionedAt}
			if err := s.ensureUser(context.Background(), cu); err != nil || !slices.Equal(log, tt.want) {
				t.Fatalf("err=%v calls=%v, want %v", err, log, tt.want)
			}
		})
	}
}
