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

package org.apache.custos.sharing.management.service;

import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.admin.client.IamAdminServiceClient;
import org.apache.custos.iam.service.CheckingResponse;
import org.apache.custos.iam.service.UserRepresentation;
import org.apache.custos.iam.service.UserSearchMetadata;
import org.apache.custos.iam.service.UserSearchRequest;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.identity.service.GetUserManagementSATokenRequest;
import org.apache.custos.integration.services.commons.utils.InterServiceModelMapper;
import org.apache.custos.sharing.client.SharingClient;
import org.apache.custos.sharing.management.exceptions.SharingException;
import org.apache.custos.sharing.management.service.SharingManagementServiceGrpc.SharingManagementServiceImplBase;
import org.apache.custos.sharing.service.Status;
import org.apache.custos.sharing.service.*;
import org.apache.custos.user.profile.client.UserProfileClient;
import org.apache.custos.user.profile.service.*;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.List;

@GRpcService
public class SharingManagementService extends SharingManagementServiceImplBase {

    private static Logger LOGGER = LoggerFactory.getLogger(SharingManagementService.class);

    @Autowired
    private UserProfileClient userProfileClient;

    @Autowired
    private IamAdminServiceClient iamAdminServiceClient;

    @Autowired
    private SharingClient sharingClient;

    @Autowired
    private IdentityClient identityClient;


    @Override
    public void createEntityType(EntityTypeRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to createEntityType in tenant " + request.getTenantId() +
                    "  with entity type Id " + request.getEntityType().getId());

            Status status = sharingClient.createEntityType(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at createEntityType " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateEntityType(EntityTypeRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to updateEntityType in tenant " + request.getTenantId() +
                    "  with entity type Id " + request.getEntityType().getId());

            Status status = sharingClient.updateEntityType(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at updateEntityType " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteEntityType(EntityTypeRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteEntityType in tenant " + request.getTenantId() +
                    "  with entity type  Id " + request.getEntityType().getId());

            Status status = sharingClient.deleteEntityType(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at deleteEntityType " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getEntityType(EntityTypeRequest request, StreamObserver<EntityType> responseObserver) {
        try {
            LOGGER.debug("Request received to getEntityType in tenant " + request.getTenantId() +
                    "  with entity  type Id " + request.getEntityType().getId());

            EntityType type = sharingClient.getEntityType(request);
            responseObserver.onNext(type);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getEntityType " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getEntityTypes(SearchRequest request, StreamObserver<EntityTypes> responseObserver) {
        try {
            LOGGER.debug("Request received to getEntityTypes in tenant " + request.getTenantId());

            EntityTypes entityTypes = sharingClient.getEntityTypes(request);
            responseObserver.onNext(entityTypes);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getEntityTypes " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void createPermissionType(PermissionTypeRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to createPermissionType in tenant " + request.getTenantId() +
                    "  with permission id  " + request.getPermissionType().getId());

            Status status = sharingClient.createPermissionType(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at createPermissionType " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updatePermissionType(PermissionTypeRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to updatePermissionType in tenant " + request.getTenantId() +
                    "  with permission id   " + request.getPermissionType().getId());
            Status status = sharingClient.updatePermissionType(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at updatePermissionType " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deletePermissionType(PermissionTypeRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to deletePermissionType in tenant " + request.getTenantId() +
                    "  with permission id  " + request.getPermissionType().getId());
            Status status = sharingClient.deletePermissionType(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at deletePermissionType " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getPermissionType(PermissionTypeRequest request, StreamObserver<PermissionType> responseObserver) {
        try {
            LOGGER.debug("Request received to getPermissionType in tenant " + request.getTenantId() +
                    "  with permission id  " + request.getPermissionType().getId());
            PermissionType permissionType = sharingClient.getPermissionType(request);
            responseObserver.onNext(permissionType);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getPermissionType " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getPermissionTypes(SearchRequest request, StreamObserver<PermissionTypes> responseObserver) {
        try {
            LOGGER.debug("Request received to getPermissionTypes in tenant " + request.getTenantId());
            PermissionTypes permissionTypes = sharingClient.getPermissionTypes(request);
            responseObserver.onNext(permissionTypes);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getPermissionTypes " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void createEntity(EntityRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to createEntity in tenant " + request.getTenantId() +
                    "  with entity id  " + request.getEntity().getId());

            long tenantId = request.getTenantId();

            String username = request.getEntity().getOwnerId();

            String clientId = request.getClientId();

            String clientSec = request.getClientSec();

            String profileId = validateAndGetUserProfile(username, clientId, clientSec, tenantId);


            Entity entity = request.getEntity();
            entity = entity.toBuilder().setOwnerId(profileId).build();

            request = request.toBuilder().setEntity(entity).build();

            Status status = sharingClient.createEntity(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at createEntity " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateEntity(EntityRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to updateEntity in tenant " + request.getTenantId() +
                    "  with entity id  " + request.getEntity().getId());

            long tenantId = request.getTenantId();

            String username = request.getEntity().getOwnerId();

            String clientId = request.getClientId();

            String clientSec = request.getClientSec();

            String profileId = validateAndGetUserProfile(username, clientId, clientSec, tenantId);


            Entity entity = request.getEntity();
            entity = entity.toBuilder().setOwnerId(profileId).build();

            request = request.toBuilder().setEntity(entity).build();

            Status status = sharingClient.updateEntity(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at updateEntity " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isEntityExists(EntityRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to isEntityExists in tenant " + request.getTenantId() +
                    "  with entity  id " + request.getEntity().getId());

            Status status = sharingClient.isEntityExists(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at isEntityExists " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getEntity(EntityRequest request, StreamObserver<Entity> responseObserver) {
        try {
            LOGGER.debug("Request received to getEntity in tenant " + request.getTenantId() +
                    "  with  entity  Id  " + request.getEntity().getId());

            Entity entity = sharingClient.getEntity(request);
            responseObserver.onNext(entity);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at getListOfSharedUsers " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteEntity(EntityRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteEntity in tenant " + request.getTenantId() +
                    "  with  entity Id " + request.getEntity().getId());

            Status status = sharingClient.deleteEntity(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at deleteEntity " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void searchEntities(SearchRequest request, StreamObserver<Entities> responseObserver) {
        try {
            LOGGER.debug("Request received to searchEntities in tenant " + request.getTenantId());

            Entities entities = sharingClient.searchEntities(request);
            responseObserver.onNext(entities);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at searchEntities " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getListOfSharedUsers(SharingRequest request, StreamObserver<SharedOwners> responseObserver) {
        try {
            LOGGER.debug("Request received to getListOfSharedUsers in tenant " + request.getTenantId() +
                    "  with entity Id  " + request.getEntity().getId());

            SharedOwners owners = sharingClient.getListOfSharedUsers(request);
            responseObserver.onNext(owners);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getListOfSharedUsers " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getListOfDirectlySharedUsers(SharingRequest request, StreamObserver<SharedOwners> responseObserver) {
        try {
            LOGGER.debug("Request received to getListOfDirectlySharedUsers in tenant " + request.getTenantId() +
                    "  for entity  " + request.getEntity().getId());

            SharedOwners owners = sharingClient.getListOfDirectlySharedUsers(request);
            responseObserver.onNext(owners);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getListOfDirectlySharedUsers " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getListOfSharedGroups(SharingRequest request, StreamObserver<SharedOwners> responseObserver) {
        try {
            LOGGER.debug("Request received to getListOfSharedGroups in tenant " + request.getTenantId() +
                    "  for entity  " + request.getEntity().getId());

            SharedOwners owners = sharingClient.getListOfSharedGroups(request);
            responseObserver.onNext(owners);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at getListOfSharedGroups " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getListOfDirectlySharedGroups(SharingRequest request, StreamObserver<SharedOwners> responseObserver) {
        try {
            LOGGER.debug("Request received to getListOfDirectlySharedGroups in tenant " + request.getTenantId() +
                    "  for entity  " + request.getEntity().getId());

            SharedOwners owners = sharingClient.getListOfDirectlySharedGroups(request);
            responseObserver.onNext(owners);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getListOfDirectlySharedGroups " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void shareEntityWithUsers(SharingRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to shareEntityWithUsers in tenant " + request.getTenantId() +
                    "  for entity  " + request.getEntity().getId());

            long tenantId = request.getTenantId();

            String clientId = request.getClientId();

            String clientSec = request.getClientSec();

            for (String username : request.getOwnerIdList()) {

                validateAndGetUserProfile(username, clientId, clientSec, tenantId);
            }

            Status status = sharingClient.shareEntityWithUsers(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at shareEntityWithUsers " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void shareEntityWithGroups(SharingRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to shareEntityWithGroups in tenant " + request.getTenantId() +
                    "  with  entity Id  " + request.getEntity().getId());

            long tenantId = request.getTenantId();

            for (String groupId : request.getOwnerIdList()) {

                validateAndGetGroupId(groupId, tenantId);
            }

            Status status = sharingClient.shareEntityWithGroups(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at shareEntityWithGroups " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void revokeEntitySharingFromUsers(SharingRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to revokeEntitySharingFromUsers in tenant " + request.getTenantId() +
                    "  for entity  " + request.getEntity().getId());

            long tenantId = request.getTenantId();

            String clientId = request.getClientId();

            String clientSec = request.getClientSec();

            for (String username : request.getOwnerIdList()) {

                validateAndGetUserProfile(username, clientId, clientSec, tenantId);
            }

            Status status = sharingClient.revokeEntitySharingFromUsers(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at revokeEntitySharingFromUsers " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void revokeEntitySharingFromGroups(SharingRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to revokeEntitySharingFromGroups in tenant " + request.getTenantId() +
                    "  for entity  " + request.getEntity().getId());

            long tenantId = request.getTenantId();


            for (String username : request.getOwnerIdList()) {

                validateAndGetGroupId(username,  tenantId);
            }

            Status status = sharingClient.revokeEntitySharingFromGroups(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at revokeEntitySharingFromGroups " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void userHasAccess(SharingRequest request, StreamObserver<Status> responseObserver) {
        try {
            LOGGER.debug("Request received to userHasAccess in tenant " + request.getTenantId() +
                    "  for entity  " + request.getEntity().getId());

            String clientId = request.getClientId();

            String clientSec = request.getClientSec();

            long tenantId = request.getTenantId();


            for (String username : request.getOwnerIdList()) {

                validateAndGetUserProfile(username, clientId, clientSec, tenantId);
            }

            UserProfile profile = UserProfile.newBuilder().setUsername(request.getOwnerId(0)).build();


            UserProfileRequest userProfileRequest = UserProfileRequest.newBuilder()
                    .setTenantId(tenantId)
                    .setProfile(profile)
                    .build();
            GetAllGroupsResponse response = userProfileClient.getAllGroupsOfUser(userProfileRequest);

            List<Group> groups = response.getGroupsList();

            if (groups != null && !groups.isEmpty()) {

                for (Group group : groups) {
                    request = request.toBuilder().addOwnerId(group.getId()).build();
                }
            }

            Status status = sharingClient.userHasAccess(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at userHasAccess " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    private String validateAndGetUserProfile(String username, String clientId, String clientSec, long tenantId) {

        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setClientSecret(clientSec)
                .setTenantId(tenantId)
                .build();
        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

        if (token != null && token.getAccessToken() != null) {

            UserSearchMetadata searchMetadata = UserSearchMetadata
                    .newBuilder()
                    .setUsername(username)
                    .build();

            UserSearchRequest searchRequest = UserSearchRequest
                    .newBuilder()
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setUser(searchMetadata)
                    .build();

            CheckingResponse response = iamAdminServiceClient.isUserExist(searchRequest);
            if (!response.getIsExist()) {
                throw new SharingException("User " + username + " not found", null);
            }

            UserProfile userProfile = UserProfile
                    .newBuilder()
                    .setUsername(username)
                    .build();

            UserProfileRequest request = UserProfileRequest
                    .newBuilder()
                    .setProfile(userProfile)
                    .setTenantId(tenantId)
                    .build();

            UserProfile profile = userProfileClient.getUser(request);

            if (profile == null || profile.getUsername() == null || profile.getUsername().equals("")) {

                UserRepresentation representation = iamAdminServiceClient.getUser(searchRequest);

                UserProfile exProfile = InterServiceModelMapper.convert(representation);

                UserProfileRequest req = UserProfileRequest
                        .newBuilder()
                        .setProfile(exProfile)
                        .setTenantId(tenantId)
                        .build();

                userProfileClient.createUserProfile(req);
                return exProfile.getUsername();

            }

            return profile.getUsername();

        } else {

            throw new SharingException("User management token not found ", null);
        }


    }

    private String validateAndGetGroupId(String groupId, long tenantId) {


        Group group = Group.newBuilder().setId(groupId).build();

        GroupRequest groupRequest = GroupRequest
                .newBuilder()
                .setTenantId(tenantId)
                .setGroup(group).build();

        Group exGroup = userProfileClient.getGroup(groupRequest);

        if (exGroup == null || exGroup.getId() == null || exGroup.getId().equals("")) {

            throw new SharingException("Group with Id  " + groupId + " not found", null);
        }

        return exGroup.getId();


    }


}
