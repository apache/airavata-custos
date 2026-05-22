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
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestValidateStatusesRejectUnsupportedValues(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "user", err: validateUserStatus(models.UserStatus("BANANA"))},
		{name: "project", err: validateProjectStatus(models.ProjectStatus("BANANA"))},
		{name: "allocation", err: validateAllocationStatus("status", models.AllocationStatus("BANANA"))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", tc.err)
			}
		})
	}
}

func TestValidateRequiredTimeRange(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	if err := validateRequiredTimeRange(start, end); err != nil {
		t.Fatalf("expected valid range: %v", err)
	}
	if err := validateRequiredTimeRange(time.Time{}, end); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected missing start_time to be invalid, got %v", err)
	}
	if err := validateRequiredTimeRange(end, start); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected reversed range to be invalid, got %v", err)
	}
}
