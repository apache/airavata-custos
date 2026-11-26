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

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/drift"
	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/subscribers"
	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type fakeSnapshots struct {
	snap subscribers.Snapshot
	err  error
}

func (f fakeSnapshots) Snapshot(context.Context, models.ComputeCluster) (subscribers.Snapshot, error) {
	return f.snap, f.err
}

type fakeSlurm struct {
	associations []client.Association
	err          error
	deletedQoS   []string
}

func (f *fakeSlurm) ListAssociations(client.AssocFilter) ([]client.Association, error) {
	return f.associations, f.err
}
func (f *fakeSlurm) DeleteQoS(name string) error {
	f.deletedQoS = append(f.deletedQoS, name)
	return nil
}
func (f *fakeSlurm) DeleteAssociation(client.AssocFilter) error { return nil }

// calls records the core writes a resolve made, in order.
type calls struct {
	list []string
}

func (c *calls) add(s string) { c.list = append(c.list, s) }

func newCore(c *calls) *service.CoreServiceMock {
	return &service.CoreServiceMock{
		GetComputeClusterFunc: func(_ context.Context, id string) (*models.ComputeCluster, error) {
			if id != "c1" {
				return nil, service.ErrNotFound
			}
			return &models.ComputeCluster{ID: "c1", Name: "alpha"}, nil
		},
		ListComputeClustersFunc: func(context.Context) ([]models.ComputeCluster, error) {
			return []models.ComputeCluster{{ID: "c1", Name: "alpha"}}, nil
		},
		ListComputeAllocationsByClusterFunc: func(_ context.Context, clusterID string) ([]models.ComputeAllocation, error) {
			return []models.ComputeAllocation{{ID: "al1", Name: "A1", ComputeClusterID: clusterID}}, nil
		},
		CreateAuditEventFunc: func(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) { return e, nil },
		SetSlurmPendingFunc: func(_ context.Context, kind service.SlurmPendingKind, id string, _ *time.Time) error {
			c.add("mark-" + string(kind) + ":" + id)
			return nil
		},
		ListComputeClusterQoSFunc: func(context.Context, string, string) ([]models.ComputeClusterQoS, error) {
			return []models.ComputeClusterQoS{{ID: "q-normal", Name: "normal", ComputeClusterID: "c1"}}, nil
		},
	}
}

var (
	qosConflict = drift.Conflict{Key: "qos:gpu", Kind: drift.KindQoSDefinition, Name: "gpu",
		Custos: models.ComputeClusterQoS{Name: "gpu", Priority: 5}, Slurm: models.ComputeClusterQoS{Name: "gpu", Priority: 9, MaxTRESPerUser: "gres/gpu=4"},
		Choices: []drift.Choice{drift.ChoiceCustos, drift.ChoiceSlurm, drift.ChoiceManual}}
	allocQoSConflict = drift.Conflict{Key: "account:a1:qos", Kind: drift.KindAllocationQoS, AllocationID: "al1",
		Custos:  drift.QoSValue{QoS: []string{"normal"}, Default: "normal"},
		Slurm:   drift.QoSValue{QoS: []string{"normal", "openfold"}, Default: "normal"},
		Choices: []drift.Choice{drift.ChoiceCustos, drift.ChoiceSlurm, drift.ChoiceManual}}
	custosOnlyConflict = drift.Conflict{Key: "member:a1:u1:account", Kind: drift.KindMemberLimit, AllocationID: "al1", MembershipID: "m1",
		Custos: float64(4), Slurm: float64(8), Choices: []drift.Choice{drift.ChoiceCustos}}
)

func serve(t *testing.T, core *service.CoreServiceMock, snaps fakeSnapshots, slurm *fakeSlurm, method, path, body string, privileges ...models.PrivilegeKey) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	router := identity.NewRouter(mux)
	NewHandlers(core, snaps, slurm).RegisterRoutes(router)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req = req.WithContext(identity.WithPrivilegesForTest(req.Context(), privileges))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func resolve(t *testing.T, core *service.CoreServiceMock, snap subscribers.Snapshot, body string, privileges ...models.PrivilegeKey) *httptest.ResponseRecorder {
	t.Helper()
	return serve(t, core, fakeSnapshots{snap: snap}, &fakeSlurm{}, "POST", "/connectors/slurm/clusters/c1/conflicts/resolve", body,
		privileges...)
}

func snapshotOf(conflicts ...drift.Conflict) subscribers.Snapshot {
	return subscribers.Snapshot{Conflicts: conflicts, Want: drift.Desired{DefaultQoS: "normal", SystemAccounts: map[string]bool{"root": true}}}
}

func seen(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestCluster_ViewsLiveStateFromOneSnapshot(t *testing.T) {
	max, yes := int64(2880), true
	snap := subscribers.Snapshot{Want: drift.Desired{SystemAccounts: map[string]bool{"root": true, "none": true, "ops": true}}, Have: drift.Live{
		Partitions: []client.Partition{{Name: "batch", Nodes: "n[1-4]", State: "UP", MaxTimeMinutes: &max}},
		Associations: []client.Association{
			{Account: "root", User: ""},
			{Account: "ops", ParentAccount: "root", QoS: []string{"normal", "debug"}},
			{Account: "ops", User: "zed", QoS: []string{"normal"}},
			{Account: "none", User: "zed", IsDefault: &yes},
			{Account: "a1", ParentAccount: "root"},
		}}, Conflicts: []drift.Conflict{allocQoSConflict}}
	rec := serve(t, newCore(&calls{}), fakeSnapshots{snap: snap}, &fakeSlurm{}, "GET", "/connectors/slurm/clusters/c1", "", models.ClustersRead)
	var got ClusterView
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if len(got.Partitions) != 1 || got.Partitions[0].Name != "batch" || len(got.Conflicts) != 1 || got.Conflicts[0].Key != allocQoSConflict.Key {
		t.Fatalf("view = %+v", got)
	}
	names := []string{}
	for _, a := range got.SystemAccounts {
		names = append(names, a.Name)
	}
	if strings.Join(names, ",") != "none,ops,root" {
		t.Fatalf("accounts = %v", names)
	}
	ops := got.SystemAccounts[1]
	if ops.Parent != "root" || strings.Join(ops.QoS, ",") != "normal,debug" || len(ops.Users) != 1 ||
		ops.Users[0].Name != "zed" || ops.Users[0].DefaultAccount != "none" {
		t.Fatalf("ops = %+v", ops)
	}

	snaps := fakeSnapshots{err: fmt.Errorf("live state: %w: connection refused", service.ErrUnavailable)}
	rec = serve(t, newCore(&calls{}), snaps, &fakeSlurm{}, "GET", "/connectors/slurm/clusters/c1", "", models.ClustersRead)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("unavailable: %d %s", rec.Code, rec.Body)
	}
}

func TestResolve_StaleSlurmValue409(t *testing.T) {
	stale := seen(models.ComputeClusterQoS{Name: "gpu", Priority: 1})
	rec := resolve(t, newCore(&calls{}), snapshotOf(qosConflict), `{"key":"qos:gpu","choice":"slurm","seen_slurm":`+stale+`}`, models.ClustersWrite)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
}

func TestResolve_PrivilegeByKind(t *testing.T) {
	body := `{"key":"account:a1:qos","choice":"custos","seen_slurm":` + seen(allocQoSConflict.Slurm) + `}`
	core := newCore(&calls{})
	if rec := resolve(t, core, snapshotOf(allocQoSConflict), body, models.ClustersWrite); rec.Code != http.StatusForbidden {
		t.Fatalf("allocation conflict with clusters:write: %d", rec.Code)
	}
	qosBody := `{"key":"qos:gpu","choice":"custos","seen_slurm":` + seen(qosConflict.Slurm) + `}`
	if rec := resolve(t, core, snapshotOf(qosConflict), qosBody, models.AllocationsWrite); rec.Code != http.StatusForbidden {
		t.Fatalf("qos conflict with allocations:write: %d", rec.Code)
	}
	if rec := resolve(t, core, snapshotOf(allocQoSConflict), body, models.AllocationsWrite); rec.Code != http.StatusNoContent {
		t.Fatalf("allocation conflict with allocations:write: %d %s", rec.Code, rec.Body)
	}
	if rec := resolve(t, core, snapshotOf(qosConflict), qosBody, models.ClustersWrite); rec.Code != http.StatusNoContent {
		t.Fatalf("qos conflict with clusters:write: %d %s", rec.Code, rec.Body)
	}
	if rec := resolve(t, core, snapshotOf(allocQoSConflict), body, models.ClustersRead); rec.Code != http.StatusForbidden {
		t.Fatalf("read-only caller: %d", rec.Code)
	}
}

func TestResolve_Custos_MarksPending(t *testing.T) {
	cases := []struct {
		name, body, want string
		conflict         drift.Conflict
		privilege        models.PrivilegeKey
	}{
		{"allocation", `{"key":"account:a1:qos","choice":"custos","seen_slurm":` + seen(allocQoSConflict.Slurm) + `}`, "mark-compute_allocations:al1", allocQoSConflict, models.AllocationsWrite},
		{"member", `{"key":"member:a1:u1:account","choice":"custos","seen_slurm":8}`, "mark-compute_allocation_memberships:m1", custosOnlyConflict, models.AllocationsWrite},
		{"qos", `{"key":"qos:gpu","choice":"custos","seen_slurm":` + seen(qosConflict.Slurm) + `}`, "mark-compute_cluster_qos:q-normal", drift.Conflict{Key: "qos:normal", Kind: drift.KindQoSDefinition, Name: "normal", Slurm: qosConflict.Slurm, Choices: qosConflict.Choices}, models.ClustersWrite},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &calls{}
			body := tc.body
			if tc.name == "qos" {
				body = `{"key":"qos:normal","choice":"custos","seen_slurm":` + seen(qosConflict.Slurm) + `}`
			}
			rec := resolve(t, newCore(c), snapshotOf(tc.conflict), body, tc.privilege)
			if rec.Code != http.StatusNoContent || strings.Join(c.list, "|") != tc.want {
				t.Fatalf("%d calls = %v, want %s", rec.Code, c.list, tc.want)
			}
		})
	}
}

// A SLURM-only QoS is deleted only when no system account uses it.
func TestResolve_Custos_SlurmOnlyQoSDeletesFromSlurm(t *testing.T) {
	only := drift.Conflict{Key: "qos:gpu", Kind: drift.KindQoSDefinition, Name: "gpu", Slurm: qosConflict.Slurm, Choices: []drift.Choice{drift.ChoiceCustos, drift.ChoiceSlurm}}
	body := `{"key":"qos:gpu","choice":"custos","seen_slurm":` + seen(only.Slurm) + `}`
	serveOnly := func(live ...client.Association) (*httptest.ResponseRecorder, *fakeSlurm) {
		slurm := &fakeSlurm{associations: live}
		return serve(t, newCore(&calls{}), fakeSnapshots{snap: snapshotOf(only)}, slurm, "POST", "/connectors/slurm/clusters/c1/conflicts/resolve",
			body, models.ClustersRead, models.ClustersWrite), slurm
	}

	usedByAllocation := client.Association{Account: "a1", Cluster: "alpha", QoS: []string{"gpu"}}
	rec, slurm := serveOnly(usedByAllocation)
	if rec.Code != http.StatusNoContent || strings.Join(slurm.deletedQoS, ",") != "gpu" {
		t.Fatalf("%d deleted = %v", rec.Code, slurm.deletedQoS)
	}

	if rec, slurm = serveOnly(client.Association{Account: "ops", Cluster: "alpha", QoS: []string{"gpu"}}); rec.Code != http.StatusConflict || len(slurm.deletedQoS) != 0 {
		t.Fatalf("system-used: %d deleted = %v", rec.Code, slurm.deletedQoS)
	}
}
