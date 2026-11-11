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

package models

import (
	"slices"
	"strings"
	"sync"
	"time"
)

type AppStatus string

const (
	AppActive   AppStatus = "ACTIVE"
	AppDisabled AppStatus = "DISABLED"
)

// App is a trusted tool that calls connector routes with its own token.
type App struct {
	ID          string    `json:"id"          db:"id"`
	Key         string    `json:"key"         db:"key"`
	Name        string    `json:"name"        db:"name"`
	Description string    `json:"description" db:"description"`
	Status      AppStatus `json:"status"      db:"status"`
	SecretHash  string    `json:"-"           db:"secret_hash"`
	SecretHint  string    `json:"secret_hint" db:"secret_hint"`
	CreatedBy   string    `json:"created_by"  db:"created_by"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`
	Actions     []string  `json:"actions"     db:"-"`
}

// AppConnector is what apps can do on one connector. Action keys have the shape
// <connector>:app:<action>. A user might have to authorize the whole app.
type AppConnector struct {
	Name           string   `json:"name"`
	NeedsUserGrant bool     `json:"needs_user_grant"`
	Actions        []string `json:"actions"`
}

var (
	appConnectorsMu sync.RWMutex
	appConnectors   = map[string]AppConnector{}
)

func RegisterAppConnector(c AppConnector) {
	appConnectorsMu.Lock()
	defer appConnectorsMu.Unlock()
	appConnectors[c.Name] = c
}

func KnownAppConnectors() []AppConnector {
	appConnectorsMu.RLock()
	defer appConnectorsMu.RUnlock()
	connectors := make([]AppConnector, 0, len(appConnectors))
	for _, c := range appConnectors {
		connectors = append(connectors, c)
	}
	slices.SortFunc(connectors, func(a, b AppConnector) int { return strings.Compare(a.Name, b.Name) })
	return connectors
}

func FindAppConnector(name string) (AppConnector, bool) {
	appConnectorsMu.RLock()
	defer appConnectorsMu.RUnlock()
	c, ok := appConnectors[name]
	return c, ok
}

// Connectors lists the connectors the app's actions belong to.
func (a App) Connectors() []string {
	var names []string
	for _, action := range a.Actions {
		name, _, _ := strings.Cut(action, ":")
		if !slices.Contains(names, name) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// NeedsUserGrant reports whether any of the app's connectors needs the user's authorization.
func (a App) NeedsUserGrant() bool {
	return slices.ContainsFunc(a.Connectors(), func(name string) bool {
		c, ok := FindAppConnector(name)
		return ok && c.NeedsUserGrant
	})
}
