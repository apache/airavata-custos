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
	// Packet resource names that are not catalog names, mapped to the catalog
	// resource they stand for in the custos.yaml
	resourceAliases map[string][]string
}

func NewRequestProjectCreateHandler(svc *service.Service, clusterID string, amieClient AmieClient, auditSvc AuditService, resourceAliases map[string][]string) *RequestProjectCreateHandler {
	return &RequestProjectCreateHandler{svc: svc, clusterID: clusterID, amieClient: amieClient, auditSvc: auditSvc, resourceAliases: resourceAliases}
}

func (h *RequestProjectCreateHandler) SupportsType() string { return "request_project_create" }

// Handle finds or creates the PI user (resolving the organization from PiOrgCode +
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
	person := getPerson(body, "Pi")
	if err := requireText(person.GlobalID, "PiGlobalID"); err != nil {
		return err
	}
	if err := requireText(person.Email, "PiEmail"); err != nil {
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

	pi, match, err := findPersonUser(ctx, h.svc, person)
	if err != nil {
		return fmt.Errorf("request_project_create: find PI: %w", err)
	}
	if match != personBound {
		org, err := findOrCreateOrganization(ctx, h.svc, person.OrgCode, person.OrgName)
		if err != nil {
			return fmt.Errorf("request_project_create: find or create PI organization: %w", err)
		}
		if match == personNew {
			pi, err = createPersonUser(ctx, h.svc, org.ID, person)
		} else {
			err = changeUserOrganizationAndName(ctx, h.svc, pi, org.ID, person)
		}
		if err != nil {
			return fmt.Errorf("request_project_create: PI: %w", err)
		}
		if err := createPersonUserIdentity(ctx, h.svc, pi.ID, person); err != nil {
			return fmt.Errorf("request_project_create: PI: %w", err)
		}
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreatePerson, "user", pi.ID, "PI"); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_PERSON: %w", err)
	}

	piClusterUser, err := findOrCreateClusterUser(ctx, h.svc, h.clusterID, pi.ID)
	if err != nil {
		return fmt.Errorf("request_project_create: find or create PI cluster user: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateAccount, "compute_cluster_user", piClusterUser.ID, piClusterUser.LocalUsername); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_ACCOUNT (PI): %w", err)
	}

	title := getString(body, "ProjectTitle")
	if title == "" {
		title = grantNumber
	}
	project, err := h.svc.GetProjectByOriginatedID(ctx, projectOriginatedID)
	switch {
	case errors.Is(err, service.ErrNotFound):
		project, err = h.svc.CreateProject(ctx, &models.Project{
			OriginatedID: projectOriginatedID,
			Title:        title,
			Origination:  amieIdentitySource,
			ProjectPIID:  pi.ID,
		})
		if err != nil {
			return fmt.Errorf("request_project_create: create project: %w", err)
		}
	case err != nil:
		return fmt.Errorf("request_project_create: find project: %w", err)
	// The spec says a project packet for an inactive project reactivates it.
	case project.Status != models.ProjectActive:
		if project, err = h.svc.UpdateProjectStatusTx(ctx, tx, project.ID, models.ProjectActive); err != nil {
			return fmt.Errorf("request_project_create: activate project: %w", err)
		}
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateProject, "project", project.ID, ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_PROJECT: %w", err)
	}

	su, err := getNumber(body, "ServiceUnitsAllocated")
	if err != nil {
		return fmt.Errorf("request_project_create: %w", err)
	}
	allocations, err := h.svc.ListComputeAllocationsByProject(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("request_project_create: list allocations: %w", err)
	}
	var allocation *models.ComputeAllocation
	// A grant has one allocation, active or not. This handler never makes a second.
	if len(allocations) == 0 {
		if allocation, err = h.createAllocation(ctx, body, project.ID, grantNumber, su); err != nil {
			return fmt.Errorf("request_project_create: create allocation: %w", err)
		}
	} else if allocation, err = h.changeAllocation(ctx, tx, body, &allocations[0], su); err != nil {
		return fmt.Errorf("request_project_create: change allocation: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditCreateAllocation, "compute_allocation", allocation.ID, ""); err != nil {
		return fmt.Errorf("request_project_create: audit CREATE_ALLOCATION: %w", err)
	}

	if err := h.attachResources(ctx, tx, body, allocation, packet, eventID); err != nil {
		return fmt.Errorf("request_project_create: attach resources: %w", err)
	}

	piMembership, err := findAllocationMembership(ctx, h.svc, allocation.ID, pi.ID)
	if err != nil {
		return fmt.Errorf("request_project_create: find PI membership: %w", err)
	}
	created := piMembership == nil
	if created {
		// The service rejects a second distinct PI on the same allocation.
		if piMembership, err = createAllocationMembership(ctx, h.svc, allocation, pi.ID); err != nil {
			return fmt.Errorf("request_project_create: create PI membership: %w", err)
		}
	} else if piMembership.MembershipStatus != models.ACTIVE {
		if piMembership, err = h.svc.UpdateMembershipStatusTx(ctx, tx, piMembership.ID, models.ACTIVE); err != nil {
			return fmt.Errorf("request_project_create: activate PI membership: %w", err)
		}
	}
	if err := h.svc.SetProjectRole(ctx, project.ID, pi.ID, "PI"); err != nil {
		return fmt.Errorf("request_project_create: set PI project role: %w", err)
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

	// NOTE - Not held for the PI's cluster account approval. This reply gives ACCESS the
	// project ID, and holding it would block the whole project there. ACCESS gets
	// the PI's login early, but the account is only created after admin approval.
	reply := map[string]any{"type": "notify_project_create", "body": replyBody}
	return sendReply(ctx, tx, h.amieClient, h.auditSvc, packet, eventID, "request_project_create", reply)
}

// createAllocation creates the project's allocation from the packet.
func (h *RequestProjectCreateHandler) createAllocation(ctx context.Context, body map[string]any, projectID, grantNumber string, su float64) (*models.ComputeAllocation, error) {
	if su < 0 {
		return nil, fmt.Errorf("%w: negative service units for a project with no allocation", model.ErrInvalidPacket)
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

// attachResources attaches to the allocation each packet resource found in
// the cluster's catalog. AMIE expresses the grant only in service units,
// already stored on the allocation, so mappings carry no native caps;
// unregistered names are skipped rather than failing the packet.
func (h *RequestProjectCreateHandler) attachResources(ctx context.Context, tx *sql.Tx, body map[string]any, allocation *models.ComputeAllocation, packet *model.Packet, eventID string) error {
	for _, packetName := range getResourceList(body) {
		// One packet resource can stand for several catalog resources, such as two GPU partitions with their own rates.
		names, ok := h.resourceAliases[packetName]
		if !ok {
			names = []string{packetName}
		}
		for _, name := range names {
			if err := h.attachResource(ctx, tx, allocation, packet, eventID, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *RequestProjectCreateHandler) attachResource(ctx context.Context, tx *sql.Tx, allocation *models.ComputeAllocation, packet *model.Packet, eventID, name string) error {
	resource, err := h.svc.GetComputeAllocationResourceByNameAndCluster(ctx, name, h.clusterID)
	if errors.Is(err, service.ErrNotFound) {
		slog.WarnContext(ctx, "amie: packet resource not in the cluster catalog, mapping skipped",
			"resource", name, "allocation_id", allocation.ID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("lookup resource %q: %w", name, err)
	}
	mapping, err := h.svc.AttachResourceToAllocation(ctx, allocation.ID, resource.ID, 0, 0)
	if errors.Is(err, service.ErrAlreadyExists) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("attach resource %q: %w", name, err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditAttachResource, "compute_allocation_resource_mapping", mapping.ID, name); err != nil {
		return fmt.Errorf("audit ATTACH_RESOURCE: %w", err)
	}
	return nil
}

// changeAllocation applies a repeat packet by its AllocationType and keeps a
// diff row as history. See "Allocation types" in ../README.md.
func (h *RequestProjectCreateHandler) changeAllocation(ctx context.Context, tx *sql.Tx, body map[string]any, alloc *models.ComputeAllocation, su float64) (*models.ComputeAllocation, error) {
	diffType := strings.ToUpper(strings.TrimSpace(getString(body, "AllocationType")))
	if diffType == "" {
		diffType = "GRANT"
	}
	switch diffType {
	case "RENEWAL":
		start, err := getDate(body, "StartDate")
		if err != nil {
			return nil, err
		}
		end, err := getDate(body, "EndDate")
		if err != nil {
			return nil, err
		}
		alloc.InitialSUAmount, alloc.StartTime, alloc.EndTime = su, start, end
	case "EXTENSION":
		end, err := getDate(body, "EndDate")
		if err != nil {
			return nil, err
		}
		alloc.EndTime = end
	default:
		// NEW, TRANSFER, SUPPLEMENT, ADJUSTMENT, ADVANCE. The amount is signed, so a negative transfer deducts.
		alloc.InitialSUAmount += su
	}
	alloc.Status = models.ACTIVE
	if err := h.svc.UpdateComputeAllocationTx(ctx, tx, alloc); err != nil {
		return nil, fmt.Errorf("update allocation: %w", err)
	}
	if _, err := h.svc.CreateComputeAllocationDiffTx(ctx, tx, &models.ComputeAllocationDiff{
		ComputeAllocationID: alloc.ID,
		DiffType:            diffType,
		NewSUAmount:         su,
		Status:              models.ACTIVE,
		Description:         fmt.Sprintf("AMIE %s of %g SUs", diffType, su),
	}); err != nil {
		return nil, fmt.Errorf("record allocation diff: %w", err)
	}
	return alloc, nil
}
