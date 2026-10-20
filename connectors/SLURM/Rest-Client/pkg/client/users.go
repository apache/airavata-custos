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

import (
	"fmt"
	"net/http"
)

// DefaultAccount is every user's default account. It runs no jobs, so a job names the account it charges, and no
// allocation's account is anyone's default, which Slurm would refuse to remove.
const DefaultAccount = "none"

// CreateUser also makes DefaultAccount the default of a user that exists.
func (c *Client) CreateUser(cluster, user string) error {
	noJobs := int64(0)
	if err := c.CreateAccount(Account{Name: DefaultAccount, Description: "Default account; runs no jobs", GrpJobs: &noJobs}, cluster); err != nil {
		return fmt.Errorf("create %s account: %w", DefaultAccount, err)
	}
	defaultAccount := map[string]string{"account": DefaultAccount}
	body := map[string]any{
		"association_condition": map[string]any{"accounts": []string{DefaultAccount}, "clusters": []string{cluster}, "users": []string{user}},
		"user":                  map[string]any{"default": defaultAccount},
	}
	if _, err := c.do("POST", "/slurmdb/v0.0."+c.apiVersion+"/users_association/", body, nil); err != nil {
		return err
	}
	// users_association leaves an existing user's default as it was.
	users := map[string]any{"users": []map[string]any{{"name": user, "default": defaultAccount}}}
	_, err := c.do("POST", "/slurmdb/v0.0."+c.apiVersion+"/users", users, nil)
	return err
}

func (c *Client) DeleteUser(name string) error {
	resp, err := c.do("DELETE", "/slurmdb/v0.0."+c.apiVersion+"/user/"+name, nil, nil)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return err
}
