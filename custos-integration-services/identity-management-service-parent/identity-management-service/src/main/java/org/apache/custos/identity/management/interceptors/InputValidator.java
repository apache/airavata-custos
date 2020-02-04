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

package org.apache.custos.identity.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.identity.management.service.AuthorizationRequest;
import org.apache.custos.identity.management.utils.Constants;
import org.apache.custos.integration.core.exceptions.MissingParameterException;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

@Component
public class InputValidator implements IntegrationServiceInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(InputValidator.class);

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {
        switch (method) {
            case "authorize":
                validateAuthorize(headers, msg, method);
                break;
            default:
        }
        return msg;
    }


    private boolean validateAuthorize(Metadata headers, Object body, String method) {
        validationAuthorizationHeader(headers);

        if (body == null || !(body instanceof AuthorizationRequest)) {

            AuthorizationRequest request = (AuthorizationRequest) body;

            if (request.getResponseType() == null || !request.getResponseType().trim().equals(Constants.AUTHORIZATION_CODE)) {
                throw new MissingParameterException("Incorrect response type", null);
            }
            if (request.getRedirectUri() == null || request.getRedirectUri().trim().equals("")) {
                throw new MissingParameterException("redirectUri is not available", null);
            }
            if (request.getScope() == null || request.getScope().trim().equals("")) {
                throw new MissingParameterException("scope is not available", null);
            }


        }
        return true;
    }


    private boolean validationAuthorizationHeader(Metadata headers) {
        if (headers.get(Metadata.Key.of(Constants.AUTHORIZATION_HEADER, Metadata.ASCII_STRING_MARSHALLER)) == null) {
            throw new MissingParameterException("authorization header not available", null);
        }

        return true;
    }

}
