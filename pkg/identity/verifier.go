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

package identity

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
)

// JWTVerifier validates bearer tokens against an OIDC provider's JWKS,
// issuer, audience, and expiry.
type JWTVerifier struct{ inner *oidc.IDTokenVerifier }

// Claims are the minimal subset the middleware needs after verification.
type Claims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

// NewJWTVerifier resolves the provider's discovery doc and builds a verifier
// pinned to the audience.
func NewJWTVerifier(ctx context.Context, issuer, audience string) (*JWTVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	return &JWTVerifier{inner: provider.Verifier(&oidc.Config{ClientID: audience})}, nil
}

// Verify checks signature, issuer, audience, and expiry.
func (j *JWTVerifier) Verify(ctx context.Context, rawToken string) (*Claims, error) {
	token, err := j.inner.Verify(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	var claims Claims
	if err := token.Claims(&claims); err != nil {
		return nil, err
	}
	return &claims, nil
}
