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

package org.apache.custos.sharing.service;

import io.grpc.stub.StreamObserver;
import org.apache.custos.sharing.mapper.EntityMapper;
import org.apache.custos.sharing.mapper.EntityTypeMapper;
import org.apache.custos.sharing.mapper.PermissionTypeMapper;
import org.apache.custos.sharing.mapper.SharingMapper;
import org.apache.custos.sharing.persistance.model.Sharing;
import org.apache.custos.sharing.persistance.repository.EntityRepository;
import org.apache.custos.sharing.persistance.repository.EntityTypeRepository;
import org.apache.custos.sharing.persistance.repository.PermissionTypeRepository;
import org.apache.custos.sharing.persistance.repository.SharingRepository;
import org.apache.custos.sharing.utils.Constants;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.sql.SQLException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.stream.Collectors;

@GRpcService
public class SharingService extends org.apache.custos.sharing.service.SharingServiceGrpc.SharingServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(SharingService.class);

    @Autowired
    private EntityRepository entityRepository;

    @Autowired
    private EntityTypeRepository entityTypeRepository;

    @Autowired
    private PermissionTypeRepository permissionTypeRepository;

    @Autowired
    private SharingRepository sharingRepository;


    @Override
    public void createEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                                 StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to createEntityType for" + request.getEntityType().getId() +
                    " in" + request.getTenantId());

            String internalId = request.getEntityType().getId() + "@" + request.getTenantId();

            Optional<org.apache.custos.sharing.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);

            if (optionalEntityType.isPresent()) {
                String msg = "Entity type  id" + request.getEntityType().getId() + "is already present";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.ALREADY_EXISTS.asRuntimeException());
                return;
            }

            org.apache.custos.sharing.persistance.model.EntityType
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
    public void updateEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                                 StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to updateEntityType for " + request.getEntityType().getId() +
                    "in " + request.getTenantId());
            String internalId = request.getEntityType().getId() + "@" + request.getTenantId();

            Optional<org.apache.custos.sharing.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);

            if (optionalEntityType.isEmpty()) {
                String msg = "Entity type  id" + request.getEntityType().getId() + "is not present";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.INTERNAL.asRuntimeException());
                return;
            }


            org.apache.custos.sharing.persistance.model.EntityType
                    type = EntityTypeMapper.createEntityType(request.getEntityType(), request.getTenantId());

            type.setCreatedAt(optionalEntityType.get().getCreatedAt());

            entityTypeRepository.save(type);

            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while updating entity type" + request.getEntityType().getId()
                    + "in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void deleteEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                                 StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteEntityType for" + request.getEntityType().getId()
                    + "in " + request.getTenantId());
            String internalId = request.getEntityType().getId() + "@" + request.getTenantId();
            Optional<org.apache.custos.sharing.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);

            if (optionalEntityType.isEmpty()) {
                String msg = "Entity type  id" + request.getEntityType().getId() + "is not present";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.INTERNAL.asRuntimeException());
                return;
            }

            entityTypeRepository.delete(optionalEntityType.get());

            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while deleting entity type" + request.getEntityType().getId()
                    + "in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                              StreamObserver<org.apache.custos.sharing.service.EntityType> responseObserver) {
        try {
            LOGGER.debug("Request received to getEntityType for " + request.getEntityType().getId() +
                    " in " + request.getTenantId());

            String internalId = request.getEntityType().getId() + "@" + request.getTenantId();
            Optional<org.apache.custos.sharing.persistance.model.EntityType> optionalEntityType =
                    entityTypeRepository.findById(internalId);


            if (optionalEntityType.isPresent()) {
                org.apache.custos.sharing.service.EntityType type =
                        EntityTypeMapper.createEntityType(optionalEntityType.get());
                responseObserver.onNext(type);
                responseObserver.onCompleted();

            } else {
                org.apache.custos.sharing.service.EntityType type =
                        org.apache.custos.sharing.service.EntityType.newBuilder().build();
                responseObserver.onNext(type);
                responseObserver.onCompleted();

            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching entity type for " + request.getEntityType().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getEntityTypes(org.apache.custos.sharing.service.SearchRequest request,
                               StreamObserver<org.apache.custos.sharing.service.EntityTypes> responseObserver) {
        try {
            LOGGER.debug("Request received to getEntityTypes " + " in " + request.getTenantId());

            List<org.apache.custos.sharing.persistance.model.EntityType> types =
                    entityTypeRepository.findAllByTenantId(request.getTenantId());

            List<org.apache.custos.sharing.service.EntityType> entityTypes = new ArrayList<>();

            if (!types.isEmpty()) {

                for (org.apache.custos.sharing.persistance.model.EntityType type : types) {

                    entityTypes.add(EntityTypeMapper.createEntityType(type));
                }

            }

            org.apache.custos.sharing.service.EntityTypes entityTy = org.apache.custos.sharing.service.EntityTypes
                    .newBuilder()
                    .addAllTypes(entityTypes)
                    .build();

            responseObserver.onNext(entityTy);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching entity types  in "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void createPermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to createPermissionType with id " + request.getPermissionType().getId()
                    + " in " + request.getTenantId());

            String internalId = request.getPermissionType().getId() + "@" + request.getTenantId();

            Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalEntityType =
                    permissionTypeRepository.findById(internalId);

            if (optionalEntityType.isPresent()) {
                String msg = "Permission type  id" + request.getPermissionType().getId() + "is already present";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.ALREADY_EXISTS.asRuntimeException());
                return;
            }

            org.apache.custos.sharing.persistance.model.PermissionType
                    type = PermissionTypeMapper.createPermissionType(request.getPermissionType(), request.getTenantId());
            permissionTypeRepository.save(type);

            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while creating permission type with id " + request.getPermissionType().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void updatePermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to updatePermissionType with id " + request.getPermissionType().getId()
                    + " in " + request.getTenantId());

            String internalId = request.getPermissionType().getId() + "@" + request.getTenantId();

            Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalEntityType =
                    permissionTypeRepository.findById(internalId);

            if (optionalEntityType.isEmpty()) {
                String msg = "Entity type  id" + request.getPermissionType().getId() + "is not present";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.INTERNAL.asRuntimeException());
                return;
            }

            org.apache.custos.sharing.persistance.model.PermissionType
                    type = PermissionTypeMapper.createPermissionType(request.getPermissionType(), request.getTenantId());

            type.setCreatedAt(optionalEntityType.get().getCreatedAt());

            permissionTypeRepository.save(type);

            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while updating permission type with id " + request.getPermissionType().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deletePermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to deletePermissionType with id " + request.getPermissionType().getId()
                    + " in " + request.getTenantId());
            String internalId = request.getPermissionType().getId() + "@" + request.getTenantId();
            Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalEntityType =
                    permissionTypeRepository.findById(internalId);

            if (optionalEntityType.isEmpty()) {
                String msg = "Entity type  id" + request.getPermissionType().getId() + "is not present";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.INTERNAL.asRuntimeException());
                return;
            }

            permissionTypeRepository.delete(optionalEntityType.get());

            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while deleting permission type with id " + request.getPermissionType().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getPermissionType(org.apache.custos.sharing.service.PermissionTypeRequest request,
                                  StreamObserver<org.apache.custos.sharing.service.PermissionType> responseObserver) {
        try {
            LOGGER.debug("Request received to getPermissionType with id " + request.getPermissionType().getId()
                    + " in " + request.getTenantId());

            String internalId = request.getPermissionType().getId() + "@" + request.getTenantId();
            Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalEntityType =
                    permissionTypeRepository.findById(internalId);


            if (optionalEntityType.isPresent()) {
                org.apache.custos.sharing.service.PermissionType type =
                        PermissionTypeMapper.createPermissionType(optionalEntityType.get());
                responseObserver.onNext(type);
                responseObserver.onCompleted();

            } else {
                org.apache.custos.sharing.service.PermissionType type =
                        org.apache.custos.sharing.service.PermissionType.newBuilder().build();
                responseObserver.onNext(type);
                responseObserver.onCompleted();

            }


        } catch (Exception ex) {
            String msg = "Error occurred while fetching permission type with id " + request.getPermissionType().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getPermissionTypes(org.apache.custos.sharing.service.SearchRequest request,
                                   StreamObserver<org.apache.custos.sharing.service.PermissionTypes> responseObserver) {
        try {
            LOGGER.debug("Request received to getPermissionTypes " + " in " + request.getTenantId());

            List<org.apache.custos.sharing.persistance.model.PermissionType> types =
                    permissionTypeRepository.findAllByTenantId(request.getTenantId());

            List<org.apache.custos.sharing.service.PermissionType> entityTypes = new ArrayList<>();

            if (!types.isEmpty()) {

                for (org.apache.custos.sharing.persistance.model.PermissionType type : types) {

                    entityTypes.add(PermissionTypeMapper.createPermissionType(type));
                }

            }

            org.apache.custos.sharing.service.PermissionTypes entityTy = org.apache.custos.sharing.service.PermissionTypes
                    .newBuilder()
                    .addAllTypes(entityTypes)
                    .build();

            responseObserver.onNext(entityTy);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while fetching permission types"
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void createEntity(org.apache.custos.sharing.service.EntityRequest request,
                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to createEntity with id" + request.getEntity().getId() + " in "
                    + request.getTenantId());


            org.apache.custos.sharing.service.Entity entity = request.getEntity();

            long tenantId = request.getTenantId();

            String entityTypeId = entity.getType() + "@" + tenantId;

            String internalParentId = null;

            if (entity.getParentId() != null && !entity.getParentId().trim().equals("")) {

                internalParentId = entity.getParentId() + "@" + tenantId;


                Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity = entityRepository.
                        findById(internalParentId);

                if (optionalEntity.isEmpty()) {
                    String msg = "Cannot find a parent Entity  with given Id " + entity.getParentId();
                    LOGGER.error(msg);
                    responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                    return;
                }

            }

            Optional<org.apache.custos.sharing.persistance.model.EntityType> entityType =
                    entityTypeRepository.findById(entityTypeId);

            if (entityType.isEmpty()) {
                String msg = "Cannot find a Entity Type with given Id " + entity.getId();
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            org.apache.custos.sharing.persistance.model.Entity enModel =
                    EntityMapper.createEntity(entity, tenantId, entityType.get());

            org.apache.custos.sharing.persistance.model.Entity savedEntity = entityRepository.save(enModel);


            Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalPermissionType =
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


            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while creating Entity with id " + request.getEntity().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void updateEntity(org.apache.custos.sharing.service.EntityRequest request,
                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to updateEntity with id" + request.getEntity().getId() + " in "
                    + request.getTenantId());

            org.apache.custos.sharing.service.Entity entity = request.getEntity();

            long tenantId = request.getTenantId();

            String entityTypeId = entity.getType() + "@" + tenantId;

            String internalEntityId = entity.getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity = entityRepository.findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find a Entity  with given Id " + entity.getId();
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            Optional<org.apache.custos.sharing.persistance.model.EntityType> entityType =
                    entityTypeRepository.findById(entityTypeId);

            if (entityType.isEmpty()) {
                String msg = "Cannot find a Entity Type with given Id " + entity.getType();
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            org.apache.custos.sharing.persistance.model.Entity oldEntity = optionalEntity.get();

            org.apache.custos.sharing.persistance.model.Entity newEntity =
                    EntityMapper.createEntity(entity, tenantId, entityType.get());

            if (!oldEntity.getOwnerId().equals(newEntity.getOwnerId())) {
                String msg = "Entity owner cannot change ";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }


            Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalPermissionType =
                    permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);


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

            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while updating Entity with id " + request.getEntity().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void isEntityExists(org.apache.custos.sharing.service.EntityRequest request,
                               StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to isEntityExists with id" + request.getEntity().getId() + " in "
                    + request.getTenantId());

            long tenantId = request.getTenantId();

            String entityId = request.getEntity().getId();

            String internalId = entityId + "@" + tenantId;


            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity =
                    entityRepository.findById(internalId);

            if (optionalEntity.isEmpty()) {
                org.apache.custos.sharing.service.Status status =
                        org.apache.custos.sharing.service.Status.newBuilder().setStatus(false).build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();
            } else {
                org.apache.custos.sharing.service.Status status =
                        org.apache.custos.sharing.service.Status.newBuilder().setStatus(true).build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Error occurred while checking isEntityExists with id " + request.getEntity().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }


    }

    @Override
    public void getEntity(org.apache.custos.sharing.service.EntityRequest request,
                          StreamObserver<org.apache.custos.sharing.service.Entity> responseObserver) {
        try {
            LOGGER.debug("Request received to getEntity with id" + request.getEntity().getId() + " in "
                    + request.getTenantId());

            long tenantId = request.getTenantId();

            String entityId = request.getEntity().getId();

            String internalId = entityId + "@" + tenantId;


            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity =
                    entityRepository.findById(internalId);

            if (optionalEntity.isEmpty()) {
                responseObserver.onError(io.grpc.Status.NOT_FOUND.
                        withDescription("Entity not found for id " + entityId).asRuntimeException());
            } else {
                responseObserver.onNext(EntityMapper.createEntity(optionalEntity.get()));
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching Entity with id " + request.getEntity().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteEntity(org.apache.custos.sharing.service.EntityRequest request,
                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteEntity with id" + request.getEntity().getId() + " in "
                    + request.getTenantId());

            long tenantId = request.getTenantId();

            String entityId = request.getEntity().getId();

            String internalId = entityId + "@" + tenantId;

            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity =
                    entityRepository.findById(internalId);


            if (optionalEntity.isPresent()) {
                entityRepository.delete(optionalEntity.get());
            }

            org.apache.custos.sharing.service.Status status =
                    org.apache.custos.sharing.service.Status.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while deleting Entity with id " + request.getEntity().getId()
                    + " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void searchEntities(org.apache.custos.sharing.service.SearchRequest request,
                               StreamObserver<org.apache.custos.sharing.service.Entities> responseObserver) {
        try {
            LOGGER.debug("Request received to search entities in tenant"
                    + request.getTenantId());

            long tenantId = request.getTenantId();
            int limit = request.getLimit() == 0 ? -1 : request.getLimit();

            List<org.apache.custos.sharing.persistance.model.Entity> entities = new ArrayList<>();
            HashMap<String, org.apache.custos.sharing.service.Entity> entryMap = new HashMap<>();

            List<String> sharingList = new ArrayList<>();
            sharingList.add(Constants.DIRECT_CASCADING);
            sharingList.add(Constants.DIRECT_NON_CASCADING);

            String initialParentId = null;
            List<String> internalEntityIds = new ArrayList<>();
            for (SearchCriteria searchCriteria : request.getSearchCriteriaList()) {
                if (searchCriteria.getSearchField().equals(EntitySearchField.SHARED_BY)) {
                    String value = searchCriteria.getValue();
                    List<Sharing> sharings = sharingRepository.findAllEntitiesSharedBy(tenantId, value, sharingList);
                    if (sharings != null && !sharings.isEmpty()) {
                        for (Sharing sharing : sharings) {
                            Entity entity = EntityMapper.createEntity(sharing.getEntity());
                            SharingMetadata sharingMetadata = entity.toBuilder().getMetadata();
                            sharingMetadata = sharingMetadata.toBuilder()
                                    .addPermissions(PermissionType.newBuilder()
                                            .setId(sharing.getPermissionType()
                                                    .getExternalId()).build()).build();
                            entity = entity.toBuilder().setMetadata(sharingMetadata).build();
                            entryMap.put(entity.getId(), entity);
                        }

                    }

                    List<org.apache.custos.sharing.service.Entity> entityList = new ArrayList<>(entryMap.values());
                    org.apache.custos.sharing.service.Entities resp = org.apache.custos.sharing.service.Entities
                            .newBuilder()
                            .addAllEntityArray(entityList)
                            .build();
                    responseObserver.onNext(resp);
                    responseObserver.onCompleted();
                    return;

                } else if (searchCriteria.getSearchField().equals(EntitySearchField.PARENT_ID)) {
                    initialParentId = searchCriteria.getValue();
                }
            }

            if (request.getSearchPermBottomUp()) {
                entities = entityRepository.
                        searchEntitiesWithParent(tenantId, initialParentId, limit, request.getOffset());
                List<org.apache.custos.sharing.persistance.model.Entity> totalEntityList = new ArrayList<>();
                totalEntityList.addAll(entities);
                while (!entities.isEmpty()) {
                    List<org.apache.custos.sharing.persistance.model.Entity> entityList = new ArrayList<>();
                    for (org.apache.custos.sharing.persistance.model.Entity entity : entities) {
                        List<org.apache.custos.sharing.persistance.model.Entity> exList = entityRepository.
                                searchEntitiesWithParent(tenantId, entity.getExternalId(), limit, request.getOffset());
                        if (!exList.isEmpty()) {
                            entityList.addAll(exList);
                        }
                    }
                    entities = entityList;
                    totalEntityList.addAll(entities);
                }
                entities = totalEntityList;

            } else {
                entities = entityRepository.
                        searchEntities(tenantId, request.getSearchCriteriaList(), limit, request.getOffset());
            }


            if ((entities != null && !entities.isEmpty())) {

                for (org.apache.custos.sharing.persistance.model.Entity entity : entities) {
                    internalEntityIds.add(entity.getId());
                }

                List<Sharing> sharings = sharingRepository.
                        findAllSharingEntitiesForUsers(tenantId, request.getAssociatingIdsList(), internalEntityIds);
                if (sharings != null && !sharings.isEmpty()) {
                    for (Sharing sharing : sharings) {
                        if (entryMap.containsKey(sharing.getEntity().getId())) {
                            Entity entity = entryMap.get(sharing.getEntity().getId());
                            SharingMetadata sharingMetadata = entity.toBuilder().getMetadata();
                            List<org.apache.custos.sharing.service.PermissionType> permissionTypes = sharingMetadata.getPermissionsList().stream().filter(perm -> {
                                if (perm.getId().equals(sharing.getPermissionType().getExternalId())) {
                                    return true;
                                } else {
                                    return false;
                                }
                            }).collect(Collectors.toList());
                            sharingMetadata = sharingMetadata.toBuilder()
                                    .addPermissions(PermissionType.newBuilder()
                                            .setId(sharing.getPermissionType()
                                                    .getExternalId()).build()).build();
                            entity = entity.toBuilder().setMetadata(sharingMetadata).build();
                            entryMap.put(entity.getId(), entity);

                        } else {
                            Entity entity = EntityMapper.createEntity(sharing.getEntity());
                            SharingMetadata sharingMetadata = SharingMetadata.newBuilder()
                                    .addPermissions(PermissionType.newBuilder()
                                            .setId(sharing.getPermissionType()
                                                    .getExternalId()).build()).build();
                            entity = entity.toBuilder().setMetadata(sharingMetadata).build();
                            entryMap.put(entity.getId(), entity);
                        }

                    }
                }
            }
            List<org.apache.custos.sharing.service.Entity> entityList = new ArrayList<>(entryMap.values());
            org.apache.custos.sharing.service.Entities resp = org.apache.custos.sharing.service.Entities
                    .newBuilder()
                    .addAllEntityArray(entityList)
                    .build();
            responseObserver.onNext(resp);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while search entities with id " +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getListOfSharedUsers(org.apache.custos.sharing.service.SharingRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.SharedOwners> responseObserver) {
        try {
            LOGGER.debug("Request received to getListOfSharedUsers " + request.getTenantId() + " for entity "
                    + request.getEntity().getId());
            List<Sharing> sharings = null;
            long tenantId = request.getTenantId();

            String entityId = request.getEntity().getId();

            String internalEntityId = entityId + "@" + tenantId;


            String permisstionType = request.getPermissionType().getId();

            String internalPermissionTypeId = request.getPermissionType().getId() + "@" + tenantId;
            Optional<org.apache.custos.sharing.persistance.model.PermissionType> permissionType =
                    permissionTypeRepository.findById(internalPermissionTypeId);

            if (permissionType.isEmpty()) {
                String msg = "Cannot find permission type" + permisstionType;
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity = entityRepository.
                    findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find given entity" + entityId;
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalPermissionType =
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

            org.apache.custos.sharing.service.SharedOwners owners = SharingMapper.getSharedOwners(sharings);

            responseObserver.onNext(owners);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching  getListOfSharedUsers  for entity "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getListOfDirectlySharedUsers(org.apache.custos.sharing.service.SharingRequest request,
                                             StreamObserver<org.apache.custos.sharing.service.SharedOwners> responseObserver) {
        try {
            LOGGER.debug("Request received to getListOfDirectlySharedUsers " + request.getTenantId() + " for entity "
                    + request.getEntity().getId());

            long tenantId = request.getTenantId();

            String entityId = request.getEntity().getId();

            String internalEntityId = entityId + "@" + tenantId;

            String permissionType = request.getPermissionType().getId();

            String internalPermissionTypeId = request.getPermissionType().getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.persistance.model.PermissionType> oPT =
                    permissionTypeRepository.findById(internalPermissionTypeId);

            if (oPT.isEmpty()) {
                String msg = "Cannot find permission type" + permissionType;
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity = entityRepository.
                    findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find given entity" + entityId;
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }


            List<String> sharingList = new ArrayList<>();
            sharingList.add(Constants.DIRECT_CASCADING);
            sharingList.add(Constants.DIRECT_NON_CASCADING);

            List<Sharing> sharings = sharingRepository.
                    findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType(tenantId, internalEntityId,
                            internalPermissionTypeId, Constants.USER, sharingList);
            org.apache.custos.sharing.service.SharedOwners owners = SharingMapper.getSharedOwners(sharings);

            responseObserver.onNext(owners);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while fetching  directly shared users  for entity "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getListOfSharedGroups(org.apache.custos.sharing.service.SharingRequest request,
                                      StreamObserver<org.apache.custos.sharing.service.SharedOwners> responseObserver) {
        try {
            LOGGER.debug("Request received to getListOfSharedGroups " + request.getTenantId() + " for entity "
                    + request.getEntity().getId());
            long tenantId = request.getTenantId();

            String entityId = request.getEntity().getId();

            String internalEntityId = entityId + "@" + tenantId;
            List<Sharing> sharings = null;

            String permisstionType = request.getPermissionType().getId();

            String internalPermissionTypeId = request.getPermissionType().getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.persistance.model.PermissionType> permissionType =
                    permissionTypeRepository.findById(internalPermissionTypeId);

            if (permissionType.isEmpty()) {
                String msg = "Cannot find permission type" + permisstionType;
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity = entityRepository.
                    findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find given entity" + entityId;
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            sharings = sharingRepository.
                    findAllByEntityAndPermissionTypeAndOwnerType(tenantId, internalEntityId,
                            internalPermissionTypeId, Constants.GROUP);

            org.apache.custos.sharing.service.SharedOwners owners = SharingMapper.getSharedOwners(sharings);

            responseObserver.onNext(owners);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while fetching   shared groups  for entity "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getListOfDirectlySharedGroups(org.apache.custos.sharing.service.SharingRequest request,
                                              StreamObserver<org.apache.custos.sharing.service.SharedOwners> responseObserver) {
        try {
            LOGGER.debug("Request received to getListOfDirectlySharedGroups " + request.getTenantId() + " for entity "
                    + request.getEntity().getId());

            long tenantId = request.getTenantId();

            String entityId = request.getEntity().getId();

            String internalEntityId = entityId + "@" + tenantId;

            String permisstionType = request.getPermissionType().getId();

            String internalPermissionTypeId = request.getPermissionType().getId() + "@" + tenantId;

            Optional<org.apache.custos.sharing.persistance.model.PermissionType> permissionType =
                    permissionTypeRepository.findById(internalPermissionTypeId);

            if (permissionType.isEmpty()) {
                String msg = "Cannot find permission type" + permisstionType;
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity = entityRepository.
                    findById(internalEntityId);

            if (optionalEntity.isEmpty()) {
                String msg = "Cannot find given entity" + entityId;
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            List<String> sharingList = new ArrayList<>();
            sharingList.add(Constants.DIRECT_CASCADING);
            sharingList.add(Constants.DIRECT_NON_CASCADING);

            List<Sharing> sharings = sharingRepository.
                    findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType(tenantId, internalEntityId,
                            internalPermissionTypeId, Constants.GROUP, sharingList);

            org.apache.custos.sharing.service.SharedOwners owners = SharingMapper.getSharedOwners(sharings);

            responseObserver.onNext(owners);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while fetching   directly shared groups  for entity "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void shareEntityWithUsers(org.apache.custos.sharing.service.SharingRequest request,
                                     StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to shareEntityWithUsers " + request.getTenantId() + " for entity "
                    + request.getEntity().getId());

            shareEntity(request, responseObserver, Constants.USER);

        } catch (Exception ex) {
            String msg = "Error occurred while sharing   entity with user  for entity "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void shareEntityWithGroups(org.apache.custos.sharing.service.SharingRequest request,
                                      StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to shareEntityWithGroups " + request.getTenantId() + " for entity "
                    + request.getEntity().getId());
            shareEntity(request, responseObserver, Constants.GROUP);

        } catch (Exception ex) {
            String msg = "Error occurred while sharing   entity with groups  for entity "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void revokeEntitySharingFromUsers(org.apache.custos.sharing.service.SharingRequest request,
                                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to revokeEntitySharingFromUsers " + request.getTenantId() + " for entity "
                    + request.getEntity().getId());


            revokePermission(request, responseObserver);

        } catch (Exception ex) {
            String msg = "Error occurred while revoking   entity from users  for entity "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void revokeEntitySharingFromGroups(org.apache.custos.sharing.service.SharingRequest request,
                                              StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to revokeEntitySharingFromGroups " + request.getTenantId() + " for entity "
                    + request.getEntity().getId());
            revokePermission(request, responseObserver);

        } catch (Exception ex) {
            String msg = "Error occurred while revoking   entity from groups  for entity "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void userHasAccess(org.apache.custos.sharing.service.SharingRequest request,
                              StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to userHasAccess in " + request.getTenantId() + " for entity "
                    + request.getEntity().getId() + " for user " + request.getOwnerId(0));

            long tenantId = request.getTenantId();
            List<String> owners = request.getOwnerIdList();
            String permission = request.getPermissionType().getId();

            String entityId = request.getEntity().getId();

            String internalPermissionId = permission + "@" + tenantId;

            String internalEntityId = entityId + "@" + tenantId;

            Optional<org.apache.custos.sharing.persistance.model.PermissionType> permissionType =
                    permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);

            List<String> permissionTypes = new ArrayList<>();
            permissionTypes.add(internalPermissionId);
            permissionTypes.add(permissionType.get().getId());


            List<Sharing> sharings = sharingRepository.findAllSharingOfEntityForGroupsUnderPermissions(tenantId,
                    internalEntityId, permissionTypes, owners);

            if (sharings != null && !sharings.isEmpty()) {

                org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                        .newBuilder()
                        .setStatus(true)
                        .build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();
            } else {
                org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                        .newBuilder()
                        .setStatus(false)
                        .build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            }


        } catch (Exception ex) {
            String msg = "Error occurred while checking access to  "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " for user" + request.getOwnerId(0) +
                    " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getAllDirectSharings(SharingRequest request, StreamObserver<GetAllDirectSharingsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllEntities in " + request.getTenantId());
            List<org.apache.custos.sharing.persistance.model.Entity> entities = entityRepository
                    .findAllByTenantId(request.getTenantId());
            List<org.apache.custos.sharing.persistance.model.PermissionType> permissionTypes =
                    permissionTypeRepository.findAllByTenantId(request.getTenantId());
            List<org.apache.custos.sharing.service.Entity> arrayList = new ArrayList<>();
            List<SharingMetadata> sharingMetadata = new ArrayList<>();
            entities.forEach(entity -> {
                permissionTypes.forEach(perm -> {
                    String permId = perm.getId();
                    List<String> sharingList = new ArrayList<>();
                    sharingList.add(Constants.DIRECT_CASCADING);
                    sharingList.add(Constants.DIRECT_NON_CASCADING);

                    List<Sharing> sharings = sharingRepository.
                            findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType(request.getTenantId(), entity.getId(),
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
                            findAllByEntityAndPermissionTypeAndOwnerTypeAndSharingType(request.getTenantId(), entity.getId(),
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
            GetAllDirectSharingsResponse response = GetAllDirectSharingsResponse
                    .newBuilder().addAllSharedData(sharingMetadata).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching all entities related to " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }


    @Override
    public void getAllSharings(SharingRequest request, StreamObserver<GetAllSharingsResponse> responseObserver) {
        try {
            List<org.apache.custos.sharing.service.Entity> arrayList = new ArrayList<>();
            List<org.apache.custos.sharing.persistance.model.Entity> entities = new ArrayList<>();
            List<SharingMetadata> sharingMetadata = new ArrayList<>();
            List<SharingMetadata> selectedList = new ArrayList<>();
            if (request.hasEntity() && !request.getEntity().getId().isEmpty()) {
                arrayList.add(request.getEntity());
                String entityId = request.getEntity().getId() + "@" + request.getTenantId();
                Optional<org.apache.custos.sharing.persistance.model.Entity> entityOptional = entityRepository.findById(entityId);
                if (entityOptional.isEmpty()) {
                    String msg = "Entity " + request.getEntity().getId() + " not found ";
                    LOGGER.error(msg);
                    responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                    return;
                }
                entities.add(entityOptional.get());
            } else {
                entities = entityRepository
                        .findAllByTenantId(request.getTenantId());
            }
            entities.forEach(entity -> {
                List<Sharing> userSharings = sharingRepository.
                        findAllByEntityAndOwnerType(request.getTenantId(), entity.getId(), Constants.USER);
                List<Sharing> groupSharings = sharingRepository.
                        findAllByEntityAndOwnerType(request.getTenantId(), entity.getId(), Constants.GROUP);
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


            GetAllSharingsResponse response = GetAllSharingsResponse
                    .newBuilder().addAllSharedData(selectedList).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching all sharings " + request.getTenantId() + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    private boolean addCascadingPermissionForEntity
            (org.apache.custos.sharing.persistance.model.Entity entity, String internalParentId, long tenantId) {
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


    private List<Sharing> getAllSharingForChildEntities(org.apache.custos.sharing.persistance.model.Entity entity,
                                                        org.apache.custos.sharing.persistance.model.Entity inheritedEntity,
                                                        List<String> userIds,
                                                        long tenantId,
                                                        org.apache.custos.sharing.persistance.model.PermissionType permissionType,
                                                        List<Sharing> sharings,
                                                        String ownerType,
                                                        String sharingType, String sharedBy) {

        List<org.apache.custos.sharing.persistance.model.Entity> entities =
                entityRepository.findAllByExternalParentIdAndTenantId(entity.getExternalId(), tenantId);


        if (entities != null && !entities.isEmpty()) {
            for (org.apache.custos.sharing.persistance.model.Entity child : entities) {
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


    private void shareEntity(org.apache.custos.sharing.service.SharingRequest request,
                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver, String ownerType) {
        long tenantId = request.getTenantId();

        List<String> userIds = request.getOwnerIdList();

        String permissionType = request.getPermissionType().getId();

        String internalPermissionId = permissionType + "@" + tenantId;

        boolean cascade = request.getCascade();

        String entityId = request.getEntity().getId();

        String internalEntityId = entityId + "@" + tenantId;


        Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalPermissionType =
                permissionTypeRepository.findById(internalPermissionId);

        if (optionalPermissionType.isEmpty()) {
            String msg = "Permission type not found";
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            return;
        }

        Optional<org.apache.custos.sharing.persistance.model.Entity> entityOptional =
                entityRepository.findById(internalEntityId);

        if (entityOptional.isEmpty()) {
            String msg = " Entity with Id " + entityId + "  not found";
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            return;
        }


        Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalOwnerPermissionType =
                permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);

        if (optionalOwnerPermissionType.get().getId().equals(internalPermissionId)) {
            String msg = "Owner permission type can not be assigned";
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.PERMISSION_DENIED.withDescription(msg).asRuntimeException());
            return;
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
                    entityOptional.get(), entityOptional.get(), userId, ownerType, sharingType, request.getSharedBy(),
                    tenantId);
            sharings.add(sharing);
        }

        if (cascade) {
            List<Sharing> childSharings = new ArrayList<>();
            childSharings = getAllSharingForChildEntities(entityOptional.get(), entityOptional.get(), userIds, tenantId,
                    optionalPermissionType.get(), childSharings, ownerType,
                    Constants.INDIRECT_CASCADING, request.getSharedBy());
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
                List<org.apache.custos.sharing.persistance.model.PermissionType> existingPermissionTypes =
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
            org.apache.custos.sharing.persistance.model.Entity entity = entityOptional.get();
            if (newSharings != null && newSharings.size() > 0) {
                entity.setSharedCount(newSharings.size());
                entityRepository.save(entity);
            }


        }

        org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                .newBuilder()
                .setStatus(true)
                .build();
        responseObserver.onNext(status);
        responseObserver.onCompleted();

    }

    private void revokePermission(org.apache.custos.sharing.service.SharingRequest request,
                                  StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        long tenantId = request.getTenantId();

        String entityId = request.getEntity().getId();

        String internalEntityId = entityId + "@" + tenantId;

        String permissionType = request.getPermissionType().getId();

        String internalPermissionType = permissionType + "@" + tenantId;

        List<String> usersList = request.getOwnerIdList();

        Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalPermissionType =
                permissionTypeRepository.findById(internalPermissionType);

        if (optionalPermissionType.isEmpty()) {
            String msg = "Permission type not found";
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            return;
        }

        Optional<org.apache.custos.sharing.persistance.model.Entity> entityOptional =
                entityRepository.findById(internalEntityId);

        if (entityOptional.isEmpty()) {
            String msg = " Entity with Id " + entityId + "  not found";
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            return;
        }


        Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalOwnerPermissionType =
                permissionTypeRepository.findByExternalIdAndTenantId(Constants.OWNER, tenantId);

        if (optionalOwnerPermissionType.get().getId().equals(internalPermissionType)) {
            String msg = "Owner permission type can not be assigned";
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.PERMISSION_DENIED.withDescription(msg).asRuntimeException());
            return;
        }

        List<String> checkTypes = new ArrayList<>();
        checkTypes.add(Constants.INDIRECT_CASCADING);
        checkTypes.add(Constants.DIRECT_CASCADING);
        checkTypes.add(Constants.DIRECT_NON_CASCADING);


        for (String userId : usersList) {

            LOGGER.debug("deleting " + userId + ":" + internalEntityId + ":" + internalPermissionType + ":" + tenantId);
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

            List<Sharing> exSharings = sharingRepository.findSharingForEntityOfTenant(tenantId, internalEntityId, checkTypes);
            if (!exSharings.isEmpty()) {
                sharingRepository.
                        deleteAllByEntityIdAndPermissionTypeIdAndAssociatingIdAndTenantIdAndInheritedParentId(
                                internalEntityId,
                                internalPermissionType,
                                userId,
                                tenantId,
                                exSharings.get(0).getInheritedParent().getId());
                revokeCascadingPermissionForChildEntities(internalEntityId,
                        internalPermissionType,
                        userId,
                        tenantId,
                        exSharings.get(0).getInheritedParent().getId());
            }

        }

        List<Sharing> newSharings = sharingRepository.findAllByEntityAndSharingType(tenantId,
                internalEntityId, checkTypes);
        org.apache.custos.sharing.persistance.model.Entity entity = entityOptional.get();
        if (newSharings != null && newSharings.size() > 0) {
            entity.setSharedCount(newSharings.size());
            entityRepository.save(entity);
        }

        org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                .newBuilder()
                .setStatus(true)
                .build();
        responseObserver.onNext(status);
        responseObserver.onCompleted();
    }


    private boolean revokeCascadingPermissionForChildEntities
            (String entityId, String internalPermissionType, String userId, long tenantId, String inheritedParentId) {
        List<String> newSharingTypes = new ArrayList<>();
        newSharingTypes.add(Constants.INDIRECT_CASCADING);

        List<org.apache.custos.sharing.persistance.model.Entity> entityList = entityRepository
                .findAllByExternalParentIdAndTenantId(entityId, tenantId);

        entityList.forEach(entity -> {
            sharingRepository.
                    deleteAllByEntityIdAndPermissionTypeIdAndAssociatingIdAndTenantIdAndInheritedParentId(
                            entity.getId(),
                            internalPermissionType,
                            userId,
                            tenantId,
                            inheritedParentId);

            revokeCascadingPermissionForChildEntities(entity.getId(),
                    internalPermissionType, userId, tenantId, inheritedParentId);
        });

        return true;
    }
}
