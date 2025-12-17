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
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */
package org.apache.custos.signer.service.auth;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.stereotype.Component;

import static org.apache.custos.signer.service.config.CacheConfig.OIDC_PROVIDER_CONFIG_CACHE;

/**
 * Resolves an {@link OidcProviderConfig} for an issuer using configured providers and OIDC discovery.
 */
@Component
public class OidcProviderConfigResolver {

    private static final Logger LOGGER = LoggerFactory.getLogger(OidcProviderConfigResolver.class);

    private final JwksResolver jwksResolver;
    private final OidcProviderConfig provider;

    public OidcProviderConfigResolver(JwksResolver jwksResolver, OidcProviderConfig provider) {
        this.jwksResolver = jwksResolver;
        this.provider = provider;
    }

    /**
     * Resolve a provider config for the given issuer.
     */
    @Cacheable(cacheNames = OIDC_PROVIDER_CONFIG_CACHE, key = "#root.args[0]", unless = "#result == null")
    public OidcProviderConfig resolveProviderConfig(String issuer) {
        if (issuer == null || issuer.isBlank()) {
            return null;
        }

        // If a provider is configured, only accept tokens from that issuer
        if (provider != null && provider.getIssuer() != null && !provider.getIssuer().isBlank()) {
            String configuredIssuer = provider.getIssuer();
            boolean matches = issuer.equals(configuredIssuer) ||
                    issuer.startsWith(configuredIssuer) ||
                    configuredIssuer.startsWith(issuer);

            if (!matches) {
                return null;
            }

            // If jwks-uri is configured, use it
            if (provider.getJwksUri() != null && !provider.getJwksUri().isBlank()) {
                return provider;
            }

            // Otherwise, discover jwks-uri for this issuer
            LOGGER.debug("OIDC provider configured for issuer '{}', but jwks-uri missing; attempting discovery", issuer);
            String jwksUri = jwksResolver.discoverJwksUri(issuer);
            if (jwksUri == null || jwksUri.isBlank()) {
                return null;
            }

            return new OidcProviderConfig(
                    configuredIssuer,
                    jwksUri,
                    provider.getClientId(),
                    provider.getTimeoutSeconds(),
                    provider.isVerifySsl()
            );
        }

        LOGGER.warn("No OIDC provider configured under signer.auth.oidc.provider; issuer resolution will rely on discovery only.");
        String jwksUri = jwksResolver.discoverJwksUri(issuer);
        if (jwksUri == null || jwksUri.isBlank()) {
            return null;
        }

        OidcProviderConfig discovered = new OidcProviderConfig(issuer, jwksUri, null, 10, true);
        LOGGER.info("Discovered JWKS URI for issuer: {} -> {}", issuer, jwksUri);
        return discovered;
    }

}


