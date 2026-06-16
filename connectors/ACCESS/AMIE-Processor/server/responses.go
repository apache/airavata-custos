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
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
)

// packetSource is the fixed source label until multi-source ingestion lands.
const packetSource = "access"

// PacketResponse is the wire shape for an AMIE packet.
type PacketResponse struct {
	ID          string  `json:"id"`
	AmieID      string  `json:"amie_id"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Source      string  `json:"source"`
	RawJSON     *string `json:"raw_json,omitempty"`
	ReceivedAt  string  `json:"received_at"`
	UpdatedAt   string  `json:"updated_at"`
	DecodedAt   *string `json:"decoded_at,omitempty"`
	ProcessedAt *string `json:"processed_at,omitempty"`
	Retries     int     `json:"retries"`
	LastError   *string `json:"last_error,omitempty"`
}

// PacketListResponse is the paginated list envelope for packets.
type PacketListResponse struct {
	Packets []PacketResponse `json:"packets"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// PacketEventResponse is the wire shape for a packet processing event.
type PacketEventResponse struct {
	ID         string  `json:"id"`
	PacketID   string  `json:"packet_id"`
	EventType  string  `json:"event_type"`
	Actor      string  `json:"actor"`
	Status     string  `json:"status"`
	Message    *string `json:"message,omitempty"`
	Timestamp  string  `json:"timestamp"`
	DurationMs *int64  `json:"duration_ms,omitempty"`
}

// PacketStatBucketResponse is a single (date, status, type) cell in the stats grid.
type PacketStatBucketResponse struct {
	Date   string `json:"date"`
	Status string `json:"status"`
	Type   string `json:"type"`
	Count  int64  `json:"count"`
}

// PacketStatsResponse is the per-day packet stats payload.
type PacketStatsResponse struct {
	ByDay []PacketStatBucketResponse `json:"byDay"`
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
		EventType: mapEventType(string(e.Type), string(e.Status)),
		Actor:     "amie-worker",
		Status:    mapEventStatus(string(e.Status)),
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

// mapEventType folds amie_processing_events.type into the portal's enum.
func mapEventType(t, status string) string {
	if strings.EqualFold(status, "FAILED") {
		return "FAILED"
	}
	switch strings.ToLower(t) {
	case "received":
		return "RECEIVED"
	case "decoded":
		return "DECODED"
	case "handled", "processed":
		return "HANDLED"
	case "retry_scheduled":
		return "RETRY_SCHEDULED"
	case "retry":
		return "RETRY"
	case "manual_resolve":
		return "MANUAL_RESOLVE"
	case "manual_link":
		return "MANUAL_LINK"
	}
	return "HANDLED"
}

func mapEventStatus(s string) string {
	switch strings.ToUpper(s) {
	case "FAILED", "ERROR":
		return "FAILED"
	case "RUNNING", "STARTED", "IN_PROGRESS":
		return "RUNNING"
	}
	return "SUCCEEDED"
}
