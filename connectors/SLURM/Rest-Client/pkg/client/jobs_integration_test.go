package client

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestListJobs(t *testing.T) {
	if !IsLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for listing jobs because local SLURM config is not available")
	}

	apiUrl := "http://localhost:6820"
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := New(apiUrl, user, token, apiVersion)

	JobSubmitRequest := JobSubmitRequest{
		JobSubmitParam: JobSubmitParam{
			Name:        "test-job",
			Account:     "root",
			Partition:   "compute",
			Tasks:       1,
			CpusPerTask: 1,
			TimeLimit: slurmNumber{
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
	}

	currentTime := time.Now().Unix()
	totalJobsToSubmit := 12
	for i := 0; i < totalJobsToSubmit; i++ {
		resp, err := client.SubmitJob(JobSubmitRequest)
		if err != nil {
			t.Fatalf("Failed to submit job: %v", err)
		}

		if resp.JobID == 0 {
			t.Fatalf("Invalid job ID returned from job submission: %d", resp.JobID)
		}

		log.Printf("Submitted job with ID: %d", resp.JobID)
	}

	sleepDuration := 20 // seconds
	log.Printf("Sleeping for %d seconds to allow job to start...", sleepDuration)
	time.Sleep(time.Duration(sleepDuration) * time.Second)

	filter := JobFilter{
		// You can set filter parameters here if needed
		Users:     []string{"root"},
		StartTime: &slurmNumber{Set: true, Infinite: false, Number: currentTime},
	}

	jobs, err := client.ListJobs(filter)
	if err != nil {
		t.Fatalf("Failed to list jobs: %v", err)
	}
	if len(jobs) == 0 {
		t.Log("No jobs found")
	} else {
		t.Logf("Found %d jobs", len(jobs))
		if len(jobs) != totalJobsToSubmit {
			t.Logf("Expected at %d jobs, but found %d", totalJobsToSubmit, len(jobs))
		} else {
			t.Log("Successfully found all submitted jobs")
		}
	}
}
