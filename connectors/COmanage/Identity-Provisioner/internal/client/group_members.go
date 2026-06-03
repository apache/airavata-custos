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

type CoGroupMemberCreateRequest struct {
	RequestType    string                   `json:"RequestType"`
	Version        string                   `json:"Version"`
	CoGroupMembers []CoGroupMemberCreateOne `json:"CoGroupMembers"`
}

type CoGroupMemberCreateOne struct {
	Version   string           `json:"Version"`
	Person    IdentifierParent `json:"Person"`
	CoGroupId int              `json:"CoGroupId"`
	Member    bool             `json:"Member"`
	Owner     bool             `json:"Owner"`
}

func (c *Client) CreateCoGroupMember(coPersonId, coGroupId int) (int, error) {
	body, err := json.Marshal(CoGroupMemberCreateRequest{
		RequestType: "CoGroupMembers",
		Version:     restAPIVersion,
		CoGroupMembers: []CoGroupMemberCreateOne{{
			Version:   restAPIVersion,
			Person:    IdentifierParent{Type: "CO", Id: coPersonId},
			CoGroupId: coGroupId,
			Member:    true,
			Owner:     true,
		}},
	})
	if err != nil {
		return 0, fmt.Errorf("marshal co_group_member: %w", err)
	}

	u := c.restAPI("/co_group_members.json")
	resp, respBody, err := c.Do(http.MethodPost, u, body)
	if err != nil {
		return 0, err
	}
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK:
		var out CoGroupCreateResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return 0, fmt.Errorf("decode co_group_member response: %w: %s", err, string(respBody))
		}
		var id int
		if _, err := fmt.Sscanf(out.Id, "%d", &id); err != nil {
			return 0, fmt.Errorf("parse co_group_member id %q: %w", out.Id, err)
		}
		return id, nil
	case http.StatusUnauthorized:
		return 0, ErrAuth401
	default:
		return 0, &HTTPError{Method: "POST", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}

type CoGroupMemberListResponse struct {
	ResponseType   string                 `json:"ResponseType"`
	Version        string                 `json:"Version"`
	CoGroupMembers []CoGroupMemberListOne `json:"CoGroupMembers"`
}

type CoGroupMemberListOne struct {
	Version   string           `json:"Version"`
	Id        int              `json:"Id"`
	Person    IdentifierParent `json:"Person"`
	CoGroupId int              `json:"CoGroupId"`
	Member    bool             `json:"Member"`
	Owner     bool             `json:"Owner"`
}

// FindCoGroupMember returns the membership id for a (group, person) pair, or
// 0 if none.
func (c *Client) FindCoGroupMember(coGroupId, coPersonId int) (int, error) {
	u := c.restAPI(fmt.Sprintf("/co_group_members.json?cogroupid=%d&copersonid=%d", coGroupId, coPersonId))
	resp, respBody, err := c.Do(http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var out CoGroupMemberListResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return 0, fmt.Errorf("decode co_group_members list: %w", err)
		}
		for _, m := range out.CoGroupMembers {
			if m.CoGroupId == coGroupId && m.Person.Id == coPersonId {
				return m.Id, nil
			}
		}
		return 0, nil
	case http.StatusNoContent:
		return 0, nil
	case http.StatusUnauthorized:
		return 0, ErrAuth401
	default:
		return 0, &HTTPError{Method: "GET", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}
