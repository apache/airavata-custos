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
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// IdentityCacheMaxTTL caps how long the caller resolver may hold a verified
// identity in process memory.
const IdentityCacheMaxTTL = 60 * time.Second

// IdentityCacheDefaultTTL is the TTL applied when core.auth.cache_ttl is unset
// or non-positive.
const IdentityCacheDefaultTTL = 30 * time.Second

type Config struct {
	Core       CoreConfig                  `yaml:"core"`
	Connectors map[string]*ConnectorConfig `yaml:"connectors"`
}

type CoreConfig struct {
	Database DatabaseConfig `yaml:"database"`
	API      APIConfig      `yaml:"api"`
	LogLevel string         `yaml:"log_level"`
	Auth     AuthConfig     `yaml:"auth"`
}

type DatabaseConfig struct {
	URL string `yaml:"url"`
}

type APIConfig struct {
	Port int `yaml:"port"`
}

type AuthConfig struct {
	OIDC     OIDCConfig    `yaml:"oidc"`
	CacheTTL time.Duration `yaml:"cache_ttl"`
}

type OIDCConfig struct {
	Issuer   string `yaml:"issuer"`
	Audience string `yaml:"audience"`
}

type ConnectorConfig struct {
	Type    string                 `yaml:"type"`
	Enabled bool                   `yaml:"enabled"`
	Config  map[string]interface{} `yaml:",inline"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	expandedData := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expandedData), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := applyAuthDefaults(&cfg.Core.Auth); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// applyAuthDefaults enforces the auth invariants:
//   - Issuer and Audience are required.
//   - CacheTTL <= 0 collapses to IdentityCacheDefaultTTL.
//   - CacheTTL > IdentityCacheMaxTTL caps at the ceiling and logs a warning.
func applyAuthDefaults(a *AuthConfig) error {
	if a.OIDC.Issuer == "" {
		return fmt.Errorf("core.auth.oidc.issuer is required")
	}
	if a.OIDC.Audience == "" {
		return fmt.Errorf("core.auth.oidc.audience is required")
	}
	if a.CacheTTL <= 0 {
		a.CacheTTL = IdentityCacheDefaultTTL
	} else if a.CacheTTL > IdentityCacheMaxTTL {
		slog.Warn("core.auth.cache_ttl above cap; capping", "configured", a.CacheTTL, "cap", IdentityCacheMaxTTL)
		a.CacheTTL = IdentityCacheMaxTTL
	}
	return nil
}

func expandEnvVars(input string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		envVar := match[2 : len(match)-1]
		if value, exists := os.LookupEnv(envVar); exists {
			return value
		}
		return match
	})
}

func (c *ConnectorConfig) GetStringField(key string) (string, error) {
	if val, ok := c.Config[key]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("field %s is not a string", key)
	}
	return "", fmt.Errorf("field %s not found", key)
}

func (c *ConnectorConfig) GetIntField(key string) (int, error) {
	if val, ok := c.Config[key]; ok {
		if num, ok := val.(int); ok {
			return num, nil
		}
		return 0, fmt.Errorf("field %s is not an integer", key)
	}
	return 0, fmt.Errorf("field %s not found", key)
}

func (c *ConnectorConfig) GetDurationField(key string) (time.Duration, error) {
	if val, ok := c.Config[key]; ok {
		if str, ok := val.(string); ok {
			return time.ParseDuration(str)
		}
		return 0, fmt.Errorf("field %s is not a string", key)
	}
	return 0, fmt.Errorf("field %s not found", key)
}

func (c *ConnectorConfig) GetNestedConfig(key string) (map[string]interface{}, error) {
	if val, ok := c.Config[key]; ok {
		if nested, ok := val.(map[string]interface{}); ok {
			return nested, nil
		}
		return nil, fmt.Errorf("field %s is not a map", key)
	}
	return nil, fmt.Errorf("field %s not found", key)
}
