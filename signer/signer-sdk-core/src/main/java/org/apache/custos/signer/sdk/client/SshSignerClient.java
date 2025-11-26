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
package org.apache.custos.signer.sdk.client;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Metadata;
import io.grpc.StatusRuntimeException;
import org.apache.custos.signer.v1.GetJWKSRequest;
import org.apache.custos.signer.v1.GetJWKSResponse;
import org.apache.custos.signer.v1.RevokeRequest;
import org.apache.custos.signer.v1.RevokeResponse;
import org.apache.custos.signer.v1.SignRequest;
import org.apache.custos.signer.v1.SignResponse;
import org.apache.custos.signer.v1.SshSignerServiceGrpc;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Instant;
import java.util.concurrent.TimeUnit;

/**
 * gRPC client wrapper for communicating with the Custos Signer Service.
 * Handles certificate requests, revocation, and JWKS retrieval.
 */
public class SshSignerClient implements AutoCloseable {

    private static final Logger logger = LoggerFactory.getLogger(SshSignerClient.class);

    // gRPC metadata keys for client authentication
    private static final Metadata.Key<String> CLIENT_ID_KEY =
            Metadata.Key.of("client-id", Metadata.ASCII_STRING_MARSHALLER);
    private static final Metadata.Key<String> CLIENT_SECRET_KEY =
            Metadata.Key.of("client-secret", Metadata.ASCII_STRING_MARSHALLER);

    private final String tenantId;
    private final ManagedChannel channel;
    private final SshSignerServiceGrpc.SshSignerServiceBlockingStub blockingStub;
    private final SshSignerServiceGrpc.SshSignerServiceStub asyncStub;

    /**
     * SSH Signer Client Constructor
     *
     * @param tenantId             Tenant ID for all operations
     * @param signerServiceAddress Address of the signer service (host:port)
     * @param tlsEnabled           Whether to use TLS for gRPC communication
     */
    public SshSignerClient(String tenantId, String signerServiceAddress, boolean tlsEnabled) {
        this.tenantId = tenantId;
        ManagedChannelBuilder<?> channelBuilder = ManagedChannelBuilder.forTarget(signerServiceAddress);

        if (!tlsEnabled) {
            channelBuilder.usePlaintext();
        }

        // Configure channel settings
        channelBuilder
                .keepAliveTime(30, TimeUnit.SECONDS)
                .keepAliveTimeout(5, TimeUnit.SECONDS)
                .keepAliveWithoutCalls(true)
                .maxInboundMessageSize(4 * 1024 * 1024) // 4MB
                .maxInboundMetadataSize(8 * 1024); // 8KB

        this.channel = channelBuilder.build();
        this.blockingStub = SshSignerServiceGrpc.newBlockingStub(channel);
        this.asyncStub = SshSignerServiceGrpc.newStub(channel);

        logger.debug("SshSignerClient initialized for address: {}, TLS: {}", signerServiceAddress, tlsEnabled);
    }

    /**
     * Request a signed SSH certificate
     */
    public CertificateResponse requestCertificate(String clientId, String clientSecret,
                                                  String principal, int ttlSeconds, byte[] publicKeyBytes,
                                                  String userToken) {
        try {
            logger.debug("Requesting certificate for tenant: {}, client: {}, principal: {}",
                    tenantId, clientId, principal);

            SignRequest request = SignRequest.newBuilder()
                    .setTenantId(tenantId)
                    .setClientId(clientId)
                    .setPrincipal(principal)
                    .setTtlSeconds(ttlSeconds)
                    .setPublicKey(com.google.protobuf.ByteString.copyFrom(publicKeyBytes))
                    .setUserAccessToken(userToken)
                    .build();

            // Add authentication metadata
            // Client ID must be in format: tenantId:clientId for the interceptor
            Metadata metadata = new Metadata();
            metadata.put(CLIENT_ID_KEY, tenantId + ":" + clientId);
            metadata.put(CLIENT_SECRET_KEY, clientSecret);

            // Make gRPC call with metadata
            SignResponse response = blockingStub
                    .withInterceptors(createMetadataInterceptor(metadata))
                    .sign(request);

            logger.debug("Received certificate response: serial={}, target={}:{}",
                    response.getSerialNumber(), response.getTargetHost(), response.getTargetPort());

            return new CertificateResponse(
                    response.getCertificate().toByteArray(),
                    response.getSerialNumber(),
                    Instant.ofEpochSecond(response.getValidAfter()),
                    Instant.ofEpochSecond(response.getValidBefore()),
                    response.getCaFingerprint(),
                    response.getTargetHost(),
                    response.getTargetPort(),
                    response.getTargetUsername()
            );

        } catch (StatusRuntimeException e) {
            logger.error("gRPC error requesting certificate: {}", e.getStatus().getDescription());
            throw new SignerClientException("Failed to request certificate: " + e.getStatus().getDescription(), e);
        } catch (Exception e) {
            logger.error("Error requesting certificate", e);
            throw new SignerClientException("Failed to request certificate", e);
        }
    }

    /**
     * Revoke a certificate
     */
    public boolean revokeCertificate(String clientId, String clientSecret,
                                     Long serialNumber, String keyId, String caFingerprint, String reason) {
        try {
            logger.debug("Revoking certificate for tenant: {}, client: {}", tenantId, clientId);

            RevokeRequest.Builder requestBuilder = RevokeRequest.newBuilder()
                    .setTenantId(tenantId)
                    .setClientId(clientId)
                    .setReason(reason != null ? reason : "Revoked by client");

            if (serialNumber != null) {
                requestBuilder.setSerialNumber(serialNumber);
            }
            if (keyId != null) {
                requestBuilder.setKeyId(keyId);
            }
            if (caFingerprint != null) {
                requestBuilder.setCaFingerprint(caFingerprint);
            }

            RevokeRequest request = requestBuilder.build();

            // Add authentication metadata
            // Client ID must be in format: tenantId:clientId for the interceptor
            Metadata metadata = new Metadata();
            metadata.put(CLIENT_ID_KEY, tenantId + ":" + clientId);
            metadata.put(CLIENT_SECRET_KEY, clientSecret);

            // Make gRPC call
            RevokeResponse response = blockingStub
                    .withInterceptors(createMetadataInterceptor(metadata))
                    .revoke(request);

            logger.debug("Revocation response: success={}, count={}",
                    response.getSuccess(), response.getRevokedCount());

            return response.getSuccess();

        } catch (StatusRuntimeException e) {
            logger.error("gRPC error revoking certificate: {}", e.getStatus().getDescription());
            throw new SignerClientException("Failed to revoke certificate: " + e.getStatus().getDescription(), e);
        } catch (Exception e) {
            logger.error("Error revoking certificate", e);
            throw new SignerClientException("Failed to revoke certificate", e);
        }
    }

    /**
     * Get JWKS (CA public keys)
     */
    public JWKSResponse getJWKS(String clientId, String clientSecret) {
        try {
            logger.debug("Getting JWKS for tenant: {}, client: {}", tenantId, clientId);

            GetJWKSRequest request = GetJWKSRequest.newBuilder()
                    .setTenantId(tenantId)
                    .setClientId(clientId)
                    .build();

            // Add authentication metadata
            // Client ID must be in format: tenantId:clientId for the interceptor
            Metadata metadata = new Metadata();
            metadata.put(CLIENT_ID_KEY, tenantId + ":" + clientId);
            metadata.put(CLIENT_SECRET_KEY, clientSecret);

            // Make gRPC call
            GetJWKSResponse response = blockingStub
                    .withInterceptors(createMetadataInterceptor(metadata))
                    .getJWKS(request);

            logger.debug("Received JWKS response: current={}, next={}",
                    response.getCurrentFingerprint(), response.getNextFingerprint());

            return new JWKSResponse(
                    response.getCurrentKey(),
                    response.getNextKey(),
                    response.getCurrentFingerprint(),
                    response.getNextFingerprint(),
                    Instant.ofEpochSecond(response.getRotationScheduledAt())
            );

        } catch (StatusRuntimeException e) {
            logger.error("gRPC error getting JWKS: {}", e.getStatus().getDescription());
            throw new SignerClientException("Failed to get JWKS: " + e.getStatus().getDescription(), e);
        } catch (Exception e) {
            logger.error("Error getting JWKS", e);
            throw new SignerClientException("Failed to get JWKS", e);
        }
    }

    @Override
    public void close() {
        if (channel != null && !channel.isShutdown()) {
            logger.debug("Shutting down gRPC channel");
            channel.shutdown();
            try {
                if (!channel.awaitTermination(5, TimeUnit.SECONDS)) {
                    channel.shutdownNow();
                }
            } catch (InterruptedException e) {
                channel.shutdownNow();
                Thread.currentThread().interrupt();
            }
        }
    }

    private io.grpc.ClientInterceptor createMetadataInterceptor(Metadata metadata) {
        return new io.grpc.ClientInterceptor() {
            @Override
            public <ReqT, RespT> io.grpc.ClientCall<ReqT, RespT> interceptCall(
                    io.grpc.MethodDescriptor<ReqT, RespT> method,
                    io.grpc.CallOptions callOptions,
                    io.grpc.Channel next) {
                return new io.grpc.ForwardingClientCall.SimpleForwardingClientCall<ReqT, RespT>(
                        next.newCall(method, callOptions)) {
                    @Override
                    public void start(io.grpc.ClientCall.Listener<RespT> responseListener, io.grpc.Metadata headers) {
                        headers.merge(metadata);
                        super.start(responseListener, headers);
                    }
                };
            }
        };
    }

    /**
     * Certificate response container
     */
    public static class CertificateResponse {
        private final byte[] certificate;
        private final long serialNumber;
        private final Instant validAfter;
        private final Instant validBefore;
        private final String caFingerprint;
        private final String targetHost;
        private final int targetPort;
        private final String targetUsername;

        public CertificateResponse(byte[] certificate, long serialNumber, Instant validAfter,
                                   Instant validBefore, String caFingerprint, String targetHost,
                                   int targetPort, String targetUsername) {
            this.certificate = certificate;
            this.serialNumber = serialNumber;
            this.validAfter = validAfter;
            this.validBefore = validBefore;
            this.caFingerprint = caFingerprint;
            this.targetHost = targetHost;
            this.targetPort = targetPort;
            this.targetUsername = targetUsername;
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

        public String getTargetHost() {
            return targetHost;
        }

        public int getTargetPort() {
            return targetPort;
        }

        public String getTargetUsername() {
            return targetUsername;
        }
    }

    /**
     * JWKS response container
     */
    public static class JWKSResponse {
        private final String currentKey;
        private final String nextKey;
        private final String currentFingerprint;
        private final String nextFingerprint;
        private final Instant rotationScheduledAt;

        public JWKSResponse(String currentKey, String nextKey, String currentFingerprint,
                            String nextFingerprint, Instant rotationScheduledAt) {
            this.currentKey = currentKey;
            this.nextKey = nextKey;
            this.currentFingerprint = currentFingerprint;
            this.nextFingerprint = nextFingerprint;
            this.rotationScheduledAt = rotationScheduledAt;
        }

        public String getCurrentKey() {
            return currentKey;
        }

        public String getNextKey() {
            return nextKey;
        }

        public String getCurrentFingerprint() {
            return currentFingerprint;
        }

        public String getNextFingerprint() {
            return nextFingerprint;
        }

        public Instant getRotationScheduledAt() {
            return rotationScheduledAt;
        }
    }

    /**
     * Exception thrown by signer client operations
     */
    public static class SignerClientException extends RuntimeException {
        public SignerClientException(String message) {
            super(message);
        }

        public SignerClientException(String message, Throwable cause) {
            super(message, cause);
        }
    }
}
