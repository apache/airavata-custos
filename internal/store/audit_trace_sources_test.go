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

package store_test

import (
	"slices"
	"testing"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/pkg/amie"
	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/pkg/comanage"
	"github.com/apache/airavata-custos/connectors/Notification/Email/pkg/email"
	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/pkg/smapper"
	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/pkg/vast"
	"github.com/apache/airavata-custos/internal/store"
)

func TestSourcesOfferEverySubscriber(t *testing.T) {
	sources, err := store.NewAuditTraceStore(nil).ListSources(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, subscriber := range []string{amie.Type, comanage.Type, email.Type, smapper.Type, vast.Type} {
		if !slices.Contains(sources, subscriber) {
			t.Errorf("missing %q", subscriber)
		}
	}
}
