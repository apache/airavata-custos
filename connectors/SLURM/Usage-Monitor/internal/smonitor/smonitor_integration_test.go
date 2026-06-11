package smonitor

import (
	"context"
	operations "github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
	"os"
	"testing"
	"time"
)

func TestSlurmMonitorIntegration(t *testing.T) {
	if !operations.IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for listing jobs because local SLURM config is not available")
	}

	apiUrl := "http://localhost:6820"
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	user1Token := os.Getenv("TEST_SLURM_TOKEN2")
	user2Token := os.Getenv("TEST_SLURM_TOKEN3")
	user3Token := os.Getenv("TEST_SLURM_TOKEN4")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	clusterName := "artisan"
	user1 := "testuser"
	user2 := "testuser2"
	user3 := "testuser3"
	rootClient := operations.New(apiUrl, user, token, apiVersion)
	user1Client := operations.New(apiUrl, user1, user1Token, apiVersion)
	user2Client := operations.New(apiUrl, user2, user2Token, apiVersion)
	user3Client := operations.New(apiUrl, user3, user3Token, apiVersion)

	currentTime := time.Now().Unix()

	accountName := "researchers"
	slurmAccount := operations.Account{
		Name:         accountName,
		Description:  "Researchers account",
		Organization: "researchers-org",
	}

	rootClient.CreateAccount(slurmAccount, clusterName)

	rootClient.UpsertAssociation(operations.Association{
		Account: accountName,
		Cluster: clusterName,
		User:    user1,
	})

	defer rootClient.DeleteAssociation(operations.AssocFilter{
		Account: accountName,
		Cluster: clusterName,
		User:    user1,
	})

	rootClient.UpsertAssociation(operations.Association{
		Account: accountName,
		Cluster: clusterName,
		User:    user2,
	})

	defer rootClient.DeleteAssociation(operations.AssocFilter{
		Account: accountName,
		Cluster: clusterName,
		User:    user2,
	})

	rootClient.UpsertAssociation(operations.Association{
		Account: accountName,
		Cluster: clusterName,
		User:    user3,
	})

	defer rootClient.DeleteAssociation(operations.AssocFilter{
		Account: accountName,
		Cluster: clusterName,
		User:    user3,
	})

	defer rootClient.DeleteAccount(accountName)

	user1Client.SubmitJob(operations.JobSubmitRequest{
		JobSubmitParam: operations.JobSubmitParam{
			Name:        "test-job",
			Account:     accountName,
			Partition:   "compute",
			Tasks:       1,
			CpusPerTask: 1,
			TimeLimit: operations.SlurmNumber{
				Set:      true,
				Infinite: false,
				Number:   20, // 10 seconds
			},
			CurrentWorkingDir: "/home/" + user1,
			Environment: []string{
				"TEST_ENV_VAR=test_value",
			},
			// You can set other job parameters here if needed
		},
		Script: "#!/bin/bash\nsleep 2", // Simple script that sleeps for 1 second
	})

	user2Client.SubmitJob(operations.JobSubmitRequest{
		JobSubmitParam: operations.JobSubmitParam{
			Name:        "test-job2",
			Account:     accountName,
			Partition:   "compute",
			Tasks:       1,
			CpusPerTask: 1,
			TimeLimit: operations.SlurmNumber{
				Set:      true,
				Infinite: false,
				Number:   20, // 10 seconds
			},
			CurrentWorkingDir: "/home/" + user2,
			Environment: []string{
				"TEST_ENV_VAR=test_value",
			},
			// You can set other job parameters here if needed
		},
		Script: "#!/bin/bash\nsleep 2", // Simple script that sleeps for 1 second
	})

	user3Client.SubmitJob(operations.JobSubmitRequest{
		JobSubmitParam: operations.JobSubmitParam{
			Name:        "test-job3",
			Account:     accountName,
			Partition:   "compute",
			Tasks:       1,
			CpusPerTask: 1,
			TimeLimit: operations.SlurmNumber{
				Set:      true,
				Infinite: false,
				Number:   20, // 10 seconds
			},
			CurrentWorkingDir: "/home/" + user3,
			Environment: []string{
				"TEST_ENV_VAR=test_value",
			},
			// You can set other job parameters here if needed
		},
		Script: "#!/bin/bash\nsleep 2", // Simple script that sleeps for 1 second
	})

	// Sleep for a while to allow the monitor to pick up the job
	sleepDuration := 8 // seconds
	t.Logf("Sleeping for %d seconds to allow monitor to pick up the job...", sleepDuration)
	time.Sleep(time.Duration(sleepDuration) * time.Second)

	filter := operations.JobFilter{
		// You can set filter parameters here if needed
		Users:     []string{"root", user1, user2, user3},
		StartTime: currentTime,
	}

	jobs, err := rootClient.ListJobs(filter)
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}
	if len(jobs) == 0 {
		t.Log("No jobs found")
	}
	t.Logf("Found %d jobs", len(jobs))

	comAllcUsages := make([]*models.ComputeAllocationUsage, 0)

	mockCoreService := &service.CoreServiceMock{

		GetComputeClusterFunc: func(ctx context.Context, clusterId string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{
				ID:   clusterId,
				Name: "artisan",
			}, nil
		},

		ListComputeAllocationsByClusterFunc: func(ctx context.Context, clusterId string) ([]models.ComputeAllocation, error) {
			return []models.ComputeAllocation{
				{
					ID:               "allocation-1",
					Name:             accountName, // Match the SLURM account name to link jobs to this allocation
					ComputeClusterID: clusterId,
				},
				{
					ID:               "allocation-2",
					Name:             "Test Allocation 2",
					ComputeClusterID: clusterId,
				},
			}, nil
		},

		GetComputeClusterUserByLocalUsernameAndClusterFunc: func(ctx context.Context, localUsername string, clusterId string) (*models.ComputeClusterUser, error) {
			return &models.ComputeClusterUser{
				ID:               "user-" + localUsername,
				ComputeClusterID: clusterId,
				UserID:           "user-id-" + localUsername,
				LocalUsername:    localUsername,
			}, nil
		},

		CreateComputeAllocationUsageFunc: func(ctx context.Context, u *models.ComputeAllocationUsage) (*models.ComputeAllocationUsage, error) {
			t.Logf("CreateComputeAllocationUsage called with: %+v", u)
			comAllcUsages = append(comAllcUsages, u)
			return u, nil
		},

		GetComputeAllocationResourceByNameAndClusterFunc: func(ctx context.Context, name string, clusterId string) (*models.ComputeAllocationResource, error) {
			return &models.ComputeAllocationResource{
				ID:               "resource-" + name,
				Name:             name,
				ResourceType:     "cpu",
				ResourceAmount:   4,
				ComputeClusterID: clusterId,
			}, nil
		},
	}

	monitor := &SlurmMonitor{
		slurmClient:     rootClient, // You can use a mock or real client here depending on your testing strategy
		eventBus:        nil,        // You can use a mock or real event bus here depending on your testing strategy
		coreService:     mockCoreService,
		clusterId:       "test-cluster",
		lastMonitorTime: currentTime, // initialize to 1 to avoid issues with zero value
	}

	// Call the poll method directly for testing
	monitor.poll()

	t.Logf("Total CreateComputeAllocationUsage calls: %d", len(comAllcUsages))

	// Validate that CreateComputeAllocationUsage was called with expected values
	if len(comAllcUsages) == 0 {
		t.Fatal("Expected CreateComputeAllocationUsage to be called at least once, but it was not called")
	}

	for _, usage := range comAllcUsages {
		t.Logf("Validating ComputeAllocationUsage: %+v", usage)
		if usage.ComputeAllocationID != "allocation-1" {
			t.Errorf("Expected ComputeAllocationID to be 'allocation-1', got '%s'", usage.ComputeAllocationID)
		}
		if usage.UsedRawAmount <= 0 {
			t.Errorf("Expected UsedRawAmount to be greater than 0, got %d", usage.UsedRawAmount)
		}
		if usage.UsedSUAmount <= 0 {
			t.Errorf("Expected UsedSUAmount to be greater than 0, got %d", usage.UsedSUAmount)
		}
		if usage.UserID != "user-"+usage.UserID[5:] { // Extract local username from UserID
			t.Errorf("Expected UserID to be 'user-%s', got '%s'", usage.UserID[5:], usage.UserID)
		}
		if usage.JobID == "" {
			t.Error("Expected JobID to be set, but it was empty")
		}
	}

}
