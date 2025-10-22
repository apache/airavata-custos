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

import org.apache.custos.signer.service.model.CertificateIssuanceLog;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

/**
 * Repository for CertificateIssuanceLog operations.
 * Provides access to certificate audit logs for compliance and debugging.
 */
@Repository
public interface CertificateIssuanceLogRepository extends JpaRepository<CertificateIssuanceLog, Long> {

    /**
     * Find certificate log by serial number
     */
    Optional<CertificateIssuanceLog> findBySerialNumber(Long serialNumber);

    /**
     * Find certificate logs by key ID
     */
    List<CertificateIssuanceLog> findByKeyId(String keyId);

    /**
     * Find certificate logs by CA fingerprint
     */
    List<CertificateIssuanceLog> findByCaFingerprint(String caFingerprint);

    /**
     * Find certificate logs by principal
     */
    List<CertificateIssuanceLog> findByPrincipal(String principal);

    /**
     * Find certificate logs by tenant and client
     */
    Page<CertificateIssuanceLog> findByTenantIdAndClientId(String tenantId, String clientId, Pageable pageable);

    /**
     * Find certificate logs within a time range
     */
    @Query("SELECT c FROM CertificateIssuanceLog c WHERE c.issuedAt BETWEEN :startTime AND :endTime")
    List<CertificateIssuanceLog> findByIssuedAtBetween(@Param("startTime") LocalDateTime startTime,
                                                       @Param("endTime") LocalDateTime endTime);

    /**
     * Find certificate logs by tenant, client, and time range
     */
    @Query("SELECT c FROM CertificateIssuanceLog c WHERE c.tenantId = :tenantId AND c.clientId = :clientId AND c.issuedAt BETWEEN :startTime AND :endTime")
    List<CertificateIssuanceLog> findByTenantIdAndClientIdAndIssuedAtBetween(@Param("tenantId") String tenantId,
                                                                             @Param("clientId") String clientId,
                                                                             @Param("startTime") LocalDateTime startTime,
                                                                             @Param("endTime") LocalDateTime endTime);

    /**
     * Find active certificates (not yet expired) by principal
     */
    @Query("SELECT c FROM CertificateIssuanceLog c WHERE c.principal = :principal AND c.validBefore > :now")
    List<CertificateIssuanceLog> findActiveCertificatesByPrincipal(@Param("principal") String principal,
                                                                   @Param("now") LocalDateTime now);

    /**
     * Count certificates issued by CA fingerprint
     */
    long countByCaFingerprint(String caFingerprint);

    /**
     * Count certificates issued by tenant and client
     */
    long countByTenantIdAndClientId(String tenantId, String clientId);
}
