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
	"net/url"
)

// CreatePersonResponseEntry is one Identifier echoed by POST /people. COmanage
// returns one row per identifier type configured on the binding, typically
// including the auto-assigned ids.
type CreatePersonResponseEntry struct {
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
	Login      bool   `json:"login"`
	Status     string `json:"status"`
}

// CreatePerson POSTs the composite body (CoPerson + Name + Email +
// Identifier blocks) to /people and returns the resulting Identifier list.
func (c *Client) CreatePerson(body []byte) ([]CreatePersonResponseEntry, error) {
	u := c.coreAPI("/people")
	resp, respBody, err := c.Do(http.MethodPost, u, body)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK:
		var out []CreatePersonResponseEntry
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, fmt.Errorf("decode create-person response: %w: %s", err, string(respBody))
		}
		return out, nil
	case http.StatusUnauthorized:
		return nil, ErrAuth401
	default:
		return nil, &HTTPError{Method: "POST", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}

// GetPersonComposite returns the raw composite JSON, suitable for round-tripping
// back into UpdatePerson.
func (c *Client) GetPersonComposite(identifier string) (json.RawMessage, error) {
	u := c.coreAPI("/people/" + identifier)
	resp, respBody, err := c.Do(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return json.RawMessage(respBody), nil
	case http.StatusUnauthorized:
		return nil, ErrAuth401
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, &HTTPError{Method: "GET", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}

// UpdatePerson PUTs a full-composite body. The body MUST include every related
// model (Name, EmailAddress, Identifier, etc.) or COmanage's deleteOmitted
// behavior will wipe them.
func (c *Client) UpdatePerson(identifier string, body []byte) (json.RawMessage, error) {
	u := c.coreAPI("/people/" + identifier)
	resp, respBody, err := c.Do(http.MethodPut, u, body)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return json.RawMessage(respBody), nil
	case http.StatusUnauthorized:
		return nil, ErrAuth401
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, &HTTPError{Method: "PUT", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}

func (c *Client) DeletePerson(identifier string) error {
	u := c.coreAPI("/people/" + identifier)
	resp, respBody, err := c.Do(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return ErrAuth401
	case http.StatusNotFound:
		return ErrNotFound
	default:
		return &HTTPError{Method: "DELETE", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}

type CoPersonListResponse struct {
	ResponseType string            `json:"ResponseType"`
	Version      string            `json:"Version"`
	CoPeople     []CoPersonListOne `json:"CoPeople"`
}

type CoPersonListOne struct {
	Version string `json:"Version"`
	Id      int    `json:"Id"`
	CoId    int    `json:"CoId"`
}

// FindCoPersonByEmail searches for CoPeople. search.mail is a LIKE match in
// COmanage; callers must post-filter for exact equality before trusting a hit.
func (c *Client) FindCoPersonByEmail(email string) ([]CoPersonListOne, error) {
	u := c.restAPI(fmt.Sprintf("/co_people.json?coid=%d&search.mail=%s", c.cfg.COID, url.QueryEscape(email)))
	resp, respBody, err := c.Do(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var out CoPersonListResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, fmt.Errorf("decode co_people list: %w", err)
		}
		return out.CoPeople, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusUnauthorized:
		return nil, ErrAuth401
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, &HTTPError{Method: "GET", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}
