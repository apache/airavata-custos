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
	"time"

	pb "github.com/apache/airavata-custos/allocations/access-amie/proto/gen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CreateDecodeStartedEvent(eventID, packetID string, amiePacketRecID int64) ([]byte, error) {
	payload := &pb.DecodeStartedPayload{
		PacketDbId:      packetID,
		AmiePacketRecId: amiePacketRecID,
	}
	event := &pb.ProcessingEvent{
		Id:              eventID,
		PacketDbId:      packetID,
		AmiePacketRecId: amiePacketRecID,
		Type:            pb.EventType_DECODE_STARTED,
		Status:          pb.ProcessingStatus_PENDING,
		Attempts:        0,
		CreatedAt:       timestamppb.New(time.Now()),
		Payload:         &pb.ProcessingEvent_DecodeStarted{DecodeStarted: payload},
	}
	return proto.Marshal(event)
}
