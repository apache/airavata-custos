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

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.nimbusds.jose.jwk.JWKSet;
import com.nimbusds.jose.jwk.source.ImmutableJWKSet;
import com.nimbusds.jose.jwk.source.JWKSource;
import com.nimbusds.jose.proc.SecurityContext;
import com.nimbusds.jose.util.DefaultResourceRetriever;
import com.nimbusds.jose.util.Resource;
import com.nimbusds.jose.util.ResourceRetriever;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.stereotype.Component;

import java.net.URL;
import java.util.concurrent.TimeUnit;

import static org.apache.custos.signer.service.config.CacheConfig.JWKS_CACHE;

/**
 * Resolves and caches JWKS for OIDC providers.
 * Provides JWKSource for token signature verification.
 */
@Component
public class JwksResolver {

    private static final Logger LOGGER = LoggerFactory.getLogger(JwksResolver.class);

    private final OkHttpClient httpClient;

    public JwksResolver() {
        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(10, TimeUnit.SECONDS)
                .readTimeout(10, TimeUnit.SECONDS)
                .build();
    }

    /**
     * Get JWKSource for a given issuer.
     * JWKS is cached and keyed by issuer.
     */
    @Cacheable(cacheNames = JWKS_CACHE, key = "#root.args[0].issuer")
    public JWKSource<SecurityContext> getJwkSource(OidcProviderConfig providerConfig) {
        return new ImmutableJWKSet<>(fetchJwkSet(providerConfig));
    }

    /**
     * Discover JWKS URI from OIDC discovery endpoint.
     * Falls back to provider config if discovery fails.
     */
    public String discoverJwksUri(String issuer) {
        try {
            String discoveryUrl = (issuer.endsWith("/") ? issuer : issuer + "/") + ".well-known/openid-configuration";
            LOGGER.debug("Discovering JWKS URI from: {}", discoveryUrl);

            Request request = new Request.Builder()
                    .url(discoveryUrl)
                    .get()
                    .build();

            try (Response response = httpClient.newCall(request).execute()) {
                if (!response.isSuccessful()) {
                    LOGGER.warn("Failed to fetch OIDC discovery document: HTTP {}", response.code());
                    return null;
                }

                String body = response.body() != null ? response.body().string() : null;
                if (body == null) {
                    return null;
                }

                ObjectMapper mapper = new ObjectMapper();
                JsonNode json = mapper.readTree(body);

                String jwksUri = json.path("jwks_uri").asText(null);
                if (jwksUri != null && !jwksUri.isEmpty()) {
                    LOGGER.debug("Discovered JWKS URI: {}", jwksUri);
                    return jwksUri;
                }

                LOGGER.warn("JWKS URI not found in discovery document");
                return null;
            }

        } catch (Exception e) {
            LOGGER.warn("Failed to discover JWKS URI from issuer: {}", issuer, e);
            return null;
        }
    }

    private JWKSet fetchJwkSet(OidcProviderConfig providerConfig) {
        try {
            URL jwksUrl = new URL(providerConfig.getJwksUri());

            ResourceRetriever resourceRetriever = new DefaultResourceRetriever(providerConfig.getTimeoutSeconds() * 1000, providerConfig.getTimeoutSeconds() * 1000, 64 * 1024);

            LOGGER.debug("Fetching JWKS JSON from: {}", jwksUrl);
            Resource resource = resourceRetriever.retrieveResource(jwksUrl);
            if (resource == null || resource.getContent() == null || resource.getContent().isBlank()) {
                throw new RuntimeException("Empty JWKS response from " + jwksUrl);
            }

            JWKSet jwkSet = JWKSet.parse(resource.getContent());
            LOGGER.info("Successfully fetched and parsed JWKS for issuer: {}", providerConfig.getIssuer());
            return jwkSet;

        } catch (Exception e) {
            LOGGER.error("Failed to fetch/parse JWKS for issuer: {} from: {}", providerConfig.getIssuer(), providerConfig.getJwksUri(), e);
            throw new RuntimeException("Failed to fetch/parse JWKS: " + e.getMessage(), e);
        }
    }
}
