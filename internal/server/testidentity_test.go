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

package server

import (
	"net/http"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// withTestCaller installs a *identity.Caller plus a privilege set onto the
// request context for handler tests that run without the auth middleware.
func withTestCaller(r *http.Request, userID string, privs ...models.PrivilegeKey) *http.Request {
	ctx := identity.WithCaller(r.Context(), &identity.Caller{UserID: userID})
	ctx = identity.WithPrivilegesForTest(ctx, privs)
	return r.WithContext(ctx)
}
