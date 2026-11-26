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
	"fmt"

	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type RequestAccountCreateHandler struct {
	svc        *service.Service
	clusterID  string
	amieClient AmieClient
	auditSvc   AuditService
}

func NewRequestAccountCreateHandler(svc *service.Service, clusterID string, amieClient AmieClient, auditSvc AuditService) *RequestAccountCreateHandler {
	return &RequestAccountCreateHandler{svc: svc, clusterID: clusterID, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestAccountCreateHandler) SupportsType() string { return "request_account_create" }

// Handle finds the User by UserIdentity or email or creates it, looks up the Project (which
// must already exist from a prior request_project_create), provisions a
// ComputeClusterUser on the configured cluster, and attaches a
// ComputeAllocationMembership against the project's allocation. Replies with
// the assigned posix username.
func (h *RequestAccountCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) (err error) {
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
	projectID := getString(body, "ProjectID")
	if err := requireText(projectID, "ProjectID"); err != nil {
		return err
	}
	if err := requireText(getString(body, "GrantNumber"), "GrantNumber"); err != nil {
		return err
	}
	person := getPerson(body, "User")
	if err := requireText(person.GlobalID, "UserGlobalID"); err != nil {
		return err
	}
	if err := requireText(person.Email, "UserEmail"); err != nil {
		return err
	}
	if h.clusterID == "" {
		return fmt.Errorf("AMIE_CLUSTER_ID not configured")
	}

	user, err := bindPersonUser(ctx, h.svc, person)
	if err != nil {
		return fmt.Errorf("request_account_create: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreatePerson, "user", user.ID, ""); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_PERSON: %w", err)
	}

	// AMIE replies to notify_project_create with project.id (Custos UUID), so
	// subsequent packets carry that id back to us as body.ProjectID.
	project, err := h.svc.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("request_account_create: project %q not found (request_project_create must precede this packet): %w", projectID, err)
	}
	if user.ID == project.ProjectPIID {
		if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, model.TransactionFailureReply("user is already the project PI")); err != nil {
			return fmt.Errorf("request_account_create: sending failure reply: %w", err)
		}
		packet.Status = model.PacketStatusRefused
		return h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", string(packet.Status))
	}

	allocations, err := h.svc.ListComputeAllocationsByProject(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("request_account_create: list allocations: %w", err)
	}
	if len(allocations) == 0 {
		return fmt.Errorf("request_account_create: project %q has no ComputeAllocation; request_project_create did not provision one", projectID)
	}
	allocation := allocations[0]

	account, err := findOrCreateClusterUser(ctx, h.svc, h.clusterID, user.ID)
	if err != nil {
		return fmt.Errorf("request_account_create: find or create compute cluster user: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateAccount, "compute_cluster_user", account.ID, account.LocalUsername); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_ACCOUNT: %w", err)
	}

	role := normalizeAMIERole(normalizeRole(getString(body, "UserRole")))
	membership, err := findAllocationMembership(ctx, h.svc, allocation.ID, user.ID)
	if err != nil {
		return fmt.Errorf("request_account_create: find membership: %w", err)
	}
	switch {
	case membership == nil:
		if membership, err = createAllocationMembership(ctx, h.svc, &allocation, user.ID); err != nil {
			return fmt.Errorf("request_account_create: create membership: %w", err)
		}
	// The spec says a create packet for a known member activates it, unless its project is inactive.
	case allocation.Status == models.ACTIVE && membership.MembershipStatus != models.ACTIVE:
		if membership, err = h.svc.UpdateMembershipStatusTx(ctx, tx, membership.ID, models.ACTIVE); err != nil {
			return fmt.Errorf("request_account_create: activate membership: %w", err)
		}
	}
	if err := h.svc.SetProjectRole(ctx, project.ID, user.ID, role); err != nil {
		return fmt.Errorf("request_account_create: set project role: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateMembership, "compute_allocation_membership", membership.ID,
		fmt.Sprintf("allocation=%s user=%s role=%s", allocation.ID, user.ID, role)); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_MEMBERSHIP: %w", err)
	}

	replyBody := map[string]any{
		"ProjectID":           projectID,
		"GrantNumber":         getString(body, "GrantNumber"),
		"UserPersonID":        user.ID,
		"UserRemoteSiteLogin": account.LocalUsername,
		"ResourceList":        getResourceList(body),
	}
	if v := getString(body, "UserOrgCode"); v != "" {
		replyBody["UserOrgCode"] = v
	}
	reply := map[string]any{"type": "notify_account_create", "body": replyBody}
	if err := replyWhenAccountApproved(ctx, tx, h.amieClient, h.auditSvc, packet, eventID, account, reply); err != nil {
		return fmt.Errorf("request_account_create: %w", err)
	}
	return nil
}

// normalizeAMIERole maps the AMIE role enum to the persisted project role
// vocabulary. Unrecognized values collapse to MEMBER (which the project
// membership service interprets as "derived membership only").
func normalizeAMIERole(amieRole string) string {
	switch amieRole {
	case "PI", "CO_PI", "ALLOCATION_MANAGER", "MEMBER":
		return amieRole
	case "USER":
		return "MEMBER"
	default:
		return "MEMBER"
	}
}
