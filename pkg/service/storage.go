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
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

func (s *Service) ListStorages(ctx context.Context, clusterID string, limit, offset int) ([]models.Storage, int, error) {
	return s.storages.List(ctx, clusterID, limit, offset)
}

// EnsureStorages makes a cluster's storages match its policies and owners, and returns those it made or changed.
func (s *Service) EnsureStorages(ctx context.Context, clusterID string) (made []models.Storage, err error) {
	err = s.inTx(ctx, func(tx *sql.Tx) (err error) {
		made, err = s.ensureStorages(ctx, tx, store.StorageOwners{ClusterID: clusterID})
		return err
	})
	return made, err
}

// SetStorageQuota puts an administrator's quota in place of the policy's until it lapses.
func (s *Service) SetStorageQuota(ctx context.Context, id string, bytes int64, until *time.Time, reason, actor string) error {
	if bytes <= 0 || until == nil || !until.After(time.Now()) || until.After(time.Now().AddDate(0, 0, 90)) {
		return fmt.Errorf("%w: quota must be positive and until within the next 90 days", ErrInvalidInput)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		st, from, err := s.storages.SetQuota(ctx, tx, id, bytes, *until)
		if err != nil || st == nil {
			return cmp.Or(err, fmt.Errorf("%w: no storage with a quota of its own", ErrNotFound))
		}
		if err := s.writeAuditTx(ctx, tx, "storage", "STORAGE_QUOTA_CHANGED", st.ComputeClusterID+":"+st.Path,
			map[string]any{"from": from, "to": bytes, "until": until, "reason": reason, "actor_id": actor}); err != nil {
			return err
		}
		return s.eventBus.Publish(ctx, tx, events.StorageUpdateEvent, *st)
	})
}

func (s *Service) DeleteStorage(ctx context.Context, id string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		deleted, err := s.storages.Delete(ctx, tx, id)
		if err != nil || len(deleted) == 0 {
			return cmp.Or(err, fmt.Errorf("%w: no pending storage, or something under it is active", ErrNotFound))
		}
		for _, st := range deleted {
			if err := s.eventBus.Publish(ctx, tx, events.StorageDeleteEvent, st); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) RecordStorageUsage(ctx context.Context, clusterID string, paths []string, used []int64) error {
	return s.storages.RecordUsage(ctx, clusterID, paths, used)
}

func (s *Service) ListStoragePolicies(ctx context.Context, clusterID string) ([]models.StoragePolicy, error) {
	return s.storages.ListPolicies(ctx, clusterID)
}

func (s *Service) CreateStoragePolicy(ctx context.Context, p *models.StoragePolicy) (*models.StoragePolicy, error) {
	if err := validateStoragePolicy(p); err != nil {
		return nil, err
	}
	p.ID = newID()
	if err := s.storages.CreatePolicy(ctx, p); err != nil {
		return nil, fmt.Errorf("create storage policy: %w", err)
	}
	return p, nil
}

// UpdateStoragePolicy replaces what a policy says; its cluster and type stay. A storage takes the change on its next create.
func (s *Service) UpdateStoragePolicy(ctx context.Context, p *models.StoragePolicy) (*models.StoragePolicy, error) {
	if err := validateStoragePolicy(p); err != nil {
		return nil, err
	}
	if found, err := s.storages.UpdatePolicy(ctx, p); err != nil {
		return nil, fmt.Errorf("update storage policy: %w", err)
	} else if !found {
		return nil, fmt.Errorf("%w: no %s storage policy with that id", ErrNotFound, p.Type)
	}
	return p, nil
}

func validateStoragePolicy(p *models.StoragePolicy) error {
	perUser, perAllocation, ok := p.Type.Takes()
	if _, err := models.ModeBits(p.Mode); !ok || err != nil || p.Owner == "" || p.Group == "" ||
		!strings.HasPrefix(p.PathPattern, "/") || p.QuotaBytes < 0 || p.QuotaInodes < 0 || p.QuotaBytes == 0 && p.QuotaInodes != 0 {
		return fmt.Errorf("%w: a storage policy needs a storage type, an absolute path_pattern, an owner, a group, an octal mode such as \"0755\" and non-negative quotas, with quota_inodes only beside quota_bytes", ErrInvalidInput)
	}
	for placeholder, used := range map[string]bool{"{user}": perUser, "{allocation}": perAllocation} {
		if strings.Contains(p.PathPattern, placeholder) != used || !used && strings.Contains(p.Owner+p.Group, placeholder) {
			return fmt.Errorf("%w: a %s storage policy's path_pattern, owner and group misuse %s", ErrInvalidInput, p.Type, placeholder)
		}
	}
	return nil
}

func (s *Service) DeleteStoragePolicy(ctx context.Context, id string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if ok, err := s.storages.DeletePolicy(ctx, tx, id); err != nil || ok {
			return err
		}
		return ErrNotFound
	})
}

// ensureStorages makes the storages the policies give the owners and publishes each made or changed, a parent before
// what is under it.
func (s *Service) ensureStorages(ctx context.Context, tx *sql.Tx, owners store.StorageOwners) ([]models.Storage, error) {
	made, err := s.storages.Ensure(ctx, tx, owners)
	for _, st := range made {
		if err := s.eventBus.Publish(ctx, tx, events.StorageCreateEvent, st); err != nil {
			return nil, err
		}
	}
	return made, err
}
