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

package org.apache.custos.sharing.core.mapper;

import org.apache.custos.sharing.core.SharingMetadata;
import org.apache.custos.sharing.core.persistance.model.Entity;
import org.apache.custos.sharing.core.persistance.model.PermissionType;
import org.apache.custos.sharing.core.persistance.model.Sharing;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.sql.SQLException;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

public class SharingMapper {

    private static final Logger LOGGER = LoggerFactory.getLogger(SharingMapper.class);


    public static Sharing createSharing(PermissionType permissionType,
                                        Entity entity,
                                        Entity inheritedEntity,
                                        String ownerId,
                                        String ownerType,
                                        String sharingType,
                                        String sharedBy,
                                        String tenantId) {

        String id = entity.getId() + "_" +
                inheritedEntity.getId() + "_" + ownerId + "_" + permissionType.getId() + "_" + tenantId;

        Sharing sharing = new Sharing();
        sharing.setSharingType(sharingType);
        sharing.setEntity(entity);
        sharing.setPermissionType(permissionType);
        sharing.setAssociatingId(ownerId);
        sharing.setAssociatingIdType(ownerType);
        sharing.setInheritedParent(inheritedEntity);
        sharing.setTenantId(tenantId);
        sharing.setId(id);
        if (sharedBy != null) {
            sharing.setSharedBy(sharedBy);
        }
        return sharing;
    }


    public static Sharing getNewSharing(Sharing oldSharing, String tenantId, Entity entity) {
        String id = entity.getId() + "_" +
                oldSharing.getInheritedParent().getId() + "_" + oldSharing.getAssociatingId() + "_" + oldSharing.getPermissionType().getId() + "_" + tenantId;

        Sharing sharing = new Sharing();
        sharing.setSharingType(oldSharing.getSharingType());
        sharing.setEntity(entity);
        sharing.setPermissionType(oldSharing.getPermissionType());
        sharing.setAssociatingId(oldSharing.getAssociatingId());
        sharing.setAssociatingIdType(oldSharing.getAssociatingIdType());
        sharing.setInheritedParent(oldSharing.getInheritedParent());
        sharing.setCreatedAt(oldSharing.getCreatedAt());
        if (oldSharing.getSharedBy() != null) {
            sharing.setSharedBy(oldSharing.getSharedBy());
        }
        sharing.setTenantId(tenantId);
        sharing.setId(id);
        return sharing;

    }


    public static List<String> getSharedOwners(List<Sharing> sharingList) {

        List<String> ownerIds = new ArrayList<>();

        if (sharingList != null && !sharingList.isEmpty()) {

            ownerIds = sharingList.stream().
                    map(shr -> shr.getAssociatingId()).collect(Collectors.toList());

        }
        return ownerIds;

    }

    public static SharingMetadata getSharingMetadata(Sharing sharing, Entity entity,
                                                     PermissionType permissionType, String type) throws SQLException {
        org.apache.custos.sharing.core.Entity en = EntityMapper.createEntity(entity);
        return SharingMetadata.newBuilder()
                .setEntity(en)
                .setOwnerId(sharing.getAssociatingId())
                .setOwnerType(type)
                .setSharedBy(sharing.getSharedBy()!=null?sharing.getSharedBy():"")
                .addPermissions(org.apache.custos.sharing.core.PermissionType.newBuilder()
                        .setId(permissionType.getExternalId()).build()).build();

    }

    public static Optional<List<SharingMetadata>> getSharingMetadata(List<Sharing> sharingList) {
        if (sharingList != null && !sharingList.isEmpty()) {

            return Optional.ofNullable(sharingList.stream().
                    map(shr -> {
                        try {
                            SharingMetadata metadata = SharingMetadata.newBuilder()
                                    .addPermissions(org.apache.custos.sharing.core.PermissionType.newBuilder().
                                            setId(shr.getPermissionType().getExternalId())
                                            .setName(shr.getPermissionType().getName()
                                            ).setDescription(shr.getPermissionType().getDescription() == null ?
                                                    "" : shr.getPermissionType().getDescription()).
                                                    setCreatedAt(shr.getCreatedAt().getTime())
                                            .setUpdatedAt(shr.getLastModifiedAt().getTime()).build())
                                    .setOwnerType(shr.getAssociatingIdType())
                                    .setOwnerId(shr.getAssociatingId())
                                    .setEntity(EntityMapper.createEntity(shr.getEntity()))
                                    .setSharedBy(shr.getSharedBy() != null ? shr.getSharedBy() : "")
                                    .build();
                            return metadata;
                        } catch (SQLException throwables) {
                            String msg = "Error occurred while creating metadata " + throwables.getMessage();
                            LOGGER.error(msg, throwables);
                        }
                        ;
                        return null;
                    }).collect(Collectors.toList()));

        }
        return Optional.empty();

    }

}
