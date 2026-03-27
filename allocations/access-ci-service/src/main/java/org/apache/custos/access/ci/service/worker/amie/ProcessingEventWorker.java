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
import io.micrometer.core.instrument.Timer;
import org.apache.custos.access.ci.service.handler.amie.PacketRouter;
import org.apache.custos.access.ci.service.metrics.AmieMetrics;
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
import org.slf4j.MDC;
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
 * Scheduled worker that polls for pending AMIE processing events and executes them.
 * Failures are recorded in a separate transaction to prevent infinite retry loops.
 *
 * <pre>
 * NEW → RUNNING → SUCCEEDED
 *                → RETRY_SCHEDULED → (backoff) → RUNNING → ...
 *                → PERMANENTLY_FAILED (after MAX_ATTEMPTS)
 * </pre>
 */
@Component
public class ProcessingEventWorker {

    private static final Logger LOGGER = LoggerFactory.getLogger(ProcessingEventWorker.class);

    /**
     * Maximum number of execution attempts before an event is permanently failed.
     */
    static final int MAX_ATTEMPTS = 3;

    /**
     * Base delay in seconds for exponential backoff between retry attempts.
     */
    private static final long BASE_BACKOFF_SECONDS = 30L;

    /**
     * Upper bound on the computed backoff delay (10 minutes).
     */
    private static final long MAX_BACKOFF_SECONDS = 600L;

    private final ProcessingEventRepository eventRepo;
    private final PacketRepository packetRepo;
    private final ProcessingErrorRepository errorRepo;
    private final PacketRouter router;
    private final AmieMetrics amieMetrics;
    private final ObjectMapper objectMapper = new ObjectMapper();
    private final ProcessingEventWorker self;

    public ProcessingEventWorker(ProcessingEventRepository eventRepo,
                                 PacketRepository packetRepo,
                                 ProcessingErrorRepository errorRepo,
                                 PacketRouter router,
                                 AmieMetrics amieMetrics,
                                 @Lazy ProcessingEventWorker self) {
        this.eventRepo = eventRepo;
        this.packetRepo = packetRepo;
        this.errorRepo = errorRepo;
        this.router = router;
        this.amieMetrics = amieMetrics;
        this.self = self;
    }

    /**
     * Polls for due events and processes each in its own transaction.
     */
    @Scheduled(fixedDelayString = "#{T(org.springframework.boot.convert.DurationStyle).detectAndParse('${access.amie.scheduler.worker-delay}').toMillis()}")
    public void processPendingEvents() {
        List<ProcessingEventEntity> eventsToProcess =
                eventRepo.findTop50EventsToProcess(
                        List.of(ProcessingStatus.NEW, ProcessingStatus.RETRY_SCHEDULED),
                        Instant.now());

        if (eventsToProcess.isEmpty()) {
            return;
        }

        LOGGER.info("Found {} event(s) to process.", eventsToProcess.size());

        for (ProcessingEventEntity event : eventsToProcess) {
            String eventId = event.getId();
            PacketEntity packet = event.getPacket();

            MDC.put("packetId", packet.getId());
            MDC.put("amieId", String.valueOf(packet.getAmieId()));
            MDC.put("packetType", packet.getType());

            Timer.Sample timerSample = amieMetrics.startProcessingTimer();
            try {
                self.executeEventInTransaction(event);
            } catch (Exception e) {
                LOGGER.error("Transaction failed for eventId [{}]. Opening recovery transaction to record failure.",
                        eventId, e);
                amieMetrics.stopProcessingTimer(timerSample, packet.getType());
                try {
                    self.recordFailureInNewTransaction(eventId, e);
                } catch (Exception recoveryEx) {
                    LOGGER.error("CRITICAL: Recovery transaction also failed for eventId [{}]. " +
                            "Event may remain stuck until the next worker cycle.", eventId, recoveryEx);
                }
                continue;
            } finally {
                MDC.remove("packetId");
                MDC.remove("amieId");
                MDC.remove("packetType");
                MDC.remove("handler");
            }
            amieMetrics.stopProcessingTimer(timerSample, packet.getType());
        }
    }

    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void executeEventInTransaction(ProcessingEventEntity event) throws Exception {
        PacketEntity packet = event.getPacket();
        LOGGER.info("Processing event [{}] for packet amie_id [{}]. Attempt: {}",
                event.getType(), packet.getAmieId(), event.getAttempts() + 1);

        event.setStatus(ProcessingStatus.RUNNING);
        event.setStartedAt(Instant.now());
        event.setAttempts(event.getAttempts() + 1);
        eventRepo.saveAndFlush(event);

        var packetJson = objectMapper.readTree(packet.getRawJson());
        router.route(packetJson, packet, event.getId());

        handleSuccess(event, packet);
    }

    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void recordFailureInNewTransaction(String eventId, Exception cause) {
        ProcessingEventEntity event = eventRepo.findById(eventId).orElse(null);
        if (event == null) {
            LOGGER.error("Cannot record failure: event [{}] not found in the database.", eventId);
            return;
        }

        PacketEntity packet = event.getPacket();

        int effectiveAttempts = event.getAttempts() + 1;
        event.setAttempts(effectiveAttempts);

        boolean isRetryable = effectiveAttempts < MAX_ATTEMPTS;
        ProcessingStatus newStatus = isRetryable
                ? ProcessingStatus.RETRY_SCHEDULED
                : ProcessingStatus.PERMANENTLY_FAILED;

        event.setStatus(newStatus);
        event.setLastError(cause.getMessage());
        event.setFinishedAt(Instant.now());

        if (isRetryable) {
            Instant nextRetryAt = computeNextRetryAt(effectiveAttempts);
            event.setNextRetryAt(nextRetryAt);
            amieMetrics.recordRetry();
            amieMetrics.recordPacketProcessed(packet.getType(), "retry_scheduled");
            LOGGER.warn("Event [{}] for packet amie_id [{}] failed on attempt {}/{}. Scheduled for retry after [{}].",
                    eventId, packet.getAmieId(), effectiveAttempts, MAX_ATTEMPTS, nextRetryAt);
        } else {
            event.setNextRetryAt(null);
            amieMetrics.recordPacketProcessed(packet.getType(), "permanently_failed");
            LOGGER.error("Event [{}] for packet amie_id [{}] is PERMANENTLY_FAILED after {} attempt(s). Manual intervention required.",
                    eventId, packet.getAmieId(), effectiveAttempts);
            packet.setStatus(PacketStatus.FAILED);
            packet.setLastError(cause.getMessage());
            packetRepo.save(packet);
        }

        eventRepo.save(event);

        ProcessingErrorEntity error = new ProcessingErrorEntity();
        error.setPacket(packet);
        error.setEvent(event);
        error.setSummary(cause.getClass().getSimpleName() + ": " + cause.getMessage());
        String stackTrace = getStackTraceAsString(cause);
        if (stackTrace.length() > 8000) {
            stackTrace = stackTrace.substring(0, 8000) + "\n... [truncated]";
        }
        error.setDetail(stackTrace);
        errorRepo.save(error);
    }

    private void handleSuccess(ProcessingEventEntity event, PacketEntity packet) {
        event.setStatus(ProcessingStatus.SUCCEEDED);
        event.setFinishedAt(Instant.now());
        event.setNextRetryAt(null);
        eventRepo.save(event);

        if (event.getType() == ProcessingEventType.DECODE_PACKET) {
            packet.setStatus(PacketStatus.DECODED);
            packet.setDecodedAt(Instant.now());
            packetRepo.save(packet);
        }

        amieMetrics.recordPacketProcessed(packet.getType(), "succeeded");

        LOGGER.info("Successfully processed event [{}] for packet amie_id [{}].",
                event.getType(), packet.getAmieId());
    }

    // Exponential backoff: BASE * 2^(attempt-1), capped at MAX_BACKOFF_SECONDS.
    static Instant computeNextRetryAt(int attemptNumber) {
        long exponent = Math.max(0, attemptNumber - 1);
        long delaySec = BASE_BACKOFF_SECONDS * (1L << exponent);
        delaySec = Math.min(delaySec, MAX_BACKOFF_SECONDS);
        return Instant.now().plusSeconds(delaySec);
    }

    private String getStackTraceAsString(Exception e) {
        StringWriter sw = new StringWriter();
        e.printStackTrace(new PrintWriter(sw));
        return sw.toString();
    }
}
