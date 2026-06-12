package subscribers

import (
	"context"
	"fmt"
	operations "github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
	"os"
	"testing"
)

func TestSubscribeToComputeAllocationCreation(t *testing.T) {
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

	associationSubscriber := NewAssociationSubscriber(client, nil, mockCoreService)

	computeAccount := models.ComputeAllocation{
		ID:               "test_compute_account_id",
		Name:             "test_compute_account",
		ProjectID:        "test_project_id",
		ComputeClusterID: "test_cluster_id",
	}

	associationSubscriber.SubscribeToComputeAllocationCreation(computeAccount)

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

func TestSubscribeToComputeAllocationCreationWrongCluster(t *testing.T) {
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

	associationSubscriber := NewAssociationSubscriber(client, nil, mockCoreService)

	computeAccount := models.ComputeAllocation{
		ID:               "test_compute_account_id",
		Name:             "test_compute_account",
		ProjectID:        "test_project_id",
		ComputeClusterID: "test_cluster_id",
	}

	associationSubscriber.SubscribeToComputeAllocationCreation(computeAccount)

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

	associationSubscriber := NewAssociationSubscriber(client, nil, mockCoreService)

	computeAllocation := models.ComputeAllocation{
		ID:               computeAllocationID,
		Name:             computeAllocationName,
		ProjectID:        projectID,
		ComputeClusterID: clusterID,
	}

	associationSubscriber.SubscribeToComputeAllocationCreation(computeAllocation)

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

	associationSubscriber.SubscribeToComputeAllocationResourceMappingCreation(resourceMapping)

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

func TestSubscribeToComputeAllocationResourceMappingCreation(t *testing.T) {
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

	err := createAllocationMapping(client, mockCoreService, clusterID, clusterName,
		computeAllocationID, computeAllocationName,
		projectID, allocationMappingId, allocationResourceId,
		partitionName)
	if err != nil {
		t.Fatalf("Failed to create allocation mapping: %v", err)
	}

}

func TestSubscribeToComputeAllocationResourceMappingWrongClusterCreation(t *testing.T) {

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

func TestSubscribeToComputeAllocationResourceMappingWrongResourceTypeCreation(t *testing.T) {

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
