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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
