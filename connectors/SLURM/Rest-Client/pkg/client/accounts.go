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

// cli/internal/client/accounts.go
package client

import "fmt"

type accountsResponse struct {
	Accounts []Account `json:"accounts"`
}

func (c *Client) ListAccounts() ([]Account, error) {
	var out accountsResponse
	if _, err := c.do("GET", "/slurmdb/v0.0."+c.apiVersion+"/accounts", nil, &out); err != nil {
		return nil, err
	}
	return out.Accounts, nil
}

func (c *Client) GetAccount(name string) (*Account, error) {
	var out accountsResponse
	if _, err := c.do("GET", "/slurmdb/v0.0."+c.apiVersion+"/account/"+name, nil, &out); err != nil {
		return nil, err
	}
	if len(out.Accounts) == 0 {
		return nil, fmt.Errorf("account %q not found", name)
	}
	return &out.Accounts[0], nil
}

// CreateAccount creates an account and a cluster-scope association in a single
// call, mirroring `sacctmgr add account <name> cluster=<cluster>`. Without the
// association, the account record exists but is unusable (subsequent attempts
// to add a user silently no-op). slurmrestd exposes this via the
// /accounts_association/ endpoint, whose body wraps the account metadata and
// an association_condition naming the accounts + clusters to wire up.
func (c *Client) CreateAccount(a Account, cluster string) error {
	body := map[string]any{
		"association_condition": map[string]any{
			"accounts": []string{a.Name},
			"clusters": []string{cluster},
		},
		"account": map[string]any{
			"description":  a.Description,
			"organization": a.Organization,
		},
	}
	_, err := c.do("POST", "/slurmdb/v0.0."+c.apiVersion+"/accounts_association/", body, nil)
	return err
}

func (c *Client) DeleteAccount(name string) error {
	_, err := c.do("DELETE", "/slurmdb/v0.0."+c.apiVersion+"/account/"+name, nil, nil)
	return err
}
