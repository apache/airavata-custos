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
	dmodel "github.com/apache/airavata-custos/allocations/domain/model"
)

type requestProjectCreatePersonService interface {
	FindOrCreateFromPacket(ctx context.Context, tx *sql.Tx, body map[string]any) (*dmodel.Person, error)
}

type requestProjectCreateAccountService interface {
	ProvisionClusterAccount(ctx context.Context, tx *sql.Tx, person *dmodel.Person) (*dmodel.ClusterAccount, error)
}

type requestProjectCreateProjectService interface {
	CreateOrFindProject(ctx context.Context, tx *sql.Tx, projectID, grantNumber string) (*dmodel.Project, error)
}

type requestProjectCreateMembershipService interface {
	CreateMembership(ctx context.Context, tx *sql.Tx, projectID, clusterAccountID, role string) (*dmodel.ProjectMembership, error)
}

type requestProjectCreateAmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

type requestProjectCreateAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type RequestProjectCreateHandler struct {
	personSvc     requestProjectCreatePersonService
	accountSvc    requestProjectCreateAccountService
	projectSvc    requestProjectCreateProjectService
	membershipSvc requestProjectCreateMembershipService
	amieClient    requestProjectCreateAmieClient
	auditSvc      requestProjectCreateAuditService
}

func NewRequestProjectCreateHandler(
	personSvc requestProjectCreatePersonService,
	accountSvc requestProjectCreateAccountService,
	projectSvc requestProjectCreateProjectService,
	membershipSvc requestProjectCreateMembershipService,
	amieClient requestProjectCreateAmieClient,
	auditSvc requestProjectCreateAuditService,
) *RequestProjectCreateHandler {
	return &RequestProjectCreateHandler{
		personSvc:     personSvc,
		accountSvc:    accountSvc,
		projectSvc:    projectSvc,
		membershipSvc: membershipSvc,
		amieClient:    amieClient,
		auditSvc:      auditSvc,
	}
}

func (h *RequestProjectCreateHandler) SupportsType() string {
	return "request_project_create"
}

func (h *RequestProjectCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	// Validate required fields.
	grantNumber := getString(body, "GrantNumber")
	if err := requireText(grantNumber, "GrantNumber"); err != nil {
		return err
	}
	piGlobalID := getString(body, "PiGlobalID")
	if err := requireText(piGlobalID, "PiGlobalID"); err != nil {
		return err
	}
	piFirstName := getString(body, "PiFirstName")
	if err := requireText(piFirstName, "PiFirstName"); err != nil {
		return err
	}
	piLastName := getString(body, "PiLastName")
	if err := requireText(piLastName, "PiLastName"); err != nil {
		return err
	}

	// Map PI fields to user fields so FindOrCreateFromPacket can process them.
	piAsUserBody := map[string]any{
		"UserGlobalID":     body["PiGlobalID"],
		"UserFirstName":    body["PiFirstName"],
		"UserLastName":     body["PiLastName"],
		"UserEmail":        body["PiEmail"],
		"UserOrganization": body["PiOrganization"],
		"UserOrgCode":      body["PiOrgCode"],
		"NsfStatusCode":    body["NsfStatusCode"],
		"UserDnList":       body["PiDnList"],
	}

	// Create or find the PI person.
	piPerson, err := h.personSvc.FindOrCreateFromPacket(ctx, tx, piAsUserBody)
	if err != nil {
		return fmt.Errorf("request_project_create: finding/creating PI person: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreatePerson, "person", piPerson.ID, ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_PERSON: %w", err)
	}

	// Provision a cluster account for the PI.
	clusterAccount, err := h.accountSvc.ProvisionClusterAccount(ctx, tx, piPerson)
	if err != nil {
		return fmt.Errorf("request_project_create: provisioning cluster account: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateAccount, "account", clusterAccount.ID, ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_ACCOUNT: %w", err)
	}

	// Create or find the project.
	localProjectID := "PRJ-" + grantNumber
	project, err := h.projectSvc.CreateOrFindProject(ctx, tx, localProjectID, grantNumber)
	if err != nil {
		return fmt.Errorf("request_project_create: creating/finding project: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateProject, "project", project.ID, ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_PROJECT: %w", err)
	}

	// Create PI membership.
	if _, err := h.membershipSvc.CreateMembership(ctx, tx, project.ID, clusterAccount.ID, "PI"); err != nil {
		return fmt.Errorf("request_project_create: creating PI membership: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateMembership, "membership", "", ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_MEMBERSHIP: %w", err)
	}

	// Build and send the reply.
	replyBody := map[string]any{
		"ProjectID":         project.ID,
		"GrantNumber":       grantNumber,
		"PiPersonID":        piPerson.ID,
		"PiRemoteSiteLogin": clusterAccount.Username,
		"PiGlobalID":        piGlobalID,
		"ResourceList":      getResourceList(body),
	}
	projectTitle := getString(body, "ProjectTitle")
	if projectTitle != "" {
		replyBody["ProjectTitle"] = projectTitle
	}

	reply := map[string]any{
		"type": "notify_project_create",
		"body": replyBody,
	}

	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_project_create: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_project_create: audit REPLY_SENT: %w", err)
	}

	return nil
}
