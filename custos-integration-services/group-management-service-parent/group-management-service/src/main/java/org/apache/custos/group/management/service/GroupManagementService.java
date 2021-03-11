/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.group.management.service;


import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.admin.client.IamAdminServiceClient;
import org.apache.custos.iam.service.GroupRequest;
import org.apache.custos.iam.service.UserAttribute;
import org.apache.custos.iam.service.*;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.identity.service.GetUserManagementSATokenRequest;
import org.apache.custos.user.profile.client.UserProfileClient;
import org.apache.custos.user.profile.service.*;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@GRpcService
public class GroupManagementService extends GroupManagementServiceGrpc.GroupManagementServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(GroupManagementService.class);

    @Autowired
    private UserProfileClient userProfileClient;

    @Autowired
    private IamAdminServiceClient iamAdminServiceClient;


    @Autowired
    private IdentityClient identityClient;


    //TODO: improve error handling to avoid database consistency
    @Override
    public void createKeycloakGroups(GroupsRequest request, StreamObserver<GroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received create Groups for tenant " + request.getTenantId());

            GroupsResponse response = iamAdminServiceClient.createGroups(request);


            List<org.apache.custos.user.profile.service.GroupRequest> groupRequests =
                    getAllGroupRequests(response.getGroupsList(), null, request.getTenantId(), request.getPerformedBy());


            for (org.apache.custos.user.profile.service.GroupRequest groupRequest : groupRequests) {

                userProfileClient.createGroup(groupRequest);
            }

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at createGroups " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    //TODO: improve error handling to avoid database consistency
    @Override
    public void updateKeycloakGroup(GroupRequest request, StreamObserver<GroupRepresentation> responseObserver) {
        try {
            LOGGER.debug("Request received to updateGroup for tenant " + request.getTenantId());

            GroupRepresentation gr = request.getGroup();

            if (request.getId() != null && !request.getId().trim().equals("")) {
                gr = gr.toBuilder().setId(request.getId()).build();
            }
            request = request.toBuilder().setGroup(gr).build();

            GroupRepresentation response = iamAdminServiceClient.updateGroup(request);

            List<GroupRepresentation> representations = new ArrayList<>();
            representations.add(response);

            Group group = Group.newBuilder()
                    .setName(response.getName())
                    .setId(response.getId())
                    .build();

            org.apache.custos.user.profile.service.GroupRequest groupRequest = org.apache.custos.user.profile.service.GroupRequest
                    .newBuilder().
                            setTenantId(request.getTenantId()).
                            setPerformedBy(request.getPerformedBy()).
                            setGroup(group).build();

            Group exGroup = userProfileClient.getGroup(groupRequest);


            List<org.apache.custos.user.profile.service.GroupRequest> groupRequests =
                    getAllGroupRequests(representations, exGroup.getParentId(), request.getTenantId(), request.getPerformedBy());


            userProfileClient.updateGroup(groupRequests.get(0));

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred at updateGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    //TODO: improve error handling to avoid database consistency
    @Override
    public void deleteKeycloakGroup(GroupRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to updateGroup for tenant " + request.getTenantId());

            GroupRepresentation gr = request.getGroup();
            if (request.getId() != null && !request.getId().trim().equals("")) {
                gr = gr.toBuilder().setId(request.getId()).build();
            }
            request = request.toBuilder().setGroup(gr).build();

            OperationStatus response = iamAdminServiceClient.deleteGroup(request);

            Group group = Group.newBuilder()
                    .setId(request.getGroup().getId())
                    .build();

            org.apache.custos.user.profile.service.GroupRequest groupRequest = org.apache.custos.user.profile.service.GroupRequest
                    .newBuilder().
                            setTenantId(request.getTenantId()).
                            setPerformedBy(request.getPerformedBy()).
                            setGroup(group).build();

            userProfileClient.deleteGroup(groupRequest);


            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at deleteGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void findKeycloakGroup(GroupRequest request, StreamObserver<GroupRepresentation> responseObserver) {
        try {
            LOGGER.debug("Request received findGroup for group Id " + request.getGroup().getId() + " of  tenant "
                    + request.getTenantId());


            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();

                GroupRepresentation representation = iamAdminServiceClient.findGroup(request);

                responseObserver.onNext(representation);
                responseObserver.onCompleted();
            } else {
                String msg = "Error occurred at findGroup, authentication token not found ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at findGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void getAllKeycloakGroups(GroupRequest request, StreamObserver<GroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received getAllGroups for  tenant "
                    + request.getTenantId());

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();

                GroupsResponse response = iamAdminServiceClient.getAllGroups(request);

                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                String msg = "Error occurred at findGroup, authentication token not found ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at getAllGroups " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void addUserToKeycloakGroup(UserGroupMappingRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to addUserToGroup for  user  " + request.getUsername() + " of tenant "
                    + request.getTenantId());

            OperationStatus status = iamAdminServiceClient.addUserToGroup(request);


            if (status.getStatus()) {

                GroupMembership membership = GroupMembership
                        .newBuilder()
                        .setGroupId(request.getGroupId())
                        .setUsername(request.getUsername())
                        .setTenantId(request.getTenantId())
                        .setType(request.getMembershipType())
                        .build();

                userProfileClient.addUserToGroup(membership);


            }

            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at addUserToGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void removeUserFromKeycloakGroup(UserGroupMappingRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to removeUserFromGroup for  user  " + request.getUsername() + " of tenant "
                    + request.getTenantId());


            OperationStatus status = iamAdminServiceClient.removeUserFromGroup(request);


            if (status.getStatus()) {

                GroupMembership membership = GroupMembership
                        .newBuilder()
                        .setGroupId(request.getGroupId())
                        .setUsername(request.getUsername())
                        .setTenantId(request.getTenantId())
                        .build();

                userProfileClient.removeUserFromGroup(membership);

            }

            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at removeUserFromGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void createGroup(org.apache.custos.user.profile.service.GroupRequest request, StreamObserver<Group> responseObserver) {
        try {
            LOGGER.debug("Request received to createGroup   " + request.getGroup().getName() + " of tenant "
                    + request.getTenantId());

            String id = request.getGroup().getId();

            updateProfile(request.getClientId(), request.getClientSec(),
                    request.getTenantId(), request.getGroup().getOwnerId());
            if (id != null && !id.trim().equals("")) {
                Group group = Group.newBuilder()
                        .setId(id)
                        .build();

                org.apache.custos.user.profile.service.GroupRequest groupRequest = org.apache.custos.user.profile.service.GroupRequest
                        .newBuilder().
                                setTenantId(request.getTenantId()).
                                setPerformedBy(request.getPerformedBy()).
                                setGroup(group).build();
                Group exGroup = userProfileClient.getGroup(groupRequest);

                if (exGroup.getName() != null && !exGroup.getName().trim().equals("")) {
                    String msg = "Group already exist with given id " + id;
                    LOGGER.error(msg);
                    responseObserver.onError(Status.ALREADY_EXISTS.withDescription(msg).asRuntimeException());
                }

            } else {
                id = request.getGroup().
                        getName().toLowerCase().replace(" ", "_") + "_" + UUID.randomUUID();
            }
            Group group = request.getGroup().toBuilder().setId(id).build();
            request = request.toBuilder().setGroup(group).build();
            Group creadredGroup = userProfileClient.createGroup(request);
            responseObserver.onNext(creadredGroup);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at createGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateGroup(org.apache.custos.user.profile.service.GroupRequest request, StreamObserver<Group> responseObserver) {
        try {
            LOGGER.debug("Request received to updateGroup for  group  " + request.getGroup().getId() + " of tenant "
                    + request.getTenantId());


            if (request.getId() != null && !request.getId().trim().equals("")) {
                Group group = request.getGroup().toBuilder().setId(request.getId()).build();
                request = request.toBuilder().setGroup(group).build();
            }

            Group exGroup = userProfileClient.getGroup(request);
            if (exGroup.getName() != null && !exGroup.getName().trim().equals("")) {
                Group group = request.getGroup().toBuilder().setParentId(exGroup.getParentId()).build();
                request = request.toBuilder().setGroup(group).build();
                Group updatedGr = userProfileClient.updateGroup(request);
                responseObserver.onNext(updatedGr);
                responseObserver.onCompleted();
            } else {
                String msg = "Cannot find a group with id " + request.getId();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at updateGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void deleteGroup(org.apache.custos.user.profile.service.GroupRequest request,
                            StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteGroup for  group  " + request.getGroup().getId() + " of tenant "
                    + request.getTenantId());
            if (request.getId() != null && !request.getId().trim().equals("")) {
                Group group = request.getGroup().toBuilder().setId(request.getId()).build();
                request = request.toBuilder().setGroup(group).build();
            }
            userProfileClient.deleteGroup(request);
            org.apache.custos.user.profile.service.Status status =
                    org.apache.custos.user.profile.service.Status.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at removeUserFromGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void findGroup(org.apache.custos.user.profile.service.GroupRequest request, StreamObserver<Group> responseObserver) {
        try {
            LOGGER.debug("Request received to findGroup  of tenant " + request.getTenantId());
            Group group = userProfileClient.getGroup(request);
            responseObserver.onNext(group);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at removeUserFromGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void getAllGroups(org.apache.custos.user.profile.service.GroupRequest request, StreamObserver<GetAllGroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllGroups of tenant " + request.getTenantId());
            GetAllGroupsResponse response = userProfileClient.getAllGroups(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at removeUserFromGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void addUserToGroup(GroupMembership request, StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to addUserToGroup for  user  " + request.getUsername() + " of tenant "
                    + request.getTenantId());

            updateProfile(request.getClientId(), request.getClientSec(), request.getTenantId(), request.getUsername());

            org.apache.custos.user.profile.service.Status status = userProfileClient.addUserToGroup(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at addUserToGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void removeUserFromGroup(GroupMembership request, StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to removeUserFromGroup for  user  " + request.getUsername() + " of tenant "
                    + request.getTenantId());
            org.apache.custos.user.profile.service.Status status = userProfileClient.removeUserFromGroup(request);
            responseObserver.onNext(status);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred at removeUserFromGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void addChildGroupToParentGroup(GroupToGroupMembership request,
                                           StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to addChildGroupToParentGroup for  group  " + request.getChildId() +
                    " to add " + request.getParentId() + " of tenant " + request.getTenantId());

            org.apache.custos.user.profile.service.Status status = userProfileClient.addChildGroupToParentGroup(request);


            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at addChildGroupToParentGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void removeChildGroupFromParentGroup(GroupToGroupMembership request,
                                                StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to removeUserFromGroup for  group  " + request.getChildId() +
                    " to remove " + request.getParentId() + " of tenant " + request.getTenantId());

            org.apache.custos.user.profile.service.Status status = userProfileClient.removeChildGroupFromParentGroup(request);

            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at removeChildGroupFromParentGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void getAllGroupsOfUser(UserProfileRequest request,
                                   StreamObserver<GetAllGroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllGroupsOfUser for  user  " + request.getProfile().getUsername() + " of tenant "
                    + request.getTenantId());

            GetAllGroupsResponse response = userProfileClient.getAllGroupsOfUser(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at getAllGroupsOfUser " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void getAllParentGroupsOfGroup(org.apache.custos.user.profile.service.GroupRequest request,
                                          StreamObserver<GetAllGroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllParentGroupsOfGroup for  group  "
                    + request.getGroup().getId() + " of tenant " + request.getTenantId());

            GetAllGroupsResponse response = userProfileClient.getAllParentGroupsOfGroup(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getAllParentGroupsOfGroup " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }


    @Override
    public void getAllChildUsers(org.apache.custos.user.profile.service.GroupRequest request, StreamObserver<GetAllUserProfilesResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllChildUsers for  group  "
                    + request.getGroup().getId() + " of tenant " + request.getTenantId());

            GetAllUserProfilesResponse response = userProfileClient.getAllChildUsers(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getAllChildUsers " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void getAllChildGroups(org.apache.custos.user.profile.service.GroupRequest request, StreamObserver<GetAllGroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllChildGroups for  group  "
                    + request.getGroup().getId() + " of tenant " + request.getTenantId());

            GetAllGroupsResponse response = userProfileClient.getAllChildGroups(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at getAllChildGroups " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void changeUserMembershipType(GroupMembership request, StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to changeUserMembershipType for  user  "
                    + request.getUsername() + " of tenant " + request.getTenantId());

            org.apache.custos.user.profile.service.Status response = userProfileClient.changeUserMembershipType(request);


            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at changeUserMembershipType " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void hasAccess(GroupMembership request, StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to hasAccess for  user  "
                    + request.getUsername() + " of tenant " + request.getTenantId());

            org.apache.custos.user.profile.service.Status response = userProfileClient.hasAccess(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at hasAccess " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }

    @Override
    public void addGroupMembershipType(UserGroupMembershipTypeRequest request,
                                       StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to addGroupMembershipType for  tenant " + request.getTenantId()
                    + ", type " + request.getType());

            org.apache.custos.user.profile.service.Status response = userProfileClient.addUserMembershipType(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at addGroupMembershipType " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }

    }

    @Override
    public void removeUserGroupMembershipType(UserGroupMembershipTypeRequest request,
                                              StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to removeUserGroupMembershipType for  tenant " + request.getTenantId()
                    + ", type " + request.getType());

            org.apache.custos.user.profile.service.Status response = userProfileClient.removeUserMembershipType(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at removeUserGroupMembershipType " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

        }

    }

    private void updateProfile(String clientId, String clientSec, long tenantId, String username) {

        UserProfile userProfile = UserProfile.newBuilder().setUsername(username).build();
        UserProfileRequest userProfileRequest = UserProfileRequest
                .newBuilder()
                .setTenantId(tenantId)
                .setProfile(userProfile)
                .build();

        UserProfile exUser = userProfileClient.getUser(userProfileRequest);
        if (exUser.getUsername().isBlank()) {

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(clientId)
                    .setClientSecret(clientSec)
                    .setTenantId(tenantId)
                    .build();

            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);
            UserSearchMetadata userSearchMetadata = UserSearchMetadata
                    .newBuilder().setUsername(username).build();

            UserSearchRequest searchRequest = UserSearchRequest
                    .newBuilder()
                    .setClientId(clientId)
                    .setTenantId(tenantId)
                    .setAccessToken(token.getAccessToken())
                    .setUser(userSearchMetadata)
                    .build();

            UserRepresentation representation = iamAdminServiceClient.getUser(searchRequest);

            UserProfile profile = UserProfile
                    .newBuilder()
                    .setUsername(username)
                    .setFirstName(representation.getFirstName())
                    .setLastName(representation.getLastName())
                    .setEmail(representation.getEmail())
                    .build();

            UserProfileRequest profileRequest = UserProfileRequest
                    .newBuilder()
                    .setTenantId(tenantId)
                    .setProfile(profile)
                    .build();

            userProfileClient.createUserProfile(profileRequest);
        }


    }


    private org.apache.custos.user.profile.service.GroupRequest createGroup(GroupRepresentation representation,
                                                                            String parentId,
                                                                            long tenantId,
                                                                            String performedBy) {

        List<GroupAttribute> attributes = new ArrayList<>();

        if (representation.getAttributesList() != null && !representation.getAttributesList().isEmpty()) {
            for (UserAttribute attribute : representation.getAttributesList()) {
                GroupAttribute groupAttribute = GroupAttribute
                        .newBuilder()
                        .setKey(attribute.getKey())
                        .addAllValue(attribute.getValuesList()).build();
                attributes.add(groupAttribute);

            }


        }

        Group group = Group
                .newBuilder()
                .setId(representation.getId())
                .setName(representation.getName())
                .addAllClientRoles(representation.getClientRolesList())
                .addAllRealmRoles(representation.getRealmRolesList())
                .addAllAttributes(attributes)
                .build();

        if (parentId != null) {
            group = group.toBuilder().setParentId(parentId).build();
        }

        if (representation.getOwnerId() != null && !representation.getOwnerId().trim().equals("")) {
            group = group.toBuilder().setOwnerId(representation.getOwnerId()).build();
        }

        if (representation.getDescription() != null && !representation.getDescription().trim().equals("")) {
            group = group.toBuilder().setDescription(representation.getDescription()).build();
        }

        org.apache.custos.user.profile.service.GroupRequest groupRequest =

                org.apache.custos.user.profile.service.GroupRequest
                        .newBuilder()
                        .setTenantId(tenantId)
                        .setPerformedBy(performedBy)
                        .setGroup(group).build();

        return groupRequest;


    }


    private List<org.apache.custos.user.profile.service.GroupRequest> getAllGroupRequests
            (List<GroupRepresentation> groupRepresentations, String parentId, long tenantId, String performedBy) {

        List<org.apache.custos.user.profile.service.GroupRequest> groupRequests = new ArrayList<>();
        for (GroupRepresentation representation : groupRepresentations) {

            org.apache.custos.user.profile.service.GroupRequest groupRequest =
                    createGroup(representation, parentId, tenantId, performedBy);
            groupRequests.add(groupRequest);

            if (representation.getSubGroupsList() != null && !representation.getSubGroupsList().isEmpty()) {

                List<org.apache.custos.user.profile.service.GroupRequest> list =
                        getAllGroupRequests(representation.getSubGroupsList(), representation.getId(), tenantId, performedBy);

                if (!list.isEmpty()) {
                    groupRequests.addAll(list);
                }

            }

        }

        return groupRequests;

    }


}
