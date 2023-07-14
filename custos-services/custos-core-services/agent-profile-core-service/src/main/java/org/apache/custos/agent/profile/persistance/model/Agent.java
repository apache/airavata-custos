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

package org.apache.custos.agent.profile.persistance.model;

import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import jakarta.persistence.*;
import java.util.Date;
import java.util.Set;

/**
 * Agent entity
 */
@jakarta.persistence.Entity
@Table(name = "agent_entity")
@EntityListeners(AuditingEntityListener.class)
public class Agent {

    @Id
    private String id;

    @Column(nullable = false)
    private Long tenantId;

    @Column(nullable = false)
    private String status;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date createdAt;


    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @LastModifiedDate
    private Date last_modified_at;


    private String agentId;

    @OneToMany(fetch = FetchType.EAGER, mappedBy = "agent", orphanRemoval = true, cascade = CascadeType.ALL)
    private Set<AgentRole> agentRole;

    @OneToMany(fetch = FetchType.EAGER, mappedBy = "agent", orphanRemoval = true, cascade = CascadeType.ALL)
    private Set<AgentAttribute> agentAttribute;


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

    public Date getLast_modified_at() {
        return last_modified_at;
    }

    public void setLast_modified_at(Date last_modified_at) {
        this.last_modified_at = last_modified_at;
    }

    public Set<AgentRole> getAgentRole() {
        return agentRole;
    }

    public void setAgentRole(Set<AgentRole> agentRole) {
        this.agentRole = agentRole;
    }

    public Set<AgentAttribute> getAgentAttribute() {
        return agentAttribute;
    }

    public void setAgentAttribute(Set<AgentAttribute> agentAttribute) {
        this.agentAttribute = agentAttribute;
    }

    public String getAgentId() {
        return agentId;
    }

    public void setAgentId(String agentId) {
        this.agentId = agentId;
    }
}
