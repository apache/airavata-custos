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
package org.apache.custos.access.ci.service.service;

import org.apache.custos.access.ci.service.model.amie.AuditAction;
import org.apache.custos.access.ci.service.model.amie.AuditLogEntity;
import org.apache.custos.access.ci.service.model.amie.PacketEntity;
import org.apache.custos.access.ci.service.model.amie.ProcessingEventEntity;
import org.apache.custos.access.ci.service.repo.amie.AuditLogRepository;
import org.apache.custos.access.ci.service.repo.amie.PacketRepository;
import org.apache.custos.access.ci.service.repo.amie.ProcessingEventRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class AuditServiceTest {

    @Mock
    private AuditLogRepository auditLogRepository;

    @Mock
    private PacketRepository packetRepository;

    @Mock
    private ProcessingEventRepository processingEventRepository;

    private AuditService auditService;

    @BeforeEach
    void setUp() {
        auditService = new AuditService(auditLogRepository, packetRepository, processingEventRepository);
    }

    @Test
    void log_withEventId_shouldCreateAuditEntryWithAllFields() {
        String packetId = "packet-001";
        String eventId = "event-abc";
        PacketEntity packetProxy = new PacketEntity();
        ProcessingEventEntity eventProxy = new ProcessingEventEntity();

        when(packetRepository.getReferenceById(packetId)).thenReturn(packetProxy);
        when(processingEventRepository.getReferenceById(eventId)).thenReturn(eventProxy);
        when(auditLogRepository.save(any(AuditLogEntity.class))).thenAnswer(inv -> inv.getArgument(0));

        auditService.log(packetId, eventId, AuditAction.CREATE_PERSON, "Person", "person-123", "Created person from packet");

        ArgumentCaptor<AuditLogEntity> captor = ArgumentCaptor.forClass(AuditLogEntity.class);
        verify(auditLogRepository).save(captor.capture());

        AuditLogEntity saved = captor.getValue();
        assertThat(saved.getPacket()).isSameAs(packetProxy);
        assertThat(saved.getEvent()).isSameAs(eventProxy);
        assertThat(saved.getAction()).isEqualTo(AuditAction.CREATE_PERSON);
        assertThat(saved.getEntityType()).isEqualTo("Person");
        assertThat(saved.getEntityId()).isEqualTo("person-123");
        assertThat(saved.getSummary()).isEqualTo("Created person from packet");
        assertThat(saved.getCreatedAt()).isNotNull();
    }

    @Test
    void log_withNullEventId_shouldCreateAuditEntryWithNullEvent() {
        String packetId = "packet-002";
        PacketEntity packetProxy = new PacketEntity();

        when(packetRepository.getReferenceById(packetId)).thenReturn(packetProxy);
        when(auditLogRepository.save(any(AuditLogEntity.class))).thenAnswer(inv -> inv.getArgument(0));

        auditService.log(packetId, null, AuditAction.REPLY_SENT, "Packet", null, "Reply dispatched");

        ArgumentCaptor<AuditLogEntity> captor = ArgumentCaptor.forClass(AuditLogEntity.class);
        verify(auditLogRepository).save(captor.capture());

        AuditLogEntity saved = captor.getValue();
        assertThat(saved.getPacket()).isSameAs(packetProxy);
        assertThat(saved.getEvent()).isNull();
        assertThat(saved.getAction()).isEqualTo(AuditAction.REPLY_SENT);
        assertThat(saved.getEntityType()).isEqualTo("Packet");
        assertThat(saved.getEntityId()).isNull();
        assertThat(saved.getSummary()).isEqualTo("Reply dispatched");
    }

    @Test
    void log_withNullSummaryAndEntityId_shouldSaveEntityWithNulls() {
        String packetId = "packet-003";
        PacketEntity packetProxy = new PacketEntity();

        when(packetRepository.getReferenceById(packetId)).thenReturn(packetProxy);
        when(auditLogRepository.save(any(AuditLogEntity.class))).thenAnswer(inv -> inv.getArgument(0));

        auditService.log(packetId, null, AuditAction.TRANSACTION_COMPLETE, "Transaction", null, null);

        ArgumentCaptor<AuditLogEntity> captor = ArgumentCaptor.forClass(AuditLogEntity.class);
        verify(auditLogRepository).save(captor.capture());

        AuditLogEntity saved = captor.getValue();
        assertThat(saved.getAction()).isEqualTo(AuditAction.TRANSACTION_COMPLETE);
        assertThat(saved.getEntityId()).isNull();
        assertThat(saved.getSummary()).isNull();
    }

    @Test
    void log_usesGetReferenceByIdForPacketProxy() {
        String packetId = "packet-004";
        PacketEntity proxy = new PacketEntity();
        when(packetRepository.getReferenceById(packetId)).thenReturn(proxy);
        when(auditLogRepository.save(any(AuditLogEntity.class))).thenAnswer(inv -> inv.getArgument(0));

        auditService.log(packetId, null, AuditAction.CREATE_PROJECT, "Project", "proj-1", null);

        verify(packetRepository).getReferenceById(packetId);
    }

    @Test
    void log_usesGetReferenceByIdForEventProxy() {
        String packetId = "packet-005";
        String eventId = "event-xyz";
        when(packetRepository.getReferenceById(packetId)).thenReturn(new PacketEntity());
        when(processingEventRepository.getReferenceById(eventId)).thenReturn(new ProcessingEventEntity());
        when(auditLogRepository.save(any(AuditLogEntity.class))).thenAnswer(inv -> inv.getArgument(0));

        auditService.log(packetId, eventId, AuditAction.CREATE_MEMBERSHIP, "Membership", "m-1", null);

        verify(processingEventRepository).getReferenceById(eventId);
    }
}
