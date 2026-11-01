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
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	userAgreementAuditCreated   = "USER_AGREEMENT_CREATED"
	userAgreementAuditUpdated   = "USER_AGREEMENT_UPDATED"
	userAgreementAuditDeleted   = "USER_AGREEMENT_DELETED"
	userAgreementAuditPublished = "USER_AGREEMENT_PUBLISHED"
	userAgreementAuditRemoved   = "USER_AGREEMENT_REMOVED"
)

// UserAgreementInput is the editable content of an agreement draft.
type UserAgreementInput struct {
	Title        string
	VersionLabel string
	Summary      string
	BodyMD       string
	ChangeNote   string
}

func (in UserAgreementInput) validate() error {
	if in.Title == "" || in.VersionLabel == "" || in.BodyMD == "" {
		return fmt.Errorf("%w: title, version_label and body_md are required", ErrInvalidInput)
	}
	return nil
}

func bodyHash(md string) string {
	sum := sha256.Sum256([]byte(md))
	return hex.EncodeToString(sum[:])
}

func agreementWriteErr(op string, err error) error {
	if strings.Contains(err.Error(), "uq_user_agreements_version") {
		return fmt.Errorf("%w: this version already exists", ErrAlreadyExists)
	}
	return fmt.Errorf("%s: %w", op, err)
}

func (s *Service) ListUserAgreements(ctx context.Context) ([]models.UserAgreement, error) {
	return s.agreements.List(ctx)
}

func (s *Service) GetUserAgreement(ctx context.Context, id string) (*models.UserAgreement, error) {
	a, err := s.agreements.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup user agreement: %w", err)
	}
	if a == nil {
		return nil, fmt.Errorf("%w: user agreement %q does not exist", ErrNotFound, id)
	}
	return a, nil
}

// CreateUserAgreementDraft requires actorID to hold agreements:manage.
func (s *Service) CreateUserAgreementDraft(ctx context.Context, in UserAgreementInput, actorID string) (*models.UserAgreement, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	a := &models.UserAgreement{
		ID:           newID(),
		Title:        in.Title,
		VersionLabel: in.VersionLabel,
		Summary:      in.Summary,
		BodyMD:       in.BodyMD,
		BodyHash:     bodyHash(in.BodyMD),
		Status:       models.UserAgreementDraft,
		ChangeNote:   in.ChangeNote,
		CreatedBy:    actorID,
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.AgreementsManage); err != nil {
			return err
		}
		if err := s.agreements.Create(ctx, tx, a); err != nil {
			return agreementWriteErr("insert user agreement", err)
		}
		return s.writeAuditTx(ctx, tx, "user_agreement", userAgreementAuditCreated, a.ID, map[string]any{
			"actor_id":      actorID,
			"version_label": a.VersionLabel,
		})
	}); err != nil {
		return nil, err
	}
	return a, nil
}

// UpdateUserAgreementDraft rewrites a draft. Only drafts can be edited.
func (s *Service) UpdateUserAgreementDraft(ctx context.Context, id string, in UserAgreementInput, actorID string) (*models.UserAgreement, error) {
	var a *models.UserAgreement
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.AgreementsManage); err != nil {
			return err
		}
		if a, err = s.draftTx(ctx, tx, id); err != nil {
			return err
		}
		if err := in.validate(); err != nil {
			return err
		}
		a.Title, a.VersionLabel = in.Title, in.VersionLabel
		a.Summary, a.BodyMD, a.BodyHash, a.ChangeNote = in.Summary, in.BodyMD, bodyHash(in.BodyMD), in.ChangeNote
		if err := s.agreements.Update(ctx, tx, a); err != nil {
			return agreementWriteErr("update user agreement", err)
		}
		return s.writeAuditTx(ctx, tx, "user_agreement", userAgreementAuditUpdated, id, map[string]any{"actor_id": actorID})
	}); err != nil {
		return nil, err
	}
	return a, nil
}

// DeleteUserAgreementDraft removes a draft. Published versions are kept.
func (s *Service) DeleteUserAgreementDraft(ctx context.Context, id, actorID string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.AgreementsManage); err != nil {
			return err
		}
		a, err := s.draftTx(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.agreements.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("delete user agreement: %w", err)
		}
		return s.writeAuditTx(ctx, tx, "user_agreement", userAgreementAuditDeleted, id, map[string]any{
			"actor_id":      actorID,
			"version_label": a.VersionLabel,
		})
	})
}

// PublishUserAgreement makes a draft the current version.
func (s *Service) PublishUserAgreement(ctx context.Context, id, actorID string) (*models.UserAgreement, error) {
	var a *models.UserAgreement
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.AgreementsManage); err != nil {
			return err
		}
		if a, err = s.draftTx(ctx, tx, id); err != nil {
			return err
		}
		now := nowUTC()
		a.Status, a.PublishedAt, a.PublishedBy = models.UserAgreementPublished, &now, &actorID
		if err := s.agreements.Update(ctx, tx, a); err != nil {
			return fmt.Errorf("publish user agreement: %w", err)
		}
		if err := s.writeAuditTx(ctx, tx, "user_agreement", userAgreementAuditPublished, id, map[string]any{
			"actor_id":      actorID,
			"version_label": a.VersionLabel,
		}); err != nil {
			return err
		}
		return s.eventBus.Publish(ctx, tx, events.UserAgreementPublishEvent, a)
	}); err != nil {
		return nil, err
	}
	return a, nil
}

// RemoveUserAgreement marks a published version REMOVED, the row is kept.
func (s *Service) RemoveUserAgreement(ctx context.Context, id, actorID string) (*models.UserAgreement, error) {
	var a *models.UserAgreement
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.AgreementsManage); err != nil {
			return err
		}
		if a, err = s.agreements.FindByIDTx(ctx, tx, id); err != nil {
			return fmt.Errorf("lookup user agreement: %w", err)
		}
		if a == nil {
			return fmt.Errorf("%w: user agreement %q does not exist", ErrNotFound, id)
		}
		if a.Status != models.UserAgreementPublished {
			return fmt.Errorf("%w: only a published agreement can be removed", ErrInvalidInput)
		}
		now := nowUTC()
		a.Status, a.RemovedAt, a.RemovedBy = models.UserAgreementRemoved, &now, &actorID
		if err := s.agreements.Update(ctx, tx, a); err != nil {
			return fmt.Errorf("remove user agreement: %w", err)
		}
		if err := s.writeAuditTx(ctx, tx, "user_agreement", userAgreementAuditRemoved, id, map[string]any{"actor_id": actorID}); err != nil {
			return err
		}
		return s.eventBus.Publish(ctx, tx, events.UserAgreementRemoveEvent, a)
	}); err != nil {
		return nil, err
	}
	return a, nil
}

// Users are pending only on the current version.
func (s *Service) isCurrentUserAgreement(ctx context.Context, id string) (bool, error) {
	current, err := s.agreements.FindCurrent(ctx)
	if err != nil {
		return false, fmt.Errorf("lookup current agreement: %w", err)
	}
	return current != nil && current.ID == id, nil
}

func (s *Service) UserAgreementCounts(ctx context.Context, id string) (accepted, pending int, err error) {
	if accepted, err = s.agreements.CountAcceptances(ctx, id); err != nil {
		return 0, 0, fmt.Errorf("count acceptances: %w", err)
	}
	isCurrent, err := s.isCurrentUserAgreement(ctx, id)
	if err != nil || !isCurrent {
		return accepted, 0, err
	}
	if pending, err = s.agreements.CountPendingUsers(ctx, id); err != nil {
		return 0, 0, fmt.Errorf("count pending users: %w", err)
	}
	return accepted, pending, nil
}

func (s *Service) ListUserAgreementAcceptances(ctx context.Context, id string, limit, offset int) ([]models.UserAgreementAcceptance, error) {
	return s.acceptances.ListByAgreement(ctx, id, limit, offset)
}

func (s *Service) ListUserAgreementPendingUsers(ctx context.Context, id string, limit, offset int) ([]models.User, error) {
	isCurrent, err := s.isCurrentUserAgreement(ctx, id)
	if err != nil || !isCurrent {
		return []models.User{}, err
	}
	return s.acceptances.ListPendingUsers(ctx, id, limit, offset)
}

func (s *Service) draftTx(ctx context.Context, tx *sql.Tx, id string) (*models.UserAgreement, error) {
	a, err := s.agreements.FindByIDTx(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup user agreement: %w", err)
	}
	if a == nil {
		return nil, fmt.Errorf("%w: user agreement %q does not exist", ErrNotFound, id)
	}
	if a.Status != models.UserAgreementDraft {
		return nil, fmt.Errorf("%w: user agreement %q is not a draft", ErrInvalidInput, id)
	}
	return a, nil
}
