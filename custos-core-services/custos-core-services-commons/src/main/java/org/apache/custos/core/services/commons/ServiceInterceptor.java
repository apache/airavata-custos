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

package org.apache.custos.core.services.commons;

import io.grpc.*;
import org.apache.custos.core.services.commons.exceptions.MissingParameterException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * This class intercepts incoming requests and forwarding for validation
 */
public class ServiceInterceptor implements ServerInterceptor {

    private final Logger LOGGER = LoggerFactory.getLogger(ServiceInterceptor.class);

    private Validator validator;

    public ServiceInterceptor(Validator validator) {
        this.validator = validator;
    }

    @Override
    public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(ServerCall<ReqT, RespT> serverCall,
                                                                 Metadata metadata, ServerCallHandler<ReqT, RespT> serverCallHandler) {

        String fullMethod = serverCall.getMethodDescriptor().getFullMethodName();
        String methodName = fullMethod.split("/")[1];

        LOGGER.debug("Calling method : " + serverCall.getMethodDescriptor().getFullMethodName());

        return new ForwardingServerCallListener.SimpleForwardingServerCallListener<ReqT>(serverCallHandler.startCall(serverCall, metadata)) {
            @Override
            public void onMessage(ReqT message) {
                try {
                    validator.validate(methodName, message);
                    super.onMessage(message);
                } catch (Exception ex) {
                    String msg = "Error while validating method " + methodName + " " + ex.getMessage();
                    LOGGER.error(msg);
                    if (ex instanceof MissingParameterException) {
                        serverCall.close(Status.FAILED_PRECONDITION.withDescription(msg), metadata);
                    } else {
                        serverCall.close(Status.UNKNOWN.withDescription(msg), metadata);
                    }
                }
            }

            @Override
            public void onHalfClose() {
                try {
                    super.onHalfClose();
                } catch (IllegalStateException e) {
                    LOGGER.debug(e.getMessage());
                } catch (Exception e) {
                    String msg = "Error while validating method " + methodName + " " + e.getMessage();
                    LOGGER.error(msg);
                    serverCall.close(Status.UNKNOWN.withDescription(msg), metadata);
                }
            }
        };
    }


}
