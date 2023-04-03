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

package org.apache.custos.logging.mapper;

import org.apache.custos.logging.service.LogEvent;

public class LogEventMapper {


    public static LogEvent transform(org.apache.custos.logging.persistance.model.LogEvent logEvent) {

        LogEvent event = LogEvent
                .newBuilder()
                .setTenantId(logEvent.getTenantId())
                .setClientId(logEvent.getClientId())
                .setCreatedTime(logEvent.getCreatedTime())
                .setServiceName(logEvent.getServiceName())
                .setEventType(logEvent.getEventType())
                .setExternalIp(logEvent.getRemoteIp())
                .setUsername(logEvent.getUsername() != null ? logEvent.getUsername() : "")
                .build();

        return event;
    }

    public static org.apache.custos.logging.persistance.model.LogEvent transform(LogEvent logEvent) {

        org.apache.custos.logging.persistance.model.LogEvent event = new
                org.apache.custos.logging.persistance.model.LogEvent();

        event.setClientId(logEvent.getClientId());
        event.setTenantId(logEvent.getTenantId());
        event.setCreatedTime(logEvent.getCreatedTime());
        event.setRemoteIp(logEvent.getExternalIp());
        event.setEventType(logEvent.getEventType());
        event.setServiceName(logEvent.getServiceName());
        event.setUsername(logEvent.getUsername());
        return event;

    }

}
