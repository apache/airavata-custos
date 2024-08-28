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

package org.apache.custos.core.model.user;

import jakarta.persistence.CascadeType;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.EntityListeners;
import jakarta.persistence.FetchType;
import jakarta.persistence.Id;
import jakarta.persistence.OneToMany;
import jakarta.persistence.Table;
import jakarta.persistence.Temporal;
import jakarta.persistence.TemporalType;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import java.util.Date;
import java.util.Set;

@Entity
@Table(name = "user_profile")
@EntityListeners(AuditingEntityListener.class)
public class UserProfile {

    @Id
    private String id;

    @Column(nullable = false)
    private Long tenantId;

    @Column(nullable = false)
    private String username;

    @Column
    private String emailAddress;

    @Column
    private String firstName;

    @Column
    private String lastName;

    @Column(nullable = false)
    private String status;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date createdAt;


    @Temporal(TemporalType.TIMESTAMP)
    @LastModifiedDate
    private Date lastModifiedAt;


    @Column
    private String type;


    @OneToMany(fetch = FetchType.EAGER, mappedBy = "userProfile", orphanRemoval = true, cascade = CascadeType.ALL)
    private Set<UserRole> userRole;

    @OneToMany(fetch = FetchType.EAGER, mappedBy = "userProfile", orphanRemoval = true, cascade = CascadeType.ALL)
    private Set<UserAttribute> userAttribute;

    @OneToMany(mappedBy = "userProfile", cascade = CascadeType.ALL)
    private Set<AttributeUpdateMetadata> attributeUpdateMetadata;

    @OneToMany(mappedBy = "userProfile", cascade = CascadeType.ALL)
    private Set<StatusUpdateMetadata> statusUpdateMetadata;


    @OneToMany(mappedBy = "userProfile", cascade = CascadeType.ALL)
    private Set<UserGroupMembership> userGroupMemberships;


    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public Long getTenantId() {
        return tenantId;
    }

    public void setTenantId(Long tenantId) {
        this.tenantId = tenantId;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getEmailAddress() {
        return emailAddress;
    }

    public void setEmailAddress(String emailAddress) {
        this.emailAddress = emailAddress;
    }

    public String getFirstName() {
        return firstName;
    }

    public void setFirstName(String firstName) {
        this.firstName = firstName;
    }

    public String getLastName() {
        return lastName;
    }

    public void setLastName(String lastName) {
        this.lastName = lastName;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public Date getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Date createdAt) {
        this.createdAt = createdAt;
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

    public Set<UserAttribute> getUserAttribute() {
        return userAttribute;
    }

    public void setUserAttribute(Set<UserAttribute> userAttribute) {
        this.userAttribute = userAttribute;
    }

    public Set<UserRole> getUserRole() {
        return userRole;
    }

    public void setUserRole(Set<UserRole> userRole) {
        this.userRole = userRole;
    }

    public Set<UserGroupMembership> getUserGroupMemberships() {
        return userGroupMemberships;
    }

    public void setUserGroupMemberships(Set<UserGroupMembership> userGroupMemberships) {
        this.userGroupMemberships = userGroupMemberships;
    }

    public Date getLastModifiedAt() {
        return lastModifiedAt;
    }

    public void setLastModifiedAt(Date lastModifiedAt) {
        this.lastModifiedAt = lastModifiedAt;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }
}
