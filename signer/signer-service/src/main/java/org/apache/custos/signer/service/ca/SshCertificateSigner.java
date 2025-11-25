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
package org.apache.custos.signer.service.ca;

import org.apache.custos.signer.service.vault.OpenBaoClient;
import org.bouncycastle.crypto.params.Ed25519PrivateKeyParameters;
import org.bouncycastle.crypto.signers.Ed25519Signer;
import org.bouncycastle.crypto.util.PrivateKeyFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.io.ByteArrayOutputStream;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.security.KeyPair;
import java.security.MessageDigest;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.time.Instant;
import java.util.Base64;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Random;

/**
 * Service for signing SSH certificates using Bouncy Castle.
 * Handles SSH certificate creation, signing, and validation.
 */
@Service
public class SshCertificateSigner {

    private static final Logger logger = LoggerFactory.getLogger(SshCertificateSigner.class);

    // SSH Certificate Types
    private static final int SSH_CERT_TYPE_USER = 1;
    private static final int SSH_CERT_TYPE_HOST = 2;

    // SSH Key Types
    private static final String SSH_KEY_TYPE_ED25519 = "ssh-ed25519";
    private static final String SSH_KEY_TYPE_RSA = "ssh-rsa";
    private static final String SSH_KEY_TYPE_ECDSA = "ecdsa-sha2-nistp256";

    @Autowired
    private OpenBaoClient openBaoClient;

    /**
     * Sign an SSH public key and return the certificate
     */
    public SshCertificateResult signCertificate(String tenantId, String clientId,
                                                String principal, int ttlSeconds,
                                                byte[] publicKeyBytes, String caFingerprint) {
        try {
            // Parse the public key
            SshPublicKey publicKey = parseSshPublicKey(publicKeyBytes);
            logger.debug("Parsed SSH public key: type={}, fingerprint={}",
                    publicKey.getKeyType(), publicKey.getFingerprint());

            // Get CA private key
            KeyPair caKeyPair = openBaoClient.getCurrentCAKey(tenantId, clientId)
                    .orElseThrow(() -> new RuntimeException("No CA key found for tenant: " + tenantId + ", client: " + clientId));

            // Get next serial number
            long serialNumber = openBaoClient.incrementSerialCounter(tenantId, clientId);

            // Calculate validity period
            Instant now = Instant.now();
            Instant validAfter = now;
            Instant validBefore = now.plusSeconds(ttlSeconds);

            // Generate key ID
            String keyId = String.format("%s@%s-%d", principal, clientId, now.getEpochSecond());

            // Build certificate
            SshCertificate certificate = buildSshCertificate(
                    publicKey, caKeyPair, serialNumber, keyId, principal,
                    validAfter, validBefore, caFingerprint
            );

            // Sign the certificate
            byte[] signature = signCertificate(certificate, caKeyPair.getPrivate());

            // Create final certificate with signature
            byte[] certificateBytes = createFinalCertificate(certificate, signature);

            logger.info("Successfully signed SSH certificate: serial={}, principal={}, ttl={}s",
                    serialNumber, principal, ttlSeconds);

            return new SshCertificateResult(
                    certificateBytes, serialNumber, validAfter, validBefore,
                    caFingerprint, publicKey.getFingerprint()
            );

        } catch (Exception e) {
            logger.error("Failed to sign SSH certificate for tenant: {}, client: {}, principal: {}",
                    tenantId, clientId, principal, e);
            throw new RuntimeException("Failed to sign SSH certificate", e);
        }
    }

    /**
     * Parse SSH public key from bytes
     */
    private SshPublicKey parseSshPublicKey(byte[] publicKeyBytes) throws Exception {
        // SSH public key format: "key-type" "base64-encoded-key" "comment"
        String publicKeyString = new String(publicKeyBytes, StandardCharsets.UTF_8).trim();
        String[] parts = publicKeyString.split("\\s+");

        if (parts.length < 2) {
            throw new IllegalArgumentException("Invalid SSH public key format");
        }

        String keyType = parts[0];
        byte[] keyData = Base64.getDecoder().decode(parts[1]);

        // Calculate fingerprint (SHA256 hash)
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        byte[] hash = digest.digest(keyData);
        String fingerprint = Base64.getEncoder().encodeToString(hash);

        return new SshPublicKey(keyType, keyData, fingerprint);
    }

    /**
     * Build SSH certificate structure
     */
    private SshCertificate buildSshCertificate(SshPublicKey publicKey, KeyPair caKeyPair,
                                               long serialNumber, String keyId, String principal,
                                               Instant validAfter, Instant validBefore,
                                               String caFingerprint) {
        SshCertificate cert = new SshCertificate();

        // Certificate header
        cert.setCertType(SSH_CERT_TYPE_USER);
        cert.setNonce(generateNonce());
        cert.setKeyType(publicKey.getKeyType());
        cert.setPublicKey(publicKey.getKeyData());
        cert.setSerial(serialNumber);
        cert.setKeyId(keyId);

        // Validity period
        cert.setValidAfter(validAfter.getEpochSecond());
        cert.setValidBefore(validBefore.getEpochSecond());

        // Critical options (empty for now, can be extended)
        cert.setCriticalOptions(new HashMap<>());

        // Extensions
        Map<String, String> extensions = new HashMap<>();
        extensions.put("permit-pty", "");
        extensions.put("permit-port-forwarding", "");
        extensions.put("permit-user-rc", "");
        cert.setExtensions(extensions);

        // Reserved field
        cert.setReserved("");

        // CA public key
        // TODO: Support multiple CA key types (RSA, ECDSA)
        cert.setCaKeyType(SSH_KEY_TYPE_ED25519); // Currently Ed25519 supported
        cert.setCaPublicKey(extractPublicKeyBytes(caKeyPair.getPublic()));

        return cert;
    }

    /**
     * Sign the certificate using CA private key
     */
    private byte[] signCertificate(SshCertificate certificate, PrivateKey caPrivateKey) throws Exception {
        // Serialize certificate for signing
        byte[] certificateData = serializeCertificate(certificate);

        // TODO: Add support for RSA and ECDSA signing algorithms
        // Sign using Ed25519
        if (caPrivateKey.getAlgorithm().equals("Ed25519")) {
            Ed25519Signer signer = new Ed25519Signer();
            Ed25519PrivateKeyParameters privateKeyParams = (Ed25519PrivateKeyParameters)
                    PrivateKeyFactory.createKey(caPrivateKey.getEncoded());
            signer.init(true, privateKeyParams);
            signer.update(certificateData, 0, certificateData.length);
            return signer.generateSignature();
        } else {
            throw new IllegalArgumentException("Unsupported CA key algorithm: " + caPrivateKey.getAlgorithm());
        }
    }

    /**
     * Create final certificate with signature
     */
    private byte[] createFinalCertificate(SshCertificate certificate, byte[] signature) throws Exception {
        ByteArrayOutputStream out = new ByteArrayOutputStream();

        // TODO: Select certificate format based on client key type (ssh-rsa-cert-v01@openssh.com, ecdsa-sha2-nistp256-cert-v01@openssh.com, etc.)
        // Write certificate type
        writeString(out, "ssh-ed25519-cert-v01@openssh.com");

        // Write nonce
        writeBytes(out, certificate.getNonce());

        // Write public key
        writeString(out, certificate.getKeyType());
        writeBytes(out, certificate.getPublicKey());

        // Write serial
        writeUint64(out, certificate.getSerial());

        // Write certificate type
        writeUint32(out, certificate.getCertType());

        // Write key ID
        writeString(out, certificate.getKeyId());

        // Write principals (single principal for now)
        writeStringList(out, Collections.singletonList(certificate.getKeyId().split("@")[0]));

        // Write validity period
        writeUint64(out, certificate.getValidAfter());
        writeUint64(out, certificate.getValidBefore());

        // Write critical options
        writeStringMap(out, certificate.getCriticalOptions());

        // Write extensions
        writeStringMap(out, certificate.getExtensions());

        // Write reserved
        writeString(out, certificate.getReserved());

        // Write CA public key
        writeString(out, certificate.getCaKeyType());
        writeBytes(out, certificate.getCaPublicKey());

        // TODO: Support signature types for RSA and ECDSA
        // Write signature
        writeString(out, "ssh-ed25519");
        writeBytes(out, signature);

        return out.toByteArray();
    }

    /**
     * Serialize certificate for signing (without signature)
     */
    private byte[] serializeCertificate(SshCertificate certificate) throws Exception {
        ByteArrayOutputStream out = new ByteArrayOutputStream();

        // TODO: Select certificate format based on client key type (currently hardcoded to Ed25519)
        // Write certificate type
        writeString(out, "ssh-ed25519-cert-v01@openssh.com");

        // Write nonce
        writeBytes(out, certificate.getNonce());

        // Write public key
        writeString(out, certificate.getKeyType());
        writeBytes(out, certificate.getPublicKey());

        // Write serial
        writeUint64(out, certificate.getSerial());

        // Write certificate type
        writeUint32(out, certificate.getCertType());

        // Write key ID
        writeString(out, certificate.getKeyId());

        // Write principals
        writeStringList(out, Collections.singletonList(certificate.getKeyId().split("@")[0]));

        // Write validity period
        writeUint64(out, certificate.getValidAfter());
        writeUint64(out, certificate.getValidBefore());

        // Write critical options
        writeStringMap(out, certificate.getCriticalOptions());

        // Write extensions
        writeStringMap(out, certificate.getExtensions());

        // Write reserved
        writeString(out, certificate.getReserved());

        // Write CA public key
        writeString(out, certificate.getCaKeyType());
        writeBytes(out, certificate.getCaPublicKey());

        return out.toByteArray();
    }

    private void writeString(ByteArrayOutputStream out, String str) throws Exception {
        byte[] bytes = str.getBytes(StandardCharsets.UTF_8);
        writeUint32(out, bytes.length);
        out.write(bytes);
    }

    private void writeBytes(ByteArrayOutputStream out, byte[] bytes) throws Exception {
        writeUint32(out, bytes.length);
        out.write(bytes);
    }

    private void writeUint32(ByteArrayOutputStream out, long value) throws Exception {
        ByteBuffer buffer = ByteBuffer.allocate(4);
        buffer.putInt((int) value);
        out.write(buffer.array());
    }

    private void writeUint64(ByteArrayOutputStream out, long value) throws Exception {
        ByteBuffer buffer = ByteBuffer.allocate(8);
        buffer.putLong(value);
        out.write(buffer.array());
    }

    private void writeStringList(ByteArrayOutputStream out, List<String> list) throws Exception {
        writeUint32(out, list.size());
        for (String str : list) {
            writeString(out, str);
        }
    }

    private void writeStringMap(ByteArrayOutputStream out, Map<String, String> map) throws Exception {
        writeUint32(out, map.size());
        for (Map.Entry<String, String> entry : map.entrySet()) {
            writeString(out, entry.getKey());
            writeString(out, entry.getValue());
        }
    }

    private byte[] generateNonce() {
        byte[] nonce = new byte[32];
        new Random().nextBytes(nonce);
        return nonce;
    }

    private byte[] extractPublicKeyBytes(PublicKey publicKey) {
        // Extract raw public key bytes from PublicKey object
        // This is a simplified implementation - TODO  need to handle different key types
        return publicKey.getEncoded();
    }

    /**
     * SSH Public Key container
     */
    public static class SshPublicKey {
        private final String keyType;
        private final byte[] keyData;
        private final String fingerprint;

        public SshPublicKey(String keyType, byte[] keyData, String fingerprint) {
            this.keyType = keyType;
            this.keyData = keyData;
            this.fingerprint = fingerprint;
        }

        public String getKeyType() {
            return keyType;
        }

        public byte[] getKeyData() {
            return keyData;
        }

        public String getFingerprint() {
            return fingerprint;
        }
    }

    /**
     * SSH Certificate container
     */
    public static class SshCertificate {
        private byte[] nonce;
        private String keyType;
        private byte[] publicKey;
        private long serial;
        private int certType;
        private String keyId;
        private long validAfter;
        private long validBefore;
        private Map<String, String> criticalOptions;
        private Map<String, String> extensions;
        private String reserved;
        private String caKeyType;
        private byte[] caPublicKey;

        // Getters and setters
        public byte[] getNonce() {
            return nonce;
        }

        public void setNonce(byte[] nonce) {
            this.nonce = nonce;
        }

        public String getKeyType() {
            return keyType;
        }

        public void setKeyType(String keyType) {
            this.keyType = keyType;
        }

        public byte[] getPublicKey() {
            return publicKey;
        }

        public void setPublicKey(byte[] publicKey) {
            this.publicKey = publicKey;
        }

        public long getSerial() {
            return serial;
        }

        public void setSerial(long serial) {
            this.serial = serial;
        }

        public int getCertType() {
            return certType;
        }

        public void setCertType(int certType) {
            this.certType = certType;
        }

        public String getKeyId() {
            return keyId;
        }

        public void setKeyId(String keyId) {
            this.keyId = keyId;
        }

        public long getValidAfter() {
            return validAfter;
        }

        public void setValidAfter(long validAfter) {
            this.validAfter = validAfter;
        }

        public long getValidBefore() {
            return validBefore;
        }

        public void setValidBefore(long validBefore) {
            this.validBefore = validBefore;
        }

        public Map<String, String> getCriticalOptions() {
            return criticalOptions;
        }

        public void setCriticalOptions(Map<String, String> criticalOptions) {
            this.criticalOptions = criticalOptions;
        }

        public Map<String, String> getExtensions() {
            return extensions;
        }

        public void setExtensions(Map<String, String> extensions) {
            this.extensions = extensions;
        }

        public String getReserved() {
            return reserved;
        }

        public void setReserved(String reserved) {
            this.reserved = reserved;
        }

        public String getCaKeyType() {
            return caKeyType;
        }

        public void setCaKeyType(String caKeyType) {
            this.caKeyType = caKeyType;
        }

        public byte[] getCaPublicKey() {
            return caPublicKey;
        }

        public void setCaPublicKey(byte[] caPublicKey) {
            this.caPublicKey = caPublicKey;
        }
    }

    /**
     * SSH Certificate signing result
     */
    public static class SshCertificateResult {
        private final byte[] certificate;
        private final long serialNumber;
        private final Instant validAfter;
        private final Instant validBefore;
        private final String caFingerprint;
        private final String publicKeyFingerprint;

        public SshCertificateResult(byte[] certificate, long serialNumber, Instant validAfter,
                                    Instant validBefore, String caFingerprint, String publicKeyFingerprint) {
            this.certificate = certificate;
            this.serialNumber = serialNumber;
            this.validAfter = validAfter;
            this.validBefore = validBefore;
            this.caFingerprint = caFingerprint;
            this.publicKeyFingerprint = publicKeyFingerprint;
        }

        public byte[] getCertificate() {
            return certificate;
        }

        public long getSerialNumber() {
            return serialNumber;
        }

        public Instant getValidAfter() {
            return validAfter;
        }

        public Instant getValidBefore() {
            return validBefore;
        }

        public String getCaFingerprint() {
            return caFingerprint;
        }

        public String getPublicKeyFingerprint() {
            return publicKeyFingerprint;
        }
    }
}
