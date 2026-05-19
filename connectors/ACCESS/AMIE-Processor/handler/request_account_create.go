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
	svc          *service.Service
	defaultOrgID string
	amieClient   AmieClient
	auditSvc     AuditService
}

func NewRequestAccountCreateHandler(svc *service.Service, defaultOrgID string, amieClient AmieClient, auditSvc AuditService) *RequestAccountCreateHandler {
	return &RequestAccountCreateHandler{svc: svc, defaultOrgID: defaultOrgID, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestAccountCreateHandler) SupportsType() string { return "request_account_create" }

// Handle is partial. It ensures the User (with ExternalIdentity) and confirms
// the Project exists in core, then audits the membership request. Two
// operations are still TODO:
//   - ClusterAccount provisioning — username-generation policy
//   - ComputeAllocationMembership creation
func (h *RequestAccountCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}
	projectOriginatedID := getString(body, "ProjectID")
	if err := requireText(projectOriginatedID, "ProjectID"); err != nil {
		return err
	}
	if err := requireText(getString(body, "GrantNumber"), "GrantNumber"); err != nil {
		return err
	}
	userGlobalID := getString(body, "UserGlobalID")
	if err := requireText(userGlobalID, "UserGlobalID"); err != nil {
		return err
	}

	user, err := h.ensureUser(ctx, body, userGlobalID)
	if err != nil {
		return fmt.Errorf("request_account_create: ensure user: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreatePerson, "user", user.ID, ""); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_PERSON: %w", err)
	}

	if _, err := h.svc.GetProjectByOriginatedID(ctx, projectOriginatedID); err != nil {
		if !errors.Is(err, service.ErrNotFound) {
			return fmt.Errorf("request_account_create: lookup project: %w", err)
		}
		// TODO(amie-integration): create the Project here when AMIE carries enough
		// metadata to do so safely (title, PI, grant number).
		// request_account_create assumes the project was created earlier via
		// request_project_create.
	}

	// TODO(amie-integration): provision a ClusterAccount via svc.CreateClusterAccount
	// once a username-generation policy is in place.

	// TODO(amie-integration): create a ComputeAllocationMembership via
	// svc.CreateComputeAllocationMembership once we have a ComputeAllocation
	// under this Project to attach to.
	role := normalizeRole(getString(body, "UserRole"))
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateMembership, "membership_request", "", fmt.Sprintf("project=%s user=%s role=%s (membership persistence pending allocation mapping)", projectOriginatedID, user.ID, role)); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_MEMBERSHIP: %w", err)
	}

	replyBody := map[string]any{
		"ProjectID":           projectOriginatedID,
		"GrantNumber":         getString(body, "GrantNumber"),
		"UserPersonID":        user.ID,
		"UserRemoteSiteLogin": getString(body, "UserGlobalID"),
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

	if h.defaultOrgID == "" {
		return nil, fmt.Errorf("cannot create user: AMIE_DEFAULT_ORG_ID not configured")
	}
	user, err := h.svc.CreateUser(ctx, &models.User{
		OrganizationID: h.defaultOrgID,
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
