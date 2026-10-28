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

package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure each deactivate and delete route takes its own privilege, and the write privilege that gated them before is not enough.
func TestDeactivateAndDeleteRoutes_RequireTheirOwnPrivilege(t *testing.T) {
	_, _, srv := setupTestStack(t)
	routes := []struct {
		method, path      string
		want, writeBefore models.PrivilegeKey
	}{
		{http.MethodPost, "/compute-cluster-users/any/deactivate", models.ClustersDeactivate, models.ClustersWrite},
		{http.MethodDelete, "/compute-cluster-users/any", models.ClustersDelete, models.ClustersWrite},
		{http.MethodPost, "/compute-allocations/any/deactivate", models.AllocationsDeactivate, models.AllocationsWrite},
		{http.MethodDelete, "/compute-allocations/any", models.AllocationsDelete, models.AllocationsWrite},
		{http.MethodPost, "/compute-allocation-memberships/any/deactivate", models.AllocationsDeactivate, models.AllocationsWrite},
		{http.MethodDelete, "/compute-allocation-memberships/any", models.AllocationsDelete, models.AllocationsWrite},
		{http.MethodPost, "/users/any/deactivate", models.UsersDeactivate, models.UsersWrite},
		{http.MethodDelete, "/users/any", models.UsersDelete, models.UsersWrite},
	}
	for _, r := range routes {
		for priv, want := range map[models.PrivilegeKey]int{r.writeBefore: http.StatusForbidden, r.want: http.StatusNotFound} {
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, withTestCaller(httptest.NewRequest(r.method, r.path, nil), "u-1", priv))
			if rr.Code != want {
				t.Errorf("%s %s with %s: got %d, want %d", r.method, r.path, priv, rr.Code, want)
			}
		}
	}
}

// Make sure deleting an active account, allocation or user is refused, and deactivating it first lets the delete through.
func TestDeleteActiveRow_ConflictsUntilDeactivated(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	cluster := seedCluster(t, svc)
	cu, err := svc.CreateComputeClusterUser(t.Context(), &models.ComputeClusterUser{ComputeClusterID: cluster,
		UserID: seedUser(t, database, fmt.Sprintf("u-%s@example.edu", uuid.NewString())), LocalUsername: "u-" + uuid.NewString()[:8],
		ApprovalStatus: models.ClusterAccountApproved, ReviewedBy: &admin})
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}
	alloc := seedAllocation(t, database, svc, cluster, "deactivate-"+uuid.NewString()[:8])
	user := seedUser(t, database, fmt.Sprintf("v-%s@example.edu", uuid.NewString()))
	steps := []struct {
		method, path string
		priv         models.PrivilegeKey
		want         int
	}{
		{http.MethodDelete, "/compute-cluster-users/" + cu.ID, models.ClustersDelete, http.StatusConflict},
		{http.MethodPost, "/compute-cluster-users/" + cu.ID + "/deactivate", models.ClustersDeactivate, http.StatusNoContent},
		{http.MethodDelete, "/compute-cluster-users/" + cu.ID, models.ClustersDelete, http.StatusNoContent},
		{http.MethodDelete, "/compute-allocations/" + alloc.ID, models.AllocationsDelete, http.StatusConflict},
		{http.MethodPost, "/compute-allocations/" + alloc.ID + "/deactivate", models.AllocationsDeactivate, http.StatusNoContent},
		{http.MethodDelete, "/compute-allocations/" + alloc.ID, models.AllocationsDelete, http.StatusNoContent},
		{http.MethodDelete, "/users/" + user, models.UsersDelete, http.StatusConflict},
		{http.MethodPost, "/users/" + user + "/deactivate", models.UsersDeactivate, http.StatusNoContent},
		{http.MethodDelete, "/users/" + user, models.UsersDelete, http.StatusNoContent},
	}
	for _, s := range steps {
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, withTestCaller(httptest.NewRequest(s.method, s.path, nil), admin, s.priv))
		if rr.Code != s.want {
			t.Fatalf("%s %s: got %d, want %d: %s", s.method, s.path, rr.Code, s.want, rr.Body.String())
		}
	}
}
