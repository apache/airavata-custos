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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// fakePacketStore implements only what the retry handler calls; any other
// method panics through the nil embedded interface.
type fakePacketStore struct {
	store.PacketStore
	packets    map[string]*model.Packet
	requeueErr error
	requeued   []string
}

func (f *fakePacketStore) FindByID(_ context.Context, id string) (*model.Packet, error) {
	return f.packets[id], nil
}

func (f *fakePacketStore) RequeueFailed(_ context.Context, id string) (*model.Packet, error) {
	if f.requeueErr != nil {
		return nil, f.requeueErr
	}
	p := f.packets[id]
	if p == nil || p.Status != model.PacketStatusFailed {
		return nil, nil
	}
	p.Status = model.PacketStatusNew
	f.requeued = append(f.requeued, id)
	return p, nil
}

func newRetryTestServer(t *testing.T, packets store.PacketStore, privileges ...models.PrivilegeKey) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	NewHandlers(nil, packets).RegisterRoutes(identity.NewRouter(mux))
	wrap := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := identity.WithCaller(r.Context(), &identity.Caller{UserID: "test-user"})
		ctx = identity.WithPrivilegesForTest(ctx, privileges)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
	srv := httptest.NewServer(wrap)
	t.Cleanup(srv.Close)
	return srv
}

func postRetry(t *testing.T, srv *httptest.Server, id string) *http.Response {
	t.Helper()
	resp, err := http.Post(srv.URL+"/connectors/amie/packets/"+id+"/retry", "application/json", nil)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestRetryPacket_QueuesFailedPacket(t *testing.T) {
	fs := &fakePacketStore{packets: map[string]*model.Packet{
		"p1": {ID: "p1", AmieID: 42, Type: "request_account_create", Status: model.PacketStatusFailed, ReceivedAt: time.Unix(1700000000, 0).UTC()},
	}}
	srv := newRetryTestServer(t, fs, PacketsWrite)

	resp := postRetry(t, srv, "p1")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Queued bool `json:"queued"`
		Packet struct {
			ID string `json:"id"`
		} `json:"packet"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.Queued || body.Packet.ID != "p1" {
		t.Errorf("body = %+v, want queued packet p1", body)
	}
	if len(fs.requeued) != 1 {
		t.Errorf("requeued = %v, want [p1]", fs.requeued)
	}
}

func TestRetryPacket_RejectsPacketThatIsNotFailed(t *testing.T) {
	for _, status := range []model.PacketStatus{
		model.PacketStatusNew,
		model.PacketStatusProcessed,
		model.PacketStatusWaitingApproval,
		model.PacketStatusRefused,
	} {
		fs := &fakePacketStore{packets: map[string]*model.Packet{"p1": {ID: "p1", Status: status}}}
		srv := newRetryTestServer(t, fs, PacketsWrite)

		if resp := postRetry(t, srv, "p1"); resp.StatusCode != http.StatusConflict {
			t.Errorf("%s: status = %d, want 409", status, resp.StatusCode)
		}
		if len(fs.requeued) != 0 {
			t.Errorf("%s: packet was requeued", status)
		}
	}
}

func TestRetryPacket_UnknownPacketReturns404(t *testing.T) {
	srv := newRetryTestServer(t, &fakePacketStore{packets: map[string]*model.Packet{}}, PacketsWrite)
	if resp := postRetry(t, srv, "missing"); resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRetryPacket_StoreErrorReturns500(t *testing.T) {
	fs := &fakePacketStore{
		packets:    map[string]*model.Packet{"p1": {ID: "p1", Status: model.PacketStatusFailed}},
		requeueErr: errors.New("boom"),
	}
	srv := newRetryTestServer(t, fs, PacketsWrite)
	if resp := postRetry(t, srv, "p1"); resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

func TestRetryPacket_RequiresWritePrivilege(t *testing.T) {
	fs := &fakePacketStore{packets: map[string]*model.Packet{"p1": {ID: "p1", Status: model.PacketStatusFailed}}}
	srv := newRetryTestServer(t, fs, PacketsRead)

	if resp := postRetry(t, srv, "p1"); resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
	if len(fs.requeued) != 0 {
		t.Errorf("packet was requeued without %s", PacketsWrite)
	}
}
