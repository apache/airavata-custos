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

package org.apache.custos.credential.store.credential;

import org.apache.custos.credential.store.exceptions.CredentialGenerationException;
import org.apache.custos.credential.store.model.Credential;
import org.apache.custos.credential.store.model.CredentialTypes;
import org.apache.custos.credential.store.persistance.model.CredentialEntity;
import org.apache.custos.credential.store.persistance.repository.CredentialRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.security.SecureRandom;
import java.util.Base64;
import java.util.Random;

/**
 * A class used to generate clientId and clientSecret for custos
 */
@Component
public class CredentialManager {

    private static final Logger LOGGER = LoggerFactory.getLogger(CredentialManager.class);

    private String credentialPrefix = "custos/";

    private static final Random RANDOM = new SecureRandom();

    private static final String ALPHABET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";

    private static final int ID_LENGTH = 20;

    private static final int SECRET_LENGTH = 40;

    @Autowired
    private CredentialRepository repository;


    public Credential generateCredential(long ownerId, CredentialTypes type, long validTime) {
        try {

            String clientId = credentialPrefix + generateRandomClientId(ID_LENGTH) + "/" + ownerId;

            String secret = generateSecret(SECRET_LENGTH);

            Credential credential = new Credential();
            CredentialEntity entity = new CredentialEntity();

            entity.setClientId(clientId);
            entity.setClientSecretExpiredAt(0);
            entity.setOwnerId(ownerId);
            entity.setType(type.name());

            credential.setId(clientId);
            credential.setSecret(secret);

            repository.save(entity);

            return credential;
        } catch (Exception exception) {
            throw new CredentialGenerationException
                    ("Error occurred while generating credentials for " + ownerId, exception);
        }

    }


    public Credential decodeToken(String token) {

        try {
            byte[] array = Base64.getDecoder().decode(token);
            if (array != null && array.length > 0) {
                String decodeString = new String(array);
                String[] idSecretPair = decodeString.split(":");
                if (idSecretPair != null && idSecretPair.length == 2) {
                    Credential credential = new Credential();
                    credential.setId(idSecretPair[0]);
                    credential.setSecret(idSecretPair[1]);
                    return credential;
                }

            }
        } catch (Exception ex) {
            throw new CredentialGenerationException
                    ("Error occurred while decoding token ", ex);

        }
        return null;
    }


    private String generateRandomClientId(int length) {
        StringBuilder returnValue = new StringBuilder(length);
        for (int i = 0; i < length; i++) {
            returnValue.append(ALPHABET.charAt(RANDOM.nextInt(ALPHABET.length())));
        }
        return new String(returnValue);
    }

    private String generateSecret(int length) {
        StringBuilder returnValue = new StringBuilder(length);
        for (int i = 0; i < length; i++) {
            returnValue.append(ALPHABET.charAt(RANDOM.nextInt(ALPHABET.length())));
        }
        return new String(returnValue);
    }
}
