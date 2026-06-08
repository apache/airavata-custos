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
	"fmt"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

func validateUserStatus(status models.UserStatus) error {
	switch status {
	case models.UserActive, models.UserInactive, models.UserSuspended, models.UserMerged:
		return nil
	default:
		return fmt.Errorf("%w: unsupported user status %q", ErrInvalidInput, status)
	}
}

func validateProjectStatus(status models.ProjectStatus) error {
	switch status {
	case models.ProjectActive, models.ProjectInactive, models.ProjectDeleted:
		return nil
	default:
		return fmt.Errorf("%w: unsupported project status %q", ErrInvalidInput, status)
	}
}

func validateAllocationStatus(field string, status models.AllocationStatus) error {
	switch status {
	case models.ACTIVE, models.INACTIVE, models.DELETED:
		return nil
	default:
		return fmt.Errorf("%w: unsupported %s %q", ErrInvalidInput, field, status)
	}
}

func validateRequiredTimeRange(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("%w: start_time and end_time are required", ErrInvalidInput)
	}
	if !start.Before(end) {
		return fmt.Errorf("%w: start_time must be before end_time", ErrInvalidInput)
	}
	return nil
}
