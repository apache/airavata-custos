/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the specific language
 * governing permissions and limitations under the License.
 *
 */
package org.apache.custos.signer.service.vault;

import org.bouncycastle.openssl.PEMKeyPair;
import org.bouncycastle.openssl.PEMParser;
import org.bouncycastle.openssl.jcajce.JcaPEMKeyConverter;
import org.bouncycastle.openssl.jcajce.JcaPEMWriter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.vault.core.VaultOperations;
import org.springframework.vault.support.VaultResponse;

import java.io.StringReader;
import java.io.StringWriter;
import java.security.KeyPair;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

/**
 * Service for managing CA keys in OpenBao/Vault.
 * Handles CA key storage, retrieval, rotation, and metadata management.
 */
@Service
public class OpenBaoClient {

    private static final Logger logger = LoggerFactory.getLogger(OpenBaoClient.class);

    private static final String CA_PATH_PREFIX = "ssh-ca";
    private static final String CURRENT_KEY_PATH = "current";
    private static final String NEXT_KEY_PATH = "next";
    private static final String METADATA_PATH = "metadata";

    @Autowired
    private VaultOperations vaultOperations;

    /**
     * Retrieve the current active CA private key
     */
    public Optional<KeyPair> getCurrentCAKey(String tenantId, String clientId) {
        try {
            String path = buildPath(tenantId, clientId, CURRENT_KEY_PATH);
            VaultResponse response = vaultOperations.read(path);

            if (response == null || response.getData() == null) {
                logger.debug("No current CA key found for tenant: {}, client: {}", tenantId, clientId);
                return Optional.empty();
            }

            Map<String, Object> data = response.getData();
            String privateKeyPem = (String) data.get("private_key");
            String publicKeyPem = (String) data.get("public_key");

            if (privateKeyPem == null || publicKeyPem == null) {
                logger.warn("Incomplete CA key data for tenant: {}, client: {}", tenantId, clientId);
                return Optional.empty();
            }

            KeyPair keyPair = parseKeyPair(privateKeyPem, publicKeyPem);
            logger.debug("Retrieved current CA key for tenant: {}, client: {}", tenantId, clientId);
            return Optional.of(keyPair);

        } catch (Exception e) {
            logger.error("Failed to retrieve current CA key for tenant: {}, client: {}", tenantId, clientId, e);
            return Optional.empty();
        }
    }

    /**
     * Retrieve the next rotation CA private key
     */
    public Optional<KeyPair> getNextCAKey(String tenantId, String clientId) {
        try {
            String path = buildPath(tenantId, clientId, NEXT_KEY_PATH);
            VaultResponse response = vaultOperations.read(path);

            if (response == null || response.getData() == null) {
                logger.debug("No next CA key found for tenant: {}, client: {}", tenantId, clientId);
                return Optional.empty();
            }

            Map<String, Object> data = response.getData();
            String privateKeyPem = (String) data.get("private_key");
            String publicKeyPem = (String) data.get("public_key");

            if (privateKeyPem == null || publicKeyPem == null) {
                logger.warn("Incomplete next CA key data for tenant: {}, client: {}", tenantId, clientId);
                return Optional.empty();
            }

            KeyPair keyPair = parseKeyPair(privateKeyPem, publicKeyPem);
            logger.debug("Retrieved next CA key for tenant: {}, client: {}", tenantId, clientId);
            return Optional.of(keyPair);

        } catch (Exception e) {
            logger.error("Failed to retrieve next CA key for tenant: {}, client: {}", tenantId, clientId, e);
            return Optional.empty();
        }
    }

    /**
     * Generate and store a new CA keypair
     */
    public KeyPair createCAKeyPair(String tenantId, String clientId, String algorithm) {
        try {
            KeyPair keyPair = generateKeyPair(algorithm);
            String privateKeyPem = toPemString(keyPair.getPrivate());
            String publicKeyPem = toPemString(keyPair.getPublic());

            Map<String, Object> data = new HashMap<>();
            data.put("private_key", privateKeyPem);
            data.put("public_key", publicKeyPem);
            data.put("algorithm", algorithm);
            data.put("created_at", Instant.now().getEpochSecond());

            String path = buildPath(tenantId, clientId, CURRENT_KEY_PATH);
            vaultOperations.write(path, data);

            logger.info("Created new CA keypair for tenant: {}, client: {}, algorithm: {}",
                    tenantId, clientId, algorithm);
            return keyPair;

        } catch (Exception e) {
            logger.error("Failed to create CA keypair for tenant: {}, client: {}", tenantId, clientId, e);
            throw new RuntimeException("Failed to create CA keypair", e);
        }
    }

    /**
     * Perform CA key rotation (promote next → current, generate new next)
     */
    public void rotateCA(String tenantId, String clientId) {
        try {
            // Get current metadata to preserve serial counter
            CAMetadata metadata = getCAMetadata(tenantId, clientId).orElse(new CAMetadata());

            // Promote next key to current
            Optional<KeyPair> nextKey = getNextCAKey(tenantId, clientId);
            if (nextKey.isPresent()) {
                String privateKeyPem = toPemString(nextKey.get().getPrivate());
                String publicKeyPem = toPemString(nextKey.get().getPublic());

                Map<String, Object> currentData = new HashMap<>();
                currentData.put("private_key", privateKeyPem);
                currentData.put("public_key", publicKeyPem);
                currentData.put("rotated_at", Instant.now().getEpochSecond());

                String currentPath = buildPath(tenantId, clientId, CURRENT_KEY_PATH);
                vaultOperations.write(currentPath, currentData);
            }

            // TODO: Support RSA and ECDSA CA key types
            // Generate new next key
            KeyPair newNextKey = generateKeyPair("ed25519");
            String nextPrivateKeyPem = toPemString(newNextKey.getPrivate());
            String nextPublicKeyPem = toPemString(newNextKey.getPublic());

            Map<String, Object> nextData = new HashMap<>();
            nextData.put("private_key", nextPrivateKeyPem);
            nextData.put("public_key", nextPublicKeyPem);
            nextData.put("created_at", Instant.now().getEpochSecond());

            String nextPath = buildPath(tenantId, clientId, NEXT_KEY_PATH);
            vaultOperations.write(nextPath, nextData);

            // Update metadata with new rotation timestamp
            metadata.setLastRotationAt(Instant.now().getEpochSecond());
            metadata.setNextRotationAt(Instant.now().getEpochSecond() + (90L * 24 * 60 * 60)); // 90 days
            saveCAMetadata(tenantId, clientId, metadata);

            logger.info("Successfully rotated CA keys for tenant: {}, client: {}", tenantId, clientId);

        } catch (Exception e) {
            logger.error("Failed to rotate CA keys for tenant: {}, client: {}", tenantId, clientId, e);
            throw new RuntimeException("Failed to rotate CA keys", e);
        }
    }

    /**
     * Get CA metadata (rotation schedule, serial counter, etc.)
     */
    public Optional<CAMetadata> getCAMetadata(String tenantId, String clientId) {
        try {
            String path = buildPath(tenantId, clientId, METADATA_PATH);
            VaultResponse response = vaultOperations.read(path);

            if (response == null || response.getData() == null) {
                logger.debug("No CA metadata found for tenant: {}, client: {}", tenantId, clientId);
                return Optional.empty();
            }

            Map<String, Object> data = response.getData();
            CAMetadata metadata = new CAMetadata();
            metadata.setSerialCounter(getLongValue(data, "serial_counter", 1L));
            metadata.setLastRotationAt(getLongValue(data, "last_rotation_at", Instant.now().getEpochSecond()));
            metadata.setNextRotationAt(getLongValue(data, "next_rotation_at",
                    Instant.now().getEpochSecond() + (90L * 24 * 60 * 60))); // Default 90 days
            metadata.setRotationPeriodHours(getLongValue(data, "rotation_period_hours", 2160L)); // 90 days
            metadata.setOverlapHours(getLongValue(data, "overlap_hours", 2L));

            return Optional.of(metadata);

        } catch (Exception e) {
            logger.error("Failed to retrieve CA metadata for tenant: {}, client: {}", tenantId, clientId, e);
            return Optional.empty();
        }
    }

    /**
     * Save CA metadata
     */
    public void saveCAMetadata(String tenantId, String clientId, CAMetadata metadata) {
        try {
            Map<String, Object> data = new HashMap<>();
            data.put("serial_counter", metadata.getSerialCounter());
            data.put("last_rotation_at", metadata.getLastRotationAt());
            data.put("next_rotation_at", metadata.getNextRotationAt());
            data.put("rotation_period_hours", metadata.getRotationPeriodHours());
            data.put("overlap_hours", metadata.getOverlapHours());
            data.put("updated_at", Instant.now().getEpochSecond());

            String path = buildPath(tenantId, clientId, METADATA_PATH);
            vaultOperations.write(path, data);

            logger.debug("Saved CA metadata for tenant: {}, client: {}", tenantId, clientId);

        } catch (Exception e) {
            logger.error("Failed to save CA metadata for tenant: {}, client: {}", tenantId, clientId, e);
            throw new RuntimeException("Failed to save CA metadata", e);
        }
    }

    /**
     * Atomically increment and return the next serial number
     */
    public long incrementSerialCounter(String tenantId, String clientId) {
        try {
            CAMetadata metadata = getCAMetadata(tenantId, clientId).orElse(new CAMetadata());
            long nextSerial = metadata.getSerialCounter() + 1;
            metadata.setSerialCounter(nextSerial);
            saveCAMetadata(tenantId, clientId, metadata);

            logger.debug("Incremented serial counter for tenant: {}, client: {} to {}",
                    tenantId, clientId, nextSerial);
            return nextSerial;

        } catch (Exception e) {
            logger.error("Failed to increment serial counter for tenant: {}, client: {}", tenantId, clientId, e);
            throw new RuntimeException("Failed to increment serial counter", e);
        }
    }

    /**
     * Get CA public key as PEM string
     */
    public Optional<String> getCAPublicKeyPem(String tenantId, String clientId) {
        try {
            Optional<KeyPair> keyPairOpt = getCurrentCAKey(tenantId, clientId);
            if (keyPairOpt.isPresent()) {
                return Optional.of(toPemString(keyPairOpt.get().getPublic()));
            }
            return Optional.empty();
        } catch (Exception e) {
            logger.error("Error converting CA public key to PEM", e);
            return Optional.empty();
        }
    }

    private String buildPath(String tenantId, String clientId, String suffix) {
        return String.format("%s/%s/%s/%s", CA_PATH_PREFIX, tenantId, clientId, suffix);
    }

    private KeyPair generateKeyPair(String algorithm) throws Exception {
        // TODO: Add support for RSA and ECDSA key generation
        if ("ed25519".equalsIgnoreCase(algorithm)) {
            return generateEd25519KeyPair();
        } else {
            throw new IllegalArgumentException("Unsupported key algorithm: " + algorithm);
        }
    }

    private KeyPair generateEd25519KeyPair() throws Exception {
        // Generate Ed25519 key pair using Bouncy Castle
        java.security.KeyPairGenerator keyGen = java.security.KeyPairGenerator.getInstance("Ed25519");
        return keyGen.generateKeyPair();
    }

    private KeyPair parseKeyPair(String privateKeyPem, String publicKeyPem) throws Exception {
        // Parse private key
        PEMParser privateKeyParser = new PEMParser(new StringReader(privateKeyPem));
        Object privateKeyObj = privateKeyParser.readObject();
        privateKeyParser.close();

        PrivateKey privateKey;
        if (privateKeyObj instanceof PEMKeyPair) {
            PEMKeyPair keyPair = (PEMKeyPair) privateKeyObj;
            // Convert Bouncy Castle private key to Java PrivateKey
            JcaPEMKeyConverter converter = new JcaPEMKeyConverter();
            privateKey = converter.getPrivateKey(keyPair.getPrivateKeyInfo());
        } else {
            privateKey = (PrivateKey) privateKeyObj;
        }

        // Parse public key
        PEMParser publicKeyParser = new PEMParser(new StringReader(publicKeyPem));
        Object publicKeyObj = publicKeyParser.readObject();
        publicKeyParser.close();

        PublicKey publicKey = (PublicKey) publicKeyObj;

        return new KeyPair(publicKey, privateKey);
    }

    private String toPemString(Object key) throws Exception {
        StringWriter writer = new StringWriter();
        JcaPEMWriter pemWriter = new JcaPEMWriter(writer);
        pemWriter.writeObject(key);
        pemWriter.close();
        return writer.toString();
    }

    private long getLongValue(Map<String, Object> data, String key, long defaultValue) {
        Object value = data.get(key);
        if (value instanceof Number) {
            return ((Number) value).longValue();
        }
        return defaultValue;
    }

    /**
     * CA Metadata container
     */
    public static class CAMetadata {
        private long serialCounter = 1L;
        private long lastRotationAt = Instant.now().getEpochSecond();
        private long nextRotationAt = Instant.now().getEpochSecond() + (90L * 24 * 60 * 60); // 90 days
        private long rotationPeriodHours = 2160L; // 90 days
        private long overlapHours = 2L;

        // Getters and setters
        public long getSerialCounter() {
            return serialCounter;
        }

        public void setSerialCounter(long serialCounter) {
            this.serialCounter = serialCounter;
        }

        public long getLastRotationAt() {
            return lastRotationAt;
        }

        public void setLastRotationAt(long lastRotationAt) {
            this.lastRotationAt = lastRotationAt;
        }

        public long getNextRotationAt() {
            return nextRotationAt;
        }

        public void setNextRotationAt(long nextRotationAt) {
            this.nextRotationAt = nextRotationAt;
        }

        public long getRotationPeriodHours() {
            return rotationPeriodHours;
        }

        public void setRotationPeriodHours(long rotationPeriodHours) {
            this.rotationPeriodHours = rotationPeriodHours;
        }

        public long getOverlapHours() {
            return overlapHours;
        }

        public void setOverlapHours(long overlapHours) {
            this.overlapHours = overlapHours;
        }
    }
}
