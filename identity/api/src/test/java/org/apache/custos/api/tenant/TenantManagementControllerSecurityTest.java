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

package org.apache.custos.api.tenant;

import org.apache.custos.core.tenant.management.api.CreateTenantResponse;
import org.apache.custos.core.tenant.profile.api.GetAllTenantsResponse;
import org.apache.custos.core.tenant.profile.api.GetAttributeUpdateAuditTrailResponse;
import org.apache.custos.core.tenant.profile.api.GetStatusUpdateAuditTrailResponse;
import org.apache.custos.core.tenant.profile.api.GetTenantsRequest;
import org.apache.custos.core.tenant.profile.api.Tenant;
import org.apache.custos.core.tenant.profile.api.UpdateStatusRequest;
import org.apache.custos.core.tenant.profile.api.UpdateStatusResponse;
import org.apache.custos.service.auth.AuthClaim;
import org.apache.custos.service.auth.TokenAuthorizer;
import org.apache.custos.service.management.ClientConfigurationOptions;
import org.apache.custos.service.management.TenantManagementService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.web.server.ResponseStatusException;

import java.util.List;
import java.util.Map;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Security-focused unit tests for {@link TenantManagementController}.
 *
 * <p>These tests verify that:
 * <ul>
 *   <li>Privileged endpoints require super-tenant credentials</li>
 *   <li>Audit endpoints enforce super-tenant or tenant-ownership authorization</li>
 *   <li>The configureClient endpoint enforces admin authorization</li>
 * </ul>
 */
@org.junit.jupiter.api.Tag("unit")
@ExtendWith(MockitoExtension.class)
class TenantManagementControllerSecurityTest {

    @Mock
    private TenantManagementService tenantManagementService;

    @Mock
    private TokenAuthorizer tokenAuthorizer;

    @InjectMocks
    private TenantManagementController controller;

    private HttpHeaders headers;
    private AuthClaim superTenantClaim;
    private AuthClaim regularTenantClaim;

    @BeforeEach
    void setUp() {
        headers = new HttpHeaders();
        headers.add("Authorization", "Bearer test-token");

        superTenantClaim = new AuthClaim();
        superTenantClaim.setSuperTenant(true);
        superTenantClaim.setTenantId(1L);

        regularTenantClaim = new AuthClaim();
        regularTenantClaim.setSuperTenant(false);
        regularTenantClaim.setTenantId(42L);
    }

    // updateTenantStatus — without auth should reject
    @Test
    @DisplayName("updateTenantStatus without auth token should return 401")
    void updateTenantStatus_withoutAuth_shouldReject() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.empty());

        UpdateStatusRequest request = UpdateStatusRequest.newBuilder().build();

        assertThatThrownBy(() -> controller.updateTenantStatus(request, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> {
                    ResponseStatusException rse = (ResponseStatusException) ex;
                    assertThat(rse.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
                    assertThat(rse.getReason()).contains("Super-tenant credentials required");
                });

        verify(tenantManagementService, never()).updateTenantStatus(any());
    }

    @Test
    @DisplayName("updateTenantStatus with non-super-tenant token should return 401")
    void updateTenantStatus_withNonSuperTenantAuth_shouldReject() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(regularTenantClaim));

        UpdateStatusRequest request = UpdateStatusRequest.newBuilder().build();

        assertThatThrownBy(() -> controller.updateTenantStatus(request, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> {
                    ResponseStatusException rse = (ResponseStatusException) ex;
                    assertThat(rse.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
                });

        verify(tenantManagementService, never()).updateTenantStatus(any());
    }

    // updateTenantStatus — super-tenant should succeed
    @Test
    @DisplayName("updateTenantStatus with super-tenant token should succeed")
    void updateTenantStatus_withSuperTenantAuth_shouldSucceed() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(superTenantClaim));

        UpdateStatusRequest request = UpdateStatusRequest.newBuilder().build();
        UpdateStatusResponse mockResponse = UpdateStatusResponse.newBuilder().build();
        when(tenantManagementService.updateTenantStatus(any(UpdateStatusRequest.class)))
                .thenReturn(mockResponse);

        var response = controller.updateTenantStatus(request, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isEqualTo(mockResponse);
        verify(tenantManagementService, times(1)).updateTenantStatus(any(UpdateStatusRequest.class));
    }


    // getAllTenants — without auth should reject
    @Test
    @DisplayName("getAllTenants without auth token should return 401")
    void getAllTenants_withoutAuth_shouldReject() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.empty());

        GetTenantsRequest request = GetTenantsRequest.newBuilder().build();

        assertThatThrownBy(() -> controller.getAllTenants(request, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> {
                    ResponseStatusException rse = (ResponseStatusException) ex;
                    assertThat(rse.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
                    assertThat(rse.getReason()).contains("Super-tenant credentials required");
                });

        verify(tenantManagementService, never()).getAllTenants(any());
    }

    @Test
    @DisplayName("getAllTenants with non-super-tenant token should return 401")
    void getAllTenants_withNonSuperTenantAuth_shouldReject() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(regularTenantClaim));

        GetTenantsRequest request = GetTenantsRequest.newBuilder().build();

        assertThatThrownBy(() -> controller.getAllTenants(request, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> assertThat(((ResponseStatusException) ex).getStatusCode())
                        .isEqualTo(HttpStatus.UNAUTHORIZED));

        verify(tenantManagementService, never()).getAllTenants(any());
    }


    // getAllTenants — super-tenant should succeed
    @Test
    @DisplayName("getAllTenants with super-tenant token should succeed")
    void getAllTenants_withSuperTenantAuth_shouldSucceed() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(superTenantClaim));

        GetTenantsRequest request = GetTenantsRequest.newBuilder().build();
        GetAllTenantsResponse mockResponse = GetAllTenantsResponse.newBuilder().build();
        when(tenantManagementService.getAllTenants(any(GetTenantsRequest.class)))
                .thenReturn(mockResponse);

        var response = controller.getAllTenants(request, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).isEqualTo(mockResponse);
        verify(tenantManagementService, times(1)).getAllTenants(any(GetTenantsRequest.class));
    }


    // configureClient — auth and authorization tests
    @Test
    @DisplayName("configureClient without auth token should return 401")
    void configureClient_withoutAuth_shouldReject() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.empty());

        ClientConfigurationOptions options = createValidOptions();

        assertThatThrownBy(() -> controller.configureClient(100, options, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> {
                    ResponseStatusException rse = (ResponseStatusException) ex;
                    assertThat(rse.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
                    assertThat(rse.getReason()).contains("Valid credentials required");
                });

        verify(tenantManagementService, never()).addClient(anyLong(), any(ClientConfigurationOptions.class));
    }

    @Test
    @DisplayName("configureClient with non-admin user should return 403")
    void configureClient_withNonAdminAuth_shouldReject() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(regularTenantClaim));

        ClientConfigurationOptions options = createValidOptions();

        assertThatThrownBy(() -> controller.configureClient(100, options, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> assertThat(((ResponseStatusException) ex).getStatusCode())
                        .isEqualTo(HttpStatus.FORBIDDEN));

        verify(tenantManagementService, never()).addClient(anyLong(), any(ClientConfigurationOptions.class));
    }

    @Test
    @DisplayName("configureClient with super-tenant token should succeed")
    void configureClient_withSuperTenantAuth_shouldSucceed() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(superTenantClaim));
        when(tenantManagementService.addClient(anyLong(), any(ClientConfigurationOptions.class)))
                .thenReturn(Map.of("clientId", "client-123", "clientSecret", "secret-456"));

        ClientConfigurationOptions options = createValidOptions();
        var response = controller.configureClient(100, options, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody()).containsEntry("clientId", "client-123");
        verify(tenantManagementService, times(1)).addClient(anyLong(), any(ClientConfigurationOptions.class));
    }

    @Test
    @DisplayName("configureClient with tenant admin of same tenant should succeed")
    void configureClient_withSameTenantAdmin_shouldSucceed() {
        AuthClaim adminClaim = new AuthClaim();
        adminClaim.setSuperTenant(false);
        adminClaim.setAdmin(true);
        adminClaim.setTenantId(42L);

        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(adminClaim));
        when(tenantManagementService.addClient(anyLong(), any(ClientConfigurationOptions.class)))
                .thenReturn(Map.of("clientId", "client-123", "clientSecret", "secret-456"));

        ClientConfigurationOptions options = createValidOptions();
        var response = controller.configureClient(42, options, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        verify(tenantManagementService, times(1)).addClient(anyLong(), any(ClientConfigurationOptions.class));
    }

    @Test
    @DisplayName("configureClient with parent tenant admin should succeed for child tenant")
    void configureClient_withParentTenantAdmin_shouldSucceed() {
        AuthClaim parentAdminClaim = new AuthClaim();
        parentAdminClaim.setSuperTenant(false);
        parentAdminClaim.setAdmin(true);
        parentAdminClaim.setTenantId(10L);

        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(parentAdminClaim));
        when(tokenAuthorizer.validateParentChildTenantRelationShip(10L, 42L))
                .thenReturn(true);
        when(tenantManagementService.addClient(anyLong(), any(ClientConfigurationOptions.class)))
                .thenReturn(Map.of("clientId", "client-123", "clientSecret", "secret-456"));

        ClientConfigurationOptions options = createValidOptions();
        var response = controller.configureClient(42, options, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        verify(tokenAuthorizer).validateParentChildTenantRelationShip(10L, 42L);
    }

    @Test
    @DisplayName("configureClient with unrelated tenant admin should return 403")
    void configureClient_withUnrelatedTenantAdmin_shouldReject() {
        AuthClaim unrelatedAdminClaim = new AuthClaim();
        unrelatedAdminClaim.setSuperTenant(false);
        unrelatedAdminClaim.setAdmin(true);
        unrelatedAdminClaim.setTenantId(99L);

        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(unrelatedAdminClaim));
        when(tokenAuthorizer.validateParentChildTenantRelationShip(99L, 42L))
                .thenReturn(false);

        ClientConfigurationOptions options = createValidOptions();

        assertThatThrownBy(() -> controller.configureClient(42, options, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> {
                    ResponseStatusException rse = (ResponseStatusException) ex;
                    assertThat(rse.getStatusCode()).isEqualTo(HttpStatus.FORBIDDEN);
                    assertThat(rse.getReason()).contains("Not authorized for this tenant");
                });

        verify(tenantManagementService, never()).addClient(anyLong(), any(ClientConfigurationOptions.class));
    }


    // Audit endpoints — without auth should reject
    @Test
    @DisplayName("getTenantStatusUpdateAuditTrail without auth token should return 401")
    void auditStatusEndpoint_withoutAuth_shouldReject() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.empty());

        assertThatThrownBy(() -> controller.getTenantStatusUpdateAuditTrail(42L, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> {
                    ResponseStatusException rse = (ResponseStatusException) ex;
                    assertThat(rse.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
                    assertThat(rse.getReason()).contains("Super-tenant credentials or tenant ownership required");
                });

        verify(tenantManagementService, never()).getTenantStatusUpdateAuditTrail(any());
    }

    @Test
    @DisplayName("getTenantAttributeUpdateAuditTrail without auth token should return 401")
    void auditAttributesEndpoint_withoutAuth_shouldReject() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.empty());

        assertThatThrownBy(() -> controller.getTenantAttributeUpdateAuditTrail(42, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> {
                    ResponseStatusException rse = (ResponseStatusException) ex;
                    assertThat(rse.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
                    assertThat(rse.getReason()).contains("Super-tenant credentials or tenant ownership required");
                });

        verify(tenantManagementService, never()).getTenantAttributeUpdateAuditTrail(any());
    }

    @Test
    @DisplayName("getTenantStatusUpdateAuditTrail with wrong-tenant token should return 401")
    void auditStatusEndpoint_withWrongTenantAuth_shouldReject() {
        // regularTenantClaim has tenantId=42, requesting audit for tenantId=99
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(regularTenantClaim));

        assertThatThrownBy(() -> controller.getTenantStatusUpdateAuditTrail(99L, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> assertThat(((ResponseStatusException) ex).getStatusCode())
                        .isEqualTo(HttpStatus.UNAUTHORIZED));

        verify(tenantManagementService, never()).getTenantStatusUpdateAuditTrail(any());
    }

    @Test
    @DisplayName("getTenantStatusUpdateAuditTrail with super-tenant token should succeed for any tenant")
    void auditStatusEndpoint_withSuperTenantAuth_shouldSucceed() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(superTenantClaim));
        when(tenantManagementService.getTenantStatusUpdateAuditTrail(any()))
                .thenReturn(GetStatusUpdateAuditTrailResponse.newBuilder().build());

        var response = controller.getTenantStatusUpdateAuditTrail(99L, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        verify(tenantManagementService, times(1)).getTenantStatusUpdateAuditTrail(any());
    }

    @Test
    @DisplayName("getTenantStatusUpdateAuditTrail with own tenant token should succeed")
    void auditStatusEndpoint_withOwnerTenantAuth_shouldSucceed() {
        // regularTenantClaim has tenantId=42, requesting audit for same tenantId=42
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(regularTenantClaim));
        when(tenantManagementService.getTenantStatusUpdateAuditTrail(any()))
                .thenReturn(GetStatusUpdateAuditTrailResponse.newBuilder().build());

        var response = controller.getTenantStatusUpdateAuditTrail(42L, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        verify(tenantManagementService, times(1)).getTenantStatusUpdateAuditTrail(any());
    }

    @Test
    @DisplayName("getTenantAttributeUpdateAuditTrail with super-tenant token should succeed for any tenant")
    void auditAttributesEndpoint_withSuperTenantAuth_shouldSucceed() {
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(superTenantClaim));
        when(tenantManagementService.getTenantAttributeUpdateAuditTrail(any()))
                .thenReturn(GetAttributeUpdateAuditTrailResponse.newBuilder().build());

        var response = controller.getTenantAttributeUpdateAuditTrail(99, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        verify(tenantManagementService, times(1)).getTenantAttributeUpdateAuditTrail(any());
    }

    @Test
    @DisplayName("getTenantAttributeUpdateAuditTrail with own tenant token should succeed")
    void auditAttributesEndpoint_withOwnerTenantAuth_shouldSucceed() {
        // regularTenantClaim has tenantId=42
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(regularTenantClaim));
        when(tenantManagementService.getTenantAttributeUpdateAuditTrail(any()))
                .thenReturn(GetAttributeUpdateAuditTrailResponse.newBuilder().build());

        var response = controller.getTenantAttributeUpdateAuditTrail(42, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        verify(tenantManagementService, times(1)).getTenantAttributeUpdateAuditTrail(any());
    }

    @Test
    @DisplayName("getTenantAttributeUpdateAuditTrail with wrong-tenant token should return 401")
    void auditAttributesEndpoint_withWrongTenantAuth_shouldReject() {
        // regularTenantClaim has tenantId=42, requesting audit for tenantId=99
        when(tokenAuthorizer.authorizeUsingUserToken(any(HttpHeaders.class)))
                .thenReturn(Optional.of(regularTenantClaim));

        assertThatThrownBy(() -> controller.getTenantAttributeUpdateAuditTrail(99, headers))
                .isInstanceOf(ResponseStatusException.class)
                .satisfies(ex -> assertThat(((ResponseStatusException) ex).getStatusCode())
                        .isEqualTo(HttpStatus.UNAUTHORIZED));

        verify(tenantManagementService, never()).getTenantAttributeUpdateAuditTrail(any());
    }


    // Additional security boundary tests
    @Test
    @DisplayName("createTenant with bearer token succeeds without parent tenant injection")
    void createTenant_withBearerToken_doesNotSetParentTenantId() {
        // For createTenant the logic is: if getToken returns non-blank (bearer present),
        // the authorize(headers) fallback to set parentTenantId is skipped.
        // This verifies the endpoint is reachable and delegates to the service.
        Tenant tenantRequest = Tenant.newBuilder()
                .setClientName("Test Client")
                .build();

        CreateTenantResponse mockResponse = CreateTenantResponse.newBuilder()
                .setClientId("new-client-id")
                .build();
        when(tenantManagementService.createTenant(any())).thenReturn(mockResponse);

        var response = controller.createTenant(tenantRequest, headers);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        assertThat(response.getBody().getClientId()).isEqualTo("new-client-id");
    }

    private ClientConfigurationOptions createValidOptions() {
        ClientConfigurationOptions options = new ClientConfigurationOptions();
        options.setTenantUrl("http://example.com");
        options.setAuthorizationCodeEnabled(true);
        options.setRedirectUris(List.of("http://example.com/callback"));
        return options;
    }
}
