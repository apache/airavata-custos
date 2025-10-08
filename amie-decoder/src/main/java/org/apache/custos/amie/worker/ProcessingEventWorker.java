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
package org.apache.custos.amie.worker;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.amie.handler.PacketRouter;
import org.apache.custos.amie.model.PacketEntity;
import org.apache.custos.amie.model.PacketStatus;
import org.apache.custos.amie.model.ProcessingErrorEntity;
import org.apache.custos.amie.model.ProcessingEventEntity;
import org.apache.custos.amie.model.ProcessingEventType;
import org.apache.custos.amie.model.ProcessingStatus;
import org.apache.custos.amie.repo.PacketRepository;
import org.apache.custos.amie.repo.ProcessingErrorRepository;
import org.apache.custos.amie.repo.ProcessingEventRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.time.Instant;

/**
 * A scheduled worker that fetches for new processing events and executes them.
 * State of an event (NEW -> RUNNING -> SUCCEEDED/FAILED)
 */
@Component
public class ProcessingEventWorker {

    private static final Logger LOGGER = LoggerFactory.getLogger(ProcessingEventWorker.class);
    private static final int MAX_ATTEMPTS = 5;

    private final ProcessingEventRepository eventRepo;
    private final PacketRepository packetRepo;
    private final ProcessingErrorRepository errorRepo;
    private final PacketRouter router;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public ProcessingEventWorker(ProcessingEventRepository eventRepo, PacketRepository packetRepo,
                                 ProcessingErrorRepository errorRepo, PacketRouter router) {
        this.eventRepo = eventRepo;
        this.packetRepo = packetRepo;
        this.errorRepo = errorRepo;
        this.router = router;
    }

    /**
     * Runs on a fixed delay, checks for NEW events, and processes them one by one.
     */
    @Scheduled(fixedDelayString = "#{T(org.springframework.boot.convert.DurationStyle).detectAndParse('${app.amie.scheduler.worker-delay}').toMillis()}")
    @Transactional
    public void processPendingEvents() {
        try (var eventStream = eventRepo.findTop50ByStatusOrderByCreatedAtAsc(ProcessingStatus.NEW)) {
            eventStream.forEach(this::executeEvent);
        }
    }

    private void executeEvent(ProcessingEventEntity event) {
        PacketEntity packet = event.getPacket();
        LOGGER.info("Processing event [{}] for packet amie_id [{}].", event.getType(), packet.getAmieId());

        event.setStatus(ProcessingStatus.RUNNING);
        event.setStartedAt(Instant.now());
        event.setAttempts(event.getAttempts() + 1);
        eventRepo.save(event);

        try {
            var packetJson = objectMapper.readTree(packet.getRawJson());
            router.route(packetJson, packet);

            handleSuccess(event, packet);

        } catch (Exception e) {
            LOGGER.error("Event processing failed for packet amie_id [{}]. Attempt {} of {}.", packet.getAmieId(), event.getAttempts(), MAX_ATTEMPTS, e);
            handleFailure(event, packet, e);
        }
    }

    private void handleSuccess(ProcessingEventEntity event, PacketEntity packet) {
        event.setStatus(ProcessingStatus.SUCCEEDED);
        event.setFinishedAt(Instant.now());
        eventRepo.save(event);

        if (event.getType() == ProcessingEventType.DECODE_PACKET) {
            packet.setStatus(PacketStatus.DECODED);
            packet.setDecodedAt(Instant.now());
            packetRepo.save(packet);
        }

        LOGGER.info("Successfully processed event [{}] for packet amie_id [{}].", event.getType(), packet.getAmieId());
    }

    private void handleFailure(ProcessingEventEntity event, PacketEntity packet, Exception e) {
        // Check if the event should be retried or marked as failed
        boolean isRetryable = event.getAttempts() < MAX_ATTEMPTS;
        ProcessingStatus newStatus = isRetryable ? ProcessingStatus.NEW : ProcessingStatus.FAILED;

        event.setStatus(newStatus);
        event.setLastError(e.getMessage());
        event.setFinishedAt(Instant.now());
        eventRepo.save(event);

        packet.setStatus(PacketStatus.FAILED);
        packet.setLastError(e.getMessage());
        packetRepo.save(packet);

        ProcessingErrorEntity error = new ProcessingErrorEntity();
        error.setPacket(packet);
        error.setEvent(event);
        error.setSummary(e.getClass().getSimpleName() + ": " + e.getMessage());
        error.setDetail(getStackTraceAsString(e));
        errorRepo.save(error);
    }

    private String getStackTraceAsString(Exception e) {
        StringWriter sw = new StringWriter();
        PrintWriter pw = new PrintWriter(sw);
        e.printStackTrace(pw);
        return sw.toString();
    }
}
