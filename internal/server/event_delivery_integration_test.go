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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

const subscriberTest = "test-subscriber"

// runWorkerUntil runs the worker until one delivery is in the given status and
// returns it. Fails the test after 5 seconds.
func runWorkerUntil(t *testing.T, bus *events.Bus, status models.EventDeliveryStatus) models.PendingDelivery {
	t.Helper()
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go bus.Run(ctx)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rows, err := bus.ListDeliveries(context.Background(), status, 10)
		if err != nil {
			t.Fatalf("list deliveries: %v", err)
		}
		if len(rows) == 1 {
			return rows[0]
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("no %s delivery within 5s", status)
	return models.PendingDelivery{}
}

// failDelivery publishes one event and returns its delivery once it has failed.
// The handler fails with ErrPermanent on the first call and succeeds
// on every call after that, so a retry can be checked.
func failDelivery(t *testing.T, svc *service.Service, bus *events.Bus) models.PendingDelivery {
	t.Helper()
	calls := 0
	bus.Subscribe(subscriberTest, events.OrganizationCreateEvent, func(context.Context, events.Event, any) error {
		calls++
		if calls == 1 {
			return fmt.Errorf("%w: first attempt", events.ErrPermanent)
		}
		return nil
	})
	if _, err := svc.CreateOrganization(context.Background(), &models.Organization{OriginatedID: "org-retry", Name: "Retry Org"}); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	return runWorkerUntil(t, bus, models.EventDeliveryFailed)
}

// Make sure a retried delivery is delivered again, with the attempt count starting over.
func TestRetryFailedDeliveryDeliversItAgain(t *testing.T) {
	_, svc, srv := setupTestStack(t)
	bus := svc.EventBus()
	failed := failDelivery(t, svc, bus)

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodPost, "/events/deliveries/"+failed.ID+"/retry", nil), "u-1", models.EventsManage)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204: %s", rr.Code, rr.Body.String())
	}

	if d := runWorkerUntil(t, bus, models.EventDeliverySucceeded); d.ID != failed.ID || d.Attempts != 1 {
		t.Fatalf("expected the same delivery succeeded on its first attempt after the retry, got %+v", d.EventDelivery)
	}
}

// Make sure only a failed delivery can be retried.
func TestRetryRefusesDeliveryThatHasNotFailed(t *testing.T) {
	_, svc, srv := setupTestStack(t)
	bus := svc.EventBus()
	bus.Subscribe(subscriberTest, events.OrganizationCreateEvent, func(context.Context, events.Event, any) error { return nil })
	if _, err := svc.CreateOrganization(context.Background(), &models.Organization{OriginatedID: "org-ok", Name: "Ok Org"}); err != nil {
		t.Fatalf("create organization: %v", err)
	}
	done := runWorkerUntil(t, bus, models.EventDeliverySucceeded)

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodPost, "/events/deliveries/"+done.ID+"/retry", nil), "u-1", models.EventsManage)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("status: got %d, want 409", rr.Code)
	}
}

// Make sure the events manage privilege is needed for retrying a delivery.
func TestRetryDeliveryRequiresEventsManage(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodPost, "/events/deliveries/any/retry", nil), "u-1", models.TracesRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want 403", rr.Code)
	}
}

// Make sure the subscription list marks a subscriber that is not loaded, so a connector that did not start can be seen.
func TestListSubscriptionsMarksSubscribersNotLoaded(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	svc.EventBus().Subscribe("loaded", events.OrganizationCreateEvent, func(context.Context, events.Event, any) error { return nil })
	other, err := events.New(context.Background(), database)
	if err != nil {
		t.Fatalf("second bus: %v", err)
	}
	other.Subscribe("not-loaded", events.OrganizationCreateEvent, func(context.Context, events.Event, any) error { return nil })

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/events/subscriptions", nil), "u-1", models.TracesRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rr.Code)
	}
	var body struct {
		Items []events.Subscription `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	loaded := map[string]bool{}
	for _, s := range body.Items {
		loaded[s.Subscriber] = s.Loaded
	}
	if len(loaded) != 2 || !loaded["loaded"] || loaded["not-loaded"] {
		t.Fatalf("expected loaded=true and not-loaded=false, got %v", loaded)
	}
}
