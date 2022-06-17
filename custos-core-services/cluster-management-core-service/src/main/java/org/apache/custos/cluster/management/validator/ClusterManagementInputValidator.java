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

package org.apache.custos.cluster.management.validator;

import org.apache.commons.lang.NotImplementedException;
import org.apache.custos.cluster.management.service.GetServerCertificateRequest;
import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.core.services.commons.exceptions.MissingParameterException;
import org.springframework.stereotype.Component;

/**
 * This class validates the  requests
 */
@Component
public class ClusterManagementInputValidator implements Validator {

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param obj
     * @return
     */
    public <ReqT> ReqT validate(String methodName, ReqT obj) {

        switch (methodName) {
            case "getCustosServerCertificate":
                validateGetCustosServerCertificate(obj, methodName);
                break;
            default:
                throw new NotImplementedException("UnImplemented method");

        }

      return obj;
    }

    private boolean validateGetCustosServerCertificate(Object obj, String method) {

        if (obj instanceof GetServerCertificateRequest) {
           return true;
        }
        return true;
    }

}
