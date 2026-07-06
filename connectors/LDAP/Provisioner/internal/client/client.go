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

// Package client is the LDAP protocol wrapper the LDAP Provisioner uses to
// read and write directory entries. It parallels the REST client the
// COmanage Identity-Provisioner uses; different wire protocol, same
// architectural role.
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

// ErrNotFound is returned when a search comes back empty. Mirrors the
// COmanage client's ErrNotFound so orchestration code can use errors.Is.
var ErrNotFound = errors.New("ldap: not found")

// Config carries the connection parameters and the per-cluster identity
// the connector serves. CustosClusterID lets a subscriber filter events so
// a single deployment can host multiple provisioner instances side by side.
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

	// MinUID is the lowest POSIX uidNumber the allocator will hand out.
	// Sites usually reserve 0..999 for system, 1000..49999 for local
	// users, and 50000+ for federated / auto-provisioned users. Default
	// 50000 when unset.
	MinUID int64

	// GroupBaseDN is the container for posixGroup entries — e.g.
	// "ou=groups,dc=example,dc=edu". When empty, the connector skips
	// posixGroup creation entirely (fine on systems using automatic
	// private groups; needed for strict SSSD setups).
	GroupBaseDN string
}

// DefaultMinUID is the value used when Config.MinUID is zero. Chosen to
// sit above the typical local-user range on RHEL / Debian derivatives.
const DefaultMinUID int64 = 50000

// Connection abstracts the go-ldap operations the client uses. Defined here
// so unit tests can substitute a fake without dialing a real server.
type Connection interface {
	Bind(username, password string) error
	Add(req *ldap.AddRequest) error
	Modify(req *ldap.ModifyRequest) error
	Search(req *ldap.SearchRequest) (*ldap.SearchResult, error)
	Close() error
}

// Dialer opens a Connection. Injected so tests can bypass real networking.
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

// PosixAccount is the subset of LDAP attributes the connector maintains
// for each user. UIDNumber and GIDNumber are required — schema-conformant
// posixAccount entries must have both.
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

// Client is a thin, connection-reusing wrapper over go-ldap. All public
// methods take Client.mu; helpers with the `Locked` suffix assume the
// caller already holds it.
type Client struct {
	cfg    Config
	dialer Dialer

	mu   sync.Mutex
	conn Connection
}

// New constructs a Client backed by the real go-ldap dialer.
func New(cfg Config) (*Client, error) {
	if cfg.URL == "" || cfg.BindDN == "" || cfg.BaseDN == "" {
		return nil, errors.New("client.New: Config requires URL, BindDN, and BaseDN")
	}
	return &Client{cfg: cfg, dialer: defaultDialer{}}, nil
}

// NewWithDialer is used by tests to inject a fake dialer.
func NewWithDialer(cfg Config, d Dialer) *Client {
	return &Client{cfg: cfg, dialer: d}
}

// Config returns the config the client was built with.
func (c *Client) Config() Config { return c.cfg }

// Close releases the underlying connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeConn()
}

// FindPosixAccount searches for a posixAccount entry by UID and returns
// its DN plus a map of the retrieved attributes. Returns (empty, nil, nil)
// when no entry matches.
func (c *Client) FindPosixAccount(uid string) (string, map[string][]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.findPosixAccountLocked(uid)
}

// AddPosixAccount creates a new posixAccount + inetOrgPerson entry.
// Callers supply the uidNumber (typically from internal/store.UIDSequence,
// which is the persistent monotonic allocator that guarantees no reuse
// after entry deletion and serialises cross-process races via InnoDB
// row locking).
func (c *Client) AddPosixAccount(a PosixAccount) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.addPosixAccountLocked(a)
}

// ModifyPosixAccount replaces the mutable attributes of an existing entry.
// The uid RDN and the numeric IDs are not modified — a change in either
// would be a different account.
func (c *Client) ModifyPosixAccount(dn string, a PosixAccount) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.modifyPosixAccountLocked(dn, a)
}

// AllocateNextUID scans BaseDN for all posixAccount entries, reads their
// uidNumber attribute, and returns max(uidNumber) + 1, floored at minUID.
//
// Used exclusively for one-time seeding of the persistent uid counter
// at connector startup — the loader calls this so a fresh deployment
// initialises above any entries already in LDAP from out-of-band
// provisioning. Steady-state allocations go through internal/store's
// UIDSequence, which is monotonic across restarts and never regresses
// when LDAP entries are deleted.
func (c *Client) AllocateNextUID(minUID int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.allocateNextUIDLocked(minUID)
}

// FindPosixGroup searches for a posixGroup entry by cn under GroupBaseDN
// and returns its DN. Returns "" when no entry matches. Errors when
// GroupBaseDN is empty — callers should check the config first.
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

// AddPosixGroup creates a posixGroup entry at cn=<cn>,<GroupBaseDN>.
// Errors when GroupBaseDN is empty. gidNumber must be positive.
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

// IsConstraintViolation reports whether err is an LDAP constraint /
// value-already-exists error — the signal that a concurrent writer
// claimed the uidNumber we chose.
func IsConstraintViolation(err error) bool {
	return ldap.IsErrorWithCode(err, ldap.LDAPResultConstraintViolation) ||
		ldap.IsErrorWithCode(err, ldap.LDAPResultAttributeOrValueExists)
}

// IsAlreadyExists reports whether err is an "entry already exists" LDAP
// error — used by group creation to treat concurrent adds as idempotent.
func IsAlreadyExists(err error) bool {
	return ldap.IsErrorWithCode(err, ldap.LDAPResultEntryAlreadyExists)
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
	if err := validate(a); err != nil {
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

// searchTimeoutSeconds returns Config.Timeout as whole seconds for use
// as the LDAP SearchRequest TimeLimit field. Clamped to a minimum of 1
// so a sub-second configured timeout does not become 0 (which the LDAP
// protocol interprets as "no time limit" — the opposite of the caller's
// intent).
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

func validate(a PosixAccount) error {
	if a.UID == "" {
		return errors.New("PosixAccount: UID is required")
	}
	if a.UIDNumber <= 0 {
		return errors.New("PosixAccount: UIDNumber must be a positive integer")
	}
	if a.GIDNumber <= 0 {
		return errors.New("PosixAccount: GIDNumber must be a positive integer")
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
