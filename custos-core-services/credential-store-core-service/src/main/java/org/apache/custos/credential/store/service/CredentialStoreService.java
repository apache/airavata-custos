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

package org.apache.custos.credential.store.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.core.services.commons.StatusUpdater;
import org.apache.custos.core.services.commons.persistance.model.StatusEntity;
import org.apache.custos.credential.store.credential.CredentialManager;
import org.apache.custos.credential.store.model.Credential;
import org.apache.custos.credential.store.model.CredentialTypes;
import org.apache.custos.credential.store.persistance.model.AgentCredentialEntity;
import org.apache.custos.credential.store.persistance.model.CredentialEntity;
import org.apache.custos.credential.store.persistance.repository.AgentCredentialRepository;
import org.apache.custos.credential.store.persistance.repository.CredentialRepository;
import org.apache.custos.credential.store.service.CredentialStoreServiceGrpc.CredentialStoreServiceImplBase;
import org.apache.custos.credential.store.utils.Operations;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.vault.core.VaultTemplate;
import org.springframework.vault.support.VaultResponseSupport;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

/**
 * This is responsible for vault operations
 */
@GRpcService
public class CredentialStoreService extends CredentialStoreServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(CredentialStoreService.class);

    private static final String BASE_PATH = "/secret/";

    @Autowired
    private VaultTemplate vaultTemplate;

    @Autowired
    private StatusUpdater statusUpdater;

    @Autowired
    private CredentialManager credentialManager;

    @Autowired
    private CredentialRepository repository;

    @Autowired
    private AgentCredentialRepository agentCredentialRepository;

    @Override
    public void putCredential(CredentialMetadata request, StreamObserver<OperationStatus> responseObserver) {

        try {
            LOGGER.debug("Calling putSecret API for owner " + request.getOwnerId() + " with credentials Id "
                    + request.getId() + " Secret " + request.getSecret());
            String path = BASE_PATH + request.getOwnerId() + "/" + request.getType().name();
            Credential credential = new Credential(request.getId(), request.getSecret());
            vaultTemplate.write(path, credential);
            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);
            if (response != null && response.getData() != null && response.getData().getId() != null) {
                OperationStatus secretStatus = OperationStatus.newBuilder().setState(true).build();

                statusUpdater.updateStatus(Operations.PUT_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.SUCCESS,
                        request.getOwnerId(),
                        null);


                responseObserver.onNext(secretStatus);
                responseObserver.onCompleted();
            } else {

                String msg = "PutSecret operation failed for "
                        + request.getOwnerId() + "with credentials Id "
                        + request.getId() + " Secret " + request.getSecret() + " Type " + request.getType().name();
                LOGGER.error(msg);

                statusUpdater.updateStatus(Operations.PUT_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                responseObserver.onError(Status.ABORTED.withDescription(msg).asRuntimeException());
            }
        } catch (Exception ex) {
            String msg = "PutSecret operation failed for "
                    + request.getOwnerId() + "with credentials Id "
                    + request.getId() + " Secret " + request.getSecret() + " Type " + request.getType().name();
            LOGGER.error(msg);

            statusUpdater.updateStatus(Operations.PUT_CREDENTIAL.name(),
                    org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                    request.getOwnerId(),
                    null);

            responseObserver.onError(Status.ABORTED.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getCredential(GetCredentialRequest request, StreamObserver<CredentialMetadata> responseObserver) {
        try {
            LOGGER.debug("Calling getSecret API for owner " + request.getOwnerId() + " for type " + request.getType());
            String path = BASE_PATH + request.getOwnerId() + "/" + request.getType().name();
            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null) {
                String msg = "Cannot find credentials for "
                        + request.getOwnerId() + " for type " + request.getType();
                CredentialMetadata secret = CredentialMetadata.newBuilder().build();
                responseObserver.onNext(secret);
                responseObserver.onCompleted();
                return;

            }
            Credential credential = response.getData();
            CredentialMetadata.Builder secret = CredentialMetadata.newBuilder()
                    .setSecret(credential.getSecret())
                    .setId(credential.getId())
                    .setOwnerId(request.getOwnerId());


            if (request.getType() == Type.CUSTOS) {
                Optional<CredentialEntity> entity = repository.findById(request.getOwnerId());
                if (entity.isPresent()) {
                    secret.setClientIdIssuedAt(entity.get().getIssuedAt().getTime());
                    secret.setClientSecretExpiredAt(entity.get().getClientSecretExpiredAt());
                }
            }

            responseObserver.onNext(secret.build());
            responseObserver.onCompleted();

        } catch (Exception ex) {

            String msg = " operation failed for "
                    + request.getOwnerId() + "with credentials Id "
                    + request.getId() + " Type " + request.getType().name();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllCredentials(GetAllCredentialsRequest request, StreamObserver<GetAllCredentialsResponse>
            responseObserver) {
        try {
            LOGGER.debug("Calling getAllSecrets API for owner " + request.getOwnerId());

            String subPath = BASE_PATH + request.getOwnerId();
            List<String> paths = vaultTemplate.list(subPath);

            List<CredentialMetadata> credentialMetadata = new ArrayList<>();

            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    if (isMainType(key)) {
                        String path = subPath + "/" + key;
                        VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);
                        CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), request.getOwnerId(), key);
                        credentialMetadata.add(metadata);
                    }
                }
            }
            GetAllCredentialsResponse response = GetAllCredentialsResponse.newBuilder()
                    .addAllSecretList(credentialMetadata).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " operation failed for " + request.getOwnerId() + " " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteCredential(DeleteCredentialRequest request, StreamObserver<OperationStatus> responseObserver) {

        try {

            LOGGER.debug("Calling deleteSecret API for owner " + request.getOwnerId());


            String subPath = BASE_PATH + request.getOwnerId() + (request.getType() != null ? "/" +
                    request.getType().name() : "");

            List<String> paths = vaultTemplate.list(subPath);

            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    String path = subPath + "/" + key;
                    vaultTemplate.delete(path);

                }
            }

            OperationStatus secretStatus = OperationStatus.newBuilder().setState(true).build();

            statusUpdater.updateStatus(Operations.DELETE_CREDENTIAL.name(),
                    org.apache.custos.core.services.commons.persistance.model.OperationStatus.SUCCESS,
                    request.getOwnerId(),
                    null);
            responseObserver.onNext(secretStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " operation failed for " + request.getOwnerId();

            statusUpdater.updateStatus(Operations.DELETE_CREDENTIAL.name(),
                    org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                    request.getOwnerId(),
                    null);

            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
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
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getNewCustosCredential(GetNewCustosCredentialRequest request, StreamObserver<CredentialMetadata> responseObserver) {
        try {
            LOGGER.debug("Generating custos credentials for  " + request.getOwnerId());

            Credential credential = credentialManager.generateCredential(request.getOwnerId(), CredentialTypes.CUSTOS, 0);

            String path = BASE_PATH + request.getOwnerId() + "/" + CredentialTypes.CUSTOS.name();

            vaultTemplate.write(path, credential);

            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null || response.getData() == null || response.getData().getId() == null) {
                String msg = "Get new custos operation failed for "
                        + request.getOwnerId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_CUSTOS_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }


            statusUpdater.updateStatus(Operations.GENERATE_CUSTOS_CREDENTIAL.name(),
                    org.apache.custos.core.services.commons.persistance.model.OperationStatus.SUCCESS,
                    request.getOwnerId(),
                    null);

            Optional<CredentialEntity> entity = repository.findById(request.getOwnerId());

            if (entity.isEmpty()) {
                String msg = "Credential is not persisted"
                        + request.getOwnerId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_CUSTOS_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }

            CredentialEntity en = entity.get();
            CredentialMetadata rep = CredentialMetadata
                    .newBuilder()
                    .setOwnerId(request.getOwnerId())
                    .setSecret(credential.getSecret())
                    .setClientIdIssuedAt(en.getIssuedAt().getTime())
                    .setClientSecretExpiredAt(en.getClientSecretExpiredAt())
                    .setId(credential.getId())
                    .build();
            responseObserver.onNext(rep);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " operation failed for " + request.getOwnerId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getOwnerIdFromToken(TokenRequest request, StreamObserver<GetOwnerIdResponse> responseObserver) {
        try {
            LOGGER.debug("Get ownerId for    " + request.getToken());
            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("" +
                        "Invalid access token");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());


            if (entity != null) {
                GetOwnerIdResponse response = GetOwnerIdResponse.
                        newBuilder().setOwnerId(entity.getOwnerId()).build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                LOGGER.error("User not found");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = " operation failed ";
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getCustosCredentialFromToken(TokenRequest request, StreamObserver<CredentialMetadata> responseObserver) {
        try {
            LOGGER.debug("Get credential for    " + request.getToken());
            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("" +
                        "Invalid access token");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());

            if (entity == null) {
                LOGGER.error("User not found");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            String path = BASE_PATH + entity.getOwnerId() + "/" + Type.CUSTOS.name();
            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null || response.getData() == null || !response.getData().getSecret().equals(credential.getSecret())) {
                String msg = "Invalid secret for id"
                        + credential.getId();
                LOGGER.error(msg);
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
                return;
            }


            CredentialMetadata secret = CredentialMetadata.newBuilder()
                    .setSecret(credential.getSecret())
                    .setId(credential.getId())
                    .setOwnerId(entity.getOwnerId())
                    .setClientSecretExpiredAt(entity.getClientSecretExpiredAt())
                    .setClientIdIssuedAt(entity.getIssuedAt().getTime())
                    .setType(Type.CUSTOS).build();
            responseObserver.onNext(secret);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " operation failed ";
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getCustosCredentialFromClientId(GetCredentialRequest request, StreamObserver<CredentialMetadata> responseObserver) {
        try {
            String clientId = request.getId();

            CredentialEntity entity = repository.findByClientId(clientId);

            if (entity == null) {
                String msg = " Credentials not found for user " + clientId;
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            String path = BASE_PATH + entity.getOwnerId() + "/" + Type.CUSTOS.name();

            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null || response.getData() == null) {
                String msg = "Cannot find credentials for "
                        + entity.getOwnerId() + " for type " + Type.CUSTOS.name();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            CredentialMetadata metadata = CredentialMetadata.newBuilder()
                    .setSecret(response.getData().getSecret())
                    .setId(request.getId())
                    .setOwnerId(entity.getOwnerId())
                    .setClientSecretExpiredAt(entity.getClientSecretExpiredAt())
                    .setClientIdIssuedAt(entity.getIssuedAt().getTime())
                    .setSuperTenant(response.getData().isSuperTenant())
                    .setType(Type.CUSTOS).build();
            responseObserver.onNext(metadata);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " Operation failed  " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getAllCredentialsFromToken(TokenRequest request, StreamObserver<GetAllCredentialsResponse> responseObserver) {
        try {
            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());

            if (entity == null) {
                LOGGER.error(" client not found");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            String subPath = BASE_PATH + entity.getOwnerId();

            String validatingPath = BASE_PATH + entity.getOwnerId() + "/" + Type.CUSTOS.name();
            VaultResponseSupport<Credential> validationResponse = vaultTemplate.read(validatingPath, Credential.class);

            if (validationResponse == null || validationResponse.getData() == null || !validationResponse.getData().getSecret().equals(credential.getSecret())) {
                String msg = "Invalid secret for id"
                        + credential.getId();
                LOGGER.error(msg);
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
                return;
            }

            List<String> paths = vaultTemplate.list(subPath);

            List<CredentialMetadata> credentialMetadata = new ArrayList<>();


            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    if (isMainType(key)) {
                        String path = subPath + "/" + key;
                        VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);
                        CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), entity.getOwnerId(), key);


                        if (key.equals(Type.CUSTOS)) {
                            metadata = metadata.toBuilder()
                                    .setClientIdIssuedAt(entity.getIssuedAt().getTime())
                                    .setClientSecretExpiredAt(entity.getClientSecretExpiredAt())
                                    .build();
                        }
                        credentialMetadata.add(metadata);
                    }
                }
            }
            GetAllCredentialsResponse response = GetAllCredentialsResponse.newBuilder()
                    .addAllSecretList(credentialMetadata).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " Operation failed " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getBasicCredentials(TokenRequest request, StreamObserver<Credentials> responseObserver) {
        try {

            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());

            if (entity == null) {
                LOGGER.error("User not found");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }
            String subPath = BASE_PATH + entity.getOwnerId();
            List<String> paths = vaultTemplate.list(subPath);


            Credentials.Builder credentialsBuilder = Credentials.newBuilder();


            if (entity == null) {
                LOGGER.error("User not found");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }


            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    String path = subPath + "/" + key;
                    VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);
                    if (key.equals(Type.CUSTOS)) {
                        if (!crRe.getData().getSecret().equals(credential.getSecret())) {
                            String msg = "Invalid secret for id"
                                    + credential.getId();
                            LOGGER.error(msg);
                            responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
                            return;
                        }

                        credentialsBuilder.setCustosClientId(crRe.getData().getId())
                                .setCustosClientSecret(crRe.getData().getSecret())
                                .setCustosClientIdIssuedAt(entity.getIssuedAt().getTime())
                                .setCustosClientSecretExpiredAt(entity.getClientSecretExpiredAt());

                    } else if (key.equals(Type.IAM)) {
                        credentialsBuilder.setIamClientId(crRe.getData().getId())
                                .setIamClientSecret(crRe.getData().getSecret());


                    } else if (key.equals(Type.CILOGON)) {

                        credentialsBuilder.setCiLogonClientId(crRe.getData().getId())
                                .setCiLogonClientSecret(crRe.getData().getSecret());

                    }

                }
            }

            Credentials credentials = credentialsBuilder.build();

            responseObserver.onNext(credentials);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " Operation failed " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getAllCredentialsFromJWTToken(TokenRequest request, StreamObserver<GetAllCredentialsResponse> responseObserver) {
        try {
            String token = request.getToken();

            Credential credential = credentialManager.decodeJWTToken(token);
            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());

            if (entity == null) {
                LOGGER.error("User not found");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            String subPath = BASE_PATH + entity.getOwnerId();
            List<String> paths = vaultTemplate.list(subPath);

            List<CredentialMetadata> credentialMetadata = new ArrayList<>();

            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    if (isMainType(key)) {
                        String path = subPath + "/" + key;
                        VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);
                        CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), entity.getOwnerId(), key);

                        if (key.equals(Type.CUSTOS.name())) {
                            metadata = metadata.toBuilder()
                                    .setClientIdIssuedAt(entity.getIssuedAt().getTime())
                                    .setClientSecretExpiredAt(entity.getClientSecretExpiredAt())
                                    .setSuperAdmin(credential.isAdmin())
                                    .setSuperTenant(crRe.getData().isSuperTenant())
                                    .build();
                        }
                        credentialMetadata.add(metadata);
                    }
                }
            }

            GetAllCredentialsResponse response = GetAllCredentialsResponse.newBuilder()
                    .addAllSecretList(credentialMetadata)
                    .setRequesterUserEmail(credential.getEmail())
                    .setRequesterUsername(credential.getUsername())
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " Operation failed  " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getMasterCredentials(GetCredentialRequest request, StreamObserver<GetAllCredentialsResponse> responseObserver) {
        try {

            String subPath = BASE_PATH + "master";
            List<String> paths = vaultTemplate.list(subPath);

            List<CredentialMetadata> credentialMetadata = new ArrayList<>();


            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    if (isMainType(key)) {
                        String path = subPath + "/" + key;
                        VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);
                        CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), 0, key);
                        credentialMetadata.add(metadata);
                    }
                }
            }
            GetAllCredentialsResponse response = GetAllCredentialsResponse.newBuilder()
                    .addAllSecretList(credentialMetadata).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " Operation failed  " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void createAgentCredential(CredentialMetadata request, StreamObserver<CredentialMetadata> responseObserver) {
        try {
            LOGGER.debug(" Request received to createAgentCredential " + request.getOwnerId());

            AgentCredentialEntity exEntity = agentCredentialRepository.findByClientIdAndOwnerId(request.getId(),
                    request.getOwnerId());

            if (exEntity != null) {
                String msg = "Duplicate client Id " + request.getId() + " for "
                        + request.getOwnerId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_AGENT_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }


            if (isMainType(request.getId())) {
                String msg = "Prohibited agent Id  " + request.getId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_AGENT_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                responseObserver.onError(Status.INVALID_ARGUMENT.withDescription(msg).asRuntimeException());
                return;
            }

            Credential credential = credentialManager.generateAgentCredential(request.getOwnerId(), request.getId(),
                    CredentialTypes.AGENT, 0);

            String path = BASE_PATH + request.getOwnerId() + "/" + request.getId();

            vaultTemplate.write(path, credential);

            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null || response.getData() == null || response.getData().getId() == null) {
                String msg = "Get new agent credentials operation failed of "
                        + request.getOwnerId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_AGENT_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }


            AgentCredentialEntity entity = agentCredentialRepository.findByClientIdAndOwnerId(request.getId(),
                    request.getOwnerId());

            if (entity == null) {
                String msg = "Credential is not persisted of "
                        + request.getId() + "at tenant " + request.getOwnerId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_AGENT_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }


            statusUpdater.updateStatus(Operations.GENERATE_AGENT_CREDENTIAL.name(),
                    org.apache.custos.core.services.commons.persistance.model.OperationStatus.SUCCESS,
                    request.getOwnerId(),
                    null);

            CredentialMetadata rep = CredentialMetadata
                    .newBuilder()
                    .setOwnerId(request.getOwnerId())
                    .setSecret(credential.getSecret())
                    .setClientIdIssuedAt(entity.getIssuedAt().getTime())
                    .setClientSecretExpiredAt(entity.getClientSecretExpiredAt())
                    .setId(credential.getId())
                    .setInternalSec(credential.getPassword())
                    .build();
            responseObserver.onNext(rep);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " Create Agent credential operation is failed  " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteAgentCredential(CredentialMetadata request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received for deleteAgentCredential for  " + request.getId() + " at tenant "
                    + request.getOwnerId());

            String subPath = BASE_PATH + request.getOwnerId() + "/" + request.getId();

            List<String> paths = vaultTemplate.list(subPath);

            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    String path = subPath + "/" + key;
                    vaultTemplate.delete(path);

                }
            }

            AgentCredentialEntity entity = agentCredentialRepository.findByClientIdAndOwnerId(request.getId(), request.getOwnerId());

            if (entity != null) {
                agentCredentialRepository.delete(entity);
            }

            OperationStatus secretStatus = OperationStatus.newBuilder().setState(true).build();

            statusUpdater.updateStatus(Operations.DELETE_AGENT_CREDENTIAL.name(),
                    org.apache.custos.core.services.commons.persistance.model.OperationStatus.SUCCESS,
                    request.getOwnerId(),
                    null);
            responseObserver.onNext(secretStatus);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " Delete agent credential failed  " + ex;
            LOGGER.error(msg);
            statusUpdater.updateStatus(Operations.DELETE_AGENT_CREDENTIAL.name(),
                    org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                    request.getOwnerId(),
                    null);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAgentCredential(GetCredentialRequest request, StreamObserver<CredentialMetadata> responseObserver) {
        try {
            LOGGER.debug(" Request received for getAgentCredential " + request.getOwnerId());

            String clientId = request.getId();

            AgentCredentialEntity entity = agentCredentialRepository.findByClientIdAndOwnerId(request.getId(),
                    request.getOwnerId());

            if (entity == null) {
                String msg = " Credentials not found for user " + clientId;
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            String path = BASE_PATH + entity.getOwnerId() + "/" + request.getId();

            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null || response.getData() == null) {
                String msg = "Cannot find credentials for "
                        + entity.getClientId() + " at tenant " + request.getOwnerId();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            CredentialMetadata metadata = CredentialMetadata.newBuilder()
                    .setSecret(response.getData().getSecret())
                    .setId(request.getId())
                    .setInternalSec(response.getData().getPassword())
                    .setOwnerId(entity.getOwnerId())
                    .setType(Type.AGENT).build();
            responseObserver.onNext(metadata);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Operation failed  " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getCredentialByAgentBasicAuth(TokenRequest request, StreamObserver<GetAllCredentialsResponse> responseObserver) {
        try {
            LOGGER.debug(" Request received for getCredentialByAgentBasicAuth ");

            String token = request.getToken();

            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            CredentialEntity exEntity = repository.findByClientId(request.getParentClientId());

            if (exEntity != null) {

                AgentCredentialEntity entity = agentCredentialRepository.findByClientIdAndOwnerId(credential.getId(),
                        exEntity.getOwnerId());

                if (entity == null) {
                    LOGGER.error("Agent not found");
                    responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                    return;
                }
                String subPath = BASE_PATH + entity.getOwnerId();
                List<String> paths = vaultTemplate.list(subPath);

                List<CredentialMetadata> metadataList = new ArrayList<>();


                if (paths != null && !paths.isEmpty()) {
                    for (String key : paths) {
                        String path = subPath + "/" + key;
                        VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);

                        if (isMainType(key)) {
                            CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), entity.getOwnerId(), key);
                            metadataList.add(metadata);

                        } else if (key.equals(credential.getId())) {
                            if (!crRe.getData().getSecret().equals(credential.getSecret())) {
                                String msg = "Invalid secret for id"
                                        + credential.getId();
                                LOGGER.error(msg);
                                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
                                return;
                            }
                            CredentialMetadata metadata = CredentialMetadata
                                    .newBuilder()
                                    .setType(Type.AGENT)
                                    .setId(credential.getId())
                                    .setSecret(crRe.getData().getSecret())
                                    .setInternalSec(crRe.getData().getPassword())
                                    .build();
                            metadataList.add(metadata);
                        }
                    }
                }

                GetAllCredentialsResponse response = GetAllCredentialsResponse
                        .newBuilder()
                        .addAllSecretList(metadataList)
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();

            } else {
                String msg = " Cannot find a valid  client Id for tenant ";
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }
        } catch (Exception ex) {
            String msg = " Operation failed  " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getCredentialByAgentJWTToken(TokenRequest request, StreamObserver<GetAllCredentialsResponse> responseObserver) {
        try {
            LOGGER.debug(" Request received to getCredentialByAgentJWTToken ");

            String token = request.getToken();

            Credential credential = credentialManager.decodeAgentJWTToken(token);

            if (credential == null || credential.getId() == null || credential.getParentId() == null) {
                LOGGER.error("Invalid agent access token");
                responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                return;
            }

            CredentialEntity exEntity = repository.findByClientId(credential.getParentId());

            if (exEntity != null) {

                AgentCredentialEntity entity = agentCredentialRepository.findByClientIdAndOwnerId(credential.getId(),
                        exEntity.getOwnerId());

                if (entity == null) {
                    LOGGER.error("Agent not found with Id " + credential.getId());
                    responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                    return;
                }
                GetAllCredentialsResponse response = getAllCredentials(exEntity, credential);
                responseObserver.onNext(response);
                responseObserver.onCompleted();

            } else {
                String msg = " Cannot find a valid  tenant for agent " + credential.getId();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }
        } catch (Exception ex) {
            String msg = " Error occurred  reason:  " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void validateAgentJWTToken(TokenRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to validateAgentJWTToken ");

            String token = request.getToken();

            Credential credential = credentialManager.decodeAgentJWTToken(token);

            if (credential == null || credential.getId() == null || credential.getParentId() == null) {
                OperationStatus operationStatus = OperationStatus
                        .newBuilder().setState(false).build();
                responseObserver.onNext(operationStatus);
                responseObserver.onCompleted();
                return;
            }
            CredentialEntity exEntity = repository.findByClientId(credential.getParentId());

            if (exEntity == null) {
                OperationStatus operationStatus = OperationStatus
                        .newBuilder().setState(false).build();
                responseObserver.onNext(operationStatus);
                responseObserver.onCompleted();
                return;

            }

            AgentCredentialEntity entity = agentCredentialRepository.findByClientIdAndOwnerId(credential.getId(),
                    exEntity.getOwnerId());

            if (entity == null) {
                OperationStatus operationStatus = OperationStatus
                        .newBuilder().setState(false).build();
                responseObserver.onNext(operationStatus);
                responseObserver.onCompleted();
                return;
            }
            OperationStatus operationStatus = OperationStatus
                    .newBuilder().setState(true).build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            OperationStatus operationStatus = OperationStatus
                    .newBuilder().setState(false).build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();
        }
    }

    private OperationMetadata convertFromEntity(StatusEntity entity) {
        return OperationMetadata.newBuilder()
                .setEvent(entity.getEvent())
                .setStatus(entity.getState())
                .setPerformedBy(entity.getPerformedBy())
                .setTimeStamp(entity.getTime().toString()).build();
    }


    private CredentialMetadata convertToCredentialMetadata(Credential credential, long ownerId, String type) {

        return CredentialMetadata.newBuilder()
                .setOwnerId(ownerId)
                .setType(Type.valueOf(type))
                .setId(credential.getId())
                .setSuperTenant(credential.isSuperTenant())
                .setSecret(credential.getSecret())
                .build();
    }


    private boolean isMainType(String str) {
        for (Type type : Type.values()) {
            if (type.name().equalsIgnoreCase(str))
                return true;
        }
        return false;
    }


    private GetAllCredentialsResponse getAllCredentials(CredentialEntity entity, Credential credential) {
        String subPath = BASE_PATH + entity.getOwnerId();
        List<String> paths = vaultTemplate.list(subPath);

        List<CredentialMetadata> metadataList = new ArrayList<>();


        if (paths != null && !paths.isEmpty()) {
            for (String key : paths) {
                String path = subPath + "/" + key;
                VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);

                if (isMainType(key)) {
                    CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), entity.getOwnerId(), key);
                    metadataList.add(metadata);

                } else if (key.equals(credential.getId())) {
                    CredentialMetadata metadata = CredentialMetadata
                            .newBuilder()
                            .setType(Type.AGENT)
                            .setId(credential.getId())
                            .setSecret(crRe.getData().getSecret())
                            .setOwnerId(entity.getOwnerId())
                            .setInternalSec(crRe.getData().getPassword())
                            .build();
                    metadataList.add(metadata);
                }
            }
        }
        GetAllCredentialsResponse response = GetAllCredentialsResponse
                .newBuilder()
                .addAllSecretList(metadataList)
                .build();
        return response;
    }
}
