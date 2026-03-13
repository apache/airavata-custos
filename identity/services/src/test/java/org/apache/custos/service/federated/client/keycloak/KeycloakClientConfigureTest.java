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

package org.apache.custos.service.federated.client.keycloak;

import org.apache.custos.core.commons.StatusUpdater;
import org.apache.custos.service.auth.TokenService;
import org.apache.custos.service.iam.IamAdminService;
import org.apache.custos.service.management.ClientConfigurationOptions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class KeycloakClientConfigureTest {

    private static final long TENANT_ID = 10000001L;
    private static final String CLIENT_ID = "test-client";
    private static final String TENANT_URL = "https://test.org";

    @Mock
    private KeycloakClient keycloakClient;

    @Mock
    private StatusUpdater statusUpdater;

    @Mock
    private TokenService tokenService;

    private IamAdminService iamAdminService;

    @BeforeEach
    void setUp() {
        iamAdminService = new IamAdminService(keycloakClient, statusUpdater, tokenService);
    }

    @Test
    @DisplayName("configureClient with options should delegate to KeycloakClient overload")
    void configureClient_withOptions_shouldDelegate() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl(TENANT_URL);
        options.setAuthorizationCodeEnabled(true);
        options.setRedirectUris(List.of("https://test.org/callback"));

        when(keycloakClient.configureClient(anyString(), anyString(), anyString(), anyList(), any(ClientConfigurationOptions.class)))
                .thenReturn(new KeycloakClientSecret(CLIENT_ID, "secret-123"));

        Map<String, String> result = iamAdminService.configureClient(TENANT_ID, CLIENT_ID, TENANT_URL, List.of("https://test.org/callback"), options);

        assertThat(result).containsEntry("clientId", CLIENT_ID);
        assertThat(result).containsEntry("clientSecret", "secret-123");
        verify(keycloakClient).configureClient(eq(String.valueOf(TENANT_ID)),
                eq(CLIENT_ID), eq(TENANT_URL), eq(List.of("https://test.org/callback")), eq(options));
    }

    @Test
    @DisplayName("configureClient with client_credentials options should pass options through")
    void configureClient_withClientCredentials_shouldPassOptions() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl(TENANT_URL);
        options.setClientCredentialsEnabled(true);

        when(keycloakClient.configureClient(anyString(), anyString(), anyString(), anyList(), any(ClientConfigurationOptions.class)))
                .thenReturn(new KeycloakClientSecret(CLIENT_ID, "secret-456"));

        Map<String, String> result = iamAdminService.configureClient(TENANT_ID, CLIENT_ID,
                TENANT_URL, List.of(), options);

        assertThat(result).containsEntry("clientId", CLIENT_ID);
        assertThat(result).containsEntry("clientSecret", "secret-456");
    }

    @Test
    @DisplayName("configureClient with PKCE options should pass through")
    void configureClient_withPkce_shouldPassOptions() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl(TENANT_URL);
        options.setAuthorizationCodeEnabled(true);
        options.setPkceEnabled(true);
        options.setRedirectUris(List.of("https://test.org/callback"));

        when(keycloakClient.configureClient(anyString(), anyString(), anyString(), anyList(), any(ClientConfigurationOptions.class)))
                .thenReturn(new KeycloakClientSecret(CLIENT_ID, "secret-789"));

        Map<String, String> result = iamAdminService.configureClient(TENANT_ID, CLIENT_ID, TENANT_URL, List.of("https://test.org/callback"), options);

        assertThat(result).containsEntry("clientId", CLIENT_ID);

        ArgumentCaptor<ClientConfigurationOptions> captor = ArgumentCaptor.forClass(ClientConfigurationOptions.class);
        verify(keycloakClient).configureClient(anyString(), anyString(), anyString(), anyList(), captor.capture());
        assertThat(captor.getValue().isPkceEnabled()).isTrue();
        assertThat(captor.getValue().isAuthorizationCodeEnabled()).isTrue();
    }

    @Test
    @DisplayName("configureClient with public client options should pass through")
    void configureClient_withPublicClient_shouldPassOptions() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl(TENANT_URL);
        options.setPublicClient(true);
        options.setAuthorizationCodeEnabled(true);
        options.setRedirectUris(List.of("https://test.org/callback"));

        when(keycloakClient.configureClient(anyString(), anyString(), anyString(), anyList(), any(ClientConfigurationOptions.class)))
                .thenReturn(new KeycloakClientSecret(CLIENT_ID, ""));

        Map<String, String> result = iamAdminService.configureClient(TENANT_ID, CLIENT_ID, TENANT_URL, List.of("https://test.org/callback"), options);

        assertThat(result).containsEntry("clientId", CLIENT_ID);

        ArgumentCaptor<ClientConfigurationOptions> captor = ArgumentCaptor.forClass(ClientConfigurationOptions.class);
        verify(keycloakClient).configureClient(anyString(), anyString(), anyString(), anyList(), captor.capture());
        assertThat(captor.getValue().isPublicClient()).isTrue();
    }

    @Test
    @DisplayName("legacy configureClient (4-arg) should still work")
    void configureClient_legacyFourArg_shouldWork() {
        when(keycloakClient.configureClient(anyString(), anyString(), anyString(), anyList()))
                .thenReturn(new KeycloakClientSecret(CLIENT_ID, "legacy-secret"));

        Map<String, String> result = iamAdminService.configureClient(TENANT_ID, CLIENT_ID, TENANT_URL, List.of("https://test.org/callback"));

        assertThat(result).containsEntry("clientId", CLIENT_ID);
        assertThat(result).containsEntry("clientSecret", "legacy-secret");
    }
}
