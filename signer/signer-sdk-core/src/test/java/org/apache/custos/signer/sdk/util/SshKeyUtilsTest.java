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

import org.apache.custos.signer.service.policy.KeyType;
import org.junit.jupiter.api.Test;

import java.security.KeyPair;
import java.util.Base64;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Unit tests for SshKeyUtils
 */
class SshKeyUtilsTest {

    @Test
    void testGenerateEd25519KeyPair() throws Exception {
        KeyPair keyPair = SshKeyUtils.generateKeyPair(KeyType.ED25519.id());

        assertNotNull(keyPair);
        assertNotNull(keyPair.getPrivate());
        assertNotNull(keyPair.getPublic());
        // Java reports Ed25519 as "EdDSA"
        assertEquals("EdDSA", keyPair.getPublic().getAlgorithm());
    }

    @Test
    void testKeyPairToPem() throws Exception {
        KeyPair keyPair = SshKeyUtils.generateKeyPair(KeyType.ED25519.id());
        String pem = SshKeyUtils.keyPairToPem(keyPair);

        assertNotNull(pem);
        assertTrue(pem.contains("BEGIN OPENSSH PRIVATE KEY"));
        assertTrue(pem.contains("END OPENSSH PRIVATE KEY"));

        // Verify OpenSSH header and basic structure
        String base64 = pem.lines()
                .filter(l -> !l.startsWith("-----"))
                .reduce("", (a, b) -> a + b);
        byte[] decoded = Base64.getDecoder().decode(base64);
        assertNotNull(decoded);
        assertTrue(new String(decoded, 0, "openssh-key-v1\0".length(), java.nio.charset.StandardCharsets.UTF_8)
                .startsWith("openssh-key-v1\0"));
    }

    @Test
    void testKeyPairToOpenSshPublicKey() throws Exception {
        KeyPair keyPair = SshKeyUtils.generateKeyPair(KeyType.ED25519.id());
        String opensshKey = SshKeyUtils.keyPairToOpenSshPublicKey(keyPair);

        assertNotNull(opensshKey);
        assertTrue(opensshKey.startsWith("ssh-ed25519 "));

        // Verify format: "ssh-ed25519 <base64-ssh-wire-blob> <comment>"
        String[] parts = opensshKey.split("\\s+");
        assertEquals(3, parts.length);
        assertEquals("ssh-ed25519", parts[0]);

        // Decode the SSH wire format blob
        byte[] wireBlob = Base64.getDecoder().decode(parts[1]);
        assertNotNull(wireBlob);
        assertTrue(wireBlob.length > 0);

        // Verify SSH wire format structure:
        // - 4 bytes: key type length (should be 11 for "ssh-ed25519")
        // - key type string bytes
        // - 4 bytes: public key length (should be 32)
        // - 32 bytes: public key
        java.nio.ByteBuffer buffer = java.nio.ByteBuffer.wrap(wireBlob);
        buffer.order(java.nio.ByteOrder.BIG_ENDIAN);

        // Read key type length
        int keyTypeLength = buffer.getInt();
        assertEquals(11, keyTypeLength); // "ssh-ed25519" is 11 bytes

        // Read key type string
        byte[] keyTypeBytes = new byte[keyTypeLength];
        buffer.get(keyTypeBytes);
        String keyType = new String(keyTypeBytes, java.nio.charset.StandardCharsets.UTF_8);
        assertEquals("ssh-ed25519", keyType);

        // Read public key length
        int publicKeyLength = buffer.getInt();
        assertEquals(32, publicKeyLength); // Ed25519 public key is 32 bytes

        // Read public key
        byte[] publicKeyBytes = new byte[publicKeyLength];
        buffer.get(publicKeyBytes);
        assertEquals(32, publicKeyBytes.length);

        // Verify we consumed the entire blob
        assertEquals(0, buffer.remaining());

        // Verify comment
        assertEquals("custos-generated", parts[2]);
    }

    @Test
    void testCertBytesToOpenSshCertString() throws Exception {
        // Create test certificate bytes (simplified - just test format)
        byte[] certBytes = new byte[]{1, 2, 3, 4, 5};
        String comment = "test-user";

        String certString = SshKeyUtils.certBytesToOpenSshCertString(certBytes, comment);

        assertNotNull(certString);
        assertTrue(certString.startsWith("ssh-ed25519-cert-v01@openssh.com "));

        // Verify format: "ssh-ed25519-cert-v01@openssh.com <base64> <comment>"
        String[] parts = certString.split("\\s+", 3);
        assertEquals(3, parts.length);
        assertEquals("ssh-ed25519-cert-v01@openssh.com", parts[0]);

        // Verify base64 decodes back to original bytes
        byte[] decoded = Base64.getDecoder().decode(parts[1]);
        assertArrayEquals(certBytes, decoded);

        // Verify comment
        assertEquals(comment, parts[2]);
    }

    @Test
    void testCertBytesToOpenSshCertStringWithoutComment() throws Exception {
        byte[] certBytes = new byte[]{1, 2, 3, 4, 5};

        String certString = SshKeyUtils.certBytesToOpenSshCertString(certBytes, null);

        assertNotNull(certString);
        assertTrue(certString.startsWith("ssh-ed25519-cert-v01@openssh.com "));

        // Should have prefix and base64, but no comment
        String[] parts = certString.split("\\s+");
        assertTrue(parts.length >= 2);
        assertEquals("ssh-ed25519-cert-v01@openssh.com", parts[0]);

        // Verify base64 decodes back to original bytes
        byte[] decoded = Base64.getDecoder().decode(parts[1]);
        assertArrayEquals(certBytes, decoded);
    }

    @Test
    void testKeyPairToPemWithNullKeyPair() {
        assertThrows(IllegalArgumentException.class, () -> {
            SshKeyUtils.keyPairToPem(null);
        });
    }

    @Test
    void testKeyPairToOpenSshPublicKeyWithNullKeyPair() {
        assertThrows(IllegalArgumentException.class, () -> {
            SshKeyUtils.keyPairToOpenSshPublicKey(null);
        });
    }

    @Test
    void testCertBytesToOpenSshCertStringWithNullBytes() {
        assertThrows(IllegalArgumentException.class, () -> {
            SshKeyUtils.certBytesToOpenSshCertString(null, "comment");
        });
    }

    @Test
    void testCertBytesToOpenSshCertStringWithEmptyBytes() {
        assertThrows(IllegalArgumentException.class, () -> {
            SshKeyUtils.certBytesToOpenSshCertString(new byte[0], "comment");
        });
    }
}
