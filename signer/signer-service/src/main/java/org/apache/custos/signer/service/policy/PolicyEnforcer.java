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
package org.apache.custos.signer.service.policy;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.signer.service.model.ClientSshConfigEntity;
import org.apache.custos.signer.v1.SignRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.Arrays;
import java.util.List;
import java.util.Map;

/**
 * Service for enforcing security policies against certificate signing requests.
 * Validates requests against client-specific policies loaded from database.
 */
@Service
public class PolicyEnforcer {

    private static final Logger logger = LoggerFactory.getLogger(PolicyEnforcer.class);

    @Autowired
    private ObjectMapper objectMapper;

    @Value("${signer.policy.defaults.max-ttl-seconds:86400}")
    private int defaultMaxTtlSeconds;

    @Value("${signer.policy.defaults.allowed-key-types:ed25519,rsa,ecdsa}")
    private String defaultAllowedKeyTypes;

    /**
     * Enforce policy against a signing request
     */
    public void enforcePolicy(SignRequest request, ClientSshConfigEntity clientConfig) {
        logger.debug("Enforcing policy for tenant: {}, client: {}, principal: {}",
                request.getTenantId(), request.getClientId(), request.getPrincipal());

        // Validate TTL
        enforceTtlPolicy(request, clientConfig);

        // Validate key type
        enforceKeyTypePolicy(request, clientConfig);

        // Validate source address restriction
        enforceSourceAddressPolicy(request, clientConfig);

        // Validate critical options
        enforceCriticalOptionsPolicy(request, clientConfig);

        logger.debug("Policy enforcement passed for tenant: {}, client: {}",
                request.getTenantId(), request.getClientId());
    }

    /**
     * Enforce TTL policy
     */
    private void enforceTtlPolicy(SignRequest request, ClientSshConfigEntity clientConfig) {
        int maxTtl = clientConfig.getMaxTtlSeconds() != null ?
                clientConfig.getMaxTtlSeconds() : defaultMaxTtlSeconds;

        if (request.getTtlSeconds() > maxTtl) {
            throw new PolicyViolationException(
                    String.format("Requested TTL %d seconds exceeds maximum allowed TTL %d seconds",
                            request.getTtlSeconds(), maxTtl));
        }

        if (request.getTtlSeconds() <= 0) {
            throw new PolicyViolationException("TTL must be positive");
        }

        logger.debug("TTL policy check passed: requested={}s, max={}s",
                request.getTtlSeconds(), maxTtl);
    }

    /**
     * Enforce key type policy
     */
    private void enforceKeyTypePolicy(SignRequest request, ClientSshConfigEntity clientConfig) {
        List<String> allowedKeyTypes = getAllowedKeyTypes(clientConfig);

        // Parse the public key to determine its type
        String keyType = extractKeyTypeFromPublicKey(request.getPublicKey().toByteArray());

        if (!allowedKeyTypes.contains(keyType.toLowerCase())) {
            throw new PolicyViolationException(
                    String.format("Key type '%s' is not allowed. Allowed types: %s",
                            keyType, allowedKeyTypes));
        }

        logger.debug("Key type policy check passed: type={}, allowed={}", keyType, allowedKeyTypes);
    }

    /**
     * Enforce source address restriction policy
     */
    private void enforceSourceAddressPolicy(SignRequest request, ClientSshConfigEntity clientConfig) {
        String sourceAddressRestriction = clientConfig.getSourceAddressRestriction();

        if (sourceAddressRestriction != null && !sourceAddressRestriction.trim().isEmpty()) {
            // TODO: Implement source address validation
            // This would require extracting the client IP from the gRPC context
            // and validating it against the CIDR or IP restriction
            logger.debug("Source address restriction configured: {}", sourceAddressRestriction);
            // For now, just log the restriction - full implementation would validate client IP
        }
    }

    /**
     * Enforce critical options policy
     */
    private void enforceCriticalOptionsPolicy(SignRequest request, ClientSshConfigEntity clientConfig) {
        String criticalOptionsJson = clientConfig.getCriticalOptions();

        if (criticalOptionsJson != null && !criticalOptionsJson.trim().isEmpty()) {
            try {
                Map<String, String> allowedCriticalOptions = objectMapper.readValue(
                        criticalOptionsJson, new TypeReference<Map<String, String>>() {
                        });

                // TODO: Validate requested critical options against allowed ones
                // This would require parsing the public key to extract requested critical options
                logger.debug("Critical options policy configured: {}", allowedCriticalOptions);

            } catch (Exception e) {
                logger.warn("Failed to parse critical options policy for client: {}",
                        clientConfig.getClientId(), e);
            }
        }
    }

    /**
     * Get allowed key types from client config or defaults
     */
    private List<String> getAllowedKeyTypes(ClientSshConfigEntity clientConfig) {
        String allowedKeyTypesJson = clientConfig.getAllowedKeyTypes();

        if (allowedKeyTypesJson != null && !allowedKeyTypesJson.trim().isEmpty()) {
            try {
                return objectMapper.readValue(allowedKeyTypesJson, new TypeReference<List<String>>() {
                });
            } catch (Exception e) {
                logger.warn("Failed to parse allowed key types for client: {}, using defaults",
                        clientConfig.getClientId(), e);
            }
        }

        // Fall back to defaults
        return Arrays.asList(defaultAllowedKeyTypes.split(","));
    }

    /**
     * Extract key type from SSH public key bytes
     */
    private String extractKeyTypeFromPublicKey(byte[] publicKeyBytes) {
        try {
            String publicKeyString = new String(publicKeyBytes, java.nio.charset.StandardCharsets.UTF_8).trim();
            String[] parts = publicKeyString.split("\\s+");

            if (parts.length >= 1) {
                return KeyType.from(parts[0]).id();
            }

            throw new IllegalArgumentException("Invalid SSH public key format");

        } catch (Exception e) {
            logger.error("Failed to extract key type from public key", e);
            throw new PolicyViolationException("Invalid SSH public key format");
        }
    }

    /**
     * Exception thrown when policy validation fails
     */
    public static class PolicyViolationException extends RuntimeException {
        public PolicyViolationException(String message) {
            super(message);
        }

        public PolicyViolationException(String message, Throwable cause) {
            super(message, cause);
        }
    }
}
