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
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateOrgIdentity POSTs a bare org identity in the configured CO and
// returns its id.
func (c *Client) CreateOrgIdentity() (int, error) {
	body, err := json.Marshal(map[string]interface{}{
		"RequestType": "OrgIdentities",
		"Version":     restAPIVersion,
		"OrgIdentities": []map[string]interface{}{
			{"Version": restAPIVersion, "CoId": c.cfg.COID},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("marshal org identity: %w", err)
	}
	u := c.restAPI("/org_identities.json")
	resp, respBody, err := c.Do(http.MethodPost, u, body)
	if err != nil {
		return 0, err
	}
	return decodeNewObjectID(resp, respBody, "POST", u)
}

// CreateIdentifierOnOrgIdentity POSTs an Identifier attached to an org
// identity. Login-identity types the registry refuses at person scope are
// accepted here.
func (c *Client) CreateIdentifierOnOrgIdentity(value, identifierType string, orgIdentityID int, login bool) (int, error) {
	body, err := json.Marshal(IdentifierCreateRequest{
		RequestType: "Identifiers",
		Version:     restAPIVersion,
		Identifiers: []IdentifierCreateOne{{
			Version:    restAPIVersion,
			Identifier: value,
			Type:       identifierType,
			Login:      login,
			Status:     "Active",
			Person:     IdentifierParent{Type: "Org", Id: orgIdentityID},
		}},
	})
	if err != nil {
		return 0, fmt.Errorf("marshal identifier: %w", err)
	}
	u := c.restAPI("/identifiers.json")
	resp, respBody, err := c.Do(http.MethodPost, u, body)
	if err != nil {
		return 0, err
	}
	return decodeNewObjectID(resp, respBody, "POST", u)
}

// CreateCoOrgIdentityLink attaches an org identity to a CoPerson.
func (c *Client) CreateCoOrgIdentityLink(coPersonID, orgIdentityID int) (int, error) {
	body, err := json.Marshal(map[string]interface{}{
		"RequestType": "CoOrgIdentityLinks",
		"Version":     restAPIVersion,
		"CoOrgIdentityLinks": []map[string]interface{}{
			{"Version": restAPIVersion, "CoPersonId": coPersonID, "OrgIdentityId": orgIdentityID},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("marshal org identity link: %w", err)
	}
	u := c.restAPI("/co_org_identity_links.json")
	resp, respBody, err := c.Do(http.MethodPost, u, body)
	if err != nil {
		return 0, err
	}
	return decodeNewObjectID(resp, respBody, "POST", u)
}

func decodeNewObjectID(resp *http.Response, respBody []byte, method, u string) (int, error) {
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK:
		var out CoGroupCreateResponse // NewObject shape
		if err := json.Unmarshal(respBody, &out); err != nil {
			return 0, fmt.Errorf("decode create response: %w: %s", err, string(respBody))
		}
		var id int
		if _, err := fmt.Sscanf(out.Id, "%d", &id); err != nil {
			return 0, fmt.Errorf("parse object id %q: %w", out.Id, err)
		}
		return id, nil
	case http.StatusUnauthorized:
		return 0, ErrAuth401
	default:
		return 0, &HTTPError{Method: method, URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}
