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
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	AMIE        AMIEConfig        `yaml:"amie"`
	Log         LogConfig         `yaml:"log"`
	Provisioner ProvisionerConfig `yaml:"provisioner"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	DSN          string `yaml:"dsn"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type AMIEConfig struct {
	BaseURL        string        `yaml:"base_url"`
	SiteCode       string        `yaml:"site_code"`
	APIKey         string        `yaml:"api_key"`
	PollInterval   time.Duration `yaml:"poll_interval"`
	WorkerInterval time.Duration `yaml:"worker_interval"`
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	PollerEnabled  bool          `yaml:"poller_enabled"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // "text" or "json"
}

type ProvisionerConfig struct {
	Type string `yaml:"type"` // "noop" or "slurm"
}

// Load reads config from a YAML file and applies environment variable overrides.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8083
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if cfg.AMIE.PollInterval == 0 {
		cfg.AMIE.PollInterval = 30 * time.Second
	}
	if cfg.AMIE.WorkerInterval == 0 {
		cfg.AMIE.WorkerInterval = 5 * time.Second
	}
	if cfg.AMIE.ConnectTimeout == 0 {
		cfg.AMIE.ConnectTimeout = 5 * time.Second
	}
	if cfg.AMIE.ReadTimeout == 0 {
		cfg.AMIE.ReadTimeout = 20 * time.Second
	}
	if !cfg.AMIE.PollerEnabled {
		cfg.AMIE.PollerEnabled = true
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "text"
	}
	if cfg.Provisioner.Type == "" {
		cfg.Provisioner.Type = "noop"
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AMIE_SITE_CODE"); v != "" {
		cfg.AMIE.SiteCode = v
	}
	if v := os.Getenv("AMIE_API_KEY"); v != "" {
		cfg.AMIE.APIKey = v
	}
	if v := os.Getenv("AMIE_BASE_URL"); v != "" {
		cfg.AMIE.BaseURL = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	}
}
