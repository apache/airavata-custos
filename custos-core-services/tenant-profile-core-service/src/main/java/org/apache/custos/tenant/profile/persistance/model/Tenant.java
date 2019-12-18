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

import javax.persistence.*;
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
    private Long Id;

    @Column(nullable = false)
    private String name;

    private String domain; //check meaning of this

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
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date requestedTime;

    @Column(nullable = false)
    private String requesterUsername;

    @Column(name = "logo_uri")
    private String logoURI;


    private String scope;


    @OneToMany(mappedBy = "tenant", cascade = CascadeType.ALL, fetch = FetchType.EAGER)
    private Set<Contact> contacts;

    @OneToMany(mappedBy = "tenant", cascade = CascadeType.ALL, fetch = FetchType.EAGER)
    private Set<RedirectURI> redirectURIS;

    @OneToMany(mappedBy = "tenant", cascade = CascadeType.ALL)
    private Set<AttributeUpdateMetadata> attributeUpdateMetadata;

    @OneToMany(mappedBy = "tenant", cascade = CascadeType.ALL)
    private Set<StatusUpdateMetadata> statusUpdateMetadata;


    public Long getId() {
        return Id;
    }

    public void setId(Long id) {
        Id = id;
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


    public Date getRequestedTime() {
        return requestedTime;
    }

    public void setRequestedTime(Date requestedTime) {
        this.requestedTime = requestedTime;
    }

    public String getRequesterUsername() {
        return requesterUsername;
    }

    public void setRequesterUsername(String requesterUsername) {
        this.requesterUsername = requesterUsername;
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

    public Set<AttributeUpdateMetadata> getAttributeUpdateMetadata() {
        return attributeUpdateMetadata;
    }

    public void setAttributeUpdateMetadata(Set<AttributeUpdateMetadata> attributeUpdateMetadata) {
        this.attributeUpdateMetadata = attributeUpdateMetadata;
    }

    public Set<StatusUpdateMetadata> getStatusUpdateMetadata() {
        return statusUpdateMetadata;
    }

    public void setStatusUpdateMetadata(Set<StatusUpdateMetadata> statusUpdateMetadata) {
        this.statusUpdateMetadata = statusUpdateMetadata;
    }
}
