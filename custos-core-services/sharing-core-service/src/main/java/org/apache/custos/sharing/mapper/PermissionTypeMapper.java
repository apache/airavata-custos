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

package org.apache.custos.sharing.mapper;

import org.apache.custos.sharing.persistance.model.PermissionType;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class PermissionTypeMapper {

    private static final Logger LOGGER = LoggerFactory.getLogger(PermissionTypeMapper.class);

    public static PermissionType createPermissionType(org.apache.custos.sharing.service.PermissionType entityType, long tenantId) {

        PermissionType type = new PermissionType();
        String id = entityType.getId() + "@" + tenantId;
        type.setId(id);
        type.setName(entityType.getName());
        type.setTenantId(tenantId);
        type.setExternalId(entityType.getId());
        if (entityType.getDescription() != null && entityType.getDescription().trim().equals("")) {
            type.setDescription(entityType.getDescription());
        }
        return type;
    }

    public static org.apache.custos.sharing.service.PermissionType createPermissionType(PermissionType type) {
        org.apache.custos.sharing.service.PermissionType.Builder builder =

                org.apache.custos.sharing.service.PermissionType
                        .newBuilder()
                        .setId(type.getExternalId())
                        .setCreatedAt(type.getCreatedAt().getTime())
                        .setName(type.getName())
                        .setUpdatedAt(type.getLastModifiedAt().getTime());

        if (type.getDescription() != null) {
            builder.setDescription(type.getDescription());
        }

        return builder.build();

    }
}
