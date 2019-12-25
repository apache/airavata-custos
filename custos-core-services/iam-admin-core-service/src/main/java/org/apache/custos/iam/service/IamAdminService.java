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
import org.apache.custos.federated.services.clients.keycloak.KeycloakClient;
import org.apache.custos.federated.services.clients.keycloak.KeycloakClientSecret;
import org.apache.custos.iam.exceptions.IAMAdminServiceException;
import org.apache.custos.iam.persistance.model.IAMEventMetadata;
import org.apache.custos.iam.persistance.repository.EventRepository;
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
    private EventRepository repository;

    @Override
    public void setUPTenant(SetUpTenantRequest request, StreamObserver<SetUpTenantResponse> responseObserver) {
        IAMEventMetadata iamEventMetadata = new IAMEventMetadata();
        iamEventMetadata.setEvent(IAMOperations.SET_UP_TENANT);
        iamEventMetadata.setPerformedBy(request.getRequesterEmail());

        try {
            LOGGER.debug("Request received to setUPTenant for " + request.getTenantId());

            keycloakClient.createRealm(request.getTenantId(), request.getTenantName());

            keycloakClient.createRealmAdminAccount(request.getTenantId(), request.getAdminUsername(),
                    request.getAdminFirstname(), request.getAdminLastname(),
                    request.getAdminEmail(), request.getAdminPassword());

            KeycloakClientSecret clientSecret = keycloakClient.configureClient(request.getTenantId(),
                    request.getTenantURL());

            SetUpTenantResponse response = SetUpTenantResponse.newBuilder()
                    .setClientId(clientSecret.getClientId())
                    .setClientSecret(clientSecret.getClientSecret())
                    .build();


            iamEventMetadata.setState(Status.SUCCESS);
            repository.save(iamEventMetadata);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during setUPTenant" + ex;
            LOGGER.error(msg, ex);
            iamEventMetadata.setState(Status.FAILED);
            repository.save(iamEventMetadata);
            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void isUsernameAvailable(IsUsernameAvailableRequest request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUsernameAvailable for " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUsernameAvailable(request.getTenantId(),
                    request.getUserName(),
                    request.getAccessToken());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isAvailable).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during isUsernameAvailable" + ex;
            LOGGER.error(msg, ex);
            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void isUserEnabled(UserAccessInfo request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUserEnabled for " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUserAccountEnabled(request.getTenantId(),
                    request.getAccessToken(),
                    request.getUsername());
            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isAvailable).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during isUserEnabled" + ex;
            LOGGER.error(msg, ex);
            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void addRoleToUser(RoleOperationsUserRequest request, StreamObserver<CheckingResponse> responseObserver) {
        IAMEventMetadata iamEventMetadata = new IAMEventMetadata();
        iamEventMetadata.setEvent(IAMOperations.ADD_ROLE_TO_USER);
        String userId = request.getAdminUsername() +"@"+request.getTenantId();
        iamEventMetadata.setPerformedBy(userId);
        try {
            LOGGER.debug("Request received to addRoleToUser for " + request.getTenantId());

            boolean isAdded = keycloakClient.addRoleToUser(request.getAdminUsername(),
                    request.getPassword(),
                    request.getTenantId(),
                    request.getUsername(),
                    request.getRole());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isAdded).build();


            iamEventMetadata.setState(Status.SUCCESS);
            repository.save(iamEventMetadata);
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during addRoleToUser" + ex;
            LOGGER.error(msg, ex);

            iamEventMetadata.setState(Status.FAILED);
            repository.save(iamEventMetadata);

            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void registerUser(RegisterUserRequest request, StreamObserver<RegisterUserResponse> responseObserver) {

        IAMEventMetadata iamEventMetadata = new IAMEventMetadata();
        iamEventMetadata.setEvent(IAMOperations.REGISTER_USER);
        iamEventMetadata.setPerformedBy(request.getTenantId());


        try {
            LOGGER.debug("Request received to registerUser for " + request.getTenantId());

            boolean registered = keycloakClient.createUser(request.getTenantId(),
                    request.getUsername(),
                    request.getPassword(),
                    request.getFirstName(),
                    request.getLastName(),
                    request.getEmail(),
                    request.getAccessToken());

            RegisterUserResponse registerUserResponse = RegisterUserResponse.newBuilder().
                    setIsRegistered(registered).build();


            iamEventMetadata.setState(Status.SUCCESS);
            repository.save(iamEventMetadata);

            responseObserver.onNext(registerUserResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during registerUser" + ex;
            LOGGER.error(msg, ex);

            iamEventMetadata.setState(Status.FAILED);
            repository.save(iamEventMetadata);

            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void enableUser(UserAccessInfo request, StreamObserver<User> responseObserver) {

        IAMEventMetadata iamEventMetadata = new IAMEventMetadata();
        iamEventMetadata.setEvent(IAMOperations.ENABLE_USER);
        iamEventMetadata.setPerformedBy(request.getTenantId());

        try {
            LOGGER.debug("Request received to enableUser for " + request.getTenantId());

            boolean accountEnabled = keycloakClient.enableUserAccount(request.getTenantId(),
                    request.getAccessToken(), request.getUsername());
            if (accountEnabled) {

                UserRepresentation representation = keycloakClient.getUser(request.getTenantId(),
                        request.getAccessToken(), request.getUsername());

                User user = getUser(representation);


                iamEventMetadata.setState(Status.SUCCESS);
                repository.save(iamEventMetadata);


                responseObserver.onNext(user);
                responseObserver.onCompleted();

            } else {

                iamEventMetadata.setState(Status.FAILED);
                repository.save(iamEventMetadata);

                IAMAdminServiceException exception = new IAMAdminServiceException("Account enabling failed ", null);
                responseObserver.onError(exception);
            }


        } catch (Exception ex) {
            String msg = "Error occurred during enableUser" + ex;
            LOGGER.error(msg, ex);

            iamEventMetadata.setState(Status.FAILED);
            repository.save(iamEventMetadata);

            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void isUserExist(UserAccessInfo request, StreamObserver<CheckingResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to isUserExist for " + request.getTenantId());

            boolean isUserExist = keycloakClient.isUserExist(request.getTenantId(),
                    request.getAccessToken(), request.getUsername());
            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isUserExist).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during isUserExist" + ex;
            LOGGER.error(msg, ex);
            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void getUser(UserAccessInfo request, StreamObserver<User> responseObserver) {
        try {
            LOGGER.debug("Request received to getUser for " + request.getTenantId());

            UserRepresentation representation = keycloakClient.getUser(request.getTenantId(),
                    request.getAccessToken(), request.getUsername());

            User user = getUser(representation);
            responseObserver.onNext(user);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during getUser" + ex;
            LOGGER.error(msg, ex);
            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void getUsers(GetUsersRequest request, StreamObserver<GetUsersResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getUsers for " + request.getInfo().getUsername());

            List<UserRepresentation> representation = keycloakClient.getUsers(request.getInfo().getAccessToken(),
                    request.getInfo().getTenantId(), request.getOffset(), request.getLimit(), request.getSearch());
            List<User> users = new ArrayList<>();
            representation.stream().forEach(r -> users.add(this.getUser(r)));


            GetUsersResponse response = GetUsersResponse.newBuilder().addAllUser(users).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during getUsers" + ex;
            LOGGER.error(msg, ex);
            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void resetPassword(ResetUserPassword request, StreamObserver<CheckingResponse> responseObserver) {
        IAMEventMetadata iamEventMetadata = new IAMEventMetadata();
        iamEventMetadata.setEvent(IAMOperations.RESET_PASSWORD);
        iamEventMetadata.setPerformedBy(request.getInfo().getUsername()+"@"+request.getInfo().getTenantId());
        try {
            LOGGER.debug("Request received to resetPassword for " + request.getInfo().getUsername());

            boolean isChanged = keycloakClient.resetUserPassword(request.getInfo().getAccessToken(),
                    request.getInfo().getTenantId(),
                    request.getInfo().getUsername(),
                    request.getPassword());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isChanged).build();


            iamEventMetadata.setState(Status.SUCCESS);
            repository.save(iamEventMetadata);


            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during resetPassword" + ex;
            LOGGER.error(msg, ex);

            iamEventMetadata.setState(Status.FAILED);
            repository.save(iamEventMetadata);

            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void findUsers(FindUsersRequest request, StreamObserver<GetUsersResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to findUsers for " + request.getInfo().getTenantId());

            List<UserRepresentation> representations = keycloakClient.findUser(request.getInfo().getAccessToken(),
                    request.getInfo().getTenantId(),
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
            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void updateUserProfile(UpdateUserProfileRequest request, StreamObserver<CheckingResponse> responseObserver) {
        IAMEventMetadata iamEventMetadata = new IAMEventMetadata();
        iamEventMetadata.setEvent(IAMOperations.UPDATE_USER_PROFILE);
        iamEventMetadata.setPerformedBy(request.getUser().getUsername()+"@"+request.getUser().getTenantId());
        try {
            LOGGER.debug("Request received to updateUserProfile for " + request.getUser().getUsername());

            keycloakClient.updateUserRepresentation(request.getAccessToken(),
                    request.getUser().getTenantId(),
                    request.getUser().getUsername(),
                    request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(true).build();


            iamEventMetadata.setState(Status.SUCCESS);
            repository.save(iamEventMetadata);


            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred during updateUserProfile" + ex;
            LOGGER.error(msg, ex);

            iamEventMetadata.setState(Status.FAILED);
            repository.save(iamEventMetadata);

            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void deleteUser(UserAccessInfo request, StreamObserver<CheckingResponse> responseObserver) {
        IAMEventMetadata iamEventMetadata = new IAMEventMetadata();
        iamEventMetadata.setEvent(IAMOperations.DELETE_USER);
        iamEventMetadata.setPerformedBy(request.getUsername()+"@"+request.getTenantId());
        try {
            LOGGER.debug("Request received to deleteUser for " + request.getTenantId());

            boolean isUpdated = keycloakClient.deleteUser(request.getAccessToken(),
                    request.getTenantId(), request.getUsername());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isUpdated).build();

            iamEventMetadata.setState(Status.SUCCESS);
            repository.save(iamEventMetadata);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred during deleteUser" + ex;
            LOGGER.error(msg, ex);

            iamEventMetadata.setState(Status.FAILED);
            repository.save(iamEventMetadata);

            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
    }

    @Override
    public void deleteRoleFromUser(RoleOperationsUserRequest request, StreamObserver<CheckingResponse> responseObserver) {

        IAMEventMetadata iamEventMetadata = new IAMEventMetadata();
        iamEventMetadata.setEvent(IAMOperations.DELETE_ROLE_FROM_USER);
        iamEventMetadata.setPerformedBy(request.getAdminUsername()+"@"+request.getTenantId());
        try {
            LOGGER.debug("Request received to deleteRoleFromUser for " + request.getTenantId());


            boolean isRemoved = keycloakClient.removeRoleFromUser(request.getAdminUsername(), request.getPassword(),
                    request.getTenantId(), request.getUsername(), request.getRole());

            CheckingResponse response = CheckingResponse.newBuilder().setIsExist(isRemoved).build();


            iamEventMetadata.setState(Status.SUCCESS);
            repository.save(iamEventMetadata);

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred during deleteRoleFromUser" + ex;
            LOGGER.error(msg, ex);

            iamEventMetadata.setState(Status.FAILED);
            repository.save(iamEventMetadata);

            IAMAdminServiceException exception = new IAMAdminServiceException(msg, ex);
            responseObserver.onError(exception);
        }
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
