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

package server

import (
	"errors"
	"net/http"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

func callerAndClaims(w http.ResponseWriter, r *http.Request) (*identity.Caller, *identity.Claims) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return nil, nil
	}
	claims := identity.ClaimsFromContext(r.Context())
	if claims == nil {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing token claims"))
		return nil, nil
	}
	return caller, claims
}

// @Summary	List the current agreement if the caller has not accepted it
// @Tags	Consent
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.UserAgreement
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Router	/me/agreements/pending [get]
func (s *Server) listCallerPendingAgreements(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	rows, err := s.svc.PendingUserAgreements(r.Context(), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.UserAgreement{}
	}
	common.WriteJSON(w, http.StatusOK, rows)
}

// @Summary	Accept a published agreement as the caller
// @Description	A first sign-in accepts through POST /me/onboarding instead.
// @Tags	Consent
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Agreement ID"
// @Success	201	{object}	models.UserAgreementAcceptance
// @Failure	400	{object}	object{error=string}	"Caller still has to onboard, or the agreement is not published"
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Failure	404	{object}	object{error=string}	"Unknown agreement"
// @Router	/me/agreements/{id}/accept [post]
func (s *Server) acceptCallerAgreement(w http.ResponseWriter, r *http.Request) {
	caller, claims := callerAndClaims(w, r)
	if caller == nil {
		return
	}
	if caller.NeedsOnboarding {
		common.WriteError(w, http.StatusBadRequest, errors.New("accept agreements through onboarding"))
		return
	}
	acc, err := s.svc.AcceptUserAgreement(r.Context(), caller.UserID, r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.svc.InvalidateIdentity(claims.Sub)
	common.WriteJSON(w, http.StatusCreated, acc)
}
