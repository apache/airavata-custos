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
 */
package org.apache.custos.signer.sdk;

import org.apache.custos.signer.sdk.config.SdkConfiguration;
import org.apache.custos.signer.sdk.keystore.InMemoryKeyStore;
import org.apache.custos.signer.sdk.keystore.KeyStoreProvider;
import org.apache.custos.signer.service.policy.KeyType;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Objects;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Integration test for SSH client certificate request functionality.
 */
public class SshClientIntegrationTest {

    @ParameterizedTest
    @MethodSource("keyTypes")
    void testRequestCertificateMaterials(String keyType) throws Exception {
        SdkConfiguration config = new SdkConfiguration.Builder()
                .tenantId("nexus")
                .signerServiceAddress("localhost:9095")
                .tlsEnabled(false)
                .keyStoreBackend("in-memory")
                .addClient("test-client", "test-client", "test-secret")
                .build();

        KeyStoreProvider keyStore = new InMemoryKeyStore();

        SshClient client = new SshClient.Builder()
                .configuration(config)
                .keyStoreBackend(keyStore)
                .build();

        try {
            System.out.println("Requesting certificate materials (this will auto-generate CA key if first time) using keyType=" + keyType);

            CertificateMaterials materials = client.requestCertificateMaterials(
                    "test-client",  // client alias
                    "exouser",    // principal (SSH username)
                    3600,          // TTL: 1 hour
                    "test-token",  // user token (for now, just a placeholder TODO)
                    keyType
            );

            assertNotNull(materials, "Certificate materials should not be null");
            assertNotNull(materials.keyPair(), "KeyPair should not be null");
            assertNotNull(materials.privateKeyPem(), "Private key PEM should not be null");
            assertNotNull(materials.publicKeyOpenSsh(), "Public key OpenSSH should not be null");
            assertNotNull(materials.opensshCert(), "OpenSSH cert should not be null");
            assertNotNull(materials.certBytes(), "Certificate bytes should not be null");
            assertTrue(Objects.requireNonNull(materials.certBytes()).length > 0, "Certificate bytes should not be empty");

            System.out.println("✓ Certificate materials received successfully!");
            System.out.println("  Serial Number: " + materials.serial());
            System.out.println("  CA Fingerprint: " + materials.caFingerprint());
            System.out.println("  Target: " + materials.targetHost() + ":" + materials.targetPort());
            System.out.println("  Valid After: " + materials.validAfter());
            System.out.println("  Valid Before: " + materials.validBefore());

            String keyFile = "/tmp/test-key-" + keyType;
            String certFile = "/tmp/test-cert-" + keyType;
            Files.write(Paths.get(keyFile), materials.privateKeyPem().getBytes());
            String certContent = materials.opensshCert().trim();
            Files.write(Paths.get(certFile), certContent.getBytes());

            System.out.println("✓ Saved private key to: " + keyFile);
            System.out.println("✓ Saved certificate to: " + certFile);
            System.out.println("\nNext steps:");
            System.out.println("1. Extract CA public key from Vault:");
            System.out.println("   vault kv get -field=public_key ssh-ca/nexus/test-client/current > /tmp/ca.pub");
            System.out.println("2. Configure HPC node to trust the CA (see Phase 4 in plan)");
            System.out.println("3. Test SSH connection:");
            System.out.println("   ssh -i " + keyFile + " -o CertificateFile=" + certFile + " exouser@149.165.174.144 -p 22");

        } finally {
            client.close();
        }
    }

    private static Stream<String> keyTypes() {
        return Stream.of(KeyType.ED25519.id(), KeyType.RSA.id(), KeyType.ECDSA.id());
    }
}
