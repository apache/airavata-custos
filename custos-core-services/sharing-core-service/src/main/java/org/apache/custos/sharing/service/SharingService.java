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

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

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
                    responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                    return;
                }

            }

            Optional<org.apache.custos.sharing.persistance.model.EntityType> entityType =
                    entityTypeRepository.findById(entityTypeId);

            if (entityType.isEmpty()) {
                String msg = "Cannot find a Entity Type with given Id " + entity.getId();
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.asRuntimeException());
                return;
            }

            org.apache.custos.sharing.persistance.model.Entity enModel =
                    EntityMapper.createEntity(entity, tenantId, entityType.get());

            entityRepository.save(enModel);

            Optional<org.apache.custos.sharing.persistance.model.PermissionType> optionalPermissionType =
                    permissionTypeRepository.findAllByExternalIdAndTenantId(Constants.OWNER, tenantId);

            if (optionalPermissionType.isPresent()) {
                Sharing sharing = SharingMapper.createSharing(optionalPermissionType.get(),
                        enModel, enModel, entity.getOwnerId(), Constants.DIRECT_CASCADING, tenantId);

                sharingRepository.save(sharing);

            }

            if (internalParentId != null) {
                addCascadingPermissionForEntity(enModel, internalParentId, tenantId);

            }

            Optional<org.apache.custos.sharing.persistance.model.Entity> optionalEntity = entityRepository.findById(enModel.getId());

            List<Sharing> sharings = sharingRepository.findAllByEntityAndSharingTypeAndPermissionType(tenantId,
                    enModel.getId(), Constants.INDIRECT_CASCADING,
                    optionalPermissionType.get().getId());

            org.apache.custos.sharing.persistance.model.Entity savedOne = optionalEntity.get();


            if (sharings != null && sharings.size() > 0) {
                savedOne.setSharedCount(sharings.size());
                entityRepository.save(savedOne);
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
                    permissionTypeRepository.findAllByExternalIdAndTenantId(Constants.OWNER, tenantId);


            if (oldEntity.getExternalParentId() != null &&
                    !oldEntity.getExternalParentId().equals(newEntity.getExternalParentId())) {
                sharingRepository.removeGivenCascadingPermissionsForEntity(tenantId, internalEntityId, Constants.INDIRECT_CASCADING);

            }

            if (newEntity.getExternalParentId() != null) {
                String internalParentId = newEntity.getExternalParentId() + "@" + tenantId;
                addCascadingPermissionForEntity(newEntity, internalParentId, tenantId);

            }


            List<Sharing> sharings = sharingRepository.findAllByEntityAndSharingTypeAndPermissionType(tenantId,
                    internalEntityId, Constants.INDIRECT_CASCADING,
                    optionalPermissionType.get().getId());

            if (sharings != null && sharings.size() > 0) {
                newEntity.setSharedCount(sharings.size());
            }

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
                        org.apache.custos.sharing.service.Status.newBuilder().setStatus(true).build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();
            } else {
                org.apache.custos.sharing.service.Status status =
                        org.apache.custos.sharing.service.Status.newBuilder().setStatus(false).build();
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

            List<org.apache.custos.sharing.persistance.model.Entity> entities = entityRepository.
                    searchEntities(tenantId, request.getSearchCriteriaList());

            List<org.apache.custos.sharing.service.Entity> entityList = new ArrayList<>();

            if (entities != null && entities.isEmpty()) {

                for (org.apache.custos.sharing.persistance.model.Entity entity : entities) {
                    entityList.add(EntityMapper.createEntity(entity));
                }

            }

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


        } catch (Exception ex) {
            String msg = "Error occurred while checking access to  "
                    + request.getEntity().getId() +
                    " in " + request.getTenantId() + " for user" + request.getOwnerId(0) +
                    " reason :" + ex.getMessage();
            LOGGER.error(msg);
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
}
