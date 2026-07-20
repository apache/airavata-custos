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
	"net/http"
	"strconv"
	"time"

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
	if err := o.findOrCreateUnixClusterGroup(ctx, cu, coGroupID, log); err != nil {
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
	if err := o.core.MarkComputeClusterUserProvisioned(ctx, cu.ID); err != nil {
		log.Warn("comanage: failed to mark cluster user provisioned", "err", err)
	}
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

// findOrCreateUnixClusterGroup binds the CoGroup to the configured UnixCluster
// idempotently. The UnixCluster plugin returns 500 on duplicate POSTs, so we
// GET first; on a 500 from POST we re-GET to recover from races.
func (o *Orchestrator) findOrCreateUnixClusterGroup(ctx context.Context, cu *models.ComputeClusterUser, coGroupID int, log *slog.Logger) error {
	ctx, span := tracing.Start(ctx, "comanage.find_or_create_unix_cluster_group")
	defer span.End()
	span.SetAttributes(attribute.Int("comanage.co_group_id", coGroupID))

	if existing, err := o.c.FindUnixClusterGroup(coGroupID); err == nil && existing != 0 {
		span.SetAttributes(attribute.Int("comanage.unix_cluster_group_id", existing))
		return nil
	} else if err != nil {
		log.Debug("comanage: FindUnixClusterGroup failed; attempting POST", "err", err)
	}

	if _, err := o.c.CreateUnixClusterGroup(coGroupID); err != nil {
		var httpErr *client.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			log.Info("comanage: UnixClusterGroup attach returned 4xx (already attached)", "status", httpErr.StatusCode)
			return nil
		}
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusInternalServerError {
			if existing, gerr := o.c.FindUnixClusterGroup(coGroupID); gerr == nil && existing != 0 {
				log.Info("comanage: UnixClusterGroup POST returned 500 but binding exists", "binding_id", existing)
				span.SetAttributes(attribute.Int("comanage.unix_cluster_group_id", existing))
				return nil
			}
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		o.dlq(ctx, cu, "create_unix_cluster_group", err)
		return err
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
			// COmanage's /people/{identifier} endpoint expects an Identifier
			// *value* (e.g. "Custos100022"), not a numeric CoPerson ID. Resolve
			// the identifier value by listing the CoPerson's Identifiers first.
			personID, err := o.c.FindIdentifierValueOnPerson(coPersonID, personIDType)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return "", nil, false, fmt.Errorf("find %s on CoPerson %d: %w", personIDType, coPersonID, err)
			}
			if personID == "" {
				// CoPerson exists but the identifier-assignment plugin hasn't
				// run yet; fall through to POST /people (idempotent against
				// repeat creation since COmanage dedups on the CoPersonRole).
			} else {
				composite, err := o.c.GetPersonComposite(personID)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return "", nil, false, fmt.Errorf("get composite by %s %s: %w", personIDType, personID, err)
				}
				span.SetAttributes(attribute.String("comanage.person_id", personID))
				return personID, composite, false, nil
			}
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
	// COmanage's identifier-assignment plugin runs asynchronously, so
	// POST /people sometimes returns an empty identifier
	// list even though the CoPerson was created. Refetch by email with a short
	// backoff and pull the identifier from the eventually consistent composite.
	if user.Email != "" {
		personID, composite, ok := o.waitForCreatedPersonID(user.Email, personIDType)
		if ok {
			span.SetAttributes(
				attribute.String("comanage.person_id", personID),
				attribute.Bool("comanage.created", true),
				attribute.Bool("comanage.eventually_consistent", true),
			)
			return personID, composite, true, nil
		}
	}
	return "", nil, false, fmt.Errorf("POST /people returned no %s identifier: %+v", personIDType, resp)
}

// waitForCreatedPersonID polls findByEmailExact then the Identifiers list a
// few times, since COmanage's identifier-assignment plugin runs asynchronously
// after POST /people. Returns the assigned person_id once present.
func (o *Orchestrator) waitForCreatedPersonID(email, personIDType string) (string, json.RawMessage, bool) {
	delays := []time.Duration{300 * time.Millisecond, 700 * time.Millisecond, 1500 * time.Millisecond, 3 * time.Second}
	for _, d := range delays {
		time.Sleep(d)
		coPersonID, err := o.findByEmailExact(email)
		if err != nil || coPersonID == 0 {
			continue
		}
		personID, err := o.c.FindIdentifierValueOnPerson(coPersonID, personIDType)
		if err != nil || personID == "" {
			continue
		}
		composite, err := o.c.GetPersonComposite(personID)
		if err != nil {
			continue
		}
		return personID, composite, true
	}
	return "", nil, false
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
		EventType:  eventType,
		EntityID:   cu.ID,
		EntityType: "compute_cluster_user",
		Details:    details,
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
		EventType:  "ComanageProvisioningFailed",
		EntityID:   cu.ID,
		EntityType: "compute_cluster_user",
		Details:    details,
	})
	// TODO: admin endpoint + CLI to re-fire ComputeClusterUserCreateEvent for a
	// specific user, and to clean up orphan CoGroups from terminal dead-letters.
}
