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

// Package connectors is the registry through which external-system connectors
// (ACCESS AMIE, NAIRR, etc.) plug into the core binary.
//
// A deployment selects which connectors are bundled by importing them
// blank in cmd/server/connectors.go. Each connector's init() function calls
// Register, contributing both its embedded migrations and its Start function
// to the runtime.
package connectors

import (
	"context"
	"io/fs"
	"sort"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
)

// Deps is the bundle the host hands to every connector at startup.
type Deps struct {
	DB      *sqlx.DB
	Service *service.Service
	Bus     *events.Bus
}

// Connector is a long-running subsystem that integrates an external system
// with core. The host applies its migrations first, then calls Start.
type Connector interface {
	// Name returns a unique slug for this connector (used to namespace the
	// per-connector schema_migrations table).
	Name() string
	// Migrations returns the embedded migration tree and the subdirectory
	// within it that holds the .sql files. Return (nil, "") if the connector
	// has no migrations.
	Migrations() (fs.FS, string)
	// Start kicks off the connector. It must return promptly; long-running
	// work belongs in goroutines that respect ctx cancellation.
	Start(ctx context.Context, deps Deps) error
}

var (
	mu       sync.Mutex
	registry = map[string]Connector{}
)

// Register adds a connector to the global registry. Typically called from an
// init() function in the connector's package. Re-registration with the same
// Name panics — connector names must be unique.
func Register(c Connector) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := registry[c.Name()]; ok {
		panic("connectors: duplicate registration for " + c.Name())
	}
	registry[c.Name()] = c
}

// All returns the registered connectors in deterministic (name-sorted) order.
func All() []Connector {
	mu.Lock()
	defer mu.Unlock()
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]Connector, 0, len(names))
	for _, n := range names {
		out = append(out, registry[n])
	}
	return out
}
