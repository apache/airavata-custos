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

package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/store"
)

// Service is a high-level façade over the underlying stores. It wraps each
// mutating operation in a transaction so callers do not need to manage
// *sql.Tx themselves.
type Service struct {
	db       *sqlx.DB
	orgs     store.OrganizationStore
	users    store.UserStore
	projs    store.ProjectStore
	clusters store.ComputeClusterStore
}

// New constructs a Service backed by the supplied database handle.
// Stores are instantiated internally using the default MySQL implementations.
func New(database *sqlx.DB) *Service {
	return &Service{
		db:       database,
		orgs:     store.NewOrganizationStore(database),
		users:    store.NewUserStore(database),
		projs:    store.NewProjectStore(database),
		clusters: store.NewComputeClusterStore(database),
	}
}

// NewWithStores constructs a Service from explicit stores. Useful for tests
// within this module — stores are an internal type and cannot be supplied by
// external callers.
func NewWithStores(database *sqlx.DB, orgs store.OrganizationStore, users store.UserStore, projs store.ProjectStore, clusters store.ComputeClusterStore) *Service {
	return &Service{db: database, orgs: orgs, users: users, projs: projs, clusters: clusters}
}

// inTx runs fn inside a database transaction managed by the Service.
func (s *Service) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return db.TxFn(ctx, s.db, fn)
}

// newID returns a new identifier for a freshly created entity.
func newID() string {
	return uuid.NewString()
}

// nowUTC returns the current time in UTC.
func nowUTC() time.Time {
	return time.Now().UTC()
}
