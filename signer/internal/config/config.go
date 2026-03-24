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
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Vault    VaultConfig    `yaml:"vault"`
	Signer   SignerConfig   `yaml:"signer"`
	DevMode  DevModeConfig  `yaml:"dev_mode"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
	CORS     CORSConfig     `yaml:"cors"`
}

// DevModeConfig disables OIDC token validation and returns a default identity
// for all requests when enabled.
type DevModeConfig struct {
	Enabled      bool   `yaml:"enabled"`
	DefaultEmail string `yaml:"default_email"`
}

type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
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
	CA         CAConfig         `yaml:"ca"`
	Policy     PolicyConfig     `yaml:"policy"`
	Auth       AuthConfig       `yaml:"auth"`
	Validation ValidationConfig `yaml:"validation"`
}

type CAConfig struct {
	Rotation RotationConfig `yaml:"rotation"`
}

type RotationConfig struct {
	PeriodHours  int `yaml:"period_hours"`
	OverlapHours int `yaml:"overlap_hours"`
}

type PolicyConfig struct {
	Defaults PolicyDefaults `yaml:"defaults"`
}

type PolicyDefaults struct {
	MaxTTLSeconds   int      `yaml:"max_ttl_seconds"`
	AllowedKeyTypes []string `yaml:"allowed_key_types"`
}

type AuthConfig struct {
	AllowedIssuers []string   `yaml:"allowed_issuers"`
	OIDC           OIDCConfig `yaml:"oidc"`
}

type OIDCConfig struct {
	JWKSCacheTTLSeconds int `yaml:"jwks_cache_ttl_seconds"`
	JWKSMaxProviders    int `yaml:"jwks_max_providers"`
	TimeoutSeconds      int `yaml:"timeout_seconds"`
}

type ValidationConfig struct {
	PrincipalValidator string         `yaml:"principal_validator"`
	COmanage           COmanageConfig `yaml:"comanage"`
}

type COmanageConfig struct {
	RegistryURL    string `yaml:"registry_url"`
	APIPath        string `yaml:"api_path"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	VerifySSL      bool   `yaml:"verify_ssl"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	AuditLevel string `yaml:"audit_level"`
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
			Port:                   3306,
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
			CA: CAConfig{
				Rotation: RotationConfig{
					PeriodHours:  2160,
					OverlapHours: 2,
				},
			},
			Policy: PolicyConfig{
				Defaults: PolicyDefaults{
					MaxTTLSeconds:   86400,
					AllowedKeyTypes: []string{"ed25519", "rsa", "ecdsa"},
				},
			},
			Auth: AuthConfig{
				AllowedIssuers: []string{},
				OIDC: OIDCConfig{
					JWKSCacheTTLSeconds: 300,
					JWKSMaxProviders:    10,
					TimeoutSeconds:      10,
				},
			},
			Validation: ValidationConfig{
				PrincipalValidator: "noop",
				COmanage: COmanageConfig{
					APIPath:        "/registry/co_people.json",
					TimeoutSeconds: 10,
					VerifySSL:      true,
				},
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			AuditLevel: "info",
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
	if v := os.Getenv("DEV_MODE"); v != "" {
		cfg.DevMode.Enabled = strings.EqualFold(v, "true")
	}
	if v := os.Getenv("DEV_DEFAULT_EMAIL"); v != "" {
		cfg.DevMode.DefaultEmail = v
	}
	if v := os.Getenv("ALLOWED_ISSUERS"); v != "" {
		issuers := strings.Split(v, ",")
		trimmed := make([]string, 0, len(issuers))
		for _, issuer := range issuers {
			s := strings.TrimSpace(issuer)
			if s != "" {
				trimmed = append(trimmed, s)
			}
		}
		cfg.Signer.Auth.AllowedIssuers = trimmed
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
}
