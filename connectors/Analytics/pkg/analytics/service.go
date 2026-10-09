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

package analytics

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

// maxDailyBuckets caps the daily series so a stray far-past start_time cannot
// balloon the response.
const maxDailyBuckets = 400

const defaultJobsLimit = 20
const maxJobsLimit = 100

// roleMember is the derived (never persisted) role for a caller who reaches a
// project only through an allocation membership.
const roleMember = "MEMBER"

// Allocation is one allocation with its consumed credits.
type Allocation struct {
	ID              string    `json:"id" binding:"required"`
	Name            string    `json:"name" binding:"required"`
	Status          string    `json:"status"`
	InitialSUAmount int64     `json:"initial_su_amount" binding:"required"`
	UsedSUAmount    float64   `json:"used_su_amount" binding:"required"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time" binding:"required"`
}

// ProjectContext is one project the caller belongs to, their role on it, and
// its allocations.
type ProjectContext struct {
	ProjectID   string       `json:"project_id" binding:"required"`
	ProjectName string       `json:"project_name" binding:"required"`
	Role        string       `json:"role" binding:"required"`
	Allocations []Allocation `json:"allocations" binding:"required"`
}

// UsageDailyBucket is one day of consumption keyed by resource id. Buckets are
// continuous from allocation start to today, so a zero-usage day carries an
// empty map.
type UsageDailyBucket struct {
	Date       string             `json:"date" binding:"required"`
	ByResource map[string]float64 `json:"by_resource" binding:"required"`
}

// UsageResource aggregates one resource's consumption, including the caller's
// own slice. Cap is null in v1 (per-resource caps are time-varying).
type UsageResource struct {
	ResourceID   string  `json:"resource_id" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	ResourceType string  `json:"resource_type"`
	Used         float64 `json:"used" binding:"required"`
	Cap          *int64  `json:"cap" extensions:"x-nullable"`
	UsedNative   float64 `json:"used_native" binding:"required"`
	NativeUnit   string  `json:"native_unit" binding:"required"`
	UsedByCaller float64 `json:"used_by_caller" binding:"required"`
}

// UsageMember is one member's consumption against an allocation.
type UsageMember struct {
	UserID string  `json:"user_id" binding:"required"`
	Name   string  `json:"name" binding:"required"`
	Used   float64 `json:"used" binding:"required"`
}

// UsageSummary is the aggregated usage for one allocation. ByMember is null
// unless the caller may see per-member data.
type UsageSummary struct {
	Total      int64              `json:"total" binding:"required"`
	Used       float64            `json:"used"`
	Daily      []UsageDailyBucket `json:"daily" binding:"required"`
	ByResource []UsageResource    `json:"by_resource" binding:"required"`
	ByMember   []UsageMember      `json:"by_member" extensions:"x-nullable"`
}

// Job is one usage record (a job's charge).
type Job struct {
	ID             string    `json:"id" binding:"required"`
	JobID          string    `json:"job_id" binding:"required"`
	CalculatedTime time.Time `json:"calculated_time" binding:"required"`
	UserID         string    `json:"user_id" binding:"required"`
	UserName       string    `json:"user_name" binding:"required"`
	ResourceID     string    `json:"resource_id" binding:"required"`
	ResourceName   string    `json:"resource_name" binding:"required"`
	ResourceType   string    `json:"resource_type" binding:"required"`
	UsedRaw        float64   `json:"used_raw" binding:"required"`
	NativeUnit     string    `json:"native_unit" binding:"required"`
	Used           float64   `json:"used" binding:"required"`
}

// AllocationJobs is a page of jobs plus the total matching count.
type AllocationJobs struct {
	Jobs  []Job `json:"jobs" binding:"required"`
	Total int   `json:"total" binding:"required"`
}

// Service aggregates usage on top of core data. It reads usage through its own
// store and allocations and roles through the core service.
type Service struct {
	core  *coreservice.Service
	store Store
}

// NewService wires the analytics service against the shared database and core
// service.
func NewService(core *coreservice.Service, database *sqlx.DB) *Service {
	return &Service{core: core, store: NewStore(database)}
}

// GetAllocation loads one allocation, returning coreservice.ErrNotFound when it
// does not exist.
func (s *Service) GetAllocation(ctx context.Context, id string) (*models.ComputeAllocation, error) {
	return s.core.GetComputeAllocation(ctx, id)
}

// AllocationRoleForUser returns the caller's role on the allocation, or "" for
// none.
func (s *Service) AllocationRoleForUser(ctx context.Context, allocationID, userID string) (models.ProjectRole, error) {
	return s.core.AllocationRoleForUser(ctx, allocationID, userID)
}

// AllocationJobsPage returns a page of the allocation's usage records, newest
// first. mineOnly restricts to the caller's own records; the handler forces it
// for callers who may not see other members' data.
func (s *Service) AllocationJobsPage(ctx context.Context, allocationID, callerID string, mineOnly bool, limit, offset int) (*AllocationJobs, error) {
	if limit <= 0 {
		limit = defaultJobsLimit
	}
	if limit > maxJobsLimit {
		limit = maxJobsLimit
	}
	if offset < 0 {
		offset = 0
	}
	var filter *string
	if mineOnly {
		filter = &callerID
	}
	rows, total, err := s.store.Jobs(ctx, allocationID, filter, limit, offset)
	if err != nil {
		return nil, err
	}
	jobs := make([]Job, 0, len(rows))
	for _, r := range rows {
		jobs = append(jobs, Job{
			ID:             r.ID,
			JobID:          r.JobID,
			CalculatedTime: r.CalculatedTime,
			UserID:         r.UserID,
			UserName:       r.UserName,
			ResourceID:     r.ResourceID,
			ResourceName:   r.ResourceName,
			ResourceType:   r.ResourceType,
			UsedRaw:        r.UsedRawAmount,
			NativeUnit:     nativeUnit(r.ResourceType),
			Used:           r.UsedSUAmount,
		})
	}
	return &AllocationJobs{Jobs: jobs, Total: total}, nil
}

// AnalyticsContexts returns the projects the caller touches, each with the
// caller's role and its allocations. Empty (not nil) when the caller belongs
// to nothing.
func (s *Service) AnalyticsContexts(ctx context.Context, userID string) ([]ProjectContext, error) {
	projRows, err := s.store.ProjectsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(projRows) == 0 {
		return []ProjectContext{}, nil
	}
	allocs, err := s.core.ListComputeAllocationsForParticipant(ctx, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(allocs))
	for i, a := range allocs {
		ids[i] = a.ID
	}
	allocRows, err := s.store.AllocationsByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	byProject := make(map[string][]Allocation, len(projRows))
	for _, a := range allocRows {
		byProject[a.ProjectID] = append(byProject[a.ProjectID], Allocation{
			ID:              a.ID,
			Name:            a.Name,
			Status:          a.Status,
			InitialSUAmount: a.InitialSUAmount,
			UsedSUAmount:    a.UsedSUAmount,
			StartTime:       a.StartTime,
			EndTime:         a.EndTime,
		})
	}

	out := make([]ProjectContext, 0, len(projRows))
	for _, p := range projRows {
		role := roleMember
		if p.Role.Valid && p.Role.String != "" {
			role = p.Role.String
		}
		allocs := byProject[p.ProjectID]
		if allocs == nil {
			allocs = []Allocation{}
		}
		out = append(out, ProjectContext{
			ProjectID:   p.ProjectID,
			ProjectName: p.Title,
			Role:        role,
			Allocations: allocs,
		})
	}
	return out, nil
}

// AllocationUsageSummary aggregates one allocation's usage. includeMembers
// controls whether the per-member breakdown is populated; when false ByMember
// is left nil so the response serializes it as null.
func (s *Service) AllocationUsageSummary(ctx context.Context, alloc *models.ComputeAllocation, callerID string, includeMembers bool) (*UsageSummary, error) {
	// Round the total the same way the per-resource and daily sums do, so the
	// headline reconciles with the breakdowns (used_su_amount is a DOUBLE).
	used, err := s.store.TotalUsed(ctx, alloc.ID)
	if err != nil {
		return nil, err
	}
	daily, err := s.store.DailyUsage(ctx, alloc.ID)
	if err != nil {
		return nil, err
	}
	resources, err := s.store.ResourceUsage(ctx, alloc.ID, callerID)
	if err != nil {
		return nil, err
	}

	summary := &UsageSummary{
		Total:      alloc.InitialSUAmount,
		Used:       used,
		Daily:      buildDailyBuckets(alloc.StartTime, daily),
		ByResource: buildResourceBreakdown(resources),
	}

	if includeMembers {
		members, err := s.store.MemberUsage(ctx, alloc.ID)
		if err != nil {
			return nil, err
		}
		out := make([]UsageMember, 0, len(members))
		for _, m := range members {
			out = append(out, UsageMember{UserID: m.UserID, Name: m.Name, Used: m.Used})
		}
		summary.ByMember = out
	}
	return summary, nil
}

func buildResourceBreakdown(rows []ResourceRow) []UsageResource {
	out := make([]UsageResource, 0, len(rows))
	for _, r := range rows {
		out = append(out, UsageResource{
			ResourceID:   r.ResourceID,
			Name:         r.Name,
			ResourceType: r.ResourceType,
			Used:         r.Used,
			Cap:          nil,
			UsedNative:   r.UsedNative,
			NativeUnit:   nativeUnit(r.ResourceType),
			UsedByCaller: r.UsedByCaller,
		})
	}
	return out
}

// buildDailyBuckets returns a continuous per-day series from the allocation
// start to today. The range is capped against a stray start_time.
func buildDailyBuckets(start time.Time, rows []DailyRow) []UsageDailyBucket {
	const layout = "2006-01-02"
	byDay := make(map[string]map[string]float64, len(rows))
	for _, r := range rows {
		day := r.Day.Format(layout)
		if byDay[day] == nil {
			byDay[day] = map[string]float64{}
		}
		byDay[day][r.ResourceID] += r.Credits
	}

	now := time.Now().UTC()
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	s := start.UTC()
	first := time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, time.UTC)
	if first.After(end) {
		first = end
	}
	if end.Sub(first) > maxDailyBuckets*24*time.Hour {
		first = end.AddDate(0, 0, -(maxDailyBuckets - 1))
	}

	buckets := make([]UsageDailyBucket, 0)
	for d := first; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format(layout)
		m := byDay[key]
		if m == nil {
			m = map[string]float64{}
		}
		buckets = append(buckets, UsageDailyBucket{Date: key, ByResource: m})
	}
	return buckets
}

// nativeUnit maps a resource_type to the unit for its native amount. Unknown
// types fall back to the raw type string.
func nativeUnit(resourceType string) string {
	switch resourceType {
	case "CPU_HOURS":
		return "core-hours"
	case "GPU_HOURS":
		return "GPU-hours"
	case "STORAGE_TB":
		return "TB-months"
	default:
		return resourceType
	}
}
