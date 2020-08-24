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

import org.apache.custos.resource.secret.utils.Constants;
import org.apache.custos.resource.secret.exceptions.CredentialStoreException;
import org.apache.custos.resource.secret.persistance.local.model.Secret;
import org.apache.custos.resource.secret.persistance.local.repository.SecretRepository;
import org.apache.custos.resource.secret.persistance.vault.Certificate;
import org.apache.custos.resource.secret.persistance.vault.PasswordSecret;
import org.apache.custos.resource.secret.persistance.vault.SSHCredentialSecrets;
import org.apache.custos.resource.secret.service.ResourceSecretType;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.vault.core.VaultTemplate;
import org.springframework.vault.support.VaultResponseSupport;

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
            throw new CredentialStoreException(msg, null);
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
            throw new CredentialStoreException(msg, null);
        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credential.getTenantId() + "/" + credential.getOwnerId()
                + "/" + Constants.PASSWORD + "/" + credential.getToken();


        PasswordSecret passwordSecret = new PasswordSecret(credential.getPassword());
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
            throw new CredentialStoreException(msg, null);
        }

        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + credential.getTenantId() + "/" + credential.getOwnerId() +
                "/" + Constants.SSH_CREDENTIALS + "/" + credential.getToken();


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
        repository.save(secret);
        return true;
    }


    /**
     * delete existing credential
     * @param tenantId
     * @param token
     * @return
     */
    public boolean deleteCredential(long tenantId, String token) {

        Optional<Secret> exSec = repository.findById(token);

        if (exSec.isEmpty()) {
            return true;
        }

        Secret secret = exSec.get();

        String type = null;

        if (secret.getSecretType().equals(ResourceSecretType.SSH.name())) {
            type = Constants.SSH_CREDENTIALS;
        } else if (secret.getSecretType().equals(ResourceSecretType.PASSWORD.name())) {
            type = Constants.PASSWORD;
        } else if (secret.getSecretType().equals(ResourceSecretType.X509_CERTIFICATE.name())) {
            type = Constants.CERTIFICATES;
        }


        String path = Constants.VAULT_RESOURCE_SECRETS_PATH + tenantId + "/" + secret.getOwnerId() +
                "/" + type + "/" + token;

        vaultTemplate.delete(path);

        repository.delete(secret);
        return true;

    }




}
