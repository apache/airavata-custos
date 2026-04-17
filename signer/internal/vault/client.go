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

// Package vault provides Vault KV v2 operations for CA key storage and serial counters.
package vault

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"time"

	vaultapi "github.com/openbao/openbao/api/v2"

	"github.com/apache/airavata-custos/signer/internal/config"
)

type CAKeyPair struct {
	PrivateKey []byte // PEM-encoded private key
	PublicKey  []byte // PEM-encoded public key
	Algorithm  string // e.g., "ed25519"
	CreatedAt  string
}

type CAMetadata struct {
	SerialCounter   int64
	LastRotationAt  string
	NextRotationAt  string
	RotationPeriodH int
	OverlapHours    int
}

type Client struct {
	client      *vaultapi.Client
	mountPath   string
	rotationCfg config.RotationConfig
}

func NewClient(cfg config.VaultConfig, rotationCfg config.RotationConfig) (*Client, error) {
	vaultCfg := vaultapi.DefaultConfig()
	vaultCfg.Address = cfg.Address
	vaultCfg.Timeout = time.Duration(cfg.TimeoutSeconds) * time.Second

	client, err := vaultapi.NewClient(vaultCfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}
	client.SetToken(cfg.Token)

	return &Client{
		client:      client,
		mountPath:   cfg.MountPath,
		rotationCfg: rotationCfg,
	}, nil
}

func (c *Client) Healthy(ctx context.Context) error {
	_, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault health check: %w", err)
	}
	return nil
}

func (c *Client) kvPath(parts ...string) string {
	path := c.mountPath + "/data"
	for _, p := range parts {
		path += "/" + p
	}
	return path
}

// GetCurrentCAKey returns the current CA key, auto-creating one if it doesn't exist.
func (c *Client) GetCurrentCAKey(ctx context.Context, tenantID, clientID string) (*CAKeyPair, error) {
	kp, err := c.readCAKey(ctx, tenantID, clientID, "current")
	if err != nil {
		return nil, err
	}
	if kp != nil {
		return kp, nil
	}

	kp, err = generateEd25519CAKey()
	if err != nil {
		return nil, err
	}

	if err := c.writeCAKey(ctx, tenantID, clientID, "current", kp); err != nil {
		return nil, err
	}

	if err := c.initMetadata(ctx, tenantID, clientID); err != nil {
		return nil, err
	}

	return kp, nil
}

// GetNextCAKey returns the next CA key, or nil if none exists.
func (c *Client) GetNextCAKey(ctx context.Context, tenantID, clientID string) (*CAKeyPair, error) {
	return c.readCAKey(ctx, tenantID, clientID, "next")
}

// IncrementSerialCounter atomically reads, increments, and writes the serial
// counter using Vault KV v2 CAS (Check-and-Set) to prevent lost updates.
func (c *Client) IncrementSerialCounter(ctx context.Context, tenantID, clientID string) (int64, error) {
	meta, version, err := c.readMetadataWithVersion(ctx, tenantID, clientID)
	if err != nil {
		return 0, err
	}
	if meta == nil {
		if err := c.initMetadata(ctx, tenantID, clientID); err != nil {
			return 0, err
		}
		return 1, nil
	}

	newSerial := meta.SerialCounter + 1
	meta.SerialCounter = newSerial
	if err := c.writeMetadataCAS(ctx, tenantID, clientID, meta, version); err != nil {
		return 0, err
	}
	return newSerial, nil
}

// RotateCA promotes next to current and generates a new next key.
func (c *Client) RotateCA(ctx context.Context, tenantID, clientID string) (currentFP, nextFP string, err error) {
	nextKey, err := c.readCAKey(ctx, tenantID, clientID, "next")
	if err != nil {
		return "", "", err
	}

	if nextKey != nil {
		// Promote next to current
		if err := c.writeCAKey(ctx, tenantID, clientID, "current", nextKey); err != nil {
			return "", "", err
		}
	} else {
		// Generate fresh current
		newCurrent, err := generateEd25519CAKey()
		if err != nil {
			return "", "", err
		}
		if err := c.writeCAKey(ctx, tenantID, clientID, "current", newCurrent); err != nil {
			return "", "", err
		}
	}

	// Generate new next key
	newNext, err := generateEd25519CAKey()
	if err != nil {
		return "", "", err
	}
	if err := c.writeCAKey(ctx, tenantID, clientID, "next", newNext); err != nil {
		return "", "", err
	}

	now := time.Now().UTC()
	meta, err := c.readMetadata(ctx, tenantID, clientID)
	if err != nil {
		return "", "", err
	}
	if meta == nil {
		meta = &CAMetadata{SerialCounter: 0}
	}
	meta.LastRotationAt = now.Format(time.RFC3339)
	meta.NextRotationAt = now.Add(time.Duration(c.rotationCfg.PeriodHours) * time.Hour).Format(time.RFC3339)
	meta.RotationPeriodH = c.rotationCfg.PeriodHours
	meta.OverlapHours = c.rotationCfg.OverlapHours
	if err := c.writeMetadata(ctx, tenantID, clientID, meta); err != nil {
		return "", "", err
	}

	// Read back keys to get fingerprints
	currentKey, err := c.readCAKey(ctx, tenantID, clientID, "current")
	if err != nil {
		return "", "", err
	}

	return string(currentKey.PublicKey), string(newNext.PublicKey), nil
}

func (c *Client) GetMetadata(ctx context.Context, tenantID, clientID string) (*CAMetadata, error) {
	return c.readMetadata(ctx, tenantID, clientID)
}

func (c *Client) readCAKey(ctx context.Context, tenantID, clientID, slot string) (*CAKeyPair, error) {
	path := c.kvPath(tenantID, clientID, slot)
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading vault path %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok || data == nil {
		return nil, nil
	}

	kp := &CAKeyPair{}
	if v, ok := data["private_key"].(string); ok {
		kp.PrivateKey = []byte(v)
	}
	if v, ok := data["public_key"].(string); ok {
		kp.PublicKey = []byte(v)
	}
	if v, ok := data["algorithm"].(string); ok {
		kp.Algorithm = v
	}
	if v, ok := data["created_at"].(string); ok {
		kp.CreatedAt = v
	}

	if len(kp.PrivateKey) == 0 {
		return nil, nil
	}
	return kp, nil
}

func (c *Client) writeCAKey(ctx context.Context, tenantID, clientID, slot string, kp *CAKeyPair) error {
	path := c.kvPath(tenantID, clientID, slot)
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"private_key": string(kp.PrivateKey),
			"public_key":  string(kp.PublicKey),
			"algorithm":   kp.Algorithm,
			"created_at":  kp.CreatedAt,
		},
	}
	_, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return fmt.Errorf("writing vault path %s: %w", path, err)
	}
	return nil
}

func (c *Client) readMetadata(ctx context.Context, tenantID, clientID string) (*CAMetadata, error) {
	meta, _, err := c.readMetadataWithVersion(ctx, tenantID, clientID)
	return meta, err
}

func (c *Client) readMetadataWithVersion(ctx context.Context, tenantID, clientID string) (*CAMetadata, int, error) {
	path := c.kvPath(tenantID, clientID, "metadata")
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, 0, fmt.Errorf("reading vault metadata: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, 0, nil
	}

	// Extract KV v2 version from metadata.
	// The Vault Go client decodes numbers as json.Number, not float64.
	version := 0
	if md, ok := secret.Data["metadata"].(map[string]interface{}); ok {
		if v, ok := md["version"].(json.Number); ok {
			if n, err := v.Int64(); err == nil {
				version = int(n)
			}
		}
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok || data == nil {
		return nil, 0, nil
	}

	meta := &CAMetadata{}
	if v, ok := data["serial_counter"].(string); ok {
		meta.SerialCounter, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := data["serial_counter"].(float64); ok {
		meta.SerialCounter = int64(v)
	}
	if v, ok := data["last_rotation_at"].(string); ok {
		meta.LastRotationAt = v
	}
	if v, ok := data["next_rotation_at"].(string); ok {
		meta.NextRotationAt = v
	}
	if v, ok := data["rotation_period_hours"].(string); ok {
		meta.RotationPeriodH, _ = strconv.Atoi(v)
	}
	if v, ok := data["rotation_period_hours"].(float64); ok {
		meta.RotationPeriodH = int(v)
	}
	if v, ok := data["overlap_hours"].(string); ok {
		meta.OverlapHours, _ = strconv.Atoi(v)
	}
	if v, ok := data["overlap_hours"].(float64); ok {
		meta.OverlapHours = int(v)
	}
	return meta, version, nil
}

func (c *Client) writeMetadata(ctx context.Context, tenantID, clientID string, meta *CAMetadata) error {
	path := c.kvPath(tenantID, clientID, "metadata")
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"serial_counter":        strconv.FormatInt(meta.SerialCounter, 10),
			"last_rotation_at":      meta.LastRotationAt,
			"next_rotation_at":      meta.NextRotationAt,
			"rotation_period_hours": strconv.Itoa(meta.RotationPeriodH),
			"overlap_hours":         strconv.Itoa(meta.OverlapHours),
		},
	}
	_, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return fmt.Errorf("writing vault metadata: %w", err)
	}
	return nil
}

func (c *Client) writeMetadataCAS(ctx context.Context, tenantID, clientID string, meta *CAMetadata, cas int) error {
	path := c.kvPath(tenantID, clientID, "metadata")
	data := map[string]interface{}{
		"options": map[string]interface{}{
			"cas": cas,
		},
		"data": map[string]interface{}{
			"serial_counter":        strconv.FormatInt(meta.SerialCounter, 10),
			"last_rotation_at":      meta.LastRotationAt,
			"next_rotation_at":      meta.NextRotationAt,
			"rotation_period_hours": strconv.Itoa(meta.RotationPeriodH),
			"overlap_hours":         strconv.Itoa(meta.OverlapHours),
		},
	}
	_, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return fmt.Errorf("writing vault metadata (CAS=%d): %w", cas, err)
	}
	return nil
}

func (c *Client) initMetadata(ctx context.Context, tenantID, clientID string) error {
	now := time.Now().UTC()
	meta := &CAMetadata{
		SerialCounter:   0,
		LastRotationAt:  now.Format(time.RFC3339),
		NextRotationAt:  now.Add(time.Duration(c.rotationCfg.PeriodHours) * time.Hour).Format(time.RFC3339),
		RotationPeriodH: c.rotationCfg.PeriodHours,
		OverlapHours:    c.rotationCfg.OverlapHours,
	}
	return c.writeMetadata(ctx, tenantID, clientID, meta)
}

func generateEd25519CAKey() (*CAKeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating ed25519 key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshaling private key: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("marshaling public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return &CAKeyPair{
		PrivateKey: privPEM,
		PublicKey:  pubPEM,
		Algorithm:  "ed25519",
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ValidationCredentials holds per-client - principal validation.
// Stored in Vault at ssh-ca/{tenant_id}/{client_id}/validation.
type ValidationCredentials struct {
	Type              string `json:"type"`                         // "ldap" or "comanage"
	LDAPUrl           string `json:"ldap_url,omitempty"`           // LDAP
	BindDN            string `json:"bind_dn,omitempty"`            // LDAP
	BindPassword      string `json:"bind_password,omitempty"`      // LDAP
	BaseDN            string `json:"base_dn,omitempty"`            // LDAP
	SearchFilter      string `json:"search_filter,omitempty"`      // LDAP — %s is the OIDC subject
	UsernameAttribute string `json:"username_attribute,omitempty"` // LDAP — attribute for POSIX username (default: "uid")
	VerifySSL         *bool  `json:"verify_ssl,omitempty"`         // LDAP / COmanage
	RegistryURL       string `json:"registry_url,omitempty"`       // COmanage
	APIUser           string `json:"api_user,omitempty"`           // COmanage
	APIKey            string `json:"api_key,omitempty"`            // COmanage
	APIPath           string `json:"api_path,omitempty"`           // COmanage
}

// GetValidationCredentials reads per-client validation credentials from Vault.
// Returns (nil, nil) if no credentials exist at the path.
func (c *Client) GetValidationCredentials(ctx context.Context, tenantID, clientID string) (*ValidationCredentials, error) {
	path := c.kvPath(tenantID, clientID, "validation")
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading validation credentials at %s: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok || data == nil {
		return nil, nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshaling validation data: %w", err)
	}

	var creds ValidationCredentials
	if err := json.Unmarshal(raw, &creds); err != nil {
		return nil, fmt.Errorf("unmarshaling validation credentials: %w", err)
	}

	if creds.Type == "" {
		return nil, nil
	}

	return &creds, nil
}
