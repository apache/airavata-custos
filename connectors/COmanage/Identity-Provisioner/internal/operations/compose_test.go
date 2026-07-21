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

package operations

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

// sampleComposite mirrors a real CoPerson composite trimmed to the fields the
// operations layer reads.
const sampleComposite = `{
    "CoPerson":{"meta":{"id":98},"co_id":2,"status":"A"},
    "Name":[{"given":"GoalE2E","family":"Throwaway","type":"official","primary_name":true}],
    "EmailAddress":[{"mail":"goal-e2e@example.invalid","type":"official","verified":false}],
    "CoGroupMember":[
        {"co_group_id":7,"member":true,"owner":false},
        {"co_group_id":25,"member":true,"owner":true}
    ],
    "Identifier":[
        {"identifier":"http://test.invalid/sub","type":"oidcsub","login":true,"status":"A"},
        {"identifier":"Person100016","type":"comanage_id","login":false,"status":"A"},
        {"identifier":"100016","type":"comanage_number","login":false,"status":"A"},
        {"identifier":"vspectes2","type":"uid","login":false,"status":"A"},
        {"identifier":"2000016","type":"uidnumber","login":false,"status":"A"},
        {"identifier":"2000016","type":"gidnumber","login":false,"status":"A"}
    ]
}`

func TestExtractIdentifier(t *testing.T) {
	tests := []struct {
		typeName string
		want     string
	}{
		{"comanage_id", "Person100016"},
		{"uidnumber", "2000016"},
		{"gidnumber", "2000016"},
		{"oidcsub", "http://test.invalid/sub"},
		{"missing-type", ""},
	}
	for _, tc := range tests {
		t.Run(tc.typeName, func(t *testing.T) {
			got, err := extractIdentifier([]byte(sampleComposite), tc.typeName)
			if err != nil {
				t.Fatalf("extractIdentifier: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractCoPersonID(t *testing.T) {
	got, err := extractCoPersonID([]byte(sampleComposite))
	if err != nil {
		t.Fatalf("extractCoPersonID: %v", err)
	}
	if got != 98 {
		t.Errorf("got %d, want 98", got)
	}
}

func TestMergeUnixClusterAccount_PreservesAllKeysAndAppendsBlock(t *testing.T) {
	block := UnixClusterAccountBlock{
		UnixClusterId:    1,
		SyncMode:         "M",
		Status:           "A",
		Username:         "custos-gthrowa",
		Uid:              2000016,
		Gecos:            "",
		LoginShell:       "/bin/bash",
		HomeDirectory:    "/home/custos-gthrowa",
		PrimaryCoGroupId: 25,
	}
	out, err := mergeUnixClusterAccount([]byte(sampleComposite), block)
	if err != nil {
		t.Fatalf("mergeUnixClusterAccount: %v", err)
	}

	var merged map[string]json.RawMessage
	if err := json.Unmarshal(out, &merged); err != nil {
		t.Fatalf("decode merged: %v", err)
	}

	for _, key := range []string{"CoPerson", "Name", "EmailAddress", "Identifier", "CoGroupMember", "UnixClusterAccount"} {
		if _, ok := merged[key]; !ok {
			t.Errorf("merged composite missing key %q", key)
		}
	}

	// UnixClusterAccount must be an array with one block matching what we sent.
	var unix []UnixClusterAccountBlock
	if err := json.Unmarshal(merged["UnixClusterAccount"], &unix); err != nil {
		t.Fatalf("decode UnixClusterAccount: %v", err)
	}
	if len(unix) != 1 || unix[0].Username != "custos-gthrowa" || unix[0].Uid != 2000016 || unix[0].PrimaryCoGroupId != 25 {
		t.Errorf("unix block: %+v", unix)
	}
}

func TestBuildCreatePersonBody_Shape(t *testing.T) {
	u := &models.User{FirstName: "GoalE2E", LastName: "Throwaway", Email: "goal-e2e@example.invalid"}
	raw, err := buildCreatePersonBody(2, u, "http://idp.invalid/users/9")
	if err != nil {
		t.Fatalf("buildCreatePersonBody: %v", err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"CoPerson", "Name", "EmailAddress", "Identifier"} {
		if _, ok := got[key]; !ok {
			t.Errorf("body missing %q", key)
		}
	}
	body := string(raw)
	if !strings.Contains(body, `"verified":true`) {
		t.Errorf("email must be created verified: %s", body)
	}
	if !strings.Contains(body, `"type":"oidcsub"`) || !strings.Contains(body, `"login":true`) {
		t.Errorf("create body must carry the login oidcsub identifier: %s", body)
	}

	raw, err = buildCreatePersonBody(2, u, "")
	if err != nil {
		t.Fatalf("buildCreatePersonBody without sub: %v", err)
	}
	if strings.Contains(string(raw), "Identifier") {
		t.Errorf("no Identifier block expected without a sub: %s", raw)
	}
}
