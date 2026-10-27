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

package notify

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// mockRelay stands in for the SMTP relay and keeps what it was given.
type mockRelay struct {
	sent []Message
	err  error
}

func (r *mockRelay) Send(_ context.Context, m Message) error {
	if r.err != nil {
		return r.err
	}
	r.sent = append(r.sent, m)
	return nil
}

// mockCore is a core holding Jane Doe with an ACCESS allocation on the
// "hpc" cluster and an unrelated one on another cluster listed first.
func mockCore(email string) *service.CoreServiceMock {
	allocations := map[string]*models.ComputeAllocation{
		"alloc-other": {ID: "alloc-other", ProjectID: "proj-other", ComputeClusterID: "other"},
		"alloc-hpc":   {ID: "alloc-hpc", ProjectID: "proj-hpc", ComputeClusterID: "hpc"},
	}
	projects := map[string]*models.Project{
		"proj-other": {ID: "proj-other", OriginatedID: "OTHER1", Origination: "nairr"},
		"proj-hpc":   {ID: "proj-hpc", OriginatedID: "CIS250123", Origination: "access"},
	}
	return &service.CoreServiceMock{
		GetUserFunc: func(_ context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, FirstName: "Jane", Email: email}, nil
		},
		ListAllocationsForUserFunc: func(context.Context, string) ([]models.ComputeAllocationMembership, error) {
			return []models.ComputeAllocationMembership{{ComputeAllocationID: "alloc-other"}, {ComputeAllocationID: "alloc-hpc"}}, nil
		},
		GetComputeAllocationFunc: func(_ context.Context, id string) (*models.ComputeAllocation, error) {
			return allocations[id], nil
		},
		GetProjectFunc: func(_ context.Context, id string) (*models.Project, error) {
			return projects[id], nil
		},
		CreateAuditEventFunc: func(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
			return e, nil
		},
	}
}

func testNotifier(core service.CoreService, relay Sender) *Notifier {
	n := NewNotifier(core, relay, Site{SiteName: "Example HPC", PortalURL: "https://portal.example.edu", ClusterHost: "login.example.edu", SupportEmail: "help@example.edu"})
	n.source = "email-notifier"
	return n
}

func pendingAccount() models.ComputeClusterUser {
	return models.ComputeClusterUser{ID: "cu-1", UserID: "u-1", ComputeClusterID: "hpc", LocalUsername: "jdoe", ApprovalStatus: models.ClusterAccountPending}
}

// Make sure only a pending account gets the request received email. The admin directly added
// accounts are already approved, and those users get only the ready email.
func TestRequestReceived_OnlyForPendingAccounts(t *testing.T) {
	relay := &mockRelay{}
	n := testNotifier(mockCore("jdoe@example.edu"), relay)

	approved := pendingAccount()
	approved.ApprovalStatus = models.ClusterAccountApproved
	if err := n.requestReceived(context.Background(), approved); err != nil {
		t.Fatal(err)
	}
	if len(relay.sent) != 0 {
		t.Fatalf("an approved account got %d request received emails", len(relay.sent))
	}

	if err := n.requestReceived(context.Background(), pendingAccount()); err != nil {
		t.Fatal(err)
	}
	if len(relay.sent) != 1 || relay.sent[0].To != "jdoe@example.edu" {
		t.Fatalf("expected one email to jdoe@example.edu, got %+v", relay.sent)
	}
}

// Make sure the email indicates the project of the allocation on the account's cluster.
func TestRequestReceived_NamesTheProjectOnTheSameCluster(t *testing.T) {
	relay := &mockRelay{}
	n := testNotifier(mockCore("jdoe@example.edu"), relay)

	if err := n.requestReceived(context.Background(), pendingAccount()); err != nil {
		t.Fatal(err)
	}
	text := relay.sent[0].Text
	if !strings.Contains(text, "Requested through: ACCESS") || !strings.Contains(text, "Project: CIS250123") {
		t.Fatalf("expected the ACCESS project CIS250123, got:\n%s", text)
	}
}

// Make sure each email sent records an audit event.
func TestSend_AuditsEachSentEmail(t *testing.T) {
	core := mockCore("jdoe@example.edu")
	n := testNotifier(core, &mockRelay{})

	if err := n.accountReady(context.Background(), pendingAccount()); err != nil {
		t.Fatal(err)
	}
	calls := core.CreateAuditEventCalls()
	if len(calls) != 1 {
		t.Fatalf("expected one audit row, got %d", len(calls))
	}
	e := calls[0].E
	if e.EventType != AuditNotificationSent || e.EntityID != "cu-1" || e.Source != "email-notifier" ||
		e.Details != `{"channel":"email","recipient":"jdoe@example.edu","template":"account-ready"}` {
		t.Fatalf("unexpected audit row %+v", e)
	}
}

// Make sure an email send failure is returned, so the event worker records the failed delivery and retries it.
func TestSend_RelayFailureIsRetriedAndNotRecordedAsSent(t *testing.T) {
	core := mockCore("jdoe@example.edu")
	n := testNotifier(core, &mockRelay{err: errors.New("421 try again later")})

	err := n.accountReady(context.Background(), pendingAccount())
	if err == nil || errors.Is(err, events.ErrPermanent) {
		t.Fatalf("expected a retryable error, got %v", err)
	}
	if len(core.CreateAuditEventCalls()) != 0 {
		t.Fatal("a failed send was audited as sent")
	}
}

// Make sure a user with no email fails the delivery at once permanently since retrying cannot fix it.
func TestSend_UserWithoutEmailFailsPermanently(t *testing.T) {
	relay := &mockRelay{}
	n := testNotifier(mockCore(""), relay)

	if err := n.accountReady(context.Background(), pendingAccount()); !errors.Is(err, events.ErrPermanent) {
		t.Fatalf("expected a permanent failure, got %v", err)
	}
	if len(relay.sent) != 0 {
		t.Fatal("an email was sent with no recipient")
	}
}

// Make sure the denied email carries the reason the admin gave.
func TestRequestDenied_CarriesTheReason(t *testing.T) {
	relay := &mockRelay{}
	n := testNotifier(mockCore("jdoe@example.edu"), relay)
	cu := pendingAccount()
	reason := "We could not confirm your allocation."
	cu.ApprovalStatus, cu.ReviewNote = models.ClusterAccountDenied, &reason

	if err := n.requestDenied(context.Background(), cu); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(relay.sent[0].Text, reason) || !strings.Contains(relay.sent[0].HTML, reason) {
		t.Fatal("the denied email does not carry the reason")
	}
}
