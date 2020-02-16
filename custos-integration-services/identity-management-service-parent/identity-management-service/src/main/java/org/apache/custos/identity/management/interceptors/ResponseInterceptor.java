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

import io.grpc.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Set;

public class ResponseInterceptor implements ServerInterceptor {
    private static final Logger LOGGER = LoggerFactory.getLogger(ResponseInterceptor.class);

    @Override
    public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(ServerCall<ReqT, RespT> serverCall, Metadata metadata,
                                                                 ServerCallHandler<ReqT, RespT> serverCallHandler) {

        LOGGER.info("right now calling");
        return serverCallHandler.startCall(new ForwardingServerCall.SimpleForwardingServerCall<ReqT, RespT>(serverCall) {


            @Override
            public void sendHeaders(Metadata responseHeaders) {
                Metadata.Key<String> CUSTOM_HEADER_KEY =
                        Metadata.Key.of("Location", Metadata.ASCII_STRING_MARSHALLER);
                Metadata.Key<String> HEADER_KEY =
                        Metadata.Key.of("grpc-status", Metadata.ASCII_STRING_MARSHALLER);
                responseHeaders.put(CUSTOM_HEADER_KEY, "https://location.com");
                responseHeaders.put(HEADER_KEY, "302");
                LOGGER.info("Header Location" + "settet");

                Set<String> keys = responseHeaders.keys();

                for (String key : keys) {
                    LOGGER.info("Key " + key);
                    Metadata.Key<String> KEY =
                            Metadata.Key.of(key, Metadata.ASCII_STRING_MARSHALLER);
                    LOGGER.info("Value " + responseHeaders.get(KEY));
                }

                super.sendHeaders(responseHeaders);
            }

        }, metadata);
    }
}
