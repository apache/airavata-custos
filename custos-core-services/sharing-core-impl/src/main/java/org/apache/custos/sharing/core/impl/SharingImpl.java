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

package org.apache.custos.sharing.core.impl;

import org.apache.custos.sharing.core.*;
import org.apache.custos.sharing.core.exceptions.CustosSharingException;
import org.apache.custos.sharing.core.persistance.repository.EntityRepository;
import org.apache.custos.sharing.core.persistance.repository.EntityTypeRepository;
import org.apache.custos.sharing.core.persistance.repository.PermissionTypeRepository;
import org.apache.custos.sharing.core.persistance.repository.SharingRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.List;
import java.util.Optional;

public class SharingImpl implements SharingAPI {

    private static final Logger LOGGER = LoggerFactory.getLogger(SharingImpl.class);

    @Autowired
    private EntityRepository entityRepository;

    @Autowired
    private EntityTypeRepository entityTypeRepository;

    @Autowired
    private PermissionTypeRepository permissionTypeRepository;

    @Autowired
    private SharingRepository sharingRepository;

    @Override
    public void createEntityType(String tenantId,EntityType entityType) {
        try {
            LOGGER.debug("Creating entity type "+ entityType.getId()+" in tenant "+ tenantId);

            String internalId = entityType.getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);

            if (optionalEntityType.isPresent()) {
                String msg = "Entity type  id" + entityType.getId() + "is already present";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            org.apache.custos.sharing.core.persistance.model.EntityType
                    type = EntityTypeMapper.createEntityType(request.getEntityType(), request.getTenantId());
            entityTypeRepository.save(type);

            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while creating entity type for " + request.getEntityType().getId() +
                    " in" + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateEntityType(String tenantId,EntityType entityType) {

    }

    @Override
    public void deleteEntityType(String tenantId, String entityTypeId) {

    }

    @Override
    public EntityType getEntityType(String tenantId, String entityTypeId) {
        return null;
    }

    @Override
    public List<EntityType> getEntityTypes(String tenantId) {
        return null;
    }

    @Override
    public void createPermissionType(PermissionType permissionType) {

    }

    @Override
    public void updatePermissionType(PermissionType permissionType) {

    }

    @Override
    public void deletePermissionType(String tenantId, String permissionTypeId) {

    }

    @Override
    public EntityType getPermissionType(String tenantId, String permissionTypeId) {
        return null;
    }

    @Override
    public List<EntityType> getPermissionTypes(String tenantId) {
        return null;
    }

    @Override
    public void createEntity(Entity entity) {

    }

    @Override
    public void updateEntity(Entity entity) {

    }

    @Override
    public void deleteEntity(String tenantId, String entityId) {

    }

    @Override
    public boolean isEntityExists(String tenantId, String entityId) {
        return false;
    }

    @Override
    public Entity getEntity(String tenantId, String entityId) {
        return null;
    }

    @Override
    public List<Entity> searchEntities(String tenantId, List<SearchCriteria> searchCriteriaList, int limit, int offset) {
        return null;
    }

    @Override
    public List<SharedOwners> getListOfSharedUsers(String tenantId, String entityId, String permissionTypeId) {
        return null;
    }

    @Override
    public List<SharedOwners> getListOfDirectlySharedUsers(String tenantId, String entityId, String permissionTypeId) {
        return null;
    }

    @Override
    public List<SharedOwners> getListOfSharedGroups(String tenantId, String entityId, String permissionTypeId) {
        return null;
    }

    @Override
    public List<SharedOwners> getListOfDirectlySharedGroups(String tenantId, String entityId, String permissionTypeId) {
        return null;
    }

    @Override
    public boolean revokePermission(String tenantId, String entityId, String permissionType, List<String> usersList) {
        return false;
    }

    @Override
    public boolean userHasAccess(String tenantId, String entityId, String permission, String username) {
        return false;
    }
}
