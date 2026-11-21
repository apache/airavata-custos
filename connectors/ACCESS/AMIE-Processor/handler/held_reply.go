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
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
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

// Subscribe answers on reviews, and on user updates, which a merge publishes for the survivor.
func (h *HeldReplies) Subscribe(bus *events.Bus, subscriber string) {
	bus.SubscribeComputeClusterUserApproved(subscriber, func(ctx context.Context, cu models.ComputeClusterUser) error { return h.answer(ctx, cu.UserID) })
	bus.SubscribeComputeClusterUserDenied(subscriber, func(ctx context.Context, cu models.ComputeClusterUser) error { return h.answer(ctx, cu.UserID) })
	bus.SubscribeUserUpdated(subscriber, func(ctx context.Context, u models.User) error { return h.answer(ctx, u.ID) })
}

// answer answers each packet waiting on one of the user's ACCESS persons, as their account on this
// connector's cluster decides. Each packet gets its own transaction, so a failed send leaves only
// that packet waiting for the bus to retry.
func (h *HeldReplies) answer(ctx context.Context, userID string) error {
	ctx, span := tracing.Start(ctx, "amie.held_reply")
	defer span.End()
	account, err := h.svc.GetComputeClusterUserByPair(ctx, h.clusterID, userID)
	if errors.Is(err, service.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get cluster account: %w", err)
	}
	if account.ApprovalStatus == models.ClusterAccountPending {
		return nil
	}
	identities, err := h.svc.ListUserIdentitiesForUser(ctx, userID)
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

// answerAccountCreate replies to a request_account_create packet as account's review decided.
// Before a review, the packet waits on the ACCESS person, which a merge keeps.
func answerAccountCreate(ctx context.Context, tx *sql.Tx, client AmieClient, auditSvc AuditService, packet *model.Packet, eventID string, body map[string]any, account *models.ComputeClusterUser) error {
	var reply map[string]any
	switch account.ApprovalStatus {
	case models.ClusterAccountApproved:
		now := time.Now().UTC()
		replyBody := map[string]any{
			"ProjectID":           getString(body, "ProjectID"),
			"GrantNumber":         getString(body, "GrantNumber"),
			"UserPersonID":        account.UserID,
			"UserRemoteSiteLogin": account.LocalUsername,
			"ResourceList":        getResourceList(body),
		}
		if v := getString(body, "UserOrgCode"); v != "" {
			replyBody["UserOrgCode"] = v
		}
		reply, packet.Status, packet.ProcessedAt = map[string]any{"type": "notify_account_create", "body": replyBody}, model.PacketStatusProcessed, &now
	case models.ClusterAccountDenied:
		note := ""
		if account.ReviewNote != nil {
			note = *account.ReviewNote
		}
		reply, packet.Status = model.TransactionFailureReply(note), model.PacketStatusRefused
	default:
		personID := getString(body, "UserGlobalID")
		packet.Status, packet.HeldFor = model.PacketStatusWaitingApproval, &personID
		return auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplyHeld, "compute_cluster_user", account.ID, account.LocalUsername)
	}
	if err := client.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("reply to packet %s: %w", packet.ID, err)
	}
	packet.HeldFor = nil
	return auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", string(packet.Status))
}
