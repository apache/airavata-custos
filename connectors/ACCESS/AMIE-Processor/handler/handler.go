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

package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type PacketHandler interface {
	Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error
	SupportsType() string
}

// AmieClient sends replies back to the AMIE server.
type AmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

// AuditService writes one audit_events row (source='amie') plus the matching
// amie_audit_extras row carrying (packet_id, event_id).
type AuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

const amieIdentitySource = "access"

func requireText(val, fieldName string) error {
	if strings.TrimSpace(val) == "" {
		return fmt.Errorf("'%s' must not be empty", fieldName)
	}
	return nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBody(packetJSON map[string]any) (map[string]any, error) {
	b, ok := packetJSON["body"]
	if !ok {
		return nil, fmt.Errorf("packet missing 'body'")
	}
	body, ok := b.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("packet 'body' is not an object")
	}
	return body, nil
}

var roleNormalizer = strings.NewReplacer(" ", "", "-", "", "_", "")

func normalizeRole(raw string) string {
	switch strings.ToUpper(roleNormalizer.Replace(raw)) {
	case "PI":
		return "PI"
	case "COPI":
		return "CO_PI"
	case "ALLOCATIONMANAGER":
		return "ALLOCATION_MANAGER"
	default:
		return "USER"
	}
}

func getResourceList(body map[string]any) []string {
	v, ok := body["ResourceList"]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// ensureUserIdentity is the idempotent upsert used by data_project_create
// and data_account_create. It is a no-op when the user already has an AMIE
// UserIdentity row for this globalID; creates one otherwise. Pre-existing
// rows are NOT touched here, attribute updates (org / orgCode / nsfStatus)
// are owned by request_user_modify.
func ensureUserIdentity(ctx context.Context, svc *service.Service, userID, globalID string) error {
	if existing, err := svc.GetUserIdentityBySourceAndExternalID(ctx, amieIdentitySource, globalID); err == nil {
		if existing.UserID == userID {
			return nil
		}
		return fmt.Errorf("user identity for %s=%s is bound to user %s (not %s)", amieIdentitySource, globalID, existing.UserID, userID)
	} else if !errors.Is(err, service.ErrNotFound) {
		return err
	}
	if _, err := svc.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     userID,
		Source:     amieIdentitySource,
		ExternalID: globalID,
	}); err != nil {
		return fmt.Errorf("create user identity: %w", err)
	}
	return nil
}

// flipUserMemberships flips every ComputeAllocationMembership the user holds
// under any allocation belonging to the given project (Custos project.id) to
// the given status. Returns the rows that were updated. Silently returns an
// empty slice when the project or user is unknown, the reply still goes back
// to AMIE but no state changes.
func flipUserMemberships(ctx context.Context, svc *service.Service, projectID, userID string, status models.AllocationStatus) ([]models.ComputeAllocationMembership, error) {
	project, err := svc.GetProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup project: %w", err)
	}
	allocations, err := svc.ListComputeAllocationsByProject(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("list allocations: %w", err)
	}
	var updated []models.ComputeAllocationMembership
	for _, a := range allocations {
		members, err := svc.ListMembersForAllocation(ctx, a.ID)
		if err != nil {
			return nil, fmt.Errorf("list memberships for allocation %s: %w", a.ID, err)
		}
		for _, m := range members {
			if m.UserID != userID {
				continue
			}
			flipped, err := svc.UpdateMembershipStatus(ctx, m.ID, status)
			if err != nil {
				return nil, fmt.Errorf("update membership %s: %w", m.ID, err)
			}
			updated = append(updated, *flipped)
		}
	}
	return updated, nil
}

// ensureOrganization looks up an Organization by its originated_id (the
// AMIE-side org code such as "TEST123"); creates one if missing, using the
// human-readable organization name from the packet.
func ensureOrganization(ctx context.Context, svc *service.Service, code, name string) (*models.Organization, error) {
	if code == "" {
		return nil, fmt.Errorf("organization code is empty")
	}
	if org, err := svc.GetOrganizationByOriginatedID(ctx, code); err == nil {
		return org, nil
	} else if !errors.Is(err, service.ErrNotFound) {
		return nil, err
	}
	if name == "" {
		name = code
	}
	return svc.CreateOrganization(ctx, &models.Organization{
		OriginatedID: code,
		Name:         name,
	})
}

// getInt64 reads a string-encoded integer from a packet body field. AMIE
// transmits numeric fields like ServiceUnitsAllocated as JSON strings.
func getInt64(body map[string]any, key string) (int64, error) {
	raw := getString(body, key)
	if raw == "" {
		return 0, fmt.Errorf("'%s' is empty", key)
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("'%s' is not an integer: %w", key, err)
	}
	return n, nil
}

// getDate reads a YYYY-MM-DD string from a packet body field. Returns the
// parsed time in UTC.
func getDate(body map[string]any, key string) (time.Time, error) {
	raw := getString(body, key)
	if raw == "" {
		return time.Time{}, fmt.Errorf("'%s' is empty", key)
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("'%s' is not a YYYY-MM-DD date: %w", key, err)
	}
	return t, nil
}

func getDNList(body map[string]any) []string {
	v, ok := body["DnList"]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
