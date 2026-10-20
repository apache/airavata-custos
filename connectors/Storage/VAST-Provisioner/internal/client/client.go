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

// Package client creates folders, views and quotas through the VAST VMS REST API.
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
	"strconv"
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

// CreateView creates a view on the existing directory path unless one exists.
// The directory comes from CreateFolder, since VMS ignores a view's owner fields.
func (c *Client) CreateView(ctx context.Context, path string, policyID int, protocols []string) error {
	return c.create(ctx, "views/", map[string]any{
		"path": path, "create_dir": false, "policy_id": policyID, "protocols": protocols, "tenant_id": c.tenantID,
	}, func(token string) bool {
		var found []json.RawMessage
		return c.do(ctx, token, http.MethodGet, "views/?path="+url.QueryEscape(path), nil, &found) == nil && len(found) > 0
	})
}

// ViewPolicyID returns the id of the corresponding view policy name.
func (c *Client) ViewPolicyID(ctx context.Context, name string) (int, error) {
	token, err := c.login(ctx)
	if err != nil {
		return 0, err
	}
	var found []struct{ ID int }
	if err := c.do(ctx, token, http.MethodGet, "viewpolicies/?name="+url.QueryEscape(name), nil, &found); err != nil {
		return 0, err
	}
	if len(found) == 0 {
		return 0, fmt.Errorf("vast: view policy %q not found", name)
	}
	return found[0].ID, nil
}

// RemoveViews deletes the views on path and under it; files and quotas stay.
func (c *Client) RemoveViews(ctx context.Context, path string) error {
	token, err := c.login(ctx)
	if err != nil {
		return err
	}
	var views []struct {
		ID   int
		Path string
	}
	if err := c.do(ctx, token, http.MethodGet, "views/?tenant_id="+strconv.Itoa(c.tenantID), nil, &views); err != nil {
		return err
	}
	for _, v := range views {
		if v.Path != path && !strings.HasPrefix(v.Path, path+"/") {
			continue
		}
		if err := c.do(ctx, token, http.MethodDelete, fmt.Sprintf("views/%d/", v.ID), nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) create(ctx context.Context, path string, body map[string]any, exists func(token string) bool) error {
	token, err := c.login(ctx)
	if err != nil {
		return err
	}
	if err := c.do(ctx, token, http.MethodPost, path, body, nil); err != nil && !exists(token) {
		return err
	}
	return nil
}

func (c *Client) login(ctx context.Context) (string, error) {
	var t struct{ Access string }
	err := c.do(ctx, "", http.MethodPost, "token/", map[string]string{"username": c.username, "password": c.password}, &t)
	return t.Access, err
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
