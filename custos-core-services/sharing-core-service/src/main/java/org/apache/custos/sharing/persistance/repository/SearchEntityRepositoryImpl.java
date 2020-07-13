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

package org.apache.custos.sharing.persistance.repository;

import org.apache.custos.sharing.persistance.model.Entity;
import org.apache.custos.sharing.service.EntitySearchField;
import org.apache.custos.sharing.service.SearchCondition;
import org.apache.custos.sharing.service.SearchCriteria;
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
public  class SearchEntityRepositoryImpl implements SearchEntityRepository {

    private static final Logger LOGGER = LoggerFactory.getLogger(SearchEntityRepositoryImpl.class);

    @PersistenceContext
    EntityManager entityManager;

    @Override
    public List<Entity> searchEntities(long tenantId, List<SearchCriteria> searchCriteria) {

        Map<String, Object> valueMap = new HashMap<>();
        String query = createSQLQuery(searchCriteria, valueMap);



        Query q = entityManager.createNativeQuery(query, Entity.class);
        for (String key : valueMap.keySet()) {
            q.setParameter(key, valueMap.get(key));
        }

        return q.getResultList();

    }


    private String createSQLQuery(List<SearchCriteria> searchCriteriaList, Map<String, Object> valueMap) {

        String query = "SELECT * FROM entity E WHERE ";

        for (SearchCriteria searchCriteria : searchCriteriaList) {

            if (searchCriteria.getSearchField().equals(EntitySearchField.ID)) {
                if (searchCriteria.getCondition().equals(SearchCondition.EQUAL)) {
                    query = query + "E.external_id = :" + EntitySearchField.ID.name() + " AND ";
                } else {
                    query = query + "E.external_id != :" + EntitySearchField.ID.name() + " AND ";
                }
                valueMap.put(EntitySearchField.ID.name(), searchCriteria.getValue());

            } else if (searchCriteria.getSearchField().equals(EntitySearchField.NAME)) {
                if (searchCriteria.getCondition().equals(SearchCondition.EQUAL)) {
                    query = query + "E.name LIKE :" + EntitySearchField.NAME.name() + " AND ";
                } else {
                    query = query + "E.name NOT LIKE :" + EntitySearchField.NAME.name() + " AND ";
                }
                valueMap.put(EntitySearchField.NAME.name(), searchCriteria.getValue());
            } else if (searchCriteria.getSearchField().equals(EntitySearchField.DESCRIPTION)) {
                if (searchCriteria.getCondition().equals(SearchCondition.EQUAL)) {
                    query = query + "E.description LIKE :" + EntitySearchField.DESCRIPTION.name() + " AND ";
                } else {
                    query = query + "E.description NOT LIKE :" + EntitySearchField.DESCRIPTION.name() + " AND ";
                }
                valueMap.put(EntitySearchField.DESCRIPTION.name(), searchCriteria.getValue());

            } else if (searchCriteria.getSearchField().equals(EntitySearchField.FULL_TEXT)) {
                if (searchCriteria.getCondition().equals(SearchCondition.EQUAL)) {
                    query = query + "E.fullText LIKE :" + EntitySearchField.DESCRIPTION.name() + " AND ";
                } else {
                    query = query + "E.fullText NOT LIKE :" + EntitySearchField.DESCRIPTION.name() + " AND ";
                }
                valueMap.put(EntitySearchField.FULL_TEXT.name(), searchCriteria.getValue());
            } else if (searchCriteria.getSearchField().equals(EntitySearchField.OWNER_ID)) {
                if (searchCriteria.getCondition().equals(SearchCondition.EQUAL)) {
                    query = query + "E.owner_id = :" + EntitySearchField.OWNER_ID.name() + " AND ";
                } else {
                    query = query + "E.owner_id != :" + EntitySearchField.OWNER_ID.name() + " AND ";
                }
                valueMap.put(EntitySearchField.OWNER_ID.name(), searchCriteria.getValue());

            } else if (searchCriteria.getSearchField().equals(EntitySearchField.ENTITY_TYPE_ID)) {
                if (searchCriteria.getCondition().equals(SearchCondition.EQUAL)) {
                    query = query + "E.entity_type_id = :" + EntitySearchField.ENTITY_TYPE_ID.name() + " AND ";
                } else {
                    query = query + "E.entity_type_id != :" + EntitySearchField.ENTITY_TYPE_ID.name() + " AND ";
                }
                valueMap.put(EntitySearchField.ENTITY_TYPE_ID.name(), searchCriteria.getValue());
            } else if (searchCriteria.getSearchField().equals(EntitySearchField.CREATED_AT)) {
                if (searchCriteria.getCondition().equals(SearchCondition.GTE)) {
                    query = query + "E.created_at >= :" + EntitySearchField.CREATED_AT.name() + " AND ";
                } else {
                    query = query + "E.created_at < :" + EntitySearchField.CREATED_AT.name() + " AND ";
                }
                valueMap.put(EntitySearchField.CREATED_AT.name(), searchCriteria.getValue());
            } else if (searchCriteria.getSearchField().equals(EntitySearchField.LAST_MODIFIED_AT)) {
                if (searchCriteria.getCondition().equals(SearchCondition.GTE)) {
                    query = query + "E.last_modified_at >= :" + EntitySearchField.ENTITY_TYPE_ID.name() + " AND ";
                } else {
                    query = query + "E.last_modified_at < :" + EntitySearchField.ENTITY_TYPE_ID.name() + " AND ";
                }
                valueMap.put(EntitySearchField.ENTITY_TYPE_ID.name(), searchCriteria.getValue());
            } else if (searchCriteria.getSearchField().equals(EntitySearchField.PARENT_ID)) {
                if (searchCriteria.getCondition().equals(SearchCondition.EQUAL)) {
                    query = query + "E.external_parent_id = :" + EntitySearchField.PARENT_ID.name() + " AND ";
                } else {
                    query = query + "E.external_parent_id != :" + EntitySearchField.PARENT_ID.name() + " AND ";
                }
                valueMap.put(EntitySearchField.PARENT_ID.name(), searchCriteria.getValue());
            } else if (searchCriteria.getSearchField().equals(EntitySearchField.SHARED_COUNT)) {
                if (searchCriteria.getCondition().equals(SearchCondition.GTE)) {
                    query = query + "E.shared_count >= :" + EntitySearchField.SHARED_COUNT.name() + " AND ";
                } else {
                    query = query + "E.shared_count < :" + EntitySearchField.SHARED_COUNT.name() + " AND ";
                }
                valueMap.put(EntitySearchField.SHARED_COUNT.name(), searchCriteria.getValue());
            }

        }
        query = query.substring(0, query.length() - 5);

        query = query + " ORDER BY E.created_at DESC";

        return query;
    }


}
