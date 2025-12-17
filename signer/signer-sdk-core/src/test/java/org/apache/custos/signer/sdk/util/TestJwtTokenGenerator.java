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
package org.apache.custos.signer.sdk.util;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.nimbusds.jose.JOSEException;
import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.JWSHeader;
import com.nimbusds.jose.JWSSigner;
import com.nimbusds.jose.JWSVerifier;
import com.nimbusds.jose.crypto.MACSigner;
import com.nimbusds.jose.crypto.MACVerifier;
import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;
import okhttp3.FormBody;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.io.InputStream;
import java.time.Instant;
import java.util.Date;
import java.util.Properties;
import java.util.concurrent.TimeUnit;

/**
 * Utility class for fetching real OIDC tokens or generating test JWT tokens for integration tests.
 */
public class TestJwtTokenGenerator {

    private static final Logger LOGGER = LoggerFactory.getLogger(TestJwtTokenGenerator.class);

    // Test shared secret for signing
    private static final String TEST_SECRET = "test-secret-key-for-jwt-signing-in-integration-tests-only";

    private static OidcTestConfig oidcConfig = null;
    private static final Object configLock = new Object();

    private static final OkHttpClient httpClient = new OkHttpClient.Builder()
            .connectTimeout(10, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .build();

    /**
     * If OIDC test configuration is available, fetches a real token from OIDC provider.
     * Otherwise, generates a test token.
     *
     * @param principal The principal (username) to include in the token
     * @return A signed JWT token string
     */
    public static String generateTestToken(String principal) {
        OidcTestConfig config = getOidcTestConfig();

        if (config != null && config.isValid()) {
            LOGGER.info("OIDC test configuration found, fetching real token from: {}", config.tokenUrl);
            try {
                return fetchOidcToken(config, principal);
            } catch (Exception e) {
                LOGGER.warn("Failed to fetch OIDC token, falling back to test token generation: {}", e.getMessage());
                // Proceed to test token generation
            }
        }

        // Fallback to test token generation
        LOGGER.debug("Using test token generation (no OIDC config or fetch failed)");
        return generateTestToken(principal, "custos-test-issuer", 3600);
    }

    /**
     * Generate a test JWT token with custom claims.
     *
     * @param principal  The principal (username) to include in the token
     * @param issuer     The issuer to include in the token
     * @param ttlSeconds Token TTL in seconds
     * @return A signed JWT token string
     */
    public static String generateTestToken(String principal, String issuer, int ttlSeconds) {
        try {
            Instant now = Instant.now();
            Instant expiry = now.plusSeconds(ttlSeconds);

            JWTClaimsSet claimsSet = new JWTClaimsSet.Builder()
                    .subject(principal)
                    .issuer(issuer)
                    .claim("preferred_username", principal)
                    .claim("email", principal + "@test.example.com")
                    .claim("sub", principal)
                    .issueTime(Date.from(now))
                    .expirationTime(Date.from(expiry))
                    .jwtID(java.util.UUID.randomUUID().toString())
                    .build();

            JWSHeader header = new JWSHeader(JWSAlgorithm.HS256);
            SignedJWT signedJWT = new SignedJWT(header, claimsSet);

            JWSSigner signer = new MACSigner(TEST_SECRET);
            signedJWT.sign(signer);

            return signedJWT.serialize();

        } catch (JOSEException e) {
            throw new RuntimeException("Failed to generate test JWT token", e);
        }
    }

    /**
     * Fetch a real OIDC token using Resource Owner Password Credentials grant.
     */
    private static String fetchOidcToken(OidcTestConfig config, String principal) throws IOException {
        RequestBody formBody = new FormBody.Builder()
                .add("grant_type", "password")
                .add("client_id", config.clientId)
                .add("username", config.username)
                .add("password", config.password)
                .add("scope", "openid profile email")
                .build();

        Request request = new Request.Builder()
                .url(config.tokenUrl)
                .post(formBody)
                .header("Content-Type", "application/x-www-form-urlencoded")
                .build();

        try (Response response = httpClient.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                String errorBody = response.body() != null ? response.body().string() : "No error body";
                throw new IOException("OIDC token request failed: HTTP " + response.code() + " - " + errorBody);
            }

            String responseBody = response.body() != null ? response.body().string() : null;
            if (responseBody == null) {
                throw new IOException("Empty response from OIDC token endpoint");
            }

            ObjectMapper mapper = new ObjectMapper();
            JsonNode json = mapper.readTree(responseBody);

            String token = json.path("access_token").asText(null);
            if (token == null || token.isEmpty()) {
                token = json.path("id_token").asText(null);
            }

            if (token == null || token.isEmpty()) {
                throw new IOException("No token found in OIDC response");
            }

            LOGGER.info("Successfully fetched OIDC token for user: {}", config.username);
            return token;

        } catch (Exception e) {
            LOGGER.error("Error fetching OIDC token", e);
            throw new IOException("Failed to fetch OIDC token: " + e.getMessage(), e);
        }
    }

    /**
     * Load OIDC test configuration from test-oidc.properties.
     */
    private static OidcTestConfig getOidcTestConfig() {
        if (oidcConfig != null) {
            return oidcConfig;
        }

        synchronized (configLock) {
            if (oidcConfig != null) {
                return oidcConfig;
            }

            try {
                InputStream is = TestJwtTokenGenerator.class.getClassLoader().getResourceAsStream("test-oidc.properties");

                if (is == null) {
                    LOGGER.debug("test-oidc.properties not found, will use test token generation");
                    oidcConfig = new OidcTestConfig();
                    return oidcConfig;
                }

                Properties props = new Properties();
                props.load(is);
                is.close();

                String tokenUrl = props.getProperty("test.oidc.token.url");
                String clientId = props.getProperty("test.oidc.client.id");
                String username = props.getProperty("test.oidc.username");
                String password = props.getProperty("test.oidc.password");
                String issuer = props.getProperty("test.oidc.issuer");

                if (tokenUrl != null && !tokenUrl.trim().isEmpty() &&
                        clientId != null && !clientId.trim().isEmpty() &&
                        username != null && !username.trim().isEmpty() &&
                        password != null && !password.trim().isEmpty()) {

                    oidcConfig = new OidcTestConfig(tokenUrl, clientId, username, password, issuer);
                    LOGGER.info("Loaded OIDC test configuration: tokenUrl={}, clientId={}, username={}", tokenUrl, clientId, username);
                } else {
                    LOGGER.debug("OIDC test configuration incomplete, will use test token generation");
                    oidcConfig = new OidcTestConfig();
                }

            } catch (Exception e) {
                LOGGER.warn("Failed to load test-oidc.properties: {}", e.getMessage());
                oidcConfig = new OidcTestConfig();
            }

            return oidcConfig;
        }
    }

    /**
     * Verify a test JWT token (for testing the generator itself).
     *
     * @param token The JWT token to verify
     * @return true if the token is valid
     */
    public static boolean verifyTestToken(String token) {
        try {
            SignedJWT signedJWT = SignedJWT.parse(token);
            JWSVerifier verifier = new MACVerifier(TEST_SECRET);
            return signedJWT.verify(verifier);
        } catch (Exception e) {
            return false;
        }
    }

    /**
     * Container for OIDC test configuration.
     */
    private static class OidcTestConfig {
        final String tokenUrl;
        final String clientId;
        final String username;
        final String password;
        final String issuer;

        OidcTestConfig() {
            this.tokenUrl = null;
            this.clientId = null;
            this.username = null;
            this.password = null;
            this.issuer = null;
        }

        OidcTestConfig(String tokenUrl, String clientId, String username, String password, String issuer) {
            this.tokenUrl = tokenUrl;
            this.clientId = clientId;
            this.username = username;
            this.password = password;
            this.issuer = issuer;
        }

        boolean isValid() {
            return tokenUrl != null && !tokenUrl.trim().isEmpty() &&
                    clientId != null && !clientId.trim().isEmpty() &&
                    username != null && !username.trim().isEmpty() &&
                    password != null && !password.trim().isEmpty();
        }
    }
}
