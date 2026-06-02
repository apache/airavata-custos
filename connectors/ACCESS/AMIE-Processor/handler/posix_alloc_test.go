//go:build integration

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
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/posix"
)

func TestAllocateAndCreateClusterUser_CollisionRetry(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	ctx := context.Background()

	org, err := svc.CreateOrganization(ctx, &models.Organization{
		OriginatedID: uuid.NewString(),
		Name:         "posix-alloc-test-org",
	})
	if err != nil {
		t.Fatalf("create org: %v", err)
	}

	const n = 10
	userIDs := make([]string, n)
	for i := range userIDs {
		u, err := svc.CreateUser(ctx, &models.User{
			OrganizationID: org.ID,
			FirstName:      "Collision",
			LastName:       "Target",
			Email:          fmt.Sprintf("col-%s@example.invalid", uuid.NewString()),
		})
		if err != nil {
			t.Fatalf("create user %d: %v", i, err)
		}
		userIDs[i] = u.ID
	}

	type result struct {
		username string
		err      error
	}
	results := make([]result, n)
	var wg sync.WaitGroup
	for i, uid := range userIDs {
		wg.Add(1)
		go func(idx int, userID string) {
			defer wg.Done()
			ccu, err := allocateAndCreateClusterUser(ctx, svc, testClusterID, userID)
			if err != nil {
				results[idx] = result{err: err}
				return
			}
			results[idx] = result{username: ccu.LocalUsername}
		}(i, uid)
	}
	wg.Wait()

	seen := make(map[string]bool)
	for i, r := range results {
		if r.err != nil {
			t.Errorf("goroutine %d: %v", i, r.err)
			continue
		}
		if seen[r.username] {
			t.Errorf("duplicate username %q across goroutines", r.username)
		}
		seen[r.username] = true
		if !strings.HasPrefix(r.username, posix.Prefix()+"-") {
			t.Errorf("username %q does not start with %q", r.username, posix.Prefix()+"-")
		}
	}
	if len(seen) != n {
		t.Errorf("distinct usernames: got %d, want %d", len(seen), n)
	}
}
