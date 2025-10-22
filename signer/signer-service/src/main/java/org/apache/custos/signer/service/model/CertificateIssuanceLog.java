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
package org.apache.custos.signer.service.model;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import org.springframework.data.annotation.CreatedDate;

import java.time.LocalDateTime;

/**
 * Entity representing comprehensive audit log for certificate issuance.
 * Tracks all SSH certificate signing events for compliance and debugging.
 */
@Entity
@Table(name = "certificate_issuance_logs")
public class CertificateIssuanceLog {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "tenant_id", nullable = false)
    private String tenantId;

    @Column(name = "client_id", nullable = false)
    private String clientId;

    @Column(name = "serial_number", nullable = false, unique = true)
    private Long serialNumber;

    @Column(name = "key_id", nullable = false)
    private String keyId; // SHA256 of public key

    @Column(name = "principal", nullable = false)
    private String principal;

    @Column(name = "public_key_fingerprint", nullable = false)
    private String publicKeyFingerprint;

    @Column(name = "ca_fingerprint", nullable = false)
    private String caFingerprint;

    @Column(name = "valid_after", nullable = false)
    private LocalDateTime validAfter;

    @Column(name = "valid_before", nullable = false)
    private LocalDateTime validBefore;

    @CreatedDate
    @Column(name = "issued_at", nullable = false, updatable = false)
    private LocalDateTime issuedAt;

    @Column(name = "source_ip", length = 45)
    private String sourceIp; // Client IP from gRPC context (IPv4/IPv6)

    @Column(name = "user_access_token_hash")
    private String userAccessTokenHash; // Token hash for correlation

    @Column(name = "request_metadata", columnDefinition = "JSON")
    private String requestMetadata; // Additional context

    public CertificateIssuanceLog() {
    }

    public CertificateIssuanceLog(String tenantId, String clientId, Long serialNumber,
                                  String keyId, String principal, String publicKeyFingerprint,
                                  String caFingerprint, LocalDateTime validAfter, LocalDateTime validBefore) {
        this.tenantId = tenantId;
        this.clientId = clientId;
        this.serialNumber = serialNumber;
        this.keyId = keyId;
        this.principal = principal;
        this.publicKeyFingerprint = publicKeyFingerprint;
        this.caFingerprint = caFingerprint;
        this.validAfter = validAfter;
        this.validBefore = validBefore;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getTenantId() {
        return tenantId;
    }

    public void setTenantId(String tenantId) {
        this.tenantId = tenantId;
    }

    public String getClientId() {
        return clientId;
    }

    public void setClientId(String clientId) {
        this.clientId = clientId;
    }

    public Long getSerialNumber() {
        return serialNumber;
    }

    public void setSerialNumber(Long serialNumber) {
        this.serialNumber = serialNumber;
    }

    public String getKeyId() {
        return keyId;
    }

    public void setKeyId(String keyId) {
        this.keyId = keyId;
    }

    public String getPrincipal() {
        return principal;
    }

    public void setPrincipal(String principal) {
        this.principal = principal;
    }

    public String getPublicKeyFingerprint() {
        return publicKeyFingerprint;
    }

    public void setPublicKeyFingerprint(String publicKeyFingerprint) {
        this.publicKeyFingerprint = publicKeyFingerprint;
    }

    public String getCaFingerprint() {
        return caFingerprint;
    }

    public void setCaFingerprint(String caFingerprint) {
        this.caFingerprint = caFingerprint;
    }

    public LocalDateTime getValidAfter() {
        return validAfter;
    }

    public void setValidAfter(LocalDateTime validAfter) {
        this.validAfter = validAfter;
    }

    public LocalDateTime getValidBefore() {
        return validBefore;
    }

    public void setValidBefore(LocalDateTime validBefore) {
        this.validBefore = validBefore;
    }

    public LocalDateTime getIssuedAt() {
        return issuedAt;
    }

    public void setIssuedAt(LocalDateTime issuedAt) {
        this.issuedAt = issuedAt;
    }

    public String getSourceIp() {
        return sourceIp;
    }

    public void setSourceIp(String sourceIp) {
        this.sourceIp = sourceIp;
    }

    public String getUserAccessTokenHash() {
        return userAccessTokenHash;
    }

    public void setUserAccessTokenHash(String userAccessTokenHash) {
        this.userAccessTokenHash = userAccessTokenHash;
    }

    public String getRequestMetadata() {
        return requestMetadata;
    }

    public void setRequestMetadata(String requestMetadata) {
        this.requestMetadata = requestMetadata;
    }

    @Override
    public String toString() {
        return "CertificateIssuanceLog{" +
                "id=" + id +
                ", tenantId='" + tenantId + '\'' +
                ", clientId='" + clientId + '\'' +
                ", serialNumber=" + serialNumber +
                ", principal='" + principal + '\'' +
                ", issuedAt=" + issuedAt +
                '}';
    }
}
