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

package org.apache.custos.tenant.management.interceptors;


import io.grpc.Metadata;
import org.apache.custos.iam.service.AddRolesRequest;
import org.apache.custos.iam.service.RoleRepresentation;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.apache.custos.tenant.management.exceptions.MissingParameterException;
import org.apache.custos.tenant.management.service.DeleteTenantRequest;
import org.apache.custos.tenant.management.service.GetTenantRequest;
import org.apache.custos.tenant.management.service.UpdateTenantRequest;
import org.apache.custos.tenant.management.utils.Constants;
import org.apache.custos.tenant.profile.service.UpdateStatusRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * This class validates the  request input parameters
 */
@Component
public class InputValidator implements IntegrationServiceInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(InputValidator.class);

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param body
     * @return
     */
    private void validate(String methodName, Object body, Metadata headers) {

        switch (methodName) {
            case "getCredentials":
                validateGetCredentials(headers, methodName);
                break;
            case "getTenant":
                validateGetTenant(headers, body, methodName);
                break;
            case "updateTenant":
                validateUpdateTenant(headers, body, methodName);
                break;
            case "deleteTenant":
                validateDeleteTenant(headers, body, methodName);
                break;
            case "addTenantRoles":
                validateAddRoleToTenant(headers, body, methodName);
                break;
            case "addProtocolMapper":
            case "configureEventPersistence":
                validateAddProtocolMapper(headers, body, methodName);
                break;
            case "updateTenantStatus":
                validateUpdateTenantStatus(headers, body, methodName);
                break;
            case "addToCache":
            case "removeFromCache":
            case "getFromCache":
            case "getInstitutions":
            case "getTenantRoles":
                validationAuthorizationHeader(headers);
                break;
            default:
        }
    }


    private boolean validateGetCredentials(Metadata headers, String method) {
        return validationAuthorizationHeader(headers);

    }

    private boolean validateGetTenant(Metadata headers, Object body, String method) {
        validationAuthorizationHeader(headers);
        GetTenantRequest tenantRequest = ((GetTenantRequest) body);

        String clientId = tenantRequest.getClientId();

        if (clientId == null || clientId.trim().equals("")) {
            throw new MissingParameterException("client_id should not be null", null);
        }
        return true;
    }

    private boolean validateUpdateTenant(Metadata headers, Object body, String method) {
        validationAuthorizationHeader(headers);


        UpdateTenantRequest tenantRequest = ((UpdateTenantRequest) body);

        String clientId = tenantRequest.getClientId();

        if (clientId == null || clientId.trim().equals("")) {
            clientId = tenantRequest.getBody().getClientId();
        }

        if (clientId == null || clientId.trim().equals("")) {
            throw new MissingParameterException("client_id should not be null", null);
        }
        return true;
    }

    private boolean validateDeleteTenant(Metadata headers, Object body, String method) {
        validationAuthorizationHeader(headers);


        DeleteTenantRequest tenantRequest = ((DeleteTenantRequest) body);

        String clientId = tenantRequest.getClientId();


        if (clientId == null || clientId.trim().equals("")) {
            throw new MissingParameterException("client_id should not be null", null);
        }
        return true;
    }

    private boolean validateAddRoleToTenant(Metadata headers, Object body, String method) {
        validationAuthorizationHeader(headers);


        AddRolesRequest addRolesRequest = ((AddRolesRequest) body);

        List<RoleRepresentation> rolesList = addRolesRequest.getRolesList();

        if (rolesList == null || rolesList.isEmpty()) {
            throw new MissingParameterException("Roles should not be null", null);
        }

        for (RoleRepresentation roleRepresentation : rolesList) {
            if (roleRepresentation.getName() == null || roleRepresentation.getName().trim().equals("")) {
                throw new MissingParameterException("Roles name should not be null", null);
            }
            if (roleRepresentation.getDescription() == null || roleRepresentation.getDescription().trim().equals("")) {
                throw new MissingParameterException("Description should not be null", null);
            }
        }


        return true;
    }

    private boolean validateAddProtocolMapper(Metadata headers, Object body, String method) {
        validationAuthorizationHeader(headers);

        return true;
    }

    private boolean validateUpdateTenantStatus(Metadata headers, Object body, String method) {
        boolean isSuperTenant = ((UpdateStatusRequest) body).getSuperTenant();
        if (!isSuperTenant) {
            validationAuthorizationHeader(headers);
        }

        UpdateStatusRequest updateStatusRequest = ((UpdateStatusRequest) body);

        if (updateStatusRequest.getClientId() == null || updateStatusRequest.getClientId().trim().equals("")) {
            throw new MissingParameterException("Client Id should not be null", null);
        }
        return true;
    }


    private boolean validationAuthorizationHeader(Metadata headers) {
        if (headers.get(Metadata.Key.of(Constants.AUTHORIZATION_HEADER, Metadata.ASCII_STRING_MARSHALLER)) == null
                || headers.get(Metadata.Key.of("Authorization", Metadata.ASCII_STRING_MARSHALLER)) == null) {
            throw new MissingParameterException("authorization header not available", null);
        }

        return true;
    }


    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {
        validate(method, msg, headers);
        return msg;
    }
}
