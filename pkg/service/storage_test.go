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

package service

import (
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure a target is refused unless its path names {user}, {allocation}, both, or {name} alone, and its owner and
// group name only those.
func TestSetStorageTargets(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		group   string
		wantErr bool
	}{
		{name: "member", path: "/project/{allocation}/{user}"},
		{name: "connector holder", path: "/trial-home/{name}"},
		{name: "no holder", path: "/shared", wantErr: true},
		{name: "name with user", path: "/trial-home/{name}/{user}", wantErr: true},
		{name: "group naming a placeholder the path lacks", path: "/project/{allocation}", group: "{user}", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&Service{}).SetStorageTargets(map[string]map[models.StorageType]models.StorageTarget{"c1": {"t": {Path: tt.path, Group: tt.group, Mode: "0700"}}})
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
