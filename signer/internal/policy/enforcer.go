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
	"net"
	"strings"

	"github.com/apache/airavata-custos/signer/internal/cert"
	"github.com/apache/airavata-custos/signer/internal/store"
)

type Enforcer struct {
	defaultMaxTTL       int
	defaultAllowedTypes []string
}

func NewEnforcer(defaultMaxTTL int, defaultAllowedTypes []string) *Enforcer {
	return &Enforcer{
		defaultMaxTTL:       defaultMaxTTL,
		defaultAllowedTypes: defaultAllowedTypes,
	}
}

type PolicyError struct {
	Message string
}

func (e *PolicyError) Error() string {
	return e.Message
}

func (e *Enforcer) Enforce(ttlSeconds int, sshKeyType string, sourceIP string, clientCfg *store.ClientConfig) error {
	maxTTL := clientCfg.MaxTTLSeconds
	if maxTTL == 0 {
		maxTTL = e.defaultMaxTTL
	}
	if ttlSeconds <= 0 {
		return &PolicyError{Message: fmt.Sprintf("TTL must be greater than 0, got %d", ttlSeconds)}
	}
	if ttlSeconds > maxTTL {
		return &PolicyError{Message: fmt.Sprintf("Requested TTL %d seconds exceeds maximum allowed TTL %d seconds", ttlSeconds, maxTTL)}
	}

	normalizedType := cert.NormalizeKeyType(sshKeyType)
	allowedTypes := clientCfg.AllowedKeyTypes
	if len(allowedTypes) == 0 {
		allowedTypes = e.defaultAllowedTypes
	}
	typeAllowed := false
	for _, at := range allowedTypes {
		if strings.EqualFold(at, normalizedType) {
			typeAllowed = true
			break
		}
	}
	if !typeAllowed {
		return &PolicyError{Message: fmt.Sprintf("Key type %q is not allowed. Allowed types: %v", normalizedType, allowedTypes)}
	}

	if clientCfg.SourceAddressRestriction != nil && *clientCfg.SourceAddressRestriction != "" {
		_, cidr, err := net.ParseCIDR(*clientCfg.SourceAddressRestriction)
		if err != nil {
			return &PolicyError{Message: fmt.Sprintf("Invalid source address restriction CIDR: %s", *clientCfg.SourceAddressRestriction)}
		}
		ip := net.ParseIP(sourceIP)
		if ip == nil {
			return &PolicyError{Message: fmt.Sprintf("Could not parse source IP: %s", sourceIP)}
		}
		if !cidr.Contains(ip) {
			return &PolicyError{Message: fmt.Sprintf("Source IP %s is not within allowed CIDR %s", sourceIP, *clientCfg.SourceAddressRestriction)}
		}
	}

	return nil
}

// BuildCriticalOptions constructs the SSH certificate critical options map from
// the client config's source address restriction and the per-request force command.
func BuildCriticalOptions(sourceAddr *string, forceCommand string) map[string]string {
	opts := make(map[string]string)
	if sourceAddr != nil && *sourceAddr != "" {
		opts["source-address"] = *sourceAddr
	}
	if forceCommand != "" {
		opts["force-command"] = forceCommand
	}
	if len(opts) == 0 {
		return nil
	}
	return opts
}
