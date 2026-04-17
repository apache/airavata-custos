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

	pb "github.com/apache/airavata-custos/allocations/access-amie/proto/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// ---------------------------------------------------------------------------
// CreateDecodeStartedEvent tests
// ---------------------------------------------------------------------------

func TestCreateDecodeStartedEvent_ValidEventDeserializeVerifyFields(t *testing.T) {
	eventID := "evt-abc-123"
	packetID := "pkt-xyz-456"
	var amieRecID int64 = 99

	data, err := CreateDecodeStartedEvent(eventID, packetID, amieRecID)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var event pb.ProcessingEvent
	require.NoError(t, proto.Unmarshal(data, &event))

	assert.Equal(t, eventID, event.GetId())
	assert.Equal(t, packetID, event.GetPacketDbId())
	assert.Equal(t, amieRecID, event.GetAmiePacketRecId())
	assert.Equal(t, pb.EventType_DECODE_STARTED, event.GetType())
	assert.Equal(t, pb.ProcessingStatus_PENDING, event.GetStatus())
	assert.Equal(t, int32(0), event.GetAttempts())
	assert.NotNil(t, event.GetCreatedAt())

	payload := event.GetDecodeStarted()
	require.NotNil(t, payload)
	assert.Equal(t, packetID, payload.GetPacketDbId())
	assert.Equal(t, amieRecID, payload.GetAmiePacketRecId())
}

func TestCreateDecodeStartedEvent_EmptyEventIDAndPacketID(t *testing.T) {
	data, err := CreateDecodeStartedEvent("", "", 1)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var event pb.ProcessingEvent
	require.NoError(t, proto.Unmarshal(data, &event))

	assert.Equal(t, "", event.GetId())
	assert.Equal(t, "", event.GetPacketDbId())
	assert.Equal(t, int64(1), event.GetAmiePacketRecId())
}

func TestCreateDecodeStartedEvent_ZeroAmiePacketRecID(t *testing.T) {
	data, err := CreateDecodeStartedEvent("evt-1", "pkt-1", 0)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var event pb.ProcessingEvent
	require.NoError(t, proto.Unmarshal(data, &event))

	assert.Equal(t, int64(0), event.GetAmiePacketRecId())
	payload := event.GetDecodeStarted()
	require.NotNil(t, payload)
	assert.Equal(t, int64(0), payload.GetAmiePacketRecId())
}
