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

package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/allocations/access-amie/config"
	custosdb "github.com/apache/airavata-custos/allocations/access-amie/db"
	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

type pollerPacketStore interface {
	FindByAmieID(ctx context.Context, amieID int64) (*model.Packet, error)
	Save(ctx context.Context, tx *sql.Tx, p *model.Packet) error
}

type pollerEventStore interface {
	Save(ctx context.Context, tx *sql.Tx, e *model.ProcessingEvent) error
}

type pollerAmieClient interface {
	FetchInProgressPackets(ctx context.Context) ([]map[string]any, error)
}

type pollerMetrics interface {
	RecordPacketReceived(packetType string)
	RecordPollerFetch(count int)
}

type Poller struct {
	client       pollerAmieClient
	packetStore  pollerPacketStore
	eventStore   pollerEventStore
	metrics      pollerMetrics
	db           *sqlx.DB
	pollInterval time.Duration
	enabled      bool
}

func NewPoller(client pollerAmieClient, packetStore pollerPacketStore, eventStore pollerEventStore, metrics pollerMetrics, db *sqlx.DB, cfg config.AMIEConfig) *Poller {
	return &Poller{
		client:       client,
		packetStore:  packetStore,
		eventStore:   eventStore,
		metrics:      metrics,
		db:           db,
		pollInterval: cfg.PollInterval,
		enabled:      cfg.PollerEnabled,
	}
}

// Run starts the polling loop. It blocks until the context is cancelled. If
// the poller is disabled via configuration, it returns immediately.
func (p *Poller) Run(ctx context.Context) {
	if !p.enabled {
		slog.Info("AMIE poller disabled")
		return
	}
	slog.Info("AMIE poller started", "interval", p.pollInterval)
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	// Poll immediately on start, then on ticker.
	p.pollForPackets(ctx)
	for {
		select {
		case <-ctx.Done():
			slog.Info("AMIE poller stopping")
			return
		case <-ticker.C:
			p.pollForPackets(ctx)
		}
	}
}

func (p *Poller) pollForPackets(ctx context.Context) {
	slog.Info("Polling for new AMIE packets...")
	packets, err := p.client.FetchInProgressPackets(ctx)
	if err != nil {
		slog.Error("failed to fetch AMIE packets", "error", err)
		return
	}
	if len(packets) == 0 {
		slog.Debug("no new AMIE packets found")
		return
	}
	slog.Info("fetched AMIE packets", "count", len(packets))
	p.metrics.RecordPollerFetch(len(packets))

	for _, packetNode := range packets {
		p.processIndividualPacket(ctx, packetNode)
	}
}

func (p *Poller) processIndividualPacket(ctx context.Context, packetNode map[string]any) {
	// Extract amie_id from header.packet_rec_id.
	amiePacketRecID := int64(-1)
	if header, ok := packetNode["header"].(map[string]any); ok {
		if v, ok := header["packet_rec_id"]; ok {
			switch n := v.(type) {
			case float64:
				amiePacketRecID = int64(n)
			case int64:
				amiePacketRecID = n
			}
		}
	}
	packetType := ""
	if t, ok := packetNode["type"].(string); ok {
		packetType = t
	}

	if amiePacketRecID < 0 || packetType == "" {
		slog.Warn("skipping packet with missing header fields", "raw", packetNode)
		return
	}

	logger := slog.With("amieId", amiePacketRecID, "packetType", packetType)

	// Deduplication check.
	existing, err := p.packetStore.FindByAmieID(ctx, amiePacketRecID)
	if err != nil {
		logger.Error("failed to check for existing packet", "error", err)
		return
	}
	if existing != nil {
		logger.Debug("packet already exists, skipping")
		return
	}

	// Persist in a transaction.
	err = custosdb.TxFn(ctx, p.db, func(tx *sql.Tx) error {
		// Create packet.
		rawJSON, _ := json.Marshal(packetNode)
		now := time.Now().UTC()
		newPacket := &model.Packet{
			ID:         uuid.NewString(),
			AmieID:     amiePacketRecID,
			Type:       packetType,
			Status:     model.PacketStatusNew,
			RawJSON:    string(rawJSON),
			ReceivedAt: now,
		}
		if err := p.packetStore.Save(ctx, tx, newPacket); err != nil {
			return fmt.Errorf("save packet: %w", err)
		}

		// Create processing event.
		eventID := uuid.NewString()
		payload, err := CreateDecodeStartedEvent(eventID, newPacket.ID, amiePacketRecID)
		if err != nil {
			return fmt.Errorf("create protobuf payload: %w", err)
		}

		event := &model.ProcessingEvent{
			ID:        eventID,
			PacketID:  newPacket.ID,
			Type:      model.EventTypeDecodePacket,
			Status:    model.ProcessingStatusNew,
			Payload:   payload,
			CreatedAt: now,
		}
		if err := p.eventStore.Save(ctx, tx, event); err != nil {
			return fmt.Errorf("save processing event: %w", err)
		}

		logger.Info("persisted new AMIE packet", "packetId", newPacket.ID)
		return nil
	})

	if err != nil {
		logger.Error("failed to persist packet", "error", err)
		return
	}

	p.metrics.RecordPacketReceived(packetType)
}
