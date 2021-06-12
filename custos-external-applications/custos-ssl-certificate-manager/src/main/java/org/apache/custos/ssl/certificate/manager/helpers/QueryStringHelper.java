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
 * specific language governing permissions and limitations
 * under the License.
 */

package org.apache.custos.ssl.certificate.manager.helpers;

import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;

/**
 * Helper class to build query strings for a url
 */
public class QueryStringHelper {
    private final StringBuilder query = new StringBuilder();

    public QueryStringHelper(String name, String value) {
        encode(name, value);
    }

    /**
     *
     * Adds a new name and value to query string
     *
     * @param name Name of the query
     * @param value Value of the query
     */
    public void add(String name, String value) {
        query.append("&");
        encode(name, value);
    }

    private void encode(String name, String value) {
        query.append(URLEncoder.encode(name, StandardCharsets.UTF_8));
        query.append("=");
        query.append(URLEncoder.encode(value, StandardCharsets.UTF_8));
    }

    /**
     *
     * Returns query string
     *
     * @return query string
     */
    public String toString() {
        return query.toString();
    }
}