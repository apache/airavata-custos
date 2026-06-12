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

	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	_ "github.com/apache/airavata-custos/pkg/service"
)

var ErrNotFoundShim = client.ErrNotFound

const comanageIdentitySource = "comanage"

func (o *Orchestrator) findStoredPersonID(ctx context.Context, userID string) (string, error) {
	ctx, span := tracing.Start(ctx, "comanage.find_stored_person_id")
	defer span.End()
	idents, err := o.core.ListUserIdentitiesForUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("list user identities: %w", err)
	}
	for _, id := range idents {
		if id.Source == comanageIdentitySource && id.ExternalID != "" {
			return id.ExternalID, nil
		}
	}
	return "", nil
}

// storePersonID writes the COmanage CoPerson identifier into user_identities.
// No-op if a row already exists.
func (o *Orchestrator) storePersonID(ctx context.Context, userID, personID string) error {
	ctx, span := tracing.Start(ctx, "comanage.store_person_id")
	defer span.End()
	if existing, err := o.findStoredPersonID(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	} else if existing != "" {
		return nil
	}
	_, err := o.core.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     userID,
		Source:     comanageIdentitySource,
		ExternalID: personID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// findByEmailExact searches by email and returns the first match in the
// configured CO, or 0 if none. COmanage's search.mail is a LIKE match: callers
// trading off precision for one round-trip should be aware that very similar
// emails could collide.
func (o *Orchestrator) findByEmailExact(email string) (int, error) {
	if email == "" {
		return 0, nil
	}
	candidates, err := o.c.FindCoPersonByEmail(email)
	if err != nil {
		if errors.Is(err, ErrNotFoundShim) {
			return 0, nil
		}
		return 0, err
	}
	for _, p := range candidates {
		if p.CoId == o.c.Config().COID {
			return p.Id, nil
		}
	}
	return 0, nil
}
