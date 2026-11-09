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

// Package config handles YAML configuration loading with environment variable overrides.
package config

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const minServiceTokenLength = 32

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Vault    VaultConfig    `yaml:"vault"`
	Signer   SignerConfig   `yaml:"signer"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

type ServerConfig struct {
	Port                   int `yaml:"port"`
	ReadTimeoutSeconds     int `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds    int `yaml:"write_timeout_seconds"`
	ShutdownTimeoutSeconds int `yaml:"shutdown_timeout_seconds"`
}

type DatabaseConfig struct {
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	Name                   string `yaml:"name"`
	Username               string `yaml:"username"`
	Password               string `yaml:"password"`
	MaxOpenConns           int    `yaml:"max_open_conns"`
	MaxIdleConns           int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeSeconds int    `yaml:"conn_max_lifetime_seconds"`
}

type VaultConfig struct {
	Address        string `yaml:"address"`
	Token          string `yaml:"token"`
	MountPath      string `yaml:"mount_path"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type SignerConfig struct {
	ServiceToken      string           `yaml:"service_token"`
	ReservedUsernames []string         `yaml:"reserved_usernames"`
	LoginHost         string           `yaml:"login_host"`
	LoginPort         int              `yaml:"login_port"`
	CA                CAConfig         `yaml:"ca"`
	Policy            PolicyConfig     `yaml:"policy"`
	Validation        ValidationConfig `yaml:"validation"`
}

type CAConfig struct {
	Rotation RotationConfig `yaml:"rotation"`
}

type RotationConfig struct {
	PeriodHours  int `yaml:"period_hours"`
	OverlapHours int `yaml:"overlap_hours"`
}

type PolicyConfig struct {
	MaxTTLSeconds    int      `yaml:"max_ttl_seconds"`
	AllowedKeyTypes  []string `yaml:"allowed_key_types"`
	DeniedExtensions []string `yaml:"denied_extensions"`
}

type ValidationConfig struct {
	PrincipalValidator string `yaml:"principal_validator"`
	CacheTTLSeconds    int    `yaml:"cache_ttl_seconds"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:                   8084,
			ReadTimeoutSeconds:     30,
			WriteTimeoutSeconds:    30,
			ShutdownTimeoutSeconds: 30,
		},
		Database: DatabaseConfig{
			Host:                   "localhost",
			Port:                   5432,
			Name:                   "custos_signer",
			Username:               "admin",
			Password:               "admin",
			MaxOpenConns:           25,
			MaxIdleConns:           5,
			ConnMaxLifetimeSeconds: 300,
		},
		Vault: VaultConfig{
			Address:        "http://localhost:8200",
			Token:          "",
			MountPath:      "ssh-ca",
			TimeoutSeconds: 10,
		},
		Signer: SignerConfig{
			ReservedUsernames: []string{"root"},
			LoginPort:         22,
			CA: CAConfig{
				Rotation: RotationConfig{
					PeriodHours:  2160,
					OverlapHours: 2,
				},
			},
			Policy: PolicyConfig{
				MaxTTLSeconds:   86400,
				AllowedKeyTypes: []string{"ed25519", "rsa", "ecdsa"},
			},
			Validation: ValidationConfig{
				PrincipalValidator: "none",
				CacheTTLSeconds:    300, // 5 minutes cache
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
}

// Load reads configuration from a YAML file, then applies environment variable overrides.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

// Validate stops startup on unsafe settings. Whoever holds the service token can get a certificate for any user that is not reserved.
func (c *Config) Validate() error {
	if len(c.Signer.ServiceToken) < minServiceTokenLength {
		return fmt.Errorf("service_token must be at least %d characters", minServiceTokenLength)
	}
	if v := c.Signer.Validation.PrincipalValidator; v != "none" && v != "ldap" {
		return fmt.Errorf("principal_validator %q must be none or ldap", v)
	}
	if !slices.ContainsFunc(c.Signer.ReservedUsernames, func(name string) bool { return strings.EqualFold(name, "root") }) {
		return fmt.Errorf("reserved_usernames must include root")
	}
	if c.Signer.LoginHost == "" {
		return fmt.Errorf("login_host is required")
	}
	return nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = p
		}
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("VAULT_ADDRESS"); v != "" {
		cfg.Vault.Address = v
	}
	if v := os.Getenv("VAULT_TOKEN"); v != "" {
		cfg.Vault.Token = v
	}
	if v := os.Getenv("SIGNER_SERVICE_TOKEN"); v != "" {
		cfg.Signer.ServiceToken = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
}
