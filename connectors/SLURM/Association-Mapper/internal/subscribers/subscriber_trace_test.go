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

package subscribers

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	prev := otel.GetTracerProvider()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
	return sr
}

func mockCoreServiceWithError(err error) *service.CoreServiceMock {
	return &service.CoreServiceMock{
		GetComputeClusterFunc: func(ctx context.Context, id string) (*models.ComputeCluster, error) {
			return nil, err
		},
		GetComputeAllocationFunc: func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
			return nil, err
		},
		GetComputeAllocationMembershipFunc: func(ctx context.Context, id string) (*models.ComputeAllocationMembership, error) {
			return nil, err
		},
		CreateAuditEventFunc: func(ctx context.Context, event *models.AuditEvent) (*models.AuditEvent, error) {
			return event, nil
		},
	}
}

func findSpan(spans []sdktrace.ReadOnlySpan, name string) sdktrace.ReadOnlySpan {
	for _, s := range spans {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func TestSubscriberSpansOpenUnderParent(t *testing.T) {
	sr := setupRecorder(t)

	core := mockCoreServiceWithError(errors.New("mock get failed"))
	sub := NewAssociationSubscriber(nil, nil, core)

	ctx, root := tracing.Start(context.Background(), "test.parent")
	rootTrace := root.SpanContext().TraceID()
	rootSpan := root.SpanContext().SpanID()

	sub.SubscribeToComputeAllocationCreation(ctx, models.ComputeAllocation{
		ID: "alloc-1", ComputeClusterID: "cluster-1",
	})
	sub.SubscribeToComputeAllocationResourceMappingCreation(ctx, models.ComputeAllocationResourceMapping{
		ID: "mapping-1", ComputeAllocationID: "alloc-1",
	})
	sub.SubscribeToComputeAllocationMembershipCreation(ctx, models.ComputeAllocationMembership{
		ID: "mem-1", ComputeAllocationID: "alloc-1", UserID: "user-1",
	})
	sub.SubscribeToComputeAllocationMembershipResourceOverrideCreation(ctx, models.ComputeAllocationMembershipResourceOverride{
		ID: "override-1", ComputeAllocationMembershipID: "mem-1",
	})
	root.End()

	want := []string{
		"slurm.compute_allocation_create",
		"slurm.compute_allocation_resource_mapping_create",
		"slurm.compute_allocation_membership_create",
		"slurm.compute_allocation_membership_resource_override_create",
	}

	ended := sr.Ended()
	for _, name := range want {
		s := findSpan(ended, name)
		if s == nil {
			t.Fatalf("expected span %q to be recorded", name)
		}
		if s.SpanContext().TraceID() != rootTrace {
			t.Fatalf("span %q trace id mismatch: got %s want %s", name, s.SpanContext().TraceID(), rootTrace)
		}
		if s.Parent().SpanID() != rootSpan {
			t.Fatalf("span %q expected parent %s, got %s", name, rootSpan, s.Parent().SpanID())
		}
	}
}

func TestSubscriberSpansRecordErrorOnDownstreamFailure(t *testing.T) {
	sr := setupRecorder(t)

	core := mockCoreServiceWithError(errors.New("downstream failure"))
	sub := NewAssociationSubscriber(nil, nil, core)

	ctx, root := tracing.Start(context.Background(), "test.parent")
	sub.SubscribeToComputeAllocationCreation(ctx, models.ComputeAllocation{
		ID: "alloc-err", ComputeClusterID: "cluster-err",
	})
	root.End()

	s := findSpan(sr.Ended(), "slurm.compute_allocation_create")
	if s == nil {
		t.Fatalf("expected slurm.compute_allocation_create span")
	}
	if s.Status().Code.String() != "Error" {
		t.Fatalf("expected span status Error, got %s", s.Status().Code.String())
	}
	if len(s.Events()) == 0 {
		t.Fatalf("expected exception event recorded on span")
	}
}
