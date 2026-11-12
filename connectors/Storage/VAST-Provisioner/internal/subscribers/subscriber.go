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

// Package subscribers acts on core's storage events: it makes a created storage on VAST, sets a changed quota and
// removes a deleted storage. Core decides when a storage is created; the sweep only reports use back.
package subscribers

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type Cluster struct {
	VMS           *client.Client
	Views         map[models.StorageType]bool
	ViewPolicyID  int
	ViewProtocols []string
}

type StorageSubscriber struct {
	bus      *events.Bus
	core     service.CoreService
	clusters map[string]Cluster
}

func NewStorageSubscriber(bus *events.Bus, core service.CoreService, clusters map[string]Cluster) *StorageSubscriber {
	return &StorageSubscriber{bus: bus, core: core, clusters: clusters}
}

func (s *StorageSubscriber) RegisterSubscribers(subscriber string) {
	s.bus.SubscribeStorageCreated(subscriber, s.createStorage)
	s.bus.SubscribeStorageUpdated(subscriber, s.updateStorage)
	s.bus.SubscribeStorageDeleted(subscriber, s.deleteStorage)
}

// createStorage makes the storage's folder, its view when its type has one, and its quota.
func (s *StorageSubscriber) createStorage(ctx context.Context, st models.Storage) error {
	c, ok := s.clusters[st.ComputeClusterID]
	if !ok {
		return nil
	}
	mode, err := models.ModeBits(st.Mode)
	if err != nil {
		return err
	}
	if err := c.VMS.CreateFolder(ctx, st.Path, st.Owner, st.Group, mode); err != nil {
		return err
	}
	if c.Views[st.Type] {
		if err := c.VMS.CreateView(ctx, st.Path, c.ViewPolicyID, c.ViewProtocols); err != nil {
			return err
		}
	}
	return s.updateStorage(ctx, st)
}

// updateStorage brings the storage's quota to its state. A storage without one counts against its parent's.
func (s *StorageSubscriber) updateStorage(ctx context.Context, st models.Storage) error {
	if c, ok := s.clusters[st.ComputeClusterID]; ok && st.QuotaBytes != 0 {
		return c.VMS.CreateQuota(ctx, st.Path, st.QuotaBytes, st.QuotaInodes)
	}
	return nil
}

func (s *StorageSubscriber) deleteStorage(ctx context.Context, st models.Storage) error {
	if c, ok := s.clusters[st.ComputeClusterID]; ok {
		return c.VMS.DeleteFolder(ctx, st.Path)
	}
	return nil
}

func (s *StorageSubscriber) StartReconciler(ctx context.Context, interval time.Duration) {
	for {
		for id, c := range s.clusters {
			s.reconcile(ctx, id, c)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}

func (s *StorageSubscriber) reconcile(ctx context.Context, id string, c Cluster) {
	ctx, span := tracing.Start(ctx, "vast.reconcile")
	defer span.End()
	// Core's rule first, then every storage again, so one missing on VAST comes back.
	_, err := s.core.EnsureStorages(ctx, id)
	for offset, total := 0, 1; offset < total; offset += 200 {
		page, n, listErr := s.core.ListStorages(ctx, id, 200, offset)
		total, err = n, errors.Join(err, listErr)
		for _, st := range page {
			if st.Status == "active" && st.PolicyID != nil {
				err = errors.Join(err, s.createStorage(ctx, st))
			}
		}
	}
	quotas, quotaErr := c.VMS.Quotas(ctx)
	err = errors.Join(err, quotaErr)
	var paths []string
	var used []int64
	for path, q := range quotas {
		paths, used = append(paths, path), append(used, q.Used)
	}
	if err = errors.Join(err, s.core.RecordStorageUsage(ctx, id, paths, used)); err != nil {
		slog.ErrorContext(ctx, "VAST reconciler", "cluster_id", id, "err", err)
	}
}
