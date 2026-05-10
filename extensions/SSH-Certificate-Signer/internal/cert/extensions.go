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

package cert

import "fmt"

// SSHExtension represents a valid OpenSSH certificate extension name.
type SSHExtension string

const (
	ExtPermitPTY             SSHExtension = "permit-pty"
	ExtPermitPortForwarding  SSHExtension = "permit-port-forwarding"
	ExtPermitUserRC          SSHExtension = "permit-user-rc"
	ExtPermitAgentForwarding SSHExtension = "permit-agent-forwarding"
	ExtPermitX11Forwarding   SSHExtension = "permit-X11-forwarding"
	ExtNoTouchRequired       SSHExtension = "no-touch-required"
)

var validExtensions = map[SSHExtension]bool{
	ExtPermitPTY:             true,
	ExtPermitPortForwarding:  true,
	ExtPermitUserRC:          true,
	ExtPermitAgentForwarding: true,
	ExtPermitX11Forwarding:   true,
	ExtNoTouchRequired:       true,
}

// AllStandardExtensions returns the 5 standard non-FIDO extensions.
// no-touch-required is excluded as it is FIDO-specific.
func AllStandardExtensions() []SSHExtension {
	return []SSHExtension{
		ExtPermitPTY,
		ExtPermitPortForwarding,
		ExtPermitUserRC,
		ExtPermitAgentForwarding,
		ExtPermitX11Forwarding,
	}
}

func (e SSHExtension) Validate() error {
	if !validExtensions[e] {
		return fmt.Errorf("unknown SSH extension: %q", string(e))
	}
	return nil
}

// ValidateExtensionList validates a list of extension name strings and returns
// them as typed SSHExtension values. Returns an error naming the first invalid entry.
func ValidateExtensionList(names []string) ([]SSHExtension, error) {
	result := make([]SSHExtension, 0, len(names))
	for _, name := range names {
		ext := SSHExtension(name)
		if err := ext.Validate(); err != nil {
			return nil, err
		}
		result = append(result, ext)
	}
	return result, nil
}

// ResolveExtensions computes the granted extension set by starting with all
// standard extensions and removing denied and excluded entries.
func ResolveExtensions(denied []string, excluded []string) ([]SSHExtension, error) {
	// Validate inputs
	if _, err := ValidateExtensionList(denied); err != nil {
		return nil, fmt.Errorf("invalid denied extension: %w", err)
	}
	if _, err := ValidateExtensionList(excluded); err != nil {
		return nil, fmt.Errorf("invalid excluded extension: %w", err)
	}

	// Build removal set
	remove := make(map[SSHExtension]bool, len(denied)+len(excluded))
	for _, d := range denied {
		remove[SSHExtension(d)] = true
	}
	for _, e := range excluded {
		remove[SSHExtension(e)] = true
	}

	// Filter
	var granted []SSHExtension
	for _, ext := range AllStandardExtensions() {
		if !remove[ext] {
			granted = append(granted, ext)
		}
	}
	return granted, nil
}

// ExtensionsToMap converts a list of SSHExtension values to the map[string]string
// format required by golang.org/x/crypto/ssh.Certificate.Permissions.Extensions.
func ExtensionsToMap(exts []SSHExtension) map[string]string {
	m := make(map[string]string, len(exts))
	for _, e := range exts {
		m[string(e)] = ""
	}
	return m
}

// ExtensionNames converts a list of SSHExtension values to plain strings.
func ExtensionNames(exts []SSHExtension) []string {
	names := make([]string, len(exts))
	for i, e := range exts {
		names[i] = string(e)
	}
	return names
}
