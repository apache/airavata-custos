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

package org.apache.custos.sharing.management.interceptors;


import io.grpc.Metadata;
import org.apache.custos.integration.core.exceptions.MissingParameterException;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.apache.custos.sharing.service.EntityRequest;
import org.apache.custos.sharing.service.SharingRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * This class validates the  request input parameters
 */
@Component
public class SharingManagementInputValidator implements IntegrationServiceInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(SharingManagementInputValidator.class);

    /**
     * Input parameter validater
     *
     * @param methodName
     * @param body
     * @return
     */
    private void validate(String methodName, Object body, Metadata headers) {


        if (methodName.equals("createEntity") || methodName.equals("updateEntity")) {
            validationAuthorizationHeader(headers);
            if (body instanceof EntityRequest) {
                EntityRequest request = (EntityRequest) body;

                if (request.getEntity() == null ||
                        request.getEntity().getOwnerId() == null || request.getEntity().getOwnerId().equals("")) {
                    throw new MissingParameterException("OwnerId not found", null);
                }

            } else {
                throw new MissingParameterException("Unexpected input type for method " + methodName, null);
            }


        } else if (methodName.equals("shareEntityWithUsers") || methodName.equals("revokeEntitySharingFromUsers")) {
            validationAuthorizationHeader(headers);
            if (body instanceof SharingRequest) {
                SharingRequest request = (SharingRequest) body;

                if (request.getOwnerIdList() == null ||
                        request.getOwnerIdList() == null || request.getOwnerIdList().size() == 0) {
                    throw new MissingParameterException("OwnerId not found", null);
                }

            } else {
                throw new MissingParameterException("Unexpected input type for method " + methodName, null);
            }

        } else if (methodName.equals("shareEntityWithGroups") || methodName.equals("revokeEntitySharingFromGroups")) {
            validationAuthorizationHeader(headers);
            if (body instanceof SharingRequest) {
                SharingRequest request = (SharingRequest) body;

                if (request.getOwnerIdList() == null ||
                        request.getOwnerIdList() == null || request.getOwnerIdList().size() == 0) {
                    throw new MissingParameterException("Group Id not found", null);
                }

            } else {
                throw new MissingParameterException("Unexpected input type for method " + methodName, null);
            }

        }

    }


    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {
        validate(method, msg, headers);
        return msg;
    }


    private boolean validationAuthorizationHeader(Metadata headers) {
        if (headers.get(Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER)) == null
                || headers.get(Metadata.Key.of("Authorization", Metadata.ASCII_STRING_MARSHALLER)) == null) {
            throw new MissingParameterException("authorization header not available", null);
        }

        return true;
    }

}
