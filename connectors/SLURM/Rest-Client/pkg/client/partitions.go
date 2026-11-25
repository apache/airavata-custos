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

// Live partitions, read-only: slurm.conf owns them, Custos never writes them.
package client

import "strings"

type Partition struct {
	Name               string `json:"name"`
	Nodes              string `json:"nodes"`
	State              string `json:"state"`
	MaxTimeMinutes     *int64 `json:"max_time_minutes"`
	DefaultTimeMinutes *int64 `json:"default_time_minutes"`
}

type partitionWire struct {
	Name  string `json:"name"`
	Nodes struct {
		Configured string `json:"configured"`
	} `json:"nodes"`
	Partition struct {
		State []string `json:"state"`
	} `json:"partition"`
	Maximums struct {
		Time SlurmNumber `json:"time"`
	} `json:"maximums"`
	Defaults struct {
		Time SlurmNumber `json:"time"`
	} `json:"defaults"`
}

type partitionsResponse struct {
	Meta struct {
		Slurm struct {
			Cluster string `json:"cluster"`
		} `json:"slurm"`
	} `json:"meta"`
	Partitions []partitionWire `json:"partitions"`
}

// ListPartitions returns the partitions of cluster: none when slurmctld serves another.
// The cluster is read from the response meta; each partition's own cluster field is empty.
func (c *Client) ListPartitions(cluster string) ([]Partition, error) {
	var out partitionsResponse
	if _, err := c.do("GET", "/slurm/v0.0."+c.apiVersion+"/partitions/", nil, &out); err != nil {
		return nil, err
	}
	ps := make([]Partition, 0, len(out.Partitions))
	if out.Meta.Slurm.Cluster != cluster {
		return ps, nil
	}
	for _, w := range out.Partitions {
		ps = append(ps, Partition{
			Name:               w.Name,
			Nodes:              w.Nodes.Configured,
			State:              strings.Join(w.Partition.State, ","),
			MaxTimeMinutes:     ptrNum(&w.Maximums.Time),
			DefaultTimeMinutes: ptrNum(&w.Defaults.Time),
		})
	}
	return ps, nil
}
