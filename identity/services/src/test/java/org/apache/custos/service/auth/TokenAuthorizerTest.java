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

package org.apache.custos.service.auth;

import org.apache.custos.core.credential.store.api.CredentialMetadata;
import org.apache.custos.core.credential.store.api.GetAllCredentialsResponse;
import org.apache.custos.core.credential.store.api.TokenRequest;
import org.apache.custos.core.credential.store.api.Type;
import org.apache.custos.core.model.tenant.Tenant;
import org.apache.custos.core.tenant.profile.api.GetTenantRequest;
import org.apache.custos.core.tenant.profile.api.GetTenantResponse;
import org.apache.custos.core.tenant.profile.api.TenantStatus;
import org.apache.custos.service.credential.store.CredentialStoreService;
import org.apache.custos.service.identity.IdentityService;
import org.apache.custos.service.profile.TenantProfileService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@Tag("unit")
class TokenAuthorizerTest {

    private static final long TENANT_ID = 10000001L;
    private static final String CLIENT_ID = "custos-test-client";
    private static final String CLIENT_SECRET = "test-secret";

    @Mock
    private CredentialStoreService credentialStoreService;

    @Mock
    private TenantProfileService tenantProfileService;

    @Mock
    private IdentityService identityService;

    private TokenAuthorizer tokenAuthorizer;

    @BeforeEach
    void setUp() {
        tokenAuthorizer = new TokenAuthorizer(credentialStoreService, tenantProfileService, identityService);
    }

    /**
     * A tenant with parentId == 0 is a root (super) tenant.
     * authorize() should result in an AuthClaim with isSuperTenant() == true.
     */
    @Test
    void authorize_withSuperTenant_shouldSetSuperTenantTrue() {
        // Arrange
        GetAllCredentialsResponse credentialsResponse = buildCredentialsResponse(TENANT_ID);
        when(credentialStoreService.getAllCredentialsFromToken(any(TokenRequest.class)))
                .thenReturn(credentialsResponse);

        Tenant superTenantEntity = buildTenantEntity(TENANT_ID, 0L);
        when(tenantProfileService.getTenantEntityByTenantId(TENANT_ID))
                .thenReturn(superTenantEntity);

        GetTenantResponse tenantResponse = buildActiveTenantResponse(TENANT_ID);
        when(tenantProfileService.getTenant(any(GetTenantRequest.class)))
                .thenReturn(tenantResponse);

        // Act
        Optional<AuthClaim> result = tokenAuthorizer.authorize("some-base64-token");

        // Assert
        assertThat(result).isPresent();
        assertThat(result.get().isSuperTenant()).isTrue();
    }

    /**
     * A tenant with parentId > 0 is a child/regular tenant.
     * authorize() should result in an AuthClaim with isSuperTenant() == false.
     */
    @Test
    void authorize_withChildTenant_shouldSetSuperTenantFalse() {
        // Arrange
        long parentTenantId = 10000000L;
        GetAllCredentialsResponse credentialsResponse = buildCredentialsResponse(TENANT_ID);
        when(credentialStoreService.getAllCredentialsFromToken(any(TokenRequest.class)))
                .thenReturn(credentialsResponse);

        Tenant childTenantEntity = buildTenantEntity(TENANT_ID, parentTenantId);
        when(tenantProfileService.getTenantEntityByTenantId(TENANT_ID))
                .thenReturn(childTenantEntity);

        GetTenantResponse tenantResponse = buildActiveTenantResponse(TENANT_ID);
        when(tenantProfileService.getTenant(any(GetTenantRequest.class)))
                .thenReturn(tenantResponse);

        // Act
        Optional<AuthClaim> result = tokenAuthorizer.authorize("some-base64-token");

        // Assert
        assertThat(result).isPresent();
        assertThat(result.get().isSuperTenant()).isFalse();
    }

    /**
     * When getTenantEntityByTenantId returns null (legacy/not-found case),
     * the tenant should NOT be treated as a super-tenant to fail safely.
     */
    @Test
    void authorize_withNullTenantEntity_shouldSetSuperTenantFalse() {
        // Arrange
        GetAllCredentialsResponse credentialsResponse = buildCredentialsResponse(TENANT_ID);
        when(credentialStoreService.getAllCredentialsFromToken(any(TokenRequest.class)))
                .thenReturn(credentialsResponse);

        when(tenantProfileService.getTenantEntityByTenantId(TENANT_ID))
                .thenReturn(null);

        GetTenantResponse tenantResponse = buildActiveTenantResponse(TENANT_ID);
        when(tenantProfileService.getTenant(any(GetTenantRequest.class)))
                .thenReturn(tenantResponse);

        // Act
        Optional<AuthClaim> result = tokenAuthorizer.authorize("some-base64-token");

        // Assert
        assertThat(result).isPresent();
        assertThat(result.get().isSuperTenant()).isFalse();
    }

    /**
     * Explicitly verifies that parentId == 0 is treated as super-tenant,
     * even when the entity's parentId field is a primitive long defaulting to 0.
     * This is a guard against any future change that might accidentally flip the comparison.
     */
    @Test
    void authorize_withParentIdZero_shouldSetSuperTenantTrue() {
        // Arrange
        GetAllCredentialsResponse credentialsResponse = buildCredentialsResponse(TENANT_ID);
        when(credentialStoreService.getAllCredentialsFromToken(any(TokenRequest.class)))
                .thenReturn(credentialsResponse);

        // parentId of 0 explicitly set — primitive long zero is the super-tenant sentinel
        Tenant tenantEntity = buildTenantEntity(TENANT_ID, 0L);
        when(tenantProfileService.getTenantEntityByTenantId(TENANT_ID))
                .thenReturn(tenantEntity);

        GetTenantResponse tenantResponse = buildActiveTenantResponse(TENANT_ID);
        when(tenantProfileService.getTenant(any(GetTenantRequest.class)))
                .thenReturn(tenantResponse);

        // Act
        Optional<AuthClaim> result = tokenAuthorizer.authorize("some-base64-token");

        // Assert
        assertThat(result).isPresent();
        assertThat(result.get().isSuperTenant())
                .as("Tenant with parentId == 0 must be identified as super-tenant")
                .isTrue();
    }

    /**
     * When the tenant profile returns an INACTIVE status, authorize() should return empty
     * (the tenant is not allowed to operate).
     */
    @Test
    void authorize_withInactiveTenant_shouldReturnEmpty() {
        // Arrange
        GetAllCredentialsResponse credentialsResponse = buildCredentialsResponse(TENANT_ID);
        when(credentialStoreService.getAllCredentialsFromToken(any(TokenRequest.class)))
                .thenReturn(credentialsResponse);

        Tenant tenantEntity = buildTenantEntity(TENANT_ID, 0L);
        when(tenantProfileService.getTenantEntityByTenantId(TENANT_ID))
                .thenReturn(tenantEntity);

        // Tenant is DEACTIVATED, not ACTIVE
        org.apache.custos.core.tenant.profile.api.Tenant protoTenant =
                org.apache.custos.core.tenant.profile.api.Tenant.newBuilder()
                        .setTenantId(TENANT_ID)
                        .setTenantStatus(TenantStatus.DEACTIVATED)
                        .build();
        GetTenantResponse inactiveResponse = GetTenantResponse.newBuilder()
                .setTenant(protoTenant)
                .build();
        when(tenantProfileService.getTenant(any(GetTenantRequest.class)))
                .thenReturn(inactiveResponse);

        // Act
        Optional<AuthClaim> result = tokenAuthorizer.authorize("some-base64-token");

        // Assert
        assertThat(result).isEmpty();
    }

    /**
     * When the credential store returns no credentials, authorize() should return empty.
     */
    @Test
    void authorize_withEmptyCredentials_shouldReturnEmpty() {
        GetAllCredentialsResponse emptyResponse = GetAllCredentialsResponse.newBuilder().build();
        when(credentialStoreService.getAllCredentialsFromToken(any(TokenRequest.class)))
                .thenReturn(emptyResponse);

        // Act
        Optional<AuthClaim> result = tokenAuthorizer.authorize("some-base64-token");

        // Assert
        assertThat(result).isEmpty();
    }

    private GetAllCredentialsResponse buildCredentialsResponse(long ownerId) {
        CredentialMetadata custosMetadata = CredentialMetadata.newBuilder()
                .setOwnerId(ownerId)
                .setId(CLIENT_ID)
                .setSecret(CLIENT_SECRET)
                .setType(Type.CUSTOS)
                .build();

        return GetAllCredentialsResponse.newBuilder()
                .addSecretList(custosMetadata)
                .build();
    }

    private Tenant buildTenantEntity(long tenantId, long parentId) {
        Tenant entity = new Tenant();
        entity.setId(tenantId);
        entity.setParentId(parentId);
        entity.setName("test-tenant");
        entity.setStatus("ACTIVE");
        entity.setDomain("test.org");
        entity.setRequesterEmail("admin@test.org");
        entity.setAdminFirstName("Admin");
        entity.setAdminLastName("User");
        entity.setAdminEmail("admin@test.org");
        entity.setAdminUsername("admin");
        entity.setScope("openid");
        entity.setApplicationType("web");
        return entity;
    }

    private GetTenantResponse buildActiveTenantResponse(long tenantId) {
        org.apache.custos.core.tenant.profile.api.Tenant protoTenant =
                org.apache.custos.core.tenant.profile.api.Tenant.newBuilder()
                        .setTenantId(tenantId)
                        .setTenantStatus(TenantStatus.ACTIVE)
                        .build();

        return GetTenantResponse.newBuilder()
                .setTenant(protoTenant)
                .build();
    }
}
