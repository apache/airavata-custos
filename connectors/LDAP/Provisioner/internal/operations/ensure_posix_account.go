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
	"errors"
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"strconv"

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// maxPosixUsernameLen is the traditional POSIX login-name cap.
const maxPosixUsernameLen = 32

// validPosixUsername also rejects every DN metacharacter, so the username
// can go into uid=<value>,<BaseDN> without escaping.
var validPosixUsername = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

// maxAllocRetries bounds the renumbering when the directory rejects a
// uidNumber that an entry created outside Custos already holds.
const maxAllocRetries = 3

// errConflict marks a state only an operator can resolve, so the bus fails
// the delivery at once instead of retrying it.
var errConflict = fmt.Errorf("%w: conflict", events.ErrPermanent)

// ensurePOSIXAccountImpl provisions the posixAccount (and, when GroupBaseDN
// is set, the matching posixGroup) for a cluster account. The uid and gid
// come from ldap_posix_accounts; gidNumber = uidNumber.
func (o *Orchestrator) ensurePOSIXAccountImpl(ctx context.Context, cu *models.ComputeClusterUser) error {
	log := slog.With(
		"correlation_id", cu.ID,
		"custos_user_id", cu.UserID,
		"local_username", cu.LocalUsername,
		"base_dn", o.c.Config().BaseDN,
	)

	if cu.LocalUsername == "" {
		err := fmt.Errorf("%w: compute_cluster_user %s has empty local_username", events.ErrPermanent, cu.ID)
		o.auditFailure(ctx, cu, "validate_local_username", err)
		return err
	}
	if len(cu.LocalUsername) > maxPosixUsernameLen || !validPosixUsername.MatchString(cu.LocalUsername) {
		err := fmt.Errorf("%w: compute_cluster_user %s has invalid local_username %q (must be POSIX-safe: lowercase, [a-z_][a-z0-9_-]*, <=%d chars)",
			events.ErrPermanent, cu.ID, cu.LocalUsername, maxPosixUsernameLen)
		o.auditFailure(ctx, cu, "validate_local_username", err)
		return err
	}

	user, err := o.core.GetUser(ctx, cu.UserID)
	if err != nil {
		o.auditFailure(ctx, cu, "get_custos_user", err)
		return fmt.Errorf("get custos user %s: %w", cu.UserID, err)
	}
	if user == nil {
		err := fmt.Errorf("custos user %s not found", cu.UserID)
		o.auditFailure(ctx, cu, "get_custos_user", err)
		return err
	}

	acct, err := o.accounts.Get(ctx, cu.ID)
	if err != nil {
		o.auditFailure(ctx, cu, "get_posix_account", err)
		return err
	}
	if acct == nil {
		if acct, err = o.createAccount(ctx, cu, log); err != nil {
			return err
		}
	}
	// An approval retried after the account was deleted must not bring the
	// entry back.
	if acct.Status == store.StatusInactive {
		log.InfoContext(ctx, "ldap provisioner: account was deactivated; skipping")
		return nil
	}
	log = log.With("uid", acct.UIDNumber)

	if err := o.ensureEntry(ctx, cu, user, acct, log); err != nil {
		return err
	}
	if err := o.ensurePrimaryGroup(ctx, cu, acct, log); err != nil {
		o.auditFailure(ctx, cu, "ensure_primary_group", err)
		return err
	}
	if err := o.core.MarkComputeClusterUserProvisioned(ctx, cu.ID); err != nil {
		log.WarnContext(ctx, "ldap provisioner: failed to mark cluster user provisioned", "err", err)
	}
	return nil
}

// createAccount records the attributes for a cluster account seen for the
// first time. A Custos user re-added after removal gets their old row back,
// so they own their files again. Any other holder of the username, here or
// in the directory, is a conflict: adopting an existing account has to be
// an explicit import.
func (o *Orchestrator) createAccount(ctx context.Context, cu *models.ComputeClusterUser, log *slog.Logger) (*store.Account, error) {
	held, err := o.accounts.GetByUsername(ctx, cu.LocalUsername)
	if err != nil {
		o.auditFailure(ctx, cu, "get_posix_account", err)
		return nil, err
	}
	if held != nil && held.Status == store.StatusInactive && held.UserID == cu.UserID {
		acct, err := o.accounts.Reclaim(ctx, held.ComputeClusterUserID, cu.ID)
		if err != nil {
			o.auditFailure(ctx, cu, "reclaim_posix_account", err)
			return nil, err
		}
		if acct != nil {
			log.InfoContext(ctx, "ldap provisioner: reclaimed deactivated account", "uid", acct.UIDNumber, "previous_correlation_id", held.ComputeClusterUserID)
			return acct, nil
		}
	}
	if held != nil {
		err := fmt.Errorf("%w: username %q is held by compute_cluster_user %s", errConflict, cu.LocalUsername, held.ComputeClusterUserID)
		o.auditFailure(ctx, cu, "claim_username", err)
		return nil, err
	}
	dn, _, err := o.c.FindPosixAccount(cu.LocalUsername)
	if err != nil {
		o.auditFailure(ctx, cu, "find_posix_account", err)
		return nil, err
	}
	if dn != "" {
		err := fmt.Errorf("%w: %s already exists and was not created by Custos", errConflict, dn)
		o.auditFailure(ctx, cu, "claim_username", err)
		return nil, err
	}

	cfg := o.c.Config()
	acct, err := o.accounts.Create(ctx, store.Account{
		ComputeClusterUserID: cu.ID,
		UserID:               cu.UserID,
		Username:             cu.LocalUsername,
		HomeDirectory:        path.Join(cfg.HomedirPrefix, cu.LocalUsername),
		LoginShell:           cfg.DefaultShell,
		EntryDN:              fmt.Sprintf("uid=%s,%s", cu.LocalUsername, cfg.BaseDN),
	}, o.floor)
	if errors.Is(err, store.ErrUsernameTaken) {
		err = fmt.Errorf("%w: username %q was claimed by another cluster account", errConflict, cu.LocalUsername)
	}
	if err != nil {
		o.auditFailure(ctx, cu, "create_posix_account", err)
		return nil, err
	}
	log.InfoContext(ctx, "ldap provisioner: uidNumber assigned", "uid", acct.UIDNumber)
	return acct, nil
}

// ensureEntry adds the posixAccount, or updates it when it exists and
// carries the stored uidNumber.
func (o *Orchestrator) ensureEntry(ctx context.Context, cu *models.ComputeClusterUser, user *models.User, acct *store.Account, log *slog.Logger) error {
	dn, attrs, err := o.c.FindPosixAccount(acct.Username)
	if err != nil {
		o.auditFailure(ctx, cu, "find_posix_account", err)
		return err
	}
	if dn != "" {
		if got := attrs["uidNumber"]; len(got) != 1 || got[0] != strconv.FormatInt(acct.UIDNumber, 10) {
			err := fmt.Errorf("%w: %s has uidNumber %v, Custos assigned %d", errConflict, dn, got, acct.UIDNumber)
			o.auditFailure(ctx, cu, "verify_posix_account", err)
			return err
		}
		if err := o.c.ModifyPosixAccount(dn, buildPosixAccount(acct, user)); err != nil {
			o.auditFailure(ctx, cu, "modify_posix_account", err)
			return err
		}
		o.audit(ctx, cu, "LDAPAccountUpdated", fmt.Sprintf("dn=%s username=%s", dn, acct.Username))
		log.InfoContext(ctx, "ldap provisioner: posixAccount updated", "dn", dn)
		return nil
	}

	for attempt := 1; ; attempt++ {
		writtenDN, err := o.c.AddPosixAccount(buildPosixAccount(acct, user))
		if err == nil {
			o.audit(ctx, cu, "LDAPAccountCreated",
				fmt.Sprintf("dn=%s username=%s uid=%d gid=%d", writtenDN, acct.Username, acct.UIDNumber, acct.GIDNumber))
			log.InfoContext(ctx, "ldap provisioner: posixAccount created", "dn", writtenDN)
			return nil
		}
		if !client.IsConstraintViolation(err) || attempt > maxAllocRetries {
			o.auditFailure(ctx, cu, "add_posix_account", err)
			return err
		}
		log.WarnContext(ctx, "ldap provisioner: uidNumber constraint violation, renumbering", "attempt", attempt, "attempted_uid", acct.UIDNumber)
		renumbered, renumberErr := o.accounts.Renumber(ctx, cu.ID, o.floor)
		if renumberErr != nil {
			o.auditFailure(ctx, cu, "renumber_posix_account", renumberErr)
			return renumberErr
		}
		*acct = *renumbered
	}
}

// ensurePrimaryGroup adds the cn=<username> posixGroup under GroupBaseDN, if
// configured. Without it the hosts are expected to use automatic private
// groups. A concurrent add is treated as success.
func (o *Orchestrator) ensurePrimaryGroup(ctx context.Context, cu *models.ComputeClusterUser, acct *store.Account, log *slog.Logger) error {
	if o.c.Config().GroupBaseDN == "" {
		return nil
	}
	existing, err := o.c.FindPosixGroup(acct.Username)
	if err != nil {
		return fmt.Errorf("find posixGroup %s: %w", acct.Username, err)
	}
	if existing != "" {
		log.InfoContext(ctx, "ldap provisioner: primary posixGroup already present", "dn", existing)
		return nil
	}
	dn, err := o.c.AddPosixGroup(acct.Username, acct.GIDNumber)
	if err != nil {
		if client.IsAlreadyExists(err) {
			log.InfoContext(ctx, "ldap provisioner: posixGroup created concurrently")
			return nil
		}
		return fmt.Errorf("add posixGroup %s: %w", acct.Username, err)
	}
	log.InfoContext(ctx, "ldap provisioner: primary posixGroup created", "dn", dn, "gid", acct.GIDNumber)
	o.audit(ctx, cu, "LDAPGroupCreated",
		fmt.Sprintf("dn=%s cn=%s gid=%d", dn, acct.Username, acct.GIDNumber))
	return nil
}

func buildPosixAccount(acct *store.Account, user *models.User) client.PosixAccount {
	// sn is required by the person schema, and a user may have no last name.
	given, surname := user.FirstName, user.LastName
	if surname == "" {
		given, surname = "", user.FirstName
	}
	if surname == "" {
		surname = acct.Username
	}
	return client.PosixAccount{
		UID:           acct.Username,
		UIDNumber:     acct.UIDNumber,
		GIDNumber:     acct.GIDNumber,
		GivenName:     given,
		Surname:       surname,
		Mail:          user.Email,
		HomeDirectory: acct.HomeDirectory,
		LoginShell:    acct.LoginShell,
	}
}
