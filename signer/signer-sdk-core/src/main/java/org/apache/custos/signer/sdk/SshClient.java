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
package org.apache.custos.signer.sdk;

import org.apache.custos.signer.sdk.client.SshSignerClient;
import org.apache.custos.signer.sdk.config.SdkConfiguration;
import org.apache.custos.signer.sdk.keystore.InMemoryKeyStore;
import org.apache.custos.signer.sdk.keystore.KeyStoreProvider;
import org.apache.custos.signer.sdk.ssh.SshSession;
import org.apache.custos.signer.sdk.util.SshKeyUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.nio.charset.StandardCharsets;
import java.security.KeyPair;
import java.time.Duration;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Main facade for Custos SSH operations.
 * Provides a high-level API for secure SSH operations using short-lived certificates.
 */
public class SshClient implements AutoCloseable {

    private static final Logger logger = LoggerFactory.getLogger(SshClient.class);

    private final SdkConfiguration configuration;
    private final SshSignerClient signerClient;
    private final KeyStoreProvider keyStore;
    private final Map<String, SshSession> activeSessions = new ConcurrentHashMap<>();

    private SshClient(Builder builder) {
        this.configuration = builder.configuration;
        this.signerClient = new SshSignerClient(
                configuration.getTenantId(),
                builder.signerServiceAddress,
                builder.tlsEnabled
        );
        this.keyStore = builder.keyStore;

        logger.debug("SshClient initialized with tenant: {}, signer: {}, keystore: {}",
                configuration.getTenantId(), builder.signerServiceAddress, builder.keyStore.getClass().getSimpleName());
    }

    /**
     * Request certificate materials (keys and certificate).
     * This method returns all SSH connection materials (private key, public key, certificate, and metadata).
     *
     * @param clientAlias Alias configured in SDK configuration
     * @param principal   SSH username (typically from OIDC token)
     * @param ttlSeconds  Certificate TTL in seconds
     * @param userToken   OIDC user access token
     * @return CertificateMaterials containing all SSH connection materials
     * @throws SshClientException if certificate request fails
     */
    public CertificateMaterials requestCertificateMaterials(String clientAlias, String principal, int ttlSeconds, String userToken) {
        return requestCertificateMaterials(clientAlias, principal, ttlSeconds, userToken, null);
    }

    /**
     * Request certificate materials specifying key type.
     *
     * @param clientAlias Alias configured in SDK configuration
     * @param principal   SSH username
     * @param ttlSeconds  TTL in seconds
     * @param userToken   OIDC user access token
     * @param keyType     Key type ("ed25519", "rsa", "ecdsa")
     * @return CertificateMaterials
     */
    public CertificateMaterials requestCertificateMaterials(String clientAlias, String principal, int ttlSeconds, String userToken, String keyType) {
        try {
            logger.debug("Requesting certificate materials for client: {}, principal: {}, ttl: {}s, keyType: {}", clientAlias, principal, ttlSeconds, keyType);

            // Resolve client alias to configuration
            SdkConfiguration.ClientConfig clientConfig = configuration.getClientConfig(clientAlias)
                    .orElseThrow(() -> new IllegalArgumentException("Client alias not found: " + clientAlias));

            String defaultKeyType = clientConfig.getKeyType() == null ? "ed25519" : clientConfig.getKeyType();
            String normalizedKeyType = keyType == null ? defaultKeyType.toLowerCase() : keyType.toLowerCase();
            KeyPair keyPair = SshKeyUtils.generateKeyPair(normalizedKeyType);

            String publicKeyOpenSsh = SshKeyUtils.keyPairToOpenSshPublicKey(keyPair);
            byte[] publicKeyBytes = publicKeyOpenSsh.getBytes(StandardCharsets.UTF_8);

            // Request certificate from signer service
            SshSignerClient.CertificateResponse certResponse = signerClient.requestCertificate(
                    clientConfig.getClientId(),
                    clientConfig.getClientSecret(),
                    principal,
                    ttlSeconds,
                    publicKeyBytes,
                    userToken
            );

            logger.debug("Received certificate: serial={}, target={}:{}", certResponse.getSerialNumber(), certResponse.getTargetHost(), certResponse.getTargetPort());

            // Convert private key to PEM format
            String privateKeyPem = SshKeyUtils.keyPairToPem(keyPair);

            // Convert certificate bytes to OpenSSH cert string format
            // Use principal as comment
            String opensshCert = SshKeyUtils.certBytesToOpenSshCertString(
                    certResponse.getCertificate(), normalizedKeyType, principal);

            // Create defensive copy of certBytes
            byte[] certBytesCopy = certResponse.getCertificate().clone();

            // Build and return CertificateMaterials
            CertificateMaterials materials = new CertificateMaterials(
                    keyPair,
                    privateKeyPem,
                    publicKeyOpenSsh,
                    opensshCert,
                    certBytesCopy,
                    certResponse.getSerialNumber(),
                    certResponse.getValidAfter(),
                    certResponse.getValidBefore(),
                    certResponse.getCaFingerprint(),
                    certResponse.getTargetHost(),
                    certResponse.getTargetPort(),
                    certResponse.getTargetUsername()
            );

            logger.info("Certificate materials prepared: client={}, principal={}, serial={}", clientAlias, principal, certResponse.getSerialNumber());

            return materials;

        } catch (Exception e) {
            logger.error("Failed to request certificate materials for client: {}, principal: {}", clientAlias, principal, e);
            throw new SshClientException("Failed to request certificate materials", e);
        }
    }

    /**
     * Open an SSH session using client alias.
     * <p>
     * This method uses {@link #requestCertificateMaterials(String, String, int, String)}
     * internally and then manages the session lifecycle with keystore tracking.
     *
     * @param clientAlias Alias configured in SDK configuration
     * @param principal   SSH username (typically from OIDC token)
     * @param ttlSeconds  Certificate TTL in seconds
     * @param userToken   OIDC user access token
     * @return SSH session for remote operations
     */
    public SshSession openSession(String clientAlias, String principal, int ttlSeconds, String userToken) {
        try {
            logger.debug("Opening SSH session for client: {}, principal: {}, ttl: {}s", clientAlias, principal, ttlSeconds);

            // Request certificate materials (shared flow)
            CertificateMaterials materials = requestCertificateMaterials(clientAlias, principal, ttlSeconds, userToken);

            // Store keypair in keystore for session tracking
            String contextId = generateContextId(clientAlias, principal);
            keyStore.store(principal, contextId, materials.keyPair(), Duration.ofSeconds(ttlSeconds));

            // Create SSH session from materials
            SshSession session = new SshSession(
                    materials.targetHost(),
                    materials.targetPort(),
                    materials.targetUsername(),
                    materials.keyPair(),
                    materials.certBytes(),
                    materials.validAfter(),
                    materials.validBefore()
            );

            // Track active session for cleanup
            String sessionId = generateSessionId(clientAlias, principal);
            activeSessions.put(sessionId, session);

            logger.info("SSH session opened: client={}, principal={}, target={}:{}", clientAlias, principal, materials.targetHost(), materials.targetPort());

            return session;

        } catch (Exception e) {
            logger.error("Failed to open SSH session for client: {}, principal: {}",
                    clientAlias, principal, e);
            throw new SshClientException("Failed to open SSH session", e);
        }
    }

    /**
     * Revoke a certificate
     */
    public boolean revokeCertificate(String clientAlias, Long serialNumber, String keyId, String reason) {
        try {
            SdkConfiguration.ClientConfig clientConfig = configuration.getClientConfig(clientAlias)
                    .orElseThrow(() -> new IllegalArgumentException("Client alias not found: " + clientAlias));

            return signerClient.revokeCertificate(
                    clientConfig.getClientId(),
                    clientConfig.getClientSecret(),
                    serialNumber,
                    keyId,
                    null, // caFingerprint
                    reason
            );

        } catch (Exception e) {
            logger.error("Failed to revoke certificate for client: {}", clientAlias, e);
            throw new SshClientException("Failed to revoke certificate", e);
        }
    }

    /**
     * Get JWKS for a client
     */
    public SshSignerClient.JWKSResponse getJWKS(String clientAlias) {
        try {
            SdkConfiguration.ClientConfig clientConfig = configuration.getClientConfig(clientAlias)
                    .orElseThrow(() -> new IllegalArgumentException("Client alias not found: " + clientAlias));

            return signerClient.getJWKS(
                    clientConfig.getClientId(),
                    clientConfig.getClientSecret()
            );

        } catch (Exception e) {
            logger.error("Failed to get JWKS for client: {}", clientAlias, e);
            throw new SshClientException("Failed to get JWKS", e);
        }
    }

    @Override
    public void close() {
        logger.debug("Closing SshClient and cleaning up resources");

        // Close all active sessions
        activeSessions.values().forEach(session -> {
            try {
                session.close();
            } catch (Exception e) {
                logger.warn("Error closing SSH session", e);
            }
        });
        activeSessions.clear();

        // Close signer client
        if (signerClient != null) {
            signerClient.close();
        }

        // Close key store
        if (keyStore != null) {
            keyStore.clear();
        }

        logger.debug("SshClient closed successfully");
    }

    private String generateContextId(String clientAlias, String principal) {
        return clientAlias + "-" + principal + "-" + System.currentTimeMillis();
    }

    private String generateSessionId(String clientAlias, String principal) {
        return clientAlias + "-" + principal + "-" + System.currentTimeMillis();
    }

    /**
     * Builder for SshClient
     */
    public static class Builder {
        private SdkConfiguration configuration;
        private String signerServiceAddress = "localhost:9095";
        private boolean tlsEnabled = false;
        private KeyStoreProvider keyStore = new InMemoryKeyStore();

        public Builder configuration(SdkConfiguration configuration) {
            this.configuration = configuration;
            return this;
        }

        public Builder signerServiceAddress(String address) {
            this.signerServiceAddress = address;
            return this;
        }

        public Builder tlsEnabled(boolean enabled) {
            this.tlsEnabled = enabled;
            return this;
        }

        public Builder keyStoreBackend(KeyStoreProvider keyStore) {
            this.keyStore = keyStore;
            return this;
        }

        public SshClient build() {
            if (configuration == null) {
                throw new IllegalArgumentException("Configuration is required");
            }
            return new SshClient(this);
        }
    }

    /**
     * Exception thrown by SshClient operations
     */
    public static class SshClientException extends RuntimeException {
        public SshClientException(String message) {
            super(message);
        }

        public SshClientException(String message, Throwable cause) {
            super(message, cause);
        }
    }
}
