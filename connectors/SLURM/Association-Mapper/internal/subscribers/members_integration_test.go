package subscribers

import (
	"context"
	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
	"os"
	"testing"
)

func TestSubscribeToComputeAllocationMembershipCreation(t *testing.T) {
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
	computeAllocationID := "compute-allocation-001"
	computeAllocationName := "MD Allocation"
	partitionName := "compute"
	allocationResourceId := "allocation-resource-001"
	//allocationMappingId := "allocation-mapping-001"
	orgName := "Georgia Institute of Technology"
	localUsername := "testuser" // testuser, testuser2, and testuser3 are the 3 users in the test cluster

	client.CreateAccount(operations.Account{
		Name:         computeAllocationName,
		Description:  "Test account for integration testing",
		Organization: orgName,
	}, clusterName)

	client.UpsertAssociation(operations.Association{
		Account:   computeAllocationName,
		Cluster:   clusterName,
		Partition: partitionName,
		Limits: operations.AssocLimits{
			GrpTRES: []operations.TRES{
				{
					Type:  "cpu",
					Count: 100,
				},
			},
			GrpTRESMins: []operations.TRES{
				{
					Type:  "cpu",
					Count: 1000,
				},
			},
		},
	})

	defer client.DeleteAccount(computeAllocationName)

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
			return &models.Organization{ID: id, Name: orgName}, nil
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
		GetComputeClusterUserByPairFunc: func(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error) {
			return &models.ComputeClusterUser{ID: "csu-001", LocalUsername: localUsername}, nil
		},
		ListResourcesForAllocationFunc: func(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error) {
			return []models.ComputeAllocationResource{
				{ID: allocationResourceId, Name: partitionName, ResourceType: "cpu", ResourceAmount: 1000},
			}, nil
		},
	}

	// In the test cluster, there are 3 users: testuser, testuser2, and testuser3

	membership := &models.ComputeAllocationMembership{
		ID:                  "membership-001",
		UserID:              "user-001",
		ComputeAllocationID: computeAllocationID,
	}
	associationSubscriber := NewAssociationSubscriber(client, nil, mockCoreService)
	associationSubscriber.SubscribeToComputeAllocationMembershipCreation(*membership)

	if len(auditEvents) != 1 {
		t.Fatalf("Expected 1 audit event, but got %d", len(auditEvents))
	}

	if auditEvents[0].EventType != "ComputeAllocationMembershipCreationSucceeded" {
		t.Fatalf("Expected audit event type 'ComputeAllocationMembershipCreationSucceeded', but got '%s'", auditEvents[0].EventType)
	}
}

func TestSubscribeToComputeAllocationMembershipResourceOverrideCreation(t *testing.T) {
	if !operations.IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for compute allocation membership resource override creation subscription because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := operations.New(apiUrl, user, token, apiVersion)

	clusterID := "cluster-001"
	clusterName := "artisan"
	computeAllocationID := "compute-allocation-001"
	computeAllocationName := "MD Allocation"
	partitionName := "compute"
	allocationResourceId := "allocation-resource-001"
	//allocationMappingId := "allocation-mapping-001"
	orgName := "Georgia Institute of Technology"
	localUsername := "testuser" // testuser, testuser2, and testuser3 are the 3 users in the test cluster

	client.CreateAccount(operations.Account{
		Name:         computeAllocationName,
		Description:  "Test account for integration testing",
		Organization: orgName,
	}, clusterName)

	client.UpsertAssociation(operations.Association{
		Account:   computeAllocationName,
		Cluster:   clusterName,
		Partition: partitionName,
		Limits: operations.AssocLimits{
			GrpTRES: []operations.TRES{
				{
					Type:  "cpu",
					Count: 100,
				},
			},
			GrpTRESMins: []operations.TRES{
				{
					Type:  "cpu",
					Count: 1000,
				},
			},
		},
	})

	defer client.DeleteAccount(computeAllocationName)

	baseAssociation := operations.Association{
		Account:   computeAllocationName,
		Cluster:   clusterName,
		Partition: partitionName,
		User:      localUsername,
	}

	err := client.UpsertAssociation(baseAssociation)
	if err != nil {
		t.Fatalf("Failed to upsert association: %v", err)
	}

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
			return &models.Organization{ID: id, Name: orgName}, nil
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
		GetComputeClusterUserByPairFunc: func(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error) {
			return &models.ComputeClusterUser{ID: "csu-001", LocalUsername: localUsername}, nil
		},
		ListResourcesForAllocationFunc: func(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error) {
			return []models.ComputeAllocationResource{
				{ID: allocationResourceId, Name: partitionName, ResourceType: "cpu", ResourceAmount: 1000},
			}, nil
		},
		GetComputeAllocationMembershipFunc: func(ctx context.Context, id string) (*models.ComputeAllocationMembership, error) {
			return &models.ComputeAllocationMembership{
				ID:                  "membership-001",
				UserID:              "user-001",
				ComputeAllocationID: computeAllocationID,
			}, nil
		},
	}

	override := &models.ComputeAllocationMembershipResourceOverride{
		ID:                            "override-001",
		ComputeAllocationMembershipID: "membership-001",
		ComputeAllocationResourceID:   allocationResourceId,
		OverrideResourceAmount:        50,
		OverrideResourceTime:          500,
	}

	associationSubscriber := NewAssociationSubscriber(client, nil, mockCoreService)

	associationSubscriber.SubscribeToComputeAllocationMembershipResourceOverrideCreation(*override)

	if len(auditEvents) != 1 {
		t.Fatalf("Expected 1 audit event, but got %d", len(auditEvents))
	}

	if auditEvents[0].EventType != "ComputeAllocationMembershipResourceOverrideCreationSucceeded" {
		t.Fatalf("Expected audit event type 'ComputeAllocationMembershipResourceOverrideCreationSucceeded', but got '%s'", auditEvents[0].EventType)
	}
}
