/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.core.commons;

import org.apache.custos.core.model.commons.OperationStatus;
import org.apache.custos.core.model.commons.StatusEntity;
import org.apache.custos.core.repo.commons.StatusUpdaterRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.List;


@Component
public class StatusUpdater {

    private final Logger LOGGER = LoggerFactory.getLogger(StatusUpdater.class);

    private final StatusUpdaterRepository repository;


    public StatusUpdater(StatusUpdaterRepository repository) {
        this.repository = repository;
    }

    public void updateStatus(String method, OperationStatus status, long traceId, String performedBy) {
        try {
            StatusEntity statusEntity = new StatusEntity();
            statusEntity.setEvent(method);
            statusEntity.setState(status.name());
            statusEntity.setTraceId(traceId);
            statusEntity.setPerformedBy(performedBy);
            repository.save(statusEntity);

        } catch (Exception ex) {
            LOGGER.error("Status update failed for event " + method + " and traceId " + traceId);
        }
    }

    public List<StatusEntity> getOperationStatus(long traceId) {
        return repository.findAllByTraceId(traceId);
    }

}
