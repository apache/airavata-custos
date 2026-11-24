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
	"github.com/apache/airavata-custos/pkg/service"
)

// @Summary	Check a login name for onboarding
// @Description	Reports whether the candidate is valid and free on every cluster the caller has an account on, and suggests a free name.
// @Tags	Onboarding
// @Security	BearerAuth
// @Produce	json
// @Param	u	query	string	false	"Candidate login name"
// @Success	200	{object}	service.UsernameCheck
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Router	/me/onboarding/username [get]
func (s *Server) checkOnboardingUsername(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	res, err := s.svc.CheckOnboardingUsername(r.Context(), caller.UserID, r.URL.Query().Get("u"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, res)
}

type onboardingRequest struct {
	Username string   `json:"username"`
	Accepted []string `json:"accepted"`
}

// @Summary	Finish a first sign-in
// @Description	Binds the caller's IdP subject to their pending user once the current agreement is accepted. Nothing is written unless all of it succeeds.
// @Tags	Onboarding
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	onboardingRequest	true	"Chosen username and accepted agreement IDs"
// @Success	200	{object}	models.User
// @Failure	400	{object}	object{error=string}	"Caller is not onboarding, an agreement is missing, or the username is invalid"
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Failure	409	{object}	object{error=string}	"Username is taken"
// @Router	/me/onboarding [post]
func (s *Server) completeOnboarding(w http.ResponseWriter, r *http.Request) {
	caller, claims := callerAndClaims(w, r)
	if caller == nil {
		return
	}
	if !caller.NeedsOnboarding {
		common.WriteError(w, http.StatusBadRequest, errors.New("caller is not waiting for onboarding"))
		return
	}
	var req onboardingRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	user, err := s.svc.CompleteOnboarding(r.Context(), claims, service.OnboardingInput{Username: req.Username, AcceptedIDs: req.Accepted})
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.svc.InvalidateIdentity(claims.Sub)
	common.WriteJSON(w, http.StatusOK, user)
}

// SetupStatusResponse lists the provisioning steps after onboarding.
type SetupStatusResponse struct {
	Steps []service.SetupStep `json:"steps"`
}

// @Summary	Report provisioning progress after onboarding
// @Description	One entry per step, with done_at once it is done. registry: the user is in the identity registry. cluster_account: the account on the named cluster is created. project_group: the named allocation's group is on the cluster.
// @Tags	Onboarding
// @Security	BearerAuth
// @Produce	json
// @Success	200	{object}	SetupStatusResponse
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Router	/me/onboarding/status [get]
func (s *Server) getSetupStatus(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	steps, err := s.svc.SetupSteps(r.Context(), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, SetupStatusResponse{Steps: steps})
}
