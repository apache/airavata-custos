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
package org.apache.custos.amie;

import com.fasterxml.jackson.databind.JsonNode;
import org.apache.custos.amie.client.AmieClient;
import org.apache.custos.amie.model.PacketEntity;
import org.apache.custos.amie.model.PacketStatus;
import org.apache.custos.amie.model.ProcessingEventEntity;
import org.apache.custos.amie.model.ProcessingEventType;
import org.apache.custos.amie.model.ProcessingStatus;
import org.apache.custos.amie.repo.PacketRepository;
import org.apache.custos.amie.repo.ProcessingEventRepository;
import org.apache.custos.amie.util.ProtoUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.List;

/**
 * A scheduled service that polls the AMIE API for new packets,
 * persists them to the database, and later for processing.
 */
@Component
public class AmiePoller {

    private static final Logger LOGGER = LoggerFactory.getLogger(AmiePoller.class);

    private final AmieClient client;
    private final PacketRepository packetRepo;
    private final ProcessingEventRepository eventRepo;

    public AmiePoller(AmieClient client, PacketRepository packetRepo, ProcessingEventRepository eventRepo) {
        this.client = client;
        this.packetRepo = packetRepo;
        this.eventRepo = eventRepo;
    }


    @Scheduled(fixedDelayString = "${app.amie.scheduler.poll-delay}")
    @Transactional
    public void pollForPackets() {
        LOGGER.info("Polling for new AMIE packets...");
        List<JsonNode> packets = client.fetchInProgressPackets();

        if (packets.isEmpty()) {
            LOGGER.info("No new packets found.");
            return;
        }

        LOGGER.info("Found {} packets to process.", packets.size());
        for (JsonNode packetNode : packets) {
            try {
                processIndividualPacket(packetNode);
            } catch (Exception e) {
                // If a malformed packet is found
                LOGGER.error("An unexpected error occurred while processing a packet. Raw packet: {}", packetNode.toString(), e);
            }
        }
    }

    /**
     * Processes a single packet from the AMIE API. Only persists new packets.
     *
     * @param packetNode The raw packet
     */
    private void processIndividualPacket(JsonNode packetNode) {
        long amiePacketRecId = packetNode.at("/header/packet_rec_id").asLong(-1);
        String packetType = packetNode.path("type").asText(null);

        if (amiePacketRecId < 0 || packetType == null) {
            LOGGER.warn("Skipping packet with missing or invalid 'packet_rec_id' or 'type'. Packet: {}", packetNode);
            return;
        }

        // Only process if this packetRecId is new
        packetRepo.findByAmieId(amiePacketRecId).ifPresentOrElse(
                (existingPacket) -> LOGGER.debug("Packet with amie_packet_rec_id {} already exists. Skip the processing.", amiePacketRecId),
                () -> {
                    LOGGER.info("Persisting new packet with amie_packet_rec_id {} and type '{}'.", amiePacketRecId, packetType);

                    PacketEntity newPacket = new PacketEntity();
                    newPacket.setAmieId(amiePacketRecId);
                    newPacket.setType(packetType);
                    newPacket.setStatus(PacketStatus.NEW);
                    newPacket.setRawJson(packetNode.toString());
                    newPacket.setReceivedAt(Instant.now());
                    packetRepo.save(newPacket);

                    ProcessingEventEntity decodeEvent = new ProcessingEventEntity();
                    decodeEvent.setPacket(newPacket);
                    decodeEvent.setType(ProcessingEventType.DECODE_PACKET);
                    decodeEvent.setStatus(ProcessingStatus.NEW);

                    byte[] payload = ProtoUtils.createDecodeStartedEvent(decodeEvent.getId(), newPacket.getId(), amiePacketRecId);
                    decodeEvent.setPayload(payload);
                    eventRepo.save(decodeEvent);

                    LOGGER.info("Successfully enqueued DECODE_PACKET event for packet {}.", amiePacketRecId);
                }
        );
    }
}