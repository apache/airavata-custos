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
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }


    }


    @Override
    public void registerAndEnableUsers(RegisterUsersRequest request, StreamObserver<RegisterUsersResponse> responseObserver) {
        try {

            RegisterUsersResponse registerUsersResponse = iamAdminServiceClient.
                    registerAndEnableUsers(request);


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

                    UserProfile profile = UserProfile.newBuilder()
                            .setFirstName(user.getFirstName())
                            .setLastName(user.getLastName())
                            .setEmail(user.getEmail())
                            .setStatus(UserStatus.valueOf(user.getState()))
                            .setTenantId(request.getTenantId())
                            .setUsername(user.getUsername())
                            .build();

                    UserProfile userProfile = userProfileClient.createUserProfile(profile);

                    if (userProfile != null && userProfile.getUserId() != null) {
                        responseObserver.onNext(user);
                        responseObserver.onCompleted();

                    } else {
                        String msg = "User enabling failed at user profile creation";
                        LOGGER.error(msg);
                        responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());

                    }

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
    public void deleteUser(UserSearchRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {


            GetUserProfileRequest req = GetUserProfileRequest
                    .newBuilder()
                    .setTenantId(request.getTenantId())
                    .setUsername(request.getUser().getUsername())
                    .build();

            UserProfile profile = userProfileClient.getUser(req);

            if (profile != null) {

                DeleteUserProfileRequest deleteUserProfileRequest = DeleteUserProfileRequest
                        .newBuilder()
                        .setTenantId(request.getTenantId())
                        .setUsername(request.getUser().getUsername())
                        .build();
                UserProfile deletedProfile = userProfileClient.deleteUser(deleteUserProfileRequest);

                if (deletedProfile != null) {

                    CheckingResponse response = iamAdminServiceClient.deleteUser(request);

                    responseObserver.onNext(response);
                    responseObserver.onCompleted();


                } else {
                    String msg = "User profile deletion failed for " + request.getUser().getUsername();
                    LOGGER.error(msg);
                    responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());
                }

            } else {
                CheckingResponse response = iamAdminServiceClient.deleteUser(request);

                responseObserver.onNext(response);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Error occurred at deleteUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUser(UserSearchRequest request, StreamObserver<UserRepresentation> responseObserver) {
        try {

            UserRepresentation user = iamAdminServiceClient.getUser(request);

            responseObserver.onNext(user);
            responseObserver.onCompleted();

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


            FindUsersResponse user = iamAdminServiceClient.getUsers(request);
            responseObserver.onNext(user);
            responseObserver.onCompleted();

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
                                           request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            CheckingResponse response = iamAdminServiceClient.deleteUserRoles(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
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

    //TODO: this is not updated
    @Override
    public void updateUserProfile(UserProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to updateUserProfile " + request.getUserProfile().getUsername() +
                    " at" + request.getUserProfile().getTenantId());


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


                UserSearchRequest info = UserSearchRequest
                        .newBuilder()
                        .setAccessToken(request.getAccessToken())
                        .setTenantId(request.getTenantId())
                        .build();

                UserRepresentation exUser = iamAdminServiceClient.getUser(info);

                CheckingResponse response = iamAdminServiceClient.updateUserProfile(updateUserProfileRequest);

                if (response != null && response.getIsExist()) {
                    try {
                        UserProfile userProfile = userProfileClient.updateUserProfile(request.getUserProfile());
                        responseObserver.onNext(userProfile);
                        responseObserver.onCompleted();
                    } catch (Exception ex) {
                        LOGGER.error("Error occurred while saving user profile in local DB, rolling back IAM service");
                        UpdateUserProfileRequest rollingRequest = UpdateUserProfileRequest
                                .newBuilder()
                                .setUser(exUser)
                                .setAccessToken(request.getAccessToken())
                                .build();
                        iamAdminServiceClient.updateUserProfile(rollingRequest);
                    }
                } else {
                    LOGGER.error("User profile  not found in IDP server");
                    responseObserver.onError(Status.CANCELLED.
                            withDescription("IAM server failed to update user profile").asRuntimeException());
                }

        } catch (Exception ex) {
            String msg = "Error occurred while updating user profile " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    //TODO: this is not updated

    @Override
    public void deleteUserProfile(DeleteProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteUserProfile " + request.getDeleteRequest().getUsername() +
                    " at" + request.getDeleteRequest().getTenantId());


            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getDeleteRequest().getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);


            if (token != null && token.getAccessToken() != null) {


                UserSearchRequest info = UserSearchRequest
                        .newBuilder()
                        .setAccessToken(token.getAccessToken())
                        .setTenantId(request.getDeleteRequest().getTenantId())
                        // .setUsername(request.getDeleteRequest().getUsername())
                        .build();


                UserProfile userProfile = userProfileClient.deleteUser(request.getDeleteRequest());

                responseObserver.onNext(userProfile);
                responseObserver.onCompleted();


                try {
                    iamAdminServiceClient.deleteUser(info);

                } catch (Exception ex) {
                    String msg = "Error occurred while deleting user profile in IDP , rolling back local DB";
                    LOGGER.error(msg);
                    userProfileClient.createUserProfile(userProfile);
                    responseObserver.onError(Status.CANCELLED.
                            withDescription("IAM server failed to update user profile").asRuntimeException());

                }

            } else {
                LOGGER.error("Error occurred retreving service account");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Service account not  found").asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred while delete user profile " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getUserProfile(GetUserProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to getUserProfile " + request.getUsername() +
                    " at" + request.getTenantId());


            UserProfile userProfile = userProfileClient.getUser(request);

            responseObserver.onNext(userProfile);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while get user profile " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getAllUserProfilesInTenant(GetAllUserProfilesRequest
                                                   request, StreamObserver<GetAllUserProfilesResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllUserProfilesInTenant " + request.getTenantId() +
                    " at" + request.getTenantId());

            GetAllUserProfilesResponse response = userProfileClient.getAllUserProfilesInTenant(request);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

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
}
