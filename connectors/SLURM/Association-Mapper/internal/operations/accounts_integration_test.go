package operations

import (
	"os"
	"testing"
)

func TestAccountCreatiion_Integration(t *testing.T) {

	if !isLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for account creation because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := New(apiUrl, user, token, apiVersion)

	client.DeleteAccount("test_account")       // clean up before test in case it was left over from a previous failed test run
	defer client.DeleteAccount("test_account") // clean up after test
	crearteAndValidateAccount(t, client)
}

func TestAccountDeletion_Integration(t *testing.T) {

	if !isLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for account deletion because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := New(apiUrl, user, token, apiVersion)

	crearteAndValidateAccount(t, client)

	err := client.DeleteAccount("test_account")
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	accounts, err := client.ListAccounts()
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}

	for _, account := range accounts {
		if account.Name == "test_account" {
			t.Fatalf("Account was not deleted: %+v\n", account)
		}
	}

	t.Logf("Successfully deleted account. Remaining accounts: %+v\n", accounts)
}

func TestGetAccount_Integration(t *testing.T) {

	if !isLocalSlurmConfigAvailable() {
		t.Skip("Skipping integration test for get account because local SLURM config is not available")
	}

	apiUrl := os.Getenv("TEST_SLURM_API")
	user := os.Getenv("TEST_SLURM_USER")
	token := os.Getenv("TEST_SLURM_TOKEN")
	apiVersion := os.Getenv("TEST_SLURM_API_VERSION")

	client := New(apiUrl, user, token, apiVersion)

	client.DeleteAccount("test_account")       // clean up before test in case it was left over from a previous failed test run
	defer client.DeleteAccount("test_account") // clean up after test
	crearteAndValidateAccount(t, client)

	account, err := client.GetAccount("test_account")
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if account.Name != "test_account" {
		t.Fatalf("Expected account name 'test_account', got '%s'", account.Name)
	}

	if account.Description != "Test account for integration testing" {
		t.Fatalf("Expected account description 'Test account for integration testing', got '%s'", account.Description)
	}

	if account.Organization != "Test Organization" {
		t.Fatalf("Expected account organization 'Test Organization', got '%s'", account.Organization)
	}

	t.Logf("Successfully retrieved account: %+v\n", account)
}
