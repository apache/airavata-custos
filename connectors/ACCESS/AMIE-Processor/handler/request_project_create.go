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
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type RequestProjectCreateHandler struct {
	svc        *service.Service
	clusterID  string
	amieClient AmieClient
	auditSvc   AuditService
}

func NewRequestProjectCreateHandler(svc *service.Service, clusterID string, amieClient AmieClient, auditSvc AuditService) *RequestProjectCreateHandler {
	return &RequestProjectCreateHandler{svc: svc, clusterID: clusterID, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestProjectCreateHandler) SupportsType() string { return "request_project_create" }

// Handle ensures the PI user (resolving the organization from PiOrgCode +
// PiOrganization), creates (or finds) the Project, and creates a
// ComputeAllocation populated from the packet body's ServiceUnitsAllocated,
// StartDate and EndDate.
func (h *RequestProjectCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) (err error) {
	ctx, span := tracing.Start(ctx, "amie.handle:"+packet.Type)
	defer span.End()
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}
	grantNumber := getString(body, "GrantNumber")
	if err := requireText(grantNumber, "GrantNumber"); err != nil {
		return err
	}
	piGlobalID := getString(body, "PiGlobalID")
	if err := requireText(piGlobalID, "PiGlobalID"); err != nil {
		return err
	}
	if h.clusterID == "" {
		return fmt.Errorf("AMIE_CLUSTER_ID not configured")
	}
	// AMIE protocol: request_project_create does not carry a ProjectID. The
	// receiving site assigns one. We use the GrantNumber as the originated_id
	// (the stable cross-site identifier on the AMIE side) so first delivery
	// and supplement/renewal re-deliveries map to the same Project row.
	//
	// TODO verify whether GrantNumber should be promoted to a first-class column on `projects`
	projectOriginatedID := grantNumber

	pi, err := h.ensurePIUser(ctx, body, piGlobalID)
	if err != nil {
		return fmt.Errorf("request_project_create: ensure PI: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreatePerson, "user", pi.ID, "PI"); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_PERSON: %w", err)
	}

	piClusterUser, err := h.ensurePIClusterUser(ctx, pi.ID)
	if err != nil {
		return fmt.Errorf("request_project_create: ensure PI cluster user: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateAccount, "compute_cluster_user", piClusterUser.ID, piClusterUser.LocalUsername); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_ACCOUNT (PI): %w", err)
	}

	project, err := h.ensureProject(ctx, projectOriginatedID, grantNumber, pi.ID)
	if err != nil {
		return fmt.Errorf("request_project_create: ensure project: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateProject, "project", project.ID, ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_PROJECT: %w", err)
	}

	allocation, err := h.ensureAllocation(ctx, body, project.ID, grantNumber)
	if err != nil {
		return fmt.Errorf("request_project_create: ensure allocation: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateAllocation, "compute_allocation", allocation.ID, ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_ALLOCATION: %w", err)
	}

	if err := h.ensureResourceMappings(ctx, tx, body, allocation, packet, eventID); err != nil {
		return fmt.Errorf("request_project_create: ensure resource mappings: %w", err)
	}

	piMembership, created, err := h.ensurePIMembership(ctx, allocation, pi.ID)
	if err != nil {
		return fmt.Errorf("request_project_create: ensure PI membership: %w", err)
	}
	if err := h.svc.EnsureProjectMembership(ctx, project.ID, pi.ID, "PI"); err != nil {
		return fmt.Errorf("request_project_create: ensure PI project membership: %w", err)
	}
	if created {
		if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateMembership, "compute_allocation_membership", piMembership.ID, "PI"); err != nil {
			return fmt.Errorf("request_project_create: audit CREATE_MEMBERSHIP (PI): %w", err)
		}
	}

	replyBody := map[string]any{
		"ProjectID":         project.ID,
		"GrantNumber":       grantNumber,
		"PiPersonID":        pi.ID,
		"PiRemoteSiteLogin": piClusterUser.LocalUsername,
		"ResourceList":      getResourceList(body),
	}
	reply := map[string]any{"type": "notify_project_create", "body": replyBody}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_project_create: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_project_create: audit REPLY_SENT: %w", err)
	}
	return nil
}

func (h *RequestProjectCreateHandler) ensurePIUser(ctx context.Context, body map[string]any, globalID string) (*models.User, error) {
	if u, err := h.svc.GetUserByUserIdentity(ctx, amieIdentitySource, globalID); err == nil {
		return u, nil
	} else if !errors.Is(err, service.ErrNotFound) {
		return nil, err
	}
	org, err := ensureOrganization(ctx, h.svc, getString(body, "PiOrgCode"), getString(body, "PiOrganization"))
	if err != nil {
		return nil, fmt.Errorf("ensure PI organization: %w", err)
	}
	email := getString(body, "PiEmail")
	user, err := h.svc.CreateUser(ctx, &models.User{
		OrganizationID: org.ID,
		FirstName:      getString(body, "PiFirstName"),
		LastName:       getString(body, "PiLastName"),
		Email:          email,
		Status:         models.UserPending,
	})
	if err != nil {
		return nil, fmt.Errorf("create PI user: %w", err)
	}
	if _, err := h.svc.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     user.ID,
		Source:     amieIdentitySource,
		ExternalID: globalID,
		Email:      email,
	}); err != nil {
		return nil, fmt.Errorf("create PI user identity: %w", err)
	}
	return user, nil
}

func (h *RequestProjectCreateHandler) ensurePIClusterUser(ctx context.Context, userID string) (*models.ComputeClusterUser, error) {
	existing, err := h.svc.ListComputeClusterUsersByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list compute cluster users: %w", err)
	}
	for _, a := range existing {
		if a.ComputeClusterID == h.clusterID {
			return &a, nil
		}
	}
	return allocateAndCreateClusterUser(ctx, h.svc, h.clusterID, userID)
}

func (h *RequestProjectCreateHandler) ensureProject(ctx context.Context, originatedID, grantNumber, piID string) (*models.Project, error) {
	if p, err := h.svc.GetProjectByOriginatedID(ctx, originatedID); err == nil {
		return p, nil
	} else if !errors.Is(err, service.ErrNotFound) {
		return nil, err
	}
	return h.svc.CreateProject(ctx, &models.Project{
		OriginatedID: originatedID,
		Title:        grantNumber,
		Origination:  amieIdentitySource,
		ProjectPIID:  piID,
	})
}

// ensureAllocation creates a ComputeAllocation for the project on first
// delivery. On repeat delivery (supplement / renewal / extension / adjustment),
// the existing row is preserved and a ComputeAllocationDiff is recorded
// capturing the grant event. InitialSUAmount on the parent row stays as the
// original grant; effective SUs = InitialSUAmount + sum(grant diffs).
func (h *RequestProjectCreateHandler) ensureAllocation(ctx context.Context, body map[string]any, projectID, grantNumber string) (*models.ComputeAllocation, error) {
	existing, err := h.svc.ListComputeAllocationsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list allocations: %w", err)
	}
	if len(existing) > 0 {
		return h.recordAllocationDiff(ctx, body, &existing[0])
	}

	su, err := getInt64(body, "ServiceUnitsAllocated")
	if err != nil {
		return nil, err
	}
	start, err := getDate(body, "StartDate")
	if err != nil {
		return nil, err
	}
	end, err := getDate(body, "EndDate")
	if err != nil {
		return nil, err
	}

	return h.svc.CreateComputeAllocation(ctx, &models.ComputeAllocation{
		ProjectID:        projectID,
		Name:             grantNumber,
		ComputeClusterID: h.clusterID,
		InitialSUAmount:  su,
		StartTime:        start,
		EndTime:          end,
	})
}

// ensureResourceMappings grants the allocation each packet resource found in
// the cluster's catalog. AMIE expresses the grant only in service units,
// already stored on the allocation, so mappings carry no native caps;
// unregistered names are skipped rather than failing the packet.
func (h *RequestProjectCreateHandler) ensureResourceMappings(ctx context.Context, tx *sql.Tx, body map[string]any, allocation *models.ComputeAllocation, packet *model.Packet, eventID string) error {
	for _, name := range getResourceList(body) {
		resource, err := h.svc.GetComputeAllocationResourceByNameAndCluster(ctx, name, h.clusterID)
		if errors.Is(err, service.ErrNotFound) {
			slog.Warn("amie: packet resource not in the cluster catalog, mapping skipped",
				"resource", name, "allocation_id", allocation.ID)
			continue
		}
		if err != nil {
			return fmt.Errorf("lookup resource %q: %w", name, err)
		}
		mapping, err := h.svc.AttachResourceToAllocation(ctx, allocation.ID, resource.ID, 0, 0)
		if errors.Is(err, service.ErrAlreadyExists) {
			continue
		}
		if err != nil {
			return fmt.Errorf("attach resource %q: %w", name, err)
		}
		if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditAttachResource, "compute_allocation_resource_mapping", mapping.ID, name); err != nil {
			return fmt.Errorf("audit ATTACH_RESOURCE: %w", err)
		}
	}
	return nil
}

// ensurePIMembership asserts the PI as a role=PI membership on the allocation.
// Returns (membership, created=true) when it inserted a new row, or
// (existing, created=false) on redelivery. The service rejects a second
// distinct PI on the same allocation.
func (h *RequestProjectCreateHandler) ensurePIMembership(ctx context.Context, allocation *models.ComputeAllocation, piID string) (*models.ComputeAllocationMembership, bool, error) {
	existing, err := h.svc.ListMembersForAllocation(ctx, allocation.ID)
	if err != nil {
		return nil, false, fmt.Errorf("list memberships: %w", err)
	}
	for _, m := range existing {
		if m.UserID == piID {
			cm := m.ComputeAllocationMembership
			return &cm, false, nil
		}
	}
	created, err := h.svc.CreateComputeAllocationMembership(ctx, &models.ComputeAllocationMembership{
		ComputeAllocationID: allocation.ID,
		UserID:              piID,
		StartTime:           allocation.StartTime,
		EndTime:             allocation.EndTime,
		MembershipStatus:    models.ACTIVE,
	})
	if err != nil {
		return nil, false, err
	}
	return created, true, nil
}

// recordAllocationDiff writes a ComputeAllocationDiff for a re-delivered
// request_project_create against an existing allocation (the AMIE pattern for
// supplements / renewals / extensions / adjustments). DiffType is the upper-cased
// AllocationType from the packet body; falls back to "GRANT" when AMIE did not supply one.
// NewSUAmount carries this packet's ServiceUnitsAllocated, the delta granted
// by this event, not the cumulative total.
func (h *RequestProjectCreateHandler) recordAllocationDiff(ctx context.Context, body map[string]any, existing *models.ComputeAllocation) (*models.ComputeAllocation, error) {
	su, err := getInt64(body, "ServiceUnitsAllocated")
	if err != nil {
		return nil, err
	}
	diffType := strings.ToUpper(strings.TrimSpace(getString(body, "AllocationType")))
	if diffType == "" {
		diffType = "GRANT"
	}
	if _, err := h.svc.CreateComputeAllocationDiff(ctx, &models.ComputeAllocationDiff{
		ComputeAllocationID: existing.ID,
		DiffType:            diffType,
		NewSUAmount:         su,
		Status:              models.ACTIVE,
		Description:         fmt.Sprintf("AMIE %s of %d SUs", diffType, su),
	}); err != nil {
		return nil, fmt.Errorf("record allocation diff: %w", err)
	}
	return existing, nil
}
