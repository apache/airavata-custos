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

package amieclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/apache/airavata-custos/allocations/access-amie/config"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	siteCode   string
	apiKey     string
}

func New(cfg config.AMIEConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.ReadTimeout,
			Transport: &http.Transport{
				ResponseHeaderTimeout: cfg.ConnectTimeout,
			},
		},
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
		siteCode: cfg.SiteCode,
		apiKey:   cfg.APIKey,
	}
}

// FetchInProgressPackets retrieves all in-progress AMIE packets for the
// configured site. On HTTP errors or non-2xx responses it logs a warning and
// returns an empty response rather than propagating the error, so the poller
// can continue retrying on the next cycle.
func (c *Client) FetchInProgressPackets(ctx context.Context) ([]map[string]any, error) {
	url := fmt.Sprintf("%s/packets/%s", c.baseURL, c.siteCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Warn("failed to create AMIE fetch request", "error", err)
		return []map[string]any{}, nil
	}
	req.Header.Set("XA-SITE", c.siteCode)
	req.Header.Set("XA-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Warn("AMIE fetch request failed", "error", err)
		return []map[string]any{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("AMIE fetch returned non-2xx status", "status", resp.StatusCode)
		return []map[string]any{}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("failed to read AMIE response body", "error", err)
		return []map[string]any{}, nil
	}

	return parsePacketsFromResponse(body)
}

// parsePacketsFromResponse extracts a slice of packet maps from the AMIE API
// JSON response. It handles three response shapes that the AMIE API may
// produce:
//  1. A root JSON array of packet objects.
//  2. A root JSON object with a "result" key containing an array.
//  3. A root JSON object (single packet or envelope with scalar "result").
func parsePacketsFromResponse(body []byte) ([]map[string]any, error) {
	// Try as a JSON array first.
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err == nil {
		return arr, nil
	}

	// Try as a JSON object.
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		slog.Error("failed to parse AMIE response JSON", "error", err)
		return []map[string]any{}, nil
	}

	result, hasResult := obj["result"]
	if !hasResult {
		// The root object itself is the single packet.
		return []map[string]any{obj}, nil
	}

	// "result" is present -- check if it is an array.
	if resultArr, ok := result.([]any); ok {
		packets := make([]map[string]any, 0, len(resultArr))
		for _, item := range resultArr {
			if m, ok := item.(map[string]any); ok {
				packets = append(packets, m)
			}
		}
		return packets, nil
	}

	// "result" is a single value -- wrap it if it is a map.
	if m, ok := result.(map[string]any); ok {
		return []map[string]any{m}, nil
	}

	// Scalar or unrecognised "result" -- return the envelope itself.
	return []map[string]any{obj}, nil
}

func (c *Client) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	url := fmt.Sprintf("%s/packets/%s/%d/reply", c.baseURL, c.siteCode, packetRecID)

	payload, err := json.Marshal(reply)
	if err != nil {
		return fmt.Errorf("marshal reply payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create reply request: %w", err)
	}
	req.Header.Set("XA-SITE", c.siteCode)
	req.Header.Set("XA-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send reply request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AMIE reply failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("AMIE health check request failed: %w", err)
	}
	defer resp.Body.Close()

	// Drain the body so the connection can be reused.
	_, _ = io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("AMIE health check returned status %d", resp.StatusCode)
}
