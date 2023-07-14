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
import org.apache.custos.sharing.core.mapper.EntityMapper;
import org.apache.custos.sharing.core.mapper.EntityTypeMapper;
import org.apache.custos.sharing.core.mapper.PermissionTypeMapper;
import org.apache.custos.sharing.core.mapper.SharingMapper;
import org.apache.custos.sharing.core.persistance.model.Sharing;
import org.apache.custos.sharing.core.persistance.repository.EntityRepository;
import org.apache.custos.sharing.core.persistance.repository.EntityTypeRepository;
import org.apache.custos.sharing.core.persistance.repository.PermissionTypeRepository;
import org.apache.custos.sharing.core.persistance.repository.SharingRepository;
import org.apache.custos.sharing.core.utils.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.sql.SQLException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;

@Component
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
    public void createEntityType(String tenantId, EntityType entityType) throws CustosSharingException {
        try {
            LOGGER.debug("Creating entity type " + entityType.getId() + " in tenant " + tenantId);

            String internalId = entityType.getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);

            if (optionalEntityType.isPresent()) {
                String msg = "Entity type  id" + entityType.getId() + "is already present";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            org.apache.custos.sharing.core.persistance.model.EntityType
                    type = EntityTypeMapper.createEntityType(entityType, tenantId);
            entityTypeRepository.save(type);

        } catch (Exception ex) {
            String msg = "Error occurred while creating entity type for " + entityType.getId() +
                    " in" + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public void updateEntityType(String tenantId, EntityType entityType) throws CustosSharingException {
        try {
            LOGGER.debug("Updating entity type  " + entityType.getId() + " in " + tenantId);

            String internalId = entityType.getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);

            if (optionalEntityType.isEmpty()) {
                String msg = "Entity type  id" + entityType.getId() + "is not present";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            org.apache.custos.sharing.core.persistance.model.EntityType
                    type = EntityTypeMapper.createEntityType(entityType, tenantId);

            type.setCreatedAt(optionalEntityType.get().getCreatedAt());

            entityTypeRepository.save(type);

        } catch (Exception ex) {
            String msg = "Error occurred while updating entity type" + entityType.getId()
                    + "in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public void deleteEntityType(String tenantId, String entityTypeId) throws CustosSharingException {
        try {
            LOGGER.debug("deleting entity" + entityTypeId + "in " + tenantId);

            String internalId = entityTypeId + "@" + tenantId;
            Optional<org.apache.custos.sharing.core.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);

            if (optionalEntityType.isEmpty()) {
                String msg = "Entity type  id" + entityTypeId + "is not present";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            entityTypeRepository.delete(optionalEntityType.get());

        } catch (Exception ex) {
            String msg = "Error occurred while deleting entity type" + entityTypeId
                    + "in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public Optional<EntityType> getEntityType(String tenantId, String entityTypeId) throws CustosSharingException {
        try {
            LOGGER.debug("Fetching entity type  " + entityTypeId + " in " + tenantId);

            String internalId = entityTypeId + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);

            if (optionalEntityType.isPresent()) {
                EntityType type =
                        EntityTypeMapper.createEntityType(optionalEntityType.get());
                return Optional.of(type);
            } else {
                return Optional.empty();
            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching entity type for " + entityTypeId +
                    " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public List<EntityType> getEntityTypes(String tenantId) throws CustosSharingException {
        try {
            LOGGER.debug(" Fetching entity types " + " in " + tenantId);

            List<org.apache.custos.sharing.core.persistance.model.EntityType> types =
                    entityTypeRepository.findAllByTenantId(tenantId);

            List<EntityType> entityTypes = new ArrayList<>();

            if (!types.isEmpty()) {

                for (org.apache.custos.sharing.core.persistance.model.EntityType type : types) {
                    entityTypes.add(EntityTypeMapper.createEntityType(type));
                }

            }
            return entityTypes;
        } catch (Exception ex) {
            String msg = "Error occurred while fetching entity types  in "
                    + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public void createPermissionType(PermissionType permissionType, String tenantId) throws CustosSharingException {
        try {
            LOGGER.debug("Creating Permission type " + permissionType.getId() + " in " + tenantId);

            String internalId = permissionType + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalEntityType =
                    permissionTypeRepository.findById(internalId);

            if (optionalEntityType.isPresent()) {
                String msg = "Permission type  id " + permissionType + "is already present";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }
            org.apache.custos.sharing.core.persistance.model.PermissionType
                    type = PermissionTypeMapper.createPermissionType(permissionType, tenantId);
            permissionTypeRepository.save(type);
        } catch (Exception ex) {
            String msg = "Error occurred while creating permission type with id " + permissionType.getId()
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public void updatePermissionType(PermissionType permissionType, String tenantId) throws CustosSharingException {
        try {
            LOGGER.debug("UpdatePermissionType " + permissionType.getId() + " in " + tenantId);

            String internalId = permissionType.getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalEntityType =
                    permissionTypeRepository.findById(internalId);

            if (optionalEntityType.isEmpty()) {
                String msg = "Entity type  id" + permissionType.getId() + "is not present";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            org.apache.custos.sharing.core.persistance.model.PermissionType
                    type = PermissionTypeMapper.createPermissionType(permissionType, tenantId);

            type.setCreatedAt(optionalEntityType.get().getCreatedAt());

            permissionTypeRepository.save(type);

        } catch (Exception ex) {
            String msg = "Error occurred while updating permission type with id " + permissionType.getId()
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }

    }

    @Override
    public void deletePermissionType(String tenantId, String permissionTypeId) throws CustosSharingException {
        try {
            LOGGER.debug("Deleting PermissionType" + permissionTypeId + " in " + tenantId);

            String internalId = permissionTypeId + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalEntityType =
                    permissionTypeRepository.findById(internalId);

            if (optionalEntityType.isEmpty()) {
                String msg = "Entity type  id" + permissionTypeId + "is not present";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            permissionTypeRepository.delete(optionalEntityType.get());

        } catch (Exception ex) {
            String msg = "Error occurred while deleting permission type with id " + permissionTypeId
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public Optional<PermissionType> getPermissionType(String tenantId, String permissionTypeId)
            throws CustosSharingException {
        try {
            LOGGER.debug("Request received to getPermissionType with id " + permissionTypeId + " in " + tenantId);

            String internalId = permissionTypeId + "@" + tenantId;
            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalEntityType =
                    permissionTypeRepository.findById(internalId);


            if (optionalEntityType.isPresent()) {
                PermissionType type =
                        PermissionTypeMapper.createPermissionType(optionalEntityType.get());
                return Optional.of(type);
            } else {
                return Optional.empty();
            }
        } catch (Exception ex) {
            String msg = "Error occurred while fetching permission type with id " + permissionTypeId
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public List<PermissionType> getPermissionTypes(String tenantId) throws CustosSharingException {
        try {
            LOGGER.debug("Fetching  permissionTypes  of " + tenantId);

            List<org.apache.custos.sharing.core.persistance.model.PermissionType> types =
                    permissionTypeRepository.findAllByTenantId(tenantId);

            List<org.apache.custos.sharing.core.PermissionType> entityTypes = new ArrayList<>();

            if (!types.isEmpty()) {
                for (org.apache.custos.sharing.core.persistance.model.PermissionType type : types) {
                    entityTypes.add(PermissionTypeMapper.createPermissionType(type));
                }
            }
            return entityTypes;
        } catch (Exception ex) {
            String msg = "Error occurred while fetching permission types"
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public void createEntity(Entity entity, String tenantId) throws CustosSharingException {
        try {
            LOGGER.debug("createEntity with id" + entity.getId() + " in " + tenantId);

            String entityTypeId = entity.getType() + "@" + tenantId;

            String internalParentId = null;

            if (entity.getParentId() != null && !entity.getParentId().isEmpty()) {

                internalParentId = entity.getParentId() + "@" + tenantId;

                Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity = entityRepository.
                        findById(internalParentId);

                if (optionalEntity.isEmpty()) {
                    String msg = "Cannot find a parent Entity  with given Id " + entity.getParentId();
                    LOGGER.error(msg);
                    throw new CustosSharingException(msg);
                }
            }

            Optional<org.apache.custos.sharing.core.persistance.model.EntityType> entityType =
                    entityTypeRepository.findById(entityTypeId);

            if (entityType.isEmpty()) {
                String msg = "Cannot find a Entity Type with given Id " + entity.getId();
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            org.apache.custos.sharing.core.persistance.model.Entity enModel =
                    EntityMapper.createEntity(entity, tenantId, entityType.get());

            org.apache.custos.sharing.core.persistance.model.Entity savedEntity = entityRepository.save(enModel);

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalPermissionType =
                    permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);

            if (optionalPermissionType.isPresent()) {
                Sharing sharing = SharingMapper.createSharing(optionalPermissionType.get(),
                        savedEntity, savedEntity, entity.getOwnerId(), Constants.USER, Constants.DIRECT_CASCADING,
                        entity.getOwnerId(), tenantId);
                sharingRepository.save(sharing);
            }

            if (internalParentId != null) {
                addCascadingPermissionForEntity(savedEntity, internalParentId, tenantId);
            }

            List<String> sharingType = new ArrayList<>();
            sharingType.add(Constants.INDIRECT_CASCADING);
            sharingType.add(Constants.DIRECT_CASCADING);
            sharingType.add(Constants.DIRECT_NON_CASCADING);

            List<Sharing> sharings = sharingRepository.findAllByEntityAndSharingType(tenantId,
                    enModel.getId(), sharingType);

            if (sharings != null && sharings.size() > 0) {
                savedEntity.setSharedCount(sharings.size());
                entityRepository.save(savedEntity);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while creating Entity with id " + entity.getId()
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public void updateEntity(Entity entity, String tenantId) throws CustosSharingException {
        try {
            LOGGER.debug("Request received to updateEntity with id" + entity.getId() + " in " + tenantId);

            String entityTypeId = entity.getType() + "@" + tenantId;

            String internalEntityId = entity.getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity =
                    entityRepository.findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find a Entity  with given Id " + entity.getId();
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            Optional<org.apache.custos.sharing.core.persistance.model.EntityType> entityType =
                    entityTypeRepository.findById(entityTypeId);

            if (entityType.isEmpty()) {
                String msg = "Cannot find a Entity type with given Id " + entity.getType();
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            org.apache.custos.sharing.core.persistance.model.Entity oldEntity = optionalEntity.get();

            org.apache.custos.sharing.core.persistance.model.Entity newEntity =
                    EntityMapper.createEntity(entity, tenantId, entityType.get());

            if (!oldEntity.getOwnerId().equals(newEntity.getOwnerId())) {
                String msg = "Entity owner cannot change ";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            if (oldEntity.getExternalParentId() != null &&
                    !oldEntity.getExternalParentId().equals(newEntity.getExternalParentId())) {

                sharingRepository.removeGivenCascadingPermissionsForEntity(tenantId, internalEntityId, Constants.INDIRECT_CASCADING);
            }

            if (newEntity.getExternalParentId() != null) {
                String internalParentId = newEntity.getExternalParentId() + "@" + tenantId;
                addCascadingPermissionForEntity(newEntity, internalParentId, tenantId);
            }

            List<String> sharingType = new ArrayList<>();
            sharingType.add(Constants.INDIRECT_CASCADING);
            sharingType.add(Constants.DIRECT_CASCADING);
            sharingType.add(Constants.DIRECT_NON_CASCADING);

            List<Sharing> sharings = sharingRepository.findAllByEntityAndSharingType(tenantId,
                    internalEntityId, sharingType);

            if (sharings != null && sharings.size() > 0) {
                newEntity.setSharedCount(sharings.size());
            }

            newEntity.setCreatedAt(oldEntity.getCreatedAt());

            entityRepository.save(newEntity);

        } catch (Exception ex) {
            String msg = "Error occurred while updating Entity with id " + entity.getId()
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public void deleteEntity(String tenantId, String entityId) throws CustosSharingException {
        try {
            LOGGER.debug("Deleting entity with id" + entityId + " in " + tenantId);

            String internalId = entityId + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity =
                    entityRepository.findById(internalId);

            if (optionalEntity.isPresent()) {
                entityRepository.delete(optionalEntity.get());
            }
        } catch (Exception ex) {
            String msg = "Error occurred while deleting entity with id " + entityId
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public boolean isEntityExists(String tenantId, String entityId) throws CustosSharingException {
        try {
            LOGGER.debug(" Check entity exists with given id" + entityId + " in " + tenantId);

            String internalId = entityId + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity =
                    entityRepository.findById(internalId);

            if (optionalEntity.isEmpty()) {
                return false;
            } else {
                return true;
            }

        } catch (Exception ex) {
            String msg = "Error occurred while checking isEntityExists with id " + entityId
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public Optional<Entity> getEntity(String tenantId, String entityId) throws CustosSharingException {
        try {
            LOGGER.debug("Fetching entity with id" + entityId + " in " + tenantId);

            String internalId = entityId + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity =
                    entityRepository.findById(internalId);

            if (optionalEntity.isEmpty()) {
                return Optional.empty();
            } else {
                return Optional.of(EntityMapper.createEntity(optionalEntity.get()));
            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching Entity with id " + entityId
                    + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public List<Entity> searchEntities(String tenantId, List<SearchCriteria> searchCriteriaList, List<String> associatingIdList,
                                       int limit, int offset,
                                       boolean searchPermBottomUp) throws CustosSharingException {
        try {
            LOGGER.debug("Search entities in tenant " + tenantId);

            List<org.apache.custos.sharing.core.persistance.model.Entity> entities = new ArrayList<>();
            if (searchPermBottomUp) {
                entities = entityRepository.
                        searchEntitiesRecursive(tenantId, searchCriteriaList);
            } else {
                entities = entityRepository.
                        searchEntities(tenantId, searchCriteriaList, limit, offset);
            }

            HashMap<String, Entity> entryMap = new HashMap<>();

            List<String> internalEntityIds = new ArrayList<>();

            if (entities != null && !entities.isEmpty()) {

                for (org.apache.custos.sharing.core.persistance.model.Entity entity : entities) {
                    internalEntityIds.add(entity.getId());
                }

                List<Sharing> sharings = sharingRepository.
                        findAllSharingEntitiesForUsers(tenantId, associatingIdList, internalEntityIds);


                if (sharings != null && !sharings.isEmpty()) {
                    if (searchPermBottomUp) {
                        for (org.apache.custos.sharing.core.persistance.model.Entity en : entities) {
                            Entity entity = EntityMapper.createEntity(en);
                            entryMap.put(entity.getId(), entity);
                        }

                    } else {
                        for (Sharing sharing : sharings) {
                            Entity entity = EntityMapper.createEntity(sharing.getEntity());
                            entryMap.put(entity.getId(), entity);
                        }
                    }
                }
            }
            List<org.apache.custos.sharing.core.Entity> entityList = new ArrayList<>(entryMap.values());
            return entityList;

        } catch (Exception ex) {
            String msg = "Error occurred while searching entities" + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public List<String> getListOfSharedUsers(String tenantId, String entityId, String permissionTypeId) throws CustosSharingException {
        try {
            LOGGER.debug("Fetching shared users in tenant " + tenantId + " for entity " + entityId);
            List<Sharing> sharings = null;

            String internalEntityId = entityId + "@" + tenantId;

            String internalPermissionTypeId = permissionTypeId + "@" + tenantId;
            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> permissionType =
                    permissionTypeRepository.findById(internalPermissionTypeId);

            if (permissionType.isEmpty()) {
                String msg = "Cannot find permission type" + permissionTypeId;
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity = entityRepository.
                    findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find given entity" + entityId;
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalPermissionType =
                    permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);

            if (optionalPermissionType.get().equals(internalPermissionTypeId)) {
                List<String> sharingList = new ArrayList<>();
                sharingList.add(Constants.DIRECT_CASCADING);
                sharingList.add(Constants.DIRECT_NON_CASCADING);

                sharings = sharingRepository.
                        findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType(tenantId, internalEntityId,
                                internalPermissionTypeId, Constants.USER, sharingList);
            } else {

                sharings = sharingRepository.
                        findAllByEntityAndPermissionTypeAndOwnerType(tenantId, internalEntityId,
                                internalPermissionTypeId, Constants.USER);
            }

            return SharingMapper.getSharedOwners(sharings);
        } catch (Exception ex) {
            String msg = "Error occurred while fetching  getListOfSharedUsers  for entity "
                    + entityId + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public List<String> getListOfDirectlySharedUsers(String tenantId, String entityId, String permissionTypeId) throws CustosSharingException {
        try {
            LOGGER.debug("Get list of directly shared users in  " + tenantId + " for entity " + entityId);

            String internalEntityId = entityId + "@" + tenantId;

            String internalPermissionTypeId = permissionTypeId + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> oPT =
                    permissionTypeRepository.findById(internalPermissionTypeId);

            if (oPT.isEmpty()) {
                String msg = "Cannot find permission type" + permissionTypeId;
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity = entityRepository.
                    findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find given entity" + entityId;
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            List<String> sharingList = new ArrayList<>();
            sharingList.add(Constants.DIRECT_CASCADING);
            sharingList.add(Constants.DIRECT_NON_CASCADING);

            List<Sharing> sharings = sharingRepository.
                    findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType(tenantId, internalEntityId,
                            internalPermissionTypeId, Constants.USER, sharingList);
            return SharingMapper.getSharedOwners(sharings);

        } catch (Exception ex) {
            String msg = "Error occurred while fetching  directly shared users  for entity "
                    + entityId + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }

    }

    @Override
    public List<String> getListOfSharedGroups(String tenantId, String entityId, String permissionTypeId) throws CustosSharingException {
        try {
            LOGGER.debug("Fetching list of shared groups " + tenantId + " for entity " + entityId);

            String internalEntityId = entityId + "@" + tenantId;
            List<Sharing> sharings = null;

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> permissionType =
                    permissionTypeRepository.findById(permissionTypeId);

            String internalPermissionTypeId = permissionTypeId + "@" + tenantId;

            if (permissionType.isEmpty()) {
                String msg = "Cannot find permission type" + permissionTypeId;
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity = entityRepository.
                    findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find given entity" + entityId;
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            sharings = sharingRepository.
                    findAllByEntityAndPermissionTypeAndOwnerType(tenantId, internalEntityId,
                            internalPermissionTypeId, Constants.GROUP);

            return SharingMapper.getSharedOwners(sharings);

        } catch (Exception ex) {
            String msg = "Error occurred while fetching   shared groups  for entity " + permissionTypeId + " in "
                    + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public List<String> getListOfDirectlySharedGroups(String tenantId, String entityId, String permissionTypeId)
            throws CustosSharingException {
        try {
            LOGGER.debug("Fetching list of shared groups  " + tenantId + " for entity " + entityId);

            String internalEntityId = entityId + "@" + tenantId;
            List<Sharing> sharings = null;

            String internalPermissionTypeId = permissionTypeId + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> permissionType =
                    permissionTypeRepository.findById(internalPermissionTypeId);

            if (permissionType.isEmpty()) {
                String msg = "Cannot find permission type" + permissionTypeId;
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            Optional<org.apache.custos.sharing.core.persistance.model.Entity> optionalEntity = entityRepository.
                    findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find given entity" + entityId;
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }

            sharings = sharingRepository.
                    findAllByEntityAndPermissionTypeAndOwnerType(tenantId, internalEntityId,
                            internalPermissionTypeId, Constants.GROUP);

            return SharingMapper.getSharedOwners(sharings);

        } catch (Exception ex) {
            String msg = "Error occurred while fetching   shared groups  for entity "
                    + entityId + " in " + tenantId + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public boolean revokePermission(String tenantId, String entityId, String permissionType, List<String> usersList) throws CustosSharingException {

        String internalEntityId = entityId + "@" + tenantId;

        String internalPermissionType = permissionType + "@" + tenantId;

        Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalPermissionType =
                permissionTypeRepository.findById(internalPermissionType);

        if (optionalPermissionType.isEmpty()) {
            String msg = "Permission type not found";
            LOGGER.error(msg);
            throw new CustosSharingException(msg);
        }

        Optional<org.apache.custos.sharing.core.persistance.model.Entity> entityOptional =
                entityRepository.findById(internalEntityId);

        if (entityOptional.isEmpty()) {
            String msg = " Entity with Id " + entityId + "  not found";
            LOGGER.error(msg);
            throw new CustosSharingException(msg);
        }

        Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalOwnerPermissionType =
                permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);

        if (optionalOwnerPermissionType.get().getId().equals(internalPermissionType)) {
            String msg = "Owner permission type can not be assigned";
            LOGGER.error(msg);
            throw new CustosSharingException(msg);
        }


        for (String userId : usersList) {

            sharingRepository.
                    deleteAllByEntityIdAndPermissionTypeIdAndAssociatingIdAndTenantIdAndInheritedParentId(
                            internalEntityId,
                            internalPermissionType,
                            userId,
                            tenantId,
                            internalEntityId);
            sharingRepository.deleteAllByInheritedParentIdAndPermissionTypeIdAndTenantIdAndSharingTypeAndAssociatingId(
                    internalEntityId,
                    internalPermissionType,
                    tenantId,
                    Constants.INDIRECT_CASCADING,
                    userId);
        }


        List<String> checkTypes = new ArrayList<>();
        checkTypes.add(Constants.INDIRECT_CASCADING);
        checkTypes.add(Constants.DIRECT_CASCADING);
        checkTypes.add(Constants.DIRECT_NON_CASCADING);

        List<Sharing> newSharings = sharingRepository.findAllByEntityAndSharingType(tenantId,
                internalEntityId, checkTypes);
        org.apache.custos.sharing.core.persistance.model.Entity entity = entityOptional.get();
        if (newSharings != null && newSharings.size() > 0) {
            entity.setSharedCount(newSharings.size());
            entityRepository.save(entity);
        }

        return true;
    }

    @Override
    public boolean userHasAccess(String tenantId, String entityId, String permission, String username) throws CustosSharingException {
        try {
            LOGGER.debug("Check user access in " + tenantId + " for entity "
                    + entityId + "  of  user " + username);

            String internalPermissionId = permission + "@" + tenantId;

            String internalEntityId = entityId + "@" + tenantId;

            Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> permissionType =
                    permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);

            List<String> permissionTypes = new ArrayList<>();
            permissionTypes.add(internalPermissionId);
            permissionTypes.add(permissionType.get().getId());

            List<String> userList = new ArrayList<>();
            userList.add(username);

            List<Sharing> sharings = sharingRepository.findAllSharingOfEntityForGroupsUnderPermissions(tenantId,
                    internalEntityId, permissionTypes, userList);

            if (sharings != null && !sharings.isEmpty()) {

                return true;
            }
            return false;

        } catch (Exception ex) {
            String msg = "Error occurred while checking access to  " + entityId + " in " + tenantId +
                    " for user" + username +
                    " reason :" + ex.getMessage();
            LOGGER.error(msg);
            throw new CustosSharingException(msg, ex);
        }
    }

    @Override
    public void shareEntity(String tenantId, String entityId, String permissionType, List<String> userIds,
                            boolean cascade, String ownerType, String sharedBy) throws CustosSharingException {

        String internalPermissionId = permissionType + "@" + tenantId;


        String internalEntityId = entityId + "@" + tenantId;


        Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalPermissionType =
                permissionTypeRepository.findById(internalPermissionId);

        if (optionalPermissionType.isEmpty()) {
            String msg = "Permission type not found";
            LOGGER.error(msg);
            throw new CustosSharingException(msg);
        }

        Optional<org.apache.custos.sharing.core.persistance.model.Entity> entityOptional =
                entityRepository.findById(internalEntityId);

        if (entityOptional.isEmpty()) {
            String msg = " Entity with Id " + entityId + "  not found";
            LOGGER.error(msg);
            throw new CustosSharingException(msg);
        }


        Optional<org.apache.custos.sharing.core.persistance.model.PermissionType> optionalOwnerPermissionType =
                permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);

        if (optionalOwnerPermissionType.get().getId().equals(internalPermissionId)) {
            String msg = "Owner permission type can not be assigned";
            LOGGER.error(msg);
            throw new CustosSharingException(msg);
        }

        String sharingType = null;
        List<Sharing> sharings = new ArrayList<>();

        if (cascade) {
            sharingType = Constants.DIRECT_CASCADING;
        } else {
            sharingType = Constants.DIRECT_NON_CASCADING;
        }


        for (String userId : userIds) {

            Sharing sharing = SharingMapper.createSharing(optionalPermissionType.get(),
                    entityOptional.get(), entityOptional.get(), userId, ownerType, sharingType, sharedBy,
                    tenantId);
            sharings.add(sharing);
        }

        if (cascade) {
            List<Sharing> childSharings = new ArrayList<>();
            childSharings = getAllSharingForChildEntities(entityOptional.get(), entityOptional.get(), userIds, tenantId,
                    optionalPermissionType.get(), childSharings, ownerType,
                    Constants.INDIRECT_CASCADING, sharedBy);
            if (!childSharings.isEmpty()) {
                sharings.addAll(childSharings);
            }
        }


        List<Sharing> effectiveSharings = new ArrayList<>();
        if (!sharings.isEmpty()) {
            sharings.forEach(shr -> {
                Optional<Sharing> sharing = sharingRepository.findById(shr.getId());
                if (sharing.isEmpty()) {
                    effectiveSharings.add(shr);
                }

            });

        }

        if (!effectiveSharings.isEmpty()) {
            sharingRepository.saveAll(effectiveSharings);

            //revoke other permissions
            for (String userId : userIds) {
                List<org.apache.custos.sharing.core.persistance.model.PermissionType> existingPermissionTypes =
                        permissionTypeRepository.findAllByTenantId(tenantId);

                existingPermissionTypes.forEach(permission -> {
                    if (!(permission.getExternalId().equals(Constants.OWNER) || permission.getId().equals(internalPermissionId))) {
                        sharingRepository.
                                deleteAllByEntityIdAndPermissionTypeIdAndAssociatingIdAndTenantIdAndInheritedParentId(
                                        internalEntityId,
                                        permission.getId(),
                                        userId,
                                        tenantId,
                                        internalEntityId);
                        sharingRepository.deleteAllByInheritedParentIdAndPermissionTypeIdAndTenantIdAndSharingTypeAndAssociatingId(
                                internalEntityId,
                                permission.getId(),
                                tenantId,
                                Constants.INDIRECT_CASCADING,
                                userId);
                    }
                });
            }


            List<String> checkTypes = new ArrayList<>();
            checkTypes.add(Constants.INDIRECT_CASCADING);
            checkTypes.add(Constants.DIRECT_CASCADING);
            checkTypes.add(Constants.DIRECT_NON_CASCADING);

            List<Sharing> newSharings = sharingRepository.findAllByEntityAndSharingType(tenantId,
                    internalEntityId, checkTypes);
            org.apache.custos.sharing.core.persistance.model.Entity entity = entityOptional.get();
            if (newSharings != null && newSharings.size() > 0) {
                entity.setSharedCount(newSharings.size());
                entityRepository.save(entity);
            }


        }
    }


    @Override
    public List<SharingMetadata> getAllDirectSharings(String tenantId) {
        List<org.apache.custos.sharing.core.persistance.model.Entity> entities = entityRepository
                .findAllByTenantId(tenantId);
        List<org.apache.custos.sharing.core.persistance.model.PermissionType> permissionTypes =
                permissionTypeRepository.findAllByTenantId(tenantId);
        List<org.apache.custos.sharing.core.Entity> arrayList = new ArrayList<>();
        List<SharingMetadata> sharingMetadata = new ArrayList<>();
        entities.forEach(entity -> {
            permissionTypes.forEach(perm -> {
                String permId = perm.getId();
                List<String> sharingList = new ArrayList<>();
                sharingList.add(Constants.DIRECT_CASCADING);
                sharingList.add(Constants.DIRECT_NON_CASCADING);

                List<Sharing> sharings = sharingRepository.
                        findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType(tenantId, entity.getId(),
                                permId, Constants.USER, sharingList);
                sharings.forEach(shr -> {
                    try {
                        sharingMetadata.add(SharingMapper.getSharingMetadata(shr, entity, perm, Constants.USER));
                    } catch (SQLException throwables) {
                        String msg = "Error occurred while transforming entity" + entity.getId();
                        LOGGER.error(msg);
                    }
                });


                List<Sharing> sharingsGroups = sharingRepository.
                        findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType(tenantId, entity.getId(),
                                permId, Constants.GROUP, sharingList);
                sharingsGroups.forEach(shr -> {
                    try {
                        sharingMetadata.add(SharingMapper.getSharingMetadata(shr, entity, perm, Constants.GROUP));
                    } catch (SQLException throwables) {
                        String msg = "Error occurred while transforming entity" + entity.getId();
                        LOGGER.error(msg);
                    }
                });
            });

        });
        return sharingMetadata;
    }

    @Override
    public List<SharingMetadata> getAllSharings(String tenantId, String entityId) throws CustosSharingException {
        List<org.apache.custos.sharing.core.persistance.model.Entity> entities = new ArrayList<>();
        List<SharingMetadata> sharingMetadata = new ArrayList<>();
        List<SharingMetadata> selectedList = new ArrayList<>();
        if (!entityId.isEmpty()) {
            String internalEntityId = entityId + "@" + tenantId;
            Optional<org.apache.custos.sharing.core.persistance.model.Entity> entityOptional = entityRepository.findById(internalEntityId);
            if (entityOptional.isEmpty()) {
                String msg = "Entity " + entityId + " not found ";
                LOGGER.error(msg);
                throw new CustosSharingException(msg);
            }
            entities.add(entityOptional.get());
        } else {
            entities = entityRepository
                    .findAllByTenantId(tenantId);
        }
        entities.forEach(entity -> {
            List<Sharing> userSharings = sharingRepository.
                    findAllByEntityAndOwnerType(tenantId, entity.getId(), Constants.USER);
            List<Sharing> groupSharings = sharingRepository.
                    findAllByEntityAndOwnerType(tenantId, entity.getId(), Constants.GROUP);
            Optional<List<SharingMetadata>> optionalUserSharings = SharingMapper.getSharingMetadata(userSharings);
            Optional<List<SharingMetadata>> optionalGroupSharings = SharingMapper.getSharingMetadata(groupSharings);
            if (optionalUserSharings.isPresent()) {
                sharingMetadata.addAll(optionalUserSharings.get());
            }
            if (optionalGroupSharings.isPresent()) {
                sharingMetadata.addAll(optionalGroupSharings.get());
            }

        });


        //TODO: replace with proper query
        sharingMetadata.forEach(shr -> {
            if (selectedList.isEmpty()) {
                selectedList.add(shr);
            } else {
                AtomicBoolean matched = new AtomicBoolean(false);
                new ArrayList<>(selectedList).forEach(selVar -> {
                    if ((shr.getEntity().getId().equals(selVar.getEntity().getId())
                            && shr.getOwnerId().equals(selVar.getOwnerId()) &&
                            shr.getPermissions(0).getId().equals(selVar.getPermissions(0).getId()))) {
                        matched.set(true);
                    }
                });
                if (!matched.get()) {
                    selectedList.add(shr);
                }
            }
        });

        return selectedList;
    }

    private boolean addCascadingPermissionForEntity
            (org.apache.custos.sharing.core.persistance.model.Entity entity, String internalParentId, String tenantId) {
        List<String> newSharingTypes = new ArrayList<>();
        newSharingTypes.add(Constants.DIRECT_CASCADING);
        newSharingTypes.add(Constants.INDIRECT_CASCADING);

        List<Sharing> sharings = sharingRepository.
                findSharingForEntityOfTenant(tenantId, internalParentId, newSharingTypes);

        if (sharings != null && !sharings.isEmpty()) {

            for (Sharing sharing : sharings) {
                Sharing newShr = SharingMapper.getNewSharing(sharing, tenantId, entity);
                newShr.setSharingType(Constants.INDIRECT_CASCADING);
                sharingRepository.save(newShr);
            }
        }
        return true;

    }

    private List<Sharing> getAllSharingForChildEntities(org.apache.custos.sharing.core.persistance.model.Entity entity,
                                                        org.apache.custos.sharing.core.persistance.model.Entity inheritedEntity,
                                                        List<String> userIds,
                                                        String tenantId,
                                                        org.apache.custos.sharing.core.persistance.model.PermissionType permissionType,
                                                        List<Sharing> sharings,
                                                        String ownerType,
                                                        String sharingType, String sharedBy) {

        List<org.apache.custos.sharing.core.persistance.model.Entity> entities =
                entityRepository.findAllByExternalParentIdAndTenantId(entity.getExternalId(), tenantId);


        if (entities != null && !entities.isEmpty()) {
            for (org.apache.custos.sharing.core.persistance.model.Entity child : entities) {
                for (String userId : userIds) {
                    Sharing sharing = SharingMapper.createSharing(permissionType,
                            child, inheritedEntity, userId, ownerType, sharingType, sharedBy, tenantId);
                    sharings.add(sharing);
                }
                getAllSharingForChildEntities(child, inheritedEntity, userIds, tenantId, permissionType, sharings, ownerType, sharingType, sharedBy);

            }
            return sharings;
        }
        return sharings;
    }
}
