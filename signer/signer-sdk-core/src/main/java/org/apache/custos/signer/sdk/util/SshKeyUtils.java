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
package org.apache.custos.signer.sdk.util;

import org.bouncycastle.crypto.params.Ed25519PrivateKeyParameters;
import org.bouncycastle.crypto.util.PrivateKeyFactory;
import org.bouncycastle.openssl.jcajce.JcaPEMWriter;

import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.StringWriter;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.util.Base64;

/**
 * Utility class for SSH key generation and format conversion.
 * TODO: Add support for RSA and ECDSA key types.
 */
public final class SshKeyUtils {

    // TODO: Support multiple key types (RSA, ECDSA) in addition to Ed25519
    private static final String SSH_KEY_TYPE_ED25519 = "ssh-ed25519";
    private static final String SSH_CERT_TYPE_ED25519 = "ssh-ed25519-cert-v01@openssh.com";

    private SshKeyUtils() {
    }

    /**
     * Generate a new Ed25519 keypair.
     *
     * @return Generated Ed25519 KeyPair
     * @throws Exception if key generation fails
     */
    public static KeyPair generateEd25519KeyPair() throws Exception {
        KeyPairGenerator keyGen = KeyPairGenerator.getInstance("Ed25519");
        return keyGen.generateKeyPair();
    }

    /**
     * Convert a private key to OpenSSH format (for Ed25519).
     * <p>
     * OpenSSH format for Ed25519 private keys uses a specific binary format:
     * - Magic string: "openssh-key-v1\0"
     * - Cipher name: "none" (for unencrypted keys)
     * - KDF name: "none"
     * - Public key (SSH wire format)
     * - Private key (32 bytes for Ed25519)
     * <p>
     * This format is required for SSH to accept Ed25519 keys.
     *
     * @param keyPair KeyPair containing the private key
     * @return OpenSSH-formatted private key string (PEM encoded)
     * @throws Exception if conversion fails
     */
    public static String keyPairToPem(KeyPair keyPair) throws Exception {
        if (keyPair == null || keyPair.getPrivate() == null) {
            throw new IllegalArgumentException("KeyPair or private key is null");
        }

        // For Ed25519, generate OpenSSH format
        String algorithm = keyPair.getPrivate().getAlgorithm();
        if ("EdDSA".equals(algorithm) || "Ed25519".equals(algorithm)) {
            return keyPairToOpenSshPrivateKey(keyPair);
        }

        // For other key types, use standard PKCS#8 format
        StringWriter writer = new StringWriter();
        try (JcaPEMWriter pemWriter = new JcaPEMWriter(writer)) {
            pemWriter.writeObject(keyPair.getPrivate());
        }
        return writer.toString();
    }

    /**
     * Convert Ed25519 key pair to OpenSSH private key format.
     * <p>
     * OpenSSH private key format structure:
     * - Magic: "openssh-key-v1\0" (15 bytes + null terminator)
     * - Cipher name length (4 bytes) + cipher name ("none" for unencrypted)
     * - KDF name length (4 bytes) + KDF name ("none")
     * - KDF options (4 bytes length + data)
     * - Number of keys (4 bytes)
     * - Public key
     * - Private key
     * <p>
     * All encoded in base64 and wrapped in PEM headers.
     *
     * @param keyPair Ed25519 KeyPair
     * @return OpenSSH private key in PEM format
     * @throws Exception if conversion fails
     */
    private static String keyPairToOpenSshPrivateKey(KeyPair keyPair) throws Exception {
        // Extract Ed25519 private key bytes (32 bytes)
        Ed25519PrivateKeyParameters privateKeyParams = (Ed25519PrivateKeyParameters) PrivateKeyFactory.createKey(keyPair.getPrivate().getEncoded());
        byte[] privateKeyBytes = privateKeyParams.getEncoded();

        // Extract Ed25519 public key bytes (32 bytes)
        byte[] encoded = keyPair.getPublic().getEncoded();
        byte[] publicKeyBytes = new byte[32];
        System.arraycopy(encoded, encoded.length - 32, publicKeyBytes, 0, 32);

        // Build SSH wire format for public key
        ByteArrayOutputStream publicKeyBlob = new ByteArrayOutputStream();
        DataOutputStream publicKeyOut = new DataOutputStream(publicKeyBlob);
        byte[] keyTypeBytes = SSH_KEY_TYPE_ED25519.getBytes(java.nio.charset.StandardCharsets.UTF_8);
        publicKeyOut.writeInt(keyTypeBytes.length);
        publicKeyOut.write(keyTypeBytes);
        publicKeyOut.writeInt(publicKeyBytes.length);
        publicKeyOut.write(publicKeyBytes);
        publicKeyOut.flush();
        byte[] publicKeyBlobBytes = publicKeyBlob.toByteArray();

        // Build unencrypted private section as defined by PROTOCOL.key:
        // uint32 checkint1, uint32 checkint2 (same value)
        // string keytype
        // string public key (same blob as above)
        // string private key (for ed25519: 32-byte priv + 32-byte pub)
        // string comment
        // padding 1,2,3.. to block boundary
        java.security.SecureRandom random = new java.security.SecureRandom();
        int checkInt = random.nextInt();

        ByteArrayOutputStream privateBuf = new ByteArrayOutputStream();
        DataOutputStream privOut = new DataOutputStream(privateBuf);

        privOut.writeInt(checkInt);
        privOut.writeInt(checkInt);

        writeString(privOut, SSH_KEY_TYPE_ED25519.getBytes(java.nio.charset.StandardCharsets.UTF_8));

        // public key (raw 32-byte Ed25519 public key)
        writeString(privOut, publicKeyBytes);

        // private key (64 bytes: private + public)
        byte[] fullPrivate = new byte[privateKeyBytes.length + publicKeyBytes.length];
        System.arraycopy(privateKeyBytes, 0, fullPrivate, 0, privateKeyBytes.length);
        System.arraycopy(publicKeyBytes, 0, fullPrivate, privateKeyBytes.length, publicKeyBytes.length);
        writeString(privOut, fullPrivate);

        // comment
        writeString(privOut, "custos-generated".getBytes(java.nio.charset.StandardCharsets.UTF_8));

        // padding with 1..n bytes
        int paddingNeeded = (8 - (privateBuf.size() % 8)) % 8;
        for (int i = 1; i <= paddingNeeded; i++) {
            privOut.write(i);
        }
        privOut.flush();
        byte[] privateKeyBlobBytes = privateBuf.toByteArray();

        // Build OpenSSH key format
        ByteArrayOutputStream opensshKey = new ByteArrayOutputStream();
        DataOutputStream out = new DataOutputStream(opensshKey);

        // Magic string: "openssh-key-v1\0"
        out.write("openssh-key-v1\0".getBytes(java.nio.charset.StandardCharsets.UTF_8));

        // Cipher name: "none" (unencrypted)
        byte[] cipherName = "none".getBytes(java.nio.charset.StandardCharsets.UTF_8);
        out.writeInt(cipherName.length);
        out.write(cipherName);

        // KDF name: "none"
        byte[] kdfName = "none".getBytes(java.nio.charset.StandardCharsets.UTF_8);
        out.writeInt(kdfName.length);
        out.write(kdfName);

        // KDF options: empty (4 bytes for length = 0)
        out.writeInt(0);

        // Number of keys: 1
        out.writeInt(1);

        // Public key (with length prefix)
        writeString(out, publicKeyBlobBytes);

        // Private key (with length prefix)
        writeString(out, privateKeyBlobBytes);

        out.flush();

        String base64Key = Base64.getEncoder().encodeToString(opensshKey.toByteArray());

        // Format as PEM
        StringBuilder pem = new StringBuilder();
        pem.append("-----BEGIN OPENSSH PRIVATE KEY-----\n");
        // Split base64 into 70-character lines
        for (int i = 0; i < base64Key.length(); i += 70) {
            int end = Math.min(i + 70, base64Key.length());
            pem.append(base64Key, i, end);
            pem.append("\n");
        }
        pem.append("-----END OPENSSH PRIVATE KEY-----\n");

        return pem.toString();
    }

    private static void writeString(DataOutputStream out, byte[] data) throws Exception {
        out.writeInt(data.length);
        out.write(data);
    }

    /**
     * Convert a public key to OpenSSH format.
     * <p>
     * Format: "ssh-ed25519 &lt;base64-encoded-ssh-wire-blob&gt; &lt;comment&gt;"
     * <p>
     * The SSH wire format blob contains:
     * - 4 bytes: length of key type string ("ssh-ed25519" = 11 bytes)
     * - key type string bytes ("ssh-ed25519")
     * - 4 bytes: length of public key (32 bytes)
     * - 32 bytes: public key
     * <p>
     * This entire blob is then base64-encoded to produce the OpenSSH public key format.
     * <p>
     * TODO: Support RSA and ECDSA key types (detect key type and format accordingly).
     *
     * @param keyPair KeyPair containing the public key
     * @return OpenSSH-formatted public key string
     * @throws Exception if conversion fails
     */
    public static String keyPairToOpenSshPublicKey(KeyPair keyPair) throws Exception {
        if (keyPair == null || keyPair.getPublic() == null) {
            throw new IllegalArgumentException("KeyPair or public key is null");
        }

        // Extract Ed25519 public key bytes (32 bytes)
        // The encoded format for Ed25519 is: OID (12 bytes) + public key (32 bytes) = 44 bytes
        byte[] encoded = keyPair.getPublic().getEncoded();
        if (encoded.length < 44) {
            throw new IllegalArgumentException("Invalid Ed25519 public key encoding");
        }

        // Extract the 32-byte public key (last 32 bytes)
        byte[] publicKeyBytes = new byte[32];
        System.arraycopy(encoded, encoded.length - 32, publicKeyBytes, 0, 32);

        // SSH wire format: [4-byte length][key-type-string][4-byte length][32-byte public key]
        java.io.ByteArrayOutputStream blobStream = new java.io.ByteArrayOutputStream();
        java.io.DataOutputStream blob = new java.io.DataOutputStream(blobStream);

        try {
            // Write key type string length (4 bytes, big-endian)
            byte[] keyTypeBytes = SSH_KEY_TYPE_ED25519.getBytes(java.nio.charset.StandardCharsets.UTF_8);
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
        return SSH_KEY_TYPE_ED25519 + " " + base64Key + " custos-generated";
    }

    /**
     * Convert certificate bytes to OpenSSH certificate string format.
     * <p>
     * Format: "ssh-ed25519-cert-v01@openssh.com &lt;base64-encoded-certificate&gt; &lt;comment&gt;"
     * <p>
     * The certificate bytes are base64-encoded and prefixed with the appropriate
     * certificate type identifier. The comment typically contains the principal
     * or client alias for identification.
     * <p>
     * TODO: Detect certificate type from bytes and use appropriate format (ssh-rsa-cert-v01@openssh.com, ecdsa-sha2-nistp256-cert-v01@openssh.com, etc.).
     *
     * @param certBytes Certificate bytes (raw binary certificate data)
     * @param comment   Comment to append (typically principal or clientAlias)
     * @return OpenSSH certificate string format
     * @throws Exception if conversion fails
     */
    public static String certBytesToOpenSshCertString(byte[] certBytes, String comment) throws Exception {
        if (certBytes == null || certBytes.length == 0) {
            throw new IllegalArgumentException("Certificate bytes cannot be null or empty");
        }

        // Encode certificate bytes to base64
        String base64Cert = Base64.getEncoder().encodeToString(certBytes);

        // Format: "ssh-ed25519-cert-v01@openssh.com <base64> <comment>"
        String certString = SSH_CERT_TYPE_ED25519 + " " + base64Cert;
        if (comment != null && !comment.isEmpty()) {
            certString += " " + comment;
        }

        return certString;
    }
}
