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
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	custosdb "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/db"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// replyWhenAccountApproved replies to ACCESS only if the cluster account is
// approved. Otherwise, the reply is held and sent once an admin approves it.
func replyWhenAccountApproved(ctx context.Context, tx *sql.Tx, client AmieClient, auditSvc AuditService, packet *model.Packet, eventID string, account *models.ComputeClusterUser, reply map[string]any) error {
	if account.ApprovalStatus == models.ClusterAccountApproved {
		if err := client.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
			return fmt.Errorf("sending reply: %w", err)
		}
		// A packet processed again can still carry an earlier held reply.
		packet.HeldReply, packet.HeldFor = nil, nil
		return auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", "")
	}
	raw, err := json.Marshal(reply)
	if err != nil {
		return fmt.Errorf("encode held reply: %w", err)
	}
	held := string(raw)
	packet.Status = model.PacketStatusWaitingApproval
	packet.HeldReply = &held
	packet.HeldFor = &account.ID
	return auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplyHeld, "compute_cluster_user", account.ID, account.LocalUsername)
}

// HeldReplies answers packets that are held for a reason, for example a
// cluster account approval.
type HeldReplies struct {
	db       *sqlx.DB
	packets  store.PacketStore
	client   AmieClient
	auditSvc AuditService
}

func NewHeldReplies(db *sqlx.DB, packets store.PacketStore, client AmieClient, auditSvc AuditService) *HeldReplies {
	return &HeldReplies{db: db, packets: packets, client: client, auditSvc: auditSvc}
}

func (h *HeldReplies) Subscribe(bus *events.Bus, subscriber string) {
	bus.SubscribeComputeClusterUserApproved(subscriber, h.approved)
	bus.SubscribeComputeClusterUserDenied(subscriber, h.denied)
}

func (h *HeldReplies) approved(ctx context.Context, cu models.ComputeClusterUser) error {
	return h.answer(ctx, cu.ID, model.PacketStatusProcessed, nil)
}

// denied ends the transaction on the ACCESS side with a failure that carries
// the admin's reason, which the protocol allows in reply to any packet.
func (h *HeldReplies) denied(ctx context.Context, cu models.ComputeClusterUser) error {
	reason := ""
	if cu.ReviewNote != nil {
		reason = *cu.ReviewNote
	}
	return h.answer(ctx, cu.ID, model.PacketStatusRefused, map[string]any{
		"type": "inform_transaction_complete",
		"body": map[string]any{"StatusCode": "Failure", "DetailCode": 2, "Message": reason},
	})
}

// answer sends each packet its held reply, or the failure when one is given.
// Each packet gets its own transaction, so a failed send leaves only that
// packet waiting for the bus to retry.
func (h *HeldReplies) answer(ctx context.Context, clusterUserID string, status model.PacketStatus, failure map[string]any) error {
	packets, err := h.packets.ListWaitingFor(ctx, clusterUserID)
	if err != nil {
		return fmt.Errorf("list held packets: %w", err)
	}
	for _, p := range packets {
		packetReply := failure
		// No failure means approval, so send the reply held on this packet.
		if packetReply == nil {
			if err := json.Unmarshal([]byte(*p.HeldReply), &packetReply); err != nil {
				return fmt.Errorf("decode held reply of packet %s: %w", p.ID, err)
			}
		}
		if err := custosdb.TxFn(ctx, h.db, func(tx *sql.Tx) error {
			if err := h.client.ReplyToPacket(ctx, p.AmieID, packetReply); err != nil {
				return fmt.Errorf("reply to packet %s: %w", p.ID, err)
			}
			p.Status = status
			if status == model.PacketStatusProcessed {
				now := time.Now().UTC()
				p.ProcessedAt = &now
			}
			p.HeldReply, p.HeldFor = nil, nil
			if err := h.packets.Update(ctx, tx, &p); err != nil {
				return err
			}
			return h.auditSvc.Log(ctx, tx, p.ID, "", model.AuditReplySent, "reply", "", string(status))
		}); err != nil {
			return err
		}
	}
	return nil
}
