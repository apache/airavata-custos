// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
