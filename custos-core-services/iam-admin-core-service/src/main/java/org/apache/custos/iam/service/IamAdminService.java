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
import org.apache.custos.federated.services.clients.keycloak.KeycloakClient;
import org.apache.custos.federated.services.clients.keycloak.KeycloakClientSecret;
import org.apache.custos.federated.services.clients.keycloak.UnauthorizedException;
import org.apache.custos.iam.service.IamAdminServiceGrpc.IamAdminServiceImplBase;
import org.apache.custos.iam.utils.IAMOperations;
import org.apache.custos.iam.utils.Status;
import org.keycloak.representations.idm.ProtocolMapperRepresentation;
import org.keycloak.representations.idm.UserRepresentation;
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
    public void isUsernameAvailable(UserSearchRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUsernameAvailable at " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUsernameAvailable(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(),
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
    public void isUserEnabled(UserSearchRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUserEnabled at " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUserAccountEnabled(String.valueOf(request.getTenantId()),
                    request.getAccessToken(),
                    request.getUser().getUsername());
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

            UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUser().getUsername());

            if (representation != null) {
                org.apache.custos.iam.service.UserRepresentation user = getUser(representation, request.getClientId());
                responseObserver.onNext(user);
                responseObserver.onCompleted();
            } else {
                String msg = "User " + request.getUser().getUsername() + "not found at " + request.getTenantId();
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

            List<UserRepresentation> representation = keycloakClient.getUsers(request.getAccessToken(),
                    String.valueOf(request.getTenantId()), request.getOffset(), request.getLimit(),
                    request.getUser().getUsername(), request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail());
            List<org.apache.custos.iam.service.UserRepresentation> users = new ArrayList<>();
            representation.stream().forEach(r -> users.add(this.getUser(r, request.getClientId())));


            FindUsersResponse response = FindUsersResponse.newBuilder().addAllUser(users).build();
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
    public void resetPassword(ResetUserPassword request, StreamObserver<CheckingResponse> responseObserver) {
        String userId = request.getUsername() + "@" + request.getTenantId();
        try {
            LOGGER.debug("Request received to resetPassword for " + request.getUsername());

            boolean isChanged = keycloakClient.resetUserPassword(request.getAccessToken(),
                    String.valueOf(request.getTenantId()),
                    request.getUsername(),
                    request.getPassword());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isChanged).build();


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
    public void updateUserProfile(UpdateUserProfileRequest request, StreamObserver<CheckingResponse> responseObserver) {
        String userId = request.getUser().getUsername() + "@" + request.getTenantId();

        try {
            LOGGER.debug("Request received to updateUserProfile for " + request.getUser().getUsername());

            keycloakClient.updateUserRepresentation(request.getAccessToken(),
                    String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(),
                    request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(true).build();


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
    public void deleteUser(UserSearchRequest request, StreamObserver<CheckingResponse> responseObserver) {

        try {
            LOGGER.debug("Request received to deleteUser for " + request.getTenantId());

            boolean isUpdated = keycloakClient.deleteUser(request.getAccessToken(),
                    String.valueOf(request.getTenantId()), request.getUser().getUsername());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isUpdated).build();

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

            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteRolesFromUser(DeleteUserRolesRequest request, StreamObserver<CheckingResponse> responseObserver) {

        try {
            LOGGER.debug("Request received to deleteRoleFromUser for " + request.getTenantId());


            if (!request.getRealmRolesList().isEmpty()) {

                keycloakClient.removeRoleFromUser(request.getAccessToken(),
                        String.valueOf(request.getTenantId()), request.getUsername(),
                        request.getRealmRolesList(), request.getClientId(), false);
            }

            if (!request.getClientRolesList().isEmpty()) {
                keycloakClient.removeRoleFromUser(request.getAccessToken(),
                        String.valueOf(request.getTenantId()), request.getUsername(),
                        request.getClientRolesList(), request.getClientId(), true);

            }
            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(true).build();


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

            keycloakClient.addRolesToUsers(request.getAccessToken(), String.valueOf(request.getTenantId()),
                    request.getUsernamesList(), request.getRolesList(), request.getClientId(), request.getClientLevel());

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
                configMap.put("usermodel.clientRoleMapping.clientId",request.getClientId());
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
            LOGGER.debug("Request received to add protocol mapper " + request.getTenantId());

            List<UserAttribute> attributes = request.getAttributesList();

            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, request.getUsersList());

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

            List<UserAttribute> attributes = request.getAttributesList();

            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            keycloakClient.deleteUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, request.getUsersList());

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

    private OperationMetadata convertFromEntity(StatusEntity entity) {
        return OperationMetadata.newBuilder()
                .setEvent(entity.getEvent())
                .setStatus(entity.getState())
                .setPerformedBy(entity.getPerformedBy())
                .setTimeStamp(entity.getTime().toString()).build();
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
}
