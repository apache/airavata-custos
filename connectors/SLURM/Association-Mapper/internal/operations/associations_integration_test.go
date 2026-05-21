package operations

import (
	"os"
	"testing"
)

func TestCreateInvalidAssocation(t *testing.T) {
	if !isLocalSlurmConfigAvailable() {
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
