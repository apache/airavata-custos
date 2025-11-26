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
package org.apache.custos.signer.service.grpc;

import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.server.service.GrpcService;
import org.apache.custos.signer.service.audit.AuditLogger;
import org.apache.custos.signer.service.auth.ClientAuthInterceptor;
import org.apache.custos.signer.service.auth.OidcTokenValidator;
import org.apache.custos.signer.service.ca.SshCertificateSigner;
import org.apache.custos.signer.service.model.ClientSshConfigEntity;
import org.apache.custos.signer.service.model.RevocationEvent;
import org.apache.custos.signer.service.policy.PolicyEnforcer;
import org.apache.custos.signer.service.repo.RevocationEventRepository;
import org.apache.custos.signer.service.vault.OpenBaoClient;
import org.apache.custos.signer.v1.GetJWKSRequest;
import org.apache.custos.signer.v1.GetJWKSResponse;
import org.apache.custos.signer.v1.RevokeRequest;
import org.apache.custos.signer.v1.RevokeResponse;
import org.apache.custos.signer.v1.SignRequest;
import org.apache.custos.signer.v1.SignResponse;
import org.apache.custos.signer.v1.SshSignerServiceGrpc;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

/**
 * gRPC service implementation for SSH certificate signing operations.
 * Handles Sign, Revoke, and GetJWKS requests.
 */
@GrpcService
public class SshSignerGrpcService extends SshSignerServiceGrpc.SshSignerServiceImplBase {

    private static final Logger logger = LoggerFactory.getLogger(SshSignerGrpcService.class);

    @Autowired
    private SshCertificateSigner certificateSigner;

    @Autowired
    private PolicyEnforcer policyEnforcer;

    @Autowired
    private OidcTokenValidator tokenValidator;

    @Autowired
    private AuditLogger auditLogger;

    @Autowired
    private RevocationEventRepository revocationEventRepository;

    @Autowired
    private OpenBaoClient openBaoClient;

    @Override
    public void sign(SignRequest request, StreamObserver<SignResponse> responseObserver) {
        try {
            logger.debug("Received sign request for tenant: {}, client: {}, principal: {}",
                    request.getTenantId(), request.getClientId(), request.getPrincipal());

            // Get authenticated client config from context
            ClientSshConfigEntity clientConfig = ClientAuthInterceptor.AUTHENTICATED_CLIENT_KEY.get();
            if (clientConfig == null) {
                throw new IllegalStateException("No authenticated client found in context");
            }

            // Validate user access token
            OidcTokenValidator.UserIdentity userIdentity = tokenValidator.validateToken(request.getUserAccessToken());

            // Enforce policy
            policyEnforcer.enforcePolicy(request, clientConfig);

            // Generate certificate
            String caFingerprint = calculateCaFingerprint(clientConfig.getTenantId(), clientConfig.getClientId());
            SshCertificateSigner.SshCertificateResult result = certificateSigner.signCertificate(
                    request.getTenantId(), request.getClientId(), request.getPrincipal(),
                    request.getTtlSeconds(), request.getPublicKey().toByteArray(), caFingerprint
            );

            // Calculate public key fingerprint for audit
            String publicKeyFingerprint = calculatePublicKeyFingerprint(request.getPublicKey().toByteArray());

            // Log certificate issuance
            auditLogger.logCertificateIssuance(
                    request.getTenantId(), request.getClientId(), result.getSerialNumber(),
                    calculateKeyId(request.getPublicKey().toByteArray()), request.getPrincipal(),
                    publicKeyFingerprint, caFingerprint,
                    LocalDateTime.ofInstant(result.getValidAfter(), ZoneOffset.UTC),
                    LocalDateTime.ofInstant(result.getValidBefore(), ZoneOffset.UTC),
                    getSourceIp(), tokenValidator.hashTokenForAudit(request.getUserAccessToken()),
                    createRequestMetadata(request)
            );

            SignResponse response = SignResponse.newBuilder()
                    .setCertificate(com.google.protobuf.ByteString.copyFrom(result.getCertificate()))
                    .setSerialNumber(result.getSerialNumber())
                    .setValidAfter(result.getValidAfter().getEpochSecond())
                    .setValidBefore(result.getValidBefore().getEpochSecond())
                    .setCaFingerprint(caFingerprint)
                    .setTargetHost(clientConfig.getTargetHost())
                    .setTargetPort(clientConfig.getTargetPort())
                    .setTargetUsername(request.getPrincipal()) // SSH username = principal
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

            logger.info("Successfully signed certificate: serial={}, principal={}, ttl={}s",
                    result.getSerialNumber(), request.getPrincipal(), request.getTtlSeconds());

        } catch (Exception e) {
            logger.error("Failed to sign certificate for tenant: {}, client: {}, principal: {}",
                    request.getTenantId(), request.getClientId(), request.getPrincipal(), e);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(e.getMessage()).asRuntimeException());
        }
    }

    @Override
    public void revoke(RevokeRequest request, StreamObserver<RevokeResponse> responseObserver) {
        try {
            logger.debug("Received revoke request for tenant: {}, client: {}",
                    request.getTenantId(), request.getClientId());

            // Get authenticated client config from context
            ClientSshConfigEntity clientConfig = ClientAuthInterceptor.AUTHENTICATED_CLIENT_KEY.get();
            if (clientConfig == null) {
                throw new IllegalStateException("No authenticated client found in context");
            }

            int revokedCount = 0;

            // Create revocation event based on revocation type
            if (request.getSerialNumber() > 0) {
                // Revoke by serial number
                RevocationEvent revocationEvent = RevocationEvent.forSerialNumber(
                        request.getTenantId(), request.getClientId(),
                        request.getSerialNumber(), request.getReason(), "client-" + request.getClientId()
                );
                revocationEventRepository.save(revocationEvent);
                revokedCount = 1;

                auditLogger.logCertificateRevocation(
                        request.getTenantId(), request.getClientId(), request.getSerialNumber(),
                        null, null, request.getReason(), "client-" + request.getClientId()
                );

            } else if (!request.getKeyId().isEmpty()) {
                // Revoke by key ID
                RevocationEvent revocationEvent = RevocationEvent.forKeyId(
                        request.getTenantId(), request.getClientId(),
                        request.getKeyId(), request.getReason(), "client-" + request.getClientId()
                );
                revocationEventRepository.save(revocationEvent);
                revokedCount = 1;

                auditLogger.logCertificateRevocation(
                        request.getTenantId(), request.getClientId(), null,
                        request.getKeyId(), null, request.getReason(), "client-" + request.getClientId()
                );

            } else if (!request.getCaFingerprint().isEmpty()) {
                // Revoke all certificates from CA
                RevocationEvent revocationEvent = RevocationEvent.forCaFingerprint(
                        request.getTenantId(), request.getClientId(),
                        request.getCaFingerprint(), request.getReason(), "client-" + request.getClientId()
                );
                revocationEventRepository.save(revocationEvent);
                revokedCount = 1; // KRL will handle the actual revocation count

                auditLogger.logCertificateRevocation(
                        request.getTenantId(), request.getClientId(), null,
                        null, request.getCaFingerprint(), request.getReason(), "client-" + request.getClientId()
                );

            } else {
                throw new IllegalArgumentException("At least one revocation identifier must be provided");
            }

            // Build response
            RevokeResponse response = RevokeResponse.newBuilder()
                    .setSuccess(true)
                    .setMessage("Certificate(s) revoked successfully")
                    .setRevokedCount(revokedCount)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

            logger.info("Successfully revoked certificate(s): tenant={}, client={}, count={}",
                    request.getTenantId(), request.getClientId(), revokedCount);

        } catch (Exception e) {
            logger.error("Failed to revoke certificate for tenant: {}, client: {}",
                    request.getTenantId(), request.getClientId(), e);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(e.getMessage()).asRuntimeException());
        }
    }

    @Override
    public void getJWKS(GetJWKSRequest request, StreamObserver<GetJWKSResponse> responseObserver) {
        try {
            logger.debug("Received getJWKS request for tenant: {}, client: {}",
                    request.getTenantId(), request.getClientId());

            // Get authenticated client config from context
            ClientSshConfigEntity clientConfig = ClientAuthInterceptor.AUTHENTICATED_CLIENT_KEY.get();
            if (clientConfig == null) {
                throw new IllegalStateException("No authenticated client found in context");
            }

            // Get current CA public key
            String currentKeyPem = openBaoClient.getCAPublicKeyPem(request.getTenantId(), request.getClientId())
                    .orElseThrow(() -> new RuntimeException("No CA public key found"));

            // Get next CA public key (if exists)
            String nextKeyPem = openBaoClient.getNextCAKey(request.getTenantId(), request.getClientId())
                    .map(keyPair -> {
                        try {
                            return toPemString(keyPair.getPublic());
                        } catch (Exception e) {
                            logger.warn("Failed to convert next CA key to PEM", e);
                            return null;
                        }
                    })
                    .orElse("");

            // Calculate fingerprints
            String currentFingerprint = calculateCaFingerprint(request.getTenantId(), request.getClientId());
            String nextFingerprint = nextKeyPem.isEmpty() ? "" : calculateKeyFingerprint(nextKeyPem);

            // Get rotation metadata
            OpenBaoClient.CAMetadata metadata = openBaoClient.getCAMetadata(request.getTenantId(), request.getClientId())
                    .orElse(new OpenBaoClient.CAMetadata());

            // Convert PEM to JWK format (simplified)
            String currentJwk = convertPemToJwk(currentKeyPem);
            String nextJwk = nextKeyPem.isEmpty() ? "" : convertPemToJwk(nextKeyPem);

            // Build response
            GetJWKSResponse response = GetJWKSResponse.newBuilder()
                    .setCurrentKey(currentJwk)
                    .setNextKey(nextJwk)
                    .setCurrentFingerprint(currentFingerprint)
                    .setNextFingerprint(nextFingerprint)
                    .setRotationScheduledAt(metadata.getNextRotationAt())
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

            logger.debug("Successfully returned JWKS for tenant: {}, client: {}",
                    request.getTenantId(), request.getClientId());

        } catch (Exception e) {
            logger.error("Failed to get JWKS for tenant: {}, client: {}",
                    request.getTenantId(), request.getClientId(), e);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(e.getMessage()).asRuntimeException());
        }
    }

    private String calculateCaFingerprint(String tenantId, String clientId) {
        try {
            String caKeyPem = openBaoClient.getCAPublicKeyPem(tenantId, clientId)
                    .orElseThrow(() -> new RuntimeException("No CA key found"));
            return calculateKeyFingerprint(caKeyPem);
        } catch (Exception e) {
            logger.error("Failed to calculate CA fingerprint", e);
            return "unknown";
        }
    }

    private String calculatePublicKeyFingerprint(byte[] publicKeyBytes) {
        try {
            String publicKeyString = new String(publicKeyBytes, StandardCharsets.UTF_8).trim();
            String[] parts = publicKeyString.split("\\s+");
            if (parts.length >= 2) {
                byte[] keyData = Base64.getDecoder().decode(parts[1]);
                MessageDigest digest = MessageDigest.getInstance("SHA-256");
                byte[] hash = digest.digest(keyData);
                return Base64.getEncoder().encodeToString(hash);
            }
            return "unknown";
        } catch (Exception e) {
            logger.error("Failed to calculate public key fingerprint", e);
            return "unknown";
        }
    }

    private String calculateKeyFingerprint(String keyPem) {
        try {
            // Extract key data from PEM format
            String[] lines = keyPem.split("\n");
            StringBuilder keyData = new StringBuilder();
            boolean inKeyData = false;
            for (String line : lines) {
                if (line.contains("BEGIN")) {
                    inKeyData = true;
                    continue;
                }
                if (line.contains("END")) {
                    break;
                }
                if (inKeyData) {
                    keyData.append(line.trim());
                }
            }

            byte[] decoded = Base64.getDecoder().decode(keyData.toString());
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            byte[] hash = digest.digest(decoded);
            return Base64.getEncoder().encodeToString(hash);
        } catch (Exception e) {
            logger.error("Failed to calculate key fingerprint", e);
            return "unknown";
        }
    }

    private String calculateKeyId(byte[] publicKeyBytes) {
        return calculatePublicKeyFingerprint(publicKeyBytes);
    }

    private String getSourceIp() {
        // TODO: Extract source IP from gRPC context
        return "unknown";
    }

    private Map<String, Object> createRequestMetadata(SignRequest request) {
        Map<String, Object> metadata = new HashMap<>();
        metadata.put("ttl_seconds", request.getTtlSeconds());
        metadata.put("public_key_length", request.getPublicKey().size());
        return metadata;
    }

    private String toPemString(java.security.PublicKey publicKey) throws Exception {
        java.io.StringWriter writer = new java.io.StringWriter();
        org.bouncycastle.openssl.jcajce.JcaPEMWriter pemWriter = new org.bouncycastle.openssl.jcajce.JcaPEMWriter(writer);
        pemWriter.writeObject(publicKey);
        pemWriter.close();
        return writer.toString();
    }

    private String convertPemToJwk(String pemKey) {
        // Simplified JWK conversion - TODO use a proper JWK library
        try {
            String fingerprint = calculateKeyFingerprint(pemKey);
            return String.format("{\"kty\":\"OKP\",\"crv\":\"Ed25519\",\"kid\":\"%s\"}", fingerprint);
        } catch (Exception e) {
            logger.error("Failed to convert PEM to JWK", e);
            return "{}";
        }
    }
}
