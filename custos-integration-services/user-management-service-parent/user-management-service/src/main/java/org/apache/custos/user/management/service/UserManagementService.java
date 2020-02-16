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
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSecret())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                RegisterUsersResponse registerUsersResponse = iamAdminServiceClient.registerAndEnableUsers(request.toBuilder().setAccessToken(token.getAccessToken()).build());

                responseObserver.onNext(registerUsersResponse);
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
    public void enableUser(UserIdentificationRequest request, StreamObserver<User> responseObserver) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getInfo().getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                UserAccessInfo info = request.getInfo()
                        .toBuilder()
                        .setAccessToken(token.getAccessToken())
                        .build();

                User user = iamAdminServiceClient.enableUser(info);

                if (user != null) {

                    UserProfile profile = UserProfile.newBuilder()
                            .setFirstName(user.getFirstName())
                            .setLastName(user.getLastName())
                            .setEmailAddress(user.getEmail())
                            .setStatus(UserStatus.valueOf(user.getState()))
                            .setTenantId(request.getInfo().getTenantId())
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

                String msg = "Cannot find service account";
                LOGGER.error(msg);
                responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());

            }

        } catch (Exception ex) {
            String msg = "Error occurred at enableUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteUser(UserIdentificationRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {


            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getInfo().getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                UserAccessInfo info = request.getInfo()
                        .toBuilder()
                        .setAccessToken(token.getAccessToken())
                        .build();

                GetUserProfileRequest req = GetUserProfileRequest
                        .newBuilder()
                        .setTenantId(request.getInfo().getTenantId())
                        .setUsername(request.getInfo().getUsername())
                        .build();

                UserProfile profile = userProfileClient.getUser(req);

                if (profile != null) {

                    DeleteUserProfileRequest deleteUserProfileRequest = DeleteUserProfileRequest
                            .newBuilder()
                            .setTenantId(request.getInfo().getTenantId())
                            .setUsername(request.getInfo().getUsername())
                            .build();
                    UserProfile deletedProfile = userProfileClient.deleteUser(deleteUserProfileRequest);

                    if (deletedProfile != null) {

                        CheckingResponse response = iamAdminServiceClient.deleteUser(info);

                        responseObserver.onNext(response);
                        responseObserver.onCompleted();


                    } else {
                        String msg = "User profile deletion failed for " + request.getInfo().getUsername();
                        LOGGER.error(msg);
                        responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());
                    }

                } else {
                    CheckingResponse response = iamAdminServiceClient.deleteUser(info);

                    responseObserver.onNext(response);
                    responseObserver.onCompleted();
                }


            } else {
                String msg = "Cannot find service account";
                LOGGER.error(msg);
                responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at deleteUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUser(GetUserRequest request, StreamObserver<User> responseObserver) {
        try {

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                UserAccessInfo info = UserAccessInfo.newBuilder()
                        .setAccessToken(token.getAccessToken())
                        .setUsername(request.getUsername())
                        .setTenantId(request.getTenantId())
                        .build();

                User user = iamAdminServiceClient.getUser(info);

                responseObserver.onNext(user);
                responseObserver.onCompleted();


            } else {

                String msg = "Cannot find service account";
                LOGGER.error(msg);
                responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred at getUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUsers(GetUsersRequest request, StreamObserver<GetUsersResponse> responseObserver) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {


                UserAccessInfo info = UserAccessInfo.newBuilder()
                        .setUsername(request.getUsername())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(token.getAccessToken())
                        .build();


                org.apache.custos.iam.service.GetUsersRequest getUsersRequest =

                        org.apache.custos.iam.service.GetUsersRequest
                                .newBuilder()
                                .setLimit(request.getLimit())
                                .setOffset(request.getOffset())
                                .setSearch(request.getSearch())
                                .setInfo(info)
                                .build();

                GetUsersResponse user = iamAdminServiceClient.getUsers(getUsersRequest);
                responseObserver.onNext(user);
                responseObserver.onCompleted();

            } else {

                String msg = "Cannot find service account";
                LOGGER.error(msg);
                responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at getUsers " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void resetPassword(ResetPasswordRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getPasswordMetadata().getIamClientId())
                    .setClientSecret(request.getPasswordMetadata().getIamClientSecret())
                    .setTenantId(request.getPasswordMetadata().getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                UserAccessInfo info = UserAccessInfo.newBuilder()
                        .setUsername(request.getPasswordMetadata().getUsername())
                        .setTenantId(request.getPasswordMetadata().getTenantId())
                        .setAccessToken(token.getAccessToken())
                        .build();

                ResetUserPassword resetUserPassword = ResetUserPassword
                        .newBuilder()
                        .setPassword(request.getPasswordMetadata().getPassword())
                        .setInfo(info)
                        .build();

                CheckingResponse response = iamAdminServiceClient.resetPassword(resetUserPassword);
                responseObserver.onNext(response);
                responseObserver.onCompleted();

            } else {
                String msg = "Cannot find service account";
                LOGGER.error(msg);
                responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred at resetPassword " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addRoleToUser(RoleOperationRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            CheckingResponse response = iamAdminServiceClient.addRoleToUser(request.getRole());
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at addRoleToUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void deleteRoleFromUser(RoleOperationRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            CheckingResponse response = iamAdminServiceClient.deleteRoleFromUser(request.getRole());
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at deleteRoleFromUser " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isUserEnabled(GetUserRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {


            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {


                UserAccessInfo info = UserAccessInfo.newBuilder()
                        .setAccessToken(token.getAccessToken())
                        .setUsername(request.getUsername())
                        .setTenantId(request.getTenantId())
                        .build();


                CheckingResponse response = iamAdminServiceClient.isUserEnabled(info);

                responseObserver.onNext(response);
                responseObserver.onCompleted();


            } else {
                String msg = "Cannot find service account";
                LOGGER.error(msg);
                responseObserver.onError(Status.CANCELLED.withDescription(msg).asRuntimeException());

            }

        } catch (Exception ex) {
            String msg = "Error occurred at isUserEnabled " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateUserProfile(UserProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to updateUserProfile " + request.getUserProfile().getUsername() +
                    " at" + request.getUserProfile().getTenantId());

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getUserProfile().getTenantId())
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);


            if (token != null && token.getAccessToken() != null) {

                User user = User.newBuilder()
                        .setFirstName(request.getUserProfile().getFirstName())
                        .setLastName(request.getUserProfile().getLastName())
                        .setEmail(request.getUserProfile().getEmailAddress())
                        .setUsername(request.getUserProfile().getUsername())
                        .setTenantId(request.getUserProfile().getTenantId())
                        .setInternalUserId(request.getUserProfile().getUserId())
                        .setCreationTime(System.currentTimeMillis())
                        .setState(request.getUserProfile().getStatus().name())
                        .build();

                UpdateUserProfileRequest updateUserProfileRequest = UpdateUserProfileRequest
                        .newBuilder()
                        .setUser(user)
                        .setAccessToken(token.getAccessToken())
                        .build();


                UserAccessInfo info = UserAccessInfo
                        .newBuilder()
                        .setAccessToken(token.getAccessToken())
                        .setTenantId(request.getUserProfile().getTenantId())
                        .setUsername(request.getUserProfile().getUsername())
                        .build();

                User exUser = iamAdminServiceClient.getUser(info);

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
                                .setAccessToken(token.getAccessToken())
                                .build();
                        iamAdminServiceClient.updateUserProfile(rollingRequest);
                    }
                } else {
                    LOGGER.error("User profile  not found in IDP server");
                    responseObserver.onError(Status.CANCELLED.
                            withDescription("IAM server failed to update user profile").asRuntimeException());
                }
            } else {
                LOGGER.error("Error occurred retreving service account");
                responseObserver.onError(Status.CANCELLED.
                        withDescription("Service account not  found").asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while updating user profile " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


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


                UserAccessInfo info = UserAccessInfo
                        .newBuilder()
                        .setAccessToken(token.getAccessToken())
                        .setTenantId(request.getDeleteRequest().getTenantId())
                        .setUsername(request.getDeleteRequest().getUsername())
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
    public void getAllUserProfilesInTenant(GetAllUserProfilesRequest request, StreamObserver<GetAllUserProfilesResponse> responseObserver) {
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
    public void getUserProfileAuditTrails(GetUpdateAuditTrailRequest request, StreamObserver<GetUpdateAuditTrailResponse> responseObserver) {
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
