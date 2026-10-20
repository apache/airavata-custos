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

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteUser(t *testing.T) {
	for status, wantErr := range map[int]bool{http.StatusOK: false, http.StatusNotFound: false, http.StatusInternalServerError: true} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" || r.URL.Path != "/slurmdb/v0.0.41/user/alice" {
				t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(status)
		}))
		err := New(srv.URL, "root", "t", "41").DeleteUser("alice")
		srv.Close()
		if (err != nil) != wantErr {
			t.Errorf("status %d: err = %v", status, err)
		}
	}
}
