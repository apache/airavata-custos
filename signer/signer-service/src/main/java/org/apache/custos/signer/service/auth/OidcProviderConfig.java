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

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.Objects;

/**
 * Configuration for an OIDC provider.
 * This class holds provider-specific configuration that can be loaded from
 * application.yml (current) or from tenant-specific storage (future) TODO.
 */
@Component
@ConfigurationProperties(prefix = "signer.auth.oidc.provider")
public class OidcProviderConfig {
    private String issuer;
    private String jwksUri;
    private String clientId;
    private int timeoutSeconds = 10;
    private boolean verifySsl = true;

    public OidcProviderConfig() {
    }

    public OidcProviderConfig(String issuer, String jwksUri, String clientId, int timeoutSeconds, boolean verifySsl) {
        this.issuer = Objects.requireNonNull(issuer, "Issuer cannot be null");
        this.jwksUri = Objects.requireNonNull(jwksUri, "JWKS URI cannot be null");
        this.clientId = clientId;
        this.timeoutSeconds = timeoutSeconds > 0 ? timeoutSeconds : 10;
        this.verifySsl = verifySsl;
    }

    public String getIssuer() {
        return issuer;
    }

    public void setIssuer(String issuer) {
        this.issuer = issuer;
    }

    public String getJwksUri() {
        return jwksUri;
    }

    public void setJwksUri(String jwksUri) {
        this.jwksUri = jwksUri;
    }

    public String getClientId() {
        return clientId;
    }

    public void setClientId(String clientId) {
        this.clientId = clientId;
    }

    public int getTimeoutSeconds() {
        return timeoutSeconds;
    }

    public void setTimeoutSeconds(int timeoutSeconds) {
        this.timeoutSeconds = timeoutSeconds > 0 ? timeoutSeconds : 10;
    }

    public boolean isVerifySsl() {
        return verifySsl;
    }

    public void setVerifySsl(boolean verifySsl) {
        this.verifySsl = verifySsl;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        OidcProviderConfig that = (OidcProviderConfig) o;
        return Objects.equals(issuer, that.issuer);
    }

    @Override
    public int hashCode() {
        return Objects.hash(issuer);
    }

    @Override
    public String toString() {
        return "OidcProviderConfig{" +
                "issuer='" + issuer + '\'' +
                ", jwksUri='" + jwksUri + '\'' +
                ", clientId='" + clientId + '\'' +
                ", timeoutSeconds=" + timeoutSeconds +
                ", verifySsl=" + verifySsl +
                '}';
    }
}
