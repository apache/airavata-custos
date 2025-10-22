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

import org.apache.custos.signer.service.model.RevocationEvent;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;

/**
 * Repository for RevocationEvent operations.
 * Provides access to revocation events for KRL generation and compliance.
 */
@Repository
public interface RevocationEventRepository extends JpaRepository<RevocationEvent, Long> {

    /**
     * Find revocation events by serial number
     */
    List<RevocationEvent> findBySerialNumber(Long serialNumber);

    /**
     * Find revocation events by key ID
     */
    List<RevocationEvent> findByKeyId(String keyId);

    /**
     * Find revocation events by CA fingerprint
     */
    List<RevocationEvent> findByCaFingerprint(String caFingerprint);

    /**
     * Find revocation events by tenant and client
     */
    Page<RevocationEvent> findByTenantIdAndClientId(String tenantId, String clientId, Pageable pageable);

    /**
     * Find revocation events within a time range
     */
    @Query("SELECT r FROM RevocationEvent r WHERE r.revokedAt BETWEEN :startTime AND :endTime")
    List<RevocationEvent> findByRevokedAtBetween(@Param("startTime") LocalDateTime startTime,
                                                 @Param("endTime") LocalDateTime endTime);

    /**
     * Find revocation events by tenant, client, and time range
     */
    @Query("SELECT r FROM RevocationEvent r WHERE r.tenantId = :tenantId AND r.clientId = :clientId AND r.revokedAt BETWEEN :startTime AND :endTime")
    List<RevocationEvent> findByTenantIdAndClientIdAndRevokedAtBetween(@Param("tenantId") String tenantId,
                                                                       @Param("clientId") String clientId,
                                                                       @Param("startTime") LocalDateTime startTime,
                                                                       @Param("endTime") LocalDateTime endTime);

    /**
     * Find all revocation events for KRL generation (since last KRL update)
     */
    @Query("SELECT r FROM RevocationEvent r WHERE r.tenantId = :tenantId AND r.clientId = :clientId AND r.revokedAt > :since")
    List<RevocationEvent> findRevocationEventsForKrl(@Param("tenantId") String tenantId,
                                                     @Param("clientId") String clientId,
                                                     @Param("since") LocalDateTime since);

    /**
     * Count revocation events by CA fingerprint
     */
    long countByCaFingerprint(String caFingerprint);

    /**
     * Count revocation events by tenant and client
     */
    long countByTenantIdAndClientId(String tenantId, String clientId);

    /**
     * Check if a certificate is revoked by serial number
     */
    @Query("SELECT COUNT(r) > 0 FROM RevocationEvent r WHERE r.serialNumber = :serialNumber")
    boolean isCertificateRevoked(@Param("serialNumber") Long serialNumber);

    /**
     * Check if a certificate is revoked by key ID
     */
    @Query("SELECT COUNT(r) > 0 FROM RevocationEvent r WHERE r.keyId = :keyId")
    boolean isKeyRevoked(@Param("keyId") String keyId);

    /**
     * Check if a CA is revoked by fingerprint
     */
    @Query("SELECT COUNT(r) > 0 FROM RevocationEvent r WHERE r.caFingerprint = :caFingerprint")
    boolean isCaRevoked(@Param("caFingerprint") String caFingerprint);
}
