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

package org.apache.custos.agent.profile.validator;

import org.apache.custos.core.services.commons.Validator;

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


    }

    private boolean validateAddTenant(Object obj) {

        return true;
    }


    private boolean validateUpdateTenant(Object obj) {

        return true;
    }


    private boolean validateGetAllTenantsForUser(Object obj) {

        return true;
    }

    private boolean validateGetTenant(Object obj) {

        return true;
    }


    private boolean validateGetMetadata(Object obj) {

        return true;

    }


    private boolean validateGetAllTenants(Object obj) {

        return true;

    }


    private boolean validateUpdateTenantStatus(Object obj) {

        return true;


    }


}
