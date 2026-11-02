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

package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

const userAgreementAuditAccepted = "USER_AGREEMENT_ACCEPTED"

// PendingUserAgreements returns the current agreement if the user has not accepted it.
func (s *Service) PendingUserAgreements(ctx context.Context, userID string) ([]models.UserAgreement, error) {
	return s.agreements.ListPendingForUser(ctx, userID)
}

// AcceptUserAgreement records the acceptance once. A repeat returns the first record.
func (s *Service) AcceptUserAgreement(ctx context.Context, userID, agreementID string) (*models.UserAgreementAcceptance, error) {
	a, err := s.GetUserAgreement(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	if a.Status != models.UserAgreementPublished {
		return nil, fmt.Errorf("%w: only a published agreement can be accepted", ErrInvalidInput)
	}
	var acc *models.UserAgreementAcceptance
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		acc, err = s.acceptTx(ctx, tx, userID, a)
		return err
	}); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *Service) acceptTx(ctx context.Context, tx *sql.Tx, userID string, a *models.UserAgreement) (*models.UserAgreementAcceptance, error) {
	acc := &models.UserAgreementAcceptance{
		ID:          newID(),
		UserID:      userID,
		AgreementID: a.ID,
		AcceptedAt:  nowUTC(),
		BodyHash:    a.BodyHash,
	}
	created, err := s.acceptances.Create(ctx, tx, acc)
	if err != nil {
		return nil, fmt.Errorf("insert acceptance: %w", err)
	}
	if !created {
		return s.acceptances.FindTx(ctx, tx, userID, a.ID)
	}
	if err := s.writeAuditTx(ctx, tx, "user", userAgreementAuditAccepted, userID, map[string]any{
		"agreement_id":  a.ID,
		"version_label": a.VersionLabel,
		"body_hash":     a.BodyHash,
	}); err != nil {
		return nil, err
	}
	return acc, s.eventBus.Publish(ctx, tx, events.UserAgreementAcceptanceCreateEvent, acc)
}
