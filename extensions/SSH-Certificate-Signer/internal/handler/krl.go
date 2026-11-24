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
	"fmt"
	"net/http"

	"github.com/apache/airavata-custos/signer/internal/krl"
)

type KRLSource interface {
	Snapshot() krl.Snapshot
}

type KRLHandler struct {
	source KRLSource
}

func NewKRLHandler(source KRLSource) *KRLHandler {
	return &KRLHandler{source: source}
}

func (h *KRLHandler) Handle(w http.ResponseWriter, r *http.Request) {
	s := h.source.Snapshot()
	etag := fmt.Sprintf(`"%d"`, s.Version)
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(s.Bytes)
}
