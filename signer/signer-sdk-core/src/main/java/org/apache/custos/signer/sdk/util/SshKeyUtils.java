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

import org.bouncycastle.openssl.jcajce.JcaPEMWriter;

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
     * Convert a private key to PEM format using Bouncy Castle.
     *
     * @param keyPair KeyPair containing the private key
     * @return PEM-formatted private key string
     * @throws Exception if conversion fails
     */
    public static String keyPairToPem(KeyPair keyPair) throws Exception {
        if (keyPair == null || keyPair.getPrivate() == null) {
            throw new IllegalArgumentException("KeyPair or private key is null");
        }

        StringWriter writer = new StringWriter();
        try (JcaPEMWriter pemWriter = new JcaPEMWriter(writer)) {
            pemWriter.writeObject(keyPair.getPrivate());
        }
        return writer.toString();
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

