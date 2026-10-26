//go:build integration

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
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

func agreementInput(version string) UserAgreementInput {
	return UserAgreementInput{Title: "Acceptable Use", VersionLabel: version, BodyMD: "# Rules " + version}
}

func seedAgreementManager(t *testing.T, database *sqlx.DB) string {
	t.Helper()
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.AgreementsManage)
	return actor
}

func newClusterAccount(t *testing.T, svc *Service, userID, name string) {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	if _, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: userID, LocalUsername: name}); err != nil {
		t.Fatalf("create cluster account: %v", err)
	}
}

func seedPublishedAgreement(t *testing.T, svc *Service, actor, version string) *models.UserAgreement {
	t.Helper()
	a, err := svc.CreateUserAgreementDraft(ctx(), agreementInput(version), actor)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	if a, err = svc.PublishUserAgreement(ctx(), a.ID, actor); err != nil {
		t.Fatalf("publish: %v", err)
	}
	return a
}

func countRows(t *testing.T, database *sqlx.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := database.Get(&n, query, args...); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// Accepting the current agreement at first sign-in binds the sub, activates the
// user, renames the cluster account, and records the acceptance together.
func TestCompleteOnboarding_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	bus := newTestBus(database)
	svc := New(database, bus)
	bus.SubscribeUserOnboarded("test-subscriber", func(context.Context, models.User) error { return nil })
	aup := seedPublishedAgreement(t, svc, seedAgreementManager(t, database), "1.0")
	email := "jdoe-" + uuid.NewString()[:8] + "@example.edu"
	user := seedPendingUserWithEmail(t, database, email)
	newClusterAccount(t, svc, user, "custos-jdoe")
	claims := &identity.Claims{Sub: "sub-" + uuid.NewString(), Email: email, EmailVerified: true}

	if _, err := svc.CompleteOnboarding(ctx(), claims, OnboardingInput{Username: "jdoe", AcceptedIDs: []string{aup.ID}}); err != nil {
		t.Fatalf("CompleteOnboarding: %v", err)
	}

	if u, err := svc.GetUser(ctx(), user); err != nil || u.Status != models.UserActive {
		t.Errorf("status: got %+v, %v, want ACTIVE", u, err)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_identities WHERE user_id = $1 AND source = 'oidc' AND oidc_sub = $2", user, claims.Sub); n != 1 {
		t.Errorf("oidc identity with sub: got %d, want 1", n)
	}
	if accounts, err := svc.ListComputeClusterUsersByUser(ctx(), user); err != nil || len(accounts) != 1 || accounts[0].LocalUsername != "jdoe" {
		t.Errorf("cluster account: got %+v, %v, want one named jdoe", accounts, err)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_agreement_acceptances WHERE user_id = $1 AND agreement_id = $2", user, aup.ID); n != 1 {
		t.Errorf("acceptance: got %d, want 1", n)
	}
	for _, ev := range []string{"IDENTITY_LINKED", "USER_AGREEMENT_ACCEPTED"} {
		if got := countAuditEventsOfType(t, database, ev, user); got != 1 {
			t.Errorf("audit %s: got %d, want 1", ev, got)
		}
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM events WHERE event_type = 'user::onboarded'"); n != 1 {
		t.Errorf("user::onboarded events: got %d, want 1", n)
	}
}

// First sign-in without accepting the current agreement is refused and writes nothing. No sub, no status change, no rename, no acceptance.
func TestCompleteOnboarding_MissingAcceptance(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	seedPublishedAgreement(t, svc, seedAgreementManager(t, database), "1.0")
	email := "jdoe-" + uuid.NewString()[:8] + "@example.edu"
	user := seedPendingUserWithEmail(t, database, email)
	newClusterAccount(t, svc, user, "custos-jdoe")
	claims := &identity.Claims{Sub: "sub-" + uuid.NewString(), Email: email, EmailVerified: true}

	_, err := svc.CompleteOnboarding(ctx(), claims, OnboardingInput{Username: "jdoe"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if u, err := svc.GetUser(ctx(), user); err != nil || u.Status != models.UserPending {
		t.Errorf("status: got %+v, %v, want PENDING", u, err)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_identities WHERE user_id = $1", user); n != 0 {
		t.Errorf("identities: got %d, want 0", n)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_agreement_acceptances WHERE user_id = $1", user); n != 0 {
		t.Errorf("acceptances: got %d, want 0", n)
	}
	if accounts, err := svc.ListComputeClusterUsersByUser(ctx(), user); err != nil || len(accounts) != 1 || accounts[0].LocalUsername != "custos-jdoe" {
		t.Errorf("cluster account: got %+v, %v, want one named custos-jdoe", accounts, err)
	}
}

// Make sure a user the registry already knows finishes onboarding with their username unchanged.
func TestCompleteOnboarding_RegisteredUserKeepsUsername(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	aup := seedPublishedAgreement(t, svc, seedAgreementManager(t, database), "1.0")
	email := "jdoe-" + uuid.NewString()[:8] + "@example.edu"
	user := seedPendingUserWithEmail(t, database, email)
	newClusterAccount(t, svc, user, "custos-jdoe")
	if _, err := svc.CreateUserIdentity(ctx(), &models.UserIdentity{UserID: user, Source: "comanage", ExternalID: "100001"}); err != nil {
		t.Fatalf("create comanage identity: %v", err)
	}
	claims := &identity.Claims{Sub: "sub-" + uuid.NewString(), Email: email, EmailVerified: true}

	if _, err := svc.CompleteOnboarding(ctx(), claims, OnboardingInput{Username: "jdoe", AcceptedIDs: []string{aup.ID}}); err != nil {
		t.Fatalf("CompleteOnboarding: %v", err)
	}
	if accounts, err := svc.ListComputeClusterUsersByUser(ctx(), user); err != nil || len(accounts) != 1 || accounts[0].LocalUsername != "custos-jdoe" {
		t.Errorf("cluster account: got %+v, %v, want one named custos-jdoe", accounts, err)
	}
}

// Make sure there's no error accepting the same version of user agreement several times and only create one record
func TestAcceptUserAgreement_SimultaneousAcceptsRecordOne(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	aup := seedPublishedAgreement(t, svc, seedAgreementManager(t, database), "1.0")
	user := seedPendingUserWithEmail(t, database, "jdoe-"+uuid.NewString()[:8]+"@example.edu")

	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for range 10 {
		wg.Go(func() {
			_, err := svc.AcceptUserAgreement(ctx(), user, aup.ID)
			errs <- err
		})
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("accept: %v", err)
		}
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_agreement_acceptances WHERE user_id = $1", user); n != 1 {
		t.Errorf("acceptances: got %d, want 1", n)
	}
	if got := countAuditEventsOfType(t, database, "USER_AGREEMENT_ACCEPTED", user); got != 1 {
		t.Errorf("audit USER_AGREEMENT_ACCEPTED: got %d, want 1", got)
	}
}
