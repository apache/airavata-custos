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

package org.apache.custos.clients.core;

import io.grpc.Metadata;

import java.io.File;
import java.util.Base64;

/**
 * Having utility methods to used by clients
 */
public class ClientUtils {


    /**
     * Creates authorization header
     *
     * @param clientId
     * @param clientSecret
     * @return
     */
    public static Metadata getAuthorizationHeader(String clientId, String clientSecret) {

        String credential = clientId + ":" + clientSecret;

        byte[] token = Base64.getEncoder().encode(credential.getBytes());

        String formattedString = new String(token);
        String headerStr = "Bearer " + formattedString;
        Metadata header = new Metadata();
        Metadata.Key<String> key = Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER);
        header.put(key, headerStr);
        return header;
    }

    /**
     * Provides file object
     *
     * @param classObj
     * @param fileName
     * @return
     */
    public static File getFile(Class classObj, final String fileName) {
        return new File(classObj
                .getResource(fileName)
                .getFile());
    }


}
