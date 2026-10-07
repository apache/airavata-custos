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

// Package client creates folders and quotas through the VAST VMS REST API.
// Each create logs in for a fresh JWT from `/api/token/`, since creates are rare.
// VMS answers most refusals, including an existing folder, with a 503, so a
// refused create counts as success when the folder or quota exists; any other
// error goes back to the bus, which retries it.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	url, username, password string
	tenantID                int
	http                    *http.Client
}

func New(vmsURL, username, password string, tenantID int) *Client {
	return &Client{
		url:      strings.TrimSuffix(vmsURL, "/") + "/api/",
		username: username,
		password: password,
		tenantID: tenantID,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

// CreateFolder creates path owned by user:group unless it exists.
func (c *Client) CreateFolder(ctx context.Context, path, user, group string, mode int) error {
	return c.create(ctx, "folders/create_folder/", map[string]any{
		"path": path, "user": user, "group": group, "owner_is_group": false, "create_dir_mode": fmt.Sprintf("%o", mode), "tenant_id": c.tenantID,
	}, func(token string) bool {
		return c.do(ctx, token, http.MethodPost, "folders/stat_path/", map[string]any{"path": path, "tenant_id": c.tenantID}, nil) == nil
	})
}

// CreateQuota creates a quota on path, named after it, unless a quota by that name exists.
func (c *Client) CreateQuota(ctx context.Context, path string, hardLimit, hardLimitInodes int64) error {
	return c.create(ctx, "quotas/", map[string]any{
		"name": path, "path": path, "create_dir": false, "hard_limit": hardLimit, "hard_limit_inodes": hardLimitInodes, "tenant_id": c.tenantID,
	}, func(token string) bool {
		var found []json.RawMessage
		return c.do(ctx, token, http.MethodGet, "quotas/?name="+url.QueryEscape(path), nil, &found) == nil && len(found) > 0
	})
}

func (c *Client) create(ctx context.Context, path string, body map[string]any, exists func(token string) bool) error {
	var t struct{ Access string }
	if err := c.do(ctx, "", http.MethodPost, "token/", map[string]string{"username": c.username, "password": c.password}, &t); err != nil {
		return err
	}
	if err := c.do(ctx, t.Access, http.MethodPost, path, body, nil); err != nil && !exists(t.Access) {
		return err
	}
	return nil
}

func (c *Client) do(ctx context.Context, token, method, path string, body, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.url+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return fmt.Errorf("vast: %s %s: %d: %s", method, path, resp.StatusCode, msg)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
