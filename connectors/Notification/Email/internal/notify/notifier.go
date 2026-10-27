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

// Package notify emails researchers about their cluster account.
package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// AuditNotificationSent is written once per email.
const AuditNotificationSent = "NOTIFICATION_SENT"

// Notifier sends one email per handled event, so retrying a failed delivery resends exactly that email.
type Notifier struct {
	core   service.CoreService
	sender Sender
	site   Site
	source string
}

func NewNotifier(core service.CoreService, sender Sender, site Site) *Notifier {
	return &Notifier{core: core, sender: sender, site: site}
}

// RegisterSubscribers subscribes the handlers.
func (n *Notifier) RegisterSubscribers(bus *events.Bus, subscriber string) {
	n.source = subscriber
	bus.SubscribeComputeClusterUserCreated(subscriber, n.requestReceived)
	bus.SubscribeComputeClusterUserProvisioned(subscriber, n.accountReady)
	bus.SubscribeComputeClusterUserDenied(subscriber, n.requestDenied)
}

// requestReceived emails that Custos received a cluster account creation request and it is pending on an admin.
func (n *Notifier) requestReceived(ctx context.Context, cu models.ComputeClusterUser) error {
	if cu.ApprovalStatus != models.ClusterAccountPending {
		return nil
	}
	return n.send(ctx, RequestReceived, cu)
}

// accountReady emails that Custos provisioned a cluster account for a user.
func (n *Notifier) accountReady(ctx context.Context, cu models.ComputeClusterUser) error {
	return n.send(ctx, AccountReady, cu)
}

// requestDenied emails that Custos admin denied a cluster account creation request.
func (n *Notifier) requestDenied(ctx context.Context, cu models.ComputeClusterUser) error {
	return n.send(ctx, RequestDenied, cu)
}

func (n *Notifier) send(ctx context.Context, name string, cu models.ComputeClusterUser) error {
	user, err := n.core.GetUser(ctx, cu.UserID)
	if err != nil {
		return fmt.Errorf("look up user %s: %w", cu.UserID, err)
	}
	if user.Email == "" {
		return fmt.Errorf("%w: user %s has no email", events.ErrPermanent, user.ID)
	}
	project, err := n.project(ctx, cu)
	if err != nil {
		return fmt.Errorf("look up project for user %s: %w", cu.UserID, err)
	}

	d := Data{Site: n.site, FirstName: user.FirstName, Username: cu.LocalUsername}
	if project != nil {
		d.Source = strings.ToUpper(project.Origination)
		d.ProjectNumber = project.OriginatedID
	}
	if cu.ReviewNote != nil {
		d.Reason = *cu.ReviewNote
	}
	msg, err := Render(name, d)
	if err != nil {
		return fmt.Errorf("%w: render %s: %v", events.ErrPermanent, name, err)
	}
	msg.To = user.Email
	if err := n.sender.Send(ctx, msg); err != nil {
		return fmt.Errorf("send %s to %s: %w", name, user.Email, err)
	}

	// The email is already out, so a failed audit write is logged and not returned.
	details, _ := json.Marshal(map[string]string{"channel": "email", "template": name, "recipient": user.Email})
	if _, err := n.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  AuditNotificationSent,
		EntityID:   cu.ID,
		EntityType: "compute_cluster_user",
		Details:    string(details),
		Source:     n.source,
	}); err != nil {
		slog.WarnContext(ctx, "email sent but not audited", "template", name, "compute_cluster_user_id", cu.ID, "error", err)
	}
	return nil
}

// project finds the project the account was requested for.
// The first allocation on the cluster is the one requested because the cluster account is created only on the user's first request.
func (n *Notifier) project(ctx context.Context, cu models.ComputeClusterUser) (*models.Project, error) {
	memberships, err := n.core.ListAllocationsForUser(ctx, cu.UserID)
	if err != nil {
		return nil, err
	}
	for _, m := range memberships {
		allocation, err := n.core.GetComputeAllocation(ctx, m.ComputeAllocationID)
		if err != nil {
			return nil, err
		}
		if allocation.ComputeClusterID == cu.ComputeClusterID {
			return n.core.GetProject(ctx, allocation.ProjectID)
		}
	}
	return nil, nil
}
