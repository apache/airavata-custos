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

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultConfig(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 8084 {
		t.Errorf("expected port 8084, got %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", cfg.Database.Host)
	}
	if cfg.Signer.Validation.PrincipalValidator != "noop" {
		t.Errorf("expected noop validator, got %s", cfg.Signer.Validation.PrincipalValidator)
	}
}

func TestLoad_YAMLFile(t *testing.T) {
	yaml := `
server:
  port: 9090
database:
  host: dbhost
  port: 3307
  password: file_password
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Database.Host != "dbhost" {
		t.Errorf("expected host dbhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Password != "file_password" {
		t.Errorf("expected password file_password, got %s", cfg.Database.Password)
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	yaml := `
database:
  password: file_password
  host: filehost
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DB_PASSWORD", "env_password")
	t.Setenv("DB_HOST", "envhost")
	t.Setenv("VAULT_ADDRESS", "http://vault:8200")
	t.Setenv("DEV_MODE", "true")
	t.Setenv("DEV_DEFAULT_EMAIL", "test@dev.local")
	t.Setenv("ALLOWED_ISSUERS", "https://a.com, https://b.com")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Password != "env_password" {
		t.Errorf("expected env_password, got %s", cfg.Database.Password)
	}
	if cfg.Database.Host != "envhost" {
		t.Errorf("expected envhost, got %s", cfg.Database.Host)
	}
	if cfg.Vault.Address != "http://vault:8200" {
		t.Errorf("expected vault address override, got %s", cfg.Vault.Address)
	}
	if !cfg.DevMode.Enabled {
		t.Error("expected dev mode enabled via env")
	}
	if cfg.DevMode.DefaultEmail != "test@dev.local" {
		t.Errorf("expected default email 'test@dev.local', got %s", cfg.DevMode.DefaultEmail)
	}
	if len(cfg.Signer.Auth.AllowedIssuers) != 2 {
		t.Errorf("expected 2 issuers, got %d", len(cfg.Signer.Auth.AllowedIssuers))
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected debug log level, got %s", cfg.Logging.Level)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing config file")
	}
}
