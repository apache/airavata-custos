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
package org.apache.custos.amie.model;

import jakarta.persistence.CascadeType;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.FetchType;
import jakarta.persistence.Id;
import jakarta.persistence.OneToMany;
import jakarta.persistence.Table;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/**
 * Maps to the 'persons' table. Stores a unique record for each person received from AMIE.
 */
@Entity
@Table(name = "persons")
public class PersonEntity {

    @Id
    @Column(name = "id")
    private String id;

    @Column(name = "access_global_id", nullable = false, unique = true)
    private String accessGlobalId;

    @Column(name = "first_name", nullable = false)
    private String firstName;

    @Column(name = "last_name", nullable = false)
    private String lastName;

    @Column(name = "email", nullable = false)
    private String email;

    @Column(name = "organization")
    private String organization;

    @Column(name = "org_code")
    private String orgCode;

    @Column(name = "nsf_status_code", length = 32)
    private String nsfStatusCode;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    @OneToMany(mappedBy = "person", cascade = CascadeType.ALL, orphanRemoval = true, fetch = FetchType.LAZY)
    private List<PersonDnsEntity> dnsEntries = new ArrayList<>();

    @OneToMany(mappedBy = "person", cascade = CascadeType.ALL, orphanRemoval = true, fetch = FetchType.LAZY)
    private List<ClusterAccountEntity> clusterAccounts = new ArrayList<>();

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getAccessGlobalId() {
        return accessGlobalId;
    }

    public void setAccessGlobalId(String accessGlobalId) {
        this.accessGlobalId = accessGlobalId;
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

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public String getOrganization() {
        return organization;
    }

    public void setOrganization(String organization) {
        this.organization = organization;
    }

    public String getOrgCode() {
        return orgCode;
    }

    public void setOrgCode(String orgCode) {
        this.orgCode = orgCode;
    }

    public String getNsfStatusCode() {
        return nsfStatusCode;
    }

    public void setNsfStatusCode(String nsfStatusCode) {
        this.nsfStatusCode = nsfStatusCode;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Instant createdAt) {
        this.createdAt = createdAt;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(Instant updatedAt) {
        this.updatedAt = updatedAt;
    }

    public List<PersonDnsEntity> getDnsEntries() {
        return dnsEntries;
    }

    public void setDnsEntries(List<PersonDnsEntity> dnsEntries) {
        this.dnsEntries = dnsEntries;
    }

    public List<ClusterAccountEntity> getClusterAccounts() {
        return clusterAccounts;
    }

    public void setClusterAccounts(List<ClusterAccountEntity> clusterAccounts) {
        this.clusterAccounts = clusterAccounts;
    }

    @Override
    public boolean equals(Object o) {
        if (o == null || getClass() != o.getClass()) return false;

        PersonEntity that = (PersonEntity) o;
        return id.equals(that.id) && accessGlobalId.equals(that.accessGlobalId);
    }

    @Override
    public int hashCode() {
        int result = id.hashCode();
        result = 31 * result + accessGlobalId.hashCode();
        return result;
    }
}

