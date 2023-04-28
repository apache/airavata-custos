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

package org.apache.custos.integration.services.commons.model;

public class AuthClaim {

    private long tenantId;

    private String iamAuthId;

    private String iamAuthSecret;


    private String ciLogonId="";

    private String ciLogonSecret="";

    private String custosId;

    private String custosSecret;

    private long custosIdIssuedAt;


    private long custosSecretExpiredAt;

    private String performedBy;

    private boolean superTenant;

    private boolean admin;

    private String username;

    private String agentClientId;

    private String agentClientSecret;

    private String agentPassword;

    private String agentId;


    public AuthClaim() {
    }

    public AuthClaim(long tenantId, String iamAuthId, String iamAuthSecret) {
        this.tenantId = tenantId;
        this.iamAuthId = iamAuthId;
        this.iamAuthSecret = iamAuthSecret;
    }

    public long getTenantId() {
        return tenantId;
    }

    public void setTenantId(long tenantId) {
        this.tenantId = tenantId;
    }

    public String getIamAuthId() {
        return iamAuthId;
    }

    public void setIamAuthId(String iamAuthId) {
        this.iamAuthId = iamAuthId;
    }

    public String getIamAuthSecret() {
        return iamAuthSecret;
    }

    public void setIamAuthSecret(String iamAuthSecret) {
        this.iamAuthSecret = iamAuthSecret;
    }

    public String getCiLogonId() {
        return ciLogonId;
    }

    public void setCiLogonId(String ciLogonId) {
        this.ciLogonId = ciLogonId;
    }

    public String getCiLogonSecret() {
        return ciLogonSecret;
    }

    public void setCiLogonSecret(String ciLogonSecret) {
        this.ciLogonSecret = ciLogonSecret;
    }

    public String getCustosId() {
        return custosId;
    }

    public void setCustosId(String custosId) {
        this.custosId = custosId;
    }

    public String getCustosSecret() {
        return custosSecret;
    }

    public void setCustosSecret(String custosSecret) {
        this.custosSecret = custosSecret;
    }

    public long getCustosIdIssuedAt() {
        return custosIdIssuedAt;
    }

    public void setCustosIdIssuedAt(long custosIdIssuedAt) {
        this.custosIdIssuedAt = custosIdIssuedAt;
    }

    public long getCustosSecretExpiredAt() {
        return custosSecretExpiredAt;
    }

    public void setCustosSecretExpiredAt(long custosSecretExpiredAt) {
        this.custosSecretExpiredAt = custosSecretExpiredAt;
    }

    public String getPerformedBy() {
        return performedBy;
    }

    public void setPerformedBy(String performedBy) {
        this.performedBy = performedBy;
    }

    public boolean isSuperTenant() {
        return superTenant;
    }

    public void setSuperTenant(boolean superTenant) {
        this.superTenant = superTenant;
    }

    public boolean isAdmin() {
        return admin;
    }

    public void setAdmin(boolean admin) {
        this.admin = admin;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }


    public String getAgentClientId() {
        return agentClientId;
    }

    public void setAgentClientId(String agentClientId) {
        this.agentClientId = agentClientId;
    }

    public String getAgentClientSecret() {
        return agentClientSecret;
    }

    public void setAgentClientSecret(String agentClientSecret) {
        this.agentClientSecret = agentClientSecret;
    }

    public String getAgentPassword() {
        return agentPassword;
    }

    public void setAgentPassword(String agentPassword) {
        this.agentPassword = agentPassword;
    }

    public String getAgentId() {
        return agentId;
    }

    public void setAgentId(String agentId) {
        this.agentId = agentId;
    }
}
