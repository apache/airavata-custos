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

package folders

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/db/dbtest"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

type fakeVMS struct {
	removed, audited []string
	quotas           map[string]client.Quota
}

func (f *fakeVMS) Remove(_ context.Context, path string) error {
	f.removed = append(f.removed, path)
	return nil
}

func (f *fakeVMS) Quotas(context.Context) (map[string]client.Quota, error) { return f.quotas, nil }

func (f *fakeVMS) CreateQuota(context.Context, string, int64, int64) error { return nil }

func (f *fakeVMS) CreateAuditEvent(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
	f.audited = append(f.audited, e.EntityID)
	return e, nil
}

// Make sure a deletion marks the folder and everything under it, a returning owner unmarks its folder, and deleting
// it removes it from VAST with everything under it once all of it is pending.
func TestReview(t *testing.T) {
	database := dbtest.Open(t, dbtest.CoreDSN(t), dbtest.Migrations{FS: Migrations, Dir: "migrations", Name: "vast"})
	dbtest.Truncate(t, database, "vast_folders")
	vast := &fakeVMS{quotas: map[string]client.Quota{"/home/jdoe": {HardLimit: 100, Used: 40}}}
	ctx, s := context.Background(), New(database, func(string) (VMS, error) { return vast, nil }, vast)
	for _, f := range [][3]string{{"/home/jdoe", "cluster_user", "jdoe"}, {"/project/a", "allocation", "a"},
		{"/project/a/jdoe", "cluster_user", "jdoe"}, {"/project/ab", "allocation", "ab"}} {
		if _, err := s.Record(ctx, "c1", f[0], f[1], f[2], f[2]+":"+f[2], nil); err != nil {
			t.Fatal(err)
		}
	}
	folders := func() map[string]Folder {
		var items []Folder
		rec := httptest.NewRecorder()
		s.list(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		_ = json.Unmarshal(rec.Body.Bytes(), &items)
		byPath := map[string]Folder{}
		for _, f := range items {
			byPath[f.Path] = f
		}
		return byPath
	}
	call := func(action http.HandlerFunc, path, body string) int {
		r := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
		r.SetPathValue("id", folders()[path].ID)
		rec := httptest.NewRecorder()
		action(rec, r.WithContext(identity.WithCaller(ctx, &identity.Caller{UserID: "admin"})))
		return rec.Code
	}
	if q := folders()["/home/jdoe"].Quota; q == nil || q.Used != 40 || folders()["/project/a"].Quota != nil {
		t.Fatalf("quotas: %v", folders())
	}
	code := call(s.setQuota, "/home/jdoe", `{"hard_limit": 500, "until": "`+time.Now().Add(time.Hour).Format(time.RFC3339)+`"}`)
	if limit, _ := s.Record(ctx, "c1", "/home/jdoe", "cluster_user", "jdoe", "jdoe:jdoe", nil); code != http.StatusNoContent || limit != 500 || !slices.Equal(vast.audited, []string{"c1:/home/jdoe"}) {
		t.Fatalf("quota edit: %d limit %d audited %v", code, limit, vast.audited)
	}
	_ = s.MarkPending(ctx, "c1", "allocation", "a")
	if got := folders(); got["/project/a/jdoe"].Status != "pending_deletion" || got["/project/ab"].Status != "active" {
		t.Fatalf("after marking: %v", got)
	}
	_, _ = s.Record(ctx, "c1", "/project/a/jdoe", "cluster_user", "jdoe", "jdoe:a", &models.ComputeAllocation{ID: "a1", Name: "a"})
	if got := folders()["/project/a/jdoe"]; got.Status != "active" || got.AllocationName == nil || *got.AllocationName != "a" {
		t.Fatalf("returning owner: %+v", got)
	}
	if code := call(s.deleteFolder, "/project/ab", ""); code != http.StatusNotFound {
		t.Fatalf("delete active: %d", code)
	}
	if code := call(s.deleteFolder, "/project/a", ""); code != http.StatusNotFound {
		t.Fatalf("delete over an active folder: %d", code)
	}
	_ = s.MarkPending(ctx, "c1", "allocation", "a")
	if code := call(s.deleteFolder, "/project/a", ""); code != http.StatusNoContent || !slices.Equal(vast.removed, []string{"/project/a"}) || len(folders()) != 2 {
		t.Fatalf("delete: %d removed %v left %v", code, vast.removed, folders())
	}
}
