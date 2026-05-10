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
	"testing"
)

func TestRevokeRequest_NoIdentifier(t *testing.T) {
	req := RevokeRequest{
		Reason: "no reason",
	}
	if req.SerialNumber != nil || req.KeyID != nil || req.CAFingerprint != nil {
		t.Error("all identifiers should be nil")
	}
}

// Validate JSON marshaling
func TestRevokeResponse_JSON(t *testing.T) {
	resp := RevokeResponse{
		Success:      true,
		Message:      "Certificate(s) revoked successfully",
		RevokedCount: 1,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	if parsed["success"] != true {
		t.Error("expected success true")
	}
	if parsed["revoked_count"].(float64) != 1 {
		t.Error("expected revoked_count 1")
	}
}
