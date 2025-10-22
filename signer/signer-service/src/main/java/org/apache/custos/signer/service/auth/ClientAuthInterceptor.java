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
package org.apache.custos.signer.service.auth;

import io.grpc.Context;
import io.grpc.Contexts;
import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;
import io.grpc.Status;
import org.apache.custos.signer.service.model.ClientSshConfigEntity;
import org.apache.custos.signer.service.repo.ClientSshConfigRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Component;

import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.Optional;

/**
 * gRPC server interceptor for client authentication.
 * Extracts client credentials from gRPC metadata and validates them against stored secrets.
 */
@Component
public class ClientAuthInterceptor implements ServerInterceptor {

    private static final Logger logger = LoggerFactory.getLogger(ClientAuthInterceptor.class);

    // gRPC metadata keys for client authentication
    private static final Metadata.Key<String> CLIENT_ID_KEY =
            Metadata.Key.of("client-id", Metadata.ASCII_STRING_MARSHALLER);
    private static final Metadata.Key<String> CLIENT_SECRET_KEY =
            Metadata.Key.of("client-secret", Metadata.ASCII_STRING_MARSHALLER);

    // Context key for storing authenticated client config
    public static final Context.Key<ClientSshConfigEntity> AUTHENTICATED_CLIENT_KEY =
            Context.key("authenticated-client");

    @Autowired
    private ClientSshConfigRepository clientConfigRepository;

    private final PasswordEncoder passwordEncoder = new BCryptPasswordEncoder();

    @Override
    public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(
            ServerCall<ReqT, RespT> call, Metadata headers, ServerCallHandler<ReqT, RespT> next) {

        try {
            // Extract client credentials from metadata
            String clientId = headers.get(CLIENT_ID_KEY);
            String clientSecret = headers.get(CLIENT_SECRET_KEY);

            if (clientId == null || clientSecret == null) {
                logger.warn("Missing client credentials in gRPC metadata");
                call.close(Status.UNAUTHENTICATED.withDescription("Missing client credentials"), headers);
                return new ServerCall.Listener<ReqT>() {
                };
            }

            // Extract tenant ID from client ID (assuming format: tenantId:clientId)
            String[] clientIdParts = clientId.split(":", 2);
            if (clientIdParts.length != 2) {
                logger.warn("Invalid client ID format: {}", clientId);
                call.close(Status.UNAUTHENTICATED.withDescription("Invalid client ID format"), headers);
                return new ServerCall.Listener<ReqT>() {
                };
            }

            String tenantId = clientIdParts[0];
            String actualClientId = clientIdParts[1];

            // Lookup client configuration
            Optional<ClientSshConfigEntity> clientConfigOpt =
                    clientConfigRepository.findByTenantIdAndClientId(tenantId, actualClientId);

            if (clientConfigOpt.isEmpty()) {
                logger.warn("Client not found: tenant={}, client={}", tenantId, actualClientId);
                call.close(Status.UNAUTHENTICATED.withDescription("Client not found"), headers);
                return new ServerCall.Listener<ReqT>() {
                };
            }

            ClientSshConfigEntity clientConfig = clientConfigOpt.get();

            // Check if client is enabled
            if (!clientConfig.getEnabled()) {
                logger.warn("Client is disabled: tenant={}, client={}", tenantId, actualClientId);
                call.close(Status.UNAUTHENTICATED.withDescription("Client is disabled"), headers);
                return new ServerCall.Listener<ReqT>() {
                };
            }

            // Validate client secret
            if (!validateClientSecret(clientSecret, clientConfig.getClientSecret())) {
                logger.warn("Invalid client secret for: tenant={}, client={}", tenantId, actualClientId);
                call.close(Status.UNAUTHENTICATED.withDescription("Invalid client secret"), headers);
                return new ServerCall.Listener<ReqT>() {
                };
            }

            logger.debug("Client authenticated successfully: tenant={}, client={}", tenantId, actualClientId);

            // Create authenticated context
            Context authenticatedContext = Context.current()
                    .withValue(AUTHENTICATED_CLIENT_KEY, clientConfig);

            // Continue with authenticated context
            return Contexts.interceptCall(authenticatedContext, call, headers, next);

        } catch (Exception e) {
            logger.error("Authentication error", e);
            call.close(Status.UNAUTHENTICATED.withDescription("Authentication error"), headers);
            return new ServerCall.Listener<ReqT>() {
            };
        }
    }

    /**
     * Validate client secret against stored encrypted secret
     */
    private boolean validateClientSecret(String providedSecret, String storedSecret) {
        try {
            // Check if stored secret is base64 encoded (for backward compatibility)
            if (isBase64Encoded(storedSecret)) {
                String decodedSecret = new String(Base64.getDecoder().decode(storedSecret), StandardCharsets.UTF_8);
                return passwordEncoder.matches(providedSecret, decodedSecret);
            } else {
                // Assume stored secret is already BCrypt hashed
                return passwordEncoder.matches(providedSecret, storedSecret);
            }
        } catch (Exception e) {
            logger.error("Error validating client secret", e);
            return false;
        }
    }

    /**
     * Check if a string is base64 encoded
     */
    private boolean isBase64Encoded(String str) {
        try {
            Base64.getDecoder().decode(str);
            return true;
        } catch (IllegalArgumentException e) {
            return false;
        }
    }

}
