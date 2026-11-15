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

package subscribers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/models"
)

// recordingVMS stands in for VMS and records each create and delete as "<kind> <path>".
func recordingVMS(t *testing.T) (Cluster, *[]string) {
	t.Helper()
	var log []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Path string }
		_ = json.NewDecoder(r.Body).Decode(&body)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		switch r.URL.Path {
		case "/api/token/":
			_, _ = w.Write([]byte(`{"access":"t"}`))
		case "/api/folders/create_folder/":
			log = append(log, "folder "+body.Path)
		case "/api/folders/delete_folder/":
			log = append(log, "delete "+body.Path)
		case "/api/quotas/":
			log = append(log, "quota "+body.Path)
		case "/api/views/":
			log = append(log, "view "+body.Path)
		}
	}))
	t.Cleanup(srv.Close)
	return Cluster{VMS: client.New(srv.URL, "u", "p", 1)}, &log
}

// Make sure a created storage gets its folder, its view when its type is one, and its quota when it has one.
func TestCreateStorage(t *testing.T) {
	tests := []struct {
		name string
		st   models.Storage
		want []string
	}{
		{name: "view and quota", st: models.Storage{ComputeClusterID: "c1", Type: "project", Path: "/p", Mode: "2750", QuotaBytes: 1}, want: []string{"folder /p", "view /p", "quota /p"}},
		{name: "not a view type", st: models.Storage{ComputeClusterID: "c1", Type: "user_home", Path: "/h", Mode: "0700"}, want: []string{"folder /h"}},
		{name: "skips unlisted cluster", st: models.Storage{ComputeClusterID: "c2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, log := recordingVMS(t)
			c.Views = map[models.StorageType]bool{"project": true, "user_home": false}
			s := NewStorageSubscriber(nil, nil, map[string]Cluster{"c1": c})
			if err := s.createStorage(context.Background(), tt.st); err != nil || !slices.Equal(*log, tt.want) {
				t.Fatalf("err=%v calls=%v, want %v", err, *log, tt.want)
			}
		})
	}
}

// Make sure an update sets the storage's quota and a delete deletes its folder.
func TestUpdateAndDeleteStorage(t *testing.T) {
	c, log := recordingVMS(t)
	s := NewStorageSubscriber(nil, nil, map[string]Cluster{"c1": c})
	st := models.Storage{ComputeClusterID: "c1", Path: "/p", QuotaBytes: 1}
	if err, want := errors.Join(s.updateStorage(context.Background(), st), s.deleteStorage(context.Background(), st)), []string{"quota /p", "delete /p"}; err != nil || !slices.Equal(*log, want) {
		t.Fatalf("err=%v calls=%v, want %v", err, *log, want)
	}
}
