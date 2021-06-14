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

package org.apache.custos.tenant.profile.persistance.respository;

import org.apache.custos.tenant.profile.persistance.model.Tenant;
import org.apache.custos.tenant.profile.service.TenantStatus;
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
public class SearchTenantRepositoryImpl implements SearchTenantRepository {

    private static final Logger LOGGER = LoggerFactory.getLogger(SearchTenantRepositoryImpl.class);

    @PersistenceContext
    EntityManager entityManager;

    @Override
    public List<Tenant> searchTenants(String requestEmail, String status, long parentId, int limit, int offset, String type) {
        Map<String, Object> valueMap = new HashMap<>();
        String query = createSQLQuery(requestEmail, status, type, parentId, limit, offset, valueMap);

        Query q = entityManager.createNativeQuery(query, Tenant.class);
        for (String key : valueMap.keySet()) {
            q.setParameter(key, valueMap.get(key));
        }

        return q.getResultList();
    }


    private String createSQLQuery(String requestEmail, String status, String type, long parentId, int limit, int offset,
                                  Map<String, Object> valueMap) {
        String query = "SELECT * FROM tenant E WHERE ";

        if (requestEmail != null && !requestEmail.isEmpty()) {
            query = query + "E.requester_email = :" + "requester_email" + " AND ";
            valueMap.put("requester_email", requestEmail);

        }

        if (status != null && !status.isEmpty()) {
            query = query + "E.status LIKE :" + "status" + " AND ";
            valueMap.put("status", status);

        } else {
            String defaultStatus = "'"+TenantStatus.REQUESTED.name()+ "'"+ "," +
                    "'"+TenantStatus.ACTIVE.name()+"'" + "," + "'"+TenantStatus.DENIED.name()+"'";
            query = query + "E.status IN (" + defaultStatus + ") AND ";
        }

        if (parentId > 0) {
            query = query + "E.parent_id = :" + "parent_id" + " AND ";
            valueMap.put("parent_id", parentId);

        }

        if (type != null && type.equals("ADMIN")) {
            query = query + "E.parent_id = :" + "parent_id" + " AND ";
            valueMap.put("parent_id", 0);
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
