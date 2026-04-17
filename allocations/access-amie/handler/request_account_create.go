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

	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

type requestAccountCreatePersonService interface {
	FindOrCreateFromPacket(ctx context.Context, tx *sql.Tx, body map[string]any) (*model.Person, error)
}

type requestAccountCreateAccountService interface {
	ProvisionClusterAccount(ctx context.Context, tx *sql.Tx, person *model.Person) (*model.ClusterAccount, error)
}

type requestAccountCreateProjectService interface {
	CreateOrFindProject(ctx context.Context, tx *sql.Tx, projectID, grantNumber string) (*model.Project, error)
}

type requestAccountCreateMembershipService interface {
	CreateMembership(ctx context.Context, tx *sql.Tx, projectID, clusterAccountID, role string) (*model.ProjectMembership, error)
}

type requestAccountCreateAmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

type requestAccountCreateAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type RequestAccountCreateHandler struct {
	personSvc     requestAccountCreatePersonService
	accountSvc    requestAccountCreateAccountService
	projectSvc    requestAccountCreateProjectService
	membershipSvc requestAccountCreateMembershipService
	amieClient    requestAccountCreateAmieClient
	auditSvc      requestAccountCreateAuditService
}

func NewRequestAccountCreateHandler(
	personSvc requestAccountCreatePersonService,
	accountSvc requestAccountCreateAccountService,
	projectSvc requestAccountCreateProjectService,
	membershipSvc requestAccountCreateMembershipService,
	amieClient requestAccountCreateAmieClient,
	auditSvc requestAccountCreateAuditService,
) *RequestAccountCreateHandler {
	return &RequestAccountCreateHandler{
		personSvc:     personSvc,
		accountSvc:    accountSvc,
		projectSvc:    projectSvc,
		membershipSvc: membershipSvc,
		amieClient:    amieClient,
		auditSvc:      auditSvc,
	}
}

func (h *RequestAccountCreateHandler) SupportsType() string {
	return "request_account_create"
}

func (h *RequestAccountCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	// Validate required fields.
	projectID := getString(body, "ProjectID")
	if err := requireText(projectID, "ProjectID"); err != nil {
		return err
	}
	grantNumber := getString(body, "GrantNumber")
	if err := requireText(grantNumber, "GrantNumber"); err != nil {
		return err
	}
	userGlobalID := getString(body, "UserGlobalID")
	if err := requireText(userGlobalID, "UserGlobalID"); err != nil {
		return err
	}

	// Create or find the person.
	person, err := h.personSvc.FindOrCreateFromPacket(ctx, tx, body)
	if err != nil {
		return fmt.Errorf("request_account_create: finding/creating person: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreatePerson, "person", person.ID, ""); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_PERSON: %w", err)
	}

	// Provision a cluster account.
	account, err := h.accountSvc.ProvisionClusterAccount(ctx, tx, person)
	if err != nil {
		return fmt.Errorf("request_account_create: provisioning cluster account: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateAccount, "account", account.ID, ""); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_ACCOUNT: %w", err)
	}

	// Ensure the project exists.
	if _, err := h.projectSvc.CreateOrFindProject(ctx, tx, projectID, grantNumber); err != nil {
		return fmt.Errorf("request_account_create: creating/finding project: %w", err)
	}

	// Create USER membership.
	if _, err := h.membershipSvc.CreateMembership(ctx, tx, projectID, account.ID, "USER"); err != nil {
		return fmt.Errorf("request_account_create: creating USER membership: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateMembership, "membership", "", ""); err != nil {
		return fmt.Errorf("request_account_create: audit CREATE_MEMBERSHIP: %w", err)
	}

	// Build and send the reply.
	replyBody := map[string]any{
		"ProjectID":           projectID,
		"GrantNumber":         grantNumber,
		"UserPersonID":        person.ID,
		"UserRemoteSiteLogin": account.Username,
		"ResourceList":        getResourceList(body),
	}
	userOrgCode := getString(body, "UserOrgCode")
	if userOrgCode != "" {
		replyBody["UserOrgCode"] = userOrgCode
	}

	reply := map[string]any{
		"type": "notify_account_create",
		"body": replyBody,
	}

	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_account_create: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_account_create: audit REPLY_SENT: %w", err)
	}

	return nil
}
