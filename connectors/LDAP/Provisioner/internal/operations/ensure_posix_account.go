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

package operations

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/models"
)

// identitySourcePrefix is the user_identities.source prefix for
// LDAP-assigned uidNumbers. The full source string is
// identitySourcePrefix + ":" + CustosClusterID so a deployment
// servicing multiple clusters caches each cluster's uidNumber
// separately — the same Custos user can have different POSIX uids
// on different clusters. Symmetric to the COmanage connector's
// source="comanage" tag but scoped per cluster to prevent the
// cross-cluster UID reuse that a flat source= would allow.
const identitySourcePrefix = "ldap"

// identitySource returns the fully-qualified user_identities.source
// value for this connector instance's cluster.
func (o *Orchestrator) identitySource() string {
	if id := o.c.Config().CustosClusterID; id != "" {
		return identitySourcePrefix + ":" + id
	}
	return identitySourcePrefix
}

// maxPosixUsernameLen is the traditional POSIX login-name cap. Matches
// what the pkg/posix allocator already enforces on the AMIE side.
const maxPosixUsernameLen = 32

// validPosixUsername matches a lowercase POSIX-conformant login name.
// Restricting to [a-z0-9_-] also happens to reject every RFC 4514 DN
// metacharacter (comma, plus, quote, backslash, less/greater, semicolon,
// equals, hash, space), so the RDN can be safely concatenated into
// uid=<value>,<BaseDN> without DN escaping.
var validPosixUsername = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

// maxAllocRetries bounds how many times the orchestrator will re-allocate
// after a uidNumber constraint violation from an out-of-band writer that
// already holds our counter's next value. The persistent monotonic
// counter serialises Custos-side races itself (InnoDB row lock on the
// ldap_uid_sequence row), so retries are only reached against uids
// claimed by processes outside this connector.
const maxAllocRetries = 3

// ensurePOSIXAccountImpl provisions a posixAccount entry (and, when
// GroupBaseDN is configured, a matching posixGroup) for the given
// ComputeClusterUser. Mirrors the COmanage connector's shape: resolve
// the identity registry's assigned id, cache it in user_identities,
// materialise the POSIX record, and add the primary group.
//
// Direct-LDAP path specifics:
//   - The identity registry is a persistent monotonic counter
//     (internal/store.UIDSequence, one row per cluster in
//     ldap_uid_sequence). Never regresses on entry deletion, so a new
//     user cannot inherit a deleted user's numeric uid.
//   - gidnumber = uidnumber (one-group-per-user), same pattern the
//     COmanage flow uses on the CoGroup identifier.
//   - user_identities(source="ldap:<clusterID>", external_id=<uidNumber>)
//     caches the assignment so re-provisioning the same user is O(1).
//   - When GroupBaseDN is set, a posixGroup with cn=<LocalUsername>,
//     gidNumber=<uidNumber> is created too; without GroupBaseDN the
//     connector assumes auto-private-groups on the client side.
func (o *Orchestrator) ensurePOSIXAccountImpl(ctx context.Context, cu *models.ComputeClusterUser) error {
	log := slog.With(
		"correlation_id", cu.ID,
		"custos_user_id", cu.UserID,
		"local_username", cu.LocalUsername,
		"base_dn", o.c.Config().BaseDN,
	)

	if cu.LocalUsername == "" {
		err := fmt.Errorf("compute_cluster_user %s has empty local_username", cu.ID)
		o.dlq(ctx, cu, "validate_local_username", err)
		return err
	}
	if len(cu.LocalUsername) > maxPosixUsernameLen || !validPosixUsername.MatchString(cu.LocalUsername) {
		err := fmt.Errorf("compute_cluster_user %s has invalid local_username %q (must be POSIX-safe: lowercase, [a-z_][a-z0-9_-]*, <=%d chars)",
			cu.ID, cu.LocalUsername, maxPosixUsernameLen)
		o.dlq(ctx, cu, "validate_local_username", err)
		return err
	}

	user, err := o.core.GetUser(ctx, cu.UserID)
	if err != nil {
		o.dlq(ctx, cu, "get_custos_user", err)
		return fmt.Errorf("get custos user %s: %w", cu.UserID, err)
	}
	if user == nil {
		err := fmt.Errorf("custos user %s not found", cu.UserID)
		o.dlq(ctx, cu, "get_custos_user", err)
		return err
	}

	// 1. Cache lookup — has this user already been assigned a uidNumber?
	cached, err := o.findCachedUID(ctx, cu.UserID)
	if err != nil {
		o.dlq(ctx, cu, "list_user_identities", err)
		return err
	}
	if cached > 0 {
		log.Info("ldap provisioner: uidNumber cache hit", "uid", cached)
		return o.ensureEntry(ctx, cu, user, cached)
	}

	// 2. LDAP lookup — the entry may already exist (out-of-band or a
	//    prior run that failed to cache). Adopt its uidNumber.
	adopted, err := o.tryAdoptExisting(ctx, cu, user)
	if err != nil {
		return err
	}
	if adopted {
		return nil
	}

	// 3. New user — pull a fresh uid from the persistent monotonic
	//    counter, then Add. The counter is a per-cluster row updated
	//    via one atomic UPDATE (InnoDB row lock), so allocations never
	//    regress even after LDAP entries are deleted and never collide
	//    with concurrent allocators. Constraint violations (from
	//    out-of-band writers or a mis-seeded counter) drive a bounded
	//    retry that pulls the next value.
	cfg := o.c.Config()
	for attempt := 1; attempt <= maxAllocRetries; attempt++ {
		uid, allocErr := o.uids.Allocate(ctx, cfg.CustosClusterID)
		if allocErr != nil {
			o.dlq(ctx, cu, "allocate_uid", allocErr)
			return allocErr
		}

		acct := buildPosixAccount(cu, user, uid, cfg)
		writtenDN, addErr := o.c.AddPosixAccount(acct)
		if addErr == nil {
			if cacheErr := o.storeUIDIdentity(ctx, cu.UserID, uid); cacheErr != nil {
				log.Error("ldap provisioner: cache write failed after LDAP add; partial state, self-heals on next event",
					"uid", uid, "err", cacheErr)
			}
			if err := o.ensurePrimaryGroup(ctx, cu, uid, log); err != nil {
				o.dlq(ctx, cu, "ensure_primary_group", err)
				return err
			}
			o.audit(ctx, cu, "LDAPAccountCreated",
				fmt.Sprintf("dn=%s username=%s uid=%d gid=%d", writtenDN, cu.LocalUsername, uid, uid))
			log.Info("ldap provisioner: posixAccount created", "dn", writtenDN, "uid", uid)
			return nil
		}
		if client.IsAlreadyExists(addErr) {
			// Another Custos instance created the entry between our
			// Find and Add. Adopt what's there rather than DLQ.
			log.Info("ldap provisioner: entry created concurrently by another instance, adopting")
			adopted, adoptErr := o.tryAdoptExisting(ctx, cu, user)
			if adoptErr != nil {
				return adoptErr
			}
			if adopted {
				return nil
			}
			log.Warn("ldap provisioner: EntryAlreadyExists but Find empty, retrying",
				"attempt", attempt)
			continue
		}
		if !client.IsConstraintViolation(addErr) {
			o.dlq(ctx, cu, "add_posix_account", addErr)
			return addErr
		}
		// uidNumber uniqueness violation from an out-of-band writer
		// (some entry in LDAP already has this number, e.g. a system
		// user provisioned outside Custos). Pull the next value from
		// the counter and try again.
		log.Warn("ldap provisioner: uidNumber constraint violation, allocating next",
			"attempt", attempt, "attempted_uid", uid)
	}

	err = fmt.Errorf("failed to write posixAccount after %d retries", maxAllocRetries)
	o.dlq(ctx, cu, "allocate_uid_retries_exhausted", err)
	return err
}

// tryAdoptExisting looks up a posixAccount by LocalUsername and, if
// present, caches its uidNumber and syncs mutable attributes via
// ensureEntry. Returns (adopted, err):
//
//   - (true, nil)  — an existing entry was adopted and fully processed
//     (audit fired, group ensured); caller returns nil.
//   - (false, nil) — no matching entry; caller falls through to
//     fresh allocation.
//   - (false, err) — an error occurred; DLQ has already been written.
//
// Used both by the initial LDAP-lookup path and by the retry loop when
// a concurrent instance's Add races ahead of ours.
func (o *Orchestrator) tryAdoptExisting(ctx context.Context, cu *models.ComputeClusterUser, user *models.User) (bool, error) {
	dn, attrs, err := o.c.FindPosixAccount(cu.LocalUsername)
	if err != nil {
		o.dlq(ctx, cu, "find_posix_account", err)
		return false, err
	}
	if dn == "" {
		return false, nil
	}
	uid, ok := parseUIDFromAttrs(attrs)
	if !ok {
		err := fmt.Errorf("existing entry %s has no parseable uidNumber", dn)
		o.dlq(ctx, cu, "parse_existing_uid", err)
		return false, err
	}
	if cacheErr := o.storeUIDIdentity(ctx, cu.UserID, uid); cacheErr != nil {
		// Adopted an existing LDAP entry but failed to cache. Next
		// event for this user will re-adopt — non-fatal, but not
		// silent. Escalated from Warn.
		slog.Error("ldap provisioner: cache write failed after adoption; will re-adopt on next event",
			"user_id", cu.UserID, "uid", uid, "err", cacheErr)
	}
	slog.Info("ldap provisioner: adopted existing entry",
		"dn", dn, "uid", uid, "user_id", cu.UserID)
	if err := o.ensureEntry(ctx, cu, user, uid); err != nil {
		return false, err
	}
	return true, nil
}

// ensureEntry writes the posixAccount with a known uidNumber (from cache
// or from a pre-existing LDAP entry) and ensures the matching posixGroup
// exists when configured.
func (o *Orchestrator) ensureEntry(ctx context.Context, cu *models.ComputeClusterUser, user *models.User, uid int64) error {
	log := slog.With(
		"correlation_id", cu.ID,
		"custos_user_id", cu.UserID,
		"local_username", cu.LocalUsername,
		"uid", uid,
	)
	acct := buildPosixAccount(cu, user, uid, o.c.Config())

	dn, _, err := o.c.FindPosixAccount(cu.LocalUsername)
	if err != nil {
		o.dlq(ctx, cu, "find_posix_account", err)
		return err
	}

	if dn == "" {
		writtenDN, err := o.c.AddPosixAccount(acct)
		if err != nil {
			if client.IsAlreadyExists(err) {
				// Race: the entry was created between our Find and
				// our Add (e.g., another Custos instance provisioning
				// the same user concurrently). Adopt the existing
				// entry rather than emit a spurious failure.
				log.Info("ldap provisioner: entry created concurrently during ensureEntry, adopting")
				adopted, adoptErr := o.tryAdoptExisting(ctx, cu, user)
				if adoptErr != nil {
					return adoptErr
				}
				if adopted {
					return nil
				}
				// EntryAlreadyExists but re-Find is empty — the racing
				// entry may have been deleted between the Add response
				// and our re-Find. Surface a real error.
				race := fmt.Errorf("ensureEntry: EntryAlreadyExists but re-Find empty for %q", cu.LocalUsername)
				o.dlq(ctx, cu, "ensure_entry_race", race)
				return race
			}
			o.dlq(ctx, cu, "add_posix_account", err)
			return err
		}
		if err := o.ensurePrimaryGroup(ctx, cu, uid, log); err != nil {
			o.dlq(ctx, cu, "ensure_primary_group", err)
			return err
		}
		o.audit(ctx, cu, "LDAPAccountCreated",
			fmt.Sprintf("dn=%s username=%s uid=%d gid=%d", writtenDN, cu.LocalUsername, uid, uid))
		return nil
	}
	if err := o.c.ModifyPosixAccount(dn, acct); err != nil {
		o.dlq(ctx, cu, "modify_posix_account", err)
		return err
	}
	if err := o.ensurePrimaryGroup(ctx, cu, uid, log); err != nil {
		o.dlq(ctx, cu, "ensure_primary_group", err)
		return err
	}
	o.audit(ctx, cu, "LDAPAccountUpdated",
		fmt.Sprintf("dn=%s username=%s", dn, cu.LocalUsername))
	return nil
}

// ensurePrimaryGroup adds a posixGroup entry with cn=<local_username>
// and gidNumber=<uid> under GroupBaseDN, if that container is
// configured. Concurrent adds are tolerated: an "entry already exists"
// response is treated as success.
//
// When GroupBaseDN is empty the connector assumes the site uses
// automatic-private-groups on the client (RHEL / Fedora default) and
// no LDAP-side group entry is needed.
func (o *Orchestrator) ensurePrimaryGroup(ctx context.Context, cu *models.ComputeClusterUser, uid int64, log *slog.Logger) error {
	if o.c.Config().GroupBaseDN == "" {
		return nil
	}
	existing, err := o.c.FindPosixGroup(cu.LocalUsername)
	if err != nil {
		return fmt.Errorf("find posixGroup %s: %w", cu.LocalUsername, err)
	}
	if existing != "" {
		log.Info("ldap provisioner: primary posixGroup already present", "dn", existing)
		return nil
	}
	dn, err := o.c.AddPosixGroup(cu.LocalUsername, uid)
	if err != nil {
		if client.IsAlreadyExists(err) {
			// A concurrent writer created it between our Find and Add.
			log.Info("ldap provisioner: posixGroup created concurrently")
			return nil
		}
		return fmt.Errorf("add posixGroup %s: %w", cu.LocalUsername, err)
	}
	log.Info("ldap provisioner: primary posixGroup created", "dn", dn, "gid", uid)
	o.audit(ctx, cu, "LDAPGroupCreated",
		fmt.Sprintf("dn=%s cn=%s gid=%d", dn, cu.LocalUsername, uid))
	return nil
}

func buildPosixAccount(cu *models.ComputeClusterUser, user *models.User, uid int64, cfg client.Config) client.PosixAccount {
	return client.PosixAccount{
		UID:           cu.LocalUsername,
		UIDNumber:     uid,
		GIDNumber:     uid, // one-group-per-user, mirrors COmanage
		GivenName:     user.FirstName,
		Surname:       user.LastName,
		Mail:          user.Email,
		HomeDirectory: cfg.HomedirPrefix + cu.LocalUsername,
		LoginShell:    cfg.DefaultShell,
	}
}

// findCachedUID returns the uidNumber previously assigned to this Custos
// user via this connector, or 0 if none is on file.
func (o *Orchestrator) findCachedUID(ctx context.Context, userID string) (int64, error) {
	identities, err := o.core.ListUserIdentitiesForUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	source := o.identitySource()
	for _, ui := range identities {
		if ui.Source != source {
			continue
		}
		n, err := strconv.ParseInt(ui.ExternalID, 10, 64)
		if err != nil {
			slog.Warn("ldap provisioner: user_identity(source=ldap) external_id not an integer",
				"user_id", userID, "external_id", ui.ExternalID, "err", err)
			continue
		}
		if n > 0 {
			return n, nil
		}
	}
	return 0, nil
}

func (o *Orchestrator) storeUIDIdentity(ctx context.Context, userID string, uid int64) error {
	_, err := o.core.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     userID,
		Source:     o.identitySource(),
		ExternalID: strconv.FormatInt(uid, 10),
	})
	return err
}

func parseUIDFromAttrs(attrs map[string][]string) (int64, bool) {
	values := attrs["uidNumber"]
	if len(values) == 0 {
		return 0, false
	}
	n, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
