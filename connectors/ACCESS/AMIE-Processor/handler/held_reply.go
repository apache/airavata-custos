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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	custosdb "github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// HeldReplies answers the packets waiting on a person once their account on the AMIE cluster changes.
type HeldReplies struct {
	db        *sqlx.DB
	svc       *service.Service
	clusterID string
	packets   store.PacketStore
	client    AmieClient
	auditSvc  AuditService
}

func NewHeldReplies(db *sqlx.DB, svc *service.Service, clusterID string, packets store.PacketStore, client AmieClient, auditSvc AuditService) *HeldReplies {
	return &HeldReplies{db: db, svc: svc, clusterID: clusterID, packets: packets, client: client, auditSvc: auditSvc}
}

// Subscribe listens for updates too, since a merge hands a person an account that is already reviewed.
func (h *HeldReplies) Subscribe(bus *events.Bus, subscriber string) {
	bus.SubscribeComputeClusterUserApproved(subscriber, h.answer)
	bus.SubscribeComputeClusterUserDenied(subscriber, h.answer)
	bus.SubscribeComputeClusterUserUpdated(subscriber, h.answer)
}

// answer answers each packet waiting on one of account's ACCESS persons. Each packet gets its own
// transaction, so a failed send leaves only that packet waiting for the bus to retry.
func (h *HeldReplies) answer(ctx context.Context, changed models.ComputeClusterUser) error {
	ctx, span := tracing.Start(ctx, "amie.held_reply")
	defer span.End()
	// The event's copy can be out of date, so the account's current state decides.
	account, err := h.svc.GetComputeClusterUser(ctx, changed.ID)
	if errors.Is(err, service.ErrNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("get cluster account: %w", err)
	}
	// Only a reviewed account on this connector's cluster decides a held packet.
	if account.ComputeClusterID != h.clusterID || account.ApprovalStatus == models.ComputeClusterUserPending {
		return nil
	}
	identities, err := h.svc.ListUserIdentitiesForUser(ctx, account.UserID)
	if err != nil {
		return fmt.Errorf("list user identities: %w", err)
	}
	for _, identity := range identities {
		if identity.Source != amieIdentitySource {
			continue
		}
		packets, err := h.packets.ListWaitingFor(ctx, identity.ExternalID)
		if err != nil {
			return fmt.Errorf("list held packets: %w", err)
		}
		for _, p := range packets {
			var packetJSON map[string]any
			if err := json.Unmarshal([]byte(p.RawJSON), &packetJSON); err != nil {
				return fmt.Errorf("decode packet %s: %w", p.ID, err)
			}
			body, err := getBody(packetJSON)
			if err != nil {
				return fmt.Errorf("packet %s: %w", p.ID, err)
			}
			if err := custosdb.TxFn(ctx, h.db, func(tx *sql.Tx) error {
				if err := answerAccountCreate(ctx, tx, h.client, h.auditSvc, &p, "", body, account); err != nil {
					return err
				}
				return h.packets.Update(ctx, tx, &p)
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
