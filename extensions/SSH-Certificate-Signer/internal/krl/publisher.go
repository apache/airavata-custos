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

package krl

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/apache/airavata-custos/signer/internal/audit"
	"github.com/apache/airavata-custos/signer/internal/store"
)

type Snapshot struct {
	Bytes   []byte
	Version uint64
	At      time.Time
	Serials int
}

// Publisher writes the current KRL to the login nodes' RevokedKeys file and keeps a copy to serve.
type Publisher struct {
	db     *store.DB
	path   string
	audit  *audit.Logger
	logger *slog.Logger
	// Holds at most one pending rebuild. Revokes that land during a build fold into the next one.
	rebuild chan struct{}

	mu      sync.Mutex
	serials []int64
	current Snapshot
}

func NewPublisher(db *store.DB, path string, audit *audit.Logger, logger *slog.Logger) *Publisher {
	return &Publisher{db: db, path: path, audit: audit, logger: logger, rebuild: make(chan struct{}, 1)}
}

// Trigger asks for a rebuild without waiting for it.
func (p *Publisher) Trigger() {
	select {
	case p.rebuild <- struct{}{}:
	default:
	}
}

// Run rebuilds on every trigger and every refresh, so expired serials drop out and a failed write is retried.
func (p *Publisher) Run(ctx context.Context, refresh time.Duration) {
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.rebuild:
		case <-ticker.C:
		}
		if err := p.Publish(ctx); err != nil {
			p.logger.Error("failed to publish KRL", "error", err)
		}
	}
}

// Publish rebuilds the KRL from every revoked certificate that has not expired.
// Nothing is written when the set is unchanged.
func (p *Publisher) Publish(ctx context.Context) error {
	now := time.Now().UTC()
	// Keep serials an hour past expiry, for safety.
	serials, err := p.db.ListRevokedSerialsValidAfter(ctx, now.Add(-time.Hour))
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current.Bytes != nil && slices.Equal(serials, p.serials) {
		return nil
	}
	// Two builds in the same second must not share a version, or a node takes a 304 for new content.
	version := max(uint64(now.Unix()), p.current.Version+1)
	b := Build(version, now, serials)
	// The file is swapped in one step and never truncated. A missing or partial file locks every login out.
	tmp := p.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("writing KRL: %w", err)
	}
	if err := os.Rename(tmp, p.path); err != nil {
		return fmt.Errorf("writing KRL: %w", err)
	}
	p.serials = serials
	p.current = Snapshot{Bytes: b, Version: version, At: now, Serials: len(serials)}
	p.audit.LogKRLPublished(ctx, version, len(serials), p.path)
	return nil
}

func (p *Publisher) Snapshot() Snapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}
