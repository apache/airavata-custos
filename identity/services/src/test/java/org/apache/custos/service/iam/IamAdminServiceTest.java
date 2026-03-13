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

package org.apache.custos.service.iam;

import org.apache.custos.core.commons.StatusUpdater;
import org.apache.custos.core.iam.api.SetUpTenantRequest;
import org.apache.custos.core.iam.api.SetUpTenantResponse;
import org.apache.custos.service.auth.TokenService;
import org.apache.custos.service.federated.client.keycloak.KeycloakClient;
import org.apache.custos.service.federated.client.keycloak.KeycloakClientSecret;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class IamAdminServiceTest {

    private static final long TENANT_ID = 10000001L;
    private static final String REALM_ID = String.valueOf(TENANT_ID);
    private static final String TENANT_NAME = "Test Tenant";
    private static final String ADMIN_USERNAME = "admin";
    private static final String ADMIN_FIRSTNAME = "Admin";
    private static final String ADMIN_LASTNAME = "User";
    private static final String ADMIN_EMAIL = "admin@test.org";
    private static final String ADMIN_PASSWORD = "secret";
    private static final String CUSTOS_CLIENT_ID = "custos-test-client";
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

    /**
     * When the Keycloak realm already exists, createRealm must NOT be called.
     * deleteRealm must never be called under any circumstances.
     * The rest of the setup (admin account, client) should still proceed.
     */
    @Test
    void setUpTenant_whenRealmExists_shouldNotDeleteAndRecreate() {
        // Arrange
        SetUpTenantRequest request = buildSetUpTenantRequest();
        when(keycloakClient.realmExists(REALM_ID)).thenReturn(true);
        when(keycloakClient.updateClient(eq(REALM_ID), anyString(), anyString(), anyList()))
                .thenReturn(new KeycloakClientSecret(CUSTOS_CLIENT_ID, "iam-secret"));

        // Act
        SetUpTenantResponse response = iamAdminService.setUPTenant(request);

        // Assert
        verify(keycloakClient, never()).deleteRealm(any());
        verify(keycloakClient, never()).createRealm(any(), any());
        assertThat(response.getClientId()).isEqualTo(CUSTOS_CLIENT_ID);
    }

    /**
     * When the Keycloak realm does NOT exist, createRealm must be called exactly once.
     * deleteRealm must never be called.
     */
    @Test
    void setUpTenant_whenRealmDoesNotExist_shouldCreateRealm() {
        // Arrange
        SetUpTenantRequest request = buildSetUpTenantRequest();
        when(keycloakClient.realmExists(REALM_ID)).thenReturn(false);
        when(keycloakClient.createRealmAdminAccount(eq(REALM_ID), anyString(), anyString(), anyString(), anyString(), anyString()))
                .thenReturn(true);
        when(keycloakClient.configureClient(eq(REALM_ID), anyString(), anyString(), anyList()))
                .thenReturn(new KeycloakClientSecret(CUSTOS_CLIENT_ID, "iam-secret"));

        // Act
        SetUpTenantResponse response = iamAdminService.setUPTenant(request);

        // Assert
        verify(keycloakClient, never()).deleteRealm(any());
        verify(keycloakClient).createRealm(eq(REALM_ID), eq(TENANT_NAME));
        assertThat(response.getClientId()).isEqualTo(CUSTOS_CLIENT_ID);
    }

    /**
     * Regardless of whether the realm exists or not, deleteRealm must never be called by setUPTenant.
     */
    @Test
    void setUpTenant_shouldNotDeleteExistingRealm() {
        SetUpTenantRequest request = buildSetUpTenantRequest();
        when(keycloakClient.realmExists(REALM_ID)).thenReturn(false);
        when(keycloakClient.createRealmAdminAccount(eq(REALM_ID), anyString(), anyString(), anyString(), anyString(), anyString()))
                .thenReturn(true);
        when(keycloakClient.configureClient(eq(REALM_ID), anyString(), anyString(), anyList()))
                .thenReturn(new KeycloakClientSecret(CUSTOS_CLIENT_ID, "iam-secret"));

        // Act
        iamAdminService.setUPTenant(request);

        // Assert
        verify(keycloakClient, never()).deleteRealm(anyString());
    }

    /**
     * Re-running setUPTenant for an existing realm (idempotent re-setup) should
     * skip realm creation but still configure the admin account and client.
     */
    @Test
    void setUpTenant_whenRealmExists_shouldStillConfigureAdminAndClient() {
        // Arrange
        SetUpTenantRequest request = buildSetUpTenantRequest();
        when(keycloakClient.realmExists(REALM_ID)).thenReturn(true);
        KeycloakClientSecret expectedSecret = new KeycloakClientSecret(CUSTOS_CLIENT_ID, "iam-secret");
        when(keycloakClient.updateClient(eq(REALM_ID), anyString(), anyString(), anyList()))
                .thenReturn(expectedSecret);

        // Act
        SetUpTenantResponse response = iamAdminService.setUPTenant(request);

        // Assert
        verify(keycloakClient).updateRealmAdminAccount(eq(REALM_ID), eq(ADMIN_USERNAME),
                eq(ADMIN_FIRSTNAME), eq(ADMIN_LASTNAME), eq(ADMIN_EMAIL), eq(ADMIN_PASSWORD));
        verify(keycloakClient).updateClient(eq(REALM_ID), eq(CUSTOS_CLIENT_ID), eq(TENANT_URL), anyList());
        assertThat(response.getClientId()).isEqualTo(CUSTOS_CLIENT_ID);
        assertThat(response.getClientSecret()).isEqualTo("iam-secret");
    }

    private SetUpTenantRequest buildSetUpTenantRequest() {
        return SetUpTenantRequest.newBuilder()
                .setTenantId(TENANT_ID)
                .setTenantName(TENANT_NAME)
                .setAdminUsername(ADMIN_USERNAME)
                .setAdminFirstname(ADMIN_FIRSTNAME)
                .setAdminLastname(ADMIN_LASTNAME)
                .setAdminEmail(ADMIN_EMAIL)
                .setAdminPassword(ADMIN_PASSWORD)
                .setCustosClientId(CUSTOS_CLIENT_ID)
                .setTenantURL(TENANT_URL)
                .build();
    }
}
