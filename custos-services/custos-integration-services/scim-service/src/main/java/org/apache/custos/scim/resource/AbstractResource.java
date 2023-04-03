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

package org.apache.custos.scim.resource;

import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.wso2.charon3.core.protocol.SCIMResponse;

import java.util.Map;

/**
 * Parent class of all resources
 */
public class AbstractResource {


    public ResponseEntity buildResponse(SCIMResponse response) {

        if (response != null) {
            ResponseEntity.BodyBuilder builder = ResponseEntity
                    .status(response.getResponseStatus());

            Map<String, String> headerMap = response.getHeaderParamMap();

            if (headerMap != null && !headerMap.isEmpty()) {
                HttpHeaders httpHeaders = new HttpHeaders();
                for (String key : headerMap.keySet()) {
                    httpHeaders.set(key, headerMap.get(key));
                }
                return builder.headers(httpHeaders).body(response.getResponseMessage());
            } else {
                return builder.build();
            }


        } else {
            ResponseEntity.BodyBuilder builder = ResponseEntity
                    .status(HttpStatus.ACCEPTED);
            return builder.build();

        }

    }

}
