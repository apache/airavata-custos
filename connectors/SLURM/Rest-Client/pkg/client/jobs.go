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

/*
curl -s -X GET \
  "http://localhost:6820/slurmdb/v0.0.41/jobs" \
  -H "X-SLURM-USER-NAME: root" \
  -H "X-SLURM-USER-TOKEN: $SLURM_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "users": ["root"],
    "start_time": {
      "set": true,
      "infinite": false,
      "number": 1746057600
    }
  }'
*/

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type jobsResponse struct {
	Jobs []JobInfo `json:"jobs"`
}

type JobFilter struct {
	Users     []string `json:"users,omitempty"`
	StartTime int64    `json:"start_time,omitempty"`
	EndTime   int64    `json:"end_time,omitempty"`
}

// slurmQueryTime formats a unix timestamp as the naive UTC datetime string
// slurmdbd's query parser wants.
func slurmQueryTime(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("2006-01-02T15:04:05")
}

func (c *Client) ListJobs(filter JobFilter) ([]JobInfo, error) {
	var out jobsResponse
	params := url.Values{}
	if len(filter.Users) > 0 {
		params.Set("users", strings.Join(filter.Users, ","))
	}
	if filter.StartTime > 0 {
		params.Set("start_time", slurmQueryTime(filter.StartTime))
	}
	if filter.EndTime > 0 {
		params.Set("end_time", slurmQueryTime(filter.EndTime))
	}
	path := "/slurmdb/v0.0." + c.apiVersion + "/jobs"
	if enc := params.Encode(); enc != "" {
		path += "?" + enc
	}
	if _, err := c.do("GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Jobs, nil
}

func (c *Client) GetJob(id int64) (*JobInfo, error) {
	var out jobsResponse
	if _, err := c.do("GET", fmt.Sprintf("/slurmdb/v0.0.%s/job/%d", c.apiVersion, id), nil, &out); err != nil {
		return nil, err
	}
	if len(out.Jobs) == 0 {
		return nil, fmt.Errorf("job %d not found", id)
	}
	return &out.Jobs[0], nil
}

func (c *Client) SubmitJob(request JobSubmitRequest) (*JobSubmitResponse, error) {
	var out JobSubmitResponse
	if _, err := c.do("POST", fmt.Sprintf("/slurm/v0.0.%s/job/submit", c.apiVersion), request, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
