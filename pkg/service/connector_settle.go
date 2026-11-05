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

// Core keeps connector requests settled against the current subscriptions, so turning a connector on or off in
// config never strands work. A request whose topic lost its subscribers is done by core, as if never sent, and its
// pending deliveries are dropped. A subscriber new to a topic is sent the topic's current state, addressed to it
// alone, so a connector enabled late or given a new topic catches up. A request whose retries ran out is sent again,
// at most once per reissueAfter. Connectors handle every request idempotently; notifications are never resent.

package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

const settleInterval = 5 * time.Minute

// StartSettle settles connector requests and corrects every stale ready_at, at once and then every interval, until ctx ends.
func (s *Service) StartSettle(ctx context.Context) {
	ticker := time.NewTicker(settleInterval)
	defer ticker.Stop()
	for {
		passCtx, span := tracing.Start(ctx, "core.settle")
		if err := errors.Join(s.settleConnectors(passCtx), s.syncStaleMemberships(passCtx)); err != nil {
			slog.ErrorContext(passCtx, "core.settle", "error", err)
		}
		span.End()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// reissueAfter is how long a request whose retries ran out waits before core sends it again.
const reissueAfter = time.Hour

// requestTopics are the topics core awaits a connector on, so the ones it resends.
var requestTopics = []events.EventType{
	events.ComputeClusterUserIdentityRequestEvent,
	events.ComputeClusterUserStorageRequestEvent,
	events.ComputeAllocationSchedulerRequestEvent,
	events.ComputeAllocationGroupRequestEvent,
	events.ComputeAllocationMembershipStorageRequestEvent,
}

// clusterState is every account and allocation on one cluster.
type clusterState struct {
	accounts []models.ComputeClusterUser
	allocs   []models.ComputeAllocation
}

// settleConnectors does the requests nobody subscribes to, resends exhausted ones and syncs each new subscriber.
func (s *Service) settleConnectors(ctx context.Context) error {
	clusters, err := s.clusters.List(ctx)
	if err != nil {
		return err
	}
	states := make([]clusterState, len(clusters))
	for i, c := range clusters {
		if states[i].accounts, err = s.clusterUsers.FindByCluster(ctx, c.ID); err != nil {
			return err
		}
		if states[i].allocs, err = s.allocs.FindByCluster(ctx, c.ID); err != nil {
			return err
		}
	}
	errs := []error{s.eventBus.PurgeUnsubscribed(ctx), s.settleUnsubscribed(ctx, states), s.reissueExhausted(ctx, clusters, states)}
	subs, err := s.eventBus.Unsynced(ctx)
	errs = append(errs, err)
	for _, sub := range subs {
		errs = append(errs, s.inTx(ctx, func(tx *sql.Tx) error {
			send := func(payload any) error {
				return s.eventBus.PublishTo(ctx, tx, sub.Subscriber, events.EventType(sub.EventType), payload)
			}
			for i, c := range clusters {
				if err := s.replayTx(ctx, tx, events.EventType(sub.EventType), c.ID, states[i], send); err != nil {
					return err
				}
			}
			return s.eventBus.MarkSynced(ctx, tx, sub.Subscriber, sub.EventType)
		}))
	}
	return errors.Join(errs...)
}

// settleUnsubscribed runs the mark of every outstanding request whose topic nobody subscribes to.
func (s *Service) settleUnsubscribed(ctx context.Context, states []clusterState) error {
	off := func(t events.EventType) bool { return !s.eventBus.Subscribed(t) }
	var errs []error
	do := func(fn func(tx *sql.Tx) error) { errs = append(errs, s.inTx(ctx, fn)) }
	for _, st := range states {
		for _, cu := range st.accounts {
			switch {
			case cu.ApprovalStatus == models.ClusterAccountApproved && cu.ProvisionedAt == nil && off(events.ComputeClusterUserIdentityRequestEvent):
				do(func(tx *sql.Tx) error { return s.markProvisionedTx(ctx, tx, &cu) })
			case cu.ProvisionedAt != nil && cu.StorageProvisionedAt == nil && off(events.ComputeClusterUserStorageRequestEvent):
				do(func(tx *sql.Tx) error { return s.markStorageProvisionedTx(ctx, tx, cu.ID) })
			}
		}
		for _, a := range st.allocs {
			if a.SchedulerProvisionedAt == nil && off(events.ComputeAllocationSchedulerRequestEvent) {
				do(func(tx *sql.Tx) error { return s.markSchedulerProvisionedTx(ctx, tx, a.ID) })
			}
			if a.GroupProvisionedAt == nil && off(events.ComputeAllocationGroupRequestEvent) {
				do(func(tx *sql.Tx) error { return s.groupWithoutRegistryTx(ctx, tx, &a) })
			}
		}
	}
	return errors.Join(errs...)
}

// reissueExhausted sends each still current entity whose latest request ran out of retries to the same connector again.
func (s *Service) reissueExhausted(ctx context.Context, clusters []models.ComputeCluster, states []clusterState) error {
	stuck, err := s.eventBus.Exhausted(ctx, requestTopics, nowUTC().Add(-reissueAfter))
	if err != nil || len(stuck) == 0 {
		return err
	}
	type key struct{ topic, id string }
	subs := map[key][]string{}
	for _, r := range stuck {
		k := key{r.EventType, r.EntityID}
		subs[k] = append(subs[k], r.Subscriber)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for i, c := range clusters {
			for _, topic := range requestTopics {
				if err := s.replayTx(ctx, tx, topic, c.ID, states[i], func(p any) error {
					for _, sub := range subs[key{string(topic), entityID(p)}] {
						if err := s.eventBus.PublishTo(ctx, tx, sub, topic, p); err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func entityID(p any) string {
	switch v := p.(type) {
	case models.ComputeClusterUser:
		return v.ID
	case models.ComputeAllocation:
		return v.ID
	case models.ComputeAllocationMembership:
		return v.ID
	}
	return ""
}

// replayTx sends the current state of one cluster for topic; topics with no state to resend send nothing.
func (s *Service) replayTx(ctx context.Context, tx *sql.Tx, topic events.EventType, clusterID string, st clusterState, send func(any) error) error {
	var payloads []any
	switch topic {
	case events.ComputeClusterUserIdentityRequestEvent:
		for _, cu := range st.accounts {
			if cu.ApprovalStatus == models.ClusterAccountApproved {
				payloads = append(payloads, cu)
			}
		}
	case events.ComputeClusterUserStorageRequestEvent:
		for _, cu := range st.accounts {
			if cu.ProvisionedAt != nil {
				payloads = append(payloads, cu)
			}
		}
	case events.ComputeAllocationSchedulerRequestEvent, events.ComputeAllocationGroupRequestEvent:
		for _, a := range st.allocs {
			payloads = append(payloads, a)
		}
	case events.ComputeAllocationGroupProvisionEvent:
		for _, a := range st.allocs {
			if a.GroupProvisionedAt != nil {
				payloads = append(payloads, a)
			}
		}
	case events.ComputeAllocationMembershipReadyEvent, events.ComputeAllocationMembershipStorageRequestEvent:
		grouped := map[string]bool{}
		for _, a := range st.allocs {
			grouped[a.ID] = a.GroupProvisionedAt != nil
		}
		ms, err := s.memberships.FindTx(ctx, tx, store.MembershipFilter{ClusterID: clusterID})
		if err != nil {
			return err
		}
		for _, m := range ms {
			if m.ReadyAt != nil && (topic == events.ComputeAllocationMembershipReadyEvent || grouped[m.ComputeAllocationID]) {
				payloads = append(payloads, m)
			}
		}
	}
	for _, p := range payloads {
		if err := send(p); err != nil {
			return err
		}
	}
	return nil
}
