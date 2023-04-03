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

package org.apache.custos.sharing.core.persistance.model;

import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import jakarta.persistence.Basic;
import jakarta.persistence.Column;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.SequenceGenerator;
import jakarta.persistence.Table;
import jakarta.persistence.EntityListeners;
import jakarta.persistence.Temporal;
import jakarta.persistence.TemporalType;
import jakarta.persistence.CascadeType;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.OneToMany;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.Lob;
import java.sql.Blob;
import java.util.Date;
import java.util.Set;

@jakarta.persistence.Entity
@Table(name = "entity")
@EntityListeners(AuditingEntityListener.class)
public class Entity {
    @Id
    @Column(nullable = false)
    private String id;

    @Column(nullable = false)
    private String externalId;

    @Column
    private String tenantId;

    @Column
    private String externalParentId;

    @Column(nullable = false)
    private String ownerId;

    @Column(nullable = false)
    private String name;

    @Column
    private String description;

    @Column
    @Lob
    private String fullText;

    @Column
    private int sharedCount;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date createdAt;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @LastModifiedDate
    private Date lastModifiedAt;


    @Column(nullable = true)
    @Temporal(TemporalType.TIMESTAMP)
    private Date originalCreatedTime;


    @Lob
    @Column
    private Blob binaryData;


    @JoinColumn(name = "entity_type_id")
    @ManyToOne
    private EntityType entityType;

    @OneToMany(mappedBy = "entity", cascade = CascadeType.ALL)
    private Set<Sharing> sharingSet;

    @OneToMany(mappedBy = "inheritedParent", cascade = CascadeType.ALL)
    private Set<Sharing> inheritedSharing;


    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }


    public String getExternalParentId() {
        return externalParentId;
    }

    public void setExternalParentId(String externalParentId) {
        this.externalParentId = externalParentId;
    }

    public String getOwnerId() {
        return ownerId;
    }

    public void setOwnerId(String ownerId) {
        this.ownerId = ownerId;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getFullText() {
        return fullText;
    }

    public void setFullText(String fullText) {
        this.fullText = fullText;
    }

    public int getSharedCount() {
        return sharedCount;
    }

    public void setSharedCount(int sharedCount) {
        this.sharedCount = sharedCount;
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

    public Date getOriginalCreatedTime() {
        return originalCreatedTime;
    }

    public void setOriginalCreatedTime(Date originalCreatedTime) {
        this.originalCreatedTime = originalCreatedTime;
    }

    public Blob getBinaryData() {
        return binaryData;
    }

    public void setBinaryData(Blob binaryData) {
        this.binaryData = binaryData;
    }

    public String getTenantId() {
        return tenantId;
    }

    public void setTenantId(String tenantId) {
        this.tenantId = tenantId;
    }

    public EntityType getEntityType() {
        return entityType;
    }

    public void setEntityType(EntityType entityType) {
        this.entityType = entityType;
    }

    public Set<Sharing> getSharingSet() {
        return sharingSet;
    }

    public void setSharingSet(Set<Sharing> sharingSet) {
        this.sharingSet = sharingSet;
    }

    public Set<Sharing> getInheritedSharing() {
        return inheritedSharing;
    }

    public void setInheritedSharing(Set<Sharing> inheritedSharing) {
        this.inheritedSharing = inheritedSharing;
    }

    public String getExternalId() {
        return externalId;
    }

    public void setExternalId(String externalId) {
        this.externalId = externalId;
    }
}
