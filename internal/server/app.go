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
	"net/http"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/models"
)

type appWithTokenView struct {
	models.App
	Token string `json:"token"`
}

type createAppRequest struct {
	Key         string                `json:"key"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Privileges  []models.PrivilegeKey `json:"privileges"`
}

type updateAppRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func writeApp(w http.ResponseWriter, a *models.App, err error) {
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, a)
}

// @Summary	List apps
// @Tags	Apps
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.App
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Router	/apps [get]
func (s *Server) listApps(w http.ResponseWriter, r *http.Request) {
	apps, err := s.svc.ListApps(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, apps)
}

// @Summary	List the privileges an app may hold
// @Tags	Apps
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	string
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Router	/apps/privileges [get]
func (s *Server) listAppPrivileges(w http.ResponseWriter, r *http.Request) {
	common.WriteJSON(w, http.StatusOK, models.AppPrivileges())
}

// @Summary	Create an app
// @Description	The token is returned only here.
// @Tags	Apps
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	createAppRequest	true	"App"
// @Success	201	{object}	appWithTokenView
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Failure	409	{object}	object{error=string}	"That key or name is taken"
// @Router	/apps [post]
func (s *Server) createApp(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	var req createAppRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	a, token, err := s.svc.CreateApp(r.Context(), req.Key, req.Name, req.Description, req.Privileges, caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, appWithTokenView{App: *a, Token: token})
}

// @Summary	Get an app
// @Tags	Apps
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"App ID"
// @Success	200	{object}	models.App
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/apps/{id} [get]
func (s *Server) getApp(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetApp(r.Context(), r.PathValue("id"))
	writeApp(w, a, err)
}

// @Summary	Delete a disabled app
// @Description	The key can never be used again.
// @Tags	Apps
// @Security	BearerAuth
// @Param	id	path	string	true	"App ID"
// @Success	204
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"The app is not disabled"
// @Router	/apps/{id} [delete]
func (s *Server) deleteApp(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	writeNoContent(w, s.svc.DeleteApp(r.Context(), r.PathValue("id"), caller.UserID))
}

// @Summary	Edit an app's name and description
// @Tags	Apps
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"App ID"
// @Param	request	body	updateAppRequest	true	"Name and description"
// @Success	200	{object}	models.App
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"That name is taken"
// @Router	/apps/{id} [put]
func (s *Server) updateApp(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	var req updateAppRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	a, err := s.svc.UpdateApp(r.Context(), r.PathValue("id"), req.Name, req.Description, caller.UserID)
	writeApp(w, a, err)
}

// @Summary	Disable an app
// @Tags	Apps
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"App ID"
// @Success	200	{object}	models.App
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/apps/{id}/disable [post]
func (s *Server) disableApp(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	a, err := s.svc.DisableApp(r.Context(), r.PathValue("id"), caller.UserID)
	writeApp(w, a, err)
}

// @Summary	Enable an app
// @Tags	Apps
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"App ID"
// @Success	200	{object}	models.App
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/apps/{id}/enable [post]
func (s *Server) enableApp(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	a, err := s.svc.EnableApp(r.Context(), r.PathValue("id"), caller.UserID)
	writeApp(w, a, err)
}

// @Summary	Issue a new secret for an app
// @Description	The old token stops working at once. The new token is returned only here.
// @Tags	Apps
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"App ID"
// @Success	200	{object}	appWithTokenView
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks apps:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/apps/{id}/secret [post]
func (s *Server) issueAppSecret(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	a, token, err := s.svc.IssueAppSecret(r.Context(), r.PathValue("id"), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, appWithTokenView{App: *a, Token: token})
}

func writeNoContent(w http.ResponseWriter, err error) {
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
