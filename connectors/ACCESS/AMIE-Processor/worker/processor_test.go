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

package worker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ComputeNextRetryAt tests
// ---------------------------------------------------------------------------

// TestComputeNextRetryAt_Attempt1Is30s verifies that after the first failure
// the backoff delay is exactly 30 seconds (BaseBackoffSeconds * 2^0).
func TestComputeNextRetryAt_Attempt1Is30s(t *testing.T) {
	before := time.Now().UTC()
	next := ComputeNextRetryAt(1)
	after := time.Now().UTC()

	lowerBound := before.Add(30 * time.Second)
	upperBound := after.Add(30 * time.Second)

	assert.True(t, !next.Before(lowerBound), "retry time %v should be >= %v", next, lowerBound)
	assert.True(t, !next.After(upperBound), "retry time %v should be <= %v", next, upperBound)
}

// TestComputeNextRetryAt_Attempt2Is60s verifies that after the second failure
// the delay doubles to 60 seconds (BaseBackoffSeconds * 2^1).
func TestComputeNextRetryAt_Attempt2Is60s(t *testing.T) {
	before := time.Now().UTC()
	next := ComputeNextRetryAt(2)
	after := time.Now().UTC()

	lowerBound := before.Add(60 * time.Second)
	upperBound := after.Add(60 * time.Second)

	assert.True(t, !next.Before(lowerBound), "retry time %v should be >= %v", next, lowerBound)
	assert.True(t, !next.After(upperBound), "retry time %v should be <= %v", next, upperBound)
}

// TestComputeNextRetryAt_Attempt3Is120s verifies the third attempt gives 120s.
func TestComputeNextRetryAt_Attempt3Is120s(t *testing.T) {
	before := time.Now().UTC()
	next := ComputeNextRetryAt(3)
	after := time.Now().UTC()

	lowerBound := before.Add(120 * time.Second)
	upperBound := after.Add(120 * time.Second)

	assert.True(t, !next.Before(lowerBound), "retry time %v should be >= %v", next, lowerBound)
	assert.True(t, !next.After(upperBound), "retry time %v should be <= %v", next, upperBound)
}

// TestComputeNextRetryAt_Attempt6IsCapped verifies that attempt 6 (and above
// where the bit shift does not overflow) is capped at MaxBackoffSeconds (600s).
// Note: attempt=6 gives 30*2^5=960 which gets capped to 600.
func TestComputeNextRetryAt_Attempt6IsCapped(t *testing.T) {
	before := time.Now().UTC()
	next := ComputeNextRetryAt(6)
	after := time.Now().UTC()

	// Must be capped at MaxBackoffSeconds (600s), NOT larger.
	lowerBound := before.Add(MaxBackoffSeconds * time.Second)
	upperBound := after.Add(MaxBackoffSeconds * time.Second)

	assert.True(t, !next.Before(lowerBound), "retry time %v should be >= %v", next, lowerBound)
	assert.True(t, !next.After(upperBound), "retry time %v should be <= %v", next, upperBound)
}

// TestComputeNextRetryAt_BackoffFormulaProducesCorrectDurations validates the
// full formula for a range of attempt values without relying on clock drift.
// Note: Go's integer bit-shift for attempt=100 overflows to 0, so delay is 0s.
func TestComputeNextRetryAt_BackoffFormulaProducesCorrectDurations(t *testing.T) {
	tests := []struct {
		attempt      int
		expectedSecs int
	}{
		{1, 30},
		{2, 60},
		{3, 120},
		{4, 240},
		{5, 480},
		{6, 600}, // 30*2^5=960 -> capped at 600
		{7, 600},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("", func(t *testing.T) {
			before := time.Now().UTC()
			next := ComputeNextRetryAt(tc.attempt)

			// Allow 200ms tolerance for test execution time.
			elapsed := next.Sub(before)
			expected := time.Duration(tc.expectedSecs) * time.Second

			assert.GreaterOrEqual(t, elapsed, expected-100*time.Millisecond,
				"attempt %d: elapsed %v should be ~%v", tc.attempt, elapsed, expected)
			assert.LessOrEqual(t, elapsed, expected+500*time.Millisecond,
				"attempt %d: elapsed %v should be ~%v", tc.attempt, elapsed, expected)
		})
	}
}
