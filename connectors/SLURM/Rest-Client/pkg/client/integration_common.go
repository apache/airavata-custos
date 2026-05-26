package client

import "os"
import "testing"

func IsLocalSlurmConfigAvailable() bool {
	if os.Getenv("TEST_SLURM_API") == "" || os.Getenv("TEST_SLURM_USER") == "" || os.Getenv("TEST_SLURM_TOKEN") == "" || os.Getenv("TEST_SLURM_API_VERSION") == "" {
		return false
	}
	return true
}

func CrearteAndValidateAccount(t *testing.T, client *Client) {

	err := client.CreateAccount(Account{
		Name:         "test_account",
		Description:  "Test account for integration testing",
		Organization: "Test Organization",
	}, "artisan")

	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	accounts, err := client.ListAccounts()
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}

	if len(accounts) == 0 {
		t.Fatal("No accounts found after creation")
	}

	for _, account := range accounts {
		if account.Name == "test_account" {
			t.Logf("Successfully created account: %+v\n", account)
			return
		}
	}
}
