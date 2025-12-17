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

import com.nimbusds.jose.JOSEException;
import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.jwk.source.JWKSource;
import com.nimbusds.jose.proc.BadJOSEException;
import com.nimbusds.jose.proc.JWSVerificationKeySelector;
import com.nimbusds.jose.proc.SecurityContext;
import com.nimbusds.jwt.JWT;
import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.JWTParser;
import com.nimbusds.jwt.SignedJWT;
import com.nimbusds.jwt.proc.ConfigurableJWTProcessor;
import com.nimbusds.jwt.proc.DefaultJWTProcessor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.text.ParseException;
import java.time.Instant;
import java.util.Date;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;


/**
 * Service for validating OIDC tokens with full JWKS signature verification.
 */
@Service
public class OidcTokenValidator {

    private static final Logger LOGGER = LoggerFactory.getLogger(OidcTokenValidator.class);

    @Autowired
    private JwksResolver jwksResolver;

    @Autowired
    private OidcProviderConfigResolver providerConfigResolver;

    @Value("${signer.auth.token-validation.enabled:true}")
    private boolean tokenValidationEnabled;

    /**
     * Validate access token with signature verification.
     */
    public UserIdentity validateAccessToken(String accessToken) {
        return validateToken(accessToken, false);
    }

    /**
     * Internal method to validate token with optional ID token specific checks.
     */
    private UserIdentity validateToken(String token, boolean isIdToken) {
        if (!tokenValidationEnabled) {
            LOGGER.debug("Token validation is disabled, returning default identity");
            return new UserIdentity("default-user", Map.of("sub", "default-user"));
        }

        try {
            JWT jwt = JWTParser.parse(token);

            if (!(jwt instanceof SignedJWT signedJWT)) {
                throw new TokenValidationException("Token is not a signed JWT");
            }

            JWTClaimsSet claimsSet = signedJWT.getJWTClaimsSet();

            String issuer = claimsSet.getIssuer();
            if (issuer == null || issuer.trim().isEmpty()) {
                throw new TokenValidationException("Token missing issuer claim");
            }

            // Find provider config for this issuer
            OidcProviderConfig providerConfig = providerConfigResolver.resolveProviderConfig(issuer);
            if (providerConfig == null) {
                throw new TokenValidationException("No OIDC provider configured for issuer: " + issuer);
            }

            // Validate token signature using JWKS
            validateTokenSignature(signedJWT, providerConfig);

            // Validate expiry
            Date exp = claimsSet.getExpirationTime();
            if (exp != null && exp.before(Date.from(Instant.now()))) {
                throw new TokenValidationException("Token has expired");
            }

            // Validate not-before (if present)
            Date nbf = claimsSet.getNotBeforeTime();
            if (nbf != null && nbf.after(Date.from(Instant.now()))) {
                throw new TokenValidationException("Token not yet valid (nbf claim)");
            }

            // ID token specific validations
            if (isIdToken) {
                validateIdTokenClaims(claimsSet, providerConfig);
            }

            Map<String, Object> claims = claimsSet.getClaims();

            // Extract principal from claims
            String principal = extractPrincipal(claims);
            if (principal == null || principal.trim().isEmpty()) {
                throw new TokenValidationException("No valid principal found in token claims");
            }

            LOGGER.debug("Token validation successful for principal: {}, issuer: {}", principal, issuer);
            return new UserIdentity(principal, claims);

        } catch (ParseException e) {
            LOGGER.error("Failed to parse JWT token", e);
            throw new TokenValidationException("Invalid JWT token format: " + e.getMessage());

        } catch (JOSEException | BadJOSEException e) {
            LOGGER.error("JWT signature validation failed", e);
            throw new TokenValidationException("Token signature validation failed: " + e.getMessage());

        } catch (Exception e) {
            LOGGER.error("Token validation failed", e);
            throw new TokenValidationException("Token validation failed: " + e.getMessage());
        }
    }

    /**
     * Validate token signature using JWKS.
     */
    private void validateTokenSignature(SignedJWT signedJWT, OidcProviderConfig providerConfig)
            throws JOSEException, BadJOSEException {
        try {
            // Get JWKSource for this provider
            JWKSource<SecurityContext> jwkSource = jwksResolver.getJwkSource(providerConfig);

            // Create JWT processor
            ConfigurableJWTProcessor<SecurityContext> jwtProcessor = new DefaultJWTProcessor<>();

            // Set JWS key selector with supported algorithms
            // Use JWSVerificationKeySelector with common JWS algorithms
            Set<JWSAlgorithm> expectedJWSAlgs = new HashSet<>();
            expectedJWSAlgs.add(JWSAlgorithm.RS256);
            expectedJWSAlgs.add(JWSAlgorithm.RS384);
            expectedJWSAlgs.add(JWSAlgorithm.RS512);
            expectedJWSAlgs.add(JWSAlgorithm.ES256);
            expectedJWSAlgs.add(JWSAlgorithm.ES384);
            expectedJWSAlgs.add(JWSAlgorithm.ES512);
            expectedJWSAlgs.add(JWSAlgorithm.EdDSA);
            expectedJWSAlgs.add(JWSAlgorithm.PS256);
            expectedJWSAlgs.add(JWSAlgorithm.PS384);
            expectedJWSAlgs.add(JWSAlgorithm.PS512);

            JWSVerificationKeySelector<SecurityContext> keySelector = new JWSVerificationKeySelector<>(expectedJWSAlgs, jwkSource);
            jwtProcessor.setJWSKeySelector(keySelector);

            // Process and validate the token
            jwtProcessor.process(signedJWT, null);

            LOGGER.debug("Token signature validated successfully for issuer: {}", providerConfig.getIssuer());

        } catch (JOSEException | BadJOSEException e) {
            LOGGER.error("Token signature validation failed for issuer: {}", providerConfig.getIssuer(), e);
            throw e;
        }
    }

    /**
     * Validate ID token specific claims (aud, azp, etc.)
     */
    private void validateIdTokenClaims(JWTClaimsSet claimsSet, OidcProviderConfig providerConfig) {
        // Validate audience if client ID is configured
        if (providerConfig.getClientId() != null && !providerConfig.getClientId().trim().isEmpty()) {
            List<String> audiences = claimsSet.getAudience();
            if (audiences == null || audiences.isEmpty()) {
                throw new TokenValidationException("ID token missing audience claim");
            }
            if (!audiences.contains(providerConfig.getClientId())) {
                throw new TokenValidationException("ID token audience does not match configured client ID. Expected: " + providerConfig.getClientId() + ", Found: " + audiences);
            }
        }

        // Validate that token is an ID token (has 'sub' claim)
        if (claimsSet.getSubject() == null || claimsSet.getSubject().trim().isEmpty()) {
            throw new TokenValidationException("ID token missing subject (sub) claim");
        }
    }

    /**
     * Extract principal from token claims.
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
     * Hash token for audit logging (privacy protection)
     */
    public String hashTokenForAudit(String token) {
        try {
            // Use SHA-256 hash of token for audit correlation
            java.security.MessageDigest digest = java.security.MessageDigest.getInstance("SHA-256");
            byte[] hash = digest.digest(token.getBytes(java.nio.charset.StandardCharsets.UTF_8));
            return java.util.Base64.getEncoder().encodeToString(hash);

        } catch (Exception e) {
            LOGGER.warn("Failed to hash token for audit", e);
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
