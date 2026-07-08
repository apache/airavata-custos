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

package common

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON_NilSliceEncodesAsEmptyArray(t *testing.T) {
	var items []string
	rr := httptest.NewRecorder()
	WriteJSON(rr, http.StatusOK, items)
	if body := strings.TrimSpace(rr.Body.String()); body != "[]" {
		t.Errorf("nil slice body: got %q, want []", body)
	}
}

func TestWriteJSON_NonNilValuesUnchanged(t *testing.T) {
	cases := []struct {
		name string
		body any
		want string
	}{
		{"populated slice", []string{"a"}, `["a"]`},
		{"empty slice", []string{}, `[]`},
		{"object", map[string]string{"k": "v"}, `{"k":"v"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			WriteJSON(rr, http.StatusOK, tc.body)
			if body := strings.TrimSpace(rr.Body.String()); body != tc.want {
				t.Errorf("got %q, want %q", body, tc.want)
			}
		})
	}
}

func TestWriteJSON_NilBodyWritesNothing(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteJSON(rr, http.StatusNoContent, nil)
	if rr.Body.Len() != 0 {
		t.Errorf("nil body: got %q, want empty", rr.Body.String())
	}
}
