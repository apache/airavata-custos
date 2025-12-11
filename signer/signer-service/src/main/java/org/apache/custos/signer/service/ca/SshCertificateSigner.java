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
import org.bouncycastle.crypto.params.Ed25519PublicKeyParameters;
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
import java.security.Signature;
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
    private static final String SSH_SIG_ALG_RSA_SHA256 = "rsa-sha2-256";

    @Autowired
    private OpenBaoClient openBaoClient;

    /**
     * Sign an SSH public key and return the certificate
     */
    public SshCertificateResult signCertificate(String tenantId, String clientId, String principal, int ttlSeconds, byte[] publicKeyBytes, String caFingerprint) {
        try {
            // Parse the public key
            SshPublicKey publicKey = parseSshPublicKey(publicKeyBytes);
            logger.debug("Parsed SSH public key: type={}, fingerprint={}",
                    publicKey.getKeyType(), publicKey.getFingerprint());

            // Get CA private key, auto-generate if it doesn't exist
            KeyPair caKeyPair = openBaoClient.getCurrentCAKey(tenantId, clientId)
                    .orElseGet(() -> {
                        logger.info("No CA key found for tenant: {}, client: {}, auto-generating new CA key", tenantId, clientId);
                        return openBaoClient.createCAKeyPair(tenantId, clientId, "ed25519");
                    });

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
            throw new RuntimeException("Failed to sign SSH certificate: " + e.getMessage(), e);
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
        byte[] decoded = Base64.getDecoder().decode(parts[1]);

        byte[] rawKeyBytes;
        byte[] keyDataForCert;
        if (SSH_KEY_TYPE_ED25519.equals(keyType)) {
            ByteBuffer buf = ByteBuffer.wrap(decoded);
            int typeLen = buf.getInt();
            byte[] typeBytes = new byte[typeLen];
            buf.get(typeBytes);
            String embeddedType = new String(typeBytes, StandardCharsets.UTF_8);
            if (!SSH_KEY_TYPE_ED25519.equals(embeddedType)) {
                throw new IllegalArgumentException("Mismatched key type inside public key blob: " + embeddedType);
            }
            int pkLen = buf.getInt();
            if (pkLen != 32) {
                throw new IllegalArgumentException("Unexpected Ed25519 public key length: " + pkLen);
            }
            rawKeyBytes = new byte[pkLen];
            buf.get(rawKeyBytes);
            // For certificates, include the length prefix for the raw 32-byte key
            ByteBuffer keyBuf = ByteBuffer.allocate(4 + pkLen);
            keyBuf.putInt(pkLen);
            keyBuf.put(rawKeyBytes);
            keyDataForCert = keyBuf.array();

        } else {
            ByteBuffer buf = ByteBuffer.wrap(decoded);
            int typeLen = buf.getInt();
            byte[] typeBytes = new byte[typeLen];
            buf.get(typeBytes);
            String embeddedType = new String(typeBytes, StandardCharsets.UTF_8);
            if (!embeddedType.equals(keyType)) {
                throw new IllegalArgumentException("Mismatched key type inside public key blob: " + embeddedType);
            }
            byte[] remaining = new byte[buf.remaining()];
            buf.get(remaining);
            rawKeyBytes = remaining;
            // For RSA/ECDSA the certificate needs the full SSH wire-format public key blob
            keyDataForCert = rawKeyBytes;
        }

        // Calculate fingerprint (SHA256 hash)
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        byte[] hash = digest.digest(rawKeyBytes);
        String fingerprint = Base64.getEncoder().encodeToString(hash);

        return new SshPublicKey(keyType, keyDataForCert, fingerprint);
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
        cert.setPublicKey(publicKey.getKeyData()); // For Ed25519, raw 32-byte public key
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
        String caKeyType = normalizeKeyType(caKeyPair.getPublic().getAlgorithm());
        cert.setCaKeyType(caKeyType);
        try {
            cert.setCaPublicKey(toSshPublicKeyBlob(caKeyPair.getPublic()));
        } catch (Exception e) {
            throw new RuntimeException("Failed to encode CA public key", e);
        }

        return cert;
    }

    /**
     * Sign the certificate using CA private key
     */
    private byte[] signCertificate(SshCertificate certificate, PrivateKey caPrivateKey) throws Exception {
        // Serialize certificate for signing
        byte[] certificateData = serializeCertificate(certificate);

        // Sign using Ed25519
        // Note: Java reports Ed25519 keys as "EdDSA" algorithm
        String algorithm = caPrivateKey.getAlgorithm();
        if ("EdDSA".equals(algorithm) || "Ed25519".equals(algorithm)) {
            Ed25519Signer signer = new Ed25519Signer();
            Ed25519PrivateKeyParameters privateKeyParams = (Ed25519PrivateKeyParameters)
                    PrivateKeyFactory.createKey(caPrivateKey.getEncoded());
            signer.init(true, privateKeyParams);
            signer.update(certificateData, 0, certificateData.length);
            return signer.generateSignature();
        } else if ("RSA".equalsIgnoreCase(algorithm)) {
            Signature signature = Signature.getInstance("SHA256withRSA");
            signature.initSign(caPrivateKey);
            signature.update(certificateData);
            return signature.sign();
        } else if ("EC".equalsIgnoreCase(algorithm)) {
            Signature signature = Signature.getInstance("SHA256withECDSA");
            signature.initSign(caPrivateKey);
            signature.update(certificateData);
            return signature.sign();
        } else {
            throw new IllegalArgumentException("Unsupported CA key algorithm: " + caPrivateKey.getAlgorithm());
        }
    }

    /**
     * Create final certificate with signature
     */
    private byte[] createFinalCertificate(SshCertificate certificate, byte[] signature) throws Exception {
        ByteArrayOutputStream out = new ByteArrayOutputStream();

        writeString(out, getCertificateType(certificate.getKeyType()));

        // Nonce
        writeBytes(out, certificate.getNonce());

        // Subject public key (SSH wire public key blob)
        out.write(certificate.getPublicKey());

        // Serial
        writeUint64(out, certificate.getSerial());

        // Certificate type (user/host)
        writeUint32(out, certificate.getCertType());

        // Key ID
        writeString(out, certificate.getKeyId());

        // Principals (list encoded inside a single string)
        writeBytes(out, encodePrincipals(Collections.singletonList(certificate.getKeyId().split("@")[0])));

        // Validity
        writeUint64(out, certificate.getValidAfter());
        writeUint64(out, certificate.getValidBefore());

        // Critical options
        writeBytes(out, encodeOptions(certificate.getCriticalOptions()));

        // Extensions
        writeBytes(out, encodeOptions(certificate.getExtensions()));

        // Reserved (empty)
        writeBytes(out, certificate.getReserved() == null
                ? new byte[0]
                : certificate.getReserved().getBytes(StandardCharsets.UTF_8));

        // CA public key (as SSH wire public key)
        writeBytes(out, certificate.getCaPublicKey());

        // Signature (wrapped as SSH signature blob)
        ByteArrayOutputStream sigBuf = new ByteArrayOutputStream();
        writeString(sigBuf, getSignatureAlgorithm(certificate.getCaKeyType()));
        writeBytes(sigBuf, signature);
        writeBytes(out, sigBuf.toByteArray());

        return out.toByteArray();
    }

    /**
     * Serialize certificate for signing (without signature)
     */
    private byte[] serializeCertificate(SshCertificate certificate) throws Exception {
        ByteArrayOutputStream out = new ByteArrayOutputStream();

        writeString(out, getCertificateType(certificate.getKeyType()));

        // Nonce
        writeBytes(out, certificate.getNonce());

        // Subject public key (SSH wire public key blob)
        out.write(certificate.getPublicKey());

        // Serial
        writeUint64(out, certificate.getSerial());

        // Certificate type (user/host)
        writeUint32(out, certificate.getCertType());

        // Key ID
        writeString(out, certificate.getKeyId());

        // Principals
        writeBytes(out, encodePrincipals(Collections.singletonList(certificate.getKeyId().split("@")[0])));

        // Validity period
        writeUint64(out, certificate.getValidAfter());
        writeUint64(out, certificate.getValidBefore());

        // Critical options
        writeBytes(out, encodeOptions(certificate.getCriticalOptions()));

        // Extensions
        writeBytes(out, encodeOptions(certificate.getExtensions()));

        // Reserved
        writeBytes(out, certificate.getReserved() == null
                ? new byte[0]
                : certificate.getReserved().getBytes(StandardCharsets.UTF_8));

        // CA public key
        writeBytes(out, certificate.getCaPublicKey());

        return out.toByteArray();
    }

    private byte[] encodePrincipals(List<String> principals) throws Exception {
        ByteArrayOutputStream principalsBuf = new ByteArrayOutputStream();
        for (String principal : principals) {
            writeString(principalsBuf, principal);
        }
        return principalsBuf.toByteArray();
    }

    private byte[] encodeOptions(Map<String, String> options) throws Exception {
        ByteArrayOutputStream optionsBuf = new ByteArrayOutputStream();
        if (options != null && !options.isEmpty()) {
            options.entrySet().stream()
                    .sorted(Map.Entry.comparingByKey())
                    .forEach(entry -> {
                        try {
                            writeString(optionsBuf, entry.getKey());
                            writeString(optionsBuf, entry.getValue());
                        } catch (Exception e) {
                            throw new RuntimeException("Failed to encode options", e);
                        }
                    });
        }
        return optionsBuf.toByteArray();
    }

    private byte[] toSshPublicKeyBlob(PublicKey publicKey) throws Exception {
        String algorithm = publicKey.getAlgorithm();
        ByteArrayOutputStream buf = new ByteArrayOutputStream();
        if ("EdDSA".equals(algorithm) || "Ed25519".equals(algorithm) || SSH_KEY_TYPE_ED25519.equals(algorithm)) {
            Ed25519PublicKeyParameters publicKeyParams = (Ed25519PublicKeyParameters)
                    org.bouncycastle.crypto.util.PublicKeyFactory.createKey(publicKey.getEncoded());
            byte[] publicKeyBytes = publicKeyParams.getEncoded();
            writeString(buf, SSH_KEY_TYPE_ED25519);
            writeBytes(buf, publicKeyBytes);
            return buf.toByteArray();

        } else if ("RSA".equalsIgnoreCase(algorithm)) {
            java.security.interfaces.RSAPublicKey rsa = (java.security.interfaces.RSAPublicKey) publicKey;
            writeString(buf, SSH_KEY_TYPE_RSA);
            writeMpInt(buf, rsa.getPublicExponent());
            writeMpInt(buf, rsa.getModulus());
            return buf.toByteArray();

        } else if ("EC".equalsIgnoreCase(algorithm)) {
            java.security.interfaces.ECPublicKey ec = (java.security.interfaces.ECPublicKey) publicKey;
            String curve = "nistp256";
            byte[] q = encodeEcPoint(ec.getW(), ec.getParams());
            writeString(buf, SSH_KEY_TYPE_ECDSA);
            writeString(buf, curve);
            writeBytes(buf, q);
            return buf.toByteArray();
        }

        throw new IllegalArgumentException("Unsupported CA public key type: " + algorithm);
    }

    private String getCertificateType(String keyType) {
        if (keyType == null) {
            return SSH_KEY_TYPE_ED25519 + "-cert-v01@openssh.com";
        }
        if (keyType.startsWith(SSH_KEY_TYPE_RSA)) {
            return SSH_KEY_TYPE_RSA + "-cert-v01@openssh.com";
        }
        if (keyType.startsWith(SSH_KEY_TYPE_ECDSA)) {
            return SSH_KEY_TYPE_ECDSA + "-cert-v01@openssh.com";
        }
        return SSH_KEY_TYPE_ED25519 + "-cert-v01@openssh.com";
    }

    private String getSignatureAlgorithm(String caKeyType) {
        if (caKeyType == null) {
            return SSH_KEY_TYPE_ED25519;
        }
        if (caKeyType.startsWith(SSH_KEY_TYPE_RSA)) {
            return SSH_SIG_ALG_RSA_SHA256;
        }
        if (caKeyType.startsWith(SSH_KEY_TYPE_ECDSA)) {
            return SSH_KEY_TYPE_ECDSA;
        }
        return SSH_KEY_TYPE_ED25519;
    }

    private String normalizeKeyType(String algorithm) {
        if (algorithm == null) {
            return SSH_KEY_TYPE_ED25519;
        }
        if ("RSA".equalsIgnoreCase(algorithm)) {
            return SSH_KEY_TYPE_RSA;
        }
        if ("EC".equalsIgnoreCase(algorithm)) {
            return SSH_KEY_TYPE_ECDSA;
        }
        if ("Ed25519".equalsIgnoreCase(algorithm) || "EdDSA".equalsIgnoreCase(algorithm)) {
            return SSH_KEY_TYPE_ED25519;
        }
        return algorithm;
    }

    private static byte[] encodeEcPoint(java.security.spec.ECPoint point, java.security.spec.ECParameterSpec params) {
        int fieldSize = (params.getCurve().getField().getFieldSize() + 7) / 8;
        byte[] x = toFixedLength(point.getAffineX(), fieldSize);
        byte[] y = toFixedLength(point.getAffineY(), fieldSize);
        byte[] q = new byte[1 + x.length + y.length];
        q[0] = 0x04;
        System.arraycopy(x, 0, q, 1, x.length);
        System.arraycopy(y, 0, q, 1 + x.length, y.length);
        return q;
    }

    private static byte[] toFixedLength(java.math.BigInteger value, int length) {
        byte[] bytes = value.toByteArray();
        if (bytes.length == length) {
            return bytes;
        }
        byte[] out = new byte[length];
        if (bytes.length > length) {
            System.arraycopy(bytes, bytes.length - length, out, 0, length);
        } else {
            System.arraycopy(bytes, 0, out, length - bytes.length, bytes.length);
        }
        return out;
    }

    private void writeMpInt(ByteArrayOutputStream out, java.math.BigInteger value) throws Exception {
        byte[] bytes = value.toByteArray();
        writeUint32(out, bytes.length);
        out.write(bytes);
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
