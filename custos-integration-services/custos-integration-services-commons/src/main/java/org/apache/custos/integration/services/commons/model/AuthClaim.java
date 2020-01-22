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

    private String ciLogonId;

    private String ciLogonSecret;

    private String custosId;

    private String custosSecret;

    private long custosIdIssuedAt;


    private long custosSecretExpiredAt;


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
}
