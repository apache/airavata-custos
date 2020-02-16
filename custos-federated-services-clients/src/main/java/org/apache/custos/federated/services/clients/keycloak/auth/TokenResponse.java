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

package org.apache.custos.federated.services.clients.keycloak.auth;

/**
 * This class wraps all the attributes provided from the token response
 */
public class TokenResponse {
    private String accessToken;
    private double expiresIn;
    private double refreshExpiresIn;
    private String refreshToken;
    private String tokenType;
    private String idToken;
    private double notBeforePolicy;
    private String sessionState;
    private String scope;

    public TokenResponse(String accessToken, double expiresIn, double refreshExpiresIn, String refreshToken,
                         String tokenType, String idToken, double notBeforePolicy, String sessionState, String scope) {
        this.accessToken = accessToken;
        this.expiresIn = expiresIn;
        this.refreshExpiresIn = refreshExpiresIn;
        this.refreshToken = refreshToken;
        this.tokenType = tokenType;
        this.idToken = idToken;
        this.notBeforePolicy = notBeforePolicy;
        this.sessionState = sessionState;
        this.scope = scope;
    }

    public String getAccessToken() {
        return accessToken;
    }

    public void setAccessToken(String accessToken) {
        this.accessToken = accessToken;
    }

    public double getExpiresIn() {
        return expiresIn;
    }

    public void setExpiresIn(double expiresIn) {
        this.expiresIn = expiresIn;
    }

    public double getRefreshExpiresIn() {
        return refreshExpiresIn;
    }

    public void setRefreshExpiresIn(double refreshExpiresIn) {
        this.refreshExpiresIn = refreshExpiresIn;
    }

    public String getRefreshToken() {
        return refreshToken;
    }

    public void setRefreshToken(String refreshToken) {
        this.refreshToken = refreshToken;
    }

    public String getTokenType() {
        return tokenType;
    }

    public void setTokenType(String tokenType) {
        this.tokenType = tokenType;
    }

    public String getIdToken() {
        return idToken;
    }

    public void setIdToken(String idToken) {
        this.idToken = idToken;
    }

    public double getNotBeforePolicy() {
        return notBeforePolicy;
    }

    public void setNotBeforePolicy(double notBeforePolicy) {
        this.notBeforePolicy = notBeforePolicy;
    }

    public String getSessionState() {
        return sessionState;
    }

    public void setSessionState(String sessionState) {
        this.sessionState = sessionState;
    }

    public String getScope() {
        return scope;
    }

    public void setScope(String scope) {
        this.scope = scope;
    }
}
