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

import org.apache.custos.commons.exceptions.ApplicationSettingsException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ServerSettings extends ApplicationSettings {

    private static final Logger log = LoggerFactory.getLogger(ServerSettings.class);
    // Iam Server Constants
    public static final String IAM_SERVER_URL = "iam.server.url";
    public static final String IAM_SERVER_SUPER_ADMIN_USERNAME = "iam.server.super.admin.username";
    public static final String IAM_SERVER_SUPER_ADMIN_PASSWORD = "iam.server.super.admin.password";

    public static int getCacheSize() throws ApplicationSettingsException {
        return Integer.valueOf(getSetting(Constants.IN_MEMORY_CACHE_SIZE));
    }
    public static String getAuthzCacheManagerClassName() throws ApplicationSettingsException {
        return getSetting(Constants.AUTHZ_CACHE_MANAGER_CLASS);
    }

    public static String getSecurityManagerClassName() throws ApplicationSettingsException {
        return getSetting(Constants.SECURITY_MANAGER_CLASS);
    }
    public static String getRemoteIDPServiceUrl() throws ApplicationSettingsException {
        return getSetting(ServerSettings.IAM_SERVER_URL);
    }
    public static boolean isAuthzCacheEnabled() throws ApplicationSettingsException {
        return Boolean.valueOf(getSetting(Constants.AUTHZ_CACHE_ENABLED));
    }
    public static  String getAuthenticationServerPort() throws ApplicationSettingsException {
        return getSetting(ServerSettings.AUTHENTICATION_SERVICE_SERVER_PORT);
    }
    public static  String getAuthenticationServerHost() throws ApplicationSettingsException {
        return getSetting(ServerSettings.AUTHENTICATION_SERVICE_SERVER_HOST);
    }


}
