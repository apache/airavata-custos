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

package org.apache.custos.tenant.profile.persistance.model;


import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import jakarta.persistence.*;
import java.util.Date;
import java.util.Set;

/**
 * Represents tenant table at DB
 */
@Entity
@Table(name = "tenant")
@EntityListeners(AuditingEntityListener.class)
public class Tenant {

    @Id
    @GeneratedValue(strategy = GenerationType.SEQUENCE, generator = "tenant_id_generator")
    @SequenceGenerator(name = "tenant_id_generator", sequenceName = "tenant_sequence", initialValue = 10000000, allocationSize = 100)
    private Long id;

    @Column(nullable = false)
    private String name;

    @Column(nullable = false)
    private String domain;

    @Column(nullable = false)
    private String requesterEmail;

    @Column(nullable = false)
    private String status;

    @Column(nullable = false)
    private String adminFirstName;

    @Column(nullable = false)
    private String adminLastName;

    @Column(nullable = false)
    private String adminEmail;

    @Column(nullable = false)
    private String adminUsername;


    @Column(name = "tenant_uri")
    private String logoURI;

    @Column(nullable = false)
    private String scope;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date createdAt;

    private String uri;

    private String comment;

    private long parentId;

    @Column(nullable = false)
    private String applicationType;

    private String jwksUri;

    private String example_extension_parameter;

    private String tosUri;

    private String policyUri;

    private String jwks;

    private String softwareId;

    private String softwareVersion;

    private long refreshTokenLifetime = 0;


    @OneToMany(mappedBy = "tenant", cascade = CascadeType.ALL,orphanRemoval = true, fetch = FetchType.EAGER)
    private Set<Contact> contacts;

    @OneToMany(mappedBy = "tenant", cascade = CascadeType.ALL, orphanRemoval = true, fetch = FetchType.EAGER)
    private Set<RedirectURI> redirectURIS;

    @OneToMany(mappedBy = "tenant", cascade = CascadeType.ALL)
    private Set<TenantAttributeUpdateMetadata> attributeUpdateMetadata;

    @OneToMany(mappedBy = "tenant", cascade = CascadeType.ALL)
    private Set<TenantStatusUpdateMetadata> statusUpdateMetadata;


    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDomain() {
        return domain;
    }

    public void setDomain(String domain) {
        this.domain = domain;
    }

    public String getRequesterEmail() {
        return requesterEmail;
    }

    public void setRequesterEmail(String requesterEmail) {
        this.requesterEmail = requesterEmail;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getAdminFirstName() {
        return adminFirstName;
    }

    public void setAdminFirstName(String adminFirstName) {
        this.adminFirstName = adminFirstName;
    }

    public String getAdminLastName() {
        return adminLastName;
    }

    public void setAdminLastName(String adminLastName) {
        this.adminLastName = adminLastName;
    }

    public String getAdminEmail() {
        return adminEmail;
    }

    public void setAdminEmail(String adminEmail) {
        this.adminEmail = adminEmail;
    }


    public String getLogoURI() {
        return logoURI;
    }

    public void setLogoURI(String logoURI) {
        this.logoURI = logoURI;
    }

    public String getScope() {
        return scope;
    }

    public void setScope(String scope) {
        this.scope = scope;
    }

    public Set<Contact> getContacts() {
        return contacts;
    }

    public void setContacts(Set<Contact> contacts) {
        this.contacts = contacts;
    }

    public Set<RedirectURI> getRedirectURIS() {
        return redirectURIS;
    }

    public void setRedirectURIS(Set<RedirectURI> redirectURIS) {
        this.redirectURIS = redirectURIS;
    }

    public Set<TenantAttributeUpdateMetadata> getAttributeUpdateMetadata() {
        return attributeUpdateMetadata;
    }

    public void setAttributeUpdateMetadata(Set<TenantAttributeUpdateMetadata> attributeUpdateMetadata) {
        this.attributeUpdateMetadata = attributeUpdateMetadata;
    }

    public Set<TenantStatusUpdateMetadata> getStatusUpdateMetadata() {
        return statusUpdateMetadata;
    }

    public void setStatusUpdateMetadata(Set<TenantStatusUpdateMetadata> statusUpdateMetadata) {
        this.statusUpdateMetadata = statusUpdateMetadata;
    }

    public String getAdminUsername() {
        return adminUsername;
    }

    public void setAdminUsername(String adminUsername) {
        this.adminUsername = adminUsername;
    }


    public String getUri() {
        return uri;
    }

    public void setUri(String uri) {
        this.uri = uri;
    }

    public String getComment() {
        return comment;
    }

    public void setComment(String comment) {
        this.comment = comment;
    }

    public long getParentId() {
        return parentId;
    }

    public void setParentId(long parentId) {
        this.parentId = parentId;
    }

    public String getApplicationType() {
        return applicationType;
    }

    public void setApplicationType(String applicationType) {
        this.applicationType = applicationType;
    }


    public String getJwksUri() {
        return jwksUri;
    }

    public void setJwksUri(String jwksUri) {
        this.jwksUri = jwksUri;
    }

    public String getExample_extension_parameter() {
        return example_extension_parameter;
    }

    public void setExample_extension_parameter(String example_extension_parameter) {
        this.example_extension_parameter = example_extension_parameter;
    }

    public String getTosUri() {
        return tosUri;
    }

    public void setTosUri(String tosUri) {
        this.tosUri = tosUri;
    }

    public String getPolicyUri() {
        return policyUri;
    }

    public void setPolicyUri(String policyUri) {
        this.policyUri = policyUri;
    }

    public String getJwks() {
        return jwks;
    }

    public void setJwks(String jwks) {
        this.jwks = jwks;
    }

    public String getSoftwareId() {
        return softwareId;
    }

    public void setSoftwareId(String softwareId) {
        this.softwareId = softwareId;
    }

    public String getSoftwareVersion() {
        return softwareVersion;
    }

    public void setSoftwareVersion(String softwareVersion) {
        this.softwareVersion = softwareVersion;
    }

    public long getRefreshTokenLifetime() {
        return refreshTokenLifetime;
    }

    public void setRefreshTokenLifetime(long refreshTokenLifetime) {
        this.refreshTokenLifetime = refreshTokenLifetime;
    }

    public Date getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Date createdAt) {
        this.createdAt = createdAt;
    }
}
