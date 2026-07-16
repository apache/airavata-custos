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

//go:build e2e && live

// Live gate: the tier-1 flow once against the REAL registry. One-shot by
// design; every run creates a real person entry that stays behind.
package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/internal/config"
)

func TestAccessRequestLiveGate(t *testing.T) {
	registryURL := os.Getenv("COMANAGE_REGISTRY_URL")
	coID, _ := strconv.Atoi(os.Getenv("COMANAGE_CO_ID"))
	apiUser := os.Getenv("COMANAGE_API_USER")
	apiKey := os.Getenv("COMANAGE_API_KEY")
	unixClusterID, _ := strconv.Atoi(os.Getenv("COMANAGE_UNIX_CLUSTER_ID"))
	personIDType := os.Getenv("COMANAGE_PERSON_ID_TYPE")
	if registryURL == "" || coID == 0 || apiUser == "" || apiKey == "" || unixClusterID == 0 || personIDType == "" {
		t.Skip("live gate needs COMANAGE_REGISTRY_URL/CO_ID/API_USER/API_KEY/UNIX_CLUSTER_ID/PERSON_ID_TYPE")
	}
	os.Setenv("POSIX_USERNAME_PREFIX", "nexus")

	admin := kcAdminToken(t)
	database := openTestDB(t)
	resetAndSeed(t, database)
	ensureE2EClient(t, admin)

	cc := &config.ConnectorConfig{
		Type:    "comanage-identity-provisioner",
		Enabled: true,
		Config: map[string]interface{}{
			"registry": map[string]interface{}{
				"url": registryURL, "co_id": coID, "api_user": apiUser, "api_key": apiKey,
			},
			"unix_cluster": map[string]interface{}{
				"id": unixClusterID, "person_id_type": personIDType,
			},
			"provisioning": map[string]interface{}{
				"custos_cluster_id": clusterID, "default_shell": "/bin/bash",
				"homedir_prefix": "/home/", "http_timeout": "30s",
			},
		},
	}
	api := bootBackend(t, database, cc)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1e8)
	userA := kcNewUser(t, admin, "e2e-live-"+suffix, "Live", "Gate"+suffix)
	tokenA := ropcToken(t, userA.username)
	subA := jwtSub(t, tokenA)

	reqBody := map[string]string{"institution": "Live Gate University", "event_code": "PEARC26", "reason": "live gate"}
	status, body := call(t, api, http.MethodPost, "/access-requests", tokenA, reqBody)
	if status != http.StatusCreated {
		t.Fatalf("POST access request: want 201, got %d %s", status, body)
	}
	var req accessRequest
	mustDecode(t, body, &req)

	approver := kcNewUser(t, admin, "e2e-live-appr-"+suffix, "Live", "Approver"+suffix)
	tokenAppr := ropcToken(t, approver.username)
	bindIdentity(t, database, "pearc26-approver", jwtSub(t, tokenAppr), approver.email)

	status, body = call(t, api, http.MethodPut, "/access-requests/"+req.ID, tokenAppr, map[string]string{"status": "APPROVED"})
	if status != http.StatusOK {
		t.Fatalf("PUT APPROVED: want 200, got %d %s", status, body)
	}
	var approved accessRequest
	mustDecode(t, body, &approved)
	assertApprovedState(t, database, req.ID, subA, userA.email)
	username := clusterUsername(t, database, approved.CreatedUserID)
	t.Logf("LIVE: approved request %s, portal user %s, cluster username %s, email %s",
		req.ID, approved.CreatedUserID, username, userA.email)

	// Independent registry reads, polled: the subscriber runs async and the
	// registry itself may lag.
	get := func(u string) (int, []byte) {
		httpReq, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			t.Fatalf("build registry request: %v", err)
		}
		httpReq.SetBasicAuth(apiUser, apiKey)
		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			t.Fatalf("registry GET %s: %v", u, err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, b
	}

	var coPersonID int
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		code, b := get(fmt.Sprintf("%s/co_people.json?coid=%d&search.mail=%s", registryURL, coID, url.QueryEscape(userA.email)))
		if code == http.StatusOK {
			var list struct {
				CoPeople []struct {
					Id     int    `json:"Id"`
					Status string `json:"Status"`
				} `json:"CoPeople"`
			}
			if err := json.Unmarshal(b, &list); err == nil && len(list.CoPeople) > 0 {
				coPersonID = list.CoPeople[0].Id
				t.Logf("LIVE: registry CoPerson id=%d status=%s", coPersonID, list.CoPeople[0].Status)
				break
			}
		}
		time.Sleep(5 * time.Second)
	}
	if coPersonID == 0 {
		t.Fatalf("registry never showed a CoPerson for %s within 90s", userA.email)
	}

	// The composite endpoint keys on the person_id_type identifier, not the
	// numeric person id.
	code, b := get(fmt.Sprintf("%s/identifiers.json?copersonid=%d", registryURL, coPersonID))
	if code != http.StatusOK {
		t.Fatalf("identifiers read for CoPerson %d: status %d body %s", coPersonID, code, string(b))
	}
	var ids struct {
		Identifiers []struct {
			Type       string `json:"Type"`
			Identifier string `json:"Identifier"`
		} `json:"Identifiers"`
	}
	if err := json.Unmarshal(b, &ids); err != nil {
		t.Fatalf("decode identifiers: %v", err)
	}
	personIdent := ""
	for _, id := range ids.Identifiers {
		if id.Type == personIDType {
			personIdent = id.Identifier
		}
	}
	if personIdent == "" {
		t.Fatalf("CoPerson %d has no identifier of type %q: %+v", coPersonID, personIDType, ids.Identifiers)
	}

	code, b = get(fmt.Sprintf("%s/api/co/%d/core/v1/people/%s", registryURL, coID, url.PathEscape(personIdent)))
	if code != http.StatusOK {
		t.Fatalf("composite read for CoPerson %d: status %d body %s", coPersonID, code, string(b))
	}
	composite := string(b)
	for _, want := range []string{"UnixClusterAccount", username} {
		if !strings.Contains(composite, want) {
			t.Errorf("composite for CoPerson %d missing %q", coPersonID, want)
		}
	}
	if idx := strings.Index(composite, "UnixClusterAccount"); idx >= 0 {
		end := idx + 600
		if end > len(composite) {
			end = len(composite)
		}
		t.Logf("LIVE: composite UnixClusterAccount excerpt: %s", composite[idx:end])
	}
	t.Logf("LIVE GATE PASS: CoPerson %d carries a UnixClusterAccount for %s; check `id %s` on the cluster next", coPersonID, username, username)
}
