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

package operations

import (
	"context"

	"github.com/apache/airavata-custos/pkg/models"
)

// RemoveAllocationGroup unbinds a deleted allocation's CoGroup from the UnixCluster, then deletes it.
func (o *Orchestrator) RemoveAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error {
	if a.PosixGroup == nil {
		return nil
	}
	groupID, err := o.c.FindCoGroupByName(*a.PosixGroup)
	if err != nil || groupID == 0 {
		return err
	}
	return o.removeCoGroup(groupID)
}
