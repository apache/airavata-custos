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

package org.apache.custos.resource.secret.manager.adaptor.outbound;

import org.apache.custos.resource.secret.exceptions.CredentialStoreException;
import org.apache.custos.resource.secret.persistance.local.model.Secret;
import org.apache.custos.resource.secret.persistance.local.repository.SecretRepository;
import org.apache.custos.resource.secret.persistance.vault.Certificate;
import org.apache.custos.resource.secret.persistance.vault.KVSecret;
import org.apache.custos.resource.secret.persistance.vault.PasswordSecret;
import org.apache.custos.resource.secret.persistance.vault.SSHCredentialSecrets;
import org.apache.custos.resource.secret.service.ResourceSecretType;
import org.apache.custos.resource.secret.utils.Constants;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.vault.core.VaultTemplate;
import org.springframework.vault.support.VaultResponseSupport;

import java.util.List;
import java.util.Optional;

/**
 * This class received gPRC credentials
 */
@Component
public class CredentialWriter {

    private static final Logger LOGGER = LoggerFactory.getLogger(CredentialWriter.class);

    @Autowired
    private SecretRepository repository;

    @Autowired
    private VaultTemplate vaultTemplate;

    /**
     * Save SSHCredential in database
     *
     * @param credential
     * @return
     */
    public boolean saveSSHCredential(SSHCredential credential) {

        Optional<Secret> exSecret = repository.findById(credential.getToken());

        if (exSecret.isPresent()) {
            String msg = " Credential with token " + credential.getToken() + " already exist";
            LOGGER.error(msg);
            throw new CredentialStoreException("Invalid token", null);
        }

        if (credential.getExternalId() != null && !credential.getExternalId().trim().equals("")) {

            Optional<Secret> exToSec = repository.findById(credential.getExternalId());

            if (exToSec.isPresent()) {
                String msg = " Credential with token " + credential.getToken() + " already exist";
                LOGGER.error(msg);
                throw new CredentialStoreException("Invalid token", null);
            }

            List<Secret> secrets = repository.findAllByExternalIdAndTenantId(credential.getExternalId(),
                    credential.getTenantId());
            if (secrets != null && !secrets.isEmpty()) {
                String msg = " Credential with externalId " + credential.getExternalId() + " already exist";
                LOGGER.error(msg);
                throw new CredentialStoreException("Invalid token", null);
            }

        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credential.getTenantId() + "/" + credential.getOwnerId()
                + "/" + Constants.SSH_CREDENTIALS + "/" + credential.getToken();

        SSHCredentialSecrets sshCredentialSecrets = new SSHCredentialSecrets
                (credential.getPrivateKey(), credential.getPublicKey(), credential.getPassPhrase());
        vaultTemplate.write(path, sshCredentialSecrets);

        VaultResponseSupport<SSHCredentialSecrets> response = vaultTemplate.read(path, SSHCredentialSecrets.class);

        if (response == null || response.getData() == null && response.getData().getPrivateKey() == null) {
            String msg = " SSH credential of tenant " + credential.getTenantId() +
                    " of user " + credential.getOwnerId() + " is not saved in vault";
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }

        Secret secret = new Secret();
        secret.setId(credential.getToken());
        secret.setDiscription(credential.getDescription());
        secret.setOwnerId(credential.getOwnerId());
        secret.setOwnerType(credential.getResourceOwnerType().name());
        secret.setSecretType(ResourceSecretType.SSH.name());
        secret.setTenantId(credential.getTenantId());
        secret.setExternalId(credential.getExternalId());
        secret.setType(credential.getType());
        repository.save(secret);
        return true;
    }

    /**
     * save Password credential in Vault and DB
     *
     * @param credential
     * @return
     */
    public boolean savePasswordCredential(PasswordCredential credential) {
        Optional<Secret> exSecret = repository.findById(credential.getToken());

        if (exSecret.isPresent()) {
            String msg = " Credential with token " + credential.getToken() + " already exist";
            LOGGER.error(msg);
            throw new CredentialStoreException("Invalid token", null);
        }

        if (credential.getExternalId() != null && !credential.getExternalId().trim().equals("")) {
            Optional<Secret> exToSec = repository.findById(credential.getExternalId());

            if (exToSec.isPresent()) {
                String msg = " Credential with token " + credential.getToken() + " already exist";
                LOGGER.error(msg);
                throw new CredentialStoreException("Invalid token", null);
            }

            List<Secret> secrets = repository.findAllByExternalIdAndTenantId(credential.getExternalId(),
                    credential.getTenantId());
            if (secrets != null && !secrets.isEmpty()) {
                String msg = " Credential with externalId " + credential.getExternalId() + " already exist";
                LOGGER.error(msg);
                throw new CredentialStoreException("Invalid token", null);
            }

        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credential.getTenantId() + "/" + credential.getOwnerId()
                + "/" + Constants.PASSWORD + "/" + credential.getToken();


        PasswordSecret passwordSecret = new PasswordSecret(credential.getPassword(), credential.getUserId());
        vaultTemplate.write(path, passwordSecret);

        VaultResponseSupport<PasswordSecret> response = vaultTemplate.read(path, PasswordSecret.class);

        if (response == null || response.getData() == null && response.getData().getPassword() == null) {
            String msg = " Password credential of tenant " + credential.getTenantId() +
                    " of user " + credential.getOwnerId() + " is not saved in vault";
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }

        Secret secret = new Secret();
        secret.setId(credential.getToken());
        secret.setDiscription(credential.getDescription());
        secret.setOwnerId(credential.getOwnerId());
        secret.setOwnerType(credential.getResourceOwnerType().name());
        secret.setSecretType(ResourceSecretType.PASSWORD.name());
        secret.setTenantId(credential.getTenantId());
        secret.setExternalId(credential.getExternalId());
        secret.setType(credential.getType());
        repository.save(secret);
        return true;
    }

    /**
     * save certificate credential in Vault and DB
     *
     * @param credential
     * @return
     */
    public boolean saveCertificateCredential(CertificateCredential credential) {
        Optional<Secret> exSecret = repository.findById(credential.getToken());

        if (exSecret.isPresent()) {
            String msg = " Credential with token " + credential.getToken() + " already exist";
            LOGGER.error(msg);
            throw new CredentialStoreException("Invalid token", null);
        }

        if (credential.getExternalId() != null && !credential.getExternalId().trim().equals("")) {
            Optional<Secret> exToSec = repository.findById(credential.getExternalId());

            if (exToSec.isPresent()) {
                String msg = " Credential with token " + credential.getToken() + " already exist";
                LOGGER.error(msg);
                throw new CredentialStoreException("Invalid token", null);
            }

            List<Secret> secrets = repository.findAllByExternalIdAndTenantId(credential.getExternalId(),
                    credential.getTenantId());
            if (secrets != null && !secrets.isEmpty()) {
                String msg = " Credential with externalId " + credential.getExternalId() + " already exist";
                LOGGER.error(msg);
                throw new CredentialStoreException("Invalid token", null);
            }

        }


        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credential.getTenantId() + "/" + credential.getOwnerId() +
                "/" + Constants.CERTIFICATES + "/" + credential.getToken();


        Certificate certificate = new Certificate(credential.getCert(),
                String.valueOf(credential.getLifetime()),
                credential.getNotBefore(),
                credential.getNotAfter(),
                credential.getPrivateKey());

        vaultTemplate.write(path, certificate);

        VaultResponseSupport<Certificate> response = vaultTemplate.read(path, Certificate.class);

        if (response == null || response.getData() == null && response.getData().getCertificate() == null) {
            String msg = " Certificate credential of tenant " + credential.getTenantId() +
                    " of user " + credential.getOwnerId() + " is not saved in vault";
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }

        Secret secret = new Secret();
        secret.setId(credential.getToken());
        secret.setDiscription(credential.getDescription());
        secret.setOwnerId(credential.getOwnerId());
        secret.setOwnerType(credential.getResourceOwnerType().name());
        secret.setSecretType(ResourceSecretType.X509_CERTIFICATE.name());
        secret.setTenantId(credential.getTenantId());
        secret.setExternalId(credential.getExternalId());
        secret.setType(credential.getType());
        repository.save(secret);
        return true;
    }


    /**
     * delete existing credential
     *
     * @param tenantId
     * @param token
     * @return
     */
    public boolean deleteCredential(long tenantId, String token) {

        Secret secret = null;
        Optional<Secret> exSec = repository.findById(token);

        if (exSec.isPresent()) {
            secret = exSec.get();
        }

        if (exSec.isEmpty()) {
            List<Secret> secrets = repository.findAllByExternalIdAndTenantId(token, tenantId);
            if (secrets != null && !secrets.isEmpty()) {
                secret = secrets.get(0);
            }
        }

        String type = null;

        if (secret.getSecretType().equals(ResourceSecretType.SSH.name())) {
            type = Constants.SSH_CREDENTIALS;
        } else if (secret.getSecretType().equals(ResourceSecretType.PASSWORD.name())) {
            type = Constants.PASSWORD;
        } else if (secret.getSecretType().equals(ResourceSecretType.X509_CERTIFICATE.name())) {
            type = Constants.CERTIFICATES;
        }


        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + tenantId + "/" + secret.getOwnerId() +
                "/" + type + "/" + secret.getId();

        vaultTemplate.delete(path);

        repository.delete(secret);
        return true;

    }

    public boolean saveKVCredential(KVCredential kvCredential) {
        Optional<Secret> exSecret = repository.findById(kvCredential.getToken());

        if (exSecret.isPresent()) {
            String msg = " Credential with token " + kvCredential.getToken() + " already exist";
            LOGGER.error(msg);
            throw new CredentialStoreException("Invalid token", null);
        }

        List<Secret> secrets = repository.findAllByExternalIdAndOwnerIdAndTenantId(kvCredential.getKey(), kvCredential.getOwnerId(), kvCredential.getTenantId());

        if (secrets != null && !secrets.isEmpty()) {
            String msg = " Credential with key " + kvCredential.getKey() + " of user " + kvCredential.getOwnerId()
                    + " in tenant " + kvCredential.getTenantId() + " is already exists, " +
                    "please update or delete before setting it";
            LOGGER.warn(msg);
            throw new CredentialStoreException(msg, null);
        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + kvCredential.getTenantId() + "/" + kvCredential.getOwnerId() +
                "/" + Constants.KV_SECRET + "/" + kvCredential.getToken();

        KVSecret kvSecret = new KVSecret(kvCredential.getKey(), kvCredential.getValue());

        vaultTemplate.write(path, kvSecret);

        VaultResponseSupport<KVSecret> response = vaultTemplate.read(path, KVSecret.class);

        if (response == null || response.getData() == null && response.getData().getKey() == null) {
            String msg = " KV credential of tenant " + kvCredential.getTenantId() +
                    " of user " + kvCredential.getOwnerId() + " is not saved in vault";
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }

        Secret secret = new Secret();
        secret.setId(kvCredential.getToken());
        secret.setDiscription(kvCredential.getDescription());
        secret.setOwnerId(kvCredential.getOwnerId());
        secret.setOwnerType(kvCredential.getResourceOwnerType().name());
        secret.setSecretType(ResourceSecretType.KV.name());
        secret.setTenantId(kvCredential.getTenantId());
        secret.setExternalId(kvCredential.getKey());
        secret.setType(kvCredential.getType());
        repository.save(secret);
        return true;

    }

    public boolean updateKVCredential(org.apache.custos.resource.secret.service.KVCredential kvCredential) {
        Secret secret = null;

        if (kvCredential.getToken() != null && !kvCredential.getToken().equals("")) {
            Optional<Secret> exSecret = repository.findById(kvCredential.getToken());
            if (!exSecret.isPresent() || (!exSecret.get().getOwnerId().equals(kvCredential.getMetadata().getOwnerId()) ||
                    !exSecret.get().getExternalId().equals(kvCredential.getKey()))) {
                String msg = " Cannot find record for token" + kvCredential.getToken()
                        + " with given key " + kvCredential.getKey();
                LOGGER.error(msg);
                throw new CredentialStoreException(msg, null);
            }
            secret = exSecret.get();
        } else {

            List<Secret> secrets = repository.
                    findAllByExternalIdAndOwnerIdAndTenantId(kvCredential.getKey(), kvCredential.getMetadata().getOwnerId(),
                            kvCredential.getMetadata().getTenantId());

            if (secrets == null && secrets.isEmpty()) {
                String msg = " Cannot find record "
                        + " with given key " + kvCredential.getKey();
                LOGGER.error(msg);
                throw new CredentialStoreException(msg, null);
            }
            secret = secrets.get(0);
        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + kvCredential.getMetadata().getTenantId() + "/" +
                kvCredential.getMetadata().getOwnerId() +
                "/" + Constants.KV_SECRET + "/" + secret.getId();

        KVSecret kvSecret = new KVSecret(kvCredential.getKey(), kvCredential.getValue());

        vaultTemplate.delete(path);
        vaultTemplate.write(path, kvSecret);

        VaultResponseSupport<KVSecret> response = vaultTemplate.read(path, KVSecret.class);

        if (response == null || response.getData() == null && response.getData().getKey() == null) {
            String msg = " KV credential of tenant " + kvCredential.getMetadata().getTenantId() +
                    " of user " + kvCredential.getMetadata().getOwnerId() + " is not saved in vault";
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }

        secret.setDiscription(kvCredential.getMetadata().getDescription());
        repository.save(secret);
        return true;

    }

    public boolean deleteKVCredential(org.apache.custos.resource.secret.service.KVCredential kvCredential) {
        Secret secret = null;
        if (kvCredential.getToken() != null && !kvCredential.getToken().equals("")) {
            Optional<Secret> exSecret = repository.findById(kvCredential.getToken());
            if (!exSecret.isPresent() || (!exSecret.get().getOwnerId().equals(kvCredential.getMetadata().getOwnerId()) ||
                    !exSecret.get().getExternalId().equals(kvCredential.getKey()))) {
                String msg = " Cannot find record for token" + kvCredential.getToken()
                        + " with given key " + kvCredential.getKey();
                LOGGER.error(msg);
                throw new CredentialStoreException(msg, null);
            }
            secret = exSecret.get();
        } else {

            List<Secret> secrets = repository.
                    findAllByExternalIdAndOwnerIdAndTenantId(kvCredential.getKey(), kvCredential.getMetadata().getOwnerId(),
                            kvCredential.getMetadata().getTenantId());

            if (secrets == null && secrets.isEmpty()) {
                String msg = " Cannot find record "
                        + " with given key " + kvCredential.getKey();
                LOGGER.error(msg);
                throw new CredentialStoreException(msg, null);
            }
            secret = secrets.get(0);
        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + kvCredential.getMetadata().getTenantId() + "/"
                + kvCredential.getMetadata().getOwnerId() +
                "/" + Constants.KV_SECRET + "/" + secret.getId();

        vaultTemplate.delete(path);

        repository.delete(secret);
        return true;
    }

    public boolean saveCredentialMap(CredentialMap credentialMap) {
        Optional<Secret> exSecret = repository.findById(credentialMap.getToken());

        if (exSecret.isPresent()) {
            String msg = " Credential with token " + credentialMap.getToken() + " already exist";
            LOGGER.error(msg);
            throw new CredentialStoreException("Invalid token", null);
        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credentialMap.getTenantId() + "/" + credentialMap.getOwnerId() +
                "/" + Constants.SECRET_MAP + "/" + credentialMap.getToken();

        KVSecret kvSecret = new KVSecret(credentialMap.getToken(), credentialMap.getCredentialString());

        vaultTemplate.write(path, kvSecret);

        VaultResponseSupport<KVSecret> response = vaultTemplate.read(path, KVSecret.class);

        if (response == null || response.getData() == null && response.getData().getKey() == null) {
            String msg = "  credential Map of tenant " + credentialMap.getTenantId() +
                    " of user " + credentialMap.getOwnerId() + " is not saved in vault";
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }

        Secret secret = new Secret();
        secret.setId(credentialMap.getToken());
        secret.setDiscription(credentialMap.getDescription());
        secret.setOwnerId(credentialMap.getOwnerId());
        secret.setOwnerType(credentialMap.getResourceOwnerType().name());
        secret.setSecretType(ResourceSecretType.CREDENTIAL_MAP.name());
        secret.setTenantId(credentialMap.getTenantId());
        secret.setExternalId(credentialMap.getExternalId());
        secret.setType(credentialMap.getType());
        repository.save(secret);
        return true;

    }

    public boolean updateCredentialMap(CredentialMap credentialMap) {

        Optional<Secret> exSecret = repository.findById(credentialMap.getExternalId());
        if (!exSecret.isPresent()) {
            String msg = " Cannot find secret for token" + credentialMap.getToken();
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }
        Secret secret = exSecret.get();

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credentialMap.getTenantId() + "/" +
                secret.getOwnerId() +
                "/" + Constants.SECRET_MAP + "/" + secret.getId();

        KVSecret kvSecret = new KVSecret(credentialMap.getExternalId(), credentialMap.getCredentialString());

        VaultResponseSupport<KVSecret> responseEx = vaultTemplate.read(path, KVSecret.class);

        vaultTemplate.delete(path);
        vaultTemplate.write(path, kvSecret);

        VaultResponseSupport<KVSecret> response = vaultTemplate.read(path, KVSecret.class);

        if (response == null || response.getData() == null && response.getData().getKey() == null) {
            // Writing back previouse data
            vaultTemplate.write(path, responseEx.getData());

            String msg = " CredentialMap  of tenant " + credentialMap.getTenantId() +
                    " of user " + credentialMap.getOwnerId() + " is not saved in vault";
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }
        secret.setDiscription(credentialMap.getDescription());
        repository.save(secret);
        return true;

    }

    public boolean deleteCredentialMap(CredentialMap credentialMap) {
        Optional<Secret> exSecret = repository.findById(credentialMap.getExternalId());
        if (!exSecret.isPresent()) {
            String msg = " Cannot find secret for token" + credentialMap.getExternalId();
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }
        Secret secret = exSecret.get();

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credentialMap.getTenantId() + "/" +
                secret.getOwnerId() +
                "/" + Constants.SECRET_MAP + "/" + secret.getId();

        vaultTemplate.delete(path);
        repository.delete(secret);
        return true;
    }


    public boolean updateCertificateCredential(CertificateCredential credential) {
        Optional<Secret> exSecret = repository.findById(credential.getToken());

        if (exSecret.isEmpty()) {
            String msg = " Credential with token " + credential.getToken() + " not found";
            LOGGER.error(msg);
            throw new CredentialStoreException("Invalid token", null);
        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credential.getTenantId() + "/" + credential.getOwnerId() +
                "/" + Constants.CERTIFICATES + "/" + credential.getToken();


        Certificate certificate = new Certificate(credential.getCert(),
                String.valueOf(credential.getLifetime()),
                credential.getNotBefore(),
                credential.getNotAfter(),
                credential.getPrivateKey());

        vaultTemplate.write(path, certificate);

        VaultResponseSupport<Certificate> response = vaultTemplate.read(path, Certificate.class);

        if (response == null || response.getData() == null && response.getData().getCertificate() == null) {
            String msg = " Certificate credential of tenant " + credential.getTenantId() +
                    " of user " + credential.getOwnerId() + " is not saved in vault";
            LOGGER.error(msg);
            throw new CredentialStoreException(msg, null);
        }

        Secret secret = new Secret();
        secret.setId(credential.getToken());
        secret.setDiscription(credential.getDescription());
        secret.setOwnerId(credential.getOwnerId());
        secret.setOwnerType(credential.getResourceOwnerType().name());
        secret.setSecretType(ResourceSecretType.X509_CERTIFICATE.name());
        secret.setTenantId(credential.getTenantId());
        secret.setExternalId(credential.getExternalId());
        secret.setType(credential.getType());
        repository.save(secret);
        return true;
    }


}
