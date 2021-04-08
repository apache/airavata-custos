/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.sharing.persistance.model;

import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import javax.persistence.Entity;
import javax.persistence.*;
import java.util.Date;

@Entity
@Table(name = "sharing")
@EntityListeners(AuditingEntityListener.class)
public class Sharing {

    @Id
    @Column(length = 1000)
    private String id;

    @Column(nullable = false)
    private String associatingId;

    @Column(nullable = false)
    private String associatingIdType;

    @Column(nullable = false)
    private String sharingType;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date createdAt;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @LastModifiedDate
    private Date lastModifiedAt;


    @Column(nullable = false)
    private long tenantId;


    @JoinColumn(name = "permission_type_id")
    @ManyToOne
    private PermissionType permissionType;


    @JoinColumn(name = "entity_id")
    @ManyToOne
    private org.apache.custos.sharing.persistance.model.Entity entity;


    @JoinColumn(name = "inherited_parent_id")
    @ManyToOne
    private org.apache.custos.sharing.persistance.model.Entity inheritedParent;


    public String getAssociatingId() {
        return associatingId;
    }

    public void setAssociatingId(String associatingId) {
        this.associatingId = associatingId;
    }

    public String getSharingType() {
        return sharingType;
    }

    public void setSharingType(String sharingType) {
        this.sharingType = sharingType;
    }

    public Date getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Date createdAt) {
        this.createdAt = createdAt;
    }

    public Date getLastModifiedAt() {
        return lastModifiedAt;
    }

    public void setLastModifiedAt(Date lastModifiedAt) {
        this.lastModifiedAt = lastModifiedAt;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public long getTenantId() {
        return tenantId;
    }

    public void setTenantId(long tenantId) {
        this.tenantId = tenantId;
    }

    public PermissionType getPermissionType() {
        return permissionType;
    }

    public void setPermissionType(PermissionType permissionType) {
        this.permissionType = permissionType;
    }

    public org.apache.custos.sharing.persistance.model.Entity getEntity() {
        return entity;
    }

    public void setEntity(org.apache.custos.sharing.persistance.model.Entity entity) {
        this.entity = entity;
    }

    public org.apache.custos.sharing.persistance.model.Entity getInheritedParent() {
        return inheritedParent;
    }

    public void setInheritedParent(org.apache.custos.sharing.persistance.model.Entity inheritedParent) {
        this.inheritedParent = inheritedParent;
    }

    public String getAssociatingIdType() {
        return associatingIdType;
    }

    public void setAssociatingIdType(String associatingIdType) {
        this.associatingIdType = associatingIdType;
    }
}
