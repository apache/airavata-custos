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

// CoGroupCreateRequest is the body envelope for POST /co_groups.json.
// GroupType "CL" = GroupEnum::Clusters (per-user cluster primary group).
type CoGroupCreateRequest struct {
	RequestType string             `json:"RequestType"`
	Version     string             `json:"Version"`
	CoGroups    []CoGroupCreateOne `json:"CoGroups"`
}

type CoGroupCreateOne struct {
	Version     string `json:"Version"`
	CoId        int    `json:"CoId"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Open        bool   `json:"Open"`
	Status      string `json:"Status"`
	GroupType   string `json:"GroupType"`
	Auto        bool   `json:"Auto"`
}

// CoGroupCreateResponse is the standard NewObject shape returned by COmanage
// REST POSTs.
type CoGroupCreateResponse struct {
	ResponseType string `json:"ResponseType"`
	Version      string `json:"Version"`
	ObjectType   string `json:"ObjectType"`
	Id           string `json:"Id"`
}

func (c *Client) CreateCoGroup(name, description string) (int, error) {
	body, err := json.Marshal(CoGroupCreateRequest{
		RequestType: "CoGroups",
		Version:     restAPIVersion,
		CoGroups: []CoGroupCreateOne{{
			Version:     restAPIVersion,
			CoId:        c.cfg.COID,
			Name:        name,
			Description: description,
			Open:        false,
			Status:      "Active",
			GroupType:   "CL",
			Auto:        false,
		}},
	})
	if err != nil {
		return 0, fmt.Errorf("marshal co_group: %w", err)
	}

	u := c.restAPI("/co_groups.json")
	resp, respBody, err := c.Do(http.MethodPost, u, body)
	if err != nil {
		return 0, err
	}
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK:
		var out CoGroupCreateResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return 0, fmt.Errorf("decode co_group create response: %w: %s", err, string(respBody))
		}
		var id int
		if _, err := fmt.Sscanf(out.Id, "%d", &id); err != nil {
			return 0, fmt.Errorf("parse co_group id %q: %w", out.Id, err)
		}
		return id, nil
	case http.StatusUnauthorized:
		return 0, ErrAuth401
	default:
		return 0, &HTTPError{Method: "POST", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}

type CoGroupListResponse struct {
	ResponseType string           `json:"ResponseType"`
	Version      string           `json:"Version"`
	CoGroups     []CoGroupReadOne `json:"CoGroups"`
}

type CoGroupReadOne struct {
	Version string `json:"Version"`
	Id      int    `json:"Id"`
	CoId    int    `json:"CoId"`
	Name    string `json:"Name"`
}

// FindCoGroupByName returns the group id, or 0 if no group with that name
// exists in the configured CO.
func (c *Client) FindCoGroupByName(name string) (int, error) {
	u := c.restAPI(fmt.Sprintf("/co_groups.json?coid=%d", c.cfg.COID))
	resp, respBody, err := c.Do(http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var out CoGroupListResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return 0, fmt.Errorf("decode co_groups list: %w", err)
		}
		for _, g := range out.CoGroups {
			if g.Name == name {
				return g.Id, nil
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

func (c *Client) DeleteCoGroup(id int) error {
	u := c.restAPI(fmt.Sprintf("/co_groups/%d.json", id))
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
