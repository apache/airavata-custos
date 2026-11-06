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

package db

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"regexp"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

// retiredTopicsQueued finds pending work the new handlers would drop: retired topics, and registry deletes that no longer resync groups.
const retiredTopicsQueued = `SELECT EXISTS (SELECT 1 FROM event_deliveries d JOIN events e ON e.id = d.event_id
	WHERE d.status = 'PENDING'
	  AND EXISTS (SELECT 1 FROM event_subscriptions s WHERE s.subscriber = d.subscriber AND s.event_type = e.event_type)
	  AND (e.event_type IN ('compute_allocation::create', 'compute_allocation::update', 'compute_cluster_user::provision',
	                        'compute_allocation_membership::create', 'compute_allocation_membership::update', 'compute_allocation_membership::delete')
	    OR (d.subscriber = 'comanage-identity-provisioner' AND e.event_type IN ('compute_cluster_user::approve', 'user::onboarded', 'compute_cluster_user::delete'))))`

// MigrateEmbedded applies all pending migrations from the embedded
// migrations/ directory against the supplied database.
// Returns nil when there is nothing to apply.
func MigrateEmbedded(database *sqlx.DB) error {
	driver, err := pgxmigrate.WithInstance(database.DB, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	source, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "pgx", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	// Checked before Up, which marks a failed migration dirty and so stops the previous version from draining.
	if v, dirty, err := m.Version(); err == nil && !dirty && v >= 7 && v < 9 {
		var queued bool
		if err := database.Get(&queued, retiredTopicsQueued); err != nil {
			return fmt.Errorf("check retired topics: %w", err)
		}
		if queued {
			return errors.New("connector deliveries are pending on topics migration 9 retires: let the running version finish them, then stop it and upgrade")
		}
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	slog.Info("database migrations applied successfully")
	return nil
}

// connectorNamePattern restricts connector names to a safe identifier shape so
// the value can be concatenated into a SQL table name (schema_migrations_<name>)
// without escaping concerns.
var connectorNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,30}$`)

// MigrateConnectorFS applies pending migrations from src against the database,
// tracking version state in schema_migrations_<name> so each connector
// versions its schema independently of core and of other connectors.
func MigrateConnectorFS(database *sqlx.DB, src fs.FS, dir, name string) error {
	if database == nil {
		return errors.New("database is required")
	}
	if src == nil {
		return errors.New("migration source is required")
	}
	if dir == "" {
		return errors.New("migration directory is required")
	}
	if !connectorNamePattern.MatchString(name) {
		return fmt.Errorf("connector name %q is invalid; must match %s", name, connectorNamePattern)
	}

	driver, err := pgxmigrate.WithInstance(database.DB, &pgxmigrate.Config{
		MigrationsTable: "schema_migrations_" + name,
	})
	if err != nil {
		return fmt.Errorf("create migration driver for %s: %w", name, err)
	}

	source, err := iofs.New(src, dir)
	if err != nil {
		return fmt.Errorf("create migration source for %s: %w", name, err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "pgx", driver)
	if err != nil {
		return fmt.Errorf("create migrator for %s: %w", name, err)
	}

	before, _, _ := m.Version()
	upErr := m.Up()
	after, _, _ := m.Version()
	if upErr != nil && !errors.Is(upErr, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations for %s: %w", name, upErr)
	}
	if errors.Is(upErr, migrate.ErrNoChange) {
		slog.Info("connector migrations already up to date", "connector", name, "version", after)
		return nil
	}
	slog.Info("connector migrations applied", "connector", name, "from", before, "to", after)
	return nil
}
