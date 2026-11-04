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

// Package client reads and writes POSIX entries in an LDAP directory.
package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
)

var ErrNotFound = errors.New("ldap: not found")

type Config struct {
	URL             string
	BindDN          string
	BindPassword    string
	BaseDN          string
	VerifySSL       bool
	CustosClusterID string
	DefaultShell    string
	HomedirPrefix   string
	Timeout         time.Duration

	// MinUID is the lowest uidNumber or gidNumber the client writes.
	MinUID int64

	// GroupBaseDN holds the posixGroup entries. When empty no groups are
	// written, for hosts that use automatic private groups.
	GroupBaseDN string
}

// DefaultMinUID sits above the usual local-user range.
const DefaultMinUID int64 = 50000

// Connection is the part of go-ldap the client uses, so tests can fake it.
type Connection interface {
	Bind(username, password string) error
	Add(req *ldap.AddRequest) error
	Modify(req *ldap.ModifyRequest) error
	Del(req *ldap.DelRequest) error
	Search(req *ldap.SearchRequest) (*ldap.SearchResult, error)
	Close() error
}

type Dialer interface {
	Dial(url string, verifySSL bool, timeout time.Duration) (Connection, error)
}

type defaultDialer struct{}

func (defaultDialer) Dial(url string, verifySSL bool, timeout time.Duration) (Connection, error) {
	opts := []ldap.DialOpt{
		ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: !verifySSL}),
	}
	if timeout > 0 {
		opts = append(opts, ldap.DialWithDialer(&net.Dialer{Timeout: timeout}))
	}
	return ldap.DialURL(url, opts...)
}

type PosixAccount struct {
	UID           string
	UIDNumber     int64
	GIDNumber     int64
	GivenName     string
	Surname       string
	Mail          string
	HomeDirectory string
	LoginShell    string
}

// Client reuses one connection. Public methods take mu; the *Locked helpers
// expect the caller to hold it.
type Client struct {
	cfg    Config
	dialer Dialer

	mu   sync.Mutex
	conn Connection
}

func New(cfg Config) (*Client, error) {
	if cfg.URL == "" || cfg.BindDN == "" || cfg.BaseDN == "" {
		return nil, errors.New("client.New: Config requires URL, BindDN, and BaseDN")
	}
	return &Client{cfg: cfg, dialer: defaultDialer{}}, nil
}

func NewWithDialer(cfg Config, d Dialer) *Client {
	return &Client{cfg: cfg, dialer: d}
}

func (c *Client) Config() Config { return c.cfg }

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeConn()
}

// FindPosixAccount returns the entry's DN and attributes, or "" when there
// is none.
func (c *Client) FindPosixAccount(uid string) (string, map[string][]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.findPosixAccountLocked(uid)
}

// GetPosixAccount returns the attributes of the posixAccount at dn, or nil
// when there is none.
func (c *Client) GetPosixAccount(dn string) (map[string][]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, err := c.ensureConn()
	if err != nil {
		return nil, err
	}
	req := ldap.NewSearchRequest(
		dn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, c.searchTimeoutSeconds(), false,
		"(objectClass=posixAccount)",
		[]string{"uid", "uidNumber", "gidNumber"},
		nil,
	)
	result, err := conn.Search(req)
	if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
		return nil, nil
	}
	if err != nil {
		c.closeConn()
		return nil, fmt.Errorf("read posixAccount %s: %w", dn, err)
	}
	if len(result.Entries) == 0 {
		return nil, nil
	}
	attrs := make(map[string][]string, len(result.Entries[0].Attributes))
	for _, a := range result.Entries[0].Attributes {
		attrs[a.Name] = a.Values
	}
	return attrs, nil
}

func (c *Client) AddPosixAccount(a PosixAccount) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.addPosixAccountLocked(a)
}

// ModifyPosixAccount leaves uid and the numeric IDs alone: changing them
// would make it a different account.
func (c *Client) ModifyPosixAccount(dn string, a PosixAccount) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.modifyPosixAccountLocked(dn, a)
}

// AllocateNextUID returns max(uidNumber) + 1 across BaseDN, floored at
// minUID. The loader calls it once at startup so assigned numbers start
// above entries created outside Custos.
func (c *Client) AllocateNextUID(minUID int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.allocateNextUIDLocked(minUID)
}

// FindPosixGroup returns the group's DN, or "" when there is none.
func (c *Client) FindPosixGroup(cn string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.GroupBaseDN == "" {
		return "", errors.New("FindPosixGroup: GroupBaseDN not configured")
	}

	conn, err := c.ensureConn()
	if err != nil {
		return "", err
	}
	filter := fmt.Sprintf("(&(objectClass=posixGroup)(cn=%s))", ldap.EscapeFilter(cn))
	req := ldap.NewSearchRequest(
		c.cfg.GroupBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		2, c.searchTimeoutSeconds(), false,
		filter,
		[]string{"cn", "gidNumber"},
		nil,
	)
	result, err := conn.Search(req)
	if err != nil {
		c.closeConn()
		return "", fmt.Errorf("search posixGroup cn=%s: %w", cn, err)
	}
	if len(result.Entries) == 0 {
		return "", nil
	}
	if len(result.Entries) > 1 {
		return "", fmt.Errorf("search posixGroup cn=%s returned %d entries", cn, len(result.Entries))
	}
	return result.Entries[0].DN, nil
}

func (c *Client) AddPosixGroup(cn string, gidNumber int64) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.GroupBaseDN == "" {
		return "", errors.New("AddPosixGroup: GroupBaseDN not configured")
	}
	if cn == "" {
		return "", errors.New("AddPosixGroup: cn is required")
	}
	if gidNumber <= 0 {
		return "", errors.New("AddPosixGroup: gidNumber must be a positive integer")
	}

	conn, err := c.ensureConn()
	if err != nil {
		return "", err
	}

	dn := fmt.Sprintf("cn=%s,%s", cn, c.cfg.GroupBaseDN)
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", []string{"top", "posixGroup"})
	req.Attribute("cn", []string{cn})
	req.Attribute("gidNumber", []string{strconv.FormatInt(gidNumber, 10)})

	if err := conn.Add(req); err != nil {
		c.closeConn()
		return "", fmt.Errorf("add posixGroup %s: %w", dn, err)
	}
	return dn, nil
}

// IsConstraintViolation reports whether the directory rejected a value
// another entry already holds, such as a uidNumber.
func IsConstraintViolation(err error) bool {
	return ldap.IsErrorWithCode(err, ldap.LDAPResultConstraintViolation) ||
		ldap.IsErrorWithCode(err, ldap.LDAPResultAttributeOrValueExists)
}

func IsAlreadyExists(err error) bool {
	return ldap.IsErrorWithCode(err, ldap.LDAPResultEntryAlreadyExists)
}

// DeleteEntry removes dn. An entry that is already gone is not an error.
func (c *Client) DeleteEntry(dn string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, err := c.ensureConn()
	if err != nil {
		return err
	}
	if err := conn.Del(ldap.NewDelRequest(dn, nil)); err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultNoSuchObject) {
			return nil
		}
		c.closeConn()
		return fmt.Errorf("delete %s: %w", dn, err)
	}
	return nil
}

// ---- private locked helpers ------------------------------------------

func (c *Client) findPosixAccountLocked(uid string) (string, map[string][]string, error) {
	conn, err := c.ensureConn()
	if err != nil {
		return "", nil, err
	}
	filter := fmt.Sprintf("(&(objectClass=posixAccount)(uid=%s))", ldap.EscapeFilter(uid))
	req := ldap.NewSearchRequest(
		c.cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		2, c.searchTimeoutSeconds(), false,
		filter,
		[]string{"uid", "uidNumber", "gidNumber", "cn", "givenName", "sn", "mail", "homeDirectory", "loginShell"},
		nil,
	)
	result, err := conn.Search(req)
	if err != nil {
		c.closeConn()
		return "", nil, fmt.Errorf("search posixAccount uid=%s: %w", uid, err)
	}
	if len(result.Entries) == 0 {
		return "", nil, nil
	}
	if len(result.Entries) > 1 {
		return "", nil, fmt.Errorf("search posixAccount uid=%s returned %d entries", uid, len(result.Entries))
	}
	entry := result.Entries[0]
	attrs := make(map[string][]string, len(entry.Attributes))
	for _, a := range entry.Attributes {
		attrs[a.Name] = a.Values
	}
	return entry.DN, attrs, nil
}

func (c *Client) addPosixAccountLocked(a PosixAccount) (string, error) {
	if err := validate(a, c.cfg.MinUID); err != nil {
		return "", err
	}
	conn, err := c.ensureConn()
	if err != nil {
		return "", err
	}

	dn := fmt.Sprintf("uid=%s,%s", a.UID, c.cfg.BaseDN)
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", []string{
		"top", "person", "organizationalPerson", "inetOrgPerson", "posixAccount",
	})
	req.Attribute("uid", []string{a.UID})
	req.Attribute("cn", []string{fullName(a)})
	req.Attribute("sn", []string{a.Surname})
	if a.GivenName != "" {
		req.Attribute("givenName", []string{a.GivenName})
	}
	req.Attribute("uidNumber", []string{strconv.FormatInt(a.UIDNumber, 10)})
	req.Attribute("gidNumber", []string{strconv.FormatInt(a.GIDNumber, 10)})
	req.Attribute("homeDirectory", []string{a.HomeDirectory})
	if a.LoginShell != "" {
		req.Attribute("loginShell", []string{a.LoginShell})
	}
	if a.Mail != "" {
		req.Attribute("mail", []string{a.Mail})
	}

	if err := conn.Add(req); err != nil {
		c.closeConn()
		return "", fmt.Errorf("add posixAccount %s: %w", dn, err)
	}
	return dn, nil
}

func (c *Client) modifyPosixAccountLocked(dn string, a PosixAccount) error {
	conn, err := c.ensureConn()
	if err != nil {
		return err
	}
	req := ldap.NewModifyRequest(dn, nil)
	req.Replace("cn", []string{fullName(a)})
	req.Replace("sn", []string{a.Surname})
	if a.GivenName != "" {
		req.Replace("givenName", []string{a.GivenName})
	}
	req.Replace("homeDirectory", []string{a.HomeDirectory})
	if a.LoginShell != "" {
		req.Replace("loginShell", []string{a.LoginShell})
	}
	if a.Mail != "" {
		req.Replace("mail", []string{a.Mail})
	}
	if err := conn.Modify(req); err != nil {
		c.closeConn()
		return fmt.Errorf("modify posixAccount %s: %w", dn, err)
	}
	return nil
}

func (c *Client) allocateNextUIDLocked(minUID int64) (int64, error) {
	if minUID <= 0 {
		minUID = DefaultMinUID
	}
	conn, err := c.ensureConn()
	if err != nil {
		return 0, err
	}
	req := ldap.NewSearchRequest(
		c.cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, c.searchTimeoutSeconds(), false,
		"(&(objectClass=posixAccount)(uidNumber=*))",
		[]string{"uidNumber"},
		nil,
	)
	result, err := conn.Search(req)
	if err != nil {
		c.closeConn()
		return 0, fmt.Errorf("search posixAccount for max uidNumber: %w", err)
	}

	var max int64
	for _, entry := range result.Entries {
		for _, v := range entry.GetAttributeValues("uidNumber") {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				continue
			}
			if n > max {
				max = n
			}
		}
	}
	next := max + 1
	if next < minUID {
		next = minUID
	}
	return next, nil
}

// searchTimeoutSeconds is at least 1: a TimeLimit of 0 means no limit.
func (c *Client) searchTimeoutSeconds() int {
	n := int(c.cfg.Timeout.Seconds())
	if n < 1 {
		return 1
	}
	return n
}

func (c *Client) ensureConn() (Connection, error) {
	if c.conn != nil {
		return c.conn, nil
	}
	conn, err := c.dialer.Dial(c.cfg.URL, c.cfg.VerifySSL, c.cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("dial LDAP %s: %w", c.cfg.URL, err)
	}
	if err := conn.Bind(c.cfg.BindDN, c.cfg.BindPassword); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("bind LDAP as %s: %w", c.cfg.BindDN, err)
	}
	c.conn = conn
	return conn, nil
}

func (c *Client) closeConn() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

func validate(a PosixAccount, minUID int64) error {
	if a.UID == "" {
		return errors.New("PosixAccount: UID is required")
	}
	if minUID <= 0 {
		minUID = DefaultMinUID
	}
	if a.UIDNumber < minUID {
		return fmt.Errorf("PosixAccount: UIDNumber %d is below min_uid %d", a.UIDNumber, minUID)
	}
	if a.GIDNumber < minUID {
		return fmt.Errorf("PosixAccount: GIDNumber %d is below min_uid %d", a.GIDNumber, minUID)
	}
	if a.HomeDirectory == "" {
		return errors.New("PosixAccount: HomeDirectory is required")
	}
	if a.Surname == "" {
		return errors.New("PosixAccount: Surname is required (sn is a MUST on person)")
	}
	return nil
}

func fullName(a PosixAccount) string {
	if a.GivenName == "" {
		return a.Surname
	}
	return a.GivenName + " " + a.Surname
}
