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

package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

func (o *Orchestrator) ensurePOSIXAccountImpl(ctx context.Context, cu *models.ComputeClusterUser) error {
	log := slog.With("correlation_id", cu.ID, "custos_user_id", cu.UserID, "unix_cluster_id", o.c.Config().UnixClusterID)

	user, err := o.core.GetUser(ctx, cu.UserID)
	if err != nil {
		return fmt.Errorf("get custos user: %w", err)
	}

	personID, composite, created, err := o.lookupOrCreateCoPerson(ctx, user)
	if err != nil {
		o.dlq(ctx, cu, "lookup_or_create_coperson", err)
		return err
	}
	log = log.With("comanage_person_id", personID)
	log.Info("comanage: CoPerson resolved", "created", created)
	if created {
		o.audit(ctx, cu, "ComanageCoPersonCreated", fmt.Sprintf("comanage_id=%s email=%s", personID, user.Email))
	}

	if err := o.storePersonID(ctx, cu.UserID, personID); err != nil {
		log.Warn("comanage: failed to store CoPerson id", "err", err)
	}

	if composite == nil {
		composite, err = o.getPersonComposite(ctx, personID)
		if err != nil {
			o.dlq(ctx, cu, "get_composite", err)
			return err
		}
	}

	uidnumber, err := extractIdentifier(composite, "uidnumber")
	if err != nil {
		o.dlq(ctx, cu, "extract_uidnumber", err)
		return err
	}
	if uidnumber == "" {
		err := fmt.Errorf("uidnumber identifier missing on CoPerson %s", personID)
		o.dlq(ctx, cu, "missing_uidnumber", err)
		return err
	}
	uidInt, err := strconv.ParseInt(uidnumber, 10, 64)
	if err != nil {
		o.dlq(ctx, cu, "parse_uidnumber", err)
		return err
	}

	coPersonID, err := extractCoPersonID(composite)
	if err != nil || coPersonID == 0 {
		o.dlq(ctx, cu, "extract_coperson_id", err)
		return fmt.Errorf("extract CoPerson.meta.id: %w", err)
	}

	coGroupID, err := o.findOrCreateCoGroup(ctx, cu.LocalUsername)
	if err != nil {
		return err
	}
	if err := o.findOrCreateIdentifier(ctx, cu, coGroupID, cu.LocalUsername, "uid"); err != nil {
		return err
	}
	if err := o.findOrCreateIdentifier(ctx, cu, coGroupID, uidnumber, "gidnumber"); err != nil {
		return err
	}
	if err := o.findOrCreateCoGroupMember(ctx, cu, coGroupID, coPersonID); err != nil {
		return err
	}
	if err := o.createUnixClusterGroup(ctx, cu, coGroupID, log); err != nil {
		return err
	}

	fresh, err := o.getPersonComposite(ctx, personID)
	if err != nil {
		o.dlq(ctx, cu, "refetch_composite", err)
		return err
	}
	block := UnixClusterAccountBlock{
		UnixClusterId:    o.c.Config().UnixClusterID,
		SyncMode:         "M",
		Status:           "A",
		Username:         cu.LocalUsername,
		Uid:              uidInt,
		Gecos:            "",
		LoginShell:       o.c.Config().DefaultShell,
		HomeDirectory:    o.c.Config().HomedirPrefix + cu.LocalUsername,
		PrimaryCoGroupId: coGroupID,
	}
	putBody, err := mergeUnixClusterAccount(fresh, block)
	if err != nil {
		o.dlq(ctx, cu, "merge_composite", err)
		return err
	}
	if err := o.updatePerson(ctx, personID, putBody); err != nil {
		o.dlq(ctx, cu, "put_composite", err)
		return err
	}
	log.Info("comanage: UnixClusterAccount attached", "username", cu.LocalUsername, "uid", uidInt, "co_group_id", coGroupID)
	o.audit(ctx, cu, "ComanageClusterAccountAttached", fmt.Sprintf("comanage_id=%s username=%s uid=%d", personID, cu.LocalUsername, uidInt))
	return nil
}

func (o *Orchestrator) getPersonComposite(ctx context.Context, personID string) (json.RawMessage, error) {
	_, span := tracing.Start(ctx, "comanage.get_person_composite")
	defer span.End()
	composite, err := o.c.GetPersonComposite(personID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return composite, nil
}

func (o *Orchestrator) findOrCreateCoGroup(ctx context.Context, name string) (int, error) {
	_, span := tracing.Start(ctx, "comanage.find_or_create_co_group")
	defer span.End()
	id, err := o.c.FindCoGroupByName(name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	if id == 0 {
		id, err = o.c.CreateCoGroup(name, "Primary group for "+name)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return 0, err
		}
	}
	span.SetAttributes(attribute.Int("comanage.co_group_id", id))
	return id, nil
}

func (o *Orchestrator) findOrCreateIdentifier(ctx context.Context, cu *models.ComputeClusterUser, coGroupID int, value, identifierType string) error {
	_, span := tracing.Start(ctx, "comanage.find_or_create_identifier")
	defer span.End()
	span.SetAttributes(
		attribute.Int("comanage.co_group_id", coGroupID),
		attribute.String("comanage.identifier_type", identifierType),
	)
	existing, err := o.c.FindIdentifierOnGroup(coGroupID, identifierType)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		o.dlq(ctx, cu, "find_"+identifierType+"_identifier", err)
		return err
	}
	if existing != 0 {
		return nil
	}
	if _, err := o.c.CreateIdentifierOnGroup(value, identifierType, coGroupID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		o.dlq(ctx, cu, "create_"+identifierType+"_identifier", err)
		return err
	}
	return nil
}

func (o *Orchestrator) findOrCreateCoGroupMember(ctx context.Context, cu *models.ComputeClusterUser, coGroupID, coPersonID int) error {
	_, span := tracing.Start(ctx, "comanage.find_or_create_co_group_member")
	defer span.End()
	span.SetAttributes(
		attribute.Int("comanage.co_group_id", coGroupID),
		attribute.Int("comanage.co_person_id", coPersonID),
	)
	existing, err := o.c.FindCoGroupMember(coGroupID, coPersonID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		o.dlq(ctx, cu, "find_co_group_member", err)
		return err
	}
	if existing != 0 {
		return nil
	}
	if _, err := o.c.CreateCoGroupMember(coPersonID, coGroupID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		o.dlq(ctx, cu, "create_co_group_member", err)
		return err
	}
	return nil
}

// createUnixClusterGroup re-attempts the bind; 4xx is treated as
// already-attached and swallowed (idempotency mechanism).
func (o *Orchestrator) createUnixClusterGroup(ctx context.Context, cu *models.ComputeClusterUser, coGroupID int, log *slog.Logger) error {
	_, span := tracing.Start(ctx, "comanage.create_unix_cluster_group")
	defer span.End()
	span.SetAttributes(attribute.Int("comanage.co_group_id", coGroupID))
	if _, err := o.c.CreateUnixClusterGroup(coGroupID); err != nil {
		var httpErr *client.HTTPError
		if !errors.As(err, &httpErr) || httpErr.StatusCode < 400 || httpErr.StatusCode >= 500 {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			o.dlq(ctx, cu, "create_unix_cluster_group", err)
			return err
		}
		log.Info("comanage: UnixClusterGroup attach returned 4xx (already attached)", "status", httpErr.StatusCode)
	}
	return nil
}

func (o *Orchestrator) updatePerson(ctx context.Context, personID string, body []byte) error {
	_, span := tracing.Start(ctx, "comanage.update_person")
	defer span.End()
	if _, err := o.c.UpdatePerson(personID, body); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

// lookupOrCreateCoPerson resolves the user's CoPerson, returning the COmanage
// person identifier, the composite (if the GET path was used), and whether a
// new CoPerson was created.
func (o *Orchestrator) lookupOrCreateCoPerson(ctx context.Context, user *models.User) (string, json.RawMessage, bool, error) {
	ctx, span := tracing.Start(ctx, "comanage.lookup_or_create_co_person")
	defer span.End()
	personIDType := o.c.Config().PersonIDType

	if stored, err := o.findStoredPersonID(ctx, user.ID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", nil, false, fmt.Errorf("stored lookup: %w", err)
	} else if stored != "" {
		composite, err := o.c.GetPersonComposite(stored)
		if err == nil {
			span.SetAttributes(attribute.String("comanage.person_id", stored))
			return stored, composite, false, nil
		}
		if !errors.Is(err, client.ErrNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return "", nil, false, fmt.Errorf("get composite for stored id: %w", err)
		}
		// stored id no longer resolves; fall through to email search
	}

	if user.Email != "" {
		coPersonID, err := o.findByEmailExact(user.Email)
		if err != nil && !errors.Is(err, client.ErrNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return "", nil, false, fmt.Errorf("email search: %w", err)
		}
		if coPersonID != 0 {
			composite, err := o.c.GetPersonComposite(strconv.Itoa(coPersonID))
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return "", nil, false, fmt.Errorf("get composite by numeric id: %w", err)
			}
			personID, err := extractIdentifier(composite, personIDType)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return "", nil, false, fmt.Errorf("extract %s from composite: %w", personIDType, err)
			}
			if personID == "" {
				return "", nil, false, fmt.Errorf("CoPerson %d has no %s identifier", coPersonID, personIDType)
			}
			span.SetAttributes(attribute.String("comanage.person_id", personID))
			return personID, composite, false, nil
		}
	}

	body, err := buildCreatePersonBody(o.c.Config().COID, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", nil, false, fmt.Errorf("build create body: %w", err)
	}
	resp, err := o.c.CreatePerson(body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", nil, false, fmt.Errorf("create coperson: %w", err)
	}
	for _, r := range resp {
		if r.Type == personIDType {
			span.SetAttributes(
				attribute.String("comanage.person_id", r.Identifier),
				attribute.Bool("comanage.created", true),
			)
			return r.Identifier, nil, true, nil
		}
	}
	return "", nil, false, fmt.Errorf("POST /people returned no %s identifier: %+v", personIDType, resp)
}

func buildCreatePersonBody(coID int, user *models.User) ([]byte, error) {
	body := map[string]interface{}{
		"CoPerson": map[string]interface{}{
			"co_id":  coID,
			"status": "A",
		},
		"Name": []map[string]interface{}{{
			"given":        user.FirstName,
			"family":       user.LastName,
			"type":         "official",
			"primary_name": true,
			"language":     "en",
		}},
		"EmailAddress": []map[string]interface{}{{
			"mail":     user.Email,
			"type":     "official",
			"verified": false,
		}},
	}
	return json.Marshal(body)
}

func (o *Orchestrator) audit(ctx context.Context, cu *models.ComputeClusterUser, eventType, details string) {
	_, _ = o.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType: eventType,
		EntityID:  cu.ID,
		Details:   details,
	})
}

func (o *Orchestrator) dlq(ctx context.Context, cu *models.ComputeClusterUser, step string, err error) {
	ctx, span := tracing.Start(ctx, "comanage.dlq")
	defer span.End()
	span.SetAttributes(
		attribute.String("comanage.cluster_user_id", cu.ID),
		attribute.String("comanage.step", step),
	)
	details := fmt.Sprintf("step=%s err=%v", step, err)
	_, _ = o.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType: "ComanageProvisioningFailed",
		EntityID:  cu.ID,
		Details:   details,
	})
	// TODO: admin endpoint + CLI to re-fire ComputeClusterUserCreateEvent for a
	// specific user, and to clean up orphan CoGroups from terminal dead-letters.
}
