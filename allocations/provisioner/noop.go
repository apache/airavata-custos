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

package provisioner

import "context"

// Compile-time interface check.
var _ Provisioner = (*Noop)(nil)

// Noop is a Provisioner that performs no operations. Used during development
// and as a fallback when no real provisioner is configured.
type Noop struct{}

// NewNoop returns a new Noop provisioner.
func NewNoop() *Noop {
	return &Noop{}
}

// ProvisionAccount is a no-op and returns a zero-value AccountResult.
func (n *Noop) ProvisionAccount(_ context.Context, _ AccountRequest) (AccountResult, error) {
	return AccountResult{}, nil
}

// DeprovisionAccount is a no-op.
func (n *Noop) DeprovisionAccount(_ context.Context, _ string) error {
	return nil
}

// ProvisionProject is a no-op.
func (n *Noop) ProvisionProject(_ context.Context, _ ProjectRequest) error {
	return nil
}

// DeprovisionProject is a no-op.
func (n *Noop) DeprovisionProject(_ context.Context, _ string) error {
	return nil
}

// HealthCheck is a no-op and always reports healthy.
func (n *Noop) HealthCheck(_ context.Context) error {
	return nil
}
