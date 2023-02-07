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
import org.apache.custos.sharing.core.Entity;
import org.apache.custos.sharing.core.EntityType;
import org.apache.custos.sharing.core.PermissionType;
import org.apache.custos.sharing.core.SharingMetadata;
import org.apache.custos.sharing.core.exceptions.CustosSharingException;
import org.apache.custos.sharing.core.impl.SharingImpl;
import org.apache.custos.sharing.core.utils.Constants;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.List;
import java.util.Optional;

@GRpcService
public class SharingService extends org.apache.custos.sharing.service.SharingServiceGrpc.SharingServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(SharingService.class);

    @Autowired
    SharingImpl sharingCore;

    @Override
    public void createEntityType(org.apache.custos.sharing.service.EntityTypeRequest request,
                                 StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received method: createEntityType Id:" + request.getEntityType().getId() +
                    "  tenant: " + request.getTenantId());

            sharingCore.createEntityType(String.valueOf(request.getTenantId()), request.getEntityType());

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
            LOGGER.debug("Request received method: updateEntityType Id: " + request.getEntityType().getId() +
                    " tenant: " + request.getTenantId());

            sharingCore.updateEntityType(String.valueOf(request.getTenantId()), request.getEntityType());

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
            LOGGER.debug("Request received method: deleteEntityType Id:" + request.getEntityType().getId()
                    + " tenant: " + request.getTenantId());


            sharingCore.deleteEntityType(String.valueOf(request.getTenantId()), request.getEntityType().getId());


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
                              StreamObserver<org.apache.custos.sharing.core.EntityType> responseObserver) {
        try {
            LOGGER.debug("Request received method: getEntityType Id:" + request.getEntityType().getId() +
                    "  tenant: " + request.getTenantId());

            Optional<org.apache.custos.sharing.core.EntityType> entityType = sharingCore.
                    getEntityType(String.valueOf(request.getTenantId()), request.getEntityType().getId());

            if (entityType.isPresent()) {
                responseObserver.onNext(entityType.get());
                responseObserver.onCompleted();

            } else {
                responseObserver.onNext(org.apache.custos.sharing.core.EntityType.newBuilder().build());
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
            LOGGER.debug("Request received method: getEntityTypes " + " Id: " + request.getTenantId());

            List<EntityType> entityTypes = sharingCore.getEntityTypes(String.valueOf(request.getTenantId()));

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
            LOGGER.debug("Request received method: createPermissionType " + " Id: " + request.getPermissionType().getId()
                    + " tenant: " + request.getTenantId());

            sharingCore.createPermissionType(request.getPermissionType(), String.valueOf(request.getTenantId()));

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
            LOGGER.debug("Request received method:  updatePermissionType with Id: " + request.getPermissionType().getId()
                    + " tenant: " + request.getTenantId());

            sharingCore.updatePermissionType(request.getPermissionType(), String.valueOf(request.getTenantId()));

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
            LOGGER.debug("Request received method: deletePermissionType with Id: " + request.getPermissionType().getId()
                    + " tenant: " + request.getTenantId());

            sharingCore.deletePermissionType(String.valueOf(request.getTenantId()), request.getPermissionType().getId());

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
                                  StreamObserver<org.apache.custos.sharing.core.PermissionType> responseObserver) {
        try {
            LOGGER.debug("Request received method: getPermissionType with Id " + request.getPermissionType().getId()
                    + " tenant: " + request.getTenantId());

            Optional<org.apache.custos.sharing.core.PermissionType> optionalPermissionType = sharingCore.
                    getPermissionType(String.valueOf(request.getTenantId()), request.getPermissionType().getId());


            if (optionalPermissionType.isPresent()) {
                responseObserver.onNext(optionalPermissionType.get());
                responseObserver.onCompleted();

            } else {
                responseObserver.onNext(org.apache.custos.sharing.core.PermissionType.newBuilder().build());
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
            LOGGER.debug("Request received method: getPermissionTypes " + " tenant: " + request.getTenantId());


            List<PermissionType> permissionTypes = sharingCore
                    .getPermissionTypes(String.valueOf(request.getTenantId()));

            org.apache.custos.sharing.service.PermissionTypes entityTy = org.apache.custos.sharing.service.PermissionTypes
                    .newBuilder()
                    .addAllTypes(permissionTypes)
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
            LOGGER.debug("Request received method: createEntity with Id" + request.getEntity().getId() + " tenant: "
                    + request.getTenantId());

            sharingCore.createEntity(request.getEntity(), String.valueOf(request.getTenantId()));

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
            LOGGER.debug("Request received method: updateEntity with id" + request.getEntity().getId() + " tenant: "
                    + request.getTenantId());

            sharingCore.updateEntity(request.getEntity(), String.valueOf(request.getTenantId()));

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
            LOGGER.debug("Request received method: isEntityExists with id: " + request.getEntity().getId() + " tenant: "
                    + request.getTenantId());

            boolean value = sharingCore.isEntityExists(String.valueOf(request.getTenantId()), request.getEntity().getId());

            if (!value) {
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
                          StreamObserver<org.apache.custos.sharing.core.Entity> responseObserver) {
        try {
            LOGGER.debug("Request received method: getEntity with id" + request.getEntity().getId() + " tenant: "
                    + request.getTenantId());

            Optional<Entity> optionalEntity = sharingCore
                    .getEntity(String.valueOf(request.getTenantId()), request.getEntity().getId());

            if (optionalEntity.isEmpty()) {
                responseObserver.onError(io.grpc.Status.NOT_FOUND.
                        withDescription("Entity not found for id " + request.getEntity().getId()).asRuntimeException());
            } else {
                responseObserver.onNext(optionalEntity.get());
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
            LOGGER.debug("Request received method: deleteEntity with id" + request.getEntity().getId() + " tenant:"
                    + request.getTenantId());

            sharingCore.deleteEntity(String.valueOf(request.getTenantId()), request.getEntity().getId());

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
            LOGGER.debug("Request received method: search entities in tenant:"
                    + request.getTenantId());

            List<Entity> entities = sharingCore.searchEntities(String.valueOf(request.getTenantId()),
                    request.getSearchCriteriaList(), request.getAssociatingIdsList(),
                    request.getLimit(), request.getOffset(), request.getSearchPermBottomUp());

            org.apache.custos.sharing.service.Entities resp = org.apache.custos.sharing.service.Entities
                    .newBuilder()
                    .addAllEntityArray(entities)
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
            LOGGER.debug("Request received method: getListOfSharedUsers " + request.getTenantId() + " for entity "
                    + request.getEntity().getId() + "  tenant: " + request.getTenantId());

            List<String> users = sharingCore.getListOfSharedUsers(String.valueOf(request.getTenantId()),
                    request.getEntity().getId(), request.getPermissionType().getId());

            responseObserver.onNext(SharedOwners.newBuilder().addAllOwnerIds(users).build());
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
            LOGGER.debug("Request received method: getListOfDirectlySharedUsers  tenant: " + request.getTenantId() + "  entity: "
                    + request.getEntity().getId());

            List<String> users = sharingCore.getListOfDirectlySharedUsers(String.valueOf(request.getTenantId()),
                    request.getEntity().getId(), request.getPermissionType().getId());

            responseObserver.onNext(SharedOwners.newBuilder().addAllOwnerIds(users).build());
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
            LOGGER.debug("Request received method: getListOfSharedGroups  tenant:" + request.getTenantId() +
                    " for entity " + request.getEntity().getId());

            List<String> groups = sharingCore.getListOfSharedGroups(String.valueOf(request.getTenantId()),
                    request.getEntity().getId(), request.getPermissionType().getId());

            responseObserver.onNext(SharedOwners.newBuilder().addAllOwnerIds(groups).build());
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
            LOGGER.debug("Request received method: getListOfDirectlySharedGroups tenant: " + request.getTenantId() +
                    " for entity " + request.getEntity().getId());

            List<String> groups = sharingCore.getListOfDirectlySharedGroups(String.valueOf(request.getTenantId()),
                    request.getEntity().getId(), request.getPermissionType().getId());

            responseObserver.onNext(SharedOwners.newBuilder().addAllOwnerIds(groups).build());
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
            LOGGER.debug("Request received method: shareEntityWithUsers tenant: " + request.getTenantId() +
                    " for entity Id " + request.getEntity().getId());

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
            LOGGER.debug("Request received method: shareEntityWithGroups " + request.getTenantId() + " for entity "
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
            LOGGER.debug("Request received method: revokeEntitySharingFromUsers " + request.getTenantId() + " for entity "
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
            LOGGER.debug("Request received method: revokeEntitySharingFromGroups " + request.getTenantId() + " for entity "
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
            LOGGER.debug("Request received method: userHasAccess tenant: " + request.getTenantId() + " for entity: "
                    + request.getEntity().getId() + " for user: " + request.getOwnerId(0));

            boolean access = sharingCore.userHasAccess(String.valueOf(request.getTenantId()),
                    request.getEntity().getId(), request.getPermissionType().getId(), request.getOwnerId(0));

            org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                    .newBuilder()
                    .setStatus(access)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

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

            List<SharingMetadata> sharingMetadata = sharingCore.getAllDirectSharings(String.valueOf(request.getTenantId()));
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

            List<SharingMetadata> metadataList = sharingCore.
                    getAllSharings(String.valueOf(request.getTenantId()), request.getEntity().getId());

            GetAllSharingsResponse response = GetAllSharingsResponse
                    .newBuilder().addAllSharedData(metadataList).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching all sharings " + request.getTenantId() + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    private void shareEntity(org.apache.custos.sharing.service.SharingRequest request,
                             StreamObserver<org.apache.custos.sharing.service.Status> responseObserver, String ownerType) throws CustosSharingException {
        sharingCore.shareEntity(String.valueOf(request.getTenantId()), request.getEntity().getId(),
                request.getPermissionType().getId(), request.getOwnerIdList(), request.getCascade(), ownerType, request.getSharedBy());

        org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                .newBuilder()
                .setStatus(true)
                .build();
        responseObserver.onNext(status);
        responseObserver.onCompleted();

    }

    private void revokePermission(org.apache.custos.sharing.service.SharingRequest request,
                                  StreamObserver<org.apache.custos.sharing.service.Status> responseObserver) throws CustosSharingException {

        boolean revoke = sharingCore.revokePermission(String.valueOf(request.getTenantId()), request.getEntity().getId(),
                request.getPermissionType().getId(), request.getOwnerIdList());

        org.apache.custos.sharing.service.Status status = org.apache.custos.sharing.service.Status
                .newBuilder()
                .setStatus(revoke)
                .build();
        responseObserver.onNext(status);
        responseObserver.onCompleted();
    }


}
