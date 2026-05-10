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

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/apache/airavata-custos/signer/internal/httputil"
)

type UserInfoResponse struct {
	Subject   string `json:"subject"`
	Issuer    string `json:"issuer"`
	Email     string `json:"email"`
	Principal string `json:"principal"`
}

type UserInfoHandler struct{}

func NewUserInfoHandler() *UserInfoHandler {
	return &UserInfoHandler{}
}

func (h *UserInfoHandler) Handle(w http.ResponseWriter, r *http.Request) {
	identity := httputil.UserIdentityFromContext(r.Context())
	if identity == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Missing user identity")
		return
	}

	resp := UserInfoResponse{
		Subject:   identity.Subject,
		Issuer:    identity.Issuer,
		Email:     identity.Email,
		Principal: identity.Principal,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
