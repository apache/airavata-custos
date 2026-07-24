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

package subscribers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// DefaultSignInCheckInterval is how often the sign-in health probe runs when
// the connector config does not set signin_check_interval. The registry is an
// external hosted service and sign-in state changes rarely, so this stays far
// slower than the cluster-side probe.
const DefaultSignInCheckInterval = 10 * time.Minute

// Pass-wide budget. Each registry call is separately bounded by the client's
// own HTTP timeout, so a slow registry degrades one user at a time instead
// of starving the tail of the list.
const signInProbeTimeout = 5 * time.Minute

// Matches the source the provisioner stores person ids under.
const registryIdentitySource = "comanage"

// registryReader is the one read the probe performs against the registry.
type registryReader interface {
	GetPersonComposite(identifier string) (json.RawMessage, error)
}

// SignInProbe reports sign-in health for every active member of this
// connector's cluster: the account must be provisioned and its person record
// must still resolve in the registry. Read-only on both sides; provisioning
// stays with the orchestrator.
//
// A person whose record resolves but has lost its username identifier is not
// detected. Catching that needs a composite parse; add it if it ever happens.
type SignInProbe struct {
	registry  registryReader
	core      service.CoreService
	clusterID string
}

func NewSignInProbe(registry registryReader, core service.CoreService, clusterID string) *SignInProbe {
	return &SignInProbe{registry: registry, core: core, clusterID: clusterID}
}

// Start runs the probe until ctx is cancelled.
func (p *SignInProbe) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultSignInCheckInterval
	}
	slog.Info("Starting sign-in access status probe", "interval", interval)

	p.probeOnce(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping sign-in access status probe", "reason", ctx.Err())
			return
		case <-ticker.C:
			p.probeOnce(ctx)
		}
	}
}

func (p *SignInProbe) probeOnce(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Sign-in probe: pass panicked", "panic", r)
		}
	}()
	probeCtx, cancel := context.WithTimeout(ctx, signInProbeTimeout)
	defer cancel()

	clusterUsers, err := p.core.ListComputeClusterUsersByCluster(probeCtx, p.clusterID)
	if err != nil {
		slog.Error("Sign-in probe: failed to list cluster users", "cluster_id", p.clusterID, "error", err)
		return
	}

	// One registry lookup per user per pass, shared across their memberships.
	type verdict struct {
		ok             bool
		detail         string
		infrastructure bool
	}
	verdicts := make(map[string]verdict, len(clusterUsers))

	for _, csu := range clusterUsers {
		memberships, err := p.core.ListAllocationsForUser(probeCtx, csu.UserID)
		if err != nil {
			slog.Error("Sign-in probe: failed to list memberships", "user_id", csu.UserID, "error", err)
			continue
		}
		for _, membership := range memberships {
			if membership.MembershipStatus != models.ACTIVE {
				continue
			}
			alloc, err := p.core.GetComputeAllocation(probeCtx, membership.ComputeAllocationID)
			if err != nil {
				slog.Warn("Sign-in probe: could not resolve allocation",
					"membership_id", membership.ID, "error", err)
				continue
			}
			if alloc.ComputeClusterID != p.clusterID || alloc.Status != models.ACTIVE {
				continue
			}
			v, seen := verdicts[csu.UserID]
			if !seen {
				var skip bool
				v.ok, v.detail, v.infrastructure, skip = p.verify(probeCtx, csu)
				if skip {
					continue
				}
				verdicts[csu.UserID] = v
			}
			err = p.core.RecordAccessCheckResult(probeCtx, membership.ComputeAllocationID, csu.UserID, models.AccessCheckSignIn, v.ok, v.detail, v.infrastructure)
			if err != nil {
				slog.Error("Sign-in probe: failed to record result",
					"membership_id", membership.ID, "error", err)
			}
		}
	}
}

// verify decides one user's sign-in health without writing anything. skip
// means the truth is unknowable right now (a core lookup failed), which must
// not be recorded as unhealthy. infrastructure marks a registry outage:
// recorded for freshness, shown as a calm retry, never escalated.
func (p *SignInProbe) verify(ctx context.Context, csu models.ComputeClusterUser) (ok bool, detail string, infrastructure bool, skip bool) {
	if csu.ProvisionedAt == nil {
		return false, "account not provisioned yet", false, false
	}
	idents, err := p.core.ListUserIdentitiesForUser(ctx, csu.UserID)
	if err != nil {
		slog.Warn("Sign-in probe: could not list identities", "user_id", csu.UserID, "error", err)
		return false, "", false, true
	}
	personID := ""
	for _, id := range idents {
		if id.Source == registryIdentitySource && id.ExternalID != "" {
			personID = id.ExternalID
			break
		}
	}
	if personID == "" {
		return false, "registry record not linked", false, false
	}
	if _, err := p.registry.GetPersonComposite(personID); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return false, "registry record missing", false, false
		}
		return false, "registry unreachable", true, false
	}
	return true, "", false, false
}
