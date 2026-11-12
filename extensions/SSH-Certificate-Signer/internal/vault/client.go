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
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	vaultapi "github.com/openbao/openbao/api/v2"
	"golang.org/x/crypto/ssh"

	"github.com/apache/airavata-custos/signer/internal/config"
)

// ErrSerialConflict marks a lost CAS race on the serial counter, not a store failure.
var ErrSerialConflict = errors.New("serial counter conflict")

var ErrCANotFound = errors.New("no CA key in vault")

func isCASMismatch(err error) bool {
	var respErr *vaultapi.ResponseError
	if !errors.As(err, &respErr) {
		return false
	}
	for _, msg := range respErr.Errors {
		if strings.Contains(msg, "check-and-set parameter did not match") {
			return true
		}
	}
	return false
}

type CAKeyPair struct {
	PrivateKey []byte // OpenSSH private key
	PublicKey  []byte // OpenSSH public key line, as in authorized_keys
	CreatedAt  string // RFC 3339, empty when the stored CA has none
}

type CAMetadata struct {
	SerialCounter   int64
	LastRotationAt  string
	NextRotationAt  string
	RotationPeriodH int
	OverlapHours    int
}

type Client struct {
	client         *vaultapi.Client
	mountPath      string
	caKeyPath      string
	caMetadataPath string
	rotationCfg    config.RotationConfig
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
		client:    client,
		mountPath: cfg.MountPath,
		caKeyPath: cfg.CAKeyPath,
		// The serial counter and rotation settings stay next to the CA key.
		caMetadataPath: path.Join(path.Dir(cfg.CAKeyPath), "metadata"),
		rotationCfg:    rotationCfg,
	}, nil
}

func (c *Client) Healthy(ctx context.Context) error {
	_, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault health check: %w", err)
	}
	return nil
}

// kvPath turns a secret path into its KV v2 API path, {mount}/data/{secretPath}.
func (c *Client) kvPath(secretPath string) string {
	return c.mountPath + "/data/" + secretPath
}

// GetCurrentCAKey reads the CA keys from the private_key and public_key fields at the configured path.
// TODO CA from file path support
func (c *Client) GetCurrentCAKey(ctx context.Context) (*CAKeyPair, error) {
	kvPath := c.kvPath(c.caKeyPath)
	secret, err := c.client.Logical().ReadWithContext(ctx, kvPath)
	if err != nil {
		return nil, fmt.Errorf("reading vault path %s: %w", kvPath, err)
	}
	var data map[string]any
	if secret != nil {
		data, _ = secret.Data["data"].(map[string]any)
	}
	privateKey, _ := data["private_key"].(string)
	publicKey, _ := data["public_key"].(string)
	createdAt, _ := data["created_at"].(string)
	if privateKey == "" {
		return nil, fmt.Errorf("%w at %s", ErrCANotFound, kvPath)
	}
	// Not ErrCANotFound, so a missing public key never makes the signer replace an existing CA.
	if publicKey == "" {
		return nil, fmt.Errorf("CA public key missing at %s", kvPath)
	}
	return &CAKeyPair{PrivateKey: []byte(privateKey), PublicKey: []byte(publicKey), CreatedAt: createdAt}, nil
}

// GenerateCAKey generates a new CA key and stores it as the current one.
// TODO - also publish the new public key to the login nodes' trusted CA file.
func (c *Client) GenerateCAKey(ctx context.Context) (*CAKeyPair, error) {
	kp, err := generateEd25519CAKey()
	if err != nil {
		return nil, err
	}

	kvPath := c.kvPath(c.caKeyPath)
	data := map[string]any{"data": map[string]any{
		"private_key": string(kp.PrivateKey),
		"public_key":  string(kp.PublicKey),
		"created_at":  kp.CreatedAt,
	}}
	if _, err := c.client.Logical().WriteWithContext(ctx, kvPath, data); err != nil {
		return nil, fmt.Errorf("writing vault path %s: %w", kvPath, err)
	}

	if err := c.initMetadata(ctx); err != nil {
		return nil, err
	}

	return kp, nil
}

// IncrementSerialCounter atomically reads, increments, and writes the serial
// counter using Vault KV v2 CAS (Check-and-Set) to prevent lost updates.
func (c *Client) IncrementSerialCounter(ctx context.Context) (int64, error) {
	meta, version, err := c.readMetadataWithVersion(ctx)
	if err != nil {
		return 0, err
	}
	if meta == nil {
		if err := c.initMetadata(ctx); err != nil {
			return 0, err
		}
		return 1, nil
	}

	newSerial := meta.SerialCounter + 1
	meta.SerialCounter = newSerial
	if err := c.writeMetadataCAS(ctx, meta, version); err != nil {
		return 0, err
	}
	return newSerial, nil
}

func (c *Client) readMetadataWithVersion(ctx context.Context) (*CAMetadata, int, error) {
	path := c.kvPath(c.caMetadataPath)
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

func (c *Client) writeMetadata(ctx context.Context, meta *CAMetadata) error {
	kvPath := c.kvPath(c.caMetadataPath)
	data := map[string]interface{}{
		"data": map[string]interface{}{
			"serial_counter":        strconv.FormatInt(meta.SerialCounter, 10),
			"last_rotation_at":      meta.LastRotationAt,
			"next_rotation_at":      meta.NextRotationAt,
			"rotation_period_hours": strconv.Itoa(meta.RotationPeriodH),
			"overlap_hours":         strconv.Itoa(meta.OverlapHours),
		},
	}
	_, err := c.client.Logical().WriteWithContext(ctx, kvPath, data)
	if err != nil {
		return fmt.Errorf("writing vault metadata: %w", err)
	}
	return nil
}

func (c *Client) writeMetadataCAS(ctx context.Context, meta *CAMetadata, cas int) error {
	kvPath := c.kvPath(c.caMetadataPath)
	data := map[string]any{
		"options": map[string]any{
			"cas": cas,
		},
		"data": map[string]any{
			"serial_counter":        strconv.FormatInt(meta.SerialCounter, 10),
			"last_rotation_at":      meta.LastRotationAt,
			"next_rotation_at":      meta.NextRotationAt,
			"rotation_period_hours": strconv.Itoa(meta.RotationPeriodH),
			"overlap_hours":         strconv.Itoa(meta.OverlapHours),
		},
	}
	_, err := c.client.Logical().WriteWithContext(ctx, kvPath, data)
	if err != nil {
		if isCASMismatch(err) {
			return fmt.Errorf("%w (CAS=%d): %v", ErrSerialConflict, cas, err)
		}
		return fmt.Errorf("writing vault metadata (CAS=%d): %w", cas, err)
	}
	return nil
}

func (c *Client) initMetadata(ctx context.Context) error {
	now := time.Now().UTC()
	meta := &CAMetadata{
		SerialCounter:   0,
		LastRotationAt:  now.Format(time.RFC3339),
		NextRotationAt:  now.Add(time.Duration(c.rotationCfg.PeriodHours) * time.Hour).Format(time.RFC3339),
		RotationPeriodH: c.rotationCfg.PeriodHours,
		OverlapHours:    c.rotationCfg.OverlapHours,
	}
	return c.writeMetadata(ctx, meta)
}

func generateEd25519CAKey() (*CAKeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating ed25519 key: %w", err)
	}

	privBlock, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, fmt.Errorf("marshaling private key: %w", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("marshaling public key: %w", err)
	}

	return &CAKeyPair{
		PrivateKey: pem.EncodeToMemory(privBlock),
		PublicKey:  ssh.MarshalAuthorizedKey(sshPub),
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ValidationCredentials holds the deployment's principal validation settings.
// Stored at {mount}/validation.
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

// GetValidationCredentials reads the principal validation settings.
// Returns (nil, nil) if none exist at the path.
func (c *Client) GetValidationCredentials(ctx context.Context) (*ValidationCredentials, error) {
	path := c.kvPath("validation")
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
