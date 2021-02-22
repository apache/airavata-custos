/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.resource.secret.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.core.services.commons.StatusUpdater;
import org.apache.custos.core.services.commons.persistance.model.OperationStatus;
import org.apache.custos.resource.secret.manager.Credential;
import org.apache.custos.resource.secret.manager.CredentialGeneratorFactory;
import org.apache.custos.resource.secret.manager.adaptor.inbound.CredentialReader;
import org.apache.custos.resource.secret.manager.adaptor.outbound.CredentialWriter;
import org.apache.custos.resource.secret.utils.Operations;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.List;
import java.util.Optional;

/**
 * This class is responsible for handling resource secrets, such as SSH credentials, password credentials
 */
@GRpcService
public class ResourceSecretService extends ResourceSecretServiceGrpc.ResourceSecretServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(ResourceSecretService.class);


    @Autowired
    private StatusUpdater statusUpdater;

    @Autowired
    private CredentialGeneratorFactory credentialGeneratorFactory;

    @Autowired
    private CredentialWriter credentialWriter;

    @Autowired
    private CredentialReader credentialReader;


    @Override
    public void getAllResourceCredentialSummaries(GetResourceCredentialSummariesRequest request, StreamObserver<ResourceCredentialSummaries> responseObserver) {
        try {
            LOGGER.debug(" Request received to getAllResourceCredentialSummaries in tenant " + request.getTenantId() +
                    " of owner " + request.getOwnerId());

            List<SecretMetadata> metadataList = credentialReader.
                    getAllCredentialSummaries(request.getTenantId(), request.getAccessibleTokensList());

            ResourceCredentialSummaries resourceCredentialSummaries = ResourceCredentialSummaries.newBuilder()
                    .addAllMetadata(metadataList)
                    .build();

            responseObserver.onNext(resourceCredentialSummaries);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Exception occurred while fetching credential summaries  " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addSSHCredential(SSHCredential request, StreamObserver<AddResourceCredentialResponse> responseObserver) {
        try {
            LOGGER.debug(" Request received to addSSHCredential in tenant " + request.getMetadata().getTenantId() +
                    " of owner " + request.getMetadata().getOwnerId() + " with token  " + request.getMetadata().getToken());

            Credential credential = credentialGeneratorFactory.getCredential(request);
            org.apache.custos.resource.secret.manager.adaptor.outbound.SSHCredential sshCredential =
                    (org.apache.custos.resource.secret.manager.adaptor.outbound.SSHCredential) credential;

            credentialWriter.
                    saveSSHCredential(sshCredential);


            statusUpdater.updateStatus(Operations.ADD_SSH_CREDENTIAL.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());


            AddResourceCredentialResponse resourceCredentialResponse = AddResourceCredentialResponse
                    .newBuilder()
                    .setToken((sshCredential.getExternalId() != null &&
                            !sshCredential.getExternalId().trim().equals("")) ? sshCredential.getExternalId() : sshCredential.getToken())
                    .build();
            responseObserver.onNext(resourceCredentialResponse);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Exception occurred while adding SSH credentials " + request.getMetadata().getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addPasswordCredential(PasswordCredential request, StreamObserver<AddResourceCredentialResponse> responseObserver) {
        try {
            LOGGER.debug(" Request received to addPasswordCredential in tenant " + request.getMetadata().getTenantId() +
                    " of owner " + request.getMetadata().getOwnerId() + " with token  " + request.getMetadata().getToken());

            Credential credential = credentialGeneratorFactory.getCredential(request);
            org.apache.custos.resource.secret.manager.adaptor.outbound.PasswordCredential passwordCredential =
                    (org.apache.custos.resource.secret.manager.adaptor.outbound.PasswordCredential) credential;

            credentialWriter.
                    savePasswordCredential(passwordCredential);


            statusUpdater.updateStatus(Operations.ADD_PASSWORD_CREDENTIAL.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());

            AddResourceCredentialResponse resourceCredentialResponse = AddResourceCredentialResponse
                    .newBuilder()
                    .setToken((passwordCredential.getExternalId() != null &&
                            !passwordCredential.getExternalId().trim().equals("")) ? passwordCredential.getExternalId() : passwordCredential.getToken())
                    .build();
            responseObserver.onNext(resourceCredentialResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while adding password credentials " + request.getMetadata().getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addCertificateCredential(CertificateCredential request, StreamObserver<AddResourceCredentialResponse> responseObserver) {
        try {
            LOGGER.debug(" Request received to addCertificateCredential in tenant " + request.getMetadata().getTenantId() +
                    " of owner " + request.getMetadata().getOwnerId() + " with token  " + request.getMetadata().getToken());


            Credential credential = credentialGeneratorFactory.getCredential(request);
            org.apache.custos.resource.secret.manager.adaptor.outbound.CertificateCredential certificateCredential =
                    (org.apache.custos.resource.secret.manager.adaptor.outbound.CertificateCredential) credential;

            credentialWriter.
                    saveCertificateCredential(certificateCredential);

            statusUpdater.updateStatus(Operations.ADD_CERTIFICATE_CREDENTIAL.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());

            AddResourceCredentialResponse resourceCredentialResponse = AddResourceCredentialResponse
                    .newBuilder()
                    .setToken((certificateCredential.getExternalId() != null &&
                            !certificateCredential.getExternalId().trim().equals("")) ?
                            certificateCredential.getExternalId() : certificateCredential.getToken())
                    .build();
            responseObserver.onNext(resourceCredentialResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while adding certificate credential secret " + request.getMetadata().getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getResourceCredentialSummary(GetResourceCredentialByTokenRequest request, StreamObserver<SecretMetadata> responseObserver) {
        try {
            LOGGER.debug(" Request received to getResourceCredentialSummary in tenant " + request.getTenantId() +
                    " with token  " + request.getToken());

            Optional<SecretMetadata> metadata = credentialReader.getCredentialSummary(request.getTenantId(), request.getToken());

            responseObserver.onNext(metadata.get());
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Exception occurred while fetching resource credential summaries " + request.getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getSSHCredential(GetResourceCredentialByTokenRequest request, StreamObserver<SSHCredential> responseObserver) {
        try {
            LOGGER.debug(" Request received to getSSHCredential in tenant " + request.getTenantId() +
                    " with token  " + request.getToken());

            Optional<SSHCredential> sshCredential = credentialReader.getSSHCredential(request.getTenantId(), request.getToken());
            responseObserver.onNext(sshCredential.get());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while fetching SSH credential " + request.getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getPasswordCredential(GetResourceCredentialByTokenRequest request, StreamObserver<PasswordCredential> responseObserver) {
        try {
            LOGGER.debug(" Request received to getPasswordCredential in tenant " + request.getTenantId() +
                    " with token  " + request.getToken());

            Optional<PasswordCredential> passwordCredential = credentialReader.
                    getPasswordCredential(request.getTenantId(), request.getToken());
            responseObserver.onNext(passwordCredential.get());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while fetching password credential " + request.getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getCertificateCredential(GetResourceCredentialByTokenRequest request, StreamObserver<CertificateCredential> responseObserver) {
        try {
            LOGGER.debug(" Request received to getCertificateCredential in tenant " + request.getTenantId() +
                    " with token  " + request.getToken());

            Optional<CertificateCredential> certificateCredential = credentialReader.
                    getCertificateCredential(request.getTenantId(), request.getToken());
            responseObserver.onNext(certificateCredential.get());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while fetching certificate credential " + request.getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteSSHCredential(GetResourceCredentialByTokenRequest request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to deleteSSHCredential in tenant " + request.getTenantId() +
                    " with token  " + request.getToken());

            boolean status = credentialWriter.deleteCredential(request.getTenantId(), request.getToken());

            statusUpdater.updateStatus(Operations.DELETE_SSH_CREDENTIAL.name(), OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            ResourceCredentialOperationStatus operationStatus = ResourceCredentialOperationStatus
                    .newBuilder()
                    .setStatus(status)
                    .build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Exception occurred while deleting SSH secret " + request.getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deletePWDCredential(GetResourceCredentialByTokenRequest request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to deletePWDCredential in tenant " + request.getTenantId() +
                    " with token  " + request.getToken());

            boolean status = credentialWriter.deleteCredential(request.getTenantId(), request.getToken());

            statusUpdater.updateStatus(Operations.DELETE_PASSWORD_CREDENTIAL.name(), OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            ResourceCredentialOperationStatus operationStatus = ResourceCredentialOperationStatus
                    .newBuilder()
                    .setStatus(status)
                    .build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while deleting password credential " + request.getToken() +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getKVCredential(KVCredential request, StreamObserver<KVCredential> responseObserver) {
        try {
            LOGGER.debug(" Request received to getKVCredential in tenant " + request.getMetadata().getTenantId() +
                    " with key  " + request.getKey());
            String token = request.getToken();
            Optional<KVCredential> kvCredential = null;
            if (token != null && !token.trim().equals("")) {
                kvCredential = credentialReader.getKVSecretByToken(token,
                        request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());
            } else {
                kvCredential = credentialReader.getKVSecretByKey(request.getKey(),
                        request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());
            }

            responseObserver.onNext(kvCredential.get());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while fetching KV credentials " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void setKVCredential(KVCredential request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to setKVCredential in tenant " + request.getMetadata().getTenantId() +
                    " with key  " + request.getKey());
            Credential credential = credentialGeneratorFactory.getCredential(request);

            credentialWriter.saveKVCredential((org.apache.custos.resource.secret.manager.adaptor.outbound.KVCredential) credential);

            statusUpdater.updateStatus(Operations.SAVE_KV_CREDENTIAL.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());

            ResourceCredentialOperationStatus operationStatus = ResourceCredentialOperationStatus
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Exception occurred while  setting KV credentials " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateKVCredential(KVCredential request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to updateKVCredential in tenant " + request.getMetadata().getTenantId() +
                    " with key  " + request.getKey());
            credentialWriter.updateKVCredential(request);

            statusUpdater.updateStatus(Operations.UPDATE_KV_CREDENTIAL.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());

            ResourceCredentialOperationStatus operationStatus = ResourceCredentialOperationStatus
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Exception occurred while updating  KV credential " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteKVCredential(KVCredential request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to deleteKVCredential in tenant " + request.getMetadata().getTenantId() +
                    " with key  " + request.getKey());

            credentialWriter.deleteKVCredential(request);

            statusUpdater.updateStatus(Operations.DELETE_KV_CREDENTIAL.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());

            ResourceCredentialOperationStatus operationStatus = ResourceCredentialOperationStatus
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while deleting KV  credential " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getCredentialMap(CredentialMap request, StreamObserver<CredentialMap> responseObserver) {
        try {
            LOGGER.debug(" Request received to getCredentialMap in tenant " + request.getMetadata().getTenantId() +
                    " with key  " + request.getMetadata().getToken());

            Optional<CredentialMap> certificateCredential = credentialReader.
                    getCredentialMapByToken(request.getMetadata().getToken(), request.getMetadata().getTenantId());
            responseObserver.onNext(certificateCredential.get());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while deleting KV  credential " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void setCredentialMap(CredentialMap request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to setCredentialMap in tenant " + request.getMetadata().getTenantId() +
                    " with key  " + request.getMetadata().getToken());

            Credential credential = credentialGeneratorFactory.getCredential(request);

            credentialWriter.saveCredentialMap((org.apache.custos.resource.secret.manager.adaptor.outbound.CredentialMap) credential);

            statusUpdater.updateStatus(Operations.SAVE_CREDENTIAL_MAP.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());

            ResourceCredentialOperationStatus operationStatus = ResourceCredentialOperationStatus
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while deleting KV  credential " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void updateCredentialMap(CredentialMap request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to updateCredentialMap in tenant " + request.getMetadata().getTenantId() +
                    " with key  " + request.getMetadata().getToken());
            Credential credential = credentialGeneratorFactory.getCredential(request);
            credentialWriter.updateCredentialMap((org.apache.custos.resource.secret.manager.adaptor.outbound.CredentialMap) credential);

            statusUpdater.updateStatus(Operations.UPDATE_CREDENTIAL_MAP.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());

            ResourceCredentialOperationStatus operationStatus = ResourceCredentialOperationStatus
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while deleting KV  credential " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteCredentialMap(CredentialMap request, StreamObserver<ResourceCredentialOperationStatus> responseObserver) {
        try {
            LOGGER.debug(" Request received to deleteCredentialMap in tenant " + request.getMetadata().getTenantId() +
                    " with key  " + request.getMetadata().getToken());
            Credential credential = credentialGeneratorFactory.getCredential(request);
            credentialWriter.deleteCredentialMap((org.apache.custos.resource.secret.manager.adaptor.outbound.CredentialMap) credential);

            statusUpdater.updateStatus(Operations.DELETE_CREDENTIAL_MAP.name(), OperationStatus.SUCCESS,
                    request.getMetadata().getTenantId(), request.getMetadata().getOwnerId());

            ResourceCredentialOperationStatus operationStatus = ResourceCredentialOperationStatus
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Exception occurred while deleting KV  credential " +
                    " : " + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }
}
