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
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */
package org.apache.custos.signer.sdk;

import org.apache.custos.signer.sdk.config.SdkConfiguration;
import org.apache.custos.signer.sdk.keystore.InMemoryKeyStore;
import org.apache.custos.signer.sdk.keystore.KeyStoreProvider;
import org.apache.custos.signer.sdk.ssh.SshSession;
import org.apache.custos.signer.sdk.util.TestJwtTokenGenerator;
import org.apache.custos.signer.service.policy.KeyType;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Instant;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.assertEquals;

/**
 * End-to-end integration test that requests certificate materials and performs SSH operations
 * (exec, upload, download, read) using the wrapped SshSession.
 */
public class SshSessionIntegrationTest {

    private static Stream<String> keyTypes() {
        return Stream.of(KeyType.ED25519.id(), KeyType.RSA.id(), KeyType.ECDSA.id());
    }

    @ParameterizedTest
    @MethodSource("keyTypes")
    void testSshSessionEndToEnd(String keyType) throws Exception {
        SdkConfiguration config = new SdkConfiguration.Builder()
                .tenantId("nexus")
                .signerServiceAddress("localhost:9095")
                .tlsEnabled(false)
                .keyStoreBackend("in-memory")
                .addClient("test-client", "test-client", "test-secret")
                .build();

        KeyStoreProvider keyStore = new InMemoryKeyStore();

        try (SshClient client = new SshClient.Builder()
                .configuration(config)
                .keyStoreBackend(keyStore)
                .build()) {

            String userToken = TestJwtTokenGenerator.generateTestToken("exouser");

            try (SshSession session = client.openSession(
                    "test-client",
                    "exouser",
                    600,
                    userToken,
                    keyType)) {

                // e¬xec
                SshSession.CommandResult whoami = session.exec("whoami");
                assertEquals(0, whoami.getExitCode());
                assertEquals("exouser", whoami.getStdout().trim());

                // Write and read back a file
                String remoteBase = "/tmp/custos-session-" + keyType + "-" + System.currentTimeMillis();
                String remoteFile = remoteBase + ".txt";
                String content = "hello from custos ssh session test (" + keyType + ") at " + Instant.now();
                session.write(content, remoteFile, 0644);
                String readBack = session.read(remoteFile).trim();
                assertEquals(content, readBack);

                // Upload then download a file
                Path uploadLocal = Files.createTempFile("custos-upload-" + keyType, ".txt");
                Path downloadLocal = Files.createTempFile("custos-download-" + keyType, ".txt");
                String uploadContent = "upload-content-" + keyType + "-" + Instant.now();
                Files.writeString(uploadLocal, uploadContent, StandardCharsets.UTF_8);

                String remoteUpload = remoteBase + ".upload";
                session.upload(uploadLocal.toString(), remoteUpload, 0644);
                session.download(remoteUpload, downloadLocal.toString());

                String downloaded = Files.readString(downloadLocal, StandardCharsets.UTF_8).trim();
                assertEquals(uploadContent, downloaded);

                // Clean up remote artifacts
                try {
                    session.rm(remoteFile, false);
                    session.rm(remoteUpload, false);
                } catch (Exception ignore) {
                    // ignore
                }

                Files.deleteIfExists(uploadLocal);
                Files.deleteIfExists(downloadLocal);
            }
        }
    }
}
