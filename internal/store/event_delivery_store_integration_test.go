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

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

func newEvent(createdAt time.Time) *models.Event {
	return &models.Event{
		ID:        uuid.NewString(),
		EventType: "compute_cluster_user::create",
		Payload:   json.RawMessage(`{"id":"cu-1","local_username":"jdoe"}`),
		Source:    "core",
		TraceID:   "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:    "00f067aa0ba902b7",
		CreatedAt: createdAt,
	}
}

func newDelivery(eventID, subscriber string, nextRunAt time.Time) *models.EventDelivery {
	return &models.EventDelivery{
		ID:         uuid.NewString(),
		EventID:    eventID,
		Subscriber: subscriber,
		Status:     models.EventDeliveryPending,
		NextRunAt:  nextRunAt,
		CreatedAt:  nextRunAt,
	}
}

// Make sure saving a subscription that already exists keeps one row.
func TestSaveSubscriptionTwiceKeepsOneRow(t *testing.T) {
	database := setupTestDB(t)
	s := NewEventDeliveryStore(database)
	ctx := context.Background()

	for range 2 {
		if err := s.SaveSubscription(ctx, "comanage-identity-provisioner", "compute_cluster_user::create"); err != nil {
			t.Fatalf("save subscription: %v", err)
		}
	}

	subs, err := s.ListSubscriptions(ctx)
	if err != nil {
		t.Fatalf("list subscriptions: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
}

// Make sure only pending deliveries past their next_run_at, for a loaded
// subscriber, come back oldest first, each with its event and the same
// payload JSON. The worker picks rows with this query, so a retry set for
// later must not run early, and rows for a connector that is not running must
// not fill the batch.
func TestFindDueDeliveriesReturnsOnlyPastDueRows(t *testing.T) {
	database := setupTestDB(t)
	s := NewEventDeliveryStore(database)
	ctx := context.Background()
	now := time.Now().UTC()

	event := newEvent(now.Add(-4 * time.Minute))
	older := newDelivery(event.ID, "slurm-association-mapper", now.Add(-4*time.Minute))
	newer := newDelivery(event.ID, "comanage-identity-provisioner", now.Add(-3*time.Minute))
	done := newDelivery(event.ID, "analytics", now.Add(-2*time.Minute))
	notLoaded := newDelivery(event.ID, "storage", now.Add(-time.Minute))
	later := newDelivery(event.ID, "slurm-association-mapper", now.Add(time.Hour))

	inTx(t, database, func(tx *sql.Tx) error {
		if err := s.CreateEvent(ctx, tx, event); err != nil {
			return err
		}
		for _, d := range []*models.EventDelivery{older, newer, done, notLoaded, later} {
			if err := s.CreateDelivery(ctx, tx, d); err != nil {
				return err
			}
		}
		return s.MarkDeliverySucceeded(ctx, tx, done.ID, 1, now)
	})

	loaded := []string{"slurm-association-mapper", "comanage-identity-provisioner", "analytics"}
	due, err := s.FindDueDeliveries(ctx, now, loaded, 10)
	if err != nil {
		t.Fatalf("find due: %v", err)
	}
	if len(due) != 2 || due[0].ID != older.ID || due[1].ID != newer.ID {
		t.Fatalf("expected the two past-due pending rows oldest first, got %+v", due)
	}
	if due[0].Event.ID != event.ID || due[0].Event.EventType != event.EventType {
		t.Fatalf("expected the delivery to carry its event, got %+v", due[0].Event)
	}
	// jsonb reformats whitespace, so compare the decoded values.
	var got, want map[string]any
	if err := json.Unmarshal(due[0].Event.Payload, &got); err != nil {
		t.Fatalf("decode stored payload: %v", err)
	}
	_ = json.Unmarshal(event.Payload, &want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("payload changed in storage: got %v, want %v", got, want)
	}
}
