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
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Core       CoreConfig `yaml:"core"`
	Connectors map[string]*ConnectorConfig `yaml:"connectors"`
}

type CoreConfig struct {
	Database DatabaseConfig `yaml:"database"`
	API      APIConfig      `yaml:"api"`
	Auth     AuthConfig     `yaml:"auth"`
	CORS     CORSConfig     `yaml:"cors"`
	LogLevel string         `yaml:"log_level"`
}

// CORSConfig drives the origin allowlist for browser callers. An empty
// AllowedOrigins makes the middleware a no-op pass-through, which is the
// right behaviour for server-to-server deployments.
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type DatabaseConfig struct {
	URL string `yaml:"url"`
}

type APIConfig struct {
	Port int `yaml:"port"`
}

// AuthConfig drives OIDC bearer token verification at the HTTP boundary.
// Issuer + Audience are required at runtime; JWKSURL is an override the
// HTTP-layer integration tests use to point at an in-process JWKS server.
type AuthConfig struct {
	Issuer   string `yaml:"issuer"`
	Audience string `yaml:"audience"`
	JWKSURL  string `yaml:"jwks_url"`
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

	return &cfg, nil
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
