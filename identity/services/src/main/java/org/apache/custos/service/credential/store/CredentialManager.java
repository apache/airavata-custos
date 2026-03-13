/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.service.credential.store;

import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;
import org.apache.custos.core.model.credential.store.CredentialEntity;
import org.apache.custos.core.repo.credential.store.CredentialRepository;
import org.apache.custos.service.exceptions.credential.store.CredentialGenerationException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.security.SecureRandom;
import java.util.Base64;
import java.util.List;
import java.util.Map;
import java.util.Random;

/**
 * The CredentialManager class provides methods for generating and decoding credentials.
 */
@Component
public class CredentialManager {

    private static final Logger LOGGER = LoggerFactory.getLogger(CredentialManager.class);
    private static final Random RANDOM = new SecureRandom();
    private static final String ALPHABET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
    private static final int ID_LENGTH = 20;
    private static final int SECRET_LENGTH = 40;
    @Value("${custos.credential.prefix:custos-}")
    private String credentialPrefix;

    @Value("${iam.server.url}")
    private String iamServerURL;

    @Autowired
    private CredentialRepository repository;

    public static Credential decodeToken(String token) {
        try {
            byte[] array = Base64.getDecoder().decode(token);
            if (array != null && array.length > 0) {
                String decodeString = new String(array);
                String[] idSecretPair = decodeString.split(":");
                if (idSecretPair.length == 2) {
                    Credential credential = new Credential();
                    credential.setId(idSecretPair[0]);
                    credential.setSecret(idSecretPair[1]);
                    return credential;
                }
            }
        } catch (Exception ex) {
            LOGGER.error("Error occurred while decoding token");
            throw new CredentialGenerationException("Error occurred while decoding token", ex);
        }
        return null;
    }

    public Credential generateCredential(long ownerId, CredentialTypes type, long validTime) {
        try {
            String clientId = credentialPrefix + generateRandomClientId().toLowerCase() + "-" + ownerId;
            String secret = generateSecret();
            Credential credential = new Credential();

            CredentialEntity entity = new CredentialEntity();
            entity.setClientId(clientId);
            entity.setClientSecretExpiredAt(0);
            entity.setOwnerId(ownerId);
            entity.setType(type.name());

            credential.setId(clientId);
            credential.setSecret(secret);

            repository.save(entity);

            return credential;

        } catch (Exception exception) {
            LOGGER.error("Error occurred while generating credentials for {}", ownerId);
            throw new CredentialGenerationException("Error occurred while generating credentials for " + ownerId, exception);
        }
    }

    /**
     * Decodes a Keycloak-issued JWT access token and extracts relevant claims into a Credential object.
     *
     * Uses nimbus-jose-jwt to properly parse the JWT structure rather than manual Base64 splitting,
     * which prevents malformed token acceptance. The issuer claim is validated to ensure the token
     * was issued by the expected Keycloak instance.
     *
     * TODO: Implement full cryptographic signature verification using RemoteJWKSet pointed at
     *   {iamServerURL}/realms/{realm}/protocol/openid-connect/certs. The realm can be extracted
     *   from the "iss" claim. Use com.nimbusds.jose.jwk.source.RemoteJWKSet with
     *   com.nimbusds.jose.proc.DefaultJWTProcessor for production-grade verification.
     *   A Caffeine cache should wrap the JWK set retrieval to avoid per-request network calls.
     *
     * @param token a signed JWT access token string
     * @return a Credential populated from the token's claims
     */
    public Credential decodeJWTToken(String token) {
        try {
            // Parse using nimbus-jose-jwt — this validates JWT structure and prevents
            // malformed tokens from being processed silently.
            SignedJWT signedJWT = SignedJWT.parse(token);
            JWTClaimsSet claims = signedJWT.getJWTClaimsSet();

            // Validate the issuer to ensure the token originates from our Keycloak instance.
            // The issuer in Keycloak tokens takes the form: {serverUrl}/realms/{realmId}
            String issuer = claims.getIssuer();
            if (issuer == null || !issuer.startsWith(iamServerURL)) {
                throw new CredentialGenerationException("JWT issuer " + issuer
                        + " does not match expected IAM server URL " + iamServerURL);
            }

            Credential credential = new Credential();

            Object azp = claims.getClaim("azp");
            if (azp == null) {
                throw new CredentialGenerationException("JWT is missing required claim 'azp'");
            }
            credential.setId(azp.toString());

            Object email = claims.getClaim("email");
            if (email != null) {
                credential.setEmail(email.toString());
            }

            Object preferredUsername = claims.getClaim("preferred_username");
            if (preferredUsername != null) {
                credential.setUsername(preferredUsername.toString());
            }

            // Check realm_access.roles for the "admin" role
            Object realmAccessObj = claims.getClaim("realm_access");
            if (realmAccessObj instanceof Map<?, ?> realmAccessMap) {
                Object rolesObj = realmAccessMap.get("roles");
                if (rolesObj instanceof List<?> roles) {
                    for (Object role : roles) {
                        if ("admin".equals(role)) {
                            credential.setAdmin(true);
                            break;
                        }
                    }
                }
            }

            return credential;

        } catch (CredentialGenerationException ex) {
            LOGGER.error("JWT validation failed: {}", ex.getMessage());
            throw ex;
        } catch (Exception ex) {
            LOGGER.error("Error occurred while decoding JWT token");
            throw new CredentialGenerationException("Error occurred while decoding token ", ex);
        }
    }

    public Credential decodeAgentJWTToken(String token) {
        try {
            SignedJWT signedJWT = SignedJWT.parse(token);
            JWTClaimsSet claims = signedJWT.getJWTClaimsSet();

            Credential credential = new Credential();
            Object agentId = claims.getClaim("agent-id");
            if (agentId == null) {
                throw new CredentialGenerationException("Agent JWT is missing required claim 'agent-id'");
            }
            credential.setId(agentId.toString());

            Object agentParentId = claims.getClaim("agent-parent-id");
            if (agentParentId != null) {
                credential.setParentId(agentParentId.toString());
            }
            return credential;
        } catch (CredentialGenerationException ex) {
            LOGGER.error("Agent JWT validation failed: {}", ex.getMessage());
            throw ex;
        } catch (Exception ex) {
            LOGGER.error("Error occurred while decoding agent token");
            throw new CredentialGenerationException("Error occurred while decoding token", ex);
        }
    }

    private String generateRandomClientId() {
        StringBuilder returnValue = new StringBuilder(ID_LENGTH);
        for (int i = 0; i < ID_LENGTH; i++) {
            returnValue.append(ALPHABET.charAt(RANDOM.nextInt(ALPHABET.length())));
        }
        return new String(returnValue);
    }

    private String generateSecret() {
        StringBuilder returnValue = new StringBuilder(SECRET_LENGTH);
        for (int i = 0; i < SECRET_LENGTH; i++) {
            returnValue.append(ALPHABET.charAt(RANDOM.nextInt(ALPHABET.length())));
        }
        return new String(returnValue);
    }
}
