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

// Pins slurmrestd v0.0.45 behavior that Custos-managed QoS and associations depend on.
package client

import (
	"os"
	"testing"
)

const pinCluster = "artisan"

func integrationClient(t *testing.T) *Client {
	t.Helper()
	if !IsLocalSlurmConfigAvailable() {
		t.Skip("local SLURM config is not available")
	}
	return New(os.Getenv("TEST_SLURM_API"), os.Getenv("TEST_SLURM_USER"), os.Getenv("TEST_SLURM_TOKEN"), os.Getenv("TEST_SLURM_API_VERSION"))
}

func slurmPath(c *Client, rest string) string { return "/slurm/v0.0." + c.apiVersion + rest }

func createQoS(t *testing.T, c *Client, name string, priority int64) {
	t.Helper()
	if err := c.PostQoS(QoS{Name: name, Priority: priority}); err != nil {
		t.Fatalf("create qos %s: %v", name, err)
	}
	t.Cleanup(func() { _ = c.DeleteQoS(name) })
}

func findQoS(t *testing.T, c *Client, name string) (QoS, bool) {
	t.Helper()
	all, err := c.ListQoS()
	if err != nil {
		t.Fatalf("list qos: %v", err)
	}
	for _, q := range all {
		if q.Name == name {
			return q, true
		}
	}
	return QoS{}, false
}

// createAccountWithMember adds an account with one member that has no QoS or limits.
func createAccountWithMember(t *testing.T, c *Client, account, member string) {
	t.Helper()
	if err := c.CreateAccount(Account{Name: account, Description: "pin", Organization: "pin"}, pinCluster); err != nil {
		t.Fatalf("create account %s: %v", account, err)
	}
	t.Cleanup(func() {
		_ = c.DeleteAccount(account)
		_ = c.DeleteUser(member)
	})
	if err := c.UpsertAssociation(Association{Account: account, User: member, Cluster: pinCluster}); err != nil {
		t.Fatalf("create member %s: %v", member, err)
	}
}

func getAssociation(t *testing.T, c *Client, account, user string) Association {
	t.Helper()
	assocs, err := c.ListAssociations(AssocFilter{Account: account, User: user, Cluster: pinCluster})
	if err != nil {
		t.Fatalf("list associations: %v", err)
	}
	for _, a := range assocs {
		if a.User == user {
			return a
		}
	}
	t.Fatalf("association %s/%q not found in %+v", account, user, assocs)
	return Association{}
}

func TestQoS_CRUD(t *testing.T) {
	c := integrationClient(t)
	name := "pin_crud_qos"
	createQoS(t, c, name, 5)
	if q, ok := findQoS(t, c, name); !ok || q.Priority != 5 {
		t.Fatalf("after create = %+v, %v; want priority 5", q, ok)
	}

	createQoS(t, c, name, 9) // POST on an existing name updates it
	if q, _ := findQoS(t, c, name); q.Priority != 9 {
		t.Fatalf("priority after update = %d, want 9", q.Priority)
	}

	if err := c.DeleteQoS(name); err != nil {
		t.Fatalf("delete qos: %v", err)
	}
	if q, ok := findQoS(t, c, name); ok {
		t.Fatalf("qos after delete = %+v, want none", q)
	}
}

// An upsert without a limit clears it: scalars are sent as infinite, an omitted TRES list clears.
func TestUpsertAssociation_ClearsRemovedLimits(t *testing.T) {
	c := integrationClient(t)
	createAccountWithMember(t, c, "pin_clear_acct", "pin_clear_user")

	jobs, wall := int64(10), int64(60)
	member := Association{Account: "pin_clear_acct", User: "pin_clear_user", Cluster: pinCluster}
	member.Limits = AssocLimits{GrpJobs: &jobs, MaxWallPM: &wall,
		GrpTRES: []TRES{{Type: "cpu", Count: 4}}, GrpTRESMins: []TRES{{Type: "cpu", Count: 600}}}
	if err := c.UpsertAssociation(member); err != nil {
		t.Fatalf("upsert limits: %v", err)
	}
	if got := getAssociation(t, c, "pin_clear_acct", "pin_clear_user").Limits; got.GrpJobs == nil || got.MaxWallPM == nil {
		t.Fatalf("limits not set: %+v", got)
	}

	member.Limits = AssocLimits{}
	if err := c.UpsertAssociation(member); err != nil {
		t.Fatalf("upsert without limits: %v", err)
	}
	if got := getAssociation(t, c, "pin_clear_acct", "pin_clear_user").Limits; got.GrpJobs != nil || got.MaxWallPM != nil ||
		len(got.GrpTRES) != 0 || len(got.GrpTRESMins) != 0 {
		t.Fatalf("limits not cleared: %+v", got)
	}
}

func partitionRequest(name string) map[string]any {
	return map[string]any{"partitions": []map[string]any{{
		"name":  name,
		"nodes": map[string]any{"configured": "c1"},
		"qos":   map[string]any{"allowed": "normal"},
	}}}
}

func createPartition(t *testing.T, c *Client, name string) {
	t.Helper()
	if _, err := c.do("POST", slurmPath(c, "/partitions/"), partitionRequest(name), nil); err != nil {
		t.Fatalf("create partition %s: %v", name, err)
	}
	t.Cleanup(func() { _, _ = c.do("DELETE", slurmPath(c, "/partition/"+name), nil, nil) })
}

func TestListPartitions_ReturnsLive(t *testing.T) {
	c := integrationClient(t)
	createPartition(t, c, "pin_list")
	if other, err := c.ListPartitions("not-" + pinCluster); err != nil || len(other) != 0 {
		t.Fatalf("another cluster's partitions = %+v, %v; want none", other, err)
	}
	all, err := c.ListPartitions(pinCluster)
	if err != nil {
		t.Fatalf("list partitions: %v", err)
	}
	for _, p := range all {
		if p.Name == "pin_list" {
			if p.Nodes != "c1" || p.State == "" {
				t.Fatalf("partition = %+v", p)
			}
			return
		}
	}
	t.Fatalf("pin_list not in %+v", all)
}
