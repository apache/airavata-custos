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

import com.nimbusds.jose.JOSEException;
import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.JWSHeader;
import com.nimbusds.jose.JWSSigner;
import com.nimbusds.jose.JWSVerifier;
import com.nimbusds.jose.crypto.MACSigner;
import com.nimbusds.jose.crypto.MACVerifier;
import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;

import java.time.Instant;
import java.util.Date;

/**
 * Utility class for generating test JWT tokens for integration tests.
 */
public class TestJwtTokenGenerator {

    // Shared secret for signing (for testing only)
    private static final String TEST_SECRET = "test-secret-key-for-jwt-signing-in-integration-tests-only";

    /**
     * Generate a test JWT token with the specified principal.
     *
     * @param principal The principal (username) to include in the token
     * @return A signed JWT token string
     */
    public static String generateTestToken(String principal) {
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
}

