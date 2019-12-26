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

import io.grpc.stub.StreamObserver;
import org.apache.custos.credential.store.exceptions.CredentialsNotFoundException;
import org.apache.custos.credential.store.exceptions.OperationFailedException;
import org.apache.custos.credential.store.model.Credential;
import org.apache.custos.credential.store.persistance.model.ServiceEvent;
import org.apache.custos.credential.store.persistance.repository.ServiceEventRepository;
import org.apache.custos.credential.store.service.CredentialStoreServiceGrpc.CredentialStoreServiceImplBase;
import org.apache.custos.credential.store.utils.Operations;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.vault.core.VaultTemplate;
import org.springframework.vault.support.VaultResponse;
import org.springframework.vault.support.VaultResponseSupport;

import java.util.ArrayList;
import java.util.List;

/**
 * This is responsible for vault operations
 */
public class CredentialStoreService extends CredentialStoreServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(CredentialStoreService.class);

    private static final String BASE_PATH = "/secret/";

    @Autowired
    private VaultTemplate vaultTemplate;

    @Autowired
    private ServiceEventRepository repository;

    @Override
    public void putCredential(CredentialMetadata request, StreamObserver<OperationStatus> responseObserver) {

        ServiceEvent serviceEvent = new ServiceEvent();
        serviceEvent.setEvent(Operations.PUT_CREDENTIAL.name());
        serviceEvent.setTraceId(request.getOwnerId());

        try {
            LOGGER.debug("Calling putSecret API for owner " + request.getOwnerId() + " with credentials Id "
                    + request.getId() + " Secret " + request.getSecret());
            String path = BASE_PATH + request.getOwnerId() + "/" + request.getType().name();
            Credential credential = new Credential(request.getId(), request.getSecret());
            VaultResponse response = vaultTemplate.write(path, credential);
            if (response != null) {
                OperationStatus secretStatus = OperationStatus.newBuilder().setState(true).build();

                serviceEvent.setState(org.apache.custos.credential.store.utils.OperationStatus.SUCCESS.name());
                repository.save(serviceEvent);

                responseObserver.onNext(secretStatus);
                responseObserver.onCompleted();
            } else {
                OperationFailedException failedException = new OperationFailedException("PutSecret operation failed for "
                        + request.getOwnerId() + "with credentials Id "
                        + request.getId() + " Secret " + request.getSecret() + " Type " + request.getType().name(), null);

                serviceEvent.setState(org.apache.custos.credential.store.utils.OperationStatus.FAILED.name());
                repository.save(serviceEvent);

                responseObserver.onError(failedException);
            }
        } catch (Exception ex) {
            OperationFailedException failedException = new OperationFailedException("PutSecret operation failed for "
                    + request.getOwnerId() + "with credentials Id "
                    + request.getId() + " Secret " + request.getSecret() + " Type " + request.getType().name(), ex);

            serviceEvent.setState(org.apache.custos.credential.store.utils.OperationStatus.FAILED.name());
            repository.save(serviceEvent);

            responseObserver.onError(failedException);
        }
    }

    @Override
    public void getCredential(GetCredentialsRequest request, StreamObserver<CredentialMetadata> responseObserver) {
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
                CredentialsNotFoundException credentialsNotFoundException =
                        new CredentialsNotFoundException("Cannot find credentials", null);
                responseObserver.onError(credentialsNotFoundException);
            }

        } catch (Exception ex) {
            OperationFailedException failedException = new OperationFailedException(" operation failed for "
                    + request.getOwnerId() + "with credentials Id "
                    + request.getId() + " Type " + request.getType().name(), ex);
            responseObserver.onError(failedException);
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
                    CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), request.getOwnerId(), path);
                    credentialMetadata.add(metadata);
                }
            }
            GetAllCredentialsResponse response = GetAllCredentialsResponse.newBuilder()
                    .addAllSecretList(credentialMetadata).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            OperationFailedException failedException = new OperationFailedException(" operation failed for "
                    + request.getOwnerId(), ex);
            responseObserver.onError(failedException);
        }
    }

    @Override
    public void deleteCredential(DeleteCredentialRequest request, StreamObserver<OperationStatus> responseObserver) {
        ServiceEvent serviceEvent = new ServiceEvent();
        serviceEvent.setEvent(Operations.DELETE_CREDENTIAL.name());
        serviceEvent.setTraceId(request.getOwnerId());
        try {

            LOGGER.debug("Calling deleteSecret API for owner " + request.getOwnerId());


            String path = BASE_PATH + request.getOwnerId() + (request.getType() != null ? "/" +
                    request.getType().name() : "");
            vaultTemplate.delete(path);
            OperationStatus secretStatus = OperationStatus.newBuilder().setState(true).build();

            serviceEvent.setState(org.apache.custos.credential.store.utils.OperationStatus.SUCCESS.name());
            repository.save(serviceEvent);

            responseObserver.onNext(secretStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            OperationFailedException failedException = new OperationFailedException(" operation failed for "
                    + request.getOwnerId(), ex);

            serviceEvent.setState(org.apache.custos.credential.store.utils.OperationStatus.FAILED.name());
            repository.save(serviceEvent);

            responseObserver.onError(failedException);
        }
    }


    private CredentialMetadata convertToCredentialMetadata(Credential credential, String ownerId, String type) {

        return CredentialMetadata.newBuilder()
                .setOwnerId(ownerId)
                .setType(Type.valueOf(type))
                .setId(credential.getId())
                .setSecret(credential.getSecret())
                .build();
    }


}
