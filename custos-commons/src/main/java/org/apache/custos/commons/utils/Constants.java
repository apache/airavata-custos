/**
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package org.apache.custos.commons.utils;

/**
 * Constants used in Custos should go here.
 */
public final class Constants {
    public static final String IN_MEMORY_CACHE_SIZE = "in.memory.cache.size";
    public static final String AUTHZ_CACHE_MANAGER_CLASS = "authz.cache.manager.class";
    public static final String SECURITY_MANAGER_CLASS = "security.manager.class";
    public static final String AUTHZ_CACHE_ENABLED = "authz.cache.enabled";

    //Names of the attributes that could be passed in the AuthzToken's claims map.
    public static final String USER_NAME = "userName";
    public static final String GATEWAY_ID = "gatewayID";

    //constants in XACML authorization response.
    public static final String PERMIT = "Permit";
}
