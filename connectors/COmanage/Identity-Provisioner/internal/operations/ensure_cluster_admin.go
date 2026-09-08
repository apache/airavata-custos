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

	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// EnsureClusterAdminGroup puts an ADMIN cluster user in the group that carries
// sudo. The group is set up by an operator, so a missing one is an error here
// rather than something to create.
func (o *Orchestrator) EnsureClusterAdminGroup(ctx context.Context, cu *models.ComputeClusterUser) error {
	ctx, span := tracing.Start(ctx, "comanage.ensure_cluster_admin_group")
	defer span.End()
	span.SetAttributes(attribute.String("comanage.cluster_user_id", cu.ID))

	fail := func(step string, err error) error {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		o.dlq(ctx, cu, step, err)
		return err
	}

	name := o.c.Config().ClusterAdminGroup
	if name == "" {
		return fail("cluster_admin_group_unset", errors.New("cluster admin group is not configured"))
	}
	span.SetAttributes(attribute.String("comanage.cluster_admin_group", name))

	groupID, err := o.c.FindCoGroupByName(name)
	if err != nil {
		return fail("find_cluster_admin_group", err)
	}
	if groupID == 0 {
		return fail("cluster_admin_group_missing", fmt.Errorf("cluster admin group %q is not in the registry", name))
	}

	// Set by EnsurePOSIXAccount, which runs first.
	personID, err := o.findStoredPersonID(ctx, cu.UserID)
	if err != nil {
		return fail("cluster_admin_find_person", err)
	}
	if personID == "" {
		return fail("cluster_admin_person_missing", fmt.Errorf("no stored CoPerson for user %s", cu.UserID))
	}

	composite, err := o.getPersonComposite(ctx, personID)
	if err != nil {
		return fail("cluster_admin_get_composite", err)
	}
	coPersonID, err := extractCoPersonID(composite)
	if err != nil || coPersonID == 0 {
		return fail("cluster_admin_extract_coperson_id", fmt.Errorf("extract CoPerson.meta.id: %w", err))
	}

	// Joined as a plain member.
	// An owner could add anyone else to the group straight in the registry, going around Custos.
	if err := o.findOrCreateCoGroupMember(ctx, cu, groupID, coPersonID, false); err != nil {
		return err
	}
	span.SetAttributes(attribute.Int("comanage.co_group_id", groupID))
	o.audit(ctx, cu, "ComanageClusterAdminGroupJoined", fmt.Sprintf("group=%s co_group_id=%d comanage_id=%s", name, groupID, personID))
	return nil
}
