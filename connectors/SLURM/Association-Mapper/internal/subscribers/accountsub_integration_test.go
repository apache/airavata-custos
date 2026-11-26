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
	"fmt"
	"os"
	"testing"
	"time"

	operations "github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

func TestHandleComputeAllocationCreated(t *testing.T) {
	if !operations.IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for compute allocation creation subscription because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := operations.New(apiUrl, user, token, apiVersion)

	auditEvents := make([]*models.AuditEvent, 0)

	mockCoreService := &service.CoreServiceMock{
		GetComputeClusterFunc: func(ctx context.Context, clusterID string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: clusterID, Name: "artisan"}, nil
		},
		GetProjectFunc: func(ctx context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Title: "sample_project", Status: "ACTIVE"}, nil
		},
		GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, OrganizationID: "test_org_id"}, nil
		},
		GetOrganizationFunc: func(ctx context.Context, id string) (*models.Organization, error) {
			return &models.Organization{ID: id, Name: "test_org"}, nil
		},
		CreateAuditEventFunc: func(ctx context.Context, event *models.AuditEvent) (*models.AuditEvent, error) {
			t.Logf("Audit event created: %+v\n", event)
			auditEvents = append(auditEvents, event)
			return event, nil
		},
	}

	associationSubscriber := NewAssociationSubscriber(client, nil, mockCoreService, 0, 0)

	computeAccount := models.ComputeAllocation{
		ID:               "test_compute_account_id",
		Name:             "test_compute_account",
		ProjectID:        "test_project_id",
		ComputeClusterID: "test_cluster_id",
	}
	stubSyncCore(mockCoreService, computeAccount, nil, nil)

	allocationCreated.run(context.Background(), associationSubscriber, computeAccount)

	if len(auditEvents) == 0 {
		t.Errorf("Expected at least one audit event, but got none")
	}

	if len(auditEvents) > 0 {
		lastEvent := auditEvents[len(auditEvents)-1]
		if lastEvent.EventType != "ComputeAllocationCreationSucceeded" {
			t.Errorf("Expected audit event type 'ComputeAllocationCreationSucceeded', got '%s'", lastEvent.EventType)
		}
		if lastEvent.EntityID != computeAccount.ID {
			t.Errorf("Expected audit event EntityID '%s', got '%s'", computeAccount.ID, lastEvent.EntityID)
		}
	}

	client.DeleteAccount(computeAccount.Name) // clean up after test

}

func TestHandleComputeAllocationCreatedWrongCluster(t *testing.T) {
	if !operations.IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for compute allocation creation subscription because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := operations.New(apiUrl, user, token, apiVersion)

	auditEvents := make([]*models.AuditEvent, 0)

	mockCoreService := &service.CoreServiceMock{
		GetComputeClusterFunc: func(ctx context.Context, clusterID string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: clusterID, Name: "wrong_cluster_name"}, nil
		},
		GetProjectFunc: func(ctx context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Title: "sample_project", Status: "ACTIVE"}, nil
		},
		GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, OrganizationID: "test_org_id"}, nil
		},
		GetOrganizationFunc: func(ctx context.Context, id string) (*models.Organization, error) {
			return &models.Organization{ID: id, Name: "test_org"}, nil
		},
		CreateAuditEventFunc: func(ctx context.Context, event *models.AuditEvent) (*models.AuditEvent, error) {
			t.Logf("Audit event created: %+v\n", event)
			auditEvents = append(auditEvents, event)
			return event, nil
		},
	}

	associationSubscriber := NewAssociationSubscriber(client, nil, mockCoreService, 0, 0)

	computeAccount := models.ComputeAllocation{
		ID:               "test_compute_account_id",
		Name:             "test_compute_account",
		ProjectID:        "test_project_id",
		ComputeClusterID: "test_cluster_id",
	}
	stubSyncCore(mockCoreService, computeAccount, nil, nil)

	allocationCreated.run(context.Background(), associationSubscriber, computeAccount)

	if len(auditEvents) == 0 {
		t.Errorf("Expected at least one audit event, but got none")
	}

	if len(auditEvents) > 0 {
		lastEvent := auditEvents[len(auditEvents)-1]
		if lastEvent.EventType != "ComputeAllocationCreationFailed" {
			t.Errorf("Expected audit event type 'ComputeAllocationCreationFailed', got '%s'", lastEvent.EventType)
		}
		if lastEvent.EntityID != computeAccount.ID {
			t.Errorf("Expected audit event EntityID '%s', got '%s'", computeAccount.ID, lastEvent.EntityID)
		}
	}

	client.DeleteAccount(computeAccount.Name) // clean up after test
}

func createAllocationMapping(client *operations.Client, mockCoreService *service.CoreServiceMock, clusterID string, clusterName string,
	computeAllocationID string, computeAllocationName string,
	projectID string, allocationMappingId string,
	allocationResourceId string, partitionName string) error {

	associationSubscriber := NewAssociationSubscriber(client, nil, mockCoreService, 0, 0)

	computeAllocation := models.ComputeAllocation{
		ID:               computeAllocationID,
		Name:             computeAllocationName,
		ProjectID:        projectID,
		ComputeClusterID: clusterID,
	}

	allocationCreated.run(context.Background(), associationSubscriber, computeAllocation)

	allAuditEvents, err := mockCoreService.ListAllAuditEvents(context.Background())

	if err != nil {
		return fmt.Errorf("Failed to list all audit events: %v", err)
	}

	if len(allAuditEvents) == 0 {
		return fmt.Errorf("Expected at least one audit event after resource mapping creation, but got none")
	}

	if len(allAuditEvents) > 0 {
		lastEvent := allAuditEvents[len(allAuditEvents)-1]
		if lastEvent.EventType != "ComputeAllocationCreationSucceeded" {
			return fmt.Errorf("Expected audit event type 'ComputeAllocationCreationSucceeded', got '%s'", lastEvent.EventType)
		}
		if lastEvent.EntityID != computeAllocation.ID {
			return fmt.Errorf("Expected audit event EntityID '%s', got '%s'", computeAllocation.ID, lastEvent.EntityID)
		}
	}

	resourceMapping := models.ComputeAllocationResourceMapping{
		ID:                          allocationMappingId,
		ComputeAllocationID:         computeAllocation.ID,
		ComputeAllocationResourceID: allocationResourceId,
		ResourceAmount:              10,
		ResourceTime:                100,
	}
	mockCoreService.ListResourceMappingsForAllocationFunc = func(context.Context, string) ([]models.ComputeAllocationResourceMapping, error) {
		return []models.ComputeAllocationResourceMapping{resourceMapping}, nil
	}

	mappingCreated.run(context.Background(), associationSubscriber, resourceMapping)

	allAuditEvents, err = mockCoreService.ListAllAuditEvents(context.Background())

	if err != nil {
		return fmt.Errorf("Failed to list all audit events: %v", err)
	}

	if len(allAuditEvents) == 0 {
		return fmt.Errorf("Expected at least one audit event after resource mapping creation, but got none")
	}
	if len(allAuditEvents) > 0 {
		lastEvent := allAuditEvents[len(allAuditEvents)-1]
		if lastEvent.EventType != "ComputeAllocationResourceMappingCreationSucceeded" {
			return fmt.Errorf("Expected audit event type 'ComputeAllocationResourceMappingCreationSucceeded', got '%s'", lastEvent.EventType)
		}
		if lastEvent.EntityID != resourceMapping.ID {
			return fmt.Errorf("Expected audit event EntityID '%s', got '%s'", resourceMapping.ID, lastEvent.EntityID)
		}
	}

	return nil
}

func TestHandleComputeAllocationResourceMappingCreated(t *testing.T) {
	if !operations.IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for compute allocation creation subscription because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := operations.New(apiUrl, user, token, apiVersion)

	clusterID := "cluster-001"
	clusterName := "artisan"
	projectID := "project-001"
	computeAllocationID := "compute-allocation-001"
	computeAllocationName := "md-allocation"
	partitionName := "compute"
	allocationResourceId := "allocation-resource-001"
	allocationMappingId := "allocation-mapping-001"

	defer client.DeleteAccount(computeAllocationName)                                      // clean up after test
	defer client.DeleteAssociation(operations.AssocFilter{Account: computeAllocationName}) // clean up after test

	auditEvents := make([]*models.AuditEvent, 0)

	mockCoreService := &service.CoreServiceMock{
		GetComputeClusterFunc: func(ctx context.Context, clusterID string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: clusterID, Name: clusterName}, nil
		},
		GetProjectFunc: func(ctx context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Title: "PROJ001", Status: "ACTIVE"}, nil
		},
		GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, OrganizationID: "GATECH"}, nil
		},
		GetOrganizationFunc: func(ctx context.Context, id string) (*models.Organization, error) {
			return &models.Organization{ID: id, Name: "Georgia Institute of Technology"}, nil
		},
		CreateAuditEventFunc: func(ctx context.Context, event *models.AuditEvent) (*models.AuditEvent, error) {
			auditEvents = append(auditEvents, event)
			return event, nil
		},
		GetComputeAllocationFunc: func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
			return &models.ComputeAllocation{ID: id, Name: computeAllocationName, ComputeClusterID: clusterID}, nil
		},
		GetComputeAllocationResourceFunc: func(ctx context.Context, id string) (*models.ComputeAllocationResource, error) {
			return &models.ComputeAllocationResource{ID: id, Name: partitionName, ResourceType: "cpu", ResourceAmount: 1000}, nil
		},
		ListAllAuditEventsFunc: func(ctx context.Context) ([]*models.AuditEvent, error) {
			return auditEvents, nil
		},
	}

	stubSyncCore(mockCoreService, models.ComputeAllocation{ID: computeAllocationID, Name: computeAllocationName, ComputeClusterID: clusterID}, nil, nil)

	err := createAllocationMapping(client, mockCoreService, clusterID, clusterName,
		computeAllocationID, computeAllocationName,
		projectID, allocationMappingId, allocationResourceId,
		partitionName)
	if err != nil {
		t.Fatalf("Failed to create allocation mapping: %v", err)
	}

}

func TestHandleComputeAllocationResourceMappingWrongClusterCreated(t *testing.T) {

	// Pass the wrong cluster name in the mock to simulate the case where the cluster
	// associated with the compute allocation does not match any cluster in SLURM,
	// which should result in a failure to create the association and an audit event
	// being created for the failure

	if !operations.IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for compute allocation creation subscription because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := operations.New(apiUrl, user, token, apiVersion)

	clusterID := "cluster-001"
	clusterName := "wrong_cluster_name" // this is intentionally wrong to simulate the error case
	projectID := "project-001"
	computeAllocationID := "compute-allocation-001"
	computeAllocationName := "md-allocation"
	partitionName := "compute"
	allocationResourceId := "allocation-resource-001"
	allocationMappingId := "allocation-mapping-001"

	defer client.DeleteAccount(computeAllocationName)                                      // clean up after test
	defer client.DeleteAssociation(operations.AssocFilter{Account: computeAllocationName}) // clean up after test

	auditEvents := make([]*models.AuditEvent, 0)

	mockCoreService := &service.CoreServiceMock{
		GetComputeClusterFunc: func(ctx context.Context, clusterID string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: clusterID, Name: clusterName}, nil
		},
		GetProjectFunc: func(ctx context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Title: "PROJ001", Status: "ACTIVE"}, nil
		},
		GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, OrganizationID: "GATECH"}, nil
		},
		GetOrganizationFunc: func(ctx context.Context, id string) (*models.Organization, error) {
			return &models.Organization{ID: id, Name: "Georgia Institute of Technology"}, nil
		},
		CreateAuditEventFunc: func(ctx context.Context, event *models.AuditEvent) (*models.AuditEvent, error) {
			auditEvents = append(auditEvents, event)
			return event, nil
		},
		GetComputeAllocationFunc: func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
			return &models.ComputeAllocation{ID: id, Name: computeAllocationName, ComputeClusterID: clusterID}, nil
		},
		GetComputeAllocationResourceFunc: func(ctx context.Context, id string) (*models.ComputeAllocationResource, error) {
			return &models.ComputeAllocationResource{ID: id, Name: partitionName, ResourceType: "cpu", ResourceAmount: 1000}, nil
		},
		ListAllAuditEventsFunc: func(ctx context.Context) ([]*models.AuditEvent, error) {
			return auditEvents, nil
		},
	}

	stubSyncCore(mockCoreService, models.ComputeAllocation{ID: computeAllocationID, Name: computeAllocationName, ComputeClusterID: clusterID}, nil, nil)

	err := createAllocationMapping(client, mockCoreService, clusterID, clusterName,
		computeAllocationID, computeAllocationName,
		projectID, allocationMappingId, allocationResourceId,
		partitionName)
	if err != nil {
		expectedErrorMsg := "Expected audit event type 'ComputeAllocationCreationSucceeded', got 'ComputeAllocationCreationFailed'"
		if err.Error() != expectedErrorMsg {
			t.Fatalf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
		}
	} else {
		t.Fatal("Expected error due to wrong cluster name, but got nil")
	}
}

func TestHandleComputeAllocationResourceMappingWrongResourceTypeCreated(t *testing.T) {

	// Passing the wrong resource type in the mock to simulate the case where the
	// resource type of the allocation resource is not compatible with SLURM,
	// which should result in a failure to create the association and an audit event
	// being created for the failure

	if !operations.IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for compute allocation creation subscription because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := operations.New(apiUrl, user, token, apiVersion)

	clusterID := "cluster-001"
	clusterName := "artisan"
	projectID := "project-001"
	computeAllocationID := "compute-allocation-001"
	computeAllocationName := "md-allocation"
	partitionName := "compute"
	allocationResourceId := "allocation-resource-001"
	allocationMappingId := "allocation-mapping-001"

	defer client.DeleteAccount(computeAllocationName)                                      // clean up after test
	defer client.DeleteAssociation(operations.AssocFilter{Account: computeAllocationName}) // clean up after test

	auditEvents := make([]*models.AuditEvent, 0)

	mockCoreService := &service.CoreServiceMock{
		GetComputeClusterFunc: func(ctx context.Context, clusterID string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: clusterID, Name: clusterName}, nil
		},
		GetProjectFunc: func(ctx context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Title: "PROJ001", Status: "ACTIVE"}, nil
		},
		GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, OrganizationID: "GATECH"}, nil
		},
		GetOrganizationFunc: func(ctx context.Context, id string) (*models.Organization, error) {
			return &models.Organization{ID: id, Name: "Georgia Institute of Technology"}, nil
		},
		CreateAuditEventFunc: func(ctx context.Context, event *models.AuditEvent) (*models.AuditEvent, error) {
			auditEvents = append(auditEvents, event)
			return event, nil
		},
		GetComputeAllocationFunc: func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
			return &models.ComputeAllocation{ID: id, Name: computeAllocationName, ComputeClusterID: clusterID}, nil
		},
		GetComputeAllocationResourceFunc: func(ctx context.Context, id string) (*models.ComputeAllocationResource, error) {
			// Return a resource type that is not compatible with SLURM to simulate the error case where the
			//resource mapping creation fails due to invalid resource type. Test cluster does not have gpu
			return &models.ComputeAllocationResource{ID: id, Name: partitionName, ResourceType: "gpu", ResourceAmount: 1000}, nil
		},
		ListAllAuditEventsFunc: func(ctx context.Context) ([]*models.AuditEvent, error) {
			return auditEvents, nil
		},
	}

	stubSyncCore(mockCoreService, models.ComputeAllocation{ID: computeAllocationID, Name: computeAllocationName, ComputeClusterID: clusterID}, nil, nil)

	err := createAllocationMapping(client, mockCoreService, clusterID, clusterName,
		computeAllocationID, computeAllocationName,
		projectID, allocationMappingId, allocationResourceId,
		partitionName)
	if err != nil {
		expectedErrorMsg := "Expected audit event type 'ComputeAllocationResourceMappingCreationSucceeded', got 'ComputeAllocationResourceMappingCreationFailed'"
		if err.Error() != expectedErrorMsg {
			t.Fatalf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
		}
	} else {
		t.Fatal("Expected error due to wrong resource type, but got nil")
	}
}

// stubSyncCore answers the core calls a cluster sync makes: one active allocation with the given mappings and members.
func stubSyncCore(m *service.CoreServiceMock, allocation models.ComputeAllocation, mappings []models.ComputeAllocationResourceMapping, members []store.MembershipWithUser) {
	allocation.Status = models.ACTIVE
	m.GetComputeAllocationFunc = func(context.Context, string) (*models.ComputeAllocation, error) { return &allocation, nil }
	m.ListComputeClustersFunc = func(context.Context) ([]models.ComputeCluster, error) {
		return []models.ComputeCluster{{ID: allocation.ComputeClusterID}}, nil
	}
	m.ListComputeClusterQoSFunc = func(context.Context, string, string) ([]models.ComputeClusterQoS, error) { return nil, nil }
	m.ListComputeAllocationsByClusterFunc = func(context.Context, string) ([]models.ComputeAllocation, error) {
		return []models.ComputeAllocation{allocation}, nil
	}
	m.ListResourceMappingsForAllocationFunc = func(context.Context, string) ([]models.ComputeAllocationResourceMapping, error) {
		return mappings, nil
	}
	m.ListMembersForAllocationFunc = func(context.Context, string) ([]store.MembershipWithUser, error) { return members, nil }
	m.ListOverridesForMembershipFunc = func(context.Context, string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
		return nil, nil
	}
	m.SetSlurmPendingFunc = func(context.Context, service.SlurmPendingKind, string, *time.Time) error { return nil }
}
