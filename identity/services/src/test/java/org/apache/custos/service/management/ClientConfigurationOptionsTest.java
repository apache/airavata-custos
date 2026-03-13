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

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatNoException;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

@Tag("unit")
class ClientConfigurationOptionsTest {

    @Test
    @DisplayName("validate should reject when no grant types are enabled")
    void validate_noGrantTypes_shouldThrow() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("http://example.com");

        assertThatThrownBy(options::validate)
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("At least one grant type must be enabled");
    }

    @Test
    @DisplayName("validate should reject public client with client_credentials")
    void validate_publicClientWithClientCredentials_shouldThrow() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("http://example.com");
        options.setPublicClient(true);
        options.setClientCredentialsEnabled(true);

        assertThatThrownBy(options::validate)
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("Public clients cannot use client_credentials");
    }

    @Test
    @DisplayName("validate should reject blank tenantUrl")
    void validate_blankTenantUrl_shouldThrow() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("   ");
        options.setClientCredentialsEnabled(true);

        assertThatThrownBy(options::validate)
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("tenantUrl must not be blank");
    }

    @Test
    @DisplayName("validate should reject null tenantUrl")
    void validate_nullTenantUrl_shouldThrow() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setClientCredentialsEnabled(true);

        assertThatThrownBy(options::validate)
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("tenantUrl must not be blank");
    }

    @Test
    @DisplayName("validate should reject authorization_code without redirectUris")
    void validate_authCodeWithoutRedirectUris_shouldThrow() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("http://example.com");
        options.setAuthorizationCodeEnabled(true);

        assertThatThrownBy(options::validate)
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("redirectUris required");
    }

    @Test
    @DisplayName("validate should reject implicit without redirectUris")
    void validate_implicitWithoutRedirectUris_shouldThrow() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("http://example.com");
        options.setImplicitEnabled(true);

        assertThatThrownBy(options::validate)
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("redirectUris required");
    }

    @Test
    @DisplayName("validate should accept valid client_credentials-only config")
    void validate_clientCredentialsOnly_shouldPass() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("http://example.com");
        options.setClientCredentialsEnabled(true);

        assertThatNoException().isThrownBy(options::validate);
    }

    @Test
    @DisplayName("validate should accept valid authorization_code with redirectUris")
    void validate_authCodeWithRedirectUris_shouldPass() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("http://example.com");
        options.setAuthorizationCodeEnabled(true);
        options.setRedirectUris(List.of("http://example.com/callback"));

        assertThatNoException().isThrownBy(options::validate);
    }

    @Test
    @DisplayName("validate should accept public client with authorization_code")
    void validate_publicClientWithAuthCode_shouldPass() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("http://example.com");
        options.setPublicClient(true);
        options.setAuthorizationCodeEnabled(true);
        options.setRedirectUris(List.of("http://example.com/callback"));

        assertThatNoException().isThrownBy(options::validate);
    }

    @Test
    @DisplayName("all boolean fields should default to false")
    void defaults_shouldBeFalse() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();

        assertThatThrownBy(() -> {
            options.setTenantUrl("http://example.com");
            options.validate();
        }).isInstanceOf(IllegalArgumentException.class)
                .hasMessageContaining("At least one grant type must be enabled");

        // Verify all defaults are false
        assertThat(options.isAuthorizationCodeEnabled()).isFalse();
        assertThat(options.isClientCredentialsEnabled()).isFalse();
        assertThat(options.isImplicitEnabled()).isFalse();
        assertThat(options.isPasswordEnabled()).isFalse();
        assertThat(options.isDeviceCodeEnabled()).isFalse();
        assertThat(options.isPkceEnabled()).isFalse();
        assertThat(options.isPublicClient()).isFalse();
    }
}
