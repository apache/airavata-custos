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

package org.apache.custos.identity.validator;


import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.core.services.commons.exceptions.MissingParameterException;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.identity.service.AuthenticationRequest;
import org.apache.custos.identity.service.Claim;
import org.apache.custos.identity.service.GetUserManagementSATokenRequest;


/**
 * This class validates the  requests
 */
public class InputValidator implements Validator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public void validate(String methodName, Object obj) {

        switch (methodName) {
            case "authenticate":
                validateAuthenticate(obj);
                break;
            case "isAuthenticate":
            case "getUser":
                validateAuthzToken(obj);
                break;
            case "getUserManagementServiceAccountAccessToken":
                validateGetUserManagementServiceAccountAccessToken(obj);
                break;
            default:
                throw new RuntimeException("Method not implemented");
        }

    }

    private boolean validateAuthenticate(Object obj) {
        if (obj instanceof AuthenticationRequest) {
            AuthenticationRequest request = (AuthenticationRequest) obj;
            if (request.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getClientId() == null || request.getClientId().trim().equals("")) {
                throw new MissingParameterException("Client Id should not be null", null);
            }
            if (request.getClientSecret() == null || request.getClientSecret().trim().equals("")) {
                throw new MissingParameterException("Client secret should not be null", null);
            }
            if (request.getUsername() == null || request.getUsername().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }
            if (request.getPassword() == null || request.getPassword().trim().equals("")) {
                throw new MissingParameterException(" password should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method setUPTenant");
        }
        return true;
    }

    private boolean validateAuthzToken(Object obj) {
        if (obj instanceof AuthToken) {
            AuthToken request = (AuthToken) obj;
            String username = null;
            String tenantId = null;

            for (Claim claim : request.getClaimsList()) {
                if (claim.getKey().equals("username")) {
                    username = claim.getValue();
                } else if (claim.getKey().equals("tenantId")) {
                    tenantId = claim.getValue();
                }
            }
            if (request.getAccessToken() == null || request.getAccessToken().trim().equals("")) {
                throw new MissingParameterException("Access token should not be null", null);
            }


            if (username == null || username.trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

            if (tenantId == null || tenantId.trim().equals("")) {
                throw new MissingParameterException("TenantId should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method isUsernameAvailable");
        }
        return true;
    }


    private boolean validateGetUserManagementServiceAccountAccessToken(Object obj) {
        if (obj instanceof GetUserManagementSATokenRequest) {
            GetUserManagementSATokenRequest request = (GetUserManagementSATokenRequest) obj;
            if (request.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getClientId() == null || request.getClientId().trim().equals("")) {
                throw new MissingParameterException("Client Id should not be null", null);
            }

            if (request.getClientSecret() == null || request.getClientSecret().trim().equals("")) {
                throw new MissingParameterException("Client secret should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method userAccess");
        }
        return true;
    }


}
