// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package auth handles client credential authentication and OIDC token validation.
package auth

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/apache/airavata-custos/signer/internal/store"
)

type ClientAuthenticator struct {
	db *store.DB
}

func NewClientAuthenticator(db *store.DB) *ClientAuthenticator {
	return &ClientAuthenticator{db: db}
}

// ParseClientID parses the X-Client-Id header value into tenant_id and client_id.
// The header format is {tenant_id}:{client_id}, split on the first colon only.
func ParseClientID(header string) (tenantID, clientID string, err error) {
	if header == "" {
		return "", "", fmt.Errorf("empty client ID header")
	}
	idx := strings.Index(header, ":")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid client ID format: no colon separator")
	}
	tenantID = header[:idx]
	clientID = header[idx+1:]
	if tenantID == "" || clientID == "" {
		return "", "", fmt.Errorf("invalid client ID format: empty tenant or client")
	}
	return tenantID, clientID, nil
}

type AuthResult struct {
	Config *store.ClientConfig
}

func (a *ClientAuthenticator) Authenticate(ctx context.Context, clientIDHeader, secretHeader string) (*AuthResult, error) {
	if clientIDHeader == "" {
		return nil, &AuthError{Code: "unauthorized", Message: "Missing X-Client-Id header", Status: 401}
	}
	if secretHeader == "" {
		return nil, &AuthError{Code: "unauthorized", Message: "Missing X-Client-Secret header", Status: 401}
	}

	tenantID, clientID, err := ParseClientID(clientIDHeader)
	if err != nil {
		return nil, &AuthError{Code: "unauthorized", Message: err.Error(), Status: 401}
	}

	cfg, err := a.db.GetClientConfig(ctx, tenantID, clientID)
	if err != nil {
		return nil, fmt.Errorf("looking up client config: %w", err)
	}
	if cfg == nil {
		return nil, &AuthError{Code: "unauthorized", Message: "Unknown client", Status: 401}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cfg.ClientSecret), []byte(secretHeader)); err != nil {
		return nil, &AuthError{Code: "unauthorized", Message: "Invalid client secret", Status: 401}
	}

	if !cfg.Enabled {
		return nil, &AuthError{Code: "client_disabled", Message: "Client is disabled", Status: 403}
	}

	return &AuthResult{Config: cfg}, nil
}

type AuthError struct {
	Code    string
	Message string
	Status  int
}

func (e *AuthError) Error() string {
	return e.Message
}
