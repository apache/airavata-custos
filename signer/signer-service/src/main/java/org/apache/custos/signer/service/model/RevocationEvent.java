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
 * Entity representing certificate and CA revocation events.
 * Tracks all revocation operations for compliance and KRL generation.
 */
@Entity
@Table(name = "revocation_events")
public class RevocationEvent {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "tenant_id", nullable = false)
    private String tenantId;

    @Column(name = "client_id", nullable = false)
    private String clientId;

    @Column(name = "serial_number")
    private Long serialNumber; // Revoke by serial (nullable for CA revocation)

    @Column(name = "key_id")
    private String keyId; // Revoke by key ID (nullable for CA revocation)

    @Column(name = "ca_fingerprint")
    private String caFingerprint; // Revoke all certs from this CA (nullable for single cert)

    @CreatedDate
    @Column(name = "revoked_at", nullable = false, updatable = false)
    private LocalDateTime revokedAt;

    @Column(name = "reason", nullable = false)
    private String reason;

    @Column(name = "revoked_by", nullable = false)
    private String revokedBy; // Admin or system identifier

    public RevocationEvent() {
    }

    public RevocationEvent(String tenantId, String clientId, String reason, String revokedBy) {
        this.tenantId = tenantId;
        this.clientId = clientId;
        this.reason = reason;
        this.revokedBy = revokedBy;
    }

    public static RevocationEvent forSerialNumber(String tenantId, String clientId,
                                                  Long serialNumber, String reason, String revokedBy) {
        RevocationEvent event = new RevocationEvent(tenantId, clientId, reason, revokedBy);
        event.setSerialNumber(serialNumber);
        return event;
    }

    public static RevocationEvent forKeyId(String tenantId, String clientId,
                                           String keyId, String reason, String revokedBy) {
        RevocationEvent event = new RevocationEvent(tenantId, clientId, reason, revokedBy);
        event.setKeyId(keyId);
        return event;
    }

    public static RevocationEvent forCaFingerprint(String tenantId, String clientId,
                                                   String caFingerprint, String reason, String revokedBy) {
        RevocationEvent event = new RevocationEvent(tenantId, clientId, reason, revokedBy);
        event.setCaFingerprint(caFingerprint);
        return event;
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

    public String getCaFingerprint() {
        return caFingerprint;
    }

    public void setCaFingerprint(String caFingerprint) {
        this.caFingerprint = caFingerprint;
    }

    public LocalDateTime getRevokedAt() {
        return revokedAt;
    }

    public void setRevokedAt(LocalDateTime revokedAt) {
        this.revokedAt = revokedAt;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }

    public String getRevokedBy() {
        return revokedBy;
    }

    public void setRevokedBy(String revokedBy) {
        this.revokedBy = revokedBy;
    }

    public RevocationType getRevocationType() {
        if (serialNumber != null) {
            return RevocationType.SERIAL_NUMBER;
        } else if (keyId != null) {
            return RevocationType.KEY_ID;
        } else if (caFingerprint != null) {
            return RevocationType.CA_FINGERPRINT;
        } else {
            throw new IllegalStateException("RevocationEvent must have at least one revocation identifier set");
        }
    }

    @Override
    public String toString() {
        return "RevocationEvent{" +
                "id=" + id +
                ", tenantId='" + tenantId + '\'' +
                ", clientId='" + clientId + '\'' +
                ", serialNumber=" + serialNumber +
                ", keyId='" + keyId + '\'' +
                ", caFingerprint='" + caFingerprint + '\'' +
                ", revokedAt=" + revokedAt +
                ", reason='" + reason + '\'' +
                ", revokedBy='" + revokedBy + '\'' +
                '}';
    }

    public enum RevocationType {
        SERIAL_NUMBER,
        KEY_ID,
        CA_FINGERPRINT
    }
}
