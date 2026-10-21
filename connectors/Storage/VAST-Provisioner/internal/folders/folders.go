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

// Package folders keeps the registry of folders the connector made. A deleted owner's folders wait as
// pending_deletion until an admin deletes them from VAST.
package folders

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

//go:embed migrations/*.sql
var Migrations embed.FS

type Folder struct {
	ID             string        `db:"id" json:"id"`
	ClusterID      string        `db:"compute_cluster_id" json:"cluster_id"`
	Path           string        `db:"path" json:"path"`
	OwnerKind      string        `db:"owner_kind" json:"owner_kind"`
	OwnerName      string        `db:"owner_name" json:"owner_name"`
	UnixOwner      string        `db:"unix_owner" json:"unix_owner"`
	Status         string        `db:"status" json:"status"`
	AllocationID   *string       `db:"allocation_id" json:"allocation_id"`
	AllocationName *string       `db:"allocation_name" json:"allocation_name"`
	StartTime      *time.Time    `db:"start_time" json:"start_time"`
	EndTime        *time.Time    `db:"end_time" json:"end_time"`
	QuotaOverride  *int64        `db:"quota_override" json:"quota_override"`
	QuotaUntil     *time.Time    `db:"quota_until" json:"quota_until"`
	Quota          *client.Quota `db:"-" json:"quota"`
}

// VMS is what the review needs from a cluster's VAST.
type VMS interface {
	Remove(ctx context.Context, path string) error
	Quotas(ctx context.Context) (map[string]client.Quota, error)
	CreateQuota(ctx context.Context, path string, hardLimit, hardLimitInodes int64) error
}

// Auditor records an admin's change.
type Auditor interface {
	CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error)
}

// Registry serves the review; vms finds a cluster's VAST.
type Registry struct {
	db    *sqlx.DB
	vms   func(clusterID string) (VMS, error)
	audit Auditor
}

func New(db *sqlx.DB, vms func(clusterID string) (VMS, error), audit Auditor) *Registry {
	return &Registry{db: db, vms: vms, audit: audit}
}

// Record notes a folder made for a live owner, so a returning owner's folder is active again.
// It keeps a copy of the folder's allocation, a nil for a user's own folder, so the review shows it once deleted, and
// returns an admin's hard limit while it lasts, else 0.
func (s *Registry) Record(ctx context.Context, clusterID, path, ownerKind, ownerName, unixOwner string, a *models.ComputeAllocation) (int64, error) {
	var id, name *string
	var start, end *time.Time
	if a != nil {
		id, name, start, end = &a.ID, &a.Name, &a.StartTime, &a.EndTime
	}
	var override int64
	err := s.db.GetContext(ctx, &override, `INSERT INTO vast_folders
		(id, compute_cluster_id, path, owner_kind, owner_name, unix_owner, allocation_id, allocation_name, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (compute_cluster_id, path) DO UPDATE SET owner_kind = EXCLUDED.owner_kind, owner_name = EXCLUDED.owner_name,
			unix_owner = EXCLUDED.unix_owner, allocation_id = EXCLUDED.allocation_id, allocation_name = EXCLUDED.allocation_name,
			start_time = EXCLUDED.start_time, end_time = EXCLUDED.end_time, status = 'active'
		RETURNING COALESCE(CASE WHEN quota_until > now() THEN quota_override END, 0)`,
		uuid.NewString(), clusterID, path, ownerKind, ownerName, unixOwner, id, name, start, end)
	return override, err
}

// MarkPending marks the owner's active folders and those under them, such as members' folders in a project.
func (s *Registry) MarkPending(ctx context.Context, clusterID, ownerKind, ownerName string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE vast_folders f SET status = 'pending_deletion'
		WHERE f.compute_cluster_id = $1 AND f.status = 'active' AND EXISTS (SELECT 1 FROM vast_folders p
			WHERE p.compute_cluster_id = $1 AND p.owner_kind = $2 AND p.owner_name = $3
			AND (f.path = p.path OR starts_with(f.path, p.path || '/')))`, clusterID, ownerKind, ownerName)
	return err
}

func (s *Registry) RegisterRoutes(router *identity.Router) {
	router.RequirePrivilege("GET /connectors/vast/folders", models.ClustersRead, s.list)
	router.RequirePrivilege("DELETE /connectors/vast/folders/{id}", models.ClustersWrite, s.deleteFolder)
	router.RequirePrivilege("PUT /connectors/vast/folders/{id}/quota", models.ClustersWrite, s.setQuota)
}

// setQuota applies an admin's hard limit to a folder's own quota until a date at most 90 days away; the reconciler
// restores the configured limit after it. The change is audited under "<cluster>:<path>", so it outlives the folder.
func (s *Registry) setQuota(w http.ResponseWriter, r *http.Request) {
	var b struct {
		HardLimit int64     `json:"hard_limit"`
		Until     time.Time `json:"until"`
		Reason    string    `json:"reason"`
	}
	if err := common.DecodeJSON(r, &b); err != nil || b.HardLimit <= 0 || !b.Until.After(time.Now()) || b.Until.After(time.Now().AddDate(0, 0, 90)) {
		common.WriteError(w, http.StatusBadRequest, errors.New("hard_limit must be positive and until within the next 90 days"))
		return
	}
	var f Folder
	if err := s.db.GetContext(r.Context(), &f, `SELECT * FROM vast_folders WHERE id = $1`, r.PathValue("id")); err != nil {
		common.WriteError(w, http.StatusNotFound, errors.New("no such folder"))
		return
	}
	vms, err := s.vms(f.ClusterID)
	var quotas map[string]client.Quota
	if err == nil {
		quotas, err = vms.Quotas(r.Context())
	}
	q, ok := quotas[f.Path]
	if err == nil && !ok {
		common.WriteError(w, http.StatusNotFound, errors.New("the folder has no quota of its own"))
		return
	}
	if err == nil {
		err = vms.CreateQuota(r.Context(), f.Path, b.HardLimit, q.HardLimitInodes)
	}
	if err != nil {
		common.WriteError(w, http.StatusBadGateway, err)
		return
	}
	details, _ := json.Marshal(map[string]any{"from": q.HardLimit, "to": b.HardLimit, "until": b.Until, "reason": b.Reason,
		"actor": identity.CallerFromContext(r.Context()).UserID})
	if _, err = s.db.ExecContext(r.Context(), `UPDATE vast_folders SET quota_override = $2, quota_until = $3 WHERE id = $1`, f.ID, b.HardLimit, b.Until); err == nil {
		_, err = s.audit.CreateAuditEvent(r.Context(), &models.AuditEvent{EventType: "VAST_QUOTA_CHANGED", EntityType: "vast_folder",
			EntityID: f.ClusterID + ":" + f.Path, Source: "vast", Details: string(details)})
	}
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Registry) list(w http.ResponseWriter, r *http.Request) {
	items := []Folder{}
	if err := s.db.SelectContext(r.Context(), &items, `SELECT * FROM vast_folders ORDER BY status <> 'pending_deletion', path`); err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	quotas := map[string]map[string]client.Quota{}
	for i, f := range items {
		q, ok := quotas[f.ClusterID]
		if !ok {
			vms, err := s.vms(f.ClusterID)
			if err == nil {
				q, err = vms.Quotas(r.Context())
			}
			if err != nil {
				slog.WarnContext(r.Context(), "VAST quotas unavailable; folders listed without them", "cluster_id", f.ClusterID, "err", err)
			}
			quotas[f.ClusterID] = q
		}
		if fq, ok := q[f.Path]; ok {
			items[i].Quota = &fq
		}
	}
	common.WriteJSON(w, http.StatusOK, items)
}

// deleteFolder drops a pending folder and the folders under it, which go with it, and commits once VAST removed it, so
// a failure leaves the folder pending and a concurrent deletion finds nothing.
func (s *Registry) deleteFolder(w http.ResponseWriter, r *http.Request) {
	tx, err := s.db.BeginTxx(r.Context(), nil)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()
	var f Folder
	err = tx.GetContext(r.Context(), &f, `DELETE FROM vast_folders f WHERE id = $1 AND status = 'pending_deletion' AND NOT EXISTS (SELECT 1 FROM vast_folders c
		WHERE c.compute_cluster_id = f.compute_cluster_id AND starts_with(c.path, f.path || '/') AND c.status <> 'pending_deletion') RETURNING *`, r.PathValue("id"))
	if errors.Is(err, sql.ErrNoRows) {
		common.WriteError(w, http.StatusNotFound, errors.New("no pending folder, or something under it is not pending"))
		return
	}
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `DELETE FROM vast_folders WHERE compute_cluster_id = $1 AND starts_with(path, $2 || '/')`, f.ClusterID, f.Path)
	}
	if err == nil {
		var vms VMS
		if vms, err = s.vms(f.ClusterID); err == nil {
			err = vms.Remove(r.Context(), f.Path)
		}
		if err != nil {
			common.WriteError(w, http.StatusBadGateway, err)
			return
		}
		err = tx.Commit()
	}
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
