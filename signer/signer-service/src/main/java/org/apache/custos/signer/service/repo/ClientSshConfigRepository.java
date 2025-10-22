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
package org.apache.custos.signer.service.repo;

import org.apache.custos.signer.service.model.ClientSshConfigEntity;
import org.apache.custos.signer.service.model.ClientSshConfigId;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

/**
 * Repository for ClientSshConfigEntity operations.
 * Provides access to client configurations and security policies.
 */
@Repository
public interface ClientSshConfigRepository extends JpaRepository<ClientSshConfigEntity, ClientSshConfigId> {

    /**
     * Find client configuration by tenant and client ID
     */
    Optional<ClientSshConfigEntity> findByTenantIdAndClientId(String tenantId, String clientId);

    /**
     * Find all enabled client configurations for a tenant
     */
    List<ClientSshConfigEntity> findByTenantIdAndEnabledTrue(String tenantId);

    /**
     * Find all enabled client configurations
     */
    List<ClientSshConfigEntity> findByEnabledTrue();

    /**
     * Check if a client configuration exists and is enabled
     */
    @Query("SELECT COUNT(c) > 0 FROM ClientSshConfigEntity c WHERE c.tenantId = :tenantId AND c.clientId = :clientId AND c.enabled = true")
    boolean existsByTenantIdAndClientIdAndEnabledTrue(@Param("tenantId") String tenantId, @Param("clientId") String clientId);

    /**
     * Find client configurations by target host (for future 1:many expansion)
     */
    List<ClientSshConfigEntity> findByTargetHostAndEnabledTrue(String targetHost);
}
