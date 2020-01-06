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
import org.apache.custos.tenant.management.exceptions.MissingParameterException;
import org.apache.custos.tenant.management.service.CreateTenantRequest;
import org.apache.custos.tenant.profile.service.Tenant;
import org.springframework.stereotype.Component;

/**
 * This class validates the  requests
 */
@Component
public class InputValidator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param body
     * @return
     */
    public  void validate(String methodName, Object body, Metadata headers) {

        switch (methodName) {
            case "createTenant":
                validateCreateTenant(body, methodName);
                break;
            case "getCredentials":
                validateGetCredentials(headers, methodName);
                break;
            default:
        }
    }

    private  boolean validateCreateTenant(Object obj, String method) {
        if (obj instanceof CreateTenantRequest) {
            CreateTenantRequest request = (CreateTenantRequest) obj;
            Tenant tenant = request.getTenant();

            if (tenant.getTenantName() == null || tenant.getTenantName().trim() == "") {
                throw new MissingParameterException("Tenant name should not be null", null);
            }

            if (tenant.getDomain() == null || tenant.getDomain().trim() == "") {
                throw new MissingParameterException("Tenant domain should not be null", null);
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

            if (tenant.getAdminUsername() == null || tenant.getAdminUsername().trim() == "") {
                throw new MissingParameterException("Tenant admin username should not be null", null);
            }
            if (tenant.getAdminPassword() == null || tenant.getAdminPassword().trim() == "") {
                throw new MissingParameterException("Tenant admin username should not be null", null);
            }

            if (tenant.getRequesterEmail() == null || tenant.getRequesterEmail().trim() == "") {
                throw new MissingParameterException("Tenant requester email  should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method " + method);
        }
        return true;
    }

    private  boolean validateGetCredentials(Metadata headers, String method) {
        if (headers.get(Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER)) == null) {
            throw new MissingParameterException("authorization header not available", null);
        }

        return true;
    }


}
