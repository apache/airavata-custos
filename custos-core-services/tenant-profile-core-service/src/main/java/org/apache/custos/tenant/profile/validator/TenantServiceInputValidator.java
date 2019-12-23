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

import org.apache.custos.tenant.profile.exceptions.MissingParameterException;
import org.apache.custos.tenant.profile.service.*;

/**
 * This class validates the  requests
 */
public class TenantServiceInputValidator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public static void  validate(String methodName, Object obj) {

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
            default:

        }


    }

    private static boolean validateAddTenant(Object obj) {
        if (obj instanceof Tenant) {
            Tenant tenant = (Tenant) obj;

            if (tenant.getTenantName() == null || tenant.getTenantName().trim() == "") {
                throw new MissingParameterException("Tenant name should not be null", null);
            }

            if (tenant.getDomain() == null || tenant.getDomain().trim() == "") {
                throw new MissingParameterException("Tenant domain should not be null", null);
            }

            if (tenant.getRequesterUsername() == null || tenant.getRequesterUsername().trim() == "") {
                throw new MissingParameterException("Tenant requester username should not be null", null);
            }

            if (tenant.getAdminFirstName() == null || tenant.getAdminFirstName().trim() == "") {
                throw new MissingParameterException("Tenant admin first name should not be null", null);
            }

            if (tenant.getAdminLastName() == null || tenant.getAdminLastName().trim() == "") {
                throw new MissingParameterException("Tenant admin last name should not be null", null);
            }

            if (tenant.getAdminEmail() == null || tenant.getAdminEmail().trim() == "") {
                throw new MissingParameterException("Tenant admin email should not be null", null);
            }

            if (tenant.getRequesterEmail() == null || tenant.getRequesterEmail().trim() == "") {
                throw new MissingParameterException("Tenant requester email  should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method addTenant");
        }
        return true;
    }


    private static boolean validateUpdateTenant(Object obj) {
        if (obj instanceof UpdateTenantRequest) {
            UpdateTenantRequest req = (UpdateTenantRequest) obj;

            Tenant tenant = req.getTenant();
            String updatedBy = req.getUpdatedBy();

            if (tenant == null || tenant.getTenantId() == null || tenant.getTenantId().trim() == "") {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (updatedBy == null || updatedBy.trim().equals("")) {
                throw new MissingParameterException("Updated by should not be null", null);
            }

            if (tenant.getTenantName() == null || tenant.getTenantName().trim() == "") {
                throw new MissingParameterException("Tenant name should not be null", null);
            }

            if (tenant.getDomain() == null || tenant.getDomain().trim() == "") {
                throw new MissingParameterException("Tenant domain should not be null", null);
            }

            if (tenant.getRequesterUsername() == null || tenant.getRequesterUsername().trim() == "") {
                throw new MissingParameterException("Tenant requester username should not be null", null);
            }

            if (tenant.getAdminFirstName() == null || tenant.getAdminFirstName().trim() == "") {
                throw new MissingParameterException("Tenant admin first name should not be null", null);
            }

            if (tenant.getAdminLastName() == null || tenant.getAdminLastName().trim() == "") {
                throw new MissingParameterException("Tenant admin last name should not be null", null);
            }

            if (tenant.getAdminEmail() == null || tenant.getAdminEmail().trim() == "") {
                throw new MissingParameterException("Tenant admin email should not be null", null);
            }

            if (tenant.getRequesterEmail() == null || tenant.getRequesterEmail().trim() == "") {
                throw new MissingParameterException("Tenant requester email  should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method updateTenant");
        }
        return true;
    }


    private static boolean validateGetAllTenantsForUser(Object obj) {
        if (obj instanceof GetAllTenantsForUserRequest) {
            GetAllTenantsForUserRequest req = (GetAllTenantsForUserRequest) obj;

            if (req.getRequesterUserName() == null || req.getRequesterUserName().trim().equals("")) {
                throw new MissingParameterException("Requester username should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method updateTenant");
        }
        return true;
    }

    private static boolean validateGetTenant(Object obj) {
        if (obj instanceof GetTenantRequest) {
            GetTenantRequest req = (GetTenantRequest) obj;

            if (req.getTenantId() == null || req.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id  should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method getTenant");
        }
        return true;
    }


    private static boolean validateGetMetadata(Object obj) {
        if (obj instanceof GetAuditTrailRequest) {
            GetAuditTrailRequest req = (GetAuditTrailRequest) obj;

            if (req.getTenantId() == null || req.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id  should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method getMetadata");
        }
        return true;

    }


    private static boolean validateUpdateTenantStatus(Object obj) {
        if (obj instanceof UpdateStatusRequest) {
            UpdateStatusRequest req = (UpdateStatusRequest) obj;

            if (req.getTenantId() == null || req.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id  should not be null", null);
            }

            if (req.getStatus() == null || req.getStatus().trim().equals("")) {
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
