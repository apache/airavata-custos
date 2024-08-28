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

package org.apache.custos.service.credential.store;

import org.apache.custos.core.commons.StatusUpdater;
import org.apache.custos.core.credential.store.api.CredentialMetadata;
import org.apache.custos.core.credential.store.api.Credentials;
import org.apache.custos.core.credential.store.api.DeleteCredentialRequest;
import org.apache.custos.core.credential.store.api.GetAllCredentialsRequest;
import org.apache.custos.core.credential.store.api.GetAllCredentialsResponse;
import org.apache.custos.core.credential.store.api.GetCredentialRequest;
import org.apache.custos.core.credential.store.api.GetNewCustosCredentialRequest;
import org.apache.custos.core.credential.store.api.GetOperationsMetadataRequest;
import org.apache.custos.core.credential.store.api.GetOperationsMetadataResponse;
import org.apache.custos.core.credential.store.api.GetOwnerIdResponse;
import org.apache.custos.core.credential.store.api.OperationMetadata;
import org.apache.custos.core.credential.store.api.OperationStatus;
import org.apache.custos.core.credential.store.api.TokenRequest;
import org.apache.custos.core.credential.store.api.Type;
import org.apache.custos.core.model.commons.StatusEntity;
import org.apache.custos.core.model.credential.store.CredentialEntity;
import org.apache.custos.core.repo.credential.store.CredentialRepository;
import org.apache.custos.service.exceptions.AuthenticationException;
import org.apache.custos.service.exceptions.InternalServerException;
import jakarta.persistence.EntityNotFoundException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.vault.core.VaultTemplate;
import org.springframework.vault.support.VaultResponseSupport;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

/**
 * The CredentialStoreService class is responsible for managing credentials in the Vault credential store.
 * It provides methods to store, retrieve, update, and delete credentials.
 */
@Service
public class CredentialStoreService {

    private static final Logger LOGGER = LoggerFactory.getLogger(CredentialStoreService.class);

    private static final String BASE_PATH = "/secret/";

    private final VaultTemplate vaultTemplate;

    private final StatusUpdater statusUpdater;

    private final CredentialManager credentialManager;

    private final CredentialRepository repository;


    public CredentialStoreService(VaultTemplate vaultTemplate, StatusUpdater statusUpdater, CredentialManager credentialManager, CredentialRepository repository) {
        this.vaultTemplate = vaultTemplate;
        this.statusUpdater = statusUpdater;
        this.credentialManager = credentialManager;
        this.repository = repository;
    }

    public OperationStatus putCredential(CredentialMetadata request) {

        try {
            LOGGER.debug("Calling putSecret API for owner " + request.getOwnerId() + " with credentials Id " + request.getId() + " Secret " + request.getSecret());

            String path = BASE_PATH + request.getOwnerId() + "/" + request.getType().name();
            Credential credential = new Credential(request.getId(), request.getSecret());
            credential.setSuperTenant(request.getSuperTenant());
            vaultTemplate.write(path, credential);
            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response != null && response.getData() != null && response.getData().getId() != null) {
                statusUpdater.updateStatus(Operations.PUT_CREDENTIAL.name(),
                        org.apache.custos.core.model.commons.OperationStatus.SUCCESS,
                        request.getOwnerId(),
                        null);
                return OperationStatus.newBuilder().setState(true).build();

            } else {
                String msg = "PutSecret operation failed for " + request.getOwnerId() + "with credentials Id " + request.getId() + " Secret " + request.getSecret() + " Type " + request.getType().name();
                LOGGER.error(msg);

                statusUpdater.updateStatus(Operations.PUT_CREDENTIAL.name(),
                        org.apache.custos.core.model.commons.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);
                throw new RuntimeException(msg);
            }

        } catch (Exception ex) {
            String msg = "PutSecret operation failed for " + request.getOwnerId() + "with credentials Id " + request.getId() + " Secret " + request.getSecret() + " Type " + request.getType().name();
            LOGGER.error(msg);

            statusUpdater.updateStatus(Operations.PUT_CREDENTIAL.name(),
                    org.apache.custos.core.model.commons.OperationStatus.FAILED,
                    request.getOwnerId(),
                    null);
            throw new InternalServerException(msg);
        }
    }

    public CredentialMetadata getCredential(GetCredentialRequest request) {
        try {
            LOGGER.debug("Calling getSecret API for owner " + request.getOwnerId() + " for type " + request.getType());
            String path = BASE_PATH + request.getOwnerId() + "/" + request.getType().name();
            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null || response.getData() == null) {
                String msg = "Cannot find credentials for " + request.getOwnerId() + " for type " + request.getType();
                LOGGER.error(msg);
                return CredentialMetadata.newBuilder().build();
            }

            Credential credential = response.getData();
            CredentialMetadata.Builder secret = CredentialMetadata.newBuilder()
                    .setSecret(credential.getSecret())
                    .setId(credential.getId())
                    .setOwnerId(request.getOwnerId());

            if (request.getType() == Type.CUSTOS) {
                Optional<CredentialEntity> entity = repository.findByOwnerIdAndClientId(request.getOwnerId(), credential.getId());
                if (entity.isPresent()) {
                    secret.setClientIdIssuedAt(entity.get().getIssuedAt().getTime());
                    secret.setClientSecretExpiredAt(entity.get().getClientSecretExpiredAt());
                }
            }
            return secret.build();

        } catch (Exception ex) {
            String msg = " operation failed for " + request.getOwnerId() + "with credentials Id " + request.getId() + " Type " + request.getType().name();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetAllCredentialsResponse getAllCredentials(GetAllCredentialsRequest request) {
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
                        if (crRe != null && crRe.getData() != null) {
                            CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), request.getOwnerId(), key);
                            credentialMetadata.add(metadata);
                        }
                    }
                }
            }
            return GetAllCredentialsResponse.newBuilder().addAllSecretList(credentialMetadata).build();

        } catch (Exception ex) {
            String msg = " operation failed for " + request.getOwnerId() + " " + ex;
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public OperationStatus deleteCredential(DeleteCredentialRequest request) {
        try {
            LOGGER.debug("Calling deleteSecret API for owner " + request.getOwnerId());

            String subPath = BASE_PATH + request.getOwnerId() + (request.getType() != null ? "/" + request.getType().name() : "");
            List<String> paths = vaultTemplate.list(subPath);
            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    String path = subPath + "/" + key;
                    vaultTemplate.delete(path);
                }
            }

            statusUpdater.updateStatus(Operations.DELETE_CREDENTIAL.name(),
                    org.apache.custos.core.model.commons.OperationStatus.SUCCESS,
                    request.getOwnerId(),
                    null);

            return OperationStatus.newBuilder().setState(true).build();

        } catch (Exception ex) {
            String msg = " operation failed for " + request.getOwnerId();
            statusUpdater.updateStatus(Operations.DELETE_CREDENTIAL.name(),
                    org.apache.custos.core.model.commons.OperationStatus.FAILED,
                    request.getOwnerId(),
                    null);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetOperationsMetadataResponse getOperationMetadata(GetOperationsMetadataRequest request) {
        try {
            LOGGER.debug("Calling getOperationMetadata API for traceId " + request.getTraceId());

            List<OperationMetadata> metadata = new ArrayList<>();
            List<StatusEntity> entities = statusUpdater.getOperationStatus(request.getTraceId());
            if (entities != null && !entities.isEmpty()) {
                for (StatusEntity statusEntity : entities) {
                    OperationMetadata data = convertFromEntity(statusEntity);
                    metadata.add(data);
                }
            }
            return GetOperationsMetadataResponse.newBuilder().addAllMetadata(metadata).build();

        } catch (Exception ex) {
            String msg = " operation failed for " + request.getTraceId();
            throw new InternalServerException(msg, ex);
        }
    }

    public CredentialMetadata getNewCustosCredential(GetNewCustosCredentialRequest request) {
        try {
            LOGGER.debug("Generating custos credentials for  " + request.getOwnerId());
            Credential credential = credentialManager.generateCredential(request.getOwnerId(), CredentialTypes.CUSTOS, 0);
            String path = BASE_PATH + request.getOwnerId() + "/" + CredentialTypes.CUSTOS.name() + "/" + credential.getId();
            vaultTemplate.write(path, credential);

            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);
            if (response == null || response.getData() == null || response.getData().getId() == null) {
                String msg = "Get new custos operation failed for " + request.getOwnerId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_CUSTOS_CREDENTIAL.name(),
                        org.apache.custos.core.model.commons.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                throw new InternalServerException(msg);
            }

            statusUpdater.updateStatus(Operations.GENERATE_CUSTOS_CREDENTIAL.name(),
                    org.apache.custos.core.model.commons.OperationStatus.SUCCESS,
                    request.getOwnerId(),
                    null);

            Optional<CredentialEntity> entity = repository.findByOwnerIdAndClientId(request.getOwnerId(), credential.getId());

            if (entity.isEmpty()) {
                String msg = "Credential is not persisted" + request.getOwnerId();
                LOGGER.error(msg);
                statusUpdater.updateStatus(Operations.GENERATE_CUSTOS_CREDENTIAL.name(),
                        org.apache.custos.core.model.commons.OperationStatus.FAILED,
                        request.getOwnerId(),
                        null);

                throw new InternalServerException(msg);
            }

            CredentialEntity en = entity.get();
            return CredentialMetadata
                    .newBuilder()
                    .setOwnerId(request.getOwnerId())
                    .setSecret(credential.getSecret())
                    .setClientIdIssuedAt(en.getIssuedAt().getTime())
                    .setClientSecretExpiredAt(en.getClientSecretExpiredAt())
                    .setId(credential.getId())
                    .build();


        } catch (Exception ex) {
            String msg = " Credential generation failed for tenant " + request.getOwnerId();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetOwnerIdResponse getOwnerIdFromToken(TokenRequest request) {
        try {
            LOGGER.debug("Get ownerId for    " + request.getToken());
            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                throw new EntityNotFoundException("Could not find a token");
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());

            if (entity != null) {
                return GetOwnerIdResponse.newBuilder().setOwnerId(entity.getOwnerId()).build();
            } else {
                LOGGER.error("Could not find a token");
                throw new EntityNotFoundException("Could not find a token");
            }

        } catch (Exception ex) {
            String msg = "operation failed";
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public CredentialMetadata getCustosCredentialFromToken(TokenRequest request) {
        try {
            LOGGER.debug("Get credential for " + request.getToken());
            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                throw new EntityNotFoundException("Could not find a token");
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());

            if (entity == null) {
                LOGGER.error("Could not find a token");
                throw new EntityNotFoundException("Could not find a token");
            }

            String path = BASE_PATH + entity.getOwnerId() + "/" + Type.CUSTOS.name();
            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null || response.getData() == null || !response.getData().getSecret().equals(credential.getSecret())) {
                String msg = "Invalid secret for Id: " + credential.getId();
                LOGGER.error(msg);
                throw new AuthenticationException(msg);
            }


            return CredentialMetadata.newBuilder()
                    .setSecret(credential.getSecret())
                    .setId(credential.getId())
                    .setOwnerId(entity.getOwnerId())
                    .setClientSecretExpiredAt(entity.getClientSecretExpiredAt())
                    .setClientIdIssuedAt(entity.getIssuedAt().getTime())
                    .setType(Type.CUSTOS).build();

        } catch (Exception ex) {
            String msg = "Error while extracting token from the CUSTOS credentials ";
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public CredentialMetadata getCustosCredentialFromClientId(GetCredentialRequest request) {
        try {
            String clientId = request.getId();
            CredentialEntity entity = repository.findByClientId(clientId);

            if (entity == null) {
                String msg = " Credentials not found for clientId " + clientId;
                LOGGER.error(msg);
                throw new EntityNotFoundException(msg);
            }

            String path = BASE_PATH + entity.getOwnerId() + "/" + Type.CUSTOS.name() +  "/" + clientId;

            VaultResponseSupport<Credential> response = vaultTemplate.read(path, Credential.class);

            if (response == null || response.getData() == null) {

                path = BASE_PATH + entity.getOwnerId() + "/" + Type.CUSTOS.name();
                response = vaultTemplate.read(path, Credential.class);
                if (response == null || response.getData() == null) {
                    String msg = "Cannot find credentials for " + entity.getOwnerId() + " for type " + Type.CUSTOS.name();
                    LOGGER.error(msg);
                    throw new EntityNotFoundException(msg);
                }
            }

            return CredentialMetadata.newBuilder()
                    .setSecret(response.getData().getSecret())
                    .setId(request.getId())
                    .setOwnerId(entity.getOwnerId())
                    .setClientSecretExpiredAt(entity.getClientSecretExpiredAt())
                    .setClientIdIssuedAt(entity.getIssuedAt().getTime())
                    .setSuperTenant(response.getData().isSuperTenant())
                    .setType(Type.CUSTOS).build();

        } catch (Exception ex) {
            String msg = "Error while extracting the credentials for Owner Id: " + request.getOwnerId();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetAllCredentialsResponse getAllCredentialsFromToken(TokenRequest request) {
        try {
            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                throw new EntityNotFoundException("Invalid access token");
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());

            if (entity == null) {
                LOGGER.error("Client not found");
                throw new EntityNotFoundException("Client not found");
            }

            String subPath = BASE_PATH + entity.getOwnerId();

            String validatingPath = BASE_PATH + entity.getOwnerId() + "/" + Type.CUSTOS.name();
            VaultResponseSupport<Credential> validationResponse = vaultTemplate.read(validatingPath, Credential.class);

            if (validationResponse == null || validationResponse.getData() == null || !validationResponse.getData().getSecret().equals(credential.getSecret())) {
                String msg = "Invalid secret for Id: " + credential.getId();
                LOGGER.error(msg);
                throw new AuthenticationException(msg);
            }

            List<String> paths = vaultTemplate.list(subPath);

            List<CredentialMetadata> credentialMetadata = new ArrayList<>();


            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    if (isMainType(key)) {
                        String path = subPath + "/" + key;
                        VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);
                        if (crRe != null && crRe.getData() != null) {
                            CredentialMetadata metadata = convertToCredentialMetadata(crRe.getData(), entity.getOwnerId(), key);
                            if (key.equals(Type.CUSTOS.name())) {
                                metadata = metadata.toBuilder()
                                        .setClientIdIssuedAt(entity.getIssuedAt().getTime())
                                        .setClientSecretExpiredAt(entity.getClientSecretExpiredAt())
                                        .build();
                            }
                            credentialMetadata.add(metadata);
                        }
                    }
                }
            }
            return GetAllCredentialsResponse.newBuilder().addAllSecretList(credentialMetadata).build();

        } catch (Exception ex) {
            String msg = "Operation failed " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public Credentials getBasicCredentials(TokenRequest request) {
        try {
            String token = request.getToken();
            Credential credential = credentialManager.decodeToken(token);

            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                throw new EntityNotFoundException("Invalid access token");
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());
            if (entity == null) {
                LOGGER.error("Could not find the credential entity with the Id: {}", credential.getId());
                throw new EntityNotFoundException("Could not find the credential entity with the Id: " + credential.getId());
            }

            String subPath = BASE_PATH + entity.getOwnerId();
            List<String> paths = vaultTemplate.list(subPath);
            Credentials.Builder credentialsBuilder = Credentials.newBuilder();

            if (paths != null && !paths.isEmpty()) {
                for (String key : paths) {
                    String path = subPath + "/" + key;
                    VaultResponseSupport<Credential> crRe = vaultTemplate.read(path, Credential.class);
                    if (crRe == null || crRe.getData() == null || crRe.getData().getSecret() == null) {
                        LOGGER.error("Cannot find Credential with the Id: " + credential.getId() + " in the Secret store");
                        throw new EntityNotFoundException("Cannot find Credential with the Id: " + credential.getId() + " in the Secret store");
                    }

                    if (key.equals(Type.CUSTOS.name())) {
                        if (!crRe.getData().getSecret().equals(credential.getSecret())) {
                            String msg = "Invalid secret for id" + credential.getId();
                            LOGGER.error(msg);
                            throw new AuthenticationException(msg);
                        }

                        credentialsBuilder.setCustosClientId(crRe.getData().getId())
                                .setCustosClientSecret(crRe.getData().getSecret())
                                .setCustosClientIdIssuedAt(entity.getIssuedAt().getTime())
                                .setCustosClientSecretExpiredAt(entity.getClientSecretExpiredAt());

                    } else if (key.equals(Type.IAM.name())) {
                        credentialsBuilder.setIamClientId(crRe.getData().getId()).setIamClientSecret(crRe.getData().getSecret());

                    } else if (key.equals(Type.CILOGON.name())) {
                        credentialsBuilder.setCiLogonClientId(crRe.getData().getId()).setCiLogonClientSecret(crRe.getData().getSecret());
                    }
                }
            }
            return credentialsBuilder.build();

        } catch (Exception ex) {
            String msg = "Operation failed " + ex;
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetAllCredentialsResponse getAllCredentialsFromJWTToken(TokenRequest request) {
        try {
            String token = request.getToken();

            Credential credential = credentialManager.decodeJWTToken(token);
            if (credential == null || credential.getId() == null) {
                LOGGER.error("Invalid access token");
                throw new EntityNotFoundException("Invalid access token");
            }

            CredentialEntity entity = repository.findByClientId(credential.getId());

            if (entity == null) {
                LOGGER.error("Cannot find a CredentialEntity with the Id: " + credential.getId());
                throw new EntityNotFoundException("Cannot find a CredentialEntity with the Id: " + credential.getId());
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

            return GetAllCredentialsResponse.newBuilder()
                    .addAllSecretList(credentialMetadata)
                    .setRequesterUserEmail(credential.getEmail())
                    .setRequesterUsername(credential.getUsername())
                    .build();

        } catch (Exception ex) {
            String msg = "Operation failed  " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    public GetAllCredentialsResponse getMasterCredentials(GetCredentialRequest request) {
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
            return GetAllCredentialsResponse.newBuilder().addAllSecretList(credentialMetadata).build();

        } catch (Exception ex) {
            String msg = "Error while extracting the master credentials";
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
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
}
