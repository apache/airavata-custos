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

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
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

// Handle ensures the User (with ExternalIdentity), looks up the Project (which
// must already exist from a prior request_project_create), provisions a
// ClusterAccount on the configured cluster, and attaches a
// ComputeAllocationMembership against the project's allocation. Replies with
// the assigned posix username.
func (h *RequestAccountCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
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
	userGlobalID := getString(body, "UserGlobalID")
	if err := requireText(userGlobalID, "UserGlobalID"); err != nil {
		return err
	}
	if h.clusterID == "" {
		return fmt.Errorf("AMIE_CLUSTER_ID not configured")
	}

	user, err := h.ensureUser(ctx, body, userGlobalID)
	if err != nil {
		return fmt.Errorf("request_account_create: ensure user: %w", err)
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

	allocations, err := h.svc.ListComputeAllocationsByProject(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("request_account_create: list allocations: %w", err)
	}
	if len(allocations) == 0 {
		return fmt.Errorf("request_account_create: project %q has no ComputeAllocation; request_project_create did not provision one", projectID)
	}
	allocation := allocations[0]

	account, err := h.ensureClusterAccount(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("request_account_create: ensure cluster account: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateAccount, "cluster_account", account.ID, account.Username); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_ACCOUNT: %w", err)
	}

	role := normalizeRole(getString(body, "UserRole"))
	membership, err := h.ensureMembership(ctx, allocation.ID, user.ID)
	if err != nil {
		return fmt.Errorf("request_account_create: ensure membership: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateMembership, "compute_allocation_membership", membership.ID,
		fmt.Sprintf("allocation=%s user=%s role=%s", allocation.ID, user.ID, role)); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_MEMBERSHIP: %w", err)
	}

	replyBody := map[string]any{
		"ProjectID":           projectID,
		"GrantNumber":         getString(body, "GrantNumber"),
		"UserPersonID":        user.ID,
		"UserRemoteSiteLogin": account.Username,
		"ResourceList":        getResourceList(body),
	}
	if v := getString(body, "UserOrgCode"); v != "" {
		replyBody["UserOrgCode"] = v
	}
	reply := map[string]any{"type": "notify_account_create", "body": replyBody}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_account_create: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_account_create: audit REPLY_SENT: %w", err)
	}
	return nil
}

func (h *RequestAccountCreateHandler) ensureUser(ctx context.Context, body map[string]any, globalID string) (*models.User, error) {
	if u, err := h.svc.GetUserByExternalIdentity(ctx, amieIdentitySource, globalID); err == nil {
		return u, nil
	} else if !errors.Is(err, service.ErrNotFound) {
		return nil, err
	}

	org, err := ensureOrganization(ctx, h.svc, getString(body, "UserOrgCode"), getString(body, "UserOrganization"))
	if err != nil {
		return nil, fmt.Errorf("ensure user organization: %w", err)
	}
	user, err := h.svc.CreateUser(ctx, &models.User{
		OrganizationID: org.ID,
		FirstName:      getString(body, "UserFirstName"),
		LastName:       getString(body, "UserLastName"),
		Email:          getString(body, "UserEmail"),
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	if _, err := h.svc.CreateExternalIdentity(ctx, &models.ExternalIdentity{
		UserID:     user.ID,
		Source:     amieIdentitySource,
		ExternalID: globalID,
	}); err != nil {
		return nil, fmt.Errorf("create external identity: %w", err)
	}
	return user, nil
}

// ensureClusterAccount returns the user's existing cluster account on the
// configured cluster, or provisions a fresh one with a temp posix username.
func (h *RequestAccountCreateHandler) ensureClusterAccount(ctx context.Context, userID string) (*models.ClusterAccount, error) {
	existing, err := h.svc.ListClusterAccountsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list cluster accounts: %w", err)
	}
	for _, a := range existing {
		if a.ComputeClusterID == h.clusterID {
			return &a, nil
		}
	}
	return h.svc.CreateClusterAccount(ctx, &models.ClusterAccount{
		UserID:           userID,
		ComputeClusterID: h.clusterID,
		Username:         generateTempPosixUsername(),
	})
}

// ensureMembership returns the existing (allocation, user) membership or
// creates a new one. Idempotent for re-delivered packets.
func (h *RequestAccountCreateHandler) ensureMembership(ctx context.Context, allocationID, userID string) (*models.ComputeAllocationMembership, error) {
	existing, err := h.svc.ListMembersForAllocation(ctx, allocationID)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	for _, m := range existing {
		if m.UserID == userID {
			return &m, nil
		}
	}
	return h.svc.CreateComputeAllocationMembership(ctx, &models.ComputeAllocationMembership{
		ComputeAllocationID: allocationID,
		UserID:              userID,
	})
}
