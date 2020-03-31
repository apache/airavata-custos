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

package org.apache.custos.user.profile.persistance.model;

import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import javax.persistence.*;
import java.util.Date;
import java.util.Set;

/**
 * Group Entity
 */
@Entity
@Table(name = "group_entity")
@EntityListeners(AuditingEntityListener.class)
public class Group {

    @Id
    private String id;

    @Column(nullable = false)
    private Long tenantId;

    @Column(nullable = false)
    private String name;


    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date createdAt;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @LastModifiedDate
    private Date lastModifiedAt;

    @Column
    private String parentId;


    @OneToMany(fetch = FetchType.EAGER, mappedBy = "group", orphanRemoval = true, cascade = CascadeType.ALL)
    private Set<GroupRole> groupRole;

    @OneToMany(fetch = FetchType.EAGER, mappedBy = "group", orphanRemoval = true, cascade = CascadeType.ALL)
    private Set<GroupAttribute> groupAttribute;


    @OneToMany(mappedBy = "group", cascade = CascadeType.ALL)
    Set<GroupMembership> groupMemberships;


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

    public Set<GroupMembership> getGroupMemberships() {
        return groupMemberships;
    }

    public void setGroupMemberships(Set<GroupMembership> groupMemberships) {
        this.groupMemberships = groupMemberships;
    }

    public Date getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Date createdAt) {
        this.createdAt = createdAt;
    }


    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Date getLastModifiedAt() {
        return lastModifiedAt;
    }

    public void setLastModifiedAt(Date lastModifiedAt) {
        this.lastModifiedAt = lastModifiedAt;
    }

    public String getParentId() {
        return parentId;
    }

    public void setParentId(String parentId) {
        this.parentId = parentId;
    }

    public Set<GroupRole> getGroupRole() {
        return groupRole;
    }

    public void setGroupRole(Set<GroupRole> groupRole) {
        this.groupRole = groupRole;
    }

    public Set<GroupAttribute> getGroupAttribute() {
        return groupAttribute;
    }

    public void setGroupAttribute(Set<GroupAttribute> groupAttribute) {
        this.groupAttribute = groupAttribute;
    }
}
