package client

import (
	"os"
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
