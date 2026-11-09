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

package policy

import (
	"fmt"
	"strings"

	"github.com/apache/airavata-custos/signer/internal/cert"
)

type Enforcer struct {
	maxTTL       int
	allowedTypes []string
}

func NewEnforcer(maxTTL int, allowedTypes []string) *Enforcer {
	return &Enforcer{maxTTL: maxTTL, allowedTypes: allowedTypes}
}

type PolicyError struct {
	Code    string
	Message string
}

func (e *PolicyError) Error() string {
	return e.Message
}

func (e *Enforcer) Enforce(ttlSeconds int, sshKeyType string) error {
	if ttlSeconds <= 0 {
		return &PolicyError{Code: "invalid_ttl", Message: fmt.Sprintf("TTL must be greater than 0, got %d", ttlSeconds)}
	}
	if ttlSeconds > e.maxTTL {
		return &PolicyError{Code: "invalid_ttl", Message: fmt.Sprintf("Requested TTL %d seconds exceeds maximum allowed TTL %d seconds", ttlSeconds, e.maxTTL)}
	}

	normalizedType := cert.NormalizeKeyType(sshKeyType)
	for _, allowed := range e.allowedTypes {
		if strings.EqualFold(allowed, normalizedType) {
			return nil
		}
	}
	return &PolicyError{Code: "key_type_not_allowed", Message: fmt.Sprintf("Key type %q is not allowed. Allowed types: %v", normalizedType, e.allowedTypes)}
}
