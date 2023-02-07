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

package org.apache.custos.tenant.profile.validator;

import org.apache.custos.core.services.api.commons.Validator;
import org.apache.custos.tenant.profile.exceptions.MissingParameterException;
import org.apache.custos.tenant.profile.service.*;
import org.springframework.stereotype.Component;

/**
 * This class validates the  requests
 */
@Component
public class TenantProfileInputValidator implements Validator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public <ReqT> ReqT validate(String methodName, ReqT obj) {

        switch (methodName) {
            case "addTenant":
                validateAddTenant(obj);
                break;
            case "updateTenant":
                validateUpdateTenant(obj);
                break;
            case "getTenant":
                validateGetTenant(obj);
                break;
            case "getTenantAttributeUpdateAuditTrail":
            case "getTenantStatusUpdateAuditTrail":
                validateGetMetadata(obj);
                break;
            case "updateTenantStatus":
                validateUpdateTenantStatus(obj);
                break;
            case "getAllTenantsForUser":
                validateGetAllTenantsForUser(obj);
                break;
            case "getAllTenants":
                validateGetAllTenants(obj);
                break;
            default:

        }
       return obj;

    }

    private boolean validateAddTenant(Object obj) {
        if (obj instanceof Tenant) {
            Tenant tenant = (Tenant) obj;

            if (tenant.getClientName() == null || tenant.getClientName().trim() == "") {
                throw new MissingParameterException("Client name should not be null", null);
            }

            if (tenant.getDomain() == null || tenant.getDomain().trim() == "") {
                throw new MissingParameterException(" Domain should not be null", null);
            }


            if (tenant.getAdminFirstName() == null || tenant.getAdminFirstName().trim() == "") {
                throw new MissingParameterException(" Admin first name should not be null", null);
            }

            if (tenant.getAdminLastName() == null || tenant.getAdminLastName().trim() == "") {
                throw new MissingParameterException(" Admin last name should not be null", null);
            }

            if (tenant.getAdminEmail() == null || tenant.getAdminEmail().trim() == "") {
                throw new MissingParameterException(" Admin email should not be null", null);
            }

            if (tenant.getAdminUsername() == null || tenant.getAdminUsername().trim() == "") {
                throw new MissingParameterException(" Admin username should not be null", null);
            }

            if (tenant.getAdminPassword() == null || tenant.getAdminPassword().trim() == "") {
                throw new MissingParameterException(" Admin password should not be null", null);
            }

            if (tenant.getRequesterEmail() == null || tenant.getRequesterEmail().trim() == "") {
                throw new MissingParameterException("Admin requester email  should not be null", null);
            }

            if (tenant.getRedirectUrisCount() == 0) {
                throw new MissingParameterException("Redirect uris cannot be  should not be null", null);
            }

            if (tenant.getJwksCount() > 0 && (tenant.getJwksUri() != null && !tenant.getJwksUri().trim().equals(""))) {
                throw new RuntimeException("jwks and jwks both should not be present in a single request", null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method addTenant");
        }
        return true;
    }


    private boolean validateUpdateTenant(Object obj) {
        if (obj instanceof Tenant) {

            Tenant tenant = (Tenant) obj;


            if (tenant == null || tenant.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }


            if (tenant.getClientName() == null || tenant.getClientName().trim() == "") {
                throw new MissingParameterException("Client name should not be null", null);
            }

            if (tenant.getDomain() == null || tenant.getDomain().trim() == "") {
                throw new MissingParameterException("Tenant domain should not be null", null);
            }


            if (tenant.getAdminFirstName() == null || tenant.getAdminFirstName().trim() == "") {
                throw new MissingParameterException(" Admin first name should not be null", null);
            }

            if (tenant.getAdminLastName() == null || tenant.getAdminLastName().trim() == "") {
                throw new MissingParameterException(" Admin last name should not be null", null);
            }

            if (tenant.getAdminEmail() == null || tenant.getAdminEmail().trim() == "") {
                throw new MissingParameterException(" Admin email should not be null", null);
            }

            if (tenant.getAdminUsername() == null || tenant.getAdminUsername().trim() == "") {
                throw new MissingParameterException(" Admin username should not be null", null);
            }

            if (tenant.getRequesterEmail() == null || tenant.getRequesterEmail().trim() == "") {
                throw new MissingParameterException("Tenant requester email  should not be null", null);
            }

            if (tenant.getRedirectUrisCount() == 0) {
                throw new MissingParameterException("Redirect uris cannot be  should not be null", null);
            }

            if (tenant.getJwksCount() > 0 && (tenant.getJwksUri() != null && !tenant.getJwksUri().trim().equals(""))) {
                throw new RuntimeException("jwks and jwks both should not be present in a single request", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method updateTenant");
        }
        return true;
    }


    private boolean validateGetAllTenantsForUser(Object obj) {
        if (obj instanceof GetAllTenantsForUserRequest) {
            GetAllTenantsForUserRequest req = (GetAllTenantsForUserRequest) obj;

            if (req.getRequesterEmail() == null || req.getRequesterEmail().trim().equals("")) {
                throw new MissingParameterException("Requester email should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method updateTenant");
        }
        return true;
    }

    private boolean validateGetTenant(Object obj) {
        if (obj instanceof GetTenantRequest) {
            GetTenantRequest req = (GetTenantRequest) obj;

            if (req.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id  should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method getTenant");
        }
        return true;
    }


    private boolean validateGetMetadata(Object obj) {
        if (obj instanceof GetAuditTrailRequest) {
            GetAuditTrailRequest req = (GetAuditTrailRequest) obj;

            if (req.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id  should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method getMetadata");
        }
        return true;

    }


    private boolean validateGetAllTenants(Object obj) {
        if (obj instanceof GetTenantsRequest) {
            GetTenantsRequest req = (GetTenantsRequest) obj;

            if (req.getLimit() == 0 && (req.getRequesterEmail() == null ||  req.getRequesterEmail().equals(""))) {
                throw new MissingParameterException("Limit should greater than 0", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method getMetadata");
        }
        return true;

    }


    private boolean validateUpdateTenantStatus(Object obj) {
        if (obj instanceof UpdateStatusRequest) {
            UpdateStatusRequest req = (UpdateStatusRequest) obj;

            if (req.getTenantId() == 0) {
                throw new MissingParameterException("Tenant Id  should not be null", null);
            }

            if (req.getStatus() == null) {
                throw new MissingParameterException("Tenant Status  should not be null", null);
            }

            if (req.getUpdatedBy() == null || req.getUpdatedBy().trim().equals("")) {
                throw new MissingParameterException("Tenant UpdatedBy  should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method updateTenantStatus");
        }
        return true;


    }


}
