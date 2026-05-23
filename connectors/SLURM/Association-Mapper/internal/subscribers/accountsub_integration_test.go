package subscribers

import (
	"os"
	"testing"

	"context"
	operations "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
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
