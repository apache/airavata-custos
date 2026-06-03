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

type UnixClusterGroupCreateRequest struct {
	RequestType       string                      `json:"RequestType"`
	Version           string                      `json:"Version"`
	UnixClusterGroups []UnixClusterGroupCreateOne `json:"UnixClusterGroups"`
}

type UnixClusterGroupCreateOne struct {
	Version       string `json:"Version"`
	UnixClusterId int    `json:"UnixClusterId"`
	CoGroupId     int    `json:"CoGroupId"`
}

// CreateUnixClusterGroup binds a CoGroup to a UnixCluster. The URL takes no
// named params for POST.
func (c *Client) CreateUnixClusterGroup(coGroupId int) (int, error) {
	body, err := json.Marshal(UnixClusterGroupCreateRequest{
		RequestType: "UnixClusterGroups",
		Version:     restAPIVersion,
		UnixClusterGroups: []UnixClusterGroupCreateOne{{
			Version:       restAPIVersion,
			UnixClusterId: c.cfg.UnixClusterID,
			CoGroupId:     coGroupId,
		}},
	})
	if err != nil {
		return 0, fmt.Errorf("marshal unix_cluster_group: %w", err)
	}

	u := c.restAPI("/unix_cluster/unix_cluster_groups.json")
	resp, respBody, err := c.Do(http.MethodPost, u, body)
	if err != nil {
		return 0, err
	}
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK:
		var out CoGroupCreateResponse // NewObject shape
		if err := json.Unmarshal(respBody, &out); err != nil {
			return 0, fmt.Errorf("decode unix_cluster_group response: %w: %s", err, string(respBody))
		}
		var id int
		if _, err := fmt.Sscanf(out.Id, "%d", &id); err != nil {
			return 0, fmt.Errorf("parse unix_cluster_group id %q: %w", out.Id, err)
		}
		return id, nil
	case http.StatusUnauthorized:
		return 0, ErrAuth401
	default:
		return 0, &HTTPError{Method: "POST", URL: u, StatusCode: resp.StatusCode, Body: string(respBody)}
	}
}

func (c *Client) DeleteUnixClusterGroup(id int) error {
	u := c.restAPI(fmt.Sprintf("/unix_cluster/unix_cluster_groups/%d.json", id))
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
