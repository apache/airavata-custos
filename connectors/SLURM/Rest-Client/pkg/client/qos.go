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

// Custos-owned QoS definitions as slurmdb stores them. Limits sit in nested
// {set, infinite, number} wrappers on the wire; unset and infinite map to nil.
package client

type QoS struct {
	Name           string   `json:"name"`
	Priority       int64    `json:"priority"`
	MaxWallMinutes *int64   `json:"max_wall_minutes"`
	MaxJobsPerUser *int64   `json:"max_jobs_per_user"`
	MaxTRESPerUser []TRES   `json:"max_tres_per_user"`
	Flags          []string `json:"flags"`
}

const qosPath = "/qos/"

type qosWire struct {
	Name     string      `json:"name"`
	Priority SlurmNumber `json:"priority"`
	Flags    []string    `json:"flags"`
	Limits   struct {
		Max struct {
			Jobs struct {
				ActiveJobs struct {
					Per struct {
						User SlurmNumber `json:"user"`
					} `json:"per"`
				} `json:"active_jobs"`
			} `json:"jobs"`
			TRES struct {
				Per struct {
					User []TRES `json:"user"`
				} `json:"per"`
			} `json:"tres"`
			WallClock struct {
				Per struct {
					Job SlurmNumber `json:"job"`
				} `json:"per"`
			} `json:"wall_clock"`
		} `json:"max"`
	} `json:"limits"`
}

type qosResponse struct {
	QoS []qosWire `json:"qos"`
}

func (q QoS) wire() qosWire {
	w := qosWire{
		Name:     q.Name,
		Priority: SlurmNumber{Set: true, Number: q.Priority},
		Flags:    q.Flags,
	}
	if w.Flags == nil {
		w.Flags = []string{}
	}
	w.Limits.Max.WallClock.Per.Job = *limitOrInfinite(q.MaxWallMinutes)
	w.Limits.Max.Jobs.ActiveJobs.Per.User = *limitOrInfinite(q.MaxJobsPerUser)
	w.Limits.Max.TRES.Per.User = q.MaxTRESPerUser
	if w.Limits.Max.TRES.Per.User == nil {
		w.Limits.Max.TRES.Per.User = []TRES{}
	}
	return w
}

func (w qosWire) qos() QoS {
	return QoS{
		Name:           w.Name,
		Priority:       w.Priority.Number,
		MaxWallMinutes: ptrNum(&w.Limits.Max.WallClock.Per.Job),
		MaxJobsPerUser: ptrNum(&w.Limits.Max.Jobs.ActiveJobs.Per.User),
		MaxTRESPerUser: w.Limits.Max.TRES.Per.User,
		Flags:          w.Flags,
	}
}

func (c *Client) ListQoS() ([]QoS, error) {
	var out qosResponse
	if _, err := c.do("GET", "/slurmdb/v0.0."+c.apiVersion+qosPath, nil, &out); err != nil {
		return nil, err
	}
	qs := make([]QoS, 0, len(out.QoS))
	for _, w := range out.QoS {
		qs = append(qs, w.qos())
	}
	return qs, nil
}

// PostQoS creates the QoS or replaces its priority, limits and flags; slurmdbd keeps an omitted TRES and clears one at count -1.
func (c *Client) PostQoS(q QoS) error {
	_, err := c.do("POST", "/slurmdb/v0.0."+c.apiVersion+qosPath, qosResponse{QoS: []qosWire{q.wire()}}, nil)
	return err
}

func (c *Client) DeleteQoS(name string) error {
	_, err := c.do("DELETE", "/slurmdb/v0.0."+c.apiVersion+qosPath+name, nil, nil)
	return err
}
