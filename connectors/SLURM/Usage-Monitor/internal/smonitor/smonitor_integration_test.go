package smonitor

import (
	operations "github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
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
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := operations.New(apiUrl, user, token, apiVersion)

	currentTime := time.Now().Unix()

	client.SubmitJob(operations.JobSubmitRequest{
		JobSubmitParam: operations.JobSubmitParam{
			Name:        "test-job",
			Account:     "root",
			Partition:   "compute",
			Tasks:       1,
			CpusPerTask: 1,
			TimeLimit: operations.SlurmNumber{
				Set:      true,
				Infinite: false,
				Number:   20, // 10 seconds
			},
			CurrentWorkingDir: "/home/testuser",
			Environment: []string{
				"TEST_ENV_VAR=test_value",
			},
			// You can set other job parameters here if needed
		},
		Script: "#!/bin/bash\nsleep 1", // Simple script that sleeps for 1 second
	})

	// Sleep for a while to allow the monitor to pick up the job
	sleepDuration := 4 // seconds
	t.Logf("Sleeping for %d seconds to allow monitor to pick up the job...", sleepDuration)
	time.Sleep(time.Duration(sleepDuration) * time.Second)

	filter := operations.JobFilter{
		// You can set filter parameters here if needed
		Users:     []string{"root"},
		StartTime: currentTime,
	}

	jobs, err := client.ListJobs(filter)
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}
	if len(jobs) == 0 {
		t.Log("No jobs found")
	}
	t.Logf("Found %d jobs", len(jobs))

}
