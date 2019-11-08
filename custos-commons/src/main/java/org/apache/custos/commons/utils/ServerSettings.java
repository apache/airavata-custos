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

    // Authentication service constants
    public static final String AUTHENTICATION_SERVICE_SERVER_PORT = "custos.authentication.server.port";
    public static final String AUTHENTICATION_SERVICE_SERVER_HOST = "custos.authentication.server.host";

    // Profile Service Constants
    public static final String PROFILE_SERVICE_SERVER_HOST = "profile.service.server.host";
    public static final String PROFILE_SERVICE_SERVER_PORT = "profile.service.server.port";

    // CI logon Constants
    private static final String CILOGON_CLIENT_ID = "cilogon.admin.client.id";
    private static final String CILOGON_CLIENT_SECRET = "cilogon.admin.client.secret";
    private static final String CILOGON_URL = "cilogon.server";

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
        return getSetting(IAM_SERVER_URL);
    }
    public static boolean isAuthzCacheEnabled() throws ApplicationSettingsException {
        return Boolean.valueOf(getSetting(Constants.AUTHZ_CACHE_ENABLED));
    }
    public static  String getAuthenticationServerPort() throws ApplicationSettingsException {
        return getSetting(AUTHENTICATION_SERVICE_SERVER_PORT);
    }
    public static  String getAuthenticationServerHost() throws ApplicationSettingsException {
        return getSetting(AUTHENTICATION_SERVICE_SERVER_HOST);
    }
    public static String getKeyStorePath() throws ApplicationSettingsException {
        return getSetting(Constants.KEYSTORE_PATH);
    }
    public static String getKeyStorePassword() throws ApplicationSettingsException {
        return getSetting(Constants.KEYSTORE_PASSWORD);
    }
    public static String getIamServerUrl() throws ApplicationSettingsException {
        return getSetting(ServerSettings.IAM_SERVER_URL);
    }
    public static String getTrustStorePath() throws ApplicationSettingsException {
        return getSetting(TRUST_STORE_PATH);
    }
    public static String getTrustStorePassword() throws ApplicationSettingsException {
        return getSetting(TRUST_STORE_PASSWORD);
    }
    public static String getIamServerSuperAdminUsername() throws ApplicationSettingsException{
        return getSetting(IAM_SERVER_SUPER_ADMIN_USERNAME);
    }
    public static String getIamServerSuperAdminPassword() throws ApplicationSettingsException{
        return getSetting(IAM_SERVER_SUPER_ADMIN_PASSWORD);
    }
    public static String getProfileServiceServerHost() throws ApplicationSettingsException {
        return getSetting(ServerSettings.PROFILE_SERVICE_SERVER_HOST);
    }
    public static String getProfileServiceServerPort() throws ApplicationSettingsException {
        return getSetting(ServerSettings.PROFILE_SERVICE_SERVER_PORT);
    }
    public static boolean isSharingTLSEnabled() throws ApplicationSettingsException {
        return Boolean.valueOf(getSetting(Constants.IS_SHARING_TLS_ENABLED));
    }
    public static int getTLSClientTimeout() throws ApplicationSettingsException {
        return Integer.valueOf(getSetting(Constants.TLS_CLIENT_TIMEOUT));
    }

    public static String getCILogonAdminClientId() throws ApplicationSettingsException{
        return getSetting(ServerSettings.CILOGON_CLIENT_ID);
    }

    public static String getCILogonAdminClientSecret() throws ApplicationSettingsException{
        return getSetting(ServerSettings.CILOGON_CLIENT_SECRET);
    }

    public static String getCILogonURL() throws ApplicationSettingsException{
        return getSetting(ServerSettings.CILOGON_URL);
    }
}
