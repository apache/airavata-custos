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
package org.apache.custos.access.ci.service.worker.amie;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.access.ci.service.handler.amie.PacketRouter;
import org.apache.custos.access.ci.service.model.amie.PacketEntity;
import org.apache.custos.access.ci.service.model.amie.PacketStatus;
import org.apache.custos.access.ci.service.model.amie.ProcessingErrorEntity;
import org.apache.custos.access.ci.service.model.amie.ProcessingEventEntity;
import org.apache.custos.access.ci.service.model.amie.ProcessingEventType;
import org.apache.custos.access.ci.service.model.amie.ProcessingStatus;
import org.apache.custos.access.ci.service.repo.amie.PacketRepository;
import org.apache.custos.access.ci.service.repo.amie.ProcessingErrorRepository;
import org.apache.custos.access.ci.service.repo.amie.ProcessingEventRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Lazy;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.time.Instant;
import java.util.List;

/**
 * A scheduled worker that fetches for new/ processing events and executes them.
 * State of an event (NEW -> RUNNING -> SUCCEEDED/FAILED)
 */
@Component
public class ProcessingEventWorker {

    private static final Logger LOGGER = LoggerFactory.getLogger(ProcessingEventWorker.class);
    private static final int MAX_ATTEMPTS = 3;

    private final ProcessingEventRepository eventRepo;
    private final PacketRepository packetRepo;
    private final ProcessingErrorRepository errorRepo;
    private final PacketRouter router;
    private final ObjectMapper objectMapper = new ObjectMapper();
    private final ProcessingEventWorker self;

    public ProcessingEventWorker(ProcessingEventRepository eventRepo, PacketRepository packetRepo,
                                 ProcessingErrorRepository errorRepo, PacketRouter router, @Lazy ProcessingEventWorker self) {
        this.eventRepo = eventRepo;
        this.packetRepo = packetRepo;
        this.errorRepo = errorRepo;
        this.router = router;
        this.self = self;
    }

    /**
     * Runs on a fixed delay, checks for NEW/RETRY_SCHEDULED events, and processes them one by one on a separate transaction.
     */
    @Scheduled(fixedDelayString = "#{T(org.springframework.boot.convert.DurationStyle).detectAndParse('${access.amie.scheduler.worker-delay}').toMillis()}")
    public void processPendingEvents() {
        List<ProcessingEventEntity> eventsToProcess = eventRepo.findTop50EventsToProcess(List.of(ProcessingStatus.NEW, ProcessingStatus.RETRY_SCHEDULED));

        if (!eventsToProcess.isEmpty()) {
            LOGGER.info("Found {} event(s) to process.", eventsToProcess.size());
            eventsToProcess.forEach(event -> {
                try {
                    self.executeEventInTransaction(event);
                } catch (Exception e) {
                    LOGGER.error("An unexpected error occurred while processing of eventId [{}].", event.getId(), e);
                }
            });
        }
    }

    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void executeEventInTransaction(ProcessingEventEntity event) {
        PacketEntity packet = event.getPacket();
        LOGGER.info("Processing event [{}] for packet amie_id [{}]. Attempt: {}", event.getType(), packet.getAmieId(), event.getAttempts() + 1);

        event.setStatus(ProcessingStatus.RUNNING);
        event.setStartedAt(Instant.now());
        event.setAttempts(event.getAttempts() + 1);
        eventRepo.saveAndFlush(event);

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
        ProcessingStatus newStatus = isRetryable ? ProcessingStatus.RETRY_SCHEDULED : ProcessingStatus.FAILED;

        event.setStatus(newStatus);
        event.setLastError(e.getMessage());
        event.setFinishedAt(Instant.now());
        eventRepo.save(event);

        if (!isRetryable) {
            LOGGER.error("Event for packet amie_id [{}] has failed permanently after {} attempts.", packet.getAmieId(), event.getAttempts());
            packet.setStatus(PacketStatus.FAILED);
            packet.setLastError(e.getMessage());
            packetRepo.save(packet);

        } else {
            LOGGER.warn("Event for packet amie_id [{}] will be retried. Status set to {}.", packet.getAmieId(), newStatus);
        }

        ProcessingErrorEntity error = new ProcessingErrorEntity();
        error.setPacket(packet);
        error.setEvent(event);
        error.setSummary(e.getClass().getSimpleName() + ": " + e.getMessage());
        error.setDetail(getStackTraceAsString(e));
        errorRepo.save(error);
    }

    private String getStackTraceAsString(Exception e) {
        StringWriter sw = new StringWriter();
        e.printStackTrace(new PrintWriter(sw));
        return sw.toString();
    }
}
