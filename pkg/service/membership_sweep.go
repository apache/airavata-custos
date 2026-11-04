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
	"errors"
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/internal/tracing"
)

const readySweepInterval = 5 * time.Minute

// StartReadySweep settles connector requests and corrects every stale ready_at, at once and then every interval, until ctx ends.
func (s *Service) StartReadySweep(ctx context.Context) {
	ticker := time.NewTicker(readySweepInterval)
	defer ticker.Stop()
	for {
		passCtx, span := tracing.Start(ctx, "core.ready_sweep")
		if err := errors.Join(s.settleConnectors(passCtx), s.syncStaleMemberships(passCtx)); err != nil {
			slog.ErrorContext(passCtx, "core.ready_sweep", "error", err)
		}
		span.End()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// syncStaleMemberships syncs each membership whose ready_at disagrees with its eligibility, one transaction each.
func (s *Service) syncStaleMemberships(ctx context.Context) error {
	ids, err := s.memberships.FindStaleReady(ctx)
	if err != nil {
		return err
	}
	errs := make([]error, len(ids))
	for i, id := range ids {
		errs[i] = s.inTx(ctx, func(tx *sql.Tx) error { return s.syncMembershipTx(ctx, tx, id) })
	}
	return errors.Join(errs...)
}
