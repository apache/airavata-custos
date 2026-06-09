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

package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/apache/airavata-custos/pkg/service"
)

const bootstrapAdminEmailEnv = "CUSTOS_BOOTSTRAP_ADMIN_EMAIL"

// tryBootstrap runs the super_admin bootstrap if the operator set
// CUSTOS_BOOTSTRAP_ADMIN_EMAIL. Idempotent: skips quietly when no env value
// is set, the user does not exist, or super_admin already has a holder.
// A bootstrap failure never blocks server start; the warning surfaces the
// issue without crashing.
func tryBootstrap(ctx context.Context, svc *service.Service) {
	email := os.Getenv(bootstrapAdminEmailEnv)
	if email == "" {
		return
	}
	if err := svc.BootstrapSuperAdmin(ctx, email, "env:"+bootstrapAdminEmailEnv); err != nil {
		slog.Warn("bootstrap super_admin failed", "email", email, "error", err)
	}
}
