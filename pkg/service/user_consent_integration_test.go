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
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

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
