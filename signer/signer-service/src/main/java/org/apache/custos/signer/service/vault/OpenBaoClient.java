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
import org.springframework.stereotype.Service;
import org.springframework.vault.core.VaultOperations;
import org.springframework.vault.support.VaultResponse;

import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.StringReader;
import java.io.StringWriter;
import java.security.KeyPair;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.time.Instant;
import java.util.Base64;
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

    private final VaultOperations vaultOperations;

    public OpenBaoClient(VaultOperations vaultOperations) {
        this.vaultOperations = vaultOperations;
    }

    /**
     * Build path for KV v2 secrets engine
     * KV v2 uses "{mount}/data/{path}" format
     */
    private String buildKv2Path(String tenantId, String clientId, String suffix) {
        // For KV v2, full path format is: {mount}/data/{tenant}/{client}/{suffix}
        return String.format("%s/data/%s/%s/%s", CA_PATH_PREFIX, tenantId, clientId, suffix);
    }

    /**
     * Retrieve the current active CA private key
     */
    public Optional<KeyPair> getCurrentCAKey(String tenantId, String clientId) {
        try {
            // KV v2 uses "data/{path}" format
            String path = buildKv2Path(tenantId, clientId, CURRENT_KEY_PATH);
            VaultResponse response = vaultOperations.read(path);

            if (response == null || response.getData() == null) {
                logger.debug("No current CA key found for tenant: {}, client: {}", tenantId, clientId);
                return Optional.empty();
            }

            // KV v2 wraps data in a "data" key
            Map<String, Object> responseData = response.getData();
            @SuppressWarnings("unchecked")
            Map<String, Object> data = (Map<String, Object>) responseData.get("data");
            if (data == null) {
                data = responseData;
            }
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
            // KV v2 uses "data/{path}" format
            String path = buildKv2Path(tenantId, clientId, NEXT_KEY_PATH);
            VaultResponse response = vaultOperations.read(path);

            if (response == null || response.getData() == null) {
                logger.debug("No next CA key found for tenant: {}, client: {}", tenantId, clientId);
                return Optional.empty();
            }

            // KV v2 wraps data in a "data" key
            Map<String, Object> responseData = response.getData();
            @SuppressWarnings("unchecked")
            Map<String, Object> data = (Map<String, Object>) responseData.get("data");
            if (data == null) {
                data = responseData;
            }
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

            // For KV v2, data must be wrapped in a "data" key
            Map<String, Object> dataMap = new HashMap<>();
            dataMap.put("private_key", privateKeyPem);
            dataMap.put("public_key", publicKeyPem);
            dataMap.put("algorithm", algorithm);
            dataMap.put("created_at", Instant.now().getEpochSecond());

            Map<String, Object> data = new HashMap<>();
            data.put("data", dataMap);

            String path = buildKv2Path(tenantId, clientId, CURRENT_KEY_PATH);
            vaultOperations.write(path, data);

            logger.info("Created new CA keypair for tenant: {}, client: {}, algorithm: {}", tenantId, clientId, algorithm);
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

                Map<String, Object> currentDataMap = new HashMap<>();
                currentDataMap.put("private_key", privateKeyPem);
                currentDataMap.put("public_key", publicKeyPem);
                currentDataMap.put("rotated_at", Instant.now().getEpochSecond());

                Map<String, Object> currentData = new HashMap<>();
                currentData.put("data", currentDataMap);

                // KV v2 uses "{mount}/data/{path}" format for write operations
                String currentPath = buildKv2Path(tenantId, clientId, CURRENT_KEY_PATH);
                vaultOperations.write(currentPath, currentData);
            }

            // TODO: Support RSA and ECDSA CA key types
            // Generate new next key
            KeyPair newNextKey = generateKeyPair("ed25519");
            String nextPrivateKeyPem = toPemString(newNextKey.getPrivate());
            String nextPublicKeyPem = toPemString(newNextKey.getPublic());

            Map<String, Object> nextDataMap = new HashMap<>();
            nextDataMap.put("private_key", nextPrivateKeyPem);
            nextDataMap.put("public_key", nextPublicKeyPem);
            nextDataMap.put("created_at", Instant.now().getEpochSecond());

            Map<String, Object> nextData = new HashMap<>();
            nextData.put("data", nextDataMap);

            String nextPath = buildKv2Path(tenantId, clientId, NEXT_KEY_PATH);
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
            String path = buildKv2Path(tenantId, clientId, METADATA_PATH);
            VaultResponse response = vaultOperations.read(path);

            if (response == null || response.getData() == null) {
                logger.debug("No CA metadata found for tenant: {}, client: {}", tenantId, clientId);
                return Optional.empty();
            }

            Map<String, Object> responseData = response.getData();
            @SuppressWarnings("unchecked")
            Map<String, Object> data = (Map<String, Object>) responseData.get("data");
            if (data == null) {
                data = responseData;
            }
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
            Map<String, Object> dataMap = new HashMap<>();
            dataMap.put("serial_counter", metadata.getSerialCounter());
            dataMap.put("last_rotation_at", metadata.getLastRotationAt());
            dataMap.put("next_rotation_at", metadata.getNextRotationAt());
            dataMap.put("rotation_period_hours", metadata.getRotationPeriodHours());
            dataMap.put("overlap_hours", metadata.getOverlapHours());
            dataMap.put("updated_at", Instant.now().getEpochSecond());

            Map<String, Object> data = new HashMap<>();
            data.put("data", dataMap);

            String path = buildKv2Path(tenantId, clientId, METADATA_PATH);
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

    /**
     * Get CA public key in OpenSSH format (for TrustedUserCAKeys configuration).
     * <p>
     * Format: "ssh-ed25519 &lt;base64-encoded-ssh-wire-blob&gt; &lt;comment&gt;"
     * <p>
     * This is the format required by SSH's TrustedUserCAKeys directive.
     *
     * @param tenantId Tenant ID
     * @param clientId Client ID
     * @return OpenSSH-formatted public key string, or empty if not found
     */
    public Optional<String> getCAPublicKeyOpenSsh(String tenantId, String clientId) {
        try {
            Optional<KeyPair> keyPairOpt = getCurrentCAKey(tenantId, clientId);
            if (keyPairOpt.isPresent()) {
                KeyPair keyPair = keyPairOpt.get();
                // Extract Ed25519 public key bytes (32 bytes)
                byte[] encoded = keyPair.getPublic().getEncoded();
                if (encoded.length < 44) {
                    logger.warn("Invalid Ed25519 public key encoding for tenant: {}, client: {}", tenantId, clientId);
                    return Optional.empty();
                }

                // Extract the 32-byte public key (last 32 bytes)
                byte[] publicKeyBytes = new byte[32];
                System.arraycopy(encoded, encoded.length - 32, publicKeyBytes, 0, 32);

                // SSH wire format: [4-byte length][key-type-string][4-byte length][32-byte public key]
                ByteArrayOutputStream blobStream = new ByteArrayOutputStream();
                DataOutputStream blob = new DataOutputStream(blobStream);

                try {
                    // Write key type string length (4 bytes, big-endian)
                    byte[] keyTypeBytes = "ssh-ed25519".getBytes(java.nio.charset.StandardCharsets.UTF_8);
                    blob.writeInt(keyTypeBytes.length);
                    blob.write(keyTypeBytes);
                    blob.writeInt(publicKeyBytes.length);
                    blob.write(publicKeyBytes);
                    blob.flush();
                } finally {
                    blob.close();
                }

                // Encode entire blob to base64
                String base64Key = Base64.getEncoder().encodeToString(blobStream.toByteArray());

                // Format: "ssh-ed25519 <base64-ssh-wire-blob> <comment>"
                String opensshKey = "ssh-ed25519 " + base64Key + " custos-ca-" + tenantId + "-" + clientId;
                return Optional.of(opensshKey);
            }
            return Optional.empty();
        } catch (Exception e) {
            logger.error("Error converting CA public key to OpenSSH format for tenant: {}, client: {}", tenantId, clientId, e);
            return Optional.empty();
        }
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
        JcaPEMKeyConverter converter = new JcaPEMKeyConverter();
        if (privateKeyObj instanceof PEMKeyPair keyPair) {
            // Convert Bouncy Castle private key to Java PrivateKey
            privateKey = converter.getPrivateKey(keyPair.getPrivateKeyInfo());
        } else if (privateKeyObj instanceof org.bouncycastle.asn1.pkcs.PrivateKeyInfo) {
            privateKey = converter.getPrivateKey((org.bouncycastle.asn1.pkcs.PrivateKeyInfo) privateKeyObj);
        } else {
            privateKey = (PrivateKey) privateKeyObj;
        }

        // Parse public key
        PEMParser publicKeyParser = new PEMParser(new StringReader(publicKeyPem));
        Object publicKeyObj = publicKeyParser.readObject();
        publicKeyParser.close();

        PublicKey publicKey;
        if (publicKeyObj instanceof org.bouncycastle.asn1.x509.SubjectPublicKeyInfo) {
            publicKey = converter.getPublicKey((org.bouncycastle.asn1.x509.SubjectPublicKeyInfo) publicKeyObj);
        } else {
            publicKey = (PublicKey) publicKeyObj;
        }

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
