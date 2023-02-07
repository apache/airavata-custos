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

package org.apache.custos.logging.persistance.repository;

import org.apache.custos.logging.persistance.model.LogEvent;
import org.apache.custos.logging.service.LogEventRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Repository;

import javax.persistence.EntityManager;
import javax.persistence.PersistenceContext;
import javax.persistence.Query;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Repository
public class LogEventRepoImpl implements LogEventRepo {

    private static final Logger LOGGER = LoggerFactory.getLogger(LogEventRepoImpl.class);

    @PersistenceContext
    EntityManager entityManager;


    @Override
    public List<LogEvent> searchEvents(LogEventRequest request) {
        Map<String, Object> valueMap = new HashMap<>();

        String query = createSQLQuery(request, valueMap);

        Query q = entityManager.createNativeQuery(query, LogEvent.class);
        for (String key : valueMap.keySet()) {
            q.setParameter(key, valueMap.get(key));
        }

        return q.getResultList();
    }


    private String createSQLQuery(LogEventRequest logEventRequest, Map<String, Object> valueMap) {

        String query = "SELECT * FROM log_events E WHERE ";


        if (logEventRequest.getTenantId() != 0) {

            query = query + "E.tenant_id = :" + "tenant_id" + " AND ";

            valueMap.put("tenant_id", logEventRequest.getTenantId());

        }

        if (logEventRequest.getClientId() != null && !logEventRequest.getClientId().equals("")) {

            query = query + "E.client_id = :" + "client_id" + " AND ";

            valueMap.put("client_id", logEventRequest.getClientId());

        }

        if (logEventRequest.getServiceName() != null && !logEventRequest.getServiceName().equals("")) {

            query = query + "E.service_name = :" + "service_name" + " AND ";

            valueMap.put("service_name", logEventRequest.getServiceName());

        }

        if (logEventRequest.getEventType() != null && !logEventRequest.getEventType().equals("")) {

            query = query + "E.event_type = :" + "event_type" + " AND ";

            valueMap.put("event_type", logEventRequest.getEventType());

        }

        if (logEventRequest.getRemoteIp() != null && !logEventRequest.getRemoteIp().equals("")) {

            query = query + "E.remote_ip = :" + "remote_ip" + " AND ";

            valueMap.put("remote_ip", logEventRequest.getEventType());

        }

        if (logEventRequest.getUsername() != null && !logEventRequest.getUsername().equals("")) {

            query = query + "E.username = :" + "username" + " AND ";

            valueMap.put("username", logEventRequest.getEventType());

        }

        if (logEventRequest.getStartTime() > 0) {

            query = query + "E.created_time >= :" + "created_time" + " AND ";

            valueMap.put("created_time", logEventRequest.getStartTime());

        }

        if (logEventRequest.getEndTime() > 0) {

            query = query + "E.created_time < :" + "created_time" + " AND ";

            valueMap.put("created_time", logEventRequest.getEndTime());

        }

        query = query.substring(0, query.length() - 5);

        query = query + " ORDER BY E.created_time DESC ";

        if (logEventRequest.getOffset() >= 0 && logEventRequest.getLimit() > 0) {
            query = query + " LIMIT " + ":limit" + " OFFSET " + ":offset";
            valueMap.put("limit", logEventRequest.getLimit());
            valueMap.put("offset", logEventRequest.getOffset());
        }

        LOGGER.debug("Query ####" + query);


        return query;

    }
}
