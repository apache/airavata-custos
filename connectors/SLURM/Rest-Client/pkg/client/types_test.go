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
	"testing"
)

func TestJobInfoDecodes(t *testing.T) {
	raw := `{"account":"proj-a","user":"user-a","partition":"debug",
        "job_id":7,"time":{"start":1000,"end":4600},
        "exit_code":{"status":["SUCCESS"],"return_code":{"set":true,"infinite":false,"number":0}},
        "tres":{"allocated":[{"type":"cpu","count":2},{"type":"node","count":1}]}}`
	var j JobInfo
	if err := json.Unmarshal([]byte(raw), &j); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(j.ExitCode.Status) != 1 || j.ExitCode.Status[0] != "SUCCESS" {
		t.Fatalf("status: %+v", j.ExitCode.Status)
	}
	if !j.ExitCode.ReturnCode.Set || j.ExitCode.ReturnCode.Number != 0 {
		t.Fatalf("return_code: %+v", j.ExitCode.ReturnCode)
	}
	if j.Time.End-j.Time.Start != 3600 || j.Tres.Allocated[0].Count != 2 {
		t.Fatalf("core fields: %+v", j)
	}
}
