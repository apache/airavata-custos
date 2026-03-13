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

package org.apache.custos.service.management;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;

public class ClientConfigurationOptions {

    private static final Logger LOGGER = LoggerFactory.getLogger(ClientConfigurationOptions.class);

    private String clientName;
    private String tenantUrl;
    private List<String> redirectUris;

    private boolean authorizationCodeEnabled;
    private boolean clientCredentialsEnabled;
    private boolean implicitEnabled;
    private boolean passwordEnabled;
    private boolean deviceCodeEnabled;

    private boolean pkceEnabled;
    private boolean publicClient;

    public void validate() {
        if (tenantUrl == null || tenantUrl.isBlank()) {
            LOGGER.error("Client configuration validation failed: tenantUrl is blank");
            throw new IllegalArgumentException("tenantUrl must not be blank");
        }

        if (!authorizationCodeEnabled && !clientCredentialsEnabled && !implicitEnabled && !passwordEnabled && !deviceCodeEnabled) {
            LOGGER.error("Client configuration validation failed: no grant type enabled for client '{}'", clientName);
            throw new IllegalArgumentException("At least one grant type must be enabled");
        }

        if (publicClient && clientCredentialsEnabled) {
            LOGGER.error("Client configuration validation failed: public client '{}' cannot use client_credentials", clientName);
            throw new IllegalArgumentException("Public clients cannot use client_credentials grant type");
        }

        if ((authorizationCodeEnabled || implicitEnabled) && (redirectUris == null || redirectUris.isEmpty())) {
            LOGGER.error("Client configuration validation failed: redirectUris missing for client '{}' with authorization_code or implicit grant", clientName);
            throw new IllegalArgumentException("redirectUris required when authorization_code or implicit grant is enabled");
        }
    }

    public String getClientName() {
        return clientName;
    }

    public void setClientName(String clientName) {
        this.clientName = clientName;
    }

    public String getTenantUrl() {
        return tenantUrl;
    }

    public void setTenantUrl(String tenantUrl) {
        this.tenantUrl = tenantUrl;
    }

    public List<String> getRedirectUris() {
        return redirectUris;
    }

    public void setRedirectUris(List<String> redirectUris) {
        this.redirectUris = redirectUris;
    }

    public boolean isAuthorizationCodeEnabled() {
        return authorizationCodeEnabled;
    }

    public void setAuthorizationCodeEnabled(boolean authorizationCodeEnabled) {
        this.authorizationCodeEnabled = authorizationCodeEnabled;
    }

    public boolean isClientCredentialsEnabled() {
        return clientCredentialsEnabled;
    }

    public void setClientCredentialsEnabled(boolean clientCredentialsEnabled) {
        this.clientCredentialsEnabled = clientCredentialsEnabled;
    }

    public boolean isImplicitEnabled() {
        return implicitEnabled;
    }

    public void setImplicitEnabled(boolean implicitEnabled) {
        this.implicitEnabled = implicitEnabled;
    }

    public boolean isPasswordEnabled() {
        return passwordEnabled;
    }

    public void setPasswordEnabled(boolean passwordEnabled) {
        this.passwordEnabled = passwordEnabled;
    }

    public boolean isDeviceCodeEnabled() {
        return deviceCodeEnabled;
    }

    public void setDeviceCodeEnabled(boolean deviceCodeEnabled) {
        this.deviceCodeEnabled = deviceCodeEnabled;
    }

    public boolean isPkceEnabled() {
        return pkceEnabled;
    }

    public void setPkceEnabled(boolean pkceEnabled) {
        this.pkceEnabled = pkceEnabled;
    }

    public boolean isPublicClient() {
        return publicClient;
    }

    public void setPublicClient(boolean publicClient) {
        this.publicClient = publicClient;
    }
}
