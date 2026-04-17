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

// Package provisioner defines the interface for HPC cluster account and project
// provisioning. It provides a common abstraction that decouples allocation
// management logic from the details of individual cluster provisioning systems.
// Implementations handle the creation and removal of user accounts and project
// allocations on target HPC resources.
package provisioner

import "context"

// AccountRequest contains the information needed to provision a user account
// on an HPC cluster.
type AccountRequest struct {
	PersonID  string
	FirstName string
	LastName  string
	Email     string
	ProjectID string
	Role      string
	DNList    []string
	Resources []string
}

// AccountResult contains the details of a successfully provisioned account.
type AccountResult struct {
	Username  string
	AccountID string
	HomeDir   string
	UID       int
}

// ProjectRequest contains the information needed to provision a project
// allocation on one or more HPC resources.
type ProjectRequest struct {
	ProjectID   string
	GrantNumber string
	Resources   []string
}

// Provisioner is the interface that HPC cluster provisioning backends must
// implement. Each method accepts a context for cancellation and deadline
// propagation. Implementations are expected to be safe for concurrent use.
type Provisioner interface {
	// ProvisionAccount creates a user account on the target HPC resource(s)
	// described by the request. It returns the resulting account details or
	// an error if provisioning fails.
	ProvisionAccount(ctx context.Context, req AccountRequest) (AccountResult, error)

	// DeprovisionAccount removes or disables the account identified by
	// username from the target HPC resource(s).
	DeprovisionAccount(ctx context.Context, username string) error

	// ProvisionProject creates a project allocation on the target HPC
	// resource(s) described by the request.
	ProvisionProject(ctx context.Context, req ProjectRequest) error

	// DeprovisionProject removes or disables the project allocation
	// identified by projectID from the target HPC resource(s).
	DeprovisionProject(ctx context.Context, projectID string) error

	// HealthCheck verifies that the provisioning backend is reachable and
	// operational. It returns nil when healthy or an error describing the
	// failure.
	HealthCheck(ctx context.Context) error
}
