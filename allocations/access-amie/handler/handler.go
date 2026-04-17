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

package handler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

type PacketHandler interface {
	Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error
	SupportsType() string
}

func requireText(val, fieldName string) error {
	if strings.TrimSpace(val) == "" {
		return fmt.Errorf("'%s' must not be empty", fieldName)
	}
	return nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBody(packetJSON map[string]any) (map[string]any, error) {
	b, ok := packetJSON["body"]
	if !ok {
		return nil, fmt.Errorf("packet missing 'body'")
	}
	body, ok := b.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("packet 'body' is not an object")
	}
	return body, nil
}

func getResourceList(body map[string]any) []string {
	v, ok := body["ResourceList"]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
