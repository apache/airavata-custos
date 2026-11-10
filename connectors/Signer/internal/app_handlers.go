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

package internal

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

const (
	AppActionGenerateCertificate = "signer:app:generate-certificate"
	AppActionRevokeCertificate   = "signer:app:revoke-certificate"
)

func appRequestedBy(r *http.Request) string {
	return "app/" + identity.CallerFromContext(r.Context()).AppKey
}

func (h *Handlers) generateCertificateForApp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserEmail     string `json:"user_email"`
		PublicKey     string `json:"public_key"`
		SourceAddress string `json:"source_address"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBytes)).Decode(&req); err != nil || req.PublicKey == "" {
		refuse(w, http.StatusBadRequest, "invalid_public_key", "A public key is required")
		return
	}
	user, subject, ok := h.requireEligibleUser(w, r, req.UserEmail)
	if !ok {
		return
	}
	h.issueCertificate(w, r, user.ID, map[string]any{
		"public_key":     req.PublicKey,
		"user_email":     user.Email,
		"user_sub":       subject,
		"requested_by":   appRequestedBy(r),
		"source_address": req.SourceAddress,
	})
}

// requireEligibleUser checks the user for the app which is asking certificates for. These are the checks a user passes at login, and an app call does not go through that.
// TODO - move to core when another connector needs these checks.
//
//	Then check the user's allow only when the calling connector needs it.
func (h *Handlers) requireEligibleUser(w http.ResponseWriter, r *http.Request, email string) (*models.User, string, bool) {
	user, err := h.core.GetUserByEmail(r.Context(), email)
	if errors.Is(err, service.ErrNotFound) {
		refuse(w, http.StatusNotFound, "user_not_found", "No user has this email")
		return nil, "", false
	}
	if err != nil {
		accountReadFailed(w, err)
		return nil, "", false
	}
	authorized, err := h.core.IsAppAuthorized(r.Context(), identity.CallerFromContext(r.Context()).AppID, user.ID)
	if err != nil {
		accountReadFailed(w, err)
		return nil, "", false
	}
	if !authorized {
		refuse(w, http.StatusForbidden, "app_not_authorized", "The user has not allowed this app")
		return nil, "", false
	}
	if user.Status != models.UserActive {
		refuse(w, http.StatusForbidden, "user_not_active", "This user is not active")
		return nil, "", false
	}
	pending, err := h.core.PendingUserAgreements(r.Context(), user.ID)
	if err != nil {
		accountReadFailed(w, err)
		return nil, "", false
	}
	if len(pending) > 0 {
		refuse(w, http.StatusForbidden, "consent_required", "The user must accept the user agreement first")
		return nil, "", false
	}
	identities, err := h.core.ListUserIdentitiesForUser(r.Context(), user.ID)
	if err != nil {
		accountReadFailed(w, err)
		return nil, "", false
	}
	for _, linked := range identities {
		if linked.OIDCSub != "" {
			return user, linked.OIDCSub, true
		}
	}
	refuse(w, http.StatusForbidden, "user_not_signed_in", "This user has not signed in to Custos yet")
	return nil, "", false
}

func (h *Handlers) listCertificatesForApp(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	query.Set("requested_by", appRequestedBy(r))
	h.relay(w, r, http.MethodGet, "/api/v1/certificates?"+query.Encode(), nil)
}

func (h *Handlers) fetchAppsCertificate(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	return h.fetchIfOwned(w, r, func(c certificateOwner) bool { return c.RequestedBy == appRequestedBy(r) })
}

func (h *Handlers) getCertificateForApp(w http.ResponseWriter, r *http.Request) {
	if body, ok := h.fetchAppsCertificate(w, r); ok {
		writeCertificate(w, body)
	}
}

func (h *Handlers) revokeCertificateForApp(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.fetchAppsCertificate(w, r); ok {
		h.revoke(w, r, appRequestedBy(r))
	}
}
