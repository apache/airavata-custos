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

package org.apache.custos.messaging.persistance.model;

import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import jakarta.persistence.*;
import java.util.Date;
import java.util.Set;

/**
 * Represents Email Template Body Params
 */
@Entity
@Table(name = "email_template")
@EntityListeners(AuditingEntityListener.class)
public class EmailTemplate {

    @Id
    @GeneratedValue(strategy = GenerationType.SEQUENCE, generator = "email_template_id_generator")
    @SequenceGenerator(name = "email_template_id_generator", sequenceName = "tenant_sequence", initialValue = 1000000000, allocationSize = 100)
    private Long id;

    @Column(nullable = false)
    private long tenantId;

    @Column(nullable = false)
    private String subject;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @CreatedDate
    private Date createdAt;

    @Column(nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    @LastModifiedDate
    private Date lastModifiedAt;

    @Column(nullable = false)
    private String custosEvent;

    @Column(nullable = false)
    private boolean status;

    @Column(nullable = false)
    private String body;


    @OneToMany(fetch = FetchType.EAGER, mappedBy = "emailTemplate", orphanRemoval = true, cascade = CascadeType.ALL)
    private Set<EmailBodyParams> bodyParams;


    @OneToMany(fetch = FetchType.EAGER, mappedBy = "emailTemplate", orphanRemoval = true, cascade = CascadeType.ALL)
    private Set<EmailReceivers> emailReceivers;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public long getTenantId() {
        return tenantId;
    }

    public void setTenantId(long tenantId) {
        this.tenantId = tenantId;
    }

    public String getSubject() {
        return subject;
    }

    public void setSubject(String subject) {
        this.subject = subject;
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

    public String getCustosEvent() {
        return custosEvent;
    }

    public void setCustosEvent(String custosEvent) {
        this.custosEvent = custosEvent;
    }

    public Set<EmailBodyParams> getBodyParams() {
        return bodyParams;
    }

    public void setBodyParams(Set<EmailBodyParams> bodyParams) {
        this.bodyParams = bodyParams;
    }

    public Set<EmailReceivers> getEmailReceivers() {
        return emailReceivers;
    }

    public void setEmailReceivers(Set<EmailReceivers> emailReceivers) {
        this.emailReceivers = emailReceivers;
    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getBody() {
        return body;
    }

    public void setBody(String body) {
        this.body = body;
    }
}
