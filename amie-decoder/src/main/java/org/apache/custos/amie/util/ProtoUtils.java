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

import com.google.protobuf.Timestamp;
import org.apache.custos.amie.internal.events.v1.DecodeStartedPayload;
import org.apache.custos.amie.internal.events.v1.EventType;
import org.apache.custos.amie.internal.events.v1.ProcessingEvent;
import org.apache.custos.amie.internal.events.v1.ProcessingStatus;

import java.time.Instant;

public final class ProtoUtils {

    private ProtoUtils() {
    }

    /**
     * Creates the serialized payload for a DECODE_STARTED event.
     * <p>
     * This method builds the full {@link ProcessingEvent} envelope with its payload
     * set to a {@link DecodeStartedPayload}. The resulting byte array is what will be
     * stored in the {@code payload} column of the {@code processing_events} table.
     *
     * @param eventId         ID of the event entity
     * @param packetId        ID of the parent packet entity
     * @param amiePacketRecId The unique record ID from the original AMIE packet header
     * @return A byte array containing the serialized Protobuf message
     */
    public static byte[] createDecodeStartedEvent(String eventId, String packetId, long amiePacketRecId) {
        DecodeStartedPayload payload = DecodeStartedPayload.newBuilder()
                .setPacketDbId(packetId)
                .setAmiePacketRecId(amiePacketRecId)
                .build();

        ProcessingEvent event = ProcessingEvent.newBuilder()
                .setId(eventId)
                .setPacketDbId(packetId)
                .setAmiePacketRecId(amiePacketRecId)
                .setType(EventType.DECODE_STARTED)
                .setStatus(ProcessingStatus.PENDING)
                .setAttempts(0)
                .setCreatedAt(instantToTimestamp(Instant.now()))
                .setDecodeStarted(payload)
                .build();

        return event.toByteArray();
    }

    private static Timestamp instantToTimestamp(Instant instant) {
        if (instant == null) {
            return Timestamp.getDefaultInstance();
        }
        return Timestamp.newBuilder()
                .setSeconds(instant.getEpochSecond())
                .setNanos(instant.getNano())
                .build();
    }
}

