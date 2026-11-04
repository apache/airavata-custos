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

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
)

// deactivatePOSIXAccountImpl removes the posixAccount, and the user's primary
// posixGroup when GroupBaseDN is set, from the directory. A group with
// another gidNumber was not created by Custos, so it is left alone. The row
// stays, marked INACTIVE, so its numbers are never given to someone else.
func (o *Orchestrator) deactivatePOSIXAccountImpl(ctx context.Context, cu *models.ComputeClusterUser) error {
	log := slog.With("correlation_id", cu.ID, "custos_user_id", cu.UserID, "local_username", cu.LocalUsername)

	acct, err := o.accounts.Get(ctx, cu.ID)
	if err != nil {
		o.auditFailure(ctx, cu, "get_posix_account", err)
		return err
	}
	if acct == nil {
		return nil
	}
	if err := o.c.DeleteEntry(acct.EntryDN); err != nil {
		o.auditFailure(ctx, cu, "delete_posix_account", err)
		return err
	}
	if o.c.Config().GroupBaseDN != "" {
		groupDN, gid, err := o.c.FindPosixGroup(acct.Username)
		if groupDN != "" && gid != acct.GIDNumber {
			log.WarnContext(ctx, "ldap provisioner: posixGroup has another gidNumber; leaving it", "dn", groupDN, "gid", gid)
			groupDN = ""
		}
		if err == nil && groupDN != "" {
			err = o.c.DeleteEntry(groupDN)
		}
		if err != nil {
			o.auditFailure(ctx, cu, "delete_posix_group", err)
			return err
		}
		if groupDN != "" {
			o.audit(ctx, cu, "LDAPGroupDeleted", fmt.Sprintf("dn=%s cn=%s gid=%d", groupDN, acct.Username, acct.GIDNumber))
		}
	}
	if err := o.accounts.SetStatus(ctx, cu.ID, store.StatusInactive); err != nil {
		o.auditFailure(ctx, cu, "deactivate_posix_account", err)
		return err
	}
	o.audit(ctx, cu, "LDAPAccountDeactivated", fmt.Sprintf("dn=%s username=%s uid=%d", acct.EntryDN, acct.Username, acct.UIDNumber))
	log.InfoContext(ctx, "ldap provisioner: posixAccount deactivated", "dn", acct.EntryDN, "uid", acct.UIDNumber)
	return nil
}
