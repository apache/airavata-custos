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

package operations

import (
	"encoding/json"
	"fmt"
)

// UnixClusterAccountBlock is appended to the composite PUT body. Field names
// mirror COmanage's JSON attributes.
type UnixClusterAccountBlock struct {
	UnixClusterId    int    `json:"unix_cluster_id"`
	SyncMode         string `json:"sync_mode"`
	Status           string `json:"status"`
	Username         string `json:"username"`
	Uid              int64  `json:"uid"`
	Gecos            string `json:"gecos"`
	LoginShell       string `json:"login_shell"`
	HomeDirectory    string `json:"home_directory"`
	PrimaryCoGroupId int    `json:"primary_co_group_id"`
}

// mergeUnixClusterAccount sets "UnixClusterAccount" on the composite while
// preserving every other top-level key.
func mergeUnixClusterAccount(composite json.RawMessage, block UnixClusterAccountBlock) ([]byte, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(composite, &top); err != nil {
		return nil, fmt.Errorf("decode composite: %w", err)
	}
	blockJSON, err := json.Marshal([]UnixClusterAccountBlock{block})
	if err != nil {
		return nil, fmt.Errorf("encode UnixClusterAccount block: %w", err)
	}
	top["UnixClusterAccount"] = blockJSON
	out, err := json.Marshal(top)
	if err != nil {
		return nil, fmt.Errorf("encode merged composite: %w", err)
	}
	return out, nil
}

// mergeLoginIdentifier sets the composite's uid identifier to username. The
// composite PUT is the only person-identifier write the registry authorizes,
// and the directory maps the login name from this entry.
func mergeLoginIdentifier(composite json.RawMessage, username string) ([]byte, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(composite, &top); err != nil {
		return nil, fmt.Errorf("decode composite: %w", err)
	}
	var idents []map[string]json.RawMessage
	if raw, ok := top["Identifier"]; ok {
		if err := json.Unmarshal(raw, &idents); err != nil {
			return nil, fmt.Errorf("decode Identifier array: %w", err)
		}
	}
	value, _ := json.Marshal(username)
	replaced := false
	for _, ident := range idents {
		var typ string
		_ = json.Unmarshal(ident["type"], &typ)
		if typ == "uid" {
			ident["identifier"] = value
			replaced = true
		}
	}
	if !replaced {
		idents = append(idents, map[string]json.RawMessage{
			"identifier": value,
			"type":       json.RawMessage(`"uid"`),
			"login":      json.RawMessage("false"),
			"status":     json.RawMessage(`"A"`),
		})
	}
	identsJSON, err := json.Marshal(idents)
	if err != nil {
		return nil, fmt.Errorf("encode Identifier array: %w", err)
	}
	top["Identifier"] = identsJSON
	out, err := json.Marshal(top)
	if err != nil {
		return nil, fmt.Errorf("encode merged composite: %w", err)
	}
	return out, nil
}

// extractIdentifier returns the first Identifier.identifier whose type matches,
// or "" if none.
func extractIdentifier(composite json.RawMessage, identifierType string) (string, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(composite, &top); err != nil {
		return "", fmt.Errorf("decode composite: %w", err)
	}
	rawIdents, ok := top["Identifier"]
	if !ok {
		return "", nil
	}
	var idents []struct {
		Identifier string `json:"identifier"`
		Type       string `json:"type"`
	}
	if err := json.Unmarshal(rawIdents, &idents); err != nil {
		return "", fmt.Errorf("decode Identifier array: %w", err)
	}
	for _, id := range idents {
		if id.Type == identifierType {
			return id.Identifier, nil
		}
	}
	return "", nil
}

// extractOrgIdentifierValues returns the identifier values of the given type
// across the composite's org identities.
func extractOrgIdentifierValues(composite json.RawMessage, identifierType string) ([]string, error) {
	var top struct {
		OrgIdentity []struct {
			Identifier []struct {
				Identifier string `json:"identifier"`
				Type       string `json:"type"`
			} `json:"Identifier"`
		} `json:"OrgIdentity"`
	}
	if err := json.Unmarshal(composite, &top); err != nil {
		return nil, fmt.Errorf("decode composite for org identifiers: %w", err)
	}
	var out []string
	for _, org := range top.OrgIdentity {
		for _, id := range org.Identifier {
			if id.Type == identifierType {
				out = append(out, id.Identifier)
			}
		}
	}
	return out, nil
}

// extractCoPersonID returns the numeric CoPerson.meta.id (distinct from the
// PersonIDType identifier string), or 0 if missing.
func extractCoPersonID(composite json.RawMessage) (int, error) {
	var top struct {
		CoPerson struct {
			Meta struct {
				Id int `json:"id"`
			} `json:"meta"`
		} `json:"CoPerson"`
	}
	if err := json.Unmarshal(composite, &top); err != nil {
		return 0, fmt.Errorf("decode composite for CoPerson.id: %w", err)
	}
	return top.CoPerson.Meta.Id, nil
}
