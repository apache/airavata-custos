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

package org.apache.custos.user.profile.persistance.repository;

import org.apache.custos.user.profile.persistance.model.Group;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Repository;

import jakarta.persistence.EntityManager;
import jakarta.persistence.PersistenceContext;
import jakarta.persistence.Query;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Repository
public class SearchGroupsRepositoryImpl implements SearchGroupsRepository {

    private static final Logger LOGGER = LoggerFactory.getLogger(SearchGroupsRepositoryImpl.class);

    @PersistenceContext
    EntityManager entityManager;

    @Override
    public List<Group> searchEntities(long tenantId, org.apache.custos.user.profile.service.Group group, int offset,
                                      int limit) {
        Map<String, Object> valueMap = new HashMap<>();
        String query = createSQLQuery(tenantId, group, valueMap,  offset,  limit);


        Query q = entityManager.createNativeQuery(query, Group.class);
        for (String key : valueMap.keySet()) {
            q.setParameter(key, valueMap.get(key));
        }

        return q.getResultList();
    }

    private String createSQLQuery(long tenantId, org.apache.custos.user.profile.service.Group group,
                                  Map<String, Object> valueMap, int offset, int limit) {

        String query = "SELECT * FROM group_entity E WHERE ";

        if (!group.getName().isBlank()) {

            query = query + "E.name  LIKE :" + "name" + " AND ";

            valueMap.put("name", group.getName());
        }

        if (!group.getDescription().isBlank()) {

            query = query + "E.description LIKE :" + "description" + " AND ";

            valueMap.put("description", group.getDescription());
        }

        if (!group.getId().isBlank()) {
            query = query + "E.external_id = :" + "external_id" + " AND ";

            valueMap.put("external_id", group.getId());
        }


        if (group.getCreatedTime() != 0) {

            Date date = new Date(group.getCreatedTime());

            query = query + "E.created_at >= :" + "created_at" + " AND ";

            valueMap.put("created_at", date);
        }

        if (group.getLastModifiedTime() != 0) {
            Date date = new Date(group.getLastModifiedTime());
            query = query + "E.last_modified_at >= :" + "last_modified_at" + " AND ";

            valueMap.put("last_modified_at", date);
        }
        if (tenantId != 0) {

            query = query + "E.tenant_id = :" + "tenant_id" + " AND ";

            valueMap.put("tenant_id", tenantId);
        }


        query = query.substring(0, query.length() - 5);

        query = query + " ORDER BY E.created_at DESC";

        if (limit > 0) {
            query = query + " LIMIT " + ":limit" + " OFFSET " + ":offset";
            valueMap.put("limit", limit);
            valueMap.put("offset", offset);
        }

        return query;
    }


}
