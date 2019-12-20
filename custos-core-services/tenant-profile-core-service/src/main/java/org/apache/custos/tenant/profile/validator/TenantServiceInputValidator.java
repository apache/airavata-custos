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
import org.apache.custos.tenant.profile.service.Tenant;

/**
 * This class validates the  requests
 */
public class TenantServiceInputValidator {

    /**
     * Input parameter validater
     * @param methodName
     * @param obj
     * @return
     */
    public static boolean validate(String methodName, Object obj) {

        switch (methodName) {
            case "addTenant":
                validateAddTenant(obj);
            default:
                return true;
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

        }else {
            throw new RuntimeException("Unexpected input type for method addTenant");
        }
        return true;
    }


}
