/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.log.management.interceptors;


import io.grpc.Metadata;
import org.apache.custos.integration.core.exceptions.MissingParameterException;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

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
        validationAuthorizationHeader(headers);

    }


    private boolean validationAuthorizationHeader(Metadata headers) {
        if (headers.get(Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER)) == null
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
