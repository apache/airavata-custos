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

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	SignerCertificatesRead  = "signer:certificates:read"
	SignerCertificatesWrite = "signer:certificates:write"
)

var (
	ErrCoreUnauthorized = errors.New("core rejected bearer token")
	ErrCoreUnavailable  = errors.New("core authorization unavailable")
)

type CoreCaller struct {
	ID         string
	Email      string
	Privileges []string
}

func (c *CoreCaller) HasPrivilege(required string) bool {
	for _, privilege := range c.Privileges {
		if privilege == required {
			return true
		}
	}
	return false
}

type CoreAuthorizer interface {
	ResolveCaller(ctx context.Context, bearer string) (*CoreCaller, error)
}

type CoreAuthorizationClient struct {
	meURL      string
	httpClient *http.Client
}

func NewCoreAuthorizationClient(baseURL string, timeout time.Duration) (*CoreAuthorizationClient, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid core API base URL %q", baseURL)
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &CoreAuthorizationClient{
		meURL:      parsed.String() + "/me",
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

func (c *CoreAuthorizationClient) ResolveCaller(ctx context.Context, bearer string) (*CoreCaller, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.meURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrCoreUnavailable, err)
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCoreUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrCoreUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("%w: core /me returned %d", ErrCoreUnavailable, resp.StatusCode)
	}

	var body struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
		Privileges []string `json:"privileges"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := decoder.Decode(&body); err != nil || body.User.ID == "" {
		return nil, fmt.Errorf("%w: invalid core /me response", ErrCoreUnavailable)
	}
	return &CoreCaller{ID: body.User.ID, Email: body.User.Email, Privileges: body.Privileges}, nil
}
