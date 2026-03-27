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
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Unit tests for ProcessingEventWorker.
 */
@ExtendWith(MockitoExtension.class)
@Tag("unit")
class ProcessingEventWorkerTest {

    @Mock
    private ProcessingEventRepository eventRepo;

    @Mock
    private PacketRepository packetRepo;

    @Mock
    private ProcessingErrorRepository errorRepo;

    @Mock
    private PacketRouter router;

    @Mock
    private ProcessingEventWorker self;

    private ProcessingEventWorker worker;

    @BeforeEach
    void setUp() {
        worker = new ProcessingEventWorker(eventRepo, packetRepo, errorRepo, router, self);
    }

    // ------------------------------------------------------------------
    // processPendingEvents — orchestration
    // ------------------------------------------------------------------

    @Test
    void processPendingEvents_whenNoEvents_doesNothing() throws Exception {
        when(eventRepo.findTop50EventsToProcess(anyList(), any(Instant.class))).thenReturn(List.of());

        worker.processPendingEvents();

        verify(self, never()).executeEventInTransaction(any());
        verify(self, never()).recordFailureInNewTransaction(any(), any());
    }

    @Test
    void processPendingEvents_callsExecuteForEachEvent() throws Exception {
        ProcessingEventEntity event1 = buildEvent("id-1", ProcessingStatus.NEW, 0);
        ProcessingEventEntity event2 = buildEvent("id-2", ProcessingStatus.RETRY_SCHEDULED, 1);
        when(eventRepo.findTop50EventsToProcess(anyList(), any(Instant.class)))
                .thenReturn(List.of(event1, event2));

        worker.processPendingEvents();

        verify(self).executeEventInTransaction(event1);
        verify(self).executeEventInTransaction(event2);
        verify(self, never()).recordFailureInNewTransaction(any(), any());
    }

    @Test
    void processPendingEvents_whenExecuteThrows_callsRecoveryTransactionWithEventId() throws Exception {
        ProcessingEventEntity event = buildEvent("id-thrown", ProcessingStatus.NEW, 0);
        when(eventRepo.findTop50EventsToProcess(anyList(), any(Instant.class)))
                .thenReturn(List.of(event));

        RuntimeException cause = new RuntimeException("handler blew up");
        doThrow(cause).when(self).executeEventInTransaction(event);

        worker.processPendingEvents();

        verify(self).recordFailureInNewTransaction("id-thrown", cause);
    }

    @Test
    void processPendingEvents_whenBothExecuteAndRecoveryThrow_continuesProcessingRemainingEvents() throws Exception {
        ProcessingEventEntity badEvent = buildEvent("bad-id", ProcessingStatus.NEW, 0);
        ProcessingEventEntity goodEvent = buildEvent("good-id", ProcessingStatus.NEW, 0);
        when(eventRepo.findTop50EventsToProcess(anyList(), any(Instant.class)))
                .thenReturn(List.of(badEvent, goodEvent));

        RuntimeException executeCause = new RuntimeException("execute failed");
        doThrow(executeCause).when(self).executeEventInTransaction(badEvent);
        doThrow(new RuntimeException("recovery also failed"))
                .when(self).recordFailureInNewTransaction("bad-id", executeCause);

        // Should not propagate — the worker logs and continues to goodEvent.
        worker.processPendingEvents();

        verify(self).executeEventInTransaction(goodEvent);
    }

    // ------------------------------------------------------------------
    // recordFailureInNewTransaction — retryable failures
    // ------------------------------------------------------------------

    @Test
    void recordFailureInNewTransaction_onFirstFailure_setsRetryScheduledWithBackoff() {
        ProcessingEventEntity event = buildEvent("evt-1", ProcessingStatus.NEW, 0);
        PacketEntity packet = event.getPacket();
        when(eventRepo.findById("evt-1")).thenReturn(Optional.of(event));

        RuntimeException cause = new IllegalArgumentException("Project not found");
        Instant before = Instant.now();

        worker.recordFailureInNewTransaction("evt-1", cause);

        Instant after = Instant.now();

        assertThat(event.getStatus()).isEqualTo(ProcessingStatus.RETRY_SCHEDULED);
        assertThat(event.getAttempts()).isEqualTo(1); // incremented from 0
        assertThat(event.getLastError()).isEqualTo("Project not found");
        assertThat(event.getNextRetryAt()).isNotNull();
        // Backoff for attempt 1 = 30 s; nextRetryAt must be in the near future
        assertThat(event.getNextRetryAt()).isAfterOrEqualTo(before.plusSeconds(29));
        assertThat(event.getNextRetryAt()).isBeforeOrEqualTo(after.plusSeconds(31));

        verify(eventRepo).save(event);
        verify(packetRepo, never()).save(any());

        ArgumentCaptor<ProcessingErrorEntity> errorCaptor = ArgumentCaptor.forClass(ProcessingErrorEntity.class);
        verify(errorRepo).save(errorCaptor.capture());
        ProcessingErrorEntity error = errorCaptor.getValue();
        assertThat(error.getSummary()).contains("IllegalArgumentException");
        assertThat(error.getSummary()).contains("Project not found");
        assertThat(error.getPacket()).isSameAs(packet);
        assertThat(error.getEvent()).isSameAs(event);
    }

    @Test
    void recordFailureInNewTransaction_onSecondFailure_setsRetryScheduledWithDoubledBackoff() {
        // Simulate: the first attempt rolled back (REQUIRES_NEW), so the event is
        // re-read with RETRY_SCHEDULED status and attempts=1 (the previous committed
        // retry count). The recovery method unconditionally increments attempts to 2
        // and applies a 60-second backoff (BASE * 2^1).
        ProcessingEventEntity event = buildEvent("evt-2", ProcessingStatus.RETRY_SCHEDULED, 1);
        when(eventRepo.findById("evt-2")).thenReturn(Optional.of(event));

        Instant before = Instant.now();
        worker.recordFailureInNewTransaction("evt-2", new RuntimeException("second failure"));
        Instant after = Instant.now();

        assertThat(event.getStatus()).isEqualTo(ProcessingStatus.RETRY_SCHEDULED);
        // Attempts unconditionally incremented from 1 to 2
        assertThat(event.getAttempts()).isEqualTo(2);
        // Backoff for attempt 2 = 60 s
        assertThat(event.getNextRetryAt()).isAfterOrEqualTo(before.plusSeconds(59));
        assertThat(event.getNextRetryAt()).isBeforeOrEqualTo(after.plusSeconds(61));
    }

    // ------------------------------------------------------------------
    // recordFailureInNewTransaction — permanent failure
    // ------------------------------------------------------------------

    @Test
    void recordFailureInNewTransaction_afterMaxAttempts_marksPermanentlyFailed() {
        // Simulate event on its last allowed retry (attempts already at MAX-1 in RETRY_SCHEDULED)
        // When we increment, it reaches MAX_ATTEMPTS = 3
        ProcessingEventEntity event = buildEvent("evt-perm", ProcessingStatus.RETRY_SCHEDULED, 2);
        PacketEntity packet = event.getPacket();
        when(eventRepo.findById("evt-perm")).thenReturn(Optional.of(event));

        RuntimeException cause = new RuntimeException("final failure");
        worker.recordFailureInNewTransaction("evt-perm", cause);

        assertThat(event.getStatus()).isEqualTo(ProcessingStatus.PERMANENTLY_FAILED);
        assertThat(event.getAttempts()).isEqualTo(3);
        assertThat(event.getNextRetryAt()).isNull();

        // Packet must also be marked FAILED
        verify(packetRepo).save(packet);
        assertThat(packet.getStatus()).isEqualTo(PacketStatus.FAILED);
        assertThat(packet.getLastError()).isEqualTo("final failure");

        verify(eventRepo).save(event);
        verify(errorRepo).save(any(ProcessingErrorEntity.class));
    }

    @Test
    void recordFailureInNewTransaction_whenEventNotFound_doesNotThrow() {
        when(eventRepo.findById("missing")).thenReturn(Optional.empty());

        // Should log and return without throwing or saving anything
        worker.recordFailureInNewTransaction("missing", new RuntimeException("cause"));

        verify(eventRepo, never()).save(any());
        verify(errorRepo, never()).save(any());
    }

    // ------------------------------------------------------------------
    // computeNextRetryAt — exponential backoff
    // ------------------------------------------------------------------

    @Test
    void computeNextRetryAt_attempt1_returns30SecondDelay() {
        Instant before = Instant.now();
        Instant result = ProcessingEventWorker.computeNextRetryAt(1);
        Instant after = Instant.now();

        assertThat(result).isAfterOrEqualTo(before.plusSeconds(29));
        assertThat(result).isBeforeOrEqualTo(after.plusSeconds(31));
    }

    @Test
    void computeNextRetryAt_attempt2_returns60SecondDelay() {
        Instant before = Instant.now();
        Instant result = ProcessingEventWorker.computeNextRetryAt(2);
        Instant after = Instant.now();

        assertThat(result).isAfterOrEqualTo(before.plusSeconds(59));
        assertThat(result).isBeforeOrEqualTo(after.plusSeconds(61));
    }

    @Test
    void computeNextRetryAt_highAttemptNumber_capsAtMaxBackoff() {
        // At attempt 100, delay would be astronomically large without capping.
        // MAX_BACKOFF_SECONDS = 600
        Instant before = Instant.now();
        Instant result = ProcessingEventWorker.computeNextRetryAt(100);
        Instant after = Instant.now();

        // Must not exceed MAX_BACKOFF_SECONDS + small tolerance
        assertThat(result).isBeforeOrEqualTo(after.plusSeconds(601));
        // Must still be at least MAX_BACKOFF_SECONDS in the future
        assertThat(result).isAfterOrEqualTo(before.plusSeconds(599));
    }

    private ProcessingEventEntity buildEvent(String id, ProcessingStatus status, int attempts) {
        PacketEntity packet = new PacketEntity();
        packet.setAmieId(42L);
        packet.setType("request_account_create");
        packet.setStatus(PacketStatus.NEW);
        // Provide minimal valid JSON so objectMapper.readTree won't NPE if called directly
        packet.setRawJson("{\"body\":{}}");

        ProcessingEventEntity event = new ProcessingEventEntity();
        // Directly set the id field via reflection to avoid UUID randomness issues in tests
        try {
            java.lang.reflect.Field idField = ProcessingEventEntity.class.getDeclaredField("id");
            idField.setAccessible(true);
            idField.set(event, id);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
        event.setPacket(packet);
        event.setType(ProcessingEventType.DECODE_PACKET);
        event.setStatus(status);
        event.setAttempts(attempts);
        event.setPayload(new byte[0]);
        return event;
    }
}
