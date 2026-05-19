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

type RequestProjectCreateHandler struct {
	svc          *service.Service
	defaultOrgID string
	amieClient   AmieClient
	auditSvc     AuditService
}

func NewRequestProjectCreateHandler(svc *service.Service, defaultOrgID string, amieClient AmieClient, auditSvc AuditService) *RequestProjectCreateHandler {
	return &RequestProjectCreateHandler{svc: svc, defaultOrgID: defaultOrgID, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestProjectCreateHandler) SupportsType() string { return "request_project_create" }

func (h *RequestProjectCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
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
	// AMIE protocol: request_project_create does not carry a ProjectID. The
	// receiving site assigns one. We use the GrantNumber as the originated_id
	// since it is the stable cross-site identifier on the AMIE side.
	projectOriginatedID := grantNumber

	pi, err := h.ensurePIUser(ctx, body, piGlobalID)
	if err != nil {
		return fmt.Errorf("request_project_create: ensure PI: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreatePerson, "user", pi.ID, "PI"); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_PERSON: %w", err)
	}

	project, err := h.ensureProject(ctx, projectOriginatedID, grantNumber, pi.ID)
	if err != nil {
		return fmt.Errorf("request_project_create: ensure project: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateProject, "project", project.ID, ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_PROJECT: %w", err)
	}

	replyBody := map[string]any{
		"ProjectID":         project.ID,
		"GrantNumber":       grantNumber,
		"PiPersonID":        pi.ID,
		"PiRemoteSiteLogin": piGlobalID,
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
	if u, err := h.svc.GetUserByExternalIdentity(ctx, amieIdentitySource, globalID); err == nil {
		return u, nil
	} else if !errors.Is(err, service.ErrNotFound) {
		return nil, err
	}
	if h.defaultOrgID == "" {
		return nil, fmt.Errorf("cannot create PI user: AMIE_DEFAULT_ORG_ID not configured")
	}
	user, err := h.svc.CreateUser(ctx, &models.User{
		OrganizationID: h.defaultOrgID,
		FirstName:      getString(body, "PiFirstName"),
		LastName:       getString(body, "PiLastName"),
		Email:          getString(body, "PiEmail"),
	})
	if err != nil {
		return nil, fmt.Errorf("create PI user: %w", err)
	}
	if _, err := h.svc.CreateExternalIdentity(ctx, &models.ExternalIdentity{
		UserID:     user.ID,
		Source:     amieIdentitySource,
		ExternalID: globalID,
	}); err != nil {
		return nil, fmt.Errorf("create PI external identity: %w", err)
	}
	return user, nil
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
