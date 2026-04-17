// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type ClientConfig struct {
	TenantID                 string
	ClientID                 string
	ClientSecret             string // BCrypt hash
	TargetHost               string
	TargetPort               int
	MaxTTLSeconds            int
	AllowedKeyTypes          []string
	SourceAddressRestriction *string
	DeniedExtensions         []string
	PrincipalSource          string
	Enabled                  bool
}

func (d *DB) GetClientConfig(ctx context.Context, tenantID, clientID string) (*ClientConfig, error) {
	var (
		cc                       ClientConfig
		allowedKeyTypesJSON      []byte
		sourceAddressRestriction sql.NullString
		deniedExtensionsJSON     sql.NullString
	)

	err := d.QueryRowContext(ctx,
		`SELECT tenant_id, client_id, client_secret, target_host, target_port,
		        max_ttl_seconds, allowed_key_types, source_address_restriction,
		        denied_extensions, principal_source, enabled
		 FROM client_ssh_configs
		 WHERE tenant_id = ? AND client_id = ?`,
		tenantID, clientID,
	).Scan(
		&cc.TenantID, &cc.ClientID, &cc.ClientSecret, &cc.TargetHost, &cc.TargetPort,
		&cc.MaxTTLSeconds, &allowedKeyTypesJSON, &sourceAddressRestriction,
		&deniedExtensionsJSON, &cc.PrincipalSource, &cc.Enabled,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying client config: %w", err)
	}

	if err := json.Unmarshal(allowedKeyTypesJSON, &cc.AllowedKeyTypes); err != nil {
		return nil, fmt.Errorf("unmarshaling allowed_key_types: %w", err)
	}

	if sourceAddressRestriction.Valid {
		cc.SourceAddressRestriction = &sourceAddressRestriction.String
	}

	if deniedExtensionsJSON.Valid && deniedExtensionsJSON.String != "" {
		if err := json.Unmarshal([]byte(deniedExtensionsJSON.String), &cc.DeniedExtensions); err != nil {
			return nil, fmt.Errorf("unmarshaling denied_extensions: %w", err)
		}
	}

	return &cc, nil
}
