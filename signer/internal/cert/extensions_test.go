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

import (
	"testing"
)

func TestSSHExtension_Validate_Valid(t *testing.T) {
	valid := []SSHExtension{
		ExtPermitPTY, ExtPermitPortForwarding, ExtPermitUserRC,
		ExtPermitAgentForwarding, ExtPermitX11Forwarding, ExtNoTouchRequired,
	}
	for _, ext := range valid {
		if err := ext.Validate(); err != nil {
			t.Errorf("expected %q to be valid, got: %v", ext, err)
		}
	}
}

func TestSSHExtension_Validate_Invalid(t *testing.T) {
	invalid := []SSHExtension{"unknown-ext", "permit-PTY", "PERMIT-PTY", ""}
	for _, ext := range invalid {
		if err := ext.Validate(); err == nil {
			t.Errorf("expected %q to be invalid", ext)
		}
	}
}

func TestValidateExtensionList_Valid(t *testing.T) {
	names := []string{"permit-pty", "permit-port-forwarding"}
	result, err := ValidateExtensionList(names)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(result))
	}
}

func TestValidateExtensionList_InvalidEntry(t *testing.T) {
	names := []string{"permit-pty", "bogus-extension"}
	_, err := ValidateExtensionList(names)
	if err == nil {
		t.Fatal("expected error for invalid extension")
	}
}

func TestValidateExtensionList_Empty(t *testing.T) {
	result, err := ValidateExtensionList(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 extensions, got %d", len(result))
	}
}

func TestAllStandardExtensions_Count(t *testing.T) {
	exts := AllStandardExtensions()
	if len(exts) != 5 {
		t.Errorf("expected 5 standard extensions, got %d", len(exts))
	}
}

func TestAllStandardExtensions_ExcludesNoTouchRequired(t *testing.T) {
	for _, ext := range AllStandardExtensions() {
		if ext == ExtNoTouchRequired {
			t.Error("AllStandardExtensions should not include no-touch-required")
		}
	}
}

func TestResolveExtensions_NoDenials(t *testing.T) {
	granted, err := ResolveExtensions(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(granted) != 5 {
		t.Errorf("expected 5 extensions, got %d", len(granted))
	}
}

func TestResolveExtensions_DenyOne(t *testing.T) {
	granted, err := ResolveExtensions([]string{"permit-port-forwarding"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(granted) != 4 {
		t.Errorf("expected 4 extensions, got %d", len(granted))
	}
	for _, ext := range granted {
		if ext == ExtPermitPortForwarding {
			t.Error("permit-port-forwarding should be denied")
		}
	}
}

func TestResolveExtensions_ExcludeOne(t *testing.T) {
	granted, err := ResolveExtensions(nil, []string{"permit-user-rc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(granted) != 4 {
		t.Errorf("expected 4 extensions, got %d", len(granted))
	}
	for _, ext := range granted {
		if ext == ExtPermitUserRC {
			t.Error("permit-user-rc should be excluded")
		}
	}
}

func TestResolveExtensions_DenyAndExcludeOverlap(t *testing.T) {
	granted, err := ResolveExtensions(
		[]string{"permit-port-forwarding"},
		[]string{"permit-port-forwarding", "permit-user-rc"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(granted) != 3 {
		t.Errorf("expected 3 extensions, got %d", len(granted))
	}
}

func TestResolveExtensions_DenyAll(t *testing.T) {
	all := ExtensionNames(AllStandardExtensions())
	granted, err := ResolveExtensions(all, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(granted) != 0 {
		t.Errorf("expected 0 extensions, got %d", len(granted))
	}
}

func TestResolveExtensions_InvalidDenied(t *testing.T) {
	_, err := ResolveExtensions([]string{"bogus"}, nil)
	if err == nil {
		t.Fatal("expected error for invalid denied extension")
	}
}

func TestResolveExtensions_InvalidExcluded(t *testing.T) {
	_, err := ResolveExtensions(nil, []string{"bogus"})
	if err == nil {
		t.Fatal("expected error for invalid excluded extension")
	}
}

func TestExtensionsToMap(t *testing.T) {
	exts := []SSHExtension{ExtPermitPTY, ExtPermitPortForwarding}
	m := ExtensionsToMap(exts)
	if len(m) != 2 {
		t.Errorf("expected 2 entries, got %d", len(m))
	}
	if _, ok := m["permit-pty"]; !ok {
		t.Error("expected permit-pty in map")
	}
	if v := m["permit-pty"]; v != "" {
		t.Errorf("expected empty value for permit-pty, got %q", v)
	}
}

func TestExtensionNames(t *testing.T) {
	exts := []SSHExtension{ExtPermitPTY, ExtPermitUserRC}
	names := ExtensionNames(exts)
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
	if names[0] != "permit-pty" || names[1] != "permit-user-rc" {
		t.Errorf("unexpected names: %v", names)
	}
}
