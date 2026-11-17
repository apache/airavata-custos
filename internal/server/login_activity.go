// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements. See the NOTICE file distributed with this work
// for additional information regarding copyright ownership. The ASF licenses
// this file to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
package server

import (
	"errors"
	"io"
	"net/http"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
)

// recordLoginEvent accepts only an empty body. Identity and time come from
// verified claims, never from browser input.
func (s *Server) recordLoginEvent(w http.ResponseWriter, r *http.Request) {
	var body struct{}
	if err := common.DecodeJSON(r, &body); err != nil && !errors.Is(err, io.EOF) {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	claims := identity.ClaimsFromContext(r.Context())
	if claims == nil {
		common.WriteError(w, http.StatusBadRequest, errors.New("verified login claims are missing"))
		return
	}
	err := s.svc.PublishPortalLogin(r.Context(), caller.UserID, claims)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusAccepted, struct {
		Accepted bool `json:"accepted"`
	}{true})
}
