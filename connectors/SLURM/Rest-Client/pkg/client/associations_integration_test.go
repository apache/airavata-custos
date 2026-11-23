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

package client

import (
	"os"
	"slices"
	"testing"
)

func TestCreateInvalidAssocation_Integration(t *testing.T) {
	if !IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for association creation because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := New(apiUrl, user, token, apiVersion)

	association := Association{
		Account:   "test_account",
		User:      "test_user",
		Cluster:   "test_cluster",
		Partition: "test_partition",
	}

	err := client.UpsertAssociation(association)
	if err != nil {
		expected := "association not found after upsert"
		if err.Error() != expected {
			t.Fatalf("Unexpected error: got %v, want %v", err.Error(), expected)
		}
	} else {
		t.Fatal("Expected error when creating association with non-existent account, user, cluster, and partition, but got nil")
	}
}

func TestCreateValidAssociation_Integration(t *testing.T) {
	if !IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for association creation because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := New(apiUrl, user, token, apiVersion)

	CrearteAndValidateAccount(t, client)

	association := Association{
		Account:   "test_account",
		User:      "test_user",
		Cluster:   "artisan",
		Partition: "compute",
	}

	err := client.UpsertAssociation(association)
	if err != nil {
		t.Fatalf("Failed to create association: %v", err)
	}

	assocs, err := client.ListAssociations(AssocFilter{
		Account:   "test_account",
		User:      "test_user",
		Cluster:   "artisan",
		Partition: "compute",
	})
	if err != nil {
		t.Fatalf("Failed to list associations after creation: %v", err)
	}
	if len(assocs) != 1 {
		t.Fatalf("Expected exactly 1 association after creation, but found %d: %+v", len(assocs), assocs)
	}
	if assocs[0].Account != "test_account" || assocs[0].User != "test_user" || assocs[0].Cluster != "artisan" || assocs[0].Partition != "compute" {
		t.Fatalf("Association fields do not match expected values: %+v", assocs[0])
	}

	defer client.DeleteAccount("test_account") // clean up after test

}

func TestDeleteAssociation_Integration(t *testing.T) {
	if !IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for association deletion because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := New(apiUrl, user, token, apiVersion)

	CrearteAndValidateAccount(t, client)
	defer client.DeleteAccount("test_account") // clean up after test

	association := Association{
		Account:   "test_account",
		User:      "test_user",
		Cluster:   "artisan",
		Partition: "compute",
	}

	err := client.UpsertAssociation(association)
	if err != nil {
		t.Fatalf("Failed to create association: %v", err)
	}

	filter := AssocFilter{
		Account:   "test_account",
		User:      "test_user",
		Cluster:   "artisan",
		Partition: "compute",
	}
	err = client.DeleteAssociation(filter)
	if err != nil {
		t.Fatalf("Failed to delete association: %v", err)
	}

	assocs, err := client.ListAssociations(AssocFilter{
		Account:   "test_account",
		User:      "test_user",
		Cluster:   "artisan",
		Partition: "compute",
	})
	if err != nil {
		t.Fatalf("Failed to list associations after deletion: %v", err)
	}
	if len(assocs) != 0 {
		t.Fatalf("Expected no associations after deletion, but found: %+v", assocs)
	}
}

// A Custos write carries no QoS and carries over TRES limits set outside Custos, so both survive it.
func TestUpsertAssociationKeepsQoSAndLimitsSetOutsideCustos_Integration(t *testing.T) {
	if !IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test because local SLURM config is not available")
	}
	client := New(os.Getenv("TEST_SLURM_API"), os.Getenv("TEST_SLURM_USER"), os.Getenv("TEST_SLURM_TOKEN"), os.Getenv("TEST_SLURM_API_VERSION"))
	qosPath := "/slurmdb/v0.0." + client.apiVersion + "/qos"
	if _, err := client.do("POST", qosPath, map[string]any{"qos": []map[string]any{{"name": "outside_qos"}}}, nil); err != nil {
		t.Fatalf("create QoS: %v", err)
	}
	defer client.do("DELETE", qosPath+"/outside_qos", nil, nil)
	CrearteAndValidateAccount(t, client)
	defer client.DeleteAccount("test_account")

	key := Association{Account: "test_account", Cluster: "artisan", User: "test_user"}
	outside := key
	outside.QoS = []string{"normal", "outside_qos"}
	outside.Limits.GrpTRES = []TRES{{Type: "mem", Count: 1024}}
	if err := client.UpsertAssociation(outside); err != nil {
		t.Fatalf("set QoS and mem limit outside Custos: %v", err)
	}

	readCurrent := func() Association {
		assocs, err := client.ListAssociations(AssocFilter{Account: key.Account, Cluster: key.Cluster, User: key.User})
		if err != nil || len(assocs) != 1 {
			t.Fatalf("list: %v %+v", err, assocs)
		}
		return assocs[0]
	}
	custos := key
	custos.Limits.GrpTRES = WithTRESSetOutside(readCurrent().Limits.GrpTRES, []TRES{{Type: "cpu", Count: 4}})
	if err := client.UpsertAssociation(custos); err != nil {
		t.Fatalf("Custos write: %v", err)
	}

	current := readCurrent()
	if len(current.QoS) != 2 || !slices.Contains(current.QoS, "outside_qos") {
		t.Errorf("expected QoS set outside Custos to be kept, got %v", current.QoS)
	}
	if NeedsTRESUpdate(current.Limits.GrpTRES, []TRES{{Type: "mem", Count: 1024}, {Type: "cpu", Count: 4}}) {
		t.Errorf("expected mem limit set outside Custos plus Custos's cpu limit, got %+v", current.Limits.GrpTRES)
	}

	custos.Limits.GrpTRES = WithTRESSetOutside(current.Limits.GrpTRES, []TRES{{Type: "cpu", Count: NoLimit}})
	if err := client.UpsertAssociation(custos); err != nil {
		t.Fatalf("Custos clears cpu: %v", err)
	}
	if got := readCurrent().Limits.GrpTRES; len(got) != 1 || got[0] != (TRES{Type: "mem", Count: 1024}) {
		t.Errorf("expected only the mem limit set outside Custos after clearing cpu, got %+v", got)
	}
}
