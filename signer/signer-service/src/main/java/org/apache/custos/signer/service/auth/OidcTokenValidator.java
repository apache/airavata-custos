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

import com.nimbusds.jwt.JWT;
import com.nimbusds.jwt.JWTParser;
import com.nimbusds.jwt.SignedJWT;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.text.ParseException;
import java.time.Instant;
import java.util.Date;
import java.util.Map;

/**
 * Service for validating OIDC tokens (basic implementation for v1).
 * TODO: Integrate with Custos identity service for full validation.
 */
@Service
public class OidcTokenValidator {

    private static final Logger logger = LoggerFactory.getLogger(OidcTokenValidator.class);

    @Value("${signer.auth.allowed-issuers:}")
    private String allowedIssuers;

    @Value("${signer.auth.token-validation.enabled:true}")
    private boolean tokenValidationEnabled;

    /**
     * Validate OIDC token and extract user identity
     */
    public UserIdentity validateToken(String token) {
        if (!tokenValidationEnabled) {
            logger.debug("Token validation is disabled, returning default identity");
            return new UserIdentity("default-user", Map.of("sub", "default-user"));
        }

        try {
            // Parse JWT token
            JWT jwt = JWTParser.parse(token);

            if (!(jwt instanceof SignedJWT)) {
                throw new TokenValidationException("Token is not a signed JWT");
            }

            SignedJWT signedJWT = (SignedJWT) jwt;

            // Extract claims
            Map<String, Object> claims = signedJWT.getJWTClaimsSet().getClaims();

            // Validate expiry
            Date exp = signedJWT.getJWTClaimsSet().getExpirationTime();
            if (exp != null && exp.before(Date.from(Instant.now()))) {
                throw new TokenValidationException("Token has expired");
            }

            // Validate issuer (if configured)
            if (allowedIssuers != null && !allowedIssuers.trim().isEmpty()) {
                String iss = signedJWT.getJWTClaimsSet().getIssuer();
                if (iss == null || !isAllowedIssuer(iss)) {
                    throw new TokenValidationException("Token issuer not allowed: " + iss);
                }
            }

            // Extract principal from claims (prefer sub, then preferred_username, then email)
            String principal = extractPrincipal(claims);
            if (principal == null || principal.trim().isEmpty()) {
                throw new TokenValidationException("No valid principal found in token claims");
            }

            logger.debug("Token validation successful for principal: {}", principal);
            return new UserIdentity(principal, claims);

        } catch (ParseException e) {
            logger.error("Failed to parse JWT token", e);
            throw new TokenValidationException("Invalid JWT token format");
        } catch (Exception e) {
            logger.error("Token validation failed", e);
            throw new TokenValidationException("Token validation failed: " + e.getMessage());
        }
    }

    /**
     * Extract principal from token claims
     */
    private String extractPrincipal(Map<String, Object> claims) {
        // Try 'sub' first (standard OIDC subject identifier)
        String principal = (String) claims.get("sub");
        if (principal != null && !principal.trim().isEmpty()) {
            return principal;
        }

        // Try 'preferred_username' (common in Keycloak)
        principal = (String) claims.get("preferred_username");
        if (principal != null && !principal.trim().isEmpty()) {
            return principal;
        }

        // Try 'email' as fallback
        principal = (String) claims.get("email");
        if (principal != null && !principal.trim().isEmpty()) {
            return principal;
        }

        // Try 'username' as last resort
        principal = (String) claims.get("username");
        if (principal != null && !principal.trim().isEmpty()) {
            return principal;
        }

        return null;
    }

    /**
     * Check if issuer is in allowed list
     */
    private boolean isAllowedIssuer(String issuer) {
        if (allowedIssuers == null || allowedIssuers.trim().isEmpty()) {
            return true; // No restriction configured
        }

        String[] allowedList = allowedIssuers.split(",");
        for (String allowed : allowedList) {
            if (issuer.trim().equals(allowed.trim())) {
                return true;
            }
        }

        return false;
    }

    /**
     * Hash token for audit logging (privacy protection)
     */
    public String hashTokenForAudit(String token) {
        try {
            // Use SHA-256 hash of token for audit correlation
            java.security.MessageDigest digest = java.security.MessageDigest.getInstance("SHA-256");
            byte[] hash = digest.digest(token.getBytes(java.nio.charset.StandardCharsets.UTF_8));
            return java.util.Base64.getEncoder().encodeToString(hash);
        } catch (Exception e) {
            logger.warn("Failed to hash token for audit", e);
            return "hash-failed";
        }
    }

    /**
     * User identity container
     */
    public static class UserIdentity {
        private final String principal;
        private final Map<String, Object> claims;

        public UserIdentity(String principal, Map<String, Object> claims) {
            this.principal = principal;
            this.claims = claims;
        }

        public String getPrincipal() {
            return principal;
        }

        public Map<String, Object> getClaims() {
            return claims;
        }

        public String getClaim(String name) {
            Object value = claims.get(name);
            return value != null ? value.toString() : null;
        }

        @Override
        public String toString() {
            return "UserIdentity{" +
                    "principal='" + principal + '\'' +
                    ", claims=" + claims.keySet() +
                    '}';
        }
    }

    /**
     * Exception thrown when token validation fails
     */
    public static class TokenValidationException extends RuntimeException {
        public TokenValidationException(String message) {
            super(message);
        }

        public TokenValidationException(String message, Throwable cause) {
            super(message, cause);
        }
    }
}
