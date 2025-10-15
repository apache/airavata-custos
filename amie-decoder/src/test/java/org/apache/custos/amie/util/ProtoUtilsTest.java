/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */
package org.apache.custos.amie.util;

import org.apache.custos.amie.internal.events.v1.DecodeStartedPayload;
import org.apache.custos.amie.internal.events.v1.EventType;
import org.apache.custos.amie.internal.events.v1.ProcessingEvent;
import org.apache.custos.amie.internal.events.v1.ProcessingStatus;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class ProtoUtilsTest {

    @Test
    void createDecodeStartedEvent_shouldCreateValidEvent() {
        String eventId = "event-123";
        String packetId = "packet-456";
        long amiePacketRecId = 789L;

        byte[] result = ProtoUtils.createDecodeStartedEvent(eventId, packetId, amiePacketRecId);

        assertThat(result).isNotNull();
        assertThat(result.length).isGreaterThan(0);

        // Verify we can deserialize it back
        try {
            ProcessingEvent event = ProcessingEvent.parseFrom(result);
            assertThat(event.getId()).isEqualTo(eventId);
            assertThat(event.getPacketDbId()).isEqualTo(packetId);
            assertThat(event.getAmiePacketRecId()).isEqualTo(amiePacketRecId);
            assertThat(event.getType()).isEqualTo(EventType.DECODE_STARTED);
            assertThat(event.getStatus()).isEqualTo(ProcessingStatus.PENDING);
            assertThat(event.getAttempts()).isEqualTo(0);
            assertThat(event.hasDecodeStarted()).isTrue();

            DecodeStartedPayload payload = event.getDecodeStarted();
            assertThat(payload.getPacketDbId()).isEqualTo(packetId);
            assertThat(payload.getAmiePacketRecId()).isEqualTo(amiePacketRecId);
        } catch (Exception e) {
            throw new RuntimeException("Failed to deserialize protobuf", e);
        }
    }

    @Test
    void createDecodeStartedEvent_withEmptyEventId_shouldCreateEventWithEmptyId() {
        String eventId = "";
        String packetId = "packet-456";
        long amiePacketRecId = 789L;

        byte[] result = ProtoUtils.createDecodeStartedEvent(eventId, packetId, amiePacketRecId);

        assertThat(result).isNotNull();

        try {
            ProcessingEvent event = ProcessingEvent.parseFrom(result);
            assertThat(event.getId()).isEmpty();
        } catch (Exception e) {
            throw new RuntimeException("Failed to deserialize protobuf", e);
        }
    }

    @Test
    void createDecodeStartedEvent_withEmptyPacketId_shouldCreateEventWithEmptyPacketId() {
        String eventId = "event-123";
        String packetId = "";
        long amiePacketRecId = 789L;

        byte[] result = ProtoUtils.createDecodeStartedEvent(eventId, packetId, amiePacketRecId);

        assertThat(result).isNotNull();

        try {
            ProcessingEvent event = ProcessingEvent.parseFrom(result);
            assertThat(event.getPacketDbId()).isEmpty();
        } catch (Exception e) {
            throw new RuntimeException("Failed to deserialize protobuf", e);
        }
    }

    @Test
    void createDecodeStartedEvent_withZeroAmiePacketRecId_shouldCreateEventWithZeroId() {
        String eventId = "event-123";
        String packetId = "packet-456";
        long amiePacketRecId = 0L;

        byte[] result = ProtoUtils.createDecodeStartedEvent(eventId, packetId, amiePacketRecId);

        assertThat(result).isNotNull();

        try {
            ProcessingEvent event = ProcessingEvent.parseFrom(result);
            assertThat(event.getAmiePacketRecId()).isEqualTo(0L);
        } catch (Exception e) {
            throw new RuntimeException("Failed to deserialize protobuf", e);
        }
    }

    @Test
    void createDecodeStartedEvent_withNegativeAmiePacketRecId_shouldCreateEventWithNegativeId() {
        String eventId = "event-123";
        String packetId = "packet-456";
        long amiePacketRecId = -1L;

        byte[] result = ProtoUtils.createDecodeStartedEvent(eventId, packetId, amiePacketRecId);

        assertThat(result).isNotNull();

        try {
            ProcessingEvent event = ProcessingEvent.parseFrom(result);
            assertThat(event.getAmiePacketRecId()).isEqualTo(-1L);
        } catch (Exception e) {
            throw new RuntimeException("Failed to deserialize protobuf", e);
        }
    }
}