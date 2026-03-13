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
import org.apache.custos.service.profile.TenantProfileService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Component;

import java.util.Arrays;
import java.util.List;

/**
 * SuperTenantBootstrapper is an {@link ApplicationRunner} that creates and activates the
 * super-tenant at application startup when {@code custos.bootstrap.enabled=true}.
 *
 * <p>If an ACTIVE super-tenant (parentId == 0) already exists,
 *  the bootstrap is skipped and the application starts normally.</p>
 *
 * <p>All admin credentials are sourced from Spring configuration properties, which in production
 * resolve from environment variables.</p>
 */
@Component
@ConditionalOnProperty(name = "custos.bootstrap.enabled", havingValue = "true")
public class SuperTenantBootstrapper implements ApplicationRunner {

    private static final Logger LOGGER = LoggerFactory.getLogger(SuperTenantBootstrapper.class);

    private static final int MAX_RETRIES = 30;
    private static final long RETRY_INTERVAL_MS = 5000;

    private final TenantManagementService tenantManagementService;
    private final TenantProfileService tenantProfileService;
    private final KeycloakClient keycloakClient;
    private final SuperTenantProperties properties;

    public SuperTenantBootstrapper(TenantManagementService tenantManagementService,
                                   TenantProfileService tenantProfileService,
                                   KeycloakClient keycloakClient,
                                   SuperTenantProperties properties) {
        this.tenantManagementService = tenantManagementService;
        this.tenantProfileService = tenantProfileService;
        this.keycloakClient = keycloakClient;
        this.properties = properties;
    }

    @Override
    public void run(ApplicationArguments args) {
        LOGGER.info("SuperTenantBootstrapper: checking whether super-tenant bootstrap is required");

        if (activeSuperTenantExists()) {
            LOGGER.info("SuperTenantBootstrapper: active super-tenant already exists, skipping bootstrap");
            return;
        }

        LOGGER.info("SuperTenantBootstrapper: no active super-tenant found, starting bootstrap");

        waitForKeycloak();

        List<String> redirectUris = parseRedirectUris(properties.redirectUris());

        Tenant tenantRequest = Tenant.newBuilder()
                .setClientName("Custos Super Tenant")
                .setRequesterEmail(properties.admin().email())
                .setAdminFirstName("CUSTOS")
                .setAdminLastName("ADMIN")
                .setAdminEmail(properties.admin().email())
                .setAdminUsername(properties.admin().username())
                .setAdminPassword(properties.admin().password())
                .addAllContacts(List.of(properties.admin().email()))
                .addAllRedirectUris(redirectUris)
                .setClientUri(properties.clientUri())
                .setScope(properties.scope())
                .setDomain(properties.domain())
                .setLogoUri(properties.clientUri())
                .setComment("Custos bootstrapping Tenant")
                .setApplicationType("web")
                .build();

        CreateTenantResponse createResponse = tenantManagementService.createTenant(tenantRequest);

        UpdateStatusRequest statusRequest = UpdateStatusRequest.newBuilder()
                .setClientId(createResponse.getClientId())
                .setStatus(TenantStatus.ACTIVE)
                .setSuperTenant(true)
                .setUpdatedBy(Constants.SYSTEM)
                .build();

        UpdateStatusResponse statusResponse = tenantManagementService.updateTenantStatus(statusRequest);

        LOGGER.info("SuperTenantBootstrapper: super-tenant bootstrapped successfully. Client ID: {}. "
                        + "Secret stored in Vault at /secret/{}/CUSTOS",
                createResponse.getClientId(),
                statusResponse.getTenantId());
    }

    private void waitForKeycloak() {
        for (int attempt = 1; attempt <= MAX_RETRIES; attempt++) {
            try {
                keycloakClient.realmExists("master");
                LOGGER.info("SuperTenantBootstrapper: Keycloak is ready");
                return;
            } catch (Exception ex) {
                LOGGER.info("SuperTenantBootstrapper: Keycloak not ready (attempt {}/{}), retrying in {}s...",
                        attempt, MAX_RETRIES, RETRY_INTERVAL_MS / 1000);
                try {
                    Thread.sleep(RETRY_INTERVAL_MS);
                } catch (InterruptedException ie) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("Bootstrap interrupted while waiting for Keycloak", ie);
                }
            }
        }
        throw new RuntimeException("Keycloak did not become available after " + MAX_RETRIES + " attempts");
    }

    /**
     * Queries the tenant profile store for any ACTIVE tenant with parentId == 0
     * Uses TenantType.ADMIN which filters by parent_id = 0.
     */
    boolean activeSuperTenantExists() {
        GetTenantsRequest request = GetTenantsRequest.newBuilder()
                .setType(TenantType.ADMIN)
                .setStatus(TenantStatus.ACTIVE)
                .setLimit(1)
                .setOffset(0)
                .build();

        GetAllTenantsResponse response = tenantProfileService.getAllTenants(request);
        return response != null && response.getTenantList() != null && !response.getTenantList().isEmpty();
    }

    /**
     * Parses a comma-separated redirect URI string into a trimmed list.
     */
    private List<String> parseRedirectUris(String redirectUrisValue) {
        if (redirectUrisValue == null || redirectUrisValue.isBlank()) {
            return List.of();
        }
        return Arrays.stream(redirectUrisValue.split(","))
                .map(String::trim)
                .filter(s -> !s.isEmpty())
                .toList();
    }
}
