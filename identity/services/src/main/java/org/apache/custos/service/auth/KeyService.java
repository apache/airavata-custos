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

package org.apache.custos.service.auth;

import org.apache.custos.service.credential.store.CredentialStoreService;
import org.springframework.stereotype.Service;

import java.security.KeyFactory;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.MessageDigest;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.spec.PKCS8EncodedKeySpec;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;
import java.util.Map;

@Service
public class KeyService {

    private final CredentialStoreService credentialStoreService;

    private volatile KeyPair keyPair;
    private volatile String keyID;

    public KeyService(CredentialStoreService credentialStoreService) {
        this.credentialStoreService = credentialStoreService;
    }

    public KeyPair getKeyPair() {
        if (this.keyPair == null) {
            synchronized (this) {
                if (this.keyPair == null) {
                    loadOrGenerateKeyPair();
                }
            }
        }
        return this.keyPair;
    }

    public String getKeyID() {
        if (this.keyID == null) {
            synchronized (this) {
                if (this.keyID == null) {
                    this.keyID = computeKeyID(getKeyPair().getPublic());
                }
            }
        }
        return this.keyID;
    }

    private void loadOrGenerateKeyPair() {
        try {
            Map<String, String> keyData = credentialStoreService.retrieveKeyPair();
            if (keyData == null || keyData.isEmpty()) {
                generateAndStoreKeyPair();
            } else {
                this.keyPair = createKeyPairFromData(keyData);
            }
        } catch (Exception e) {
            throw new RuntimeException("Failed to load or generate key pair", e);
        }
    }

    private void generateAndStoreKeyPair() {
        try {
            KeyPairGenerator keyGen = KeyPairGenerator.getInstance("RSA");
            keyGen.initialize(2048);
            KeyPair generatedKeyPair = keyGen.generateKeyPair();

            String privateKey = Base64.getEncoder().encodeToString(generatedKeyPair.getPrivate().getEncoded());
            String publicKey = Base64.getEncoder().encodeToString(generatedKeyPair.getPublic().getEncoded());

            credentialStoreService.storeKeyPair(privateKey, publicKey);

            this.keyPair = generatedKeyPair;
            this.keyID = computeKeyID(generatedKeyPair.getPublic());
        } catch (Exception e) {
            throw new RuntimeException("Failed to generate and store key pair", e);
        }
    }

    private KeyPair createKeyPairFromData(Map<String, String> keyData) throws Exception {
        byte[] privateKeyBytes = Base64.getDecoder().decode(keyData.get("privateKey"));
        byte[] publicKeyBytes = Base64.getDecoder().decode(keyData.get("publicKey"));

        KeyFactory keyFactory = KeyFactory.getInstance("RSA");
        PrivateKey privateKey = keyFactory.generatePrivate(new PKCS8EncodedKeySpec(privateKeyBytes));
        PublicKey publicKey = keyFactory.generatePublic(new X509EncodedKeySpec(publicKeyBytes));

        return new KeyPair(publicKey, privateKey);
    }

    private String computeKeyID(PublicKey publicKey) {
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            byte[] hash = digest.digest(publicKey.getEncoded());
            return Base64.getUrlEncoder().withoutPadding().encodeToString(hash);
        } catch (Exception e) {
            throw new RuntimeException("Failed to compute key ID", e);
        }
    }
}
