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
	"slices"
	"testing"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

var deletedAccount = models.ComputeClusterUser{ID: "csu-1", ComputeClusterID: "cluster-1", UserID: "user-1", LocalUsername: "testuser"}

func TestAccountDeletionRemovesTheSlurmUser(t *testing.T) {
	slurm := &fakeSlurmClient{existing: []client.Association{
		{Cluster: "testcluster", Account: noAllocation, User: "testuser"},
		{Cluster: "testcluster", Account: "test-alloc", User: "testuser"},
	}}
	if err := NewAssociationSubscriber(slurm, nil, coreMock(mockOpts{}), 0, 0).
		handleComputeClusterUserDeleted(context.Background(), deletedAccount); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !slices.Equal(slurm.deletedUsers, []string{"testuser"}) || len(slurm.allDeletes()) != 0 {
		t.Fatalf("deleted users %v, association deletes %+v; want the user only", slurm.deletedUsers, slurm.allDeletes())
	}
}

func TestAccountDeletionKeepsAUserOnAnotherCluster(t *testing.T) {
	slurm := &fakeSlurmClient{existing: []client.Association{
		{Cluster: "testcluster", Account: noAllocation, User: "testuser"},
		{Cluster: "testcluster", Account: "test-alloc", User: "testuser"},
		{Cluster: "othercluster", Account: noAllocation, User: "testuser"},
	}}
	if err := NewAssociationSubscriber(slurm, nil, coreMock(mockOpts{}), 0, 0).
		handleComputeClusterUserDeleted(context.Background(), deletedAccount); err != nil {
		t.Fatalf("delete: %v", err)
	}
	want := []client.AssocFilter{
		{Cluster: "testcluster", Account: "test-alloc", User: "testuser"},
		{Cluster: "testcluster", Account: noAllocation, User: "testuser"},
	}
	if len(slurm.deletedUsers) != 0 || !slices.Equal(slurm.allDeletes(), want) {
		t.Fatalf("deleted users %v, association deletes %+v; want %+v", slurm.deletedUsers, slurm.allDeletes(), want)
	}
}
