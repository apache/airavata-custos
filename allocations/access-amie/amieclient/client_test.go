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

package amieclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/airavata-custos/allocations/access-amie/amieclient"
	"github.com/apache/airavata-custos/allocations/access-amie/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient builds a Client pointed at the given httptest.Server URL.
func newTestClient(serverURL string) *amieclient.Client {
	return amieclient.New(config.AMIEConfig{
		BaseURL:        serverURL,
		SiteCode:       "TESTSITE",
		APIKey:         "test-key",
		ConnectTimeout: 2 * time.Second,
		ReadTimeout:    5 * time.Second,
	})
}

// ---------------------------------------------------------------------------
// FetchInProgressPackets tests
// ---------------------------------------------------------------------------

func TestFetchInProgressPackets_JSONArrayResponse(t *testing.T) {
	packets := []map[string]any{
		{"PacketType": "request_account", "PacketRecID": float64(1)},
		{"PacketType": "notify_person_modify", "PacketRecID": float64(2)},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "TESTSITE", r.Header.Get("XA-SITE"))
		assert.Equal(t, "test-key", r.Header.Get("XA-API-KEY"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(packets)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.FetchInProgressPackets(context.Background())

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "request_account", got[0]["PacketType"])
	assert.Equal(t, "notify_person_modify", got[1]["PacketType"])
}

func TestFetchInProgressPackets_ResultKeyResponse(t *testing.T) {
	inner := []any{
		map[string]any{"PacketType": "request_account", "PacketRecID": float64(10)},
	}
	body := map[string]any{"result": inner, "status": "ok"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.FetchInProgressPackets(context.Background())

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "request_account", got[0]["PacketType"])
}

func TestFetchInProgressPackets_ReturnsEmptyOnNon2xxStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.FetchInProgressPackets(context.Background())

	// The client swallows HTTP errors and returns an empty slice.
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestFetchInProgressPackets_ReturnsEmptyOnNetworkError(t *testing.T) {
	// Immediately closed server causes a network error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close before the request is made

	client := newTestClient(srv.URL)
	got, err := client.FetchInProgressPackets(context.Background())

	require.NoError(t, err)
	assert.Empty(t, got)
}

// ---------------------------------------------------------------------------
// ReplyToPacket tests
// ---------------------------------------------------------------------------

func TestReplyToPacket_Success2xx(t *testing.T) {
	reply := map[string]any{"status": "ok", "message": "accepted"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "TESTSITE", r.Header.Get("XA-SITE"))

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "ok", body["status"])

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	err := client.ReplyToPacket(context.Background(), 42, reply)

	require.NoError(t, err)
}

func TestReplyToPacket_ErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	err := client.ReplyToPacket(context.Background(), 99, map[string]any{"key": "val"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

// ---------------------------------------------------------------------------
// CheckHealth tests
// ---------------------------------------------------------------------------

func TestCheckHealth_UpOn200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	err := client.CheckHealth(context.Background())

	require.NoError(t, err)
}

func TestCheckHealth_DownOn500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	err := client.CheckHealth(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
