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
	"reflect"
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

// The composite PUT deletes anything the body leaves out, so merging an
// identifier must return the person whole.
func TestMergeIdentifier_KeepsEveryPropertyOfThePerson(t *testing.T) {
	person := `{
      "CoPerson": {"meta": {"id": 244}, "co_id": 2, "status": "A", "date_of_birth": null, "timezone": null},
      "Name": [{"given": "Test", "family": "Person", "type": "official", "primary_name": true, "language": "en"}],
      "EmailAddress": [{"mail": "user@example.invalid", "type": "official", "verified": true}],
      "Identifier": [
        {"identifier": "Person100099", "type": "comanage_id", "login": false, "status": "A"},
        {"identifier": "2000093", "type": "uidnumber", "login": false, "status": "A"},
        {"identifier": "2000093", "type": "gidnumber", "login": false, "status": "A"},
        {"identifier": "custos-tperson", "type": "uid", "login": false, "status": "A"}
      ],
      "CoPersonRole": [{"affiliation": "member", "status": "A"}],
      "CoGroupMember": [{"co_group_id": 64, "member": true, "owner": false}],
      "UnixClusterAccount": [{"username": "custos-tperson", "uid": 2000093, "unix_cluster_id": 1}],
      "Url": [],
      "SshKey": []
    }`

	merged, err := mergeIdentifier(json.RawMessage(person), "oidcsub", "http://idp.invalid/users/9", true)
	if err != nil {
		t.Fatalf("mergeIdentifier: %v", err)
	}

	var before, after map[string]json.RawMessage
	if err := json.Unmarshal([]byte(person), &before); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if err := json.Unmarshal(merged, &after); err != nil {
		t.Fatalf("decode merged: %v", err)
	}

	// Every section survives, and only Identifier is allowed to differ.
	for key, want := range before {
		got, ok := after[key]
		if !ok {
			t.Fatalf("PUT body dropped %q, the registry would delete it", key)
		}
		if key == "Identifier" {
			continue
		}
		if !sameJSON(t, got, want) {
			t.Errorf("%s changed\n got: %s\nwant: %s", key, got, want)
		}
	}
	for key := range after {
		if _, ok := before[key]; !ok {
			t.Errorf("PUT body invented %q", key)
		}
	}

	// Every identifier the person already had is still there, plus the new one.
	var idents []struct {
		Identifier string `json:"identifier"`
		Type       string `json:"type"`
		Login      bool   `json:"login"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(after["Identifier"], &idents); err != nil {
		t.Fatalf("decode identifiers: %v", err)
	}
	found := map[string]string{}
	for _, id := range idents {
		found[id.Type] = id.Identifier
		if id.Status != "A" {
			t.Errorf("%s lost its status: %+v", id.Type, id)
		}
	}
	for typ, want := range map[string]string{
		"comanage_id": "Person100099",
		"uidnumber":   "2000093",
		"gidnumber":   "2000093",
		"uid":         "custos-tperson",
		"oidcsub":     "http://idp.invalid/users/9",
	} {
		if found[typ] != want {
			t.Errorf("identifier %s = %q, want %q", typ, found[typ], want)
		}
	}
	if len(idents) != 5 {
		t.Errorf("identifier count = %d, want 5", len(idents))
	}
}

// A person whose sub changes must end with one oidcsub, not two.
func TestMergeIdentifier_ReplacesSameType(t *testing.T) {
	person := `{"CoPerson":{"co_id":2},"Identifier":[{"identifier":"old-sub","type":"oidcsub","login":true,"status":"A"}]}`
	merged, err := mergeIdentifier(json.RawMessage(person), "oidcsub", "new-sub", true)
	if err != nil {
		t.Fatalf("mergeIdentifier: %v", err)
	}
	var out struct {
		Identifier []struct {
			Identifier string `json:"identifier"`
			Type       string `json:"type"`
		} `json:"Identifier"`
	}
	if err := json.Unmarshal(merged, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Identifier) != 1 {
		t.Fatalf("want 1 oidcsub, got %d: %s", len(out.Identifier), merged)
	}
	if out.Identifier[0].Identifier != "new-sub" {
		t.Errorf("identifier = %q, want new-sub", out.Identifier[0].Identifier)
	}
}

// sameJSON compares two raw messages by value, since marshalling compacts
// whitespace and would make identical content look different.
func sameJSON(t *testing.T, a, b json.RawMessage) bool {
	t.Helper()
	var x, y interface{}
	if err := json.Unmarshal(a, &x); err != nil {
		t.Fatalf("decode a: %v", err)
	}
	if err := json.Unmarshal(b, &y); err != nil {
		t.Fatalf("decode b: %v", err)
	}
	return reflect.DeepEqual(x, y)
}
