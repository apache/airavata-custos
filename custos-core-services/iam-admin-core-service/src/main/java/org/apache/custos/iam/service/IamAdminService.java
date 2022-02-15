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

package org.apache.custos.iam.service;

import com.google.protobuf.Empty;
import io.grpc.stub.StreamObserver;
import org.apache.custos.core.services.commons.StatusUpdater;
import org.apache.custos.core.services.commons.persistance.model.OperationStatus;
import org.apache.custos.core.services.commons.persistance.model.StatusEntity;
import org.apache.custos.core.services.commons.util.Constants;
import org.apache.custos.federated.services.clients.keycloak.KeycloakClient;
import org.apache.custos.federated.services.clients.keycloak.KeycloakClientSecret;
import org.apache.custos.federated.services.clients.keycloak.UnauthorizedException;
import org.apache.custos.iam.service.IamAdminServiceGrpc.IamAdminServiceImplBase;
import org.apache.custos.iam.utils.IAMOperations;
import org.apache.custos.iam.utils.Status;
import org.keycloak.representations.idm.UserRepresentation;
import org.keycloak.representations.idm.*;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@GRpcService
public class IamAdminService extends IamAdminServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(IamAdminService.class);

    @Autowired
    private KeycloakClient keycloakClient;

    @Autowired
    private StatusUpdater statusUpdater;

    @Override
    public void setUPTenant(SetUpTenantRequest request, StreamObserver<SetUpTenantResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to setUPTenant  " + request.getTenantId());

            keycloakClient.deleteRealm(String.valueOf(request.getTenantId()));

            keycloakClient.createRealm(String.valueOf(request.getTenantId()), request.getTenantName());

            keycloakClient.createRealmAdminAccount(String.valueOf(request.getTenantId()), request.getAdminUsername(),
                    request.getAdminFirstname(), request.getAdminLastname(),
                    request.getAdminEmail(), request.getAdminPassword());

            KeycloakClientSecret clientSecret = keycloakClient.configureClient(String.valueOf(request.getTenantId()),
                    request.getCustosClientId(),
                    request.getTenantURL(), request.getRedirectURIsList());

            SetUpTenantResponse response = SetUpTenantResponse.newBuilder()
                    .setClientId(clientSecret.getClientId())
                    .setClientSecret(clientSecret.getClientSecret())
                    .build();


            statusUpdater.updateStatus(IAMOperations.SET_UP_TENANT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getRequesterEmail());

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during setUPTenant" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.SET_UP_TENANT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getRequesterEmail());

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void updateTenant(SetUpTenantRequest request, StreamObserver<SetUpTenantResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to updateTenant  " + request.getTenantId());

            keycloakClient.updateRealm(String.valueOf(request.getTenantId()), request.getTenantName());

            keycloakClient.updateRealmAdminAccount(String.valueOf(request.getTenantId()), request.getAdminUsername(),
                    request.getAdminFirstname(), request.getAdminLastname(),
                    request.getAdminEmail(), request.getAdminPassword());

            KeycloakClientSecret clientSecret = keycloakClient.updateClient(String.valueOf(request.getTenantId()),
                    request.getCustosClientId(),
                    request.getTenantURL(), request.getRedirectURIsList());

            SetUpTenantResponse response = SetUpTenantResponse.newBuilder()
                    .setClientId(clientSecret.getClientId())
                    .setClientSecret(clientSecret.getClientSecret())
                    .build();


            statusUpdater.updateStatus(IAMOperations.UPDATE_TENANT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getRequesterEmail());

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during updateTenant" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.UPDATE_TENANT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getRequesterEmail());

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteTenant(DeleteTenantRequest request, StreamObserver<Empty> responseObserver) {
        try {
            LOGGER.debug("Request received to delete tenant  " + request.getTenantId());

            keycloakClient.deleteRealm(String.valueOf(request.getTenantId()));

            responseObserver.onNext(Empty.newBuilder().build());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during delete tenant" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isUsernameAvailable(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to isUsernameAvailable at " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUsernameAvailable(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(),
                    request.getAccessToken());

            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(isAvailable).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during isUsernameAvailable" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isUserEnabled(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to isUserEnabled at " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUserAccountEnabled(String.valueOf(request.getTenantId()),
                    request.getAccessToken(),
                    request.getUser().getUsername());
            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(isAvailable).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during isUserEnabled" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void registerUser(RegisterUserRequest request, StreamObserver<RegisterUserResponse> responseObserver) {

        try {
            LOGGER.debug("Request received to registerUser for " + request.getTenantId());

            boolean registered = keycloakClient.createUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(),
                    request.getUser().getPassword(),
                    request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail(),
                    request.getUser().getTemporaryPassword(),
                    request.getAccessToken());


            RegisterUserResponse registerUserResponse = RegisterUserResponse.newBuilder().
                    setIsRegistered(registered).build();


            statusUpdater.updateStatus(IAMOperations.REGISTER_USER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));

            responseObserver.onNext(registerUserResponse);
            responseObserver.onCompleted();

        } catch (UnauthorizedException ex) {
            String msg = "Error occurred during registerUser" + ex;
            statusUpdater.updateStatus(IAMOperations.REGISTER_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
        } catch (Exception ex) {
            String msg = "Error occurred during registerUser" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.REGISTER_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void enableUser(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.UserRepresentation> responseObserver) {

        try {
            LOGGER.debug("Request received to enableUser for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());

            if (!status) {
                statusUpdater.updateStatus(IAMOperations.ENABLE_USER.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                String msg = "User not valid ";
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }

            boolean accountEnabled = keycloakClient.enableUserAccount(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUser().getUsername());
            if (accountEnabled) {

                UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()),
                        request.getAccessToken(), request.getUser().getUsername());

                org.apache.custos.iam.service.UserRepresentation user = getUser(representation, request.getClientId());


                statusUpdater.updateStatus(IAMOperations.ENABLE_USER.name(),
                        OperationStatus.SUCCESS,
                        request.getTenantId(),
                        String.valueOf(request.getTenantId()));


                responseObserver.onNext(user);
                responseObserver.onCompleted();

            } else {

                statusUpdater.updateStatus(IAMOperations.ENABLE_USER.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                String msg = "Account enabling failed ";
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred during enableUser" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.ENABLE_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void disableUser(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.UserRepresentation> responseObserver) {
        try {
            LOGGER.debug("Request received to disable for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());


            if (!status) {
                statusUpdater.updateStatus(IAMOperations.DISABLE_USER.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                String msg = "User not valid ";
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }


            boolean accountDisabled = keycloakClient.disableUserAccount(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUser().getUsername());
            if (accountDisabled) {

                UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()),
                        request.getAccessToken(), request.getUser().getUsername());

                org.apache.custos.iam.service.UserRepresentation user = getUser(representation, request.getClientId());


                statusUpdater.updateStatus(IAMOperations.DISABLE_USER.name(),
                        OperationStatus.SUCCESS,
                        request.getTenantId(),
                        String.valueOf(request.getTenantId()));


                responseObserver.onNext(user);
                responseObserver.onCompleted();

            } else {

                statusUpdater.updateStatus(IAMOperations.DISABLE_USER.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                String msg = "Account enabling failed ";
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred during disabling user" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DISABLE_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isUserExist(UserSearchRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUserExist for " + request.getTenantId());


            boolean isUserExist = keycloakClient.isUserExist(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUser().getUsername());
            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isUserExist).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during isUserExist" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUser(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.UserRepresentation> responseObserver) {
        try {
            LOGGER.debug("Request received to getUser for " + request.getTenantId());


            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername());


            if (!status) {
                String msg = "User " + request.getUser().getUsername() + "not found at " + request.getTenantId();
                responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }


            UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername());

            if (representation != null) {
                org.apache.custos.iam.service.UserRepresentation user = getUser(representation, request.getClientId());

                UserSessionRepresentation sessionRepresentation = keycloakClient.getLatestSession(String.valueOf(request.getTenantId()),
                        request.getClientId(), request.getAccessToken(), request.getUser().getUsername());

                if (sessionRepresentation != null) {
                    user = user.toBuilder().setLastLoginAt(sessionRepresentation.getLastAccess()).build();
                } else {
                    EventRepresentation eventRepresentation = keycloakClient.
                            getLastLoginEvent(String.valueOf(request.getTenantId()), request.getClientId()
                                    , request.getUser().getUsername());
                    if (eventRepresentation != null) {
                        user = user.toBuilder().setLastLoginAt(eventRepresentation.getTime()).build();
                    }
                }
                responseObserver.onNext(user);
                responseObserver.onCompleted();
            } else {
                String msg = "User " + request.getUser().getUsername() + " not found in " + request.getTenantId();
                responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred during getUser" + ex;
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void findUsers(FindUsersRequest request, StreamObserver<FindUsersResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getUsers for " + request.getUser().getUsername());
             long initiationTime = System.currentTimeMillis();
            List<UserRepresentation> representation = keycloakClient.getUsers(request.getAccessToken(),
                    String.valueOf(request.getTenantId()), request.getOffset(), request.getLimit(),
                    request.getUser().getUsername(), request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail(),
                    request.getUser().getId());
            List<org.apache.custos.iam.service.UserRepresentation> users = new ArrayList<>();
            representation.stream().forEach(r -> {
                boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                        r.getUsername(), request.getAccessToken());

                if (status) {

                    org.apache.custos.iam.service.UserRepresentation user = this.getUser(r, request.getClientId());

                    UserSessionRepresentation sessionRepresentation = keycloakClient.getLatestSession(String.valueOf(request.getTenantId()),
                            request.getClientId(), null, r.getUsername());
                    if (sessionRepresentation != null) {
                        user = user.toBuilder().setLastLoginAt(sessionRepresentation.getLastAccess()).build();
                    } else {
                        EventRepresentation eventRepresentation = keycloakClient.
                                getLastLoginEvent(String.valueOf(request.getTenantId()), request.getClientId()
                                        , r.getUsername());

                        if (eventRepresentation != null) {
                            user = user.toBuilder().setLastLoginAt(eventRepresentation.getTime()).build();
                        }
                    }

                    users.add(user);
                }

            });
            long endTime = System.currentTimeMillis();
            long total = endTime - initiationTime;
            LOGGER.info("request received: "+ initiationTime+" request end time"+ endTime+" difference " + total);
            FindUsersResponse response = FindUsersResponse.newBuilder().addAllUsers(users).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during getUsers" + ex;
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void resetPassword(ResetUserPassword request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        String userId = request.getUsername() + "@" + request.getTenantId();
        try {
            LOGGER.debug("Request received to resetPassword for " + request.getUsername());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUsername(), request.getAccessToken());


            if (!status) {
                statusUpdater.updateStatus(IAMOperations.RESET_PASSWORD.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), userId);
                String msg = "User not valid ";
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }

            boolean isChanged = keycloakClient.resetUserPassword(request.getAccessToken(),
                    String.valueOf(request.getTenantId()),
                    request.getUsername(),
                    request.getPassword());

            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(isChanged).build();


            statusUpdater.updateStatus(IAMOperations.RESET_PASSWORD.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), userId);


            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during resetPassword" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.RESET_PASSWORD.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), userId);

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void deleteExternalIDPLinksOfUsers(DeleteExternalIDPsRequest request,
                                              StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            long tenantId = request.getTenantId();
            boolean status = false;
            if (request.getUserIdList().isEmpty()) {
                status = keycloakClient.deleteExternalIDPLinks(String.valueOf(tenantId));
            } else {
                status = keycloakClient.deleteExternalIDPLinks(String.valueOf(tenantId), request.getUserIdList());
            }
            responseObserver.onNext(org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build());
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while deletingExternalIDPLinksOfUsers" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getExternalIDPLinksOfUsers(GetExternalIDPsRequest request, StreamObserver<GetExternalIDPsResponse> responseObserver) {
        try {
            long tenantId = request.getTenantId();
            List<FederatedIdentityRepresentation> identityRepresentations = keycloakClient.getExternalIDPLinks(String.valueOf(tenantId), request.getUserId());
            GetExternalIDPsResponse.Builder response = GetExternalIDPsResponse.newBuilder();
            identityRepresentations.forEach(rep -> {
                response.addIdpLinks(ExternalIDPLink.newBuilder()
                        .setProviderAlias(rep.getIdentityProvider())
                        .setProviderUsername(rep.getUserName())
                        .setProviderUserId(rep.getUserId()));
            });

            responseObserver.onNext(response.build());
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while getExternalIDPLinksOfUsers" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addExternalIDPLinksOfUsers(AddExternalIDPLinksRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            long tenantId = request.getTenantId();
            List<ExternalIDPLink> externalIDPLinkList = request.getIdpLinksList();
            List<FederatedIdentityRepresentation> federatedIdentityRepresentations = new ArrayList<>();
            externalIDPLinkList.forEach(link -> {
                FederatedIdentityRepresentation representation = new FederatedIdentityRepresentation();
                representation.setUserId(link.getProviderUserId());
                representation.setUserName(link.getProviderUsername());
                representation.setIdentityProvider(link.getProviderAlias());
                federatedIdentityRepresentations.add(representation);

            });
            keycloakClient.addExternalIDPLinks(String.valueOf(tenantId), federatedIdentityRepresentations);
            org.apache.custos.iam.service.OperationStatus status = org.apache.custos.iam.service.OperationStatus
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while getExternalIDPLinksOfUsers" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void updateUserProfile(UpdateUserProfileRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        String userId = request.getUser().getUsername() + "@" + request.getTenantId();

        try {
            LOGGER.debug("Request received to updateUserProfile for " + request.getUser().getUsername());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());


            if (!status) {
                statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), userId);
                String msg = "User not valid ";
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }

            keycloakClient.updateUserRepresentation(request.getAccessToken(),
                    String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(),
                    request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail());

            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(true).build();


            statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), userId);


            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during updateUserProfile" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), userId);

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteUser(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {

        try {
            LOGGER.debug("Request received to deleteUser for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());


            if (!status) {
                statusUpdater.updateStatus(IAMOperations.DELETE_USER.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), request.getPerformedBy());
                String msg = "User not valid ";
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }

            boolean isUpdated = keycloakClient.deleteUser(request.getAccessToken(),
                    String.valueOf(request.getTenantId()), request.getUser().getUsername());

            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(isUpdated).build();

            statusUpdater.updateStatus(IAMOperations.DELETE_USER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred during deleteUser" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.DELETE_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());

            if (ex.getMessage().contains("Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void deleteRolesFromUser(DeleteUserRolesRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {

        try {
            LOGGER.debug("Request received to deleteRoleFromUser for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUsername());


            if (!status) {
                statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_USER.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), request.getPerformedBy());
                String msg = "User not valid ";
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }


            if (!request.getRolesList().isEmpty()) {

                keycloakClient.removeRoleFromUser(request.getAccessToken(),
                        String.valueOf(request.getTenantId()), request.getUsername(),
                        request.getRolesList(), request.getClientId(), false);
            }

            if (!request.getClientRolesList().isEmpty()) {
                keycloakClient.removeRoleFromUser(request.getAccessToken(),
                        String.valueOf(request.getTenantId()), request.getUsername(),
                        request.getClientRolesList(), request.getClientId(), true);

            }
            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(true).build();


            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_USER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during deleteRoleFromUser" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());

            if (ex.getMessage().contains("Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void getOperationMetadata(GetOperationsMetadataRequest request, StreamObserver<GetOperationsMetadataResponse> responseObserver) {
        try {
            LOGGER.debug("Calling getOperationMetadata API for traceId " + request.getTraceId());

            List<OperationMetadata> metadata = new ArrayList<>();
            List<StatusEntity> entities = statusUpdater.getOperationStatus(request.getTraceId());
            if (entities == null || entities.size() > 0) {

                for (StatusEntity statusEntity : entities) {
                    OperationMetadata data = convertFromEntity(statusEntity);
                    metadata.add(data);
                }
            }

            GetOperationsMetadataResponse response = GetOperationsMetadataResponse
                    .newBuilder()
                    .addAllMetadata(metadata)
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " operation failed for " + request.getTraceId();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void configureFederatedIDP(ConfigureFederateIDPRequest request, StreamObserver<FederateIDPResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to configureFederatedIDP for " + request.getTenantId());


            keycloakClient.configureOIDCFederatedIDP(String.valueOf(request.getTenantId()), "CILogon", request.getScope(),
                    new KeycloakClientSecret(request.getClientID(), request.getClientSec()), null);


            FederateIDPResponse federateIDPResponse = FederateIDPResponse.newBuilder().setStatus(true).build();

            statusUpdater.updateStatus(IAMOperations.CONFIGURE_IDP.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getRequesterEmail());

            responseObserver.onNext(federateIDPResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Configure Federated IDP failed for " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void registerAndEnableUsers(RegisterUsersRequest request, StreamObserver<RegisterUsersResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to registerMultipleUsers for " + request.getTenantId());

            List<org.apache.custos.iam.service.UserRepresentation> userRepresentations = request.getUsersList();

            List<org.apache.custos.iam.service.UserRepresentation> failedList = new ArrayList<>();


            for (org.apache.custos.iam.service.UserRepresentation userRepresentation : userRepresentations) {

                try {
                    keycloakClient.createUser(String.valueOf(request.getTenantId()),
                            userRepresentation.getUsername(),
                            userRepresentation.getPassword(),
                            userRepresentation.getFirstName(),
                            userRepresentation.getLastName(),
                            userRepresentation.getEmail(),
                            userRepresentation.getTemporaryPassword(),
                            request.getAccessToken());

                    keycloakClient.enableUserAccount
                            (String.valueOf(request.getTenantId()),
                                    request.getAccessToken(), userRepresentation.getUsername().toLowerCase());
                    List<String> userList = new ArrayList<>();
                    userList.add(userRepresentation.getUsername());
                    if (!userRepresentation.getRealmRolesList().isEmpty()) {
                        keycloakClient.addRolesToUsers(request.getAccessToken(),
                                String.valueOf(request.getTenantId()), userList, userRepresentation.getRealmRolesList(),
                                request.getClientId(), false);
                    }
                    if (!userRepresentation.getClientRolesList().isEmpty()) {
                        keycloakClient.addRolesToUsers(request.getAccessToken(),
                                String.valueOf(request.getTenantId()), userList, userRepresentation.getClientRolesList(),
                                request.getClientId(), true);
                    }

                    if (!userRepresentation.getAttributesList().isEmpty()) {

                        Map<String, List<String>> map = new HashMap<>();
                        for (UserAttribute attribute : userRepresentation.getAttributesList()) {

                            if (attribute.getKey().equals(Constants.CUSTOS_REALM_AGENT)) {
                                // Constants.CUSTOS_REALM_AGENT + " cannot be used as a valid attribute";
                                continue;
                            }
                            map.put(attribute.getKey(), attribute.getValuesList());
                        }

                        keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()),
                                request.getAccessToken(), map, userList);

                    }

                } catch (UnauthorizedException ex) {
                    LOGGER.error(" Error occurred while adding user " + userRepresentation.getUsername() +
                            " to realm" + request.getTenantId());
                    responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(ex.getMessage())
                            .asRuntimeException());
                    return;
                } catch (Exception ex) {
                    if (ex.getMessage().contains("Unauthorized")) {
                        responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(ex.getMessage())
                                .asRuntimeException());
                        return;
                    }
                    LOGGER.error(" Error occurred while adding user " + userRepresentation.getUsername() +
                            " to realm" + request.getTenantId());
                    failedList.add(userRepresentation);
                }
            }

            if (failedList.isEmpty()) {
                statusUpdater.updateStatus(IAMOperations.REGISTER_ENABLE_USERS.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(),
                        request.getPerformedBy());
            }


            RegisterUsersResponse response = RegisterUsersResponse.newBuilder()
                    .setAllUseresRegistered(failedList.isEmpty())
                    .addAllFailedUsers(failedList).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.REGISTER_ENABLE_USERS.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            String msg = " Register  multiple users  failed for " + request.getTenantId();
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());

        }
    }


    @Override
    public void addRolesToUsers(AddUserRolesRequest request,
                                StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to addRolesToUsers for " + request.getTenantId());

            List<String> validUserNames = new ArrayList<>();

            for (String username : request.getUsernamesList()) {
                boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                        username);

                if (status) {
                    validUserNames.add(username);
                }
            }

            keycloakClient.addRolesToUsers(request.getAccessToken(), String.valueOf(request.getTenantId()),
                    validUserNames, request.getRolesList(), request.getClientId(), request.getClientLevel());

            statusUpdater.updateStatus(IAMOperations.ADD_ROLES_TO_USERS.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            org.apache.custos.iam.service.OperationStatus status = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_ROLES_TO_USERS.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Add  multiple users  failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void addRolesToTenant(AddRolesRequest request, StreamObserver<AllRoles> responseObserver) {
        try {
            LOGGER.debug("Request received to add roles to tenant for " + request.getTenantId());

            List<RoleRepresentation> rolesRepresentations = request.getRolesList();

            List<org.keycloak.representations.idm.RoleRepresentation> keycloakRolesList = new ArrayList<>();

            for (RoleRepresentation roleRepresentation : rolesRepresentations) {
                org.keycloak.representations.idm.RoleRepresentation role = new org.keycloak.representations.idm.RoleRepresentation();
                role.setName(roleRepresentation.getName());
                role.setDescription(roleRepresentation.getDescription());
                role.setComposite(roleRepresentation.getComposite());
                keycloakRolesList.add(role);
            }

            keycloakClient.addRoles(keycloakRolesList, String.valueOf(request.getTenantId()),
                    request.getClientId(), request.getClientLevel());

            statusUpdater.updateStatus(IAMOperations.ADD_ROLES_TO_TENANT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));

            List<org.keycloak.representations.idm.RoleRepresentation> allKeycloakRoles = keycloakClient.
                    getAllRoles(String.valueOf(request.getTenantId()), (request.getClientLevel()) ? request.getClientId() : null);
            AllRoles.Builder builder = AllRoles.newBuilder();
            if (allKeycloakRoles != null && !allKeycloakRoles.isEmpty()) {

                List<RoleRepresentation> roleRepresentations = new ArrayList<>();
                for (org.keycloak.representations.idm.RoleRepresentation role : allKeycloakRoles) {
                    RoleRepresentation roleRepresentation = RoleRepresentation.
                            newBuilder().setName(role.getName())
                            .setComposite(role.isComposite())
                            .setId(role.getId())
                            .build();
                    if (role.getDescription() != null) {
                        roleRepresentation = roleRepresentation.toBuilder().setDescription(role.getDescription()).build();
                    }
                    roleRepresentations.add(roleRepresentation);


                }

                builder.addAllRoles(roleRepresentations);
                if (request.getClientLevel()) {
                    builder.setScope("client_level");
                } else {
                    builder.setScope("realm_level");
                }

            }

            responseObserver.onNext(builder.build());
            responseObserver.onCompleted();


        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_ROLES_TO_TENANT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            String msg = " Add roles   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteRole(DeleteRoleRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to add roles to tenant for " + request.getTenantId());

            keycloakClient.deleteRole(request.getRole().getId(), String.valueOf(request.getTenantId()),
                    request.getClientId(), request.getClientLevel());
            org.apache.custos.iam.service.OperationStatus operationStatus =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(true).build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = " Deleting role" + request.getRole().getName() + "   " +
                    "failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getRolesOfTenant(GetRolesRequest request, StreamObserver<AllRoles> responseObserver) {
        try {
            LOGGER.debug("Request received to add roles to tenant for " + request.getTenantId());

            List<org.keycloak.representations.idm.RoleRepresentation> allKeycloakRoles = keycloakClient.
                    getAllRoles(String.valueOf(request.getTenantId()), (request.getClientLevel()) ? request.getClientId() : null);
            AllRoles.Builder builder = AllRoles.newBuilder();
            if (allKeycloakRoles != null && !allKeycloakRoles.isEmpty()) {

                List<RoleRepresentation> roleRepresentations = new ArrayList<>();
                for (org.keycloak.representations.idm.RoleRepresentation role : allKeycloakRoles) {
                    RoleRepresentation roleRepresentation = RoleRepresentation.
                            newBuilder().setName(role.getName())
                            .setComposite(role.isComposite())
                            .setId(role.getId())
                            .build();
                    if (role.getDescription() != null) {
                        roleRepresentation = roleRepresentation.toBuilder().setDescription(role.getDescription()).build();
                    }
                    roleRepresentations.add(roleRepresentation);

                }

                builder.addAllRoles(roleRepresentations);
                if (request.getClientLevel()) {
                    builder.setScope("client_level");
                } else {
                    builder.setScope("realm_level");
                }
            }

            responseObserver.onNext(builder.build());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Get roles   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addProtocolMapper(AddProtocolMapperRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to add protocol mapper " + request.getTenantId());

            String mapperModel = "oidc-usermodel-attribute-mapper";
            Map<String, String> configMap = new HashMap<>();
            if (request.getMapperType().equals(MapperTypes.USER_ATTRIBUTE)) {
                mapperModel = "oidc-usermodel-attribute-mapper";
                configMap.put("user.attribute", request.getAttributeName());
            } else if (request.getMapperType().equals(MapperTypes.USER_REALM_ROLE)) {
                mapperModel = "oidc-usermodel-realm-role-mapper";
            } else if (request.getMapperType().equals(MapperTypes.USER_CLIENT_ROLE)) {
                mapperModel = "oidc-usermodel-client-role-mapper";
                configMap.put("usermodel.clientRoleMapping.clientId", request.getClientId());
            } else {
                responseObserver.onError(io.grpc.Status.UNIMPLEMENTED.
                        withDescription("Mapping type not supported").asRuntimeException());
                return;
            }

            ProtocolMapperRepresentation protocolMapperRepresentation = new ProtocolMapperRepresentation();
            protocolMapperRepresentation.setName(request.getName());
            protocolMapperRepresentation.setProtocol("openid-connect");
            protocolMapperRepresentation.setProtocolMapper(mapperModel);

            configMap.put("user.session.note", request.getClaimName());
            configMap.put("id.token.claim", String.valueOf(request.getAddToIdToken()));
            configMap.put("access.token.claim", String.valueOf(request.getAddToAccessToken()));
            configMap.put("claim.name", request.getClaimName());
            switch (request.getClaimType()) {
                case JSON:
                    configMap.put("jsonType.label", "JSON");
                    break;
                case LONG:
                    configMap.put("jsonType.label", "long");
                    break;
                case STRING:
                    configMap.put("jsonType.label", "String");
                    break;
                case BOOLEAN:
                    configMap.put("jsonType.label", "boolean");
                    break;
                case INTEGER:
                    configMap.put("jsonType.label", "int");
                    break;
                default: {
                    responseObserver.onError(io.grpc.Status.INVALID_ARGUMENT.
                            withDescription("Unknown claim type").asRuntimeException());
                    return;
                }
            }

            configMap.put("aggregate.attrs", String.valueOf(request.getAggregateAttributeValues()));
            configMap.put("userinfo.token.claim", String.valueOf(request.getAddToUserInfo()));
            configMap.put("multivalued", String.valueOf(request.getMultiValued()));


            protocolMapperRepresentation.setConfig(configMap);

            keycloakClient.addProtocolMapper(protocolMapperRepresentation, String.valueOf(request.getTenantId()),
                    request.getClientId());

            statusUpdater.updateStatus(IAMOperations.ADD_PROTOCOL_MAPPER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));

            org.apache.custos.iam.service.OperationStatus status =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_PROTOCOL_MAPPER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            String msg = " Add protocol mapper   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addUserAttributes(AddUserAttributesRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to addUserAttributes " + request.getTenantId());

            List<UserAttribute> attributes = request.getAttributesList();

            List<String> validUserNames = new ArrayList<>();

            for (String username : request.getUsersList()) {
                boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                        username, request.getAccessToken());

                if (status) {
                    validUserNames.add(username);
                }
            }

            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                if (attribute.getKey().equals(Constants.CUSTOS_REALM_AGENT)) {
                    String msg = Constants.CUSTOS_REALM_AGENT + " cannot be used as a valid attribute";
                    LOGGER.error(msg);
                    responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
                }

                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, validUserNames);

            statusUpdater.updateStatus(IAMOperations.ADD_USER_ATTRIBUTE.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            org.apache.custos.iam.service.OperationStatus status =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_USER_ATTRIBUTE.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Add attributes   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void deleteUserAttributes(DeleteUserAttributeRequest request,
                                     StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to delete user attributes " + request.getTenantId());

            List<String> validUserNames = new ArrayList<>();

            for (String username : request.getUsersList()) {
                boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                        username, request.getAccessToken());

                if (status) {
                    validUserNames.add(username);
                }
            }


            List<UserAttribute> attributes = request.getAttributesList();

            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                attributeMap.put(attribute.getKey(), attribute.getValuesList());

            }

            keycloakClient.deleteUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, validUserNames);

            statusUpdater.updateStatus(IAMOperations.DELETE_USER_ATTRIBUTES.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            org.apache.custos.iam.service.OperationStatus status =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.DELETE_USER_ATTRIBUTES.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Add attributes   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void configureEventPersistence(EventPersistenceRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to configureEventPersistence " + request.getTenantId());

            keycloakClient.configureEventPersistence(
                    String.valueOf(request.getTenantId()),
                    request.getEvent(),
                    request.getPersistenceTime(),
                    request.getEnable(),
                    request.getAdminEvent()
            );
            org.apache.custos.iam.service.OperationStatus status =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.CONFIGURE_PERSISTANCE.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Configure Event Persistence   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void createGroups(GroupsRequest request, StreamObserver<GroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to createGroup " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();

            List<org.keycloak.representations.idm.GroupRepresentation> groupRepresentations =
                    transformToKeycloakGroups(request.getClientId(), request.getGroupsList());

            List<org.keycloak.representations.idm.GroupRepresentation> representations =
                    keycloakClient.createGroups(String.valueOf(tenantId), request.getClientId(), request.getClientSec(), groupRepresentations);

            List<GroupRepresentation> groups = transformKeycloakGroupsToGroups(request.getClientId(), representations);

            statusUpdater.updateStatus(IAMOperations.CREATE_GROUP.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            GroupsResponse response = GroupsResponse.newBuilder().addAllGroups(groups).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.CREATE_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Create Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void updateGroup(GroupRequest request, StreamObserver<GroupRepresentation> responseObserver) {
        try {
            LOGGER.debug("Request received to updateGroup " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();

            List<GroupRepresentation> representations = new ArrayList<>();
            representations.add(request.getGroup());
            List<org.keycloak.representations.idm.GroupRepresentation> groupRepresentations =
                    transformToKeycloakGroups(request.getClientId(), representations);

            org.keycloak.representations.idm.GroupRepresentation groupRepresentation =
                    keycloakClient.updateGroup(String.valueOf(tenantId), request.getClientId(), request.getClientSec(), groupRepresentations.get(0));

            GroupRepresentation group = transformKeycloakGroupToGroup(request.getClientId(), groupRepresentation, null);

            statusUpdater.updateStatus(IAMOperations.UPDATE_GROUP.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            responseObserver.onNext(group);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.UPDATE_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Update Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void deleteGroup(GroupRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteGroup " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();

            keycloakClient.deleteGroup(String.valueOf(tenantId)
                    , request.getClientId(), request.getClientSec(), request.getGroup().getId());


            statusUpdater.updateStatus(IAMOperations.DELETE_GROUP.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            org.apache.custos.iam.service.OperationStatus operationStatus = org.apache.custos.iam.service.OperationStatus
                    .newBuilder().setStatus(true).build();

            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.DELETE_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Delete Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void findGroup(GroupRequest request, StreamObserver<GroupRepresentation> responseObserver) {
        try {
            LOGGER.debug("Request received to findGroup " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();


            org.keycloak.representations.idm.GroupRepresentation groupRepresentation =
                    keycloakClient.findGroup(String.valueOf(tenantId)
                            , accessToken, request.getGroup().getId(), request.getGroup().getName());

            if (groupRepresentation != null) {
                GroupRepresentation group = transformKeycloakGroupToGroup(request.getClientId(), groupRepresentation, null);
                responseObserver.onNext(group);
                responseObserver.onCompleted();
            } else {
                responseObserver.onNext(null);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = " find Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void getAllGroups(GroupRequest request, StreamObserver<GroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllGroups " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();


            List<org.keycloak.representations.idm.GroupRepresentation> groupRepresentation =
                    keycloakClient.getAllGroups(String.valueOf(tenantId)
                            , accessToken);

            List<GroupRepresentation> groups = transformKeycloakGroupsToGroups(request.getClientId(), groupRepresentation);


            if (groups != null && !groups.isEmpty()) {
                GroupsResponse response = GroupsResponse.newBuilder().addAllGroups(groups).build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();

            } else {
                GroupsResponse response = GroupsResponse.newBuilder().build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = " Get Groups   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void addUserToGroup(UserGroupMappingRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllGroups " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();


            boolean status = keycloakClient.
                    addUserToGroup(String.valueOf(tenantId), request.getUsername(), request.getGroupId(), accessToken);

            if (status) {
                statusUpdater.updateStatus(IAMOperations.ADD_USER_TO_GROUP.name(),
                        OperationStatus.SUCCESS,
                        request.getTenantId(),
                        request.getPerformedBy());
            } else {
                statusUpdater.updateStatus(IAMOperations.ADD_USER_TO_GROUP.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(),
                        request.getPerformedBy());
            }

            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus
                    .newBuilder()
                    .setStatus(status)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_USER_TO_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = "  Groups   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }


    }

    @Override
    public void removeUserFromGroup(UserGroupMappingRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllGroups " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();


            boolean status = keycloakClient.
                    removeUserFromGroup(String.valueOf(tenantId), request.getUsername(), request.getGroupId(), accessToken);


            if (status) {
                statusUpdater.updateStatus(IAMOperations.REMOVE_USER_FROM_GROUP.name(),
                        OperationStatus.SUCCESS,
                        request.getTenantId(),
                        request.getPerformedBy());
            } else {
                statusUpdater.updateStatus(IAMOperations.REMOVE_USER_FROM_GROUP.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(),
                        request.getPerformedBy());
            }


            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus
                    .newBuilder()
                    .setStatus(status)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {

            statusUpdater.updateStatus(IAMOperations.REMOVE_USER_FROM_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());

            String msg = "  Remove user from Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }


    }


    @Override
    public void createAgentClient(AgentClientMetadata request, StreamObserver<SetUpTenantResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to configureAgentClient " + request.getTenantId());

            KeycloakClientSecret secret = keycloakClient.configureClient(
                    String.valueOf(request.getTenantId()),
                    request.getClientName(),
                    request.getTenantURL(),
                    request.getRedirectURIsList());

            if (secret != null) {

                statusUpdater.updateStatus(IAMOperations.CONFIGURE_AGENT_CLIENT.name(),
                        OperationStatus.SUCCESS,
                        request.getTenantId(),
                        request.getPerformedBy());

                SetUpTenantResponse response = SetUpTenantResponse.newBuilder()
                        .setClientId(secret.getClientId())
                        .setClientSecret(secret.getClientSecret())
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                String msg = " Configure agent client  failed for " + request.getTenantId();
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.CONFIGURE_AGENT_CLIENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Configure agent client  failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void configureAgentClient(AgentClientMetadata request,
                                     StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to configureAgentClient " + request.getTenantId());

            boolean status = keycloakClient.configureAgentClient(String.valueOf(request.getTenantId()), request.getClientName(),
                    request.getAccessTokenLifeTime());

            org.apache.custos.iam.service.OperationStatus operationStatus =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Register and configureAgentClient user   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.REGISTER_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void registerAndEnableAgent(RegisterUserRequest request, StreamObserver<RegisterUserResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to registerAndEnableAgent " + request.getTenantId());

            boolean status = keycloakClient.createUser(String.valueOf(request.getTenantId()),
                    request.getUser().getId(),
                    request.getUser().getPassword(),
                    null,
                    null,
                    null,
                    false,
                    request.getAccessToken());

            if (status) {

                List<String> userList = new ArrayList<>();
                userList.add(request.getUser().getId());
                if (!request.getUser().getRealmRolesList().isEmpty()) {
                    keycloakClient.addRolesToUsers(request.getAccessToken(),
                            String.valueOf(request.getTenantId()), userList, request.getUser().getRealmRolesList(),
                            request.getClientId(), false);
                }
                if (!request.getUser().getClientRolesList().isEmpty()) {
                    keycloakClient.addRolesToUsers(request.getAccessToken(),
                            String.valueOf(request.getTenantId()), userList, request.getUser().getClientRolesList(),
                            request.getClientId(), true);
                }

                if (!request.getUser().getAttributesList().isEmpty()) {

                    Map<String, List<String>> map = new HashMap<>();
                    for (UserAttribute attribute : request.getUser().getAttributesList()) {

                        map.put(attribute.getKey(), attribute.getValuesList());
                    }

                    keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()),
                            request.getAccessToken(), map, userList);

                }

                status = keycloakClient.enableUserAccount(String.valueOf(request.getTenantId()),
                        request.getAccessToken(), request.getUser().getId());
                statusUpdater.updateStatus(IAMOperations.REGISTER_AGENT.name(),
                        OperationStatus.SUCCESS,
                        request.getTenantId(), request.getPerformedBy());

                RegisterUserResponse response = RegisterUserResponse.newBuilder().setIsRegistered(status).build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();

            } else {
                String msg = " Register and enable user   failed for  user " + request.getUser().getId() + "of tenant"
                        + request.getTenantId();
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = " Register and enable user   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.REGISTER_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void deleteAgent(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteAgent " + request.getTenantId());

            boolean status = keycloakClient.deleteUser(request.getAccessToken(),
                    String.valueOf(request.getTenantId()), request.getUser().getId());

            statusUpdater.updateStatus(IAMOperations.DELETE_AGENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            org.apache.custos.iam.service.OperationStatus operationStatus =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();

            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Delete agent  failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void disableAgent(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to disableAgent " + request.getTenantId());
            boolean status = keycloakClient.disableUserAccount(
                    String.valueOf(request.getTenantId()), request.getAccessToken(), request.getUser().getId());

            statusUpdater.updateStatus(IAMOperations.DISABLE_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());

            org.apache.custos.iam.service.OperationStatus operationStatus =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();

            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Disable agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_USER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void isAgentNameAvailable(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to isAgentNameAvailable " + request.getTenantId());
            boolean status = keycloakClient.isUsernameAvailable(
                    String.valueOf(request.getTenantId()), request.getUser().getId(), request.getAccessToken());

            org.apache.custos.iam.service.OperationStatus operationStatus =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();

            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Is agent name available   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void addAgentAttributes(AddUserAttributesRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to addAgentAttributes " + request.getTenantId());

            List<UserAttribute> attributes = request.getAttributesList();

            Map<String, List<String>> attributeMap = new HashMap<>();

            for (UserAttribute attribute : attributes) {
                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            boolean status = keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(),
                    attributeMap, request.getAgentsList());

            statusUpdater.updateStatus(IAMOperations.ADD_AGENT_ATTRIBUTES.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            org.apache.custos.iam.service.OperationStatus response =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " Add agent attributes   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.ADD_AGENT_ATTRIBUTES.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void deleteAgentAttributes(DeleteUserAttributeRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteAgentAttributes " + request.getTenantId());
            List<UserAttribute> attributes = request.getAttributesList();

            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            boolean status = keycloakClient.deleteUserAttributes
                    (String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, request.getAgentsList());
            statusUpdater.updateStatus(IAMOperations.DELETE_AGENT_ATTRIBUTES.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            org.apache.custos.iam.service.OperationStatus response =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Delete agent attributes   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_AGENT_ATTRIBUTES.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void addRolesToAgent(AddUserRolesRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to addRolesToAgent " + request.getTenantId());
            boolean status = keycloakClient.addRolesToUsers(request.getAccessToken(), String.valueOf(request.getTenantId()),
                    request.getAgentsList(), request.getRolesList(), request.getClientId(), request.getClientLevel());
            statusUpdater.updateStatus(IAMOperations.ADD_ROLE_TO_AGENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(status).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Add roles to agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.ADD_ROLE_TO_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void deleteAgentRoles(DeleteUserRolesRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteRolesFromAgent " + request.getTenantId());

            if (!request.getRolesList().isEmpty()) {

                keycloakClient.removeRoleFromUser(request.getAccessToken(),
                        String.valueOf(request.getTenantId()), request.getId(),
                        request.getRolesList(), request.getClientId(), false);
            }

            if (!request.getClientRolesList().isEmpty()) {
                keycloakClient.removeRoleFromUser(request.getAccessToken(),
                        String.valueOf(request.getTenantId()), request.getId(),
                        request.getClientRolesList(), request.getClientId(), true);

            }
            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_AGENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            org.apache.custos.iam.service.OperationStatus response = org.apache.custos.iam.service.OperationStatus.
                    newBuilder().setStatus(true).build();


            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Delete roles from agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void getAgent(UserSearchRequest request, StreamObserver<Agent> responseObserver) {
        try {
            LOGGER.debug("Request received to getAgent " + request.getTenantId());

            UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()), request.getAccessToken(),
                    request.getUser().getId());

            if (representation != null) {
                if (representation.getAttributes() == null || representation.getAttributes().isEmpty() ||
                        representation.getAttributes().get(Constants.CUSTOS_REALM_AGENT).get(0) == null ||
                        !representation.getAttributes().get(Constants.CUSTOS_REALM_AGENT).get(0).equals("true")) {
                    responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription("Agent not found ").asRuntimeException());
                    return;
                } else {

                    Agent agent = getAgent(representation);
                    responseObserver.onNext(agent);
                    responseObserver.onCompleted();
                }
            }
            {
                responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription("Agent not found ").asRuntimeException());
                return;
            }


        } catch (Exception ex) {
            String msg = " Delete roles from agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void enableAgent(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to enableAgent " + request.getTenantId());
            boolean status = keycloakClient.enableUserAccount(
                    String.valueOf(request.getTenantId()), request.getAccessToken(), request.getUser().getId());

            statusUpdater.updateStatus(IAMOperations.ENABLE_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());

            org.apache.custos.iam.service.OperationStatus operationStatus =
                    org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();

            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Enable agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.ENABLE_AGENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void grantAdminPrivilege(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to grantAdminPrivilege " + request.getTenantId());

            boolean validationStatus = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());

            if (validationStatus) {

                boolean status = keycloakClient.grantAdminPrivilege(String.valueOf(request.getTenantId()), request.getUser().getUsername());

                statusUpdater.updateStatus(IAMOperations.GRANT_ADMIN_PRIVILEGE.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), request.getPerformedBy());

                org.apache.custos.iam.service.OperationStatus operationStatus =
                        org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();

                responseObserver.onNext(operationStatus);
                responseObserver.onCompleted();
            } else {
                String msg = " Not a valid user";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = " Grant admin privilege " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.GRANT_ADMIN_PRIVILEGE.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void removeAdminPrivilege(UserSearchRequest request, StreamObserver<org.apache.custos.iam.service.OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to removeAdminPrivilege " + request.getTenantId());

            boolean validationStatus = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());

            if (validationStatus) {

                boolean status = keycloakClient.removeAdminPrivilege(String.valueOf(request.getTenantId()), request.getUser().getUsername());

                statusUpdater.updateStatus(IAMOperations.REMOVE_ADMIN_PRIVILEGE.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), request.getPerformedBy());

                org.apache.custos.iam.service.OperationStatus operationStatus =
                        org.apache.custos.iam.service.OperationStatus.newBuilder().setStatus(status).build();

                responseObserver.onNext(operationStatus);
                responseObserver.onCompleted();
            } else {
                String msg = " Not a valid user";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = " Remove admin privilege " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.REMOVE_ADMIN_PRIVILEGE.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void getAllResources(GetAllResources request, StreamObserver<GetAllResourcesResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllResources for tenant " + request.getTenantId());

            List<UserRepresentation> representations = keycloakClient.getAllUsers(String.valueOf(request.getTenantId()));
            GetAllResourcesResponse resourcesResponse = GetAllResourcesResponse.newBuilder().build();
            if (!representations.isEmpty()) {
                if (request.getResourceType().name().equals(ResourceTypes.USER.name())) {
                    List<org.apache.custos.iam.service.UserRepresentation> users = new ArrayList<>();
                    for (UserRepresentation userRepresentation : representations) {

                        boolean validationStatus = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                                userRepresentation.getUsername());
                        if (validationStatus) {
                            users.add(getUser(userRepresentation, request.getClientId()));

                        }

                    }

                    resourcesResponse = resourcesResponse.toBuilder().addAllUsers(users).build();
                    responseObserver.onNext(resourcesResponse);
                    responseObserver.onCompleted();

                } else {
                    List<Agent> agents = new ArrayList<>();
                    for (UserRepresentation userRepresentation : representations) {
                        boolean validationStatus = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                                userRepresentation.getUsername());
                        if (!validationStatus) {
                            agents.add(getAgent(userRepresentation));
                        }
                    }

                    resourcesResponse = resourcesResponse.toBuilder().addAllAgents(agents).build();
                    responseObserver.onNext(resourcesResponse);
                    responseObserver.onCompleted();

                }

            } else {
                String msg = " Empty resources";
                LOGGER.error(msg);
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = " Get all resources failed";
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                responseObserver.onError(io.grpc.Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    private OperationMetadata convertFromEntity(StatusEntity entity) {
        return OperationMetadata.newBuilder()
                .setEvent(entity.getEvent())
                .setStatus(entity.getState())
                .setPerformedBy(entity.getPerformedBy())
                .setTimeStamp(entity.getTime().toString()).build();
    }


    private Agent getAgent(UserRepresentation representation) {

        Agent.Builder builder = Agent.newBuilder().setId(representation.getUsername())
                .setIsEnabled(representation.isEnabled())
                .setCreationTime(representation.getCreatedTimestamp());

        for (String key : representation.getAttributes().keySet()) {
            UserAttribute attribute = UserAttribute.newBuilder().setKey(key)
                    .addAllValues(representation.getAttributes().get(key))
                    .build();
            builder.addAttributes(attribute);
        }

        if (representation.getRealmRoles() != null && !representation.getRealmRoles().isEmpty()) {
            builder.addAllRealmRoles(representation.getRealmRoles());
        }

        if (representation.getClientRoles() != null && !representation.getClientRoles().isEmpty() &&
                representation.getClientRoles().get(Constants.AGENT_CLIENT) != null &&
                !representation.getClientRoles().get(Constants.AGENT_CLIENT).isEmpty()) {
            builder.addAllClientRoles(representation.getClientRoles().get(Constants.AGENT_CLIENT));
        }

        return builder.build();
    }


    private org.apache.custos.iam.service.UserRepresentation getUser(UserRepresentation representation, String clientId) {
        String state = Status.PENDING_CONFIRMATION;
        if (representation.isEnabled()) {
            state = Status.ACTIVE;
        } else if (representation.isEmailVerified()) {
            state = Status.CONFIRMED;
        }

        Map<String, List<String>> attributes = representation.getAttributes();

        List<UserAttribute> attributeList = new ArrayList<>();

        if (attributes != null && !attributes.isEmpty()) {
            for (String key : attributes.keySet()) {
                UserAttribute attribute = UserAttribute.
                        newBuilder()
                        .setKey(key)
                        .addAllValues(attributes.get(key)).build();
                attributeList.add(attribute);
            }
        }

        org.apache.custos.iam.service.UserRepresentation.Builder builder = org.apache.custos.iam.service.UserRepresentation.newBuilder()
                .setUsername(representation.getUsername())
                .setFirstName(representation.getFirstName())
                .setLastName(representation.getLastName())
                .setState(state)
                .setCreationTime(representation.getCreatedTimestamp())
                .setEmail(representation.getEmail());


        if (representation.getAttributes() != null && !representation.getAttributes().isEmpty()) {
            builder.addAllAttributes(attributeList);
        }


        if (representation.getRealmRoles() != null && !representation.getRealmRoles().isEmpty()) {
            builder.addAllRealmRoles(representation.getRealmRoles());
        }

        if (representation.getClientRoles() != null && representation.getClientRoles().get(clientId) != null &&
                !representation.getClientRoles().get(clientId).isEmpty()) {
            builder.addAllClientRoles(representation.getClientRoles().get(clientId));
        }

        return builder.build();

    }

    private List<org.keycloak.representations.idm.GroupRepresentation> transformToKeycloakGroups(String clientId, List<GroupRepresentation> groupRepresentations) {
        List<org.keycloak.representations.idm.GroupRepresentation> groupsList = new ArrayList<>();
        for (GroupRepresentation groupRepresentation : groupRepresentations) {
            groupsList.add(transformSingleGroupToKeycloakGroup(clientId, groupRepresentation, null));
        }
        return groupsList;
    }

    private org.keycloak.representations.idm.GroupRepresentation
    transformSingleGroupToKeycloakGroup(String clientId, GroupRepresentation groupRepresentation,
                                        org.keycloak.representations.idm.GroupRepresentation parentGroup) {
        String name = groupRepresentation.getName();
        String id = groupRepresentation.getId();

        if (groupRepresentation.getOwnerId() != null && !groupRepresentation.getOwnerId().equals("")) {

            groupRepresentation = groupRepresentation.toBuilder().addAttributes(UserAttribute.newBuilder()
                    .setKey(Constants.OWNER_ID).addValues(groupRepresentation.getOwnerId()).build()).build();
        }

        if (groupRepresentation.getDescription() != null && !groupRepresentation.getDescription().equals("")) {

            groupRepresentation = groupRepresentation.toBuilder().addAttributes(UserAttribute.newBuilder()
                    .setKey(Constants.DESCRIPTION).addValues(groupRepresentation.getDescription()).build()).build();
        }


        List<UserAttribute> attributeList = groupRepresentation.getAttributesList();
        List<String> realmRoles = groupRepresentation.getRealmRolesList();
        List<String> clientRoles = groupRepresentation.getClientRolesList();

        org.keycloak.representations.idm.GroupRepresentation keycloakGroup =
                new org.keycloak.representations.idm.GroupRepresentation();

        keycloakGroup.setName(name);

        if (id != null && !id.trim().equals("")) {
            keycloakGroup.setId(id);
        }


        Map<String, List<String>> map = new HashMap<>();
        if (attributeList != null && !attributeList.isEmpty()) {
            for (UserAttribute attribute : attributeList) {
                map.put(attribute.getKey(), attribute.getValuesList());
            }
            keycloakGroup.setAttributes(map);
        }


        if (realmRoles != null && !realmRoles.isEmpty()) {
            keycloakGroup.setRealmRoles(realmRoles);

        }

        Map<String, List<String>> clientMap = new HashMap<>();
        if (clientRoles != null && !clientRoles.isEmpty()) {
            clientMap.put(clientId, clientRoles);
            keycloakGroup.setClientRoles(clientMap);
        }

        if (groupRepresentation.getSubGroupsList() != null && !groupRepresentation.getSubGroupsList().isEmpty()) {
            for (GroupRepresentation representation : groupRepresentation.getSubGroupsList()) {
                transformSingleGroupToKeycloakGroup(clientId, representation, keycloakGroup);
            }

        }

        if (parentGroup != null) {
            List<org.keycloak.representations.idm.GroupRepresentation> groupRepList = parentGroup.getSubGroups();
            if (groupRepList == null) {
                groupRepList = new ArrayList<>();
            }
            String path = parentGroup.getPath() + "/" + keycloakGroup.getName();
            keycloakGroup.setPath(path);
            groupRepList.add(keycloakGroup);
            parentGroup.setSubGroups(groupRepList);
            return parentGroup;
        }

        String path = "/" + keycloakGroup.getName();
        keycloakGroup.setPath(path);

        return keycloakGroup;
    }

    private List<GroupRepresentation> transformKeycloakGroupsToGroups
            (String clientId, List<org.keycloak.representations.idm.GroupRepresentation> groupRepresentations) {
        List<GroupRepresentation> groupsList = new ArrayList<>();
        for (org.keycloak.representations.idm.GroupRepresentation groupRepresentation : groupRepresentations) {
            groupsList.add(transformKeycloakGroupToGroup(clientId, groupRepresentation, null));
        }
        return groupsList;
    }

    private GroupRepresentation transformKeycloakGroupToGroup(String clientId,
                                                              org.keycloak.representations.idm.GroupRepresentation group,
                                                              GroupRepresentation parent) {
        String name = group.getName();
        String id = group.getId();
        List<String> realmRoles = group.getRealmRoles();
        Map<String, List<String>> clientRoles = group.getClientRoles();
        Map<String, List<String>> atrs = group.getAttributes();

        GroupRepresentation representation = GroupRepresentation.newBuilder()
                .setName(name)
                .setId(id)
                .build();

        if (realmRoles != null && !realmRoles.isEmpty()) {
            representation = representation.toBuilder().addAllRealmRoles(realmRoles).build();
        }

        if (clientRoles != null && !clientRoles.isEmpty() && !clientRoles.get(clientId).isEmpty()) {

            representation = representation.toBuilder().addAllClientRoles(clientRoles.get(clientId)).build();
        }

        if (atrs != null && !atrs.isEmpty()) {
            List<UserAttribute> attributeList = new ArrayList<>();
            for (String key : atrs.keySet()) {
                if (key.equals(Constants.OWNER_ID)) {
                    representation = representation.toBuilder().setOwnerId(atrs.get(key).get(0)).build();
                } else if (key.equals(Constants.DESCRIPTION)) {
                    representation = representation.toBuilder().setDescription(atrs.get(key).get(0)).build();
                } else {
                    UserAttribute attribute = UserAttribute.newBuilder().setKey(key).addAllValues(atrs.get(key)).build();
                    attributeList.add(attribute);
                }
            }
            representation = representation.toBuilder().addAllAttributes(attributeList).build();

        }

        if (group.getSubGroups() != null && !group.getSubGroups().isEmpty()) {
            for (org.keycloak.representations.idm.GroupRepresentation subGroup : group.getSubGroups()) {
                representation = transformKeycloakGroupToGroup(clientId, subGroup, representation);
            }

        }


        if (parent != null) {
            parent = parent.toBuilder().addSubGroups(representation).build();
        }

        if (parent != null) {
            return parent;
        } else {
            return representation;
        }

    }


}
