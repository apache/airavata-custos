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

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTempYAML writes contents to a temp file and returns its path.
func writeTempYAML(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "custos.yaml")
	if err := os.WriteFile(p, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp yaml: %v", err)
	}
	return p
}

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig("../../config/custos.yaml")
	require.NoError(t, err)

	assert.Equal(t, "info", cfg.Core.LogLevel)
	assert.Equal(t, 8080, cfg.Core.API.Port)
	assert.NotEmpty(t, cfg.Core.Database.URL)

	assert.NotNil(t, cfg.Connectors)
	assert.Greater(t, len(cfg.Connectors), 0)

	slurmMapperCfg, ok := cfg.Connectors["slurm-mapper"]
	assert.True(t, ok)
	assert.True(t, slurmMapperCfg.Enabled)
	assert.Equal(t, "slurm-association-mapper", slurmMapperCfg.Type)
}

func TestExpandEnvVars(t *testing.T) {
	t.Run("expands environment variables", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		input := "url: ${TEST_VAR}"
		output := expandEnvVars(input)
		assert.Equal(t, "url: test_value", output)
	})

	t.Run("leaves unexpanded vars when env not set", func(t *testing.T) {
		os.Unsetenv("UNDEFINED_VAR")
		input := "url: ${UNDEFINED_VAR}"
		output := expandEnvVars(input)
		assert.Equal(t, "url: ${UNDEFINED_VAR}", output)
	})

	t.Run("expands multiple variables", func(t *testing.T) {
		os.Setenv("VAR1", "value1")
		os.Setenv("VAR2", "value2")
		defer os.Unsetenv("VAR1")
		defer os.Unsetenv("VAR2")

		input := "url: ${VAR1}\nkey: ${VAR2}"
		output := expandEnvVars(input)
		assert.Equal(t, "url: value1\nkey: value2", output)
	})
}

func TestConnectorConfigGetters(t *testing.T) {
	cfg, err := LoadConfig("../../config/custos.yaml")
	require.NoError(t, err)

	t.Run("verify connector type field", func(t *testing.T) {
		slurmMapper, ok := cfg.Connectors["slurm-mapper"]
		require.True(t, ok)
		assert.Equal(t, "slurm-association-mapper", slurmMapper.Type)
	})

	t.Run("GetNestedConfig", func(t *testing.T) {
		slurmMapper, ok := cfg.Connectors["slurm-mapper"]
		require.True(t, ok)
		nested, err := slurmMapper.GetNestedConfig("slurm_api")
		require.NoError(t, err)
		assert.NotNil(t, nested)
		assert.Contains(t, nested, "url")
		assert.Contains(t, nested, "username")
	})

	t.Run("GetNestedConfig from provisioner", func(t *testing.T) {
		provisioner, ok := cfg.Connectors["comanage-provisioner"]
		if ok && provisioner != nil {
			nested, err := provisioner.GetNestedConfig("registry")
			require.NoError(t, err)
			assert.NotNil(t, nested)
		}
	})
}

func TestLoadConfig_AuthDefaults(t *testing.T) {
	cfg, err := LoadConfig("../../config/custos.yaml")
	require.NoError(t, err)

	assert.Equal(t, 30*time.Second, cfg.Core.Auth.CacheTTL)
	assert.Equal(t, []string{"/healthz", "/ready"}, cfg.Core.Auth.PublicPaths)
	assert.NotEmpty(t, cfg.Core.Auth.OIDC.Issuer, "issuer should not be empty (env-unset stays as ${...} literal)")
	assert.NotEmpty(t, cfg.Core.Auth.OIDC.Audience)
}

func TestLoadConfig_AuthRejectsEmptyIssuer(t *testing.T) {
	yaml := `core:
  database:
    url: "dsn"
  api:
    port: 8080
  log_level: "info"
  auth:
    oidc:
      issuer: ""
      audience: "aud"
`
	_, err := LoadConfig(writeTempYAML(t, yaml))
	if err == nil {
		t.Fatal("expected error for empty issuer, got nil")
	}
	assert.Contains(t, err.Error(), "core.auth.oidc.issuer")
}

func TestLoadConfig_AuthRejectsEmptyAudience(t *testing.T) {
	yaml := `core:
  database:
    url: "dsn"
  api:
    port: 8080
  log_level: "info"
  auth:
    oidc:
      issuer: "https://idp.example"
      audience: ""
`
	_, err := LoadConfig(writeTempYAML(t, yaml))
	if err == nil {
		t.Fatal("expected error for empty audience, got nil")
	}
	assert.Contains(t, err.Error(), "core.auth.oidc.audience")
}

func TestLoadConfig_AuthClampsTTLAboveMax(t *testing.T) {
	yaml := `core:
  database:
    url: "dsn"
  api:
    port: 8080
  log_level: "info"
  auth:
    oidc:
      issuer: "https://idp.example"
      audience: "aud"
    cache_ttl: "5m"
`
	cfg, err := LoadConfig(writeTempYAML(t, yaml))
	require.NoError(t, err)
	assert.Equal(t, IdentityCacheMaxTTL, cfg.Core.Auth.CacheTTL)
}

func TestLoadConfig_AuthDefaultsTTLWhenZero(t *testing.T) {
	yaml := `core:
  database:
    url: "dsn"
  api:
    port: 8080
  log_level: "info"
  auth:
    oidc:
      issuer: "https://idp.example"
      audience: "aud"
`
	cfg, err := LoadConfig(writeTempYAML(t, yaml))
	require.NoError(t, err)
	assert.Equal(t, IdentityCacheDefaultTTL, cfg.Core.Auth.CacheTTL)
}
