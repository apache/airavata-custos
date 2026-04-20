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

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/allocations/access-amie/config"
	custosdb "github.com/apache/airavata-custos/allocations/access-amie/db"
	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

const (
	// MaxAttempts is the maximum number of processing attempts before an
	// event is marked as permanently failed.
	MaxAttempts = 3

	// BaseBackoffSeconds is the base delay in seconds for exponential
	// backoff retry scheduling.
	BaseBackoffSeconds = 30

	// MaxBackoffSeconds is the upper bound on retry delay in seconds.
	MaxBackoffSeconds = 600
)

type processorEventStore interface {
	FindTop50EventsToProcess(ctx context.Context, statuses []model.ProcessingStatus, now time.Time) ([]model.EventWithPacket, error)
	FindByID(ctx context.Context, id string) (*model.ProcessingEvent, error)
	Update(ctx context.Context, tx *sql.Tx, e *model.ProcessingEvent) error
}

type processorPacketStore interface {
	FindByID(ctx context.Context, id string) (*model.Packet, error)
	Update(ctx context.Context, tx *sql.Tx, p *model.Packet) error
}

type processorErrorStore interface {
	Save(ctx context.Context, tx *sql.Tx, e *model.ProcessingError) error
}

type processorRouter interface {
	Route(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error
}

type processorMetrics interface {
	RecordPacketProcessed(packetType, outcome string)
	RecordRetry()
	StartProcessingTimer() func(handlerType string)
}

type Processor struct {
	eventStore     processorEventStore
	packetStore    processorPacketStore
	errorStore     processorErrorStore
	router         processorRouter
	metrics        processorMetrics
	db             *sqlx.DB
	workerInterval time.Duration
}

func NewProcessor(eventStore processorEventStore, packetStore processorPacketStore, errorStore processorErrorStore, router processorRouter, metrics processorMetrics, db *sqlx.DB, cfg config.AMIEConfig) *Processor {
	return &Processor{
		eventStore:     eventStore,
		packetStore:    packetStore,
		errorStore:     errorStore,
		router:         router,
		metrics:        metrics,
		db:             db,
		workerInterval: cfg.WorkerInterval,
	}
}

func (p *Processor) Run(ctx context.Context) {
	slog.Info("AMIE processor started", "interval", p.workerInterval)
	ticker := time.NewTicker(p.workerInterval)
	defer ticker.Stop()

	p.processPendingEvents(ctx)
	for {
		select {
		case <-ctx.Done():
			slog.Info("AMIE processor stopping")
			return
		case <-ticker.C:
			p.processPendingEvents(ctx)
		}
	}
}

func (p *Processor) processPendingEvents(ctx context.Context) {
	events, err := p.eventStore.FindTop50EventsToProcess(
		ctx,
		[]model.ProcessingStatus{model.ProcessingStatusNew, model.ProcessingStatusRetryScheduled},
		time.Now(),
	)
	if err != nil {
		slog.Error("failed to query pending events", "error", err)
		return
	}
	if len(events) == 0 {
		return
	}

	slog.Info("processing pending events", "count", len(events))

	for _, ewp := range events {
		logger := slog.With(
			"eventId", ewp.ID,
			"packetId", ewp.PacketID,
			"amieId", ewp.PacketAmieID,
			"packetType", ewp.PacketType,
		)

		stopTimer := p.metrics.StartProcessingTimer()

		if err := p.executeInTransaction(ctx, ewp); err != nil {
			logger.Error("event processing failed", "error", err)
			if recordErr := p.recordFailureInNewTransaction(ctx, ewp.ID, err); recordErr != nil {
				logger.Error("failed to record processing failure", "error", recordErr)
			}
		}

		stopTimer(ewp.PacketType)
	}
}

// executeInTransaction processes a single event within a database transaction.
// If the handler returns an error, the entire transaction is rolled back
// (including the attempt increment), and the caller should record the failure
// in a separate transaction.
func (p *Processor) executeInTransaction(ctx context.Context, ewp model.EventWithPacket) error {
	return custosdb.TxFn(ctx, p.db, func(tx *sql.Tx) error {
		slog.Info("Processing event",
			"eventId", ewp.ID,
			"type", ewp.Type,
			"attempt", ewp.Attempts+1,
		)

		now := time.Now().UTC()
		ewp.Status = model.ProcessingStatusRunning
		ewp.StartedAt = &now
		ewp.Attempts++
		if err := p.eventStore.Update(ctx, tx, &ewp.ProcessingEvent); err != nil {
			return fmt.Errorf("update event to RUNNING: %w", err)
		}

		var packetJSON map[string]any
		if err := json.Unmarshal([]byte(ewp.PacketRawJSON), &packetJSON); err != nil {
			return fmt.Errorf("unmarshal packet raw JSON: %w", err)
		}

		// Load the full packet from DB to preserve all fields (e.g. retries).
		packet, err := p.packetStore.FindByID(ctx, ewp.PacketID)
		if err != nil {
			return fmt.Errorf("load packet %s: %w", ewp.PacketID, err)
		}

		if err := p.router.Route(ctx, tx, packetJSON, packet, ewp.ID); err != nil {
			return fmt.Errorf("route packet: %w", err)
		}

		finishedAt := time.Now().UTC()
		ewp.Status = model.ProcessingStatusSucceeded
		ewp.FinishedAt = &finishedAt
		ewp.NextRetryAt = nil
		if err := p.eventStore.Update(ctx, tx, &ewp.ProcessingEvent); err != nil {
			return fmt.Errorf("update event to SUCCEEDED: %w", err)
		}

		decodedAt := time.Now().UTC()
		packet.Status = model.PacketStatusDecoded
		packet.DecodedAt = &decodedAt
		if err := p.packetStore.Update(ctx, tx, packet); err != nil {
			return fmt.Errorf("update packet to DECODED: %w", err)
		}

		p.metrics.RecordPacketProcessed(ewp.PacketType, "succeeded")
		return nil
	})
}

// recordFailureInNewTransaction records a processing failure in a separate
// transaction from the one that was rolled back. This ensures the failure is
// persisted even though the processing transaction was rolled back. It reads
// the event afresh (with the pre-rollback attempt count) and schedules a retry
// or marks the event as permanently failed.
func (p *Processor) recordFailureInNewTransaction(ctx context.Context, eventID string, cause error) error {
	return custosdb.TxFn(ctx, p.db, func(tx *sql.Tx) error {
		// Re-read the event -- the processing transaction was rolled back,
		// so the attempt count is still the pre-increment value.
		event, err := p.eventStore.FindByID(ctx, eventID)
		if err != nil {
			return fmt.Errorf("find event for failure recording: %w", err)
		}
		if event == nil {
			slog.Warn("event not found for failure recording", "eventId", eventID)
			return nil
		}

		packet, err := p.packetStore.FindByID(ctx, event.PacketID)
		if err != nil {
			return fmt.Errorf("find packet for failure recording: %w", err)
		}

		// The processing transaction rolled back, so event.Attempts is the
		// old value. Compute the effective attempt count.
		effectiveAttempts := event.Attempts + 1
		isRetryable := effectiveAttempts < MaxAttempts

		event.Attempts = effectiveAttempts
		errMsg := cause.Error()
		event.LastError = &errMsg
		now := time.Now().UTC()
		event.FinishedAt = &now

		if isRetryable {
			event.Status = model.ProcessingStatusRetryScheduled
			nextRetry := ComputeNextRetryAt(effectiveAttempts)
			event.NextRetryAt = &nextRetry

			packet.Retries = effectiveAttempts
			packet.LastError = &errMsg
			if err := p.packetStore.Update(ctx, tx, packet); err != nil {
				return fmt.Errorf("update packet retries: %w", err)
			}

			p.metrics.RecordRetry()
			p.metrics.RecordPacketProcessed(packet.Type, "retry_scheduled")
			slog.Warn("event failed, scheduling retry",
				"eventId", eventID,
				"attempt", effectiveAttempts,
				"nextRetryAt", nextRetry,
			)
		} else {
			event.Status = model.ProcessingStatusPermanentlyFailed
			event.NextRetryAt = nil
			p.metrics.RecordPacketProcessed(packet.Type, "permanently_failed")
			slog.Error("event permanently failed after max attempts", "eventId", eventID)

			packet.Status = model.PacketStatusFailed
			packet.Retries = effectiveAttempts
			packet.LastError = &errMsg
			if err := p.packetStore.Update(ctx, tx, packet); err != nil {
				return fmt.Errorf("update packet status: %w", err)
			}
		}

		if err := p.eventStore.Update(ctx, tx, event); err != nil {
			return fmt.Errorf("update event: %w", err)
		}

		detail := cause.Error()
		if len(detail) > 8000 {
			detail = detail[:8000]
		}
		summary := fmt.Sprintf("%s", cause.Error())
		pErr := &model.ProcessingError{
			PacketID:   &event.PacketID,
			EventID:    &event.ID,
			OccurredAt: now,
			Summary:    summary,
			Detail:     &detail,
		}
		return p.errorStore.Save(ctx, tx, pErr)
	})
}

// ComputeNextRetryAt calculates the next retry time using exponential backoff.
// The delay is BaseBackoffSeconds * 2^(attempt-1), capped at MaxBackoffSeconds.
func ComputeNextRetryAt(attempt int) time.Time {
	exp := attempt - 1
	if exp < 0 {
		exp = 0
	}
	delaySec := BaseBackoffSeconds * (1 << exp) // 30 * 2^exp
	if delaySec > MaxBackoffSeconds {
		delaySec = MaxBackoffSeconds
	}
	return time.Now().UTC().Add(time.Duration(delaySec) * time.Second)
}
