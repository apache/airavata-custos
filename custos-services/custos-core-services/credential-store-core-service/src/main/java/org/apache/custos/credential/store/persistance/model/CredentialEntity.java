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

package org.apache.custos.credential.store.persistance.model;

import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import jakarta.persistence.*;
import java.util.Date;
import java.util.Set;

@Entity
@Table(name = "custos_credentials")
@EntityListeners(AuditingEntityListener.class)
public class CredentialEntity {

    @Column(nullable = false)
    private String clientId;


    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date issuedAt;

     @Column(nullable = false)
    private long clientSecretExpiredAt;

    @Id
    @Column(nullable = false)
    private long ownerId;

     @Column(nullable = false)
     private String type;

    @OneToMany(mappedBy = "credentialEntity", cascade = CascadeType.ALL)
    Set<AgentCredentialEntity> agentCredentialEntities;


    public String getClientId() {
        return clientId;
    }

    public void setClientId(String clientId) {
        this.clientId = clientId;
    }

    public Date getIssuedAt() {
        return issuedAt;
    }

    public void setIssuedAt(Date issuedAt) {
        this.issuedAt = issuedAt;
    }

    public long getClientSecretExpiredAt() {
        return clientSecretExpiredAt;
    }

    public void setClientSecretExpiredAt(long clientSecretExpiredAt) {
        this.clientSecretExpiredAt = clientSecretExpiredAt;
    }

    public long getOwnerId() {
        return ownerId;
    }

    public void setOwnerId(long ownerId) {
        this.ownerId = ownerId;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }


    public Set<AgentCredentialEntity> getAgentCredentialEntities() {
        return agentCredentialEntities;
    }

    public void setAgentCredentialEntities(Set<AgentCredentialEntity> agentCredentialEntities) {
        this.agentCredentialEntities = agentCredentialEntities;
    }
}
