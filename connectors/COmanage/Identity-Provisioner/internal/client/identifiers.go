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

// IdentifierParent is the nested {"Type":"Group"|"CO","Id":<n>} block that
// attaches an Identifier to a CoGroup or CoPerson.
type IdentifierParent struct {
	Type string `json:"Type"`
	Id   int    `json:"Id"`
}

type IdentifierCreateRequest struct {
	RequestType string                `json:"RequestType"`
	Version     string                `json:"Version"`
	Identifiers []IdentifierCreateOne `json:"Identifiers"`
}

type IdentifierCreateOne struct {
	Version    string           `json:"Version"`
	Identifier string           `json:"Identifier"`
	Type       string           `json:"Type"`
	Login      bool             `json:"Login"`
	Status     string           `json:"Status"`
	Person     IdentifierParent `json:"Person"`
}

// CreateIdentifierOnGroup POSTs an Identifier attached to a CoGroup.
func (c *Client) CreateIdentifierOnGroup(value, identifierType string, coGroupId int) (int, error) {
	body, err := json.Marshal(IdentifierCreateRequest{
		RequestType: "Identifiers",
		Version:     restAPIVersion,
		Identifiers: []IdentifierCreateOne{{
			Version:    restAPIVersion,
			Identifier: value,
			Type:       identifierType,
			Login:      false,
			Status:     "Active",
			Person:     IdentifierParent{Type: "Group", Id: coGroupId},
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
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK:
		var out CoGroupCreateResponse // same NewObject shape
		if err := json.Unmarshal(respBody, &out); err != nil {
			return 0, fmt.Errorf("decode identifier create response: %w: %s", err, string(respBody))
		}
		var id int
		if _, err := fmt.Sscanf(out.Id, "%d", &id); err != nil {
			return 0, fmt.Errorf("parse identifier id %q: %w", out.Id, err)
		}
		return id, nil
	case http.StatusUnauthorized:
		return 0, ErrAuth401
	default:
		return 0, &HTTPError{Method: "POST", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}

type IdentifierListResponse struct {
	ResponseType string              `json:"ResponseType"`
	Version      string              `json:"Version"`
	Identifiers  []IdentifierListOne `json:"Identifiers"`
}

type IdentifierListOne struct {
	Version    string `json:"Version"`
	Id         int    `json:"Id"`
	Identifier string `json:"Identifier"`
	Type       string `json:"Type"`
	Status     string `json:"Status"`
}

// FindIdentifierOnGroup returns the existing Identifier id with the given
// type on a CoGroup, or 0 if none.
func (c *Client) FindIdentifierOnGroup(coGroupId int, identifierType string) (int, error) {
	u := c.restAPI(fmt.Sprintf("/identifiers.json?cogroupid=%d", coGroupId))
	resp, respBody, err := c.Do(http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var out IdentifierListResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return 0, fmt.Errorf("decode identifiers list: %w", err)
		}
		for _, ident := range out.Identifiers {
			if ident.Type == identifierType {
				return ident.Id, nil
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

// FindIdentifierValueOnPerson returns the Identifier *value* (e.g. "Custos100022")
// of the given type on a CoPerson, or "" if none is assigned yet.
func (c *Client) FindIdentifierValueOnPerson(coPersonID int, identifierType string) (string, error) {
	u := c.restAPI(fmt.Sprintf("/identifiers.json?copersonid=%d", coPersonID))
	resp, respBody, err := c.Do(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var out IdentifierListResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return "", fmt.Errorf("decode identifiers list: %w", err)
		}
		for _, ident := range out.Identifiers {
			if ident.Type == identifierType {
				return ident.Identifier, nil
			}
		}
		return "", nil
	case http.StatusNoContent:
		return "", nil
	case http.StatusUnauthorized:
		return "", ErrAuth401
	default:
		return "", &HTTPError{Method: "GET", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}
