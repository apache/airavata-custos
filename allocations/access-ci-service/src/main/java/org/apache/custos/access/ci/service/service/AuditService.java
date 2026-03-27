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
import org.apache.custos.access.ci.service.repo.amie.AuditLogRepository;
import org.apache.custos.access.ci.service.repo.amie.PacketRepository;
import org.apache.custos.access.ci.service.repo.amie.ProcessingEventRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

/**
 * Service for recording audit log entries during AMIE packet processing.
 *
 * <p>This service is not annotated with {@code @Transactional},
 * so that this be in the caller's active transaction without introducing a new one.
 * There must be an active transaction when invoking {@link #log}.
 */
@Service
public class AuditService {

    private static final Logger LOGGER = LoggerFactory.getLogger(AuditService.class);

    private final AuditLogRepository auditLogRepository;
    private final PacketRepository packetRepository;
    private final ProcessingEventRepository processingEventRepository;

    public AuditService(AuditLogRepository auditLogRepository,
                        PacketRepository packetRepository,
                        ProcessingEventRepository processingEventRepository) {
        this.auditLogRepository = auditLogRepository;
        this.packetRepository = packetRepository;
        this.processingEventRepository = processingEventRepository;
    }

    /**
     * Records an audit log entry for an AMIE packet processing action.
     *
     * @param packetId   the ID of the packet being processed (not null)
     * @param eventId    the ID of the processing event, or null if not associated with an event
     * @param action     the auditable action performed
     * @param entityType a label describing the type of entity affected (e.g., "Person", "Project")
     * @param entityId   the ID of the affected entity, or null if not applicable
     * @param summary    a human-readable description of what happened, or null
     */
    public void log(String packetId,
                    String eventId,
                    AuditAction action,
                    String entityType,
                    String entityId,
                    String summary) {

        AuditLogEntity entry = new AuditLogEntity();
        entry.setPacket(packetRepository.getReferenceById(packetId));
        if (eventId != null) {
            entry.setEvent(processingEventRepository.getReferenceById(eventId));
        }
        entry.setAction(action);
        entry.setEntityType(entityType);
        entry.setEntityId(entityId);
        entry.setSummary(summary);

        auditLogRepository.save(entry);

        LOGGER.debug("Audit entry recorded: packetId={}, eventId={}, action={}, entityType={}, entityId={}",
                packetId, eventId, action, entityType, entityId);
    }
}
