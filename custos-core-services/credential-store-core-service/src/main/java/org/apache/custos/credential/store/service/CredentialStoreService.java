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
import org.apache.custos.credential.store.persistance.model.CredentialEntity;
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

            if (response != null) {
                Credential credential = response.getData();
                CredentialMetadata secret = CredentialMetadata.newBuilder()
                        .setSecret(credential.getSecret())
                        .setId(credential.getId())
                        .setOwnerId(request.getOwnerId())
                        .setType(request.getType()).build();
                responseObserver.onNext(secret);
                responseObserver.onCompleted();
            } else {
                String msg = "Cannot find credentials for "
                        + request.getOwnerId() + " for type " + request.getType();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }

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
                    String path = subPath + "/" + key;
                    VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);
                    CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), request.getOwnerId(), key);
                    credentialMetadata.add(metadata);
                }
            }
            GetAllCredentialsResponse response = GetAllCredentialsResponse.newBuilder()
                    .addAllSecretList(credentialMetadata).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = " operation failed for " + request.getOwnerId()+ " " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteCredential(DeleteCredentialRequest request, StreamObserver<OperationStatus> responseObserver) {

        try {

            LOGGER.debug("Calling deleteSecret API for owner " + request.getOwnerId());


            String path = BASE_PATH + request.getOwnerId() + (request.getType() != null ? "/" +
                    request.getType().name() : "");
            vaultTemplate.delete(path);
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
    public void getNewCustosCredential(GetNewCustosCredentialRequest request, StreamObserver<GetNewCustosCredentialResponse> responseObserver) {
        try {
            LOGGER.debug("Generating custos credentials for  " + request.getOwnerId());

            Credential credential = credentialManager.generateCredential(request.getOwnerId(), CredentialTypes.CUSTOS, 0);

            String path = BASE_PATH + request.getOwnerId() + "/" + CredentialTypes.CUSTOS.name();

            vaultTemplate.write(path, credential);

            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);
            if (response != null && response.getData() != null && response.getData().getId() != null) {

                statusUpdater.updateStatus(Operations.GENERATE_CUSTOS_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.SUCCESS,
                        request.getOwnerId(),
                        null);

                GetNewCustosCredentialResponse rep = GetNewCustosCredentialResponse
                        .newBuilder()
                        .setClientId(credential.getId())
                        .setClientSecret(credential.getSecret())
                        .build();
                responseObserver.onNext(rep);
                responseObserver.onCompleted();
            } else {

                String msg = "get new custos operation failed for "
                        + request.getOwnerId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_CUSTOS_CREDENTIAL.name(),
                        org.apache.custos.core.services.commons.persistance.model.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = " operation failed for " + request.getOwnerId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getOwnerIdFromToken(GetOwnerIdFromTokenRequest request, StreamObserver<GetOwnerIdResponse> responseObserver) {
        try {
            LOGGER.debug("Get ownerId for    " + request.getToken());
            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);
            if (credential != null && credential.getId() != null) {
                CredentialEntity entity = repository.findByClientId(credential.getId());
                if (entity != null) {
                    GetOwnerIdResponse response = GetOwnerIdResponse.
                            newBuilder().setOwnerId(entity.getOwnerId()).build();
                    responseObserver.onNext(response);
                    responseObserver.onCompleted();
                } else {
                    responseObserver.onError(Status.NOT_FOUND.asRuntimeException());
                }
            } else {
                responseObserver.onNext(null);
                responseObserver.onCompleted();
            }


        } catch (Exception ex) {
            String msg = " operation failed ";
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
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
                .setSecret(credential.getSecret())
                .build();
    }


}
