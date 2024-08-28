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

import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.KeyFactory;
import java.security.KeyPair;
import java.security.MessageDigest;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.spec.PKCS8EncodedKeySpec;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;

@Component
public class KeyLoader {

    private KeyPair keyPair;
    private String keyID;

    @Value("${custos.keys.location}")
    private String keysLocation;

    @PostConstruct
    public void init() {
        try {
            this.keyPair = loadKeyPair();
            this.keyID = computeKeyID(keyPair.getPublic());
        } catch (Exception e) {
            throw new RuntimeException("Failed to load key pair", e);
        }
    }

    public KeyPair getKeyPair() {
        return this.keyPair;
    }

    public String getKeyID() {
        return this.keyID;
    }

    private KeyPair loadKeyPair() throws Exception {
        System.out.println("Key location - " + keysLocation);
        String privateKeyContent = new String(Files.readAllBytes(Paths.get(keysLocation + "/private_key.pem")));
        String publicKeyContent = new String(Files.readAllBytes(Paths.get(keysLocation + "/public_key.pem")));

        privateKeyContent = privateKeyContent.replaceAll("-----BEGIN PRIVATE KEY-----", "")
                .replaceAll("-----END PRIVATE KEY-----", "")
                .replaceAll("\\s+", "");

        publicKeyContent = publicKeyContent.replaceAll("-----BEGIN PUBLIC KEY-----", "")
                .replaceAll("-----END PUBLIC KEY-----", "")
                .replaceAll("\\s+", "");

        KeyFactory keyFactory = KeyFactory.getInstance("RSA");

        byte[] privateKeyBytes = Base64.getDecoder().decode(privateKeyContent);
        byte[] publicKeyBytes = Base64.getDecoder().decode(publicKeyContent);

        PKCS8EncodedKeySpec privateKeySpec = new PKCS8EncodedKeySpec(privateKeyBytes);
        X509EncodedKeySpec publicKeySpec = new X509EncodedKeySpec(publicKeyBytes);

        PrivateKey privateKey = keyFactory.generatePrivate(privateKeySpec);
        PublicKey publicKey = keyFactory.generatePublic(publicKeySpec);

        return new KeyPair(publicKey, privateKey);
    }

    private String computeKeyID(PublicKey publicKey) throws Exception {
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        byte[] hash = digest.digest(publicKey.getEncoded());
        return Base64.getUrlEncoder().withoutPadding().encodeToString(hash);
    }
}
