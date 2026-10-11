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
	"fmt"

	"github.com/apache/airavata-custos/internal/store"
)

const (
	// The dashboard offers 7/30/90 presets plus a custom window; anything
	// outside this range is rejected rather than silently clamped.
	minActivityWindowDays = 1
	maxActivityWindowDays = 365

	defaultActivityLimit = 10
	maxActivityLimit     = 200
)

// UserActivityPage is one page of the activity audit table together with the
// normalized filter it was produced from. Echoing the filter lets the portal
// reject a response that ignored the requested window or status.
type UserActivityPage struct {
	Items      []store.UserActivityRow
	Total      int
	Limit      int
	Offset     int
	WindowDays int
	Status     string
}

// ListUserActivity validates the requested filter and returns the matching
// page of users with an OIDC-linked identity.
func (s *Service) ListUserActivity(ctx context.Context, f store.UserActivityFilter) (*UserActivityPage, error) {
	normalized, err := normalizeActivityFilter(f)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.users.ListActivity(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("list user activity: %w", err)
	}
	if rows == nil {
		rows = []store.UserActivityRow{}
	}
	return &UserActivityPage{
		Items:      rows,
		Total:      total,
		Limit:      normalized.Limit,
		Offset:     normalized.Offset,
		WindowDays: normalized.WindowDays,
		Status:     normalized.Status,
	}, nil
}

// GetActivityAnalytics returns population-wide sign-in engagement metrics.
func (s *Service) GetActivityAnalytics(ctx context.Context, windowDays int) (*store.UserActivityAnalytics, error) {
	if err := validateActivityWindow(windowDays); err != nil {
		return nil, err
	}
	analytics, err := s.users.GetActivityAnalytics(ctx, nowUTC(), windowDays)
	if err != nil {
		return nil, fmt.Errorf("user activity analytics: %w", err)
	}
	return analytics, nil
}

// GetUserActivityAnalytics returns the same metrics for one OIDC-linked user.
func (s *Service) GetUserActivityAnalytics(ctx context.Context, userID string, windowDays int) (*store.UserActivityAnalytics, error) {
	if err := validateActivityWindow(windowDays); err != nil {
		return nil, err
	}
	if userID == "" {
		return nil, ErrNotFound
	}
	analytics, err := s.users.GetUserActivityAnalytics(ctx, userID, nowUTC(), windowDays)
	if err != nil {
		return nil, fmt.Errorf("user activity analytics: %w", err)
	}
	if analytics == nil {
		return nil, ErrNotFound
	}
	return analytics, nil
}

func normalizeActivityFilter(f store.UserActivityFilter) (store.UserActivityFilter, error) {
	if err := validateActivityWindow(f.WindowDays); err != nil {
		return f, err
	}
	switch f.Status {
	case store.UserActivityStatusAll, store.UserActivityStatusActive,
		store.UserActivityStatusDormant, store.UserActivityStatusNever:
	default:
		return f, fmt.Errorf("%w: status must be all, active, dormant or never", ErrInvalidInput)
	}
	switch f.Sort {
	case "", "last_login":
		f.Sort = "last_login"
	case "name", "login_count":
	default:
		return f, fmt.Errorf("%w: sort must be name, last_login or login_count", ErrInvalidInput)
	}
	switch f.Direction {
	case "", "desc":
		f.Direction = "desc"
	case "asc":
	default:
		return f, fmt.Errorf("%w: direction must be asc or desc", ErrInvalidInput)
	}
	if f.Limit <= 0 {
		f.Limit = defaultActivityLimit
	}
	if f.Limit > maxActivityLimit {
		f.Limit = maxActivityLimit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	if f.Now.IsZero() {
		f.Now = nowUTC()
	}
	return f, nil
}

func validateActivityWindow(windowDays int) error {
	if windowDays < minActivityWindowDays || windowDays > maxActivityWindowDays {
		return fmt.Errorf("%w: window must be between %d and %d days",
			ErrInvalidInput, minActivityWindowDays, maxActivityWindowDays)
	}
	return nil
}
