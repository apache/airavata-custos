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

import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.tenant.management.api.CreateTenantResponse;
import org.apache.custos.core.tenant.profile.api.GetAllTenantsResponse;
import org.apache.custos.core.tenant.profile.api.GetTenantsRequest;
import org.apache.custos.core.tenant.profile.api.Tenant;
import org.apache.custos.core.tenant.profile.api.TenantStatus;
import org.apache.custos.core.tenant.profile.api.TenantType;
import org.apache.custos.core.tenant.profile.api.UpdateStatusRequest;
import org.apache.custos.core.tenant.profile.api.UpdateStatusResponse;
import org.apache.custos.service.federated.client.keycloak.KeycloakClient;
import org.apache.custos.service.federated.client.keycloak.KeycloakClientSecret;
import org.apache.custos.service.profile.TenantProfileService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.boot.ApplicationArguments;
import org.springframework.test.util.ReflectionTestUtils;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.lenient;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class SuperTenantBootstrapperTest {

    private static final String TEST_USERNAME = "custosadmin";
    private static final String TEST_PASSWORD = "testpass123";
    private static final String TEST_EMAIL = "admin@test.local";
    private static final String TEST_CLIENT_ID = "custos-abc123";
    private static final long TEST_TENANT_ID = 10000001L;

    private static final String CILOGON_CLIENT_ID = "cilogon:/client_id/test123";
    private static final String CILOGON_CLIENT_SECRET = "test-cilogon-secret";

    @Mock
    private TenantManagementService tenantManagementService;
    @Mock
    private TenantProfileService tenantProfileService;
    @Mock
    private KeycloakClient keycloakClient;
    @Mock
    private ApplicationArguments applicationArguments;
    private SuperTenantProperties properties;
    private SuperTenantBootstrapper bootstrapper;

    @BeforeEach
    void setUp() {
        properties = new SuperTenantProperties(
                new SuperTenantProperties.Admin(TEST_USERNAME, TEST_PASSWORD, TEST_EMAIL),
                "http://localhost:8080/,http://localhost:5173/callback/",
                "http://localhost:8080/",
                "openid email profile cilogon",
                "localhost"
        );

        bootstrapper = new SuperTenantBootstrapper(tenantManagementService, tenantProfileService, keycloakClient, properties);

        // Keycloak readiness check
        lenient().when(keycloakClient.realmExists("master")).thenReturn(true);
    }

    /**
     * When an active super-tenant already exists, createTenant must NOT be called.
     */
    @Test
    void run_whenSuperTenantExists_shouldSkipBootstrap() throws Exception {
        Tenant existingSuperTenant = Tenant.newBuilder()
                .setTenantId(TEST_TENANT_ID)
                .setParentTenantId(0)
                .setTenantStatus(TenantStatus.ACTIVE)
                .build();

        GetAllTenantsResponse existingResponse = GetAllTenantsResponse.newBuilder()
                .addTenant(existingSuperTenant)
                .build();

        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(existingResponse);

        bootstrapper.run(applicationArguments);

        verify(tenantManagementService, never()).createTenant(any());
        verify(tenantManagementService, never()).updateTenantStatus(any());
    }

    /**
     * When no super-tenant exists, createTenant and updateTenantStatus must be called in sequence.
     */
    @Test
    void run_whenNoSuperTenantExists_shouldCreateAndActivate() throws Exception {
        GetAllTenantsResponse emptyResponse = GetAllTenantsResponse.newBuilder().build();
        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(emptyResponse);

        CreateTenantResponse createResponse = CreateTenantResponse.newBuilder()
                .setClientId(TEST_CLIENT_ID)
                .setClientSecret("should-not-appear-in-logs")
                .build();
        when(tenantManagementService.createTenant(any(
                org.apache.custos.core.tenant.profile.api.Tenant.class))).thenReturn(createResponse);

        UpdateStatusResponse statusResponse = UpdateStatusResponse.newBuilder()
                .setTenantId(TEST_TENANT_ID)
                .setStatus(TenantStatus.ACTIVE)
                .build();
        when(tenantManagementService.updateTenantStatus(any(UpdateStatusRequest.class))).thenReturn(statusResponse);

        bootstrapper.run(applicationArguments);

        verify(tenantManagementService).createTenant(any(org.apache.custos.core.tenant.profile.api.Tenant.class));
        verify(tenantManagementService).updateTenantStatus(any(UpdateStatusRequest.class));
    }

    /**
     * When createTenant throws an exception, the exception must propagate so the application
     * fails to start rather than silently continuing without a super-tenant.
     */
    @Test
    void run_whenCreateTenantFails_shouldPropagateException() {
        GetAllTenantsResponse emptyResponse = GetAllTenantsResponse.newBuilder().build();
        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(emptyResponse);

        RuntimeException cause = new IllegalArgumentException("IAM server unreachable during tenant creation");
        when(tenantManagementService.createTenant(any(
                org.apache.custos.core.tenant.profile.api.Tenant.class))).thenThrow(cause);

        assertThatThrownBy(() -> bootstrapper.run(applicationArguments))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("IAM server unreachable during tenant creation");

        verify(tenantManagementService, never()).updateTenantStatus(any());
    }

    /**
     * Verifies the admin credentials configured via properties are forwarded verbatim to createTenant.
     */
    @Test
    void run_shouldPassConfiguredCredentials() throws Exception {
        GetAllTenantsResponse emptyResponse = GetAllTenantsResponse.newBuilder().build();
        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(emptyResponse);

        CreateTenantResponse createResponse = CreateTenantResponse.newBuilder()
                .setClientId(TEST_CLIENT_ID)
                .build();
        when(tenantManagementService.createTenant(any(
                org.apache.custos.core.tenant.profile.api.Tenant.class))).thenReturn(createResponse);

        UpdateStatusResponse statusResponse = UpdateStatusResponse.newBuilder()
                .setTenantId(TEST_TENANT_ID)
                .build();
        when(tenantManagementService.updateTenantStatus(any(UpdateStatusRequest.class))).thenReturn(statusResponse);

        bootstrapper.run(applicationArguments);

        ArgumentCaptor<org.apache.custos.core.tenant.profile.api.Tenant> tenantCaptor =
                ArgumentCaptor.forClass(org.apache.custos.core.tenant.profile.api.Tenant.class);
        verify(tenantManagementService).createTenant(tenantCaptor.capture());

        org.apache.custos.core.tenant.profile.api.Tenant captured = tenantCaptor.getValue();
        assertThat(captured.getAdminUsername()).isEqualTo(TEST_USERNAME);
        assertThat(captured.getAdminPassword()).isEqualTo(TEST_PASSWORD);
        assertThat(captured.getAdminEmail()).isEqualTo(TEST_EMAIL);
        assertThat(captured.getRequesterEmail()).isEqualTo(TEST_EMAIL);
    }

    /**
     * Verifies that the UpdateStatusRequest sent to updateTenantStatus has superTenant=true,
     * ACTIVE status, and the clientId returned from createTenant.
     */
    @Test
    void run_shouldSetSuperTenantTrueInStatusUpdate() throws Exception {
        GetAllTenantsResponse emptyResponse = GetAllTenantsResponse.newBuilder().build();
        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(emptyResponse);

        CreateTenantResponse createResponse = CreateTenantResponse.newBuilder()
                .setClientId(TEST_CLIENT_ID)
                .build();
        when(tenantManagementService.createTenant(any(
                org.apache.custos.core.tenant.profile.api.Tenant.class))).thenReturn(createResponse);

        UpdateStatusResponse statusResponse = UpdateStatusResponse.newBuilder()
                .setTenantId(TEST_TENANT_ID)
                .setStatus(TenantStatus.ACTIVE)
                .build();

        ArgumentCaptor<UpdateStatusRequest> statusCaptor = ArgumentCaptor.forClass(UpdateStatusRequest.class);
        when(tenantManagementService.updateTenantStatus(statusCaptor.capture())).thenReturn(statusResponse);

        bootstrapper.run(applicationArguments);

        UpdateStatusRequest capturedStatus = statusCaptor.getValue();
        assertThat(capturedStatus.getSuperTenant()).isTrue();
        assertThat(capturedStatus.getStatus()).isEqualTo(TenantStatus.ACTIVE);
        assertThat(capturedStatus.getClientId()).isEqualTo(TEST_CLIENT_ID);
        assertThat(capturedStatus.getUpdatedBy()).isEqualTo(Constants.SYSTEM);
    }

    /**
     * Verifies that the idempotency check queries for ADMIN type tenants with ACTIVE status.
     */
    @Test
    void activeSuperTenantExists_shouldQueryWithAdminTypeAndActiveStatus() {
        GetAllTenantsResponse emptyResponse = GetAllTenantsResponse.newBuilder().build();

        ArgumentCaptor<GetTenantsRequest> requestCaptor = ArgumentCaptor.forClass(GetTenantsRequest.class);
        when(tenantProfileService.getAllTenants(requestCaptor.capture())).thenReturn(emptyResponse);

        boolean result = bootstrapper.activeSuperTenantExists();

        assertThat(result).isFalse();
        GetTenantsRequest captured = requestCaptor.getValue();
        assertThat(captured.getType()).isEqualTo(TenantType.ADMIN);
        assertThat(captured.getStatus()).isEqualTo(TenantStatus.ACTIVE);
        assertThat(captured.getLimit()).isEqualTo(1);
    }

    /**
     * Verifies that activeSuperTenantExists returns false when the profile service
     * returns null (defensive null check).
     */
    @Test
    void activeSuperTenantExists_whenNullResponse_shouldReturnFalse() {
        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(null);

        boolean result = bootstrapper.activeSuperTenantExists();

        assertThat(result).isFalse();
    }

    /**
     * Verifies that comma-separated redirect URIs configured via properties are forwarded
     * as a list (trimmed) to createTenant.
     */
    @Test
    void run_shouldParseRedirectUrisAsList() throws Exception {
        GetAllTenantsResponse emptyResponse = GetAllTenantsResponse.newBuilder().build();
        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(emptyResponse);

        CreateTenantResponse createResponse = CreateTenantResponse.newBuilder()
                .setClientId(TEST_CLIENT_ID)
                .build();
        when(tenantManagementService.createTenant(any(
                org.apache.custos.core.tenant.profile.api.Tenant.class))).thenReturn(createResponse);

        UpdateStatusResponse statusResponse = UpdateStatusResponse.newBuilder()
                .setTenantId(TEST_TENANT_ID)
                .build();
        when(tenantManagementService.updateTenantStatus(any(UpdateStatusRequest.class))).thenReturn(statusResponse);

        bootstrapper.run(applicationArguments);

        ArgumentCaptor<org.apache.custos.core.tenant.profile.api.Tenant> tenantCaptor =
                ArgumentCaptor.forClass(org.apache.custos.core.tenant.profile.api.Tenant.class);
        verify(tenantManagementService).createTenant(tenantCaptor.capture());

        org.apache.custos.core.tenant.profile.api.Tenant captured = tenantCaptor.getValue();
        assertThat(captured.getRedirectUrisList())
                .containsExactly("http://localhost:8080/", "http://localhost:5173/callback/");
    }

    /**
     * When redirect URIs property is blank, the tenant must be created with an empty list.
     */
    @Test
    void run_whenRedirectUrisBlank_shouldPassEmptyList() throws Exception {
        properties = new SuperTenantProperties(
                properties.admin(), "", properties.clientUri(), properties.scope(), properties.domain());
        bootstrapper = new SuperTenantBootstrapper(tenantManagementService, tenantProfileService, keycloakClient, properties);

        GetAllTenantsResponse emptyResponse = GetAllTenantsResponse.newBuilder().build();
        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(emptyResponse);

        CreateTenantResponse createResponse = CreateTenantResponse.newBuilder()
                .setClientId(TEST_CLIENT_ID)
                .build();
        when(tenantManagementService.createTenant(any(
                org.apache.custos.core.tenant.profile.api.Tenant.class))).thenReturn(createResponse);

        UpdateStatusResponse statusResponse = UpdateStatusResponse.newBuilder()
                .setTenantId(TEST_TENANT_ID)
                .build();
        when(tenantManagementService.updateTenantStatus(any(UpdateStatusRequest.class))).thenReturn(statusResponse);

        bootstrapper.run(applicationArguments);

        ArgumentCaptor<org.apache.custos.core.tenant.profile.api.Tenant> tenantCaptor =
                ArgumentCaptor.forClass(org.apache.custos.core.tenant.profile.api.Tenant.class);
        verify(tenantManagementService).createTenant(tenantCaptor.capture());

        assertThat(tenantCaptor.getValue().getRedirectUrisList()).isEmpty();
    }

    private void stubFullBootstrap() {
        GetAllTenantsResponse emptyResponse = GetAllTenantsResponse.newBuilder().build();
        when(tenantProfileService.getAllTenants(any(GetTenantsRequest.class))).thenReturn(emptyResponse);

        CreateTenantResponse createResponse = CreateTenantResponse.newBuilder()
                .setClientId(TEST_CLIENT_ID)
                .build();
        when(tenantManagementService.createTenant(any(Tenant.class))).thenReturn(createResponse);

        UpdateStatusResponse statusResponse = UpdateStatusResponse.newBuilder()
                .setTenantId(TEST_TENANT_ID)
                .setStatus(TenantStatus.ACTIVE)
                .build();
        when(tenantManagementService.updateTenantStatus(any(UpdateStatusRequest.class))).thenReturn(statusResponse);
    }

    @Test
    void configureCILogonIDP_whenEnabled_callsKeycloakClient() {
        stubFullBootstrap();
        when(keycloakClient.configureOIDCFederatedIDP(anyString(), anyString(), anyString(), any(KeycloakClientSecret.class), isNull()))
                .thenReturn(true);

        ReflectionTestUtils.setField(bootstrapper, "ciLogonEnabled", true);
        ReflectionTestUtils.setField(bootstrapper, "ciLogonClientId", CILOGON_CLIENT_ID);
        ReflectionTestUtils.setField(bootstrapper, "ciLogonClientSecret", CILOGON_CLIENT_SECRET);

        bootstrapper.run(applicationArguments);

        ArgumentCaptor<KeycloakClientSecret> secretCaptor = ArgumentCaptor.forClass(KeycloakClientSecret.class);
        verify(keycloakClient).configureOIDCFederatedIDP(
                eq(String.valueOf(TEST_TENANT_ID)),
                eq("CILogon"),
                eq("openid profile email org.cilogon.userinfo"),
                secretCaptor.capture(),
                isNull()
        );

        KeycloakClientSecret capturedSecret = secretCaptor.getValue();
        assertThat(capturedSecret.clientId()).isEqualTo(CILOGON_CLIENT_ID);
        assertThat(capturedSecret.clientSecret()).isEqualTo(CILOGON_CLIENT_SECRET);
    }

    @Test
    void configureCILogonIDP_whenDisabled_skipsConfiguration() throws Exception {
        stubFullBootstrap();

        ReflectionTestUtils.setField(bootstrapper, "ciLogonEnabled", false);
        ReflectionTestUtils.setField(bootstrapper, "ciLogonClientId", CILOGON_CLIENT_ID);
        ReflectionTestUtils.setField(bootstrapper, "ciLogonClientSecret", CILOGON_CLIENT_SECRET);

        bootstrapper.run(applicationArguments);

        verify(keycloakClient, never()).configureOIDCFederatedIDP(anyString(), anyString(), anyString(), any(KeycloakClientSecret.class), any());
    }

    @Test
    void configureCILogonIDP_whenEnabledButCredentialsMissing_skipsWithWarning() {
        stubFullBootstrap();

        ReflectionTestUtils.setField(bootstrapper, "ciLogonEnabled", true);
        ReflectionTestUtils.setField(bootstrapper, "ciLogonClientId", "");
        ReflectionTestUtils.setField(bootstrapper, "ciLogonClientSecret", "");

        bootstrapper.run(applicationArguments);

        verify(keycloakClient, never()).configureOIDCFederatedIDP(anyString(), anyString(), anyString(), any(KeycloakClientSecret.class), any());
    }

    @Test
    void fullBootstrap_whenCILogonEnabled_configuresIDPAfterTenantCreation() {
        stubFullBootstrap();
        when(keycloakClient.configureOIDCFederatedIDP(anyString(), anyString(), anyString(), any(KeycloakClientSecret.class), isNull()))
                .thenReturn(true);

        ReflectionTestUtils.setField(bootstrapper, "ciLogonEnabled", true);
        ReflectionTestUtils.setField(bootstrapper, "ciLogonClientId", CILOGON_CLIENT_ID);
        ReflectionTestUtils.setField(bootstrapper, "ciLogonClientSecret", CILOGON_CLIENT_SECRET);

        bootstrapper.run(applicationArguments);

        // Verify full sequence: create tenant -> activate -> configure CILogon IDP
        var inOrder = org.mockito.Mockito.inOrder(tenantManagementService, keycloakClient);
        inOrder.verify(tenantManagementService).createTenant(any(Tenant.class));
        inOrder.verify(tenantManagementService).updateTenantStatus(any(UpdateStatusRequest.class));
        inOrder.verify(keycloakClient).configureOIDCFederatedIDP(anyString(), anyString(), anyString(), any(KeycloakClientSecret.class), isNull());
    }
}
