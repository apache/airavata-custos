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

import io.grpc.stub.StreamObserver;
import org.apache.custos.core.services.commons.StatusUpdater;
import org.apache.custos.core.services.commons.persistance.model.OperationStatus;
import org.apache.custos.core.services.commons.persistance.model.StatusEntity;
import org.apache.custos.federated.services.clients.keycloak.KeycloakClient;
import org.apache.custos.federated.services.clients.keycloak.KeycloakClientSecret;
import org.apache.custos.iam.service.IamAdminServiceGrpc.IamAdminServiceImplBase;
import org.apache.custos.iam.utils.IAMOperations;
import org.apache.custos.iam.utils.Status;
import org.keycloak.representations.idm.UserRepresentation;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.List;

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
            LOGGER.debug("Request received to setUPTenant for " + request.getTenantId());

            keycloakClient.createRealm(String.valueOf(request.getTenantId()), request.getTenantName());

            keycloakClient.createRealmAdminAccount(String.valueOf(request.getTenantId()), request.getAdminUsername(),
                    request.getAdminFirstname(), request.getAdminLastname(),
                    request.getAdminEmail(), request.getAdminPassword());

            KeycloakClientSecret clientSecret = keycloakClient.configureClient(String.valueOf(request.getTenantId()),
                    request.getTenantURL());

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
    public void isUsernameAvailable(IsUsernameAvailableRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUsernameAvailable for " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUsernameAvailable(String.valueOf(request.getTenantId()),
                    request.getUserName(),
                    request.getAccessToken());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isAvailable).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during isUsernameAvailable" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isUserEnabled(UserAccessInfo request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUserEnabled for " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUserAccountEnabled(String.valueOf(request.getTenantId()),
                    request.getAccessToken(),
                    request.getUsername());
            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isAvailable).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during isUserEnabled" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addRoleToUser(RoleOperationsUserRequest request, StreamObserver<CheckingResponse> responseObserver) {

        String userId = request.getAdminUsername() + "@" + request.getTenantId();

        try {
            LOGGER.debug("Request received to addRoleToUser for " + request.getTenantId());

            boolean isAdded = keycloakClient.addRoleToUser(request.getAdminUsername(),
                    request.getPassword(),
                    String.valueOf(request.getTenantId()),
                    request.getUsername(),
                    request.getRole());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isAdded).build();


            statusUpdater.updateStatus(IAMOperations.ADD_ROLE_TO_USER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    userId);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during addRoleToUser" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.ADD_ROLE_TO_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    userId);

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void registerUser(RegisterUserRequest request, StreamObserver<RegisterUserResponse> responseObserver) {

        try {
            LOGGER.debug("Request received to registerUser for " + request.getTenantId());

            boolean registered = keycloakClient.createUser(String.valueOf(request.getTenantId()),
                    request.getUsername(),
                    request.getPassword(),
                    request.getFirstName(),
                    request.getLastName(),
                    request.getEmail(),
                    request.getAccessToken());

            RegisterUserResponse registerUserResponse = RegisterUserResponse.newBuilder().
                    setIsRegistered(registered).build();


            statusUpdater.updateStatus(IAMOperations.REGISTER_USER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));

            responseObserver.onNext(registerUserResponse);
            responseObserver.onCompleted();

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
    public void enableUser(UserAccessInfo request, StreamObserver<User> responseObserver) {

        try {
            LOGGER.debug("Request received to enableUser for " + request.getTenantId());

            boolean accountEnabled = keycloakClient.enableUserAccount(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUsername());
            if (accountEnabled) {

                UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()),
                        request.getAccessToken(), request.getUsername());

                User user = getUser(representation);


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
    public void isUserExist(UserAccessInfo request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUserExist for " + request.getTenantId());

            boolean isUserExist = keycloakClient.isUserExist(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUsername());
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
    public void getUser(UserAccessInfo request, StreamObserver<User> responseObserver) {
        try {
            LOGGER.debug("Request received to getUser for " + request.getTenantId());

            UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUsername());

            User user = getUser(representation);
            responseObserver.onNext(user);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during getUser" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUsers(GetUsersRequest request, StreamObserver<GetUsersResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getUsers for " + request.getInfo().getUsername());

            List<UserRepresentation> representation = keycloakClient.getUsers(request.getInfo().getAccessToken(),
                    String.valueOf(request.getInfo().getTenantId()), request.getOffset(), request.getLimit(), request.getSearch());
            List<User> users = new ArrayList<>();
            representation.stream().forEach(r -> users.add(this.getUser(r)));


            GetUsersResponse response = GetUsersResponse.newBuilder().addAllUser(users).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during getUsers" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void resetPassword(ResetUserPassword request, StreamObserver<CheckingResponse> responseObserver) {
        String userId = request.getInfo().getUsername() + "@" + request.getInfo().getTenantId();
        try {
            LOGGER.debug("Request received to resetPassword for " + request.getInfo().getUsername());

            boolean isChanged = keycloakClient.resetUserPassword(request.getInfo().getAccessToken(),
                    String.valueOf(request.getInfo().getTenantId()),
                    request.getInfo().getUsername(),
                    request.getPassword());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isChanged).build();


            statusUpdater.updateStatus(IAMOperations.RESET_PASSWORD.name(),
                    OperationStatus.SUCCESS,
                    request.getInfo().getTenantId(), userId);


            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during resetPassword" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.RESET_PASSWORD.name(),
                    OperationStatus.FAILED,
                    request.getInfo().getTenantId(), userId);

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void findUsers(FindUsersRequest request, StreamObserver<GetUsersResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to findUsers for " + request.getInfo().getTenantId());

            List<UserRepresentation> representations = keycloakClient.findUser(request.getInfo().getAccessToken(),
                    String.valueOf(request.getInfo().getTenantId()),
                    request.getInfo().getUsername(),
                    request.getEmail());
            List<User> users = new ArrayList<>();
            representations.stream().forEach(r -> users.add(this.getUser(r)));


            GetUsersResponse response = GetUsersResponse.newBuilder().addAllUser(users).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during findUsers" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateUserProfile(UpdateUserProfileRequest request, StreamObserver<CheckingResponse> responseObserver) {
        String userId = request.getUser().getUsername() + "@" + request.getUser().getTenantId();

        try {
            LOGGER.debug("Request received to updateUserProfile for " + request.getUser().getUsername());

            keycloakClient.updateUserRepresentation(request.getAccessToken(),
                    String.valueOf(request.getUser().getTenantId()),
                    request.getUser().getUsername(),
                    request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(true).build();


            statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(),
                    OperationStatus.SUCCESS,
                    request.getUser().getTenantId(), userId);


            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during updateUserProfile" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(),
                    OperationStatus.FAILED,
                    request.getUser().getTenantId(), userId);

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteUser(UserAccessInfo request, StreamObserver<CheckingResponse> responseObserver) {

        String userId = request.getUsername() + "@" + request.getTenantId();

        try {
            LOGGER.debug("Request received to deleteUser for " + request.getTenantId());

            boolean isUpdated = keycloakClient.deleteUser(request.getAccessToken(),
                    String.valueOf(request.getTenantId()), request.getUsername());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isUpdated).build();

            statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), userId);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred during deleteUser" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), userId);

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteRoleFromUser(RoleOperationsUserRequest request, StreamObserver<CheckingResponse> responseObserver) {

        String userId = request.getAdminUsername() + "@" + request.getTenantId();

        try {
            LOGGER.debug("Request received to deleteRoleFromUser for " + request.getTenantId());


            boolean isRemoved = keycloakClient.removeRoleFromUser(request.getAdminUsername(), request.getPassword(),
                    String.valueOf(request.getTenantId()), request.getUsername(), request.getRole());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isRemoved).build();


            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_USER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), userId);


            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during deleteRoleFromUser" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), userId);

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
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

    private OperationMetadata convertFromEntity(StatusEntity entity) {
        return OperationMetadata.newBuilder()
                .setEvent(entity.getEvent())
                .setStatus(entity.getState())
                .setPerformedBy(entity.getPerformedBy())
                .setTimeStamp(entity.getTime().toString()).build();
    }


    private User getUser(UserRepresentation representation) {
        String state = Status.PENDING_CONFIRMATION;
        if (representation.isEnabled()) {
            state = Status.ACTIVE;
        } else if (representation.isEmailVerified()) {
            state = Status.CONFIRMED;
        }

        return User.newBuilder()
                .setInternalUserId(representation.getUsername() + "@" + representation.getId())
                .setUsername(representation.getUsername())
                .setFirstName(representation.getFirstName())
                .setLastName(representation.getLastName())
                .setState(state)
                .setCreationTime(representation.getCreatedTimestamp())
                .setEmail(representation.getEmail())
                .build();

    }
}
