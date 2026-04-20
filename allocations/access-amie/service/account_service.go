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

package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/apache/airavata-custos/allocations/access-amie/db"
	"github.com/apache/airavata-custos/allocations/domain/model"
	"github.com/google/uuid"
)

type clusterAccountStore interface {
	FindByUsername(ctx context.Context, username string) (*model.ClusterAccount, error)
	FindByPerson(ctx context.Context, personID string) ([]model.ClusterAccount, error)
	Save(ctx context.Context, tx *sql.Tx, a *model.ClusterAccount) error
}

type UserAccountService struct {
	accounts clusterAccountStore
}

func NewUserAccountService(accounts clusterAccountStore) *UserAccountService {
	return &UserAccountService{accounts: accounts}
}

// ProvisionClusterAccount returns an existing cluster account for the person
// or creates a new one with a unique username derived from the person's name.
func (s *UserAccountService) ProvisionClusterAccount(ctx context.Context, tx *sql.Tx, person *model.Person) (*model.ClusterAccount, error) {
	existing, err := s.accounts.FindByPerson(ctx, person.ID)
	if err != nil {
		return nil, fmt.Errorf("account_service: finding accounts for person %s: %w", person.ID, err)
	}
	if len(existing) > 0 {
		return &existing[0], nil
	}

	username := generateUsername(person)
	uniqueUsername, err := s.ensureUniqueUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("account_service: ensuring unique username for person %s: %w", person.ID, err)
	}

	now := time.Now().UTC()
	acct := &model.ClusterAccount{
		ID:        uuid.NewString(),
		PersonID:  person.ID,
		Username:  uniqueUsername,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.accounts.Save(ctx, tx, acct); err != nil {
		// Handle MariaDB duplicate key race condition:
		// another transaction may have inserted a record between our check and insert.
		if db.IsDuplicateKeyError(err) {
			retryExisting, retryErr := s.accounts.FindByPerson(ctx, person.ID)
			if retryErr != nil {
				return nil, fmt.Errorf("account_service: retry finding accounts for person %s: %w", person.ID, retryErr)
			}
			if len(retryExisting) > 0 {
				return &retryExisting[0], nil
			}
		}
		return nil, fmt.Errorf("account_service: saving cluster account for person %s: %w", person.ID, err)
	}

	slog.DebugContext(ctx, "provisioned cluster account", "account_id", acct.ID, "person_id", person.ID, "username", uniqueUsername)
	return acct, nil
}

// generateUsername builds a candidate username from the person's first initial
// and last name (lowercased, spaces replaced with hyphens).
func generateUsername(p *model.Person) string {
	initial := ""
	if len(p.FirstName) > 0 {
		initial = string(p.FirstName[0])
	}
	last := strings.ReplaceAll(strings.TrimSpace(p.LastName), " ", "-")
	return strings.ToLower(initial + last)
}

// ensureUniqueUsername appends a numeric suffix if the candidate username is
// already taken.
func (s *UserAccountService) ensureUniqueUsername(ctx context.Context, base string) (string, error) {
	candidate := base
	suffix := 1
	for {
		existing, err := s.accounts.FindByUsername(ctx, candidate)
		if err != nil {
			return "", fmt.Errorf("checking username %s: %w", candidate, err)
		}
		if existing == nil {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s%d", base, suffix)
		suffix++
	}
}
