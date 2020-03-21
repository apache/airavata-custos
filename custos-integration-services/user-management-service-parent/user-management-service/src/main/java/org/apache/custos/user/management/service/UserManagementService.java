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

package org.apache.custos.user.management.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.admin.client.IamAdminServiceClient;
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
import java.util.stream.Collectors;

/**
 * Service class of User Management Service
 */
@GRpcService
public class UserManagementService extends UserManagementServiceGrpc.UserManagementServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(UserManagementService.class);


    @Autowired
    private UserProfileClient userProfileClient;

    @Autowired
    private IamAdminServiceClient iamAdminServiceClient;

    @Autowired
    private IdentityClient identityClient;


    @Override
    public void registerUser(RegisterUserRequest request, StreamObserver<RegisterUserResponse> responseObserver) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                org.apache.custos.iam.service.RegisterUserRequest registerUserRequest = request.
                        toBuilder()
                        .setAccessToken(token.getAccessToken())
                        .build();

                RegisterUserResponse registerUserResponse = iamAdminServiceClient.registerUser(registerUserRequest);

                responseObserver.onNext(registerUserResponse);
                responseObserver.onCompleted();

            } else {
                LOGGER.error("Cannot find service token");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Cannot find service token").asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred at registerUser " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("CredentialGenerationException")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }


    }


    @Override
    public void registerAndEnableUsers(RegisterUsersRequest request, StreamObserver<RegisterUsersResponse> responseObserver) {
        try {

            RegisterUsersResponse registerUsersResponse = iamAdminServiceClient.
                    registerAndEnableUsers(request);


            if (request.getUsersList() != null && !request.getUsersList().isEmpty() &&
                    registerUsersResponse.getAllUseresRegistered()) {
                try {


                    request.getUsersList().forEach(user -> {
                        List<org.apache.custos.user.profile.service.UserAttribute> userAtrList = new ArrayList<>();
                        if (user.getAttributesList() != null && !user.getAttributesList().isEmpty()) {

                            user.getAttributesList().forEach(atr -> {
                                org.apache.custos.user.profile.service.UserAttribute userAttribute =
                                        org.apache.custos.user.profile.service.UserAttribute
                                                .newBuilder()
                                                .setKey(atr.getKey())
                                                .addAllValue(atr.getValuesList())
                                                .build();

                                userAtrList.add(userAttribute);
                            });
                        }


                        UserProfile profile = UserProfile.newBuilder()
                                .setFirstName(user.getFirstName())
                                .setLastName(user.getLastName())
                                .setEmail(user.getEmail())
                                .setStatus(UserStatus.valueOf("ACTIVE"))
                                .addAllAttributes(userAtrList)
                                .addAllRealmRoles(user.getRealmRolesList())
                                .addAllClientRoles(user.getClientRolesList())
                                .setUsername(user.getUsername().toLowerCase())
                                .build();
                        org.apache.custos.user.profile.service.UserProfileRequest profileRequest =
                                org.apache.custos.user.profile.service.UserProfileRequest.newBuilder()
                                        .setProfile(profile)
                                        .setTenantId(request.getTenantId())
                                        .build();

                        userProfileClient.createUserProfile(profileRequest);


                    });
                } catch (Exception ex) {

                    request.getUsersList().forEach(user -> {

                        UserSearchMetadata metadata = UserSearchMetadata
                                .newBuilder()
                                .setUsername(user.getUsername())
                                .build();
                        UserSearchRequest searchRequest = UserSearchRequest
                                .newBuilder()
                                .setTenantId(request.getTenantId())
                                .setClientId(request.getClientId())
                                .setAccessToken(request.getAccessToken())
                                .setUser(metadata)
                                .build();
                        iamAdminServiceClient.deleteUser(searchRequest);

                    });

                }

            }

            responseObserver.onNext(registerUsersResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at registerAndEnableUsers " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }

    }


    @Override
    public void addUserAttributes(AddUserAttributesRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {

            OperationStatus status = iamAdminServiceClient.addUserAttributes(request);


            for (String user : request.getUsersList()) {

                UserSearchMetadata metadata = UserSearchMetadata
                        .newBuilder()
                        .setUsername(user).build();

                UserSearchRequest searchRequest = UserSearchRequest
                        .newBuilder()
                        .setClientId(request.getClientId())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(request.getAccessToken())
                        .setUser(metadata)
                        .build();

                UserRepresentation representation = iamAdminServiceClient.getUser(searchRequest);


                if (representation != null) {

                    UserProfile profile = this.convertToProfile(representation);

                    org.apache.custos.user.profile.service.UserProfileRequest req =
                            org.apache.custos.user.profile.service.UserProfileRequest
                                    .newBuilder()
                                    .setTenantId(request.getTenantId())
                                    .setProfile(profile)
                                    .build();


                    UserProfile existingProfile = userProfileClient.getUser(req);

                    if (existingProfile == null || existingProfile.getUsername().trim().equals("")) {
                        userProfileClient.createUserProfile(req);
                    } else {

                        userProfileClient.updateUserProfile(req);
                    }


                }

            }
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at addUserAttributes " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void deleteUserAttributes(DeleteUserAttributeRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {

            OperationStatus status = iamAdminServiceClient.deleteUserAttributes(request);


            for (String user : request.getUsersList()) {

                UserSearchMetadata metadata = UserSearchMetadata
                        .newBuilder()
                        .setUsername(user).build();

                UserSearchRequest searchRequest = UserSearchRequest
                        .newBuilder()
                        .setClientId(request.getClientId())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(request.getAccessToken())
                        .setUser(metadata)
                        .build();

                UserRepresentation representation = iamAdminServiceClient.getUser(searchRequest);


                if (representation != null) {

                    UserProfile profile = this.convertToProfile(representation);

                    org.apache.custos.user.profile.service.UserProfileRequest req =
                            org.apache.custos.user.profile.service.UserProfileRequest
                                    .newBuilder()
                                    .setTenantId(request.getTenantId())
                                    .setProfile(profile)
                                    .build();


                    userProfileClient.updateUserProfile(req);


                }

            }
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at deleteUserAttributes " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void enableUser(UserSearchRequest request, StreamObserver<UserRepresentation> responseObserver) {
        try {

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();

                UserRepresentation user = iamAdminServiceClient.enableUser(request);

                if (user != null) {

                    UserProfile profile = this.convertToProfile(user);

                    org.apache.custos.user.profile.service.UserProfileRequest profileRequest =
                            org.apache.custos.user.profile.service.UserProfileRequest.newBuilder()
                                    .setProfile(profile)
                                    .setTenantId(request.getTenantId())
                                    .build();

                    UserProfile exProfile = userProfileClient.getUser(profileRequest);

                    if (exProfile.getUsername().equals("")) {
                        userProfileClient.createUserProfile(profileRequest);
                    } else {
                        userProfileClient.updateUserProfile(profileRequest);
                    }


                    responseObserver.onNext(user);
                    responseObserver.onCompleted();


                } else {
                    String msg = "User enabling failed at IDP server";
                    LOGGER.error(msg);
                    responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());

                }
            } else {
                LOGGER.error("Cannot find service token");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Cannot find service token").asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred at enableUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void disableUser(UserSearchRequest request, StreamObserver<UserRepresentation> responseObserver) {
        try {

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();

                UserRepresentation user = iamAdminServiceClient.disableUser(request);

                if (user != null) {

                    UserProfile profile = this.convertToProfile(user);

                    org.apache.custos.user.profile.service.UserProfileRequest profileRequest =
                            org.apache.custos.user.profile.service.UserProfileRequest.newBuilder()
                                    .setProfile(profile)
                                    .setTenantId(request.getTenantId())
                                    .build();

                    UserProfile exProfile = userProfileClient.getUser(profileRequest);

                    if (exProfile.getUsername().equals("")) {
                        userProfileClient.createUserProfile(profileRequest);
                    } else {
                        userProfileClient.updateUserProfile(profileRequest);
                    }


                    responseObserver.onNext(user);
                    responseObserver.onCompleted();


                } else {
                    String msg = "User enabling failed at IDP server";
                    LOGGER.error(msg);
                    responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());

                }
            } else {
                LOGGER.error("Cannot find service token");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Cannot find service token").asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred at enableUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteUser(UserSearchRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            UserProfile profileReq = UserProfile.newBuilder().setUsername(request.getUser().getUsername().toLowerCase()).build();


            org.apache.custos.user.profile.service.UserProfileRequest req =
                    org.apache.custos.user.profile.service.UserProfileRequest
                            .newBuilder()
                            .setTenantId(request.getTenantId())
                            .setProfile(profileReq)
                            .build();

            UserProfile profile = userProfileClient.getUser(req);

            if (profile != null && !profile.getUsername().trim().equals("")) {


                UserProfile deletedProfile = userProfileClient.deleteUser(req);

                if (deletedProfile != null) {

                    CheckingResponse response = iamAdminServiceClient.deleteUser(request);

                    OperationStatus status = OperationStatus.newBuilder().setStatus(response.getIsExist()).build();
                    responseObserver.onNext(status);
                    responseObserver.onCompleted();


                } else {
                    String msg = "User profile deletion failed for " + request.getUser().getUsername();
                    LOGGER.error(msg);
                    responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());
                }

            } else {
                CheckingResponse response = iamAdminServiceClient.deleteUser(request);

                OperationStatus status = OperationStatus.newBuilder().setStatus(response.getIsExist()).build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Error occurred at deleteUser " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void getUser(UserSearchRequest request, StreamObserver<UserRepresentation> responseObserver) {
        try {

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();

                UserRepresentation user = iamAdminServiceClient.getUser(request);

                responseObserver.onNext(user);
                responseObserver.onCompleted();
            } else {
                LOGGER.error("Cannot find service token");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Cannot find service token").asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at getUser " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void findUsers(FindUsersRequest request, StreamObserver<FindUsersResponse> responseObserver) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                FindUsersResponse user = iamAdminServiceClient.getUsers(request);
                responseObserver.onNext(user);
                responseObserver.onCompleted();
            } else {
                LOGGER.error("Cannot find service token");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Cannot find service token").asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at getUsers " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void resetPassword(ResetUserPassword request, StreamObserver<CheckingResponse> responseObserver) {
        try {

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();

                CheckingResponse response = iamAdminServiceClient.resetPassword(request);
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                LOGGER.error("Cannot find service token");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Cannot find service token").asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at resetPassword " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addRolesToUsers(AddUserRolesRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {

            OperationStatus response = iamAdminServiceClient.addRolesToUsers(request);

            for (String user : request.getUsernamesList()) {

                UserSearchMetadata metadata = UserSearchMetadata
                        .newBuilder()
                        .setUsername(user).build();

                UserSearchRequest searchRequest = UserSearchRequest
                        .newBuilder()
                        .setClientId(request.getClientId())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(request.getAccessToken())
                        .setUser(metadata)
                        .build();

                UserRepresentation representation = iamAdminServiceClient.getUser(searchRequest);


                if (representation != null) {

                    UserProfile profile = this.convertToProfile(representation);

                    org.apache.custos.user.profile.service.UserProfileRequest req =
                            org.apache.custos.user.profile.service.UserProfileRequest
                                    .newBuilder()
                                    .setTenantId(request.getTenantId())
                                    .setProfile(profile)
                                    .build();


                    UserProfile exsistingProfile = userProfileClient.getUser(req);

                    if (exsistingProfile == null || exsistingProfile.getUsername().trim().equals("")) {
                        userProfileClient.createUserProfile(req);
                    } else {

                        userProfileClient.updateUserProfile(req);
                    }
                }

            }

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at addRolesToUsers " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void deleteUserRoles(DeleteUserRolesRequest
                                        request, StreamObserver<OperationStatus> responseObserver) {
        try {
            CheckingResponse response = iamAdminServiceClient.deleteUserRoles(request);


            UserSearchMetadata metadata = UserSearchMetadata
                    .newBuilder()
                    .setUsername(request.getUsername()).build();

            UserSearchRequest searchRequest = UserSearchRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setTenantId(request.getTenantId())
                    .setAccessToken(request.getAccessToken())
                    .setUser(metadata)
                    .build();

            UserRepresentation representation = iamAdminServiceClient.getUser(searchRequest);


            if (representation != null) {

                UserProfile profile = this.convertToProfile(representation);

                org.apache.custos.user.profile.service.UserProfileRequest req =
                        org.apache.custos.user.profile.service.UserProfileRequest
                                .newBuilder()
                                .setTenantId(request.getTenantId())
                                .setProfile(profile)
                                .build();


                userProfileClient.updateUserProfile(req);


            }


            OperationStatus status = OperationStatus.newBuilder().setStatus(response.getIsExist()).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (
                Exception ex) {
            String msg = "Error occurred at deleteRoleFromUser " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }

    }

    @Override
    public void isUserEnabled(UserSearchRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {
                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                CheckingResponse response = iamAdminServiceClient.isUserEnabled(request);

                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {

                LOGGER.error("Cannot find service token");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Cannot find service token").asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred at isUserEnabled " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void isUsernameAvailable(UserSearchRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {
                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                CheckingResponse response = iamAdminServiceClient.isUsernameAvailable(request);

                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {

                LOGGER.error("Cannot find service token");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Cannot find service token").asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred at isUsernameAvailable " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void updateUserProfile(UserProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to updateUserProfile " + request.getUserProfile().getUsername() +
                    " at" + request.getTenantId());


            UserRepresentation.Builder builder = UserRepresentation.newBuilder()
                    .setFirstName(request.getUserProfile().getFirstName())
                    .setLastName(request.getUserProfile().getLastName())
                    .setEmail(request.getUserProfile().getEmail())
                    .setUsername(request.getUserProfile().getUsername());

            if (request.getUserProfile().getStatus() != null) {
                builder.setState(request.getUserProfile().getStatus().name());
            }


            UpdateUserProfileRequest updateUserProfileRequest = UpdateUserProfileRequest
                    .newBuilder()
                    .setUser(builder.build())
                    .setAccessToken(request.getAccessToken())
                    .setTenantId(request.getTenantId())
                    .build();

            UserSearchMetadata metadata = UserSearchMetadata
                    .newBuilder()
                    .setUsername(request.getUserProfile().getUsername()).build();

            UserSearchRequest info = UserSearchRequest
                    .newBuilder()
                    .setAccessToken(request.getAccessToken())
                    .setTenantId(request.getTenantId())
                    .setUser(metadata)
                    .build();

            UserRepresentation exUser = iamAdminServiceClient.getUser(info);

            CheckingResponse response = iamAdminServiceClient.updateUserProfile(updateUserProfileRequest);

            if (response != null && response.getIsExist()) {
                try {
                    org.apache.custos.user.profile.service.UserProfileRequest userProfileRequest =
                            org.apache.custos.user.profile.service.UserProfileRequest.
                                    newBuilder()
                                    .setProfile(request.getUserProfile())
                                    .setTenantId(request.getTenantId())
                                    .build();

                    UserProfile profile = userProfileClient.getUser(userProfileRequest);

                    if (profile != null && !profile.getUsername().equals("")) {
                        profile = profile.toBuilder()
                                .setEmail(request.getUserProfile().getEmail())
                                .setFirstName(request.getUserProfile().getFirstName())
                                .setLastName(request.getUserProfile().getLastName())
                                .setUsername(request.getUserProfile().getUsername())
                                .build();

                        userProfileClient.
                                updateUserProfile(userProfileRequest.toBuilder().setProfile(profile).build());

                        responseObserver.onNext(profile);
                        responseObserver.onCompleted();

                    } else {
                        String msg = "Error occurred while saving user profile in local DB, " +
                                "rolling back IAM service" + "User profile not found";
                        LOGGER.error(msg);
                        UpdateUserProfileRequest rollingRequest = UpdateUserProfileRequest
                                .newBuilder()
                                .setUser(exUser)
                                .setAccessToken(request.getAccessToken())
                                .setTenantId(request.getTenantId())
                                .build();
                        iamAdminServiceClient.updateUserProfile(rollingRequest);
                        responseObserver.onError(Status.CANCELLED.
                                withDescription(msg).asRuntimeException());
                    }


                } catch (Exception ex) {
                    String msg = "Error occurred while saving user profile in local DB, " +
                            "rolling back IAM service" + ex.getMessage();
                    LOGGER.error(msg);
                    UpdateUserProfileRequest rollingRequest = UpdateUserProfileRequest
                            .newBuilder()
                            .setUser(exUser)
                            .setAccessToken(request.getAccessToken())
                            .setTenantId(request.getTenantId())
                            .build();
                    iamAdminServiceClient.updateUserProfile(rollingRequest);
                    responseObserver.onError(Status.CANCELLED.
                            withDescription(msg).asRuntimeException());
                }
            } else {
                LOGGER.error("User profile  not found in IDP server");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("IAM server failed to update user profile").asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while updating user profile " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void deleteUserProfile(UserProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteUserProfile " + request.getUserProfile().getUsername() +
                    " at" + request.getTenantId());

            UserSearchMetadata metadata = UserSearchMetadata
                    .newBuilder()
                    .setUsername(request.getUserProfile().getUsername()).build();

            UserSearchRequest info = UserSearchRequest
                    .newBuilder()
                    .setAccessToken(request.getAccessToken())
                    .setTenantId(request.getTenantId())
                    .setUser(metadata)
                    .build();


            org.apache.custos.user.profile.service.UserProfileRequest userProfileRequest =

                    org.apache.custos.user.profile.service.UserProfileRequest
                            .newBuilder()
                            .setProfile(request.getUserProfile())
                            .setTenantId(request.getTenantId())
                            .build();
            UserProfile userProfile = userProfileClient.deleteUser(userProfileRequest);

            responseObserver.onNext(userProfile);
            responseObserver.onCompleted();

            iamAdminServiceClient.deleteUser(info);


        } catch (Exception ex) {
            String msg = "Error occurred while delete user profile " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void getUserProfile(UserProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to getUserProfile " + request.getUserProfile().getUsername() +
                    " at" + request.getTenantId());

            org.apache.custos.user.profile.service.UserProfileRequest userProfileRequest =

                    org.apache.custos.user.profile.service.UserProfileRequest
                            .newBuilder()
                            .setProfile(request.getUserProfile())
                            .setTenantId(request.getTenantId())
                            .build();

            UserProfile userProfile = userProfileClient.getUser(userProfileRequest);

            responseObserver.onNext(userProfile);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while get user profile " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getAllUserProfilesInTenant(UserProfileRequest
                                                   request, StreamObserver<GetAllUserProfilesResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllUserProfilesInTenant " + request.getTenantId() +
                    " at" + request.getTenantId());

            org.apache.custos.user.profile.service.UserProfileRequest userProfileRequest =

                    org.apache.custos.user.profile.service.UserProfileRequest
                            .newBuilder()
                            .setProfile(request.getUserProfile())
                            .setTenantId(request.getTenantId())
                            .build();


            if (request.getUserProfile().getAttributesList() == null || request.getUserProfile().getAttributesList().isEmpty()) {
                GetAllUserProfilesResponse response = userProfileClient.getAllUserProfilesInTenant(userProfileRequest);

                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                GetAllUserProfilesResponse response = userProfileClient.findUserProfilesByUserAttributes(userProfileRequest);

                responseObserver.onNext(response);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Error occurred while get all  user profiles in tenant " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUserProfileAuditTrails(GetUpdateAuditTrailRequest
                                                  request, StreamObserver<GetUpdateAuditTrailResponse> responseObserver) {
        try {

            LOGGER.debug("Request received to getUserProfileAuditTrails " + request.getUsername() +
                    " at" + request.getTenantId());

            GetUpdateAuditTrailResponse response = userProfileClient.getUserProfileUpdateAuditTrail(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while get user profile audit trails " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void linkUserProfile(LinkUserProfileRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to linkUserProfile   at " + request.getTenantId());

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {


                UserSearchMetadata metadata = UserSearchMetadata
                        .newBuilder()
                        .setUsername(request.getCurrentUsername()).build();

                UserSearchRequest searchRequest = UserSearchRequest
                        .newBuilder()
                        .setClientId(request.getIamClientId())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(token.getAccessToken())
                        .setUser(metadata)
                        .build();


                UserRepresentation userTobeLinked = iamAdminServiceClient.getUser(searchRequest);


                if (userTobeLinked != null && !userTobeLinked.getUsername().equals("")) {

                    UserSearchMetadata exMetadata = UserSearchMetadata
                            .newBuilder().setUsername(request.getPreviousUsername()).build();

                    UserSearchRequest exSearchRequest = UserSearchRequest
                            .newBuilder()
                            .setClientId(request.getIamClientId())
                            .setTenantId(request.getTenantId())
                            .setAccessToken(token.getAccessToken())
                            .setUser(exMetadata)
                            .build();

                    UserRepresentation exRep = iamAdminServiceClient.getUser(exSearchRequest);

                    if (exRep != null && !exRep.getUsername().equals("")) {

                        boolean profileUpdate = false;

                        List<UserAttribute> userAttributeList = new ArrayList<>();

                        for (String attribute : request.getLinkingAttributesList()) {

                            if ("name".equals(attribute)) {
                                profileUpdate = true;
                                userTobeLinked = userTobeLinked.toBuilder()
                                        .setFirstName(exRep.getFirstName())
                                        .setLastName(exRep.getLastName())
                                        .build();

                            } else if (("email").equals(attribute)) {
                                profileUpdate = true;
                                userTobeLinked = userTobeLinked.toBuilder().setEmail(exRep.getEmail()).build();

                            } else {
                                List<UserAttribute> userAttributes = exRep.getAttributesList().stream().
                                        filter(atr -> atr.getKey().equals(attribute)).collect(Collectors.toList());

                                if (!userAttributes.isEmpty()) {
                                    UserAttribute userAttribute = userAttributes.get(0);
                                    userAttributeList.add(userAttribute);
                                }
                            }
                        }

                        if (profileUpdate) {
                            UpdateUserProfileRequest updateUserProfileRequest = UpdateUserProfileRequest
                                    .newBuilder()
                                    .setUser(userTobeLinked)
                                    .setAccessToken(token.getAccessToken())
                                    .setTenantId(request.getTenantId())
                                    .build();
                            iamAdminServiceClient.updateUserProfile(updateUserProfileRequest);
                        }

                        if (!userAttributeList.isEmpty()) {

                            AddUserAttributesRequest addUserAttributesRequest = AddUserAttributesRequest
                                    .newBuilder()
                                    .addUsers(request.getCurrentUsername())
                                    .addAllAttributes(userAttributeList)
                                    .setTenantId(request.getTenantId())
                                    .setAccessToken(token.getAccessToken())
                                    .setClientId(request.getIamClientId())
                                    .setPerformedBy(request.getPerformedBy())
                                    .build();
                            iamAdminServiceClient.addUserAttributes(addUserAttributesRequest);

                        }


                        UserRepresentation updatedUser = iamAdminServiceClient.getUser(searchRequest);

                        if (updatedUser != null) {

                            UserProfile profile = this.convertToProfile(updatedUser);

                            org.apache.custos.user.profile.service.UserProfileRequest req =
                                    org.apache.custos.user.profile.service.UserProfileRequest
                                            .newBuilder()
                                            .setTenantId(request.getTenantId())
                                            .setProfile(profile)
                                            .build();

                            userProfileClient.updateUserProfile(req);

                        }

                        CheckingResponse response = CheckingResponse.newBuilder().setIsExist(true).build();
                        OperationStatus status = OperationStatus.newBuilder().setStatus(response.getIsExist()).build();
                        responseObserver.onNext(status);
                        responseObserver.onCompleted();

                    } else {
                        String msg = "Cannot found existing user ";
                        LOGGER.error(msg);
                        responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                    }

                }
            }

        } catch (Exception ex) {
            String msg = "Error occurred while get all  user profiles in tenant " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    private UserProfile convertToProfile(UserRepresentation representation) {
        UserProfile.Builder profileBuilder = UserProfile.newBuilder();


        if (representation.getRealmRolesCount() > 0) {
            profileBuilder.addAllRealmRoles(representation.getRealmRolesList());

        }

        if (representation.getClientRolesCount() > 0) {
            profileBuilder.addAllClientRoles(representation.getClientRolesList());

        }

        if (representation.getAttributesCount() > 0) {
            List<UserAttribute> attributeList = representation.getAttributesList();

            List<org.apache.custos.user.profile.service.UserAttribute> userAtrList = new ArrayList<>();
            attributeList.forEach(atr -> {
                org.apache.custos.user.profile.service.UserAttribute userAttribute =
                        org.apache.custos.user.profile.service.UserAttribute
                                .newBuilder()
                                .setKey(atr.getKey())
                                .addAllValue(atr.getValuesList())
                                .build();

                userAtrList.add(userAttribute);
            });
            profileBuilder.addAllAttributes(userAtrList);


        }

        profileBuilder.setUsername(representation.getUsername().toLowerCase());
        profileBuilder.setFirstName(representation.getFirstName());
        profileBuilder.setLastName(representation.getLastName());
        profileBuilder.setEmail(representation.getEmail());

        return profileBuilder.build();

    }


}
