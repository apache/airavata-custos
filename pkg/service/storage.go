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
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

const storageAuditQuotaChanged = "STORAGE_QUOTA_CHANGED"

// StorageUsage reads the bytes used under each storage path on a cluster, keyed by path.
type StorageUsage func(ctx context.Context) (map[string]int64, error)

// ListStorages lists every storage with its use, read from its cluster's storage connector.
func (s *Service) ListStorages(ctx context.Context) ([]models.Storage, error) {
	items, err := s.storages.List(ctx)
	if err != nil {
		return items, err
	}
	s.claimedMu.RLock()
	readers := maps.Clone(s.claimedClusters)
	s.claimedMu.RUnlock()
	used := map[string]map[string]int64{}
	for i, st := range items {
		byPath, read := used[st.ComputeClusterID]
		if usage := readers[st.ComputeClusterID]; !read && usage != nil {
			if byPath, err = usage(ctx); err != nil {
				slog.WarnContext(ctx, "storage usage unavailable; storages listed without it", "compute_cluster_id", st.ComputeClusterID, "err", err)
			}
			used[st.ComputeClusterID] = byPath
		}
		if b, ok := byPath[st.Path]; ok {
			items[i].UsedBytes = &b
		}
	}
	return items, nil
}

// storagePlaceholders are the holders a target's path can name: an account, an allocation, and a connector's user.
var storagePlaceholders = []string{"{user}", "{allocation}", "{name}"}

// SetStorageTargets sets how each cluster's storages of each type are made, keyed by cluster id. A path names {user},
// {allocation}, both, or {name} alone, and owner and group name no placeholder the path lacks.
func (s *Service) SetStorageTargets(targets map[string]map[models.StorageType]models.StorageTarget) error {
	for clusterID, clusterTargets := range targets {
		paths := map[string]bool{}
		for typ, target := range clusterTargets {
			named := strings.Contains(target.Path, "{name}")
			if named == (strings.Contains(target.Path, "{user}") || strings.Contains(target.Path, "{allocation}")) || paths[target.Path] ||
				slices.ContainsFunc(storagePlaceholders, func(p string) bool {
					return strings.Contains(target.Owner+target.Group, p) && !strings.Contains(target.Path, p)
				}) {
				return fmt.Errorf("%w: cluster %s storage target %s needs a path of its own naming {user}, {allocation}, both, or {name} alone, "+
					"and an owner and group naming only those", ErrInvalidInput, clusterID, typ)
			}
			paths[target.Path] = true
			if _, err := strconv.ParseUint(target.Mode, 8, 12); err != nil {
				return fmt.Errorf("%w: cluster %s storage target %s needs an octal mode, such as \"0700\"", ErrInvalidInput, clusterID, typ)
			}
		}
	}
	s.storageTargets = targets
	return nil
}

// StorageHolderOf says whether user u holds a storage of a type, and the {name} its target fills in.
type StorageHolderOf func(u models.User) (name string, ok bool)

// ClaimStorageHolders has core make a type's storages for the users holderOf accepts. Until a connector claims a type
// whose path names {name}, core leaves its storages as they are.
func (s *Service) ClaimStorageHolders(typ models.StorageType, holderOf StorageHolderOf) {
	s.claimedMu.Lock()
	defer s.claimedMu.Unlock()
	s.storageHolders[typ] = holderOf
}

// ClaimStorageTargets claims the cluster for the storage connector that makes its storages and reads their usage, and
// returns its targets. Core makes storages only for a claimed cluster.
func (s *Service) ClaimStorageTargets(clusterID string, usage StorageUsage) map[models.StorageType]models.StorageTarget {
	s.claimedMu.Lock()
	defer s.claimedMu.Unlock()
	s.claimedClusters[clusterID] = usage
	return s.storageTargets[clusterID]
}

// SyncStorages makes a cluster's storages match its targets and holders, and returns its active ones.
func (s *Service) SyncStorages(ctx context.Context, clusterID string) ([]models.Storage, error) {
	if err := s.inTx(ctx, func(tx *sql.Tx) error { return s.syncStoragesTx(ctx, tx, store.StorageScope{ClusterID: clusterID}) }); err != nil {
		return nil, err
	}
	storages, err := s.storages.List(ctx)
	return slices.DeleteFunc(storages, func(st models.Storage) bool { return st.ComputeClusterID != clusterID || st.Status != "active" }), err
}

// UpdateStorage puts an administrator's quota in place of the target's until it lapses.
func (s *Service) UpdateStorage(ctx context.Context, id string, bytes int64, until *time.Time, reason, actorID string) error {
	if bytes <= 0 || until == nil || !until.After(time.Now()) || until.After(time.Now().AddDate(0, 0, 90)) {
		return fmt.Errorf("%w: quota must be positive and until within the next 90 days", ErrInvalidInput)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		st, previousBytes, err := s.storages.SetQuota(ctx, tx, id, bytes, *until)
		if err != nil {
			return err
		}
		if st == nil {
			return fmt.Errorf("%w: no storage with a quota of its own", ErrNotFound)
		}
		if err := s.writeAuditTx(ctx, tx, "storage", storageAuditQuotaChanged, st.ComputeClusterID+":"+st.Path,
			map[string]any{"from": previousBytes, "to": bytes, "until": until, "reason": reason, "actor_id": actorID}); err != nil {
			return err
		}
		return s.eventBus.Publish(ctx, tx, events.StorageUpdateEvent, *st)
	})
}

func (s *Service) DeleteStorage(ctx context.Context, id string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.storages.Lock(ctx, tx); err != nil {
			return err
		}
		deleted, err := s.storages.Delete(ctx, tx, id)
		if err != nil {
			return err
		}
		if deleted == nil {
			return fmt.Errorf("%w: no pending storage, or something under it is active", ErrNotFound)
		}
		return s.eventBus.Publish(ctx, tx, events.StorageDeleteEvent, *deleted)
	})
}

// syncStoragesTx makes the storages the claimed clusters' targets give the scope's holders, retires the scope's
// others, and publishes each one made or changed, a parent first.
func (s *Service) syncStoragesTx(ctx context.Context, tx *sql.Tx, scope store.StorageScope) error {
	s.claimedMu.RLock()
	claimedClusterIDs := slices.Collect(maps.Keys(s.claimedClusters))
	holders := maps.Clone(s.storageHolders)
	s.claimedMu.RUnlock()
	if len(claimedClusterIDs) == 0 {
		return nil
	}
	if err := s.storages.Lock(ctx, tx); err != nil {
		return err
	}
	desired, err := s.desiredStorages(ctx, tx, scope, claimedClusterIDs, holders)
	if err != nil {
		return err
	}
	var unclaimedTypes []models.StorageType
	for _, clusterID := range claimedClusterIDs {
		for typ, target := range s.storageTargets[clusterID] {
			if _, ok := holders[typ]; !ok && strings.Contains(target.Path, "{name}") {
				unclaimedTypes = append(unclaimedTypes, typ)
			}
		}
	}
	if err := s.storages.Retire(ctx, tx, scope, claimedClusterIDs, unclaimedTypes, desired); err != nil {
		return err
	}
	changed, err := s.storages.Upsert(ctx, tx, desired)
	if err != nil {
		return err
	}
	for _, st := range changed {
		if err := s.eventBus.Publish(ctx, tx, events.StorageCreateEvent, st); err != nil {
			return err
		}
	}
	return nil
}

// desiredStorages fills each claimed cluster's target once per holder of its type.
func (s *Service) desiredStorages(ctx context.Context, tx *sql.Tx, scope store.StorageScope, claimedClusterIDs []string,
	holders map[models.StorageType]StorageHolderOf) ([]models.Storage, error) {
	var accounts, allocations []store.StorageHolder
	var err error
	// An allocation's scope holds no account's own storage, and a user's none of an allocation's.
	if scope.AllocationID == "" {
		if accounts, err = s.storages.AccountHolders(ctx, tx, scope); err != nil {
			return nil, err
		}
	}
	if scope.UserID == "" {
		if allocations, err = s.storages.AllocationHolders(ctx, tx, scope); err != nil {
			return nil, err
		}
	}
	members, err := s.storages.MemberHolders(ctx, tx, scope)
	if err != nil {
		return nil, err
	}
	var users []models.User
	if len(holders) > 0 && scope.AllocationID == "" {
		if users, err = s.storages.Users(ctx, tx, scope.UserID); err != nil {
			return nil, err
		}
	}
	holdersOf := func(typ models.StorageType, path string) []store.StorageHolder {
		switch user, allocation := strings.Contains(path, "{user}"), strings.Contains(path, "{allocation}"); {
		case user && allocation:
			return members
		case user:
			return accounts
		case allocation:
			return allocations
		}
		var named []store.StorageHolder
		if holderOf := holders[typ]; holderOf != nil {
			for _, u := range users {
				if name, ok := holderOf(u); ok {
					named = append(named, store.StorageHolder{UserID: &u.ID, Name: name})
				}
			}
		}
		return named
	}

	var desired []models.Storage
	seenPaths := map[[2]string]bool{}
	for _, clusterID := range claimedClusterIDs {
		for _, typ := range slices.Sorted(maps.Keys(s.storageTargets[clusterID])) {
			target := s.storageTargets[clusterID][typ]
			for _, holder := range holdersOf(typ, target.Path) {
				if holder.ClusterID != "" && holder.ClusterID != clusterID {
					continue
				}
				values := map[string]string{"{user}": holder.Username, "{allocation}": holder.PosixGroup, "{name}": holder.Name}
				if slices.ContainsFunc(storagePlaceholders, func(p string) bool { return strings.Contains(target.Path, p) && !pathPart(values[p]) }) {
					slog.WarnContext(ctx, "storage holder's name is no directory name; it gets no storage", "compute_cluster_id", clusterID, "type", typ)
					continue
				}
				expand := strings.NewReplacer("{user}", holder.Username, "{allocation}", holder.PosixGroup, "{name}", holder.Name).Replace
				key := [2]string{clusterID, expand(target.Path)}
				if seenPaths[key] {
					slog.WarnContext(ctx, "storage path given twice; the first holder keeps it", "compute_cluster_id", clusterID, "path", key[1], "type", typ)
					continue
				}
				seenPaths[key] = true
				desired = append(desired, models.Storage{ID: newID(), ComputeClusterID: clusterID, Type: typ, Path: key[1],
					ComputeClusterUserID: holder.ComputeClusterUserID, AllocationID: holder.AllocationID, UserID: holder.UserID,
					Owner: expand(target.Owner), Group: expand(target.Group), Mode: target.Mode, QuotaBytes: target.HardLimitSize, QuotaInodes: target.HardLimitInodes})
			}
		}
	}
	return desired, nil
}

// publishUserTx publishes a change to user u and brings the storages claimed holders give u in line with it.
func (s *Service) publishUserTx(ctx context.Context, tx *sql.Tx, topic events.EventType, u *models.User) error {
	if err := s.eventBus.Publish(ctx, tx, topic, u); err != nil {
		return err
	}
	s.claimedMu.RLock()
	claimedClusterIDs := slices.Collect(maps.Keys(s.claimedClusters))
	s.claimedMu.RUnlock()
	if topic == events.UserDeleteEvent {
		return s.storages.Retire(ctx, tx, store.StorageScope{UserID: u.ID}, claimedClusterIDs, nil, nil)
	}
	return s.syncStoragesTx(ctx, tx, store.StorageScope{UserID: u.ID})
}

// pathPart holds for a placeholder's value that names one directory under its parent.
func pathPart(v string) bool {
	return v != "" && v != "." && v != ".." && !strings.Contains(v, "/")
}
