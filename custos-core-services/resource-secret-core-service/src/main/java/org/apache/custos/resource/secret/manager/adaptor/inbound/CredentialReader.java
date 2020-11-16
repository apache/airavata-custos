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

package org.apache.custos.resource.secret.manager.adaptor.inbound;

import org.apache.custos.resource.secret.manager.adaptor.outbound.CredentialWriter;
import org.apache.custos.resource.secret.persistance.local.model.Secret;
import org.apache.custos.resource.secret.persistance.local.repository.SecretRepository;
import org.apache.custos.resource.secret.persistance.vault.Certificate;
import org.apache.custos.resource.secret.persistance.vault.KVSecret;
import org.apache.custos.resource.secret.persistance.vault.PasswordSecret;
import org.apache.custos.resource.secret.persistance.vault.SSHCredentialSecrets;
import org.apache.custos.resource.secret.service.*;
import org.apache.custos.resource.secret.utils.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.vault.core.VaultTemplate;
import org.springframework.vault.support.VaultResponseSupport;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@Component
public class CredentialReader {


    private static final Logger LOGGER = LoggerFactory.getLogger(CredentialWriter.class);

    @Autowired
    private SecretRepository repository;

    @Autowired
    private VaultTemplate vaultTemplate;


    /**
     * get SSH credentials
     *
     * @param tenantId
     * @param token
     * @return
     */
    public SSHCredential getSSHCredential(long tenantId, String token) {

        Secret secret = null;

        if (token != null && !token.trim().equals("")) {
            Optional<Secret> exSecret = repository.findById(token);
            if (exSecret.isPresent()){
                secret = exSecret.get();
            }
        }
        if (secret == null) {
            List<Secret> secrets = repository.findAllByExternalIdAndTenantId(token, tenantId);
            if (secrets != null && !secrets.isEmpty()) {
                secret = secrets.get(0);
            }
        }

        if (secret == null) {
            return null;
        }

        String vaultPath = Constants.VAULT_RESOURCE_SECRETS_PATH + tenantId + "/" + secret.getOwnerId() +
                "/" + Constants.SSH_CREDENTIALS + "/" + secret.getId();

        VaultResponseSupport<SSHCredentialSecrets> response = vaultTemplate.read(vaultPath, SSHCredentialSecrets.class);

        if (response == null || response.getData() == null && response.getData().getPrivateKey() == null) {
            repository.delete(secret);
            return null;
        }

        SSHCredentialSecrets sshCredentialSecrets = response.getData();

        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setOwnerId(secret.getOwnerId())
                .setTenantId(tenantId)
                .setPersistedTime(secret.getCreatedAt().getTime())
                .setDescription(secret.getDiscription())
                .setResourceType(ResourceType.VAULT_CREDENTIAL)
                .setSource(ResourceSource.EXTERNAL)
                .setToken(
                        (secret.getExternalId() != null &&
                                !secret.getExternalId().trim().equals(""))? secret.getExternalId(): secret.getId())
                .build();

        SSHCredential credential = SSHCredential.newBuilder()
                .setPassphrase(sshCredentialSecrets.getPassphrase())
                .setPrivateKey(sshCredentialSecrets.getPrivateKey())
                .setPublicKey(sshCredentialSecrets.getPublicKey())
                .setMetadata(metadata)
                .build();

        return credential;

    }


    /**
     * get password credential
     *
     * @param tenantId
     * @param token
     * @return
     */
    public org.apache.custos.resource.secret.service.PasswordCredential getPasswordCredential(long tenantId,
                                                                                              String token) {
        Secret secret = null;

        if (token != null && !token.trim().equals("")) {
            Optional<Secret> exSecret = repository.findById(token);
            if (exSecret.isPresent()){
                secret = exSecret.get();
            }
        } if (secret == null ) {
            List<Secret> secrets = repository.findAllByExternalIdAndTenantId(token, tenantId);
            if (secrets != null && !secrets.isEmpty()) {
                secret = secrets.get(0);
            }
        }

        if (secret == null) {
            return null;
        }

        String vaultPath = Constants.VAULT_RESOURCE_SECRETS_PATH + tenantId + "/" + secret.getOwnerId() +
                "/" + Constants.PASSWORD + "/" + secret.getId();


        VaultResponseSupport<PasswordSecret> response = vaultTemplate.read(vaultPath, PasswordSecret.class);

        if (response == null || response.getData() == null && response.getData().getPassword() == null) {
            repository.delete(secret);
            return null;
        }

        PasswordSecret passwordSecret = response.getData();

        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setOwnerId(secret.getOwnerId())
                .setTenantId(tenantId)
                .setPersistedTime(secret.getCreatedAt().getTime())
                .setDescription(secret.getDiscription())
                .setResourceType(ResourceType.VAULT_CREDENTIAL)
                .setSource(ResourceSource.EXTERNAL)
                .setType(ResourceSecretType.PASSWORD)
                .setToken(
                        (secret.getExternalId() != null ||
                                !secret.getExternalId().trim().equals(""))? secret.getExternalId(): secret.getId())
                .build();

        org.apache.custos.resource.secret.service.PasswordCredential credential =
                org.apache.custos.resource.secret.service.PasswordCredential.newBuilder()
                        .setPassword(passwordSecret.getPassword())
                        .setMetadata(metadata)
                        .build();

        return credential;


    }

    /**
     * get certificate credential
     *
     * @param tenantId
     * @param token
     * @return
     */
    public CertificateCredential getCertificateCredential(long tenantId, String token) {
        Secret secret = null;

        if (token != null && !token.trim().equals("")) {
            Optional<Secret> exSecret = repository.findById(token);
            if (exSecret.isPresent()){
                secret = exSecret.get();
            }
        } if (secret == null) {
            List<Secret> secrets = repository.findAllByExternalIdAndTenantId(token, tenantId);
            if (secrets != null && !secrets.isEmpty()) {
                secret = secrets.get(0);
            }
        }

        if (secret == null) {
            return null;
        }


        String vaultPath = Constants.VAULT_RESOURCE_SECRETS_PATH + tenantId + "/" + secret.getOwnerId() +
                "/" + Constants.PASSWORD + "/" + secret.getId();

        VaultResponseSupport<Certificate> response = vaultTemplate.read(vaultPath, Certificate.class);

        if (response == null || response.getData() == null && response.getData().getCertificate() == null) {
            repository.delete(secret);
            return null;
        }

        Certificate certificate = response.getData();

        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setOwnerId(secret.getOwnerId())
                .setTenantId(tenantId)
                .setPersistedTime(secret.getCreatedAt().getTime())
                .setDescription(secret.getDiscription())
                .setResourceType(ResourceType.VAULT_CREDENTIAL)
                .setSource(ResourceSource.EXTERNAL)
                .setType(ResourceSecretType.X509_CERTIFICATE)
                .setToken(
                        (secret.getExternalId() != null &&
                                !secret.getExternalId().trim().equals(""))? secret.getExternalId(): secret.getId())
                .build();

        CertificateCredential certificateCredential = CertificateCredential.newBuilder()
                .setLifeTime(Long.valueOf(certificate.getLifetime()))
                .setNotAfter(certificate.getNotAfter())
                .setNotBefore(certificate.getNotBefore())
                .setPrivateKey(certificate.getPrivateKey())
                .setX509Cert(certificate.getCertificate())
                .setMetadata(metadata)
                .build();

        return certificateCredential;

    }

    /**
     * Get credential summary
     *
     * @param tenantId
     * @param token
     * @return
     */
    public SecretMetadata getCredentialSummary(long tenantId, String token) {

        Secret secret = null;

        if (token != null && !token.trim().equals("")) {
            Optional<Secret> exSecret = repository.findById(token);
            if (exSecret.isPresent()){
                secret = exSecret.get();
            }
        }
        if (secret == null) {
            List<Secret> secrets = repository.findAllByExternalIdAndTenantId(token, tenantId);
            if (secrets != null && !secrets.isEmpty()) {
                secret = secrets.get(0);
            }
        }

        if (secret == null) {
            return null;
        }

        return SecretMetadata.newBuilder()
                .setToken(
                        (secret.getExternalId() != null &&
                                !secret.getExternalId().trim().equals(""))? secret.getExternalId(): secret.getId())
                .setTenantId(tenantId)
                .setDescription(secret.getDiscription())
                .setPersistedTime(secret.getCreatedAt().getTime())
                .setType(ResourceSecretType.valueOf(secret.getSecretType()))
                .setResourceType(ResourceType.VAULT_CREDENTIAL)
                .setSource(ResourceSource.EXTERNAL)
                .setOwnerId(secret.getOwnerId())
                .build();

    }

    /**
     * get All credential summaries
     *
     * @param tenantId
     * @param tokens
     * @return
     */
    public List<SecretMetadata> getAllCredentialSummaries(long tenantId, List<String> tokens) {

        List<Secret> secrets = repository.getAllSecretsByIdOrExternalId(tenantId, tokens, tokens);
        List<SecretMetadata> metadata = new ArrayList<>();

        if (secrets != null && !secrets.isEmpty()) {


            secrets.forEach(secret -> {
                metadata.add(SecretMetadata.newBuilder()
                        .setToken(
                                (secret.getExternalId() != null &&
                                        !secret.getExternalId().trim().equals(""))? secret.getExternalId(): secret.getId())
                        .setTenantId(tenantId)
                        .setDescription(secret.getDiscription())
                        .setPersistedTime(secret.getCreatedAt().getTime())
                        .setType(ResourceSecretType.valueOf(secret.getSecretType()))
                        .setResourceType(ResourceType.VAULT_CREDENTIAL)
                        .setSource(ResourceSource.EXTERNAL)
                        .setOwnerId(secret.getOwnerId())
                        .build());
            });


        }

        return metadata;

    }


    public KVCredential getKVSecretByToken(String token, long tenantId, String ownerId) {
        Optional<Secret> secret = repository.findById(token);

        if (secret.isEmpty()) {
            return null;
        }

        Secret exSec = secret.get();

        if (!exSec.getOwnerId().equals(ownerId)) {
            return null;
        }

        String vaultPath = Constants.VAULT_RESOURCE_SECRETS_PATH + tenantId + "/" + ownerId +
                "/" + Constants.KV_SECRET + "/" + token;

        VaultResponseSupport<KVSecret> response = vaultTemplate.read(vaultPath, KVSecret.class);

        KVSecret kvSecret = response.getData();
        if (kvSecret == null || kvSecret.getValue() == null) {
            repository.delete(exSec);
            return null;
        }
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setToken(token)
                .setTenantId(tenantId)
                .setDescription(exSec.getDiscription())
                .setPersistedTime(exSec.getCreatedAt().getTime())
                .setType(ResourceSecretType.valueOf(exSec.getSecretType()))
                .setResourceType(ResourceType.OTHER)
                .setSource(ResourceSource.EXTERNAL)
                .setOwnerId(exSec.getOwnerId())
                .build();

        KVCredential kvCredential = KVCredential.newBuilder()
                .setKey(exSec.getExternalId())
                .setToken(exSec.getId())
                .setValue(kvSecret.getValue())
                .setMetadata(metadata).build();

        return kvCredential;
    }

    public KVCredential getKVSecretByKey(String key, long tenantId, String ownerId) {

        List<Secret> secrets = repository.findAllByExternalIdAndOwnerIdAndTenantId(key, ownerId, tenantId);

        if (secrets != null && secrets.isEmpty()) {
            return null;
        }
        Secret exSec = secrets.get(0);

        String vaultPath = Constants.VAULT_RESOURCE_SECRETS_PATH + tenantId + "/" + ownerId +
                "/" + Constants.KV_SECRET + "/" + exSec.getId();

        VaultResponseSupport<KVSecret> response = vaultTemplate.read(vaultPath, KVSecret.class);

        KVSecret kvSecret = response.getData();
        if (kvSecret == null || kvSecret.getValue() == null) {
            repository.delete(exSec);
            return null;
        }
        SecretMetadata metadata = SecretMetadata.newBuilder()
                .setToken(exSec.getId())
                .setTenantId(tenantId)
                .setDescription(exSec.getDiscription())
                .setPersistedTime(exSec.getCreatedAt().getTime())
                .setType(ResourceSecretType.valueOf(exSec.getSecretType()))
                .setResourceType(ResourceType.OTHER)
                .setSource(ResourceSource.EXTERNAL)
                .setOwnerId(exSec.getOwnerId())
                .build();

        KVCredential kvCredential = KVCredential.newBuilder()
                .setKey(exSec.getExternalId())
                .setToken(exSec.getId())
                .setValue(kvSecret.getValue())
                .setMetadata(metadata).build();

        return kvCredential;
    }
}



