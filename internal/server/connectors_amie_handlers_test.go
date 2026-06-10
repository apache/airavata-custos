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
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

type fakeAmiePacketAuditStore struct {
	events  []models.TraceEvent
	listArg string
	listErr error
}

func (f *fakeAmiePacketAuditStore) ListAuditsForPacket(_ context.Context, packetID string) ([]models.TraceEvent, error) {
	f.listArg = packetID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.events, nil
}

func TestListAmiePacketAudits_HappyPath(t *testing.T) {
	span := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	fs := &fakeAmiePacketAuditStore{events: []models.TraceEvent{{
		SpanID:    span,
		Source:    "amie",
		EventType: "CREATE_ACCOUNT",
		Status:    "ok",
		CreatedAt: time.Unix(1700000000, 0).UTC(),
	}}}
	srv := newAuditServer(t, &AdminDeps{AmiePacketAudits: fs})

	resp, err := http.Get(srv.URL + "/connectors/amie/packets/packet-1/audits")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	var body struct {
		PacketID string           `json:"packet_id"`
		Events   []map[string]any `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.PacketID != "packet-1" {
		t.Errorf("packet_id = %q", body.PacketID)
	}
	if len(body.Events) != 1 {
		t.Fatalf("events len = %d, want 1", len(body.Events))
	}
	if body.Events[0]["span_id"].(string) != hex.EncodeToString(span) {
		t.Errorf("span_id mismatch")
	}
	if fs.listArg != "packet-1" {
		t.Errorf("store called with packet_id = %q, want packet-1", fs.listArg)
	}
}

func TestListAmiePacketAudits_NoAdminReturns503(t *testing.T) {
	srv := newAuditServer(t, nil)
	resp, err := http.Get(srv.URL + "/connectors/amie/packets/p/audits")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}
}

func TestListAmiePacketAudits_StoreErrorReturns500(t *testing.T) {
	fs := &fakeAmiePacketAuditStore{listErr: errors.New("boom")}
	srv := newAuditServer(t, &AdminDeps{AmiePacketAudits: fs})
	resp, err := http.Get(srv.URL + "/connectors/amie/packets/p/audits")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
