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


import com.google.protobuf.ByteString;
import org.apache.custos.sharing.persistance.model.Entity;
import org.apache.custos.sharing.persistance.model.EntityType;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.sql.rowset.serial.SerialBlob;
import java.sql.Blob;
import java.sql.SQLException;

public class EntityMapper {

    private static final Logger LOGGER = LoggerFactory.getLogger(EntityMapper.class);


    public static Entity createEntity(org.apache.custos.sharing.service.Entity entity, long tenantId,
                                      EntityType type) throws SQLException {

        Entity perEntity = new Entity();

        String internalId = entity.getId() + "@" + tenantId;


        perEntity.setEntityType(type);
        perEntity.setId(internalId);
        perEntity.setExternalId(entity.getId());
        perEntity.setName(entity.getName());
        perEntity.setTenantId(tenantId);
        perEntity.setOwnerId(entity.getOwnerId());
        perEntity.setSharedCount(entity.getSharedCount());

        if (entity.getBinaryData() != null && !entity.getBinaryData().isEmpty()) {
            perEntity.setBinaryData(new SerialBlob(entity.getBinaryData().toByteArray()));
        }

        if (entity.getDescription() != null && !entity.getDescription().trim().equals("")) {
            perEntity.setDescription(entity.getDescription());
        }

        if (entity.getFullText() != null && entity.getFullText().trim().equals("")) {
            perEntity.setFullText(entity.getFullText());
        }

        if (entity.getParentId() != null && entity.getParentId().trim().equals("")) {
            perEntity.setExternalParentId(entity.getParentId());
        }


        return perEntity;


    }


    public static org.apache.custos.sharing.service.Entity createEntity(Entity entity) throws SQLException {

        org.apache.custos.sharing.service.Entity.Builder builder =
                org.apache.custos.sharing.service.Entity
                        .newBuilder();

        builder
                .setId(entity.getExternalId())
                .setCreatedAt(entity.getCreatedAt().getTime())
                .setUpdatedAt(entity.getLastModifiedAt().getTime())
                .setName(entity.getName())
                .setOriginalCreationTime(entity.getCreatedAt().getTime())
                .setOwnerId(entity.getOwnerId())
                .setType(entity.getEntityType().getId())
                .setSharedCount(entity.getSharedCount());

        if (entity.getDescription() != null) {
            builder.setDescription(entity.getDescription());
        }

        if (entity.getBinaryData() != null) {
            Blob blob = entity.getBinaryData();
            int len = (int) blob.length();
            byte[] blobAsBytes = blob.getBytes(1, len);
            builder.setBinaryData(ByteString.copyFrom(blobAsBytes));
        }

        if (entity.getFullText() != null) {
            builder.setFullText(entity.getFullText());
        }

        return builder.build();

    }


}
