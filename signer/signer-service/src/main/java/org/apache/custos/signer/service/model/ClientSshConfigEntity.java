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
import jakarta.persistence.EntityListeners;
import jakarta.persistence.Id;
import jakarta.persistence.IdClass;
import jakarta.persistence.Table;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import java.time.LocalDateTime;

/**
 * Entity representing client SSH configuration and security policies.
 * Maps clientId to target system with 1:1 relationship (designed for future expansion to 1:many).
 */
@Entity
@Table(name = "client_ssh_configs")
@EntityListeners(AuditingEntityListener.class)
@IdClass(ClientSshConfigId.class)
public class ClientSshConfigEntity {

    @Id
    @Column(name = "tenant_id", nullable = false)
    private String tenantId;

    @Id
    @Column(name = "client_id", nullable = false)
    private String clientId;

    @Column(name = "client_secret", nullable = false, length = 512)
    private String clientSecret; // Encrypted client secret

    @Column(name = "target_host", nullable = false)
    private String targetHost; // SSH target hostname

    @Column(name = "target_port", nullable = false)
    private Integer targetPort = 22; // SSH port

    @Column(name = "max_ttl_seconds", nullable = false)
    private Integer maxTtlSeconds = 86400; // Max certificate TTL (24h)

    @Column(name = "allowed_key_types", nullable = false, columnDefinition = "JSON")
    private String allowedKeyTypes; // e.g., ["ed25519", "rsa", "ecdsa"]

    @Column(name = "source_address_restriction")
    private String sourceAddressRestriction; // CIDR or IP for source-address critical option

    @Column(name = "critical_options", columnDefinition = "JSON")
    private String criticalOptions; // Additional critical options

    @Column(name = "extensions", columnDefinition = "JSON")
    private String extensions; // Certificate extensions

    @Column(name = "enabled", nullable = false)
    private Boolean enabled = true; // Active/inactive flag

    @CreatedDate
    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @LastModifiedDate
    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;

    public ClientSshConfigEntity() {
    }

    public ClientSshConfigEntity(String tenantId, String clientId, String clientSecret,
                                 String targetHost, Integer targetPort) {
        this.tenantId = tenantId;
        this.clientId = clientId;
        this.clientSecret = clientSecret;
        this.targetHost = targetHost;
        this.targetPort = targetPort;
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

    public String getClientSecret() {
        return clientSecret;
    }

    public void setClientSecret(String clientSecret) {
        this.clientSecret = clientSecret;
    }

    public String getTargetHost() {
        return targetHost;
    }

    public void setTargetHost(String targetHost) {
        this.targetHost = targetHost;
    }

    public Integer getTargetPort() {
        return targetPort;
    }

    public void setTargetPort(Integer targetPort) {
        this.targetPort = targetPort;
    }

    public Integer getMaxTtlSeconds() {
        return maxTtlSeconds;
    }

    public void setMaxTtlSeconds(Integer maxTtlSeconds) {
        this.maxTtlSeconds = maxTtlSeconds;
    }

    public String getAllowedKeyTypes() {
        return allowedKeyTypes;
    }

    public void setAllowedKeyTypes(String allowedKeyTypes) {
        this.allowedKeyTypes = allowedKeyTypes;
    }

    public String getSourceAddressRestriction() {
        return sourceAddressRestriction;
    }

    public void setSourceAddressRestriction(String sourceAddressRestriction) {
        this.sourceAddressRestriction = sourceAddressRestriction;
    }

    public String getCriticalOptions() {
        return criticalOptions;
    }

    public void setCriticalOptions(String criticalOptions) {
        this.criticalOptions = criticalOptions;
    }

    public String getExtensions() {
        return extensions;
    }

    public void setExtensions(String extensions) {
        this.extensions = extensions;
    }

    public Boolean getEnabled() {
        return enabled;
    }

    public void setEnabled(Boolean enabled) {
        this.enabled = enabled;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }

    public LocalDateTime getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(LocalDateTime updatedAt) {
        this.updatedAt = updatedAt;
    }

    @Override
    public String toString() {
        return "ClientSshConfigEntity{" +
                "tenantId='" + tenantId + '\'' +
                ", clientId='" + clientId + '\'' +
                ", targetHost='" + targetHost + '\'' +
                ", targetPort=" + targetPort +
                ", enabled=" + enabled +
                '}';
    }
}

