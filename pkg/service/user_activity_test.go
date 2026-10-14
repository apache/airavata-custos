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
	"errors"
	"testing"

	"github.com/apache/airavata-custos/internal/store"
)

func baseActivityFilter() store.UserActivityFilter {
	return store.UserActivityFilter{WindowDays: 30, Status: store.UserActivityStatusAll}
}

func TestNormalizeActivityFilterAppliesDefaults(t *testing.T) {
	normalized, err := normalizeActivityFilter(baseActivityFilter())
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if normalized.Sort != "last_login" || normalized.Direction != "desc" {
		t.Errorf("defaults: sort=%q direction=%q, want last_login/desc",
			normalized.Sort, normalized.Direction)
	}
	if normalized.Limit != defaultActivityLimit || normalized.Offset != 0 {
		t.Errorf("paging defaults: limit=%d offset=%d", normalized.Limit, normalized.Offset)
	}
	// One anchor per request keeps the status filter, inactive_days and the
	// window counts on the same calendar day.
	if normalized.Now.IsZero() {
		t.Error("the filter must carry the instant every local boundary is derived from")
	}
}

func TestNormalizeActivityFilterClampsPaging(t *testing.T) {
	f := baseActivityFilter()
	f.Limit = maxActivityLimit + 50
	f.Offset = -10
	normalized, err := normalizeActivityFilter(f)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if normalized.Limit != maxActivityLimit {
		t.Errorf("limit: got %d, want %d", normalized.Limit, maxActivityLimit)
	}
	if normalized.Offset != 0 {
		t.Errorf("offset: got %d, want 0", normalized.Offset)
	}
}

func TestNormalizeActivityFilterRejectsUnknownValues(t *testing.T) {
	cases := map[string]func(*store.UserActivityFilter){
		"window below the range": func(f *store.UserActivityFilter) { f.WindowDays = 0 },
		"window above the range": func(f *store.UserActivityFilter) { f.WindowDays = 366 },
		"unknown status":         func(f *store.UserActivityFilter) { f.Status = "stale" },
		"empty status":           func(f *store.UserActivityFilter) { f.Status = "" },
		"unknown sort key":       func(f *store.UserActivityFilter) { f.Sort = "email" },
		"unknown direction":      func(f *store.UserActivityFilter) { f.Direction = "sideways" },
	}
	for name, mutate := range cases {
		f := baseActivityFilter()
		mutate(&f)
		if _, err := normalizeActivityFilter(f); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%s: got %v, want ErrInvalidInput", name, err)
		}
	}
}

func TestNormalizeActivityFilterAcceptsEveryStatus(t *testing.T) {
	for _, status := range []string{
		store.UserActivityStatusAll,
		store.UserActivityStatusActive,
		store.UserActivityStatusDormant,
		store.UserActivityStatusNever,
	} {
		f := baseActivityFilter()
		f.Status = status
		normalized, err := normalizeActivityFilter(f)
		if err != nil {
			t.Errorf("status %q: %v", status, err)
			continue
		}
		if normalized.Status != status {
			t.Errorf("status %q was rewritten to %q", status, normalized.Status)
		}
	}
}

func TestValidateActivityWindowBoundaries(t *testing.T) {
	for _, window := range []int{minActivityWindowDays, 7, 30, 90, maxActivityWindowDays} {
		if err := validateActivityWindow(window); err != nil {
			t.Errorf("window %d should be accepted: %v", window, err)
		}
	}
	for _, window := range []int{-1, 0, maxActivityWindowDays + 1} {
		if err := validateActivityWindow(window); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("window %d: got %v, want ErrInvalidInput", window, err)
		}
	}
}

func TestActivityAnalyticsRejectBadWindowBeforeTouchingTheStore(t *testing.T) {
	// A nil store proves validation happens first: reaching it would panic.
	svc := &Service{}
	if _, err := svc.GetActivityAnalytics(t.Context(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("population analytics: got %v, want ErrInvalidInput", err)
	}
	if _, err := svc.GetUserActivityAnalytics(t.Context(), "u-1", 400); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("per-user analytics: got %v, want ErrInvalidInput", err)
	}
	if _, err := svc.GetUserActivityAnalytics(t.Context(), "", 30); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing user id: got %v, want ErrNotFound", err)
	}
	if _, err := svc.ListUserActivity(t.Context(), store.UserActivityFilter{WindowDays: 30, Status: "stale"}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("list: got %v, want ErrInvalidInput", err)
	}
}
