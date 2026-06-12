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

package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/posix"
	"github.com/apache/airavata-custos/pkg/service"
)

func allocateAndCreateClusterUser(ctx context.Context, svc *service.Service, clusterID, userID string) (*models.ComputeClusterUser, error) {
	user, err := svc.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lookup user %q: %w", userID, err)
	}

	base, truncated, err := posix.BuildBase(user, posix.Prefix())
	if err != nil {
		_, _ = svc.CreateAuditEvent(ctx, &models.AuditEvent{
			EventType:  "PosixUsernameUnbuildable",
			EntityID:   userID,
			EntityType: "user",
			Details:    err.Error(),
		})
		return nil, err
	}
	if truncated {
		_, _ = svc.CreateAuditEvent(ctx, &models.AuditEvent{
			EventType:  "PosixUsernameTruncated",
			EntityID:   userID,
			EntityType: "user",
			Details:    base,
		})
	}

	for n := 0; n < posix.MaxCollisionSuffix; n++ {
		candidate := base
		if n > 0 {
			candidate = base + strconv.Itoa(n+1)
		}
		ccu, err := svc.CreateComputeClusterUser(ctx, &models.ComputeClusterUser{
			ComputeClusterID: clusterID,
			UserID:           userID,
			LocalUsername:    candidate,
		})
		if err == nil {
			return ccu, nil
		}
		if errors.Is(err, service.ErrAlreadyExists) {
			continue
		}
		return nil, err
	}

	_, _ = svc.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  "PosixUsernameAllocatorExhausted",
		EntityID:   userID,
		EntityType: "user",
		Details:    base,
	})
	return nil, fmt.Errorf("posix username allocator exhausted for base %q", base)
}
