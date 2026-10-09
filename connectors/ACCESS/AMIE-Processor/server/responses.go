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

package server

import (
	"strconv"
	"time"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/pkg/models"
)

// packetSource is the fixed source label until multi-source ingestion lands.
const packetSource = "access"

// PacketResponse is the wire shape for an AMIE packet.
type PacketResponse struct {
	ID          string  `json:"id" binding:"required"`
	AmieID      string  `json:"amie_id" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	Status      string  `json:"status" binding:"required" enums:"NEW,DECODED,PROCESSED,FAILED,WAITING_APPROVAL,REFUSED"`
	Source      string  `json:"source" binding:"required"`
	RawJSON     *string `json:"raw_json,omitempty"`
	ReceivedAt  string  `json:"received_at" binding:"required"`
	UpdatedAt   string  `json:"updated_at" binding:"required"`
	DecodedAt   *string `json:"decoded_at,omitempty"`
	ProcessedAt *string `json:"processed_at,omitempty"`
	Retries     int     `json:"retries" binding:"required"`
	LastError   *string `json:"last_error,omitempty"`
}

// PacketListResponse is the paginated list envelope for packets.
type PacketListResponse struct {
	Packets []PacketResponse `json:"packets" binding:"required"`
	Total   int              `json:"total" binding:"required"`
	Limit   int              `json:"limit" binding:"required"`
	Offset  int              `json:"offset" binding:"required"`
}

// PacketEventResponse is the wire shape for a packet processing event.
type PacketEventResponse struct {
	ID         string  `json:"id" binding:"required"`
	PacketID   string  `json:"packet_id" binding:"required"`
	EventType  string  `json:"event_type" binding:"required" enums:"RECEIVED,DECODED,HANDLED,FAILED,RETRY,RETRY_SCHEDULED,MANUAL_RESOLVE,MANUAL_LINK"`
	Actor      string  `json:"actor" binding:"required"`
	Status     string  `json:"status" binding:"required" enums:"SUCCEEDED,RUNNING,FAILED"`
	Message    *string `json:"message,omitempty"`
	Timestamp  string  `json:"timestamp" binding:"required"`
	DurationMs *int64  `json:"duration_ms,omitempty"`
}

// PacketAuditsResponse is the audit trail of one packet.
type PacketAuditsResponse struct {
	PacketID string              `json:"packet_id" binding:"required"`
	Events   []models.TraceEvent `json:"events" binding:"required"`
}

// PacketStatBucketResponse is a single (date, status, type) cell in the stats grid.
type PacketStatBucketResponse struct {
	Date   string `json:"date" binding:"required"`
	Status string `json:"status" binding:"required" enums:"NEW,DECODED,PROCESSED,FAILED,WAITING_APPROVAL,REFUSED"`
	Type   string `json:"type" binding:"required"`
	Count  int64  `json:"count" binding:"required"`
}

// PacketStatsResponse is the per-day packet stats payload.
type PacketStatsResponse struct {
	ByDay []PacketStatBucketResponse `json:"byDay" binding:"required"`
}

// packetResponseFrom maps a stored Packet to its wire shape.
func packetResponseFrom(p model.Packet) PacketResponse {
	out := PacketResponse{
		ID:         p.ID,
		AmieID:     strconv.FormatInt(p.AmieID, 10),
		Type:       p.Type,
		Status:     string(p.Status),
		Source:     packetSource,
		ReceivedAt: p.ReceivedAt.UTC().Format(time.RFC3339Nano),
		Retries:    p.Retries,
		LastError:  p.LastError,
	}
	if p.RawJSON != "" {
		v := p.RawJSON
		out.RawJSON = &v
	}
	if p.DecodedAt != nil {
		v := p.DecodedAt.UTC().Format(time.RFC3339Nano)
		out.DecodedAt = &v
	}
	if p.ProcessedAt != nil {
		v := p.ProcessedAt.UTC().Format(time.RFC3339Nano)
		out.ProcessedAt = &v
	}
	out.UpdatedAt = mostRecent(p.ReceivedAt, p.DecodedAt, p.ProcessedAt).UTC().Format(time.RFC3339Nano)
	return out
}

// packetEventResponseFrom maps a stored ProcessingEvent to its wire shape.
func packetEventResponseFrom(e model.ProcessingEvent) PacketEventResponse {
	ts := e.CreatedAt
	if e.StartedAt != nil {
		ts = *e.StartedAt
	}
	out := PacketEventResponse{
		ID:        e.ID,
		PacketID:  e.PacketID,
		EventType: mapEventType(e.Status),
		Actor:     "amie-worker",
		Status:    mapEventStatus(e.Status),
		Timestamp: ts.UTC().Format(time.RFC3339Nano),
		Message:   e.LastError,
	}
	if e.StartedAt != nil && e.FinishedAt != nil {
		d := e.FinishedAt.Sub(*e.StartedAt).Milliseconds()
		if d >= 0 {
			out.DurationMs = &d
		}
	}
	return out
}

// packetStatsResponseFrom wraps neutral stat buckets in the portal envelope.
func packetStatsResponseFrom(buckets []store.StatBucket) PacketStatsResponse {
	out := PacketStatsResponse{ByDay: make([]PacketStatBucketResponse, 0, len(buckets))}
	for _, b := range buckets {
		out.ByDay = append(out.ByDay, PacketStatBucketResponse{
			Date:   b.Date,
			Status: b.Status,
			Type:   b.Type,
			Count:  b.Count,
		})
	}
	return out
}

// mostRecent returns the latest of received and any provided optional times.
func mostRecent(received time.Time, others ...*time.Time) time.Time {
	latest := received
	for _, o := range others {
		if o != nil && o.After(latest) {
			latest = *o
		}
	}
	return latest
}

// mapEventType folds a processing event into the portal's enum; DECODE_PACKET is the only type.
func mapEventType(s model.ProcessingStatus) string {
	if mapEventStatus(s) == "FAILED" {
		return "FAILED"
	}
	return "HANDLED"
}

func mapEventStatus(s model.ProcessingStatus) string {
	switch s {
	case model.ProcessingStatusFailed, model.ProcessingStatusPermanentlyFailed:
		return "FAILED"
	case model.ProcessingStatusNew, model.ProcessingStatusRunning, model.ProcessingStatusRetryScheduled:
		return "RUNNING"
	}
	return "SUCCEEDED"
}
