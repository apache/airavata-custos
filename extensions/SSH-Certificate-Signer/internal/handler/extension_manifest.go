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

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/apache/airavata-custos/signer/internal/config"
)

type ExtensionManifestHandler struct {
	webURL string
}

type extensionManifest struct {
	SchemaVersion int                       `json:"schema_version"`
	ID            string                    `json:"id"`
	Name          string                    `json:"name"`
	BasePath      string                    `json:"base_path"`
	WebURL        string                    `json:"web_url"`
	Navigation    []extensionNavigationItem `json:"navigation"`
}

type extensionNavigationItem struct {
	Href              string `json:"href"`
	Label             string `json:"label"`
	Group             string `json:"group"`
	Icon              string `json:"icon"`
	RequiredPrivilege string `json:"required_privilege"`
}

func NewExtensionManifestHandler(cfg config.WebConfig) *ExtensionManifestHandler {
	return &ExtensionManifestHandler{webURL: cfg.BaseURL}
}

func (h *ExtensionManifestHandler) Handle(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=60")
	_ = json.NewEncoder(w).Encode(extensionManifest{
		SchemaVersion: 1,
		ID:            "ssh-certificate-signer",
		Name:          "SSH Certificate Signer",
		BasePath:      "/signer",
		WebURL:        h.webURL,
		Navigation: []extensionNavigationItem{{
			Href:              "/signer/certificates",
			Label:             "SSH Certificates",
			Group:             "admin",
			Icon:              "key-round",
			RequiredPrivilege: "signer:certificates:read",
		}},
	})
}
