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
// KIND, either express or implied. See the License for the
// specific language governing permissions and limitations
// under the License.

// Package db holds the LDAP Provisioner connector's private schema.
// Applied via internal/db.MigrateConnectorFS at loader time. Mirrors
// the pattern used by the AMIE-Processor connector.
package db

import "embed"

//go:embed migrations/*.sql
var migrationFS embed.FS

// MigrationFS exposes the embedded migrations to the host's migration runner.
func MigrationFS() embed.FS { return migrationFS }
