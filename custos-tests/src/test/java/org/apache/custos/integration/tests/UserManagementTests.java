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

package org.apache.custos.integration.tests;

import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import org.apache.custos.iam.service.*;
import org.apache.custos.identity.management.client.IdentityManagementClient;
import org.apache.custos.tenant.manamgement.client.TenantManagementClient;
import org.apache.custos.user.management.client.UserManagementClient;
import org.junit.Assert;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.testng.annotations.AfterClass;
import org.testng.annotations.BeforeClass;
import org.testng.annotations.Parameters;
import org.testng.annotations.Test;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

/**
 * This class contains User management tests
 */
public class UserManagementTests {

    private static final Logger LOGGER = LoggerFactory.getLogger(UserManagementTests.class);

    private String LOG_SUFFIX = "...........................";


    private UserManagementClient userManagementClient;
    private IdentityManagementClient identityManagementClient;
    private TenantManagementClient tenantManagementClient;

    private String username;


    @Parameters({"server-host", "server-port", "client-id", "client-sec"})
    @BeforeClass(groups = {"user-management"})
    public void setup(String serverHost, String serverPort, String clientId, String clientSec) throws IOException {
        LOGGER.info("Initiating user management test cases  " + LOG_SUFFIX);

        userManagementClient = new UserManagementClient(serverHost, Integer.valueOf(serverPort), clientId, clientSec);
        identityManagementClient = new IdentityManagementClient(serverHost, Integer.valueOf(serverPort), clientId, clientSec);
        tenantManagementClient = new TenantManagementClient(serverHost, Integer.valueOf(serverPort), clientId, clientSec);


    }

    @Test(groups = {"user-management"})
    public void registerUser() {
        LOGGER.info("Executing registerUser test case");

        username = getAlphaNumericString(7);

        RegisterUserResponse response = userManagementClient.
                registerUser(username, "TestUserF",
                        "TestUserL", "abcdef",
                        username + "@gmail.com", false);

        Assert.assertTrue(response.getIsRegistered());

    }

    @Test(groups = {"user-management"}, dependsOnMethods = {"registerUser"})
    public void enableUser() {
        LOGGER.info("Executing enableUser test case");
        UserRepresentation representation = userManagementClient.enableUser(username);
        OperationStatus status = userManagementClient.isUserEnabled(username);
        Assert.assertEquals(representation.getFirstName(), "TestUserF");
        Assert.assertTrue(status.getStatus());
    }

    @Test(groups = {"user-management"}, dependsOnMethods = {"enableUser"})
    public void findUser() {
        LOGGER.info("Executing findUser test case");
        FindUsersResponse response = userManagementClient.findUser(username,
                null, null, null, 0, 2);
        Assert.assertTrue(response.getUsersCount() > 0);

    }

    @Parameters({"admin-username", "admin-password"})
    @Test(groups = {"user-management"}, dependsOnMethods = {"findUser"})
    public void addUserAttributes(String adminUsername, String adminPassword) {
        LOGGER.info("Executing addUserAttributes test case");

        Struct struct = identityManagementClient.getToken(null, null,
                adminUsername, adminPassword, null, "password");
        Value value = struct.getFieldsMap().get("access_token");
        List<String> arrayList = new ArrayList<>();
        arrayList.add("123456789");
        UserAttribute attribute = UserAttribute.newBuilder()
                .setKey("phone")
                .addAllValues(arrayList)
                .build();
        UserAttribute[] attributes = {attribute};
        String[] users = {username};
        OperationStatus status = userManagementClient.addUserAttributes(value.getStringValue(), attributes, users);
        Assert.assertTrue(status.getStatus());
        UserRepresentation representation = userManagementClient.getUser(username);

        List<UserAttribute> userAttributes = representation.getAttributesList();

        Assert.assertTrue(!userAttributes.isEmpty());
        boolean attributeAdded = false;
        for (UserAttribute userAttribute : userAttributes) {
            if (userAttribute.getKey().equals("phone") && userAttribute.getValuesList().contains("123456789")) {
                attributeAdded = true;
            }
        }

        Assert.assertTrue(attributeAdded);


    }

    @Parameters({"admin-username", "admin-password"})
    @Test(groups = {"user-management"}, dependsOnMethods = {"addUserAttributes"})
    public void deleteUserAttribute(String adminUsername, String adminPassword) {
        LOGGER.info("Executing deleteUserAttribute test case");
        Struct struct = identityManagementClient.getToken(null, null,
                adminUsername, adminPassword, null, "password");
        Value value = struct.getFieldsMap().get("access_token");
        List<String> arrayList = new ArrayList<>();
        arrayList.add("123456789");
        UserAttribute attribute = UserAttribute.newBuilder()
                .setKey("phone")
                .addAllValues(arrayList)
                .build();
        UserAttribute[] attributes = {attribute};
        String[] users = {username};
        OperationStatus status = userManagementClient.deleteUserAttributes(value.getStringValue(), attributes, users);
        Assert.assertTrue(status.getStatus());
        UserRepresentation representation = userManagementClient.getUser(username);

        List<UserAttribute> userAttributes = representation.getAttributesList();

        boolean attributeDeleted = true;
        for (UserAttribute userAttribute : userAttributes) {
            if (userAttribute.getKey().equals("phone") && userAttribute.getValuesList().contains("123456789")) {
                attributeDeleted = false;
            }
        }

        Assert.assertTrue(attributeDeleted);

    }

    @Parameters({"admin-username", "admin-password"})
    @Test(groups = {"user-management"}, dependsOnMethods = {"deleteUserAttribute"})
    public void addUserRoles(String adminUsername, String adminPassword) {
        LOGGER.info("Executing addUserRoles test case");

        Struct struct = identityManagementClient.getToken(null, null,
                adminUsername, adminPassword, null, "password");
        Value value = struct.getFieldsMap().get("access_token");

        String[] users = {username};
        String[] roles = {"testrole"};

        OperationStatus status = userManagementClient.addRolesToUsers(value.getStringValue(), roles, users, false);
        Assert.assertTrue(status.getStatus());
        UserRepresentation representation = userManagementClient.getUser(username);

        List<String> realmRolesList = representation.getRealmRolesList();
        Assert.assertTrue(!realmRolesList.isEmpty());
        boolean roleAdded = false;
        for (String role : realmRolesList) {
            if (role.equals("testrole")) {
                roleAdded = true;
            }
        }
        Assert.assertTrue(roleAdded);

    }


    @Parameters({"admin-username", "admin-password"})
    @Test(groups = {"user-management"}, dependsOnMethods = {"deleteUserAttribute"})
    public void deleteUserRoles(String adminUsername, String adminPassword) {
        LOGGER.info("Executing deleteUserRoles test case");
        Struct struct = identityManagementClient.getToken(null, null,
                adminUsername, adminPassword, null, "password");
        Value value = struct.getFieldsMap().get("access_token");

        String[] roles = {"testrole"};

        OperationStatus status = userManagementClient.deleteUserRoles(value.getStringValue(), new String[0], roles, username);
        Assert.assertTrue(status.getStatus());
        UserRepresentation representation = userManagementClient.getUser(username);

        List<String> realmRolesList = representation.getRealmRolesList();

        boolean roleDeleted = true;
        for (String role : realmRolesList) {
            if (role.equals("testrole")) {
                roleDeleted = false;
            }
        }
        Assert.assertTrue(roleDeleted);

    }

    @Parameters({"admin-username", "admin-password"})
    @Test(groups = {"user-management"}, dependsOnMethods = {"deleteUserRoles"})
    public void deleteUser(String adminUsername, String adminPassword) {
        LOGGER.info("Executing deleteUser test case");
        Struct struct = identityManagementClient.getToken(null, null,
                adminUsername, adminPassword, null, "password");
        Value value = struct.getFieldsMap().get("access_token");

        OperationStatus status = userManagementClient.deleteUser(value.getStringValue(), username);
        Assert.assertTrue(status.getStatus());

    }


    @AfterClass(groups = {"user-management"})
    void cleanup() {
        LOGGER.info("Completing user management tests " + LOG_SUFFIX);
        userManagementClient = null;
        identityManagementClient = null;
        tenantManagementClient = null;

    }


    static String getAlphaNumericString(int n) {

        // chose a Character random from this String
        String AlphaNumericString = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
                + "0123456789"
                + "abcdefghijklmnopqrstuvxyz";

        // create StringBuffer size of AlphaNumericString
        StringBuilder sb = new StringBuilder(n);

        for (int i = 0; i < n; i++) {

            // generate a random number between
            // 0 to AlphaNumericString variable length
            int index
                    = (int) (AlphaNumericString.length()
                    * Math.random());

            // add Character one by one in end of sb
            sb.append(AlphaNumericString
                    .charAt(index));
        }

        return sb.toString();
    }


}
