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

package airavata

import "testing"

// Only a CILogon subject names an Airavata user.
func TestAiravataUserID(t *testing.T) {
	for sub, want := range map[string]string{
		"http://cilogon.org/serverE/users/194529": "cilogon:194529",
		"http://cilogon.org/serverE/users/":       "",
		"https://accounts.example.org/u/194529":   "",
	} {
		if got, ok := airavataUserID(sub); got != want || ok != (want != "") {
			t.Errorf("%q: got %q, want %q", sub, got, want)
		}
	}
}
