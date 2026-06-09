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

// Package client wraps the COmanage Core API and REST surfaces. Auth is HTTP
// Basic.
package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// restAPIVersion is the COmanage REST v1 model version.
const restAPIVersion = "1.0"

type Config struct {
	RegistryURL string
	COID        int
	APIUser     string
	APIKey      string
	// PersonIDType is the COmanage Identifier Type used to look up and tag a
	// CoPerson (e.g. the type name configured in the registry's Identifier
	// Types). Required value.
	PersonIDType    string
	UnixClusterID   int
	CustosClusterID string
	DefaultShell    string
	HomedirPrefix   string
	HTTPTimeout     time.Duration
}

type Client struct {
	cfg  Config
	http *http.Client
}

func New(cfg Config) *Client {
	timeout := cfg.HTTPTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Config() Config { return c.cfg }

// coreAPI and restAPI are the two URL families. Centralized so an upstream
// version bump flips one line.
func (c *Client) coreAPI(path string) string {
	return fmt.Sprintf("%s/api/co/%d/core/v1%s", c.cfg.RegistryURL, c.cfg.COID, path)
}

func (c *Client) restAPI(path string) string {
	return c.cfg.RegistryURL + path
}

// Do issue an authenticated request and retries 5xx with backoff.
func (c *Client) Do(method, url string, body []byte) (*http.Response, []byte, error) {
	resp, respBody, err := c.doOnce(method, url, body, c.cfg.APIKey)
	if err == nil && resp.StatusCode >= 500 && resp.StatusCode < 600 {
		// 1s, 2s, 4s
		for attempt := 1; attempt <= 3; attempt++ {
			_ = resp.Body.Close()
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
			resp, respBody, err = c.doOnce(method, url, body, c.cfg.APIKey)
			if err != nil || resp.StatusCode < 500 {
				break
			}
		}
	}
	return resp, respBody, err
}

func (c *Client) doOnce(method, url string, body []byte, apiKey string) (*http.Response, []byte, error) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.SetBasicAuth(c.cfg.APIUser, apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http: %w", err)
	}
	respBody, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		return resp, nil, fmt.Errorf("read response: %w", readErr)
	}
	return resp, respBody, nil
}

var (
	ErrAuth401  = errors.New("comanage: 401 unauthorized")
	ErrNotFound = errors.New("comanage: not found")
)

type HTTPError struct {
	Method     string
	URL        string
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("comanage %s %s: %d: %s", e.Method, e.URL, e.StatusCode, truncate(e.Body, 200))
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
