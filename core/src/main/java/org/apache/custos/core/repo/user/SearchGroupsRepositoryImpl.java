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

package org.apache.custos.core.repo.user;

import org.apache.custos.core.model.user.Group;
import jakarta.persistence.EntityManager;
import jakarta.persistence.PersistenceContext;
import jakarta.persistence.Query;
import org.apache.commons.lang3.StringUtils;
import org.springframework.stereotype.Repository;

import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Repository
public class SearchGroupsRepositoryImpl implements SearchGroupsRepository {

    @PersistenceContext
    EntityManager entityManager;

    @Override
    public List<Group> searchEntities(long tenantId, org.apache.custos.core.user.profile.api.Group group, int offset, int limit) {
        Map<String, Object> valueMap = new HashMap<>();
        String query = createSQLQuery(tenantId, group, valueMap, offset, limit);

        Query q = entityManager.createNativeQuery(query, Group.class);
        valueMap.forEach(q::setParameter);

        return q.getResultList();
    }

    private String createSQLQuery(long tenantId, org.apache.custos.core.user.profile.api.Group group, Map<String, Object> valueMap, int offset, int limit) {
        String query = "SELECT * FROM group_entity E WHERE ";

        if (!group.getName().isBlank()) {
            query = query + "E.name LIKE :" + "name" + " AND ";
            valueMap.put("name", group.getName());
        }

        if (StringUtils.isNotBlank(group.getDescription())) {
            query = query + "E.description LIKE :" + "description" + " AND ";
            valueMap.put("description", group.getDescription());
        }

        if (StringUtils.isNotBlank(group.getId())) {
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
