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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// RemovePOSIXAccount removes what provisioning created for a cluster account:
// the account and login name on the person, the admin group membership, and
// the account's primary group. The person stays in the registry. It is safe
// to run again, anything already removed is skipped.
func (o *Orchestrator) RemovePOSIXAccount(ctx context.Context, cu *models.ComputeClusterUser) (err error) {
	ctx, span := tracing.Start(ctx, "comanage.remove_posix_account")
	defer span.End()
	span.SetAttributes(attribute.String("comanage.cluster_user_id", cu.ID))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	user, err := o.core.GetUser(ctx, cu.UserID)
	if err != nil {
		return fmt.Errorf("get Custos user: %w", err)
	}
	personID, composite, err := o.findCoPerson(ctx, user)
	if err != nil {
		return err
	}
	if personID != "" {
		if err := o.removeAccountFromPerson(ctx, cu, personID, composite); err != nil {
			return err
		}
	}

	// The account names this group as its primary group, so it goes last.
	groupID, err := o.c.FindCoGroupByName(cu.LocalUsername)
	if err != nil {
		return fmt.Errorf("find primary group: %w", err)
	}
	if groupID != 0 {
		if err := o.removeCoGroup(groupID); err != nil {
			return err
		}
	}

	o.audit(ctx, cu, "ComanageClusterAccountRemoved", fmt.Sprintf("comanage_id=%s username=%s", personID, cu.LocalUsername))
	return nil
}

// removeAccountFromPerson takes the cluster account, its login name and the
// admin group membership off the person. The person record itself stays.
func (o *Orchestrator) removeAccountFromPerson(ctx context.Context, cu *models.ComputeClusterUser, personID string, composite []byte) error {
	body, changed, err := removeLogin(composite, cu.LocalUsername)
	if err != nil {
		return err
	}
	if changed {
		if err := o.updatePerson(ctx, personID, body); err != nil {
			return fmt.Errorf("remove account from person: %w", err)
		}
	}

	// Left in place, the membership would give sudo to a later account of the
	// same person.
	if cu.AccessLevel != models.ClusterAccessAdmin {
		return nil
	}
	adminGroupID, err := o.c.FindCoGroupByName(o.c.Config().ClusterAdminGroup)
	if err != nil {
		return fmt.Errorf("find cluster admin group: %w", err)
	}
	coPersonID, err := extractCoPersonID(composite)
	if err != nil {
		return err
	}
	if adminGroupID == 0 || coPersonID == 0 {
		return nil
	}
	memberID, err := o.c.FindCoGroupMember(adminGroupID, coPersonID)
	if err != nil {
		return fmt.Errorf("find cluster admin membership: %w", err)
	}
	if memberID == 0 {
		return nil
	}
	if err := ignoreNotFound(o.c.DeleteCoGroupMember(memberID)); err != nil {
		return fmt.Errorf("delete cluster admin membership: %w", err)
	}
	return nil
}

// ignoreNotFound treats a delete of something already deleted as done.
func ignoreNotFound(err error) error {
	if errors.Is(err, client.ErrNotFound) {
		return nil
	}
	return err
}

// removeCoGroup unbinds a group from the UnixCluster, then deletes it.
func (o *Orchestrator) removeCoGroup(groupID int) error {
	bindingID, err := o.c.FindUnixClusterGroup(groupID)
	if err != nil {
		return fmt.Errorf("find cluster group binding: %w", err)
	}
	if bindingID != 0 {
		if err := ignoreNotFound(o.c.DeleteUnixClusterGroup(bindingID)); err != nil {
			return fmt.Errorf("delete cluster group binding: %w", err)
		}
	}
	if err := ignoreNotFound(o.c.DeleteCoGroup(groupID)); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	return nil
}
