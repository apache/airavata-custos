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
package org.apache.custos.profile.service.samples;

import org.apache.custos.client.profile.service.CustosProfileServiceClientFactory;
import org.apache.custos.commons.exceptions.CustosSecurityException;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.commons.utils.Constants;
import org.apache.custos.profile.model.user.*;
import org.apache.custos.profile.service.samples.utils.ProfileServiceClientUtil;
import org.apache.custos.profile.user.cpi.UserProfileService;
import org.apache.custos.security.manager.KeyCloakSecurityManager;
import org.apache.http.Consts;
import org.apache.http.NameValuePair;
import org.apache.http.client.entity.UrlEncodedFormEntity;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.message.BasicNameValuePair;
import org.apache.http.util.EntityUtils;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.concurrent.ThreadLocalRandom;

/**
 * Created by goshenoy on 3/23/17.
 */
public class UserProfileSample {

    private static final Logger logger = LoggerFactory.getLogger(UserProfileSample.class);
    private static UserProfileService.Client userProfileClient;
    private static UserProfile testUser = null;
    private static String testGatewayId = "default";
    private static String username = "default-admin";
    private static AuthzToken authzToken;


    /**
     * Performs the following operations in sequence:
     *  1. create new user
     *  2. find user created
     *  3. update created user's name
     *  4. find all users in gateway
     *  5. find created user by name
     *  6. delete created user
     *  7. check if user exists
     * @param args
     * @throws Exception
     */
    public static void main(String[] args) throws Exception {
        try {
            String profileServiceServerHost = ProfileServiceClientUtil.getProfileServiceServerHost();
            int profileServiceServerPort = ProfileServiceClientUtil.getProfileServiceServerPort();
            authzToken = initializeAuthzToken();
            userProfileClient = CustosProfileServiceClientFactory.createCustosUserProfileServiceClient(profileServiceServerHost, profileServiceServerPort);

            // test add-user-profile
            testUser = userProfileClient.addUserProfile(authzToken, getUserProfile(username));
            assert (testUser != null) : "User creation failed. Null userId returned!";
            System.out.println("User created with userId: " + testUser);

            // test find-user-profile
            UserProfile userProfile = userProfileClient.getUserProfileById(authzToken, testUser.getUserId(), testGatewayId);
            assert (userProfile != null) : "Could not find user with userId: " + testUser + ", and gatewayID: " + testGatewayId;
            System.out.println("UserProfile: " + userProfile);

            // test update-user-profile : update name
            userProfile = getUserProfile(testUser.getUserId());
            String newFName = userProfile.getFirstName().replaceAll("fname", "fname-updated");
            userProfile.setFirstName(newFName);
            UserProfile updatedUserProfile = userProfileClient.updateUserProfile(authzToken, userProfile);
            boolean updateSuccess = updatedUserProfile != null;
            assert (updateSuccess) : "User update with new firstName: [" + newFName + "], Failed!";
            System.out.println("User update with new firstName: [" + newFName + "], Successful!");

            // test get-all-userprofiles
            List<UserProfile> userProfileList = userProfileClient.getAllUserProfilesInGateway(authzToken, testGatewayId, 0, 5);
            assert (userProfileList != null && !userProfileList.isEmpty()) : "Failed to retrieve users for gateway!";
            System.out.println("Printing userList retrieved..");
            for (UserProfile userProfile1 : userProfileList) {
                System.out.println("\t [UserProfile] userId: " + userProfile1.getUserId());
            }

            // test delete-user-profile
            boolean deleteSuccess = userProfileClient.deleteUserProfile(authzToken, testUser.getUserId(), testGatewayId) != null;
            assert (deleteSuccess) : "Delete user failed for userId: " + testUser.getUserId();
            System.out.println("Successfully deleted user with userId: " + testUser.getUserId());

            // test-check-user-exist
            boolean userExists = userProfileClient.doesUserExist(authzToken, testUser.getUserId(), testGatewayId);
            assert (!userExists) : "User should not exist, but it does.";
            System.out.println("User was deleted, hence does not exist!");
            System.out.println("*** DONE ***");
        } catch (Exception ex) {
            ex.printStackTrace();
            logger.error("UserProfile client-sample Exception: " + ex, ex);
        }
    }

    private static UserProfile getUserProfile(String userId) {
        // get random value for userId
        int userIdValue = ThreadLocalRandom.current().nextInt(1000);
        
        // construct userProfile object
        UserProfile userProfile = new UserProfile();
        userProfile.setUserModelVersion("model-" + userIdValue);
        userProfile.setCustosInternalUserId("test-user-internal-" + userIdValue);
        userProfile.setUserId(userId);
        userProfile.setFirstName("test-user-fname");
        userProfile.setLastName("test-user-lname");
        userProfile.setGatewayId(testGatewayId);
        userProfile.addToEmails("test-user-" + userIdValue + "@domain1.com");
        userProfile.addToEmails("test-user-" + userIdValue + "@domain2.com");
        userProfile.setCreationTime(System.currentTimeMillis());
        userProfile.setLastAccessTime(System.currentTimeMillis());
        userProfile.setValidUntil(System.currentTimeMillis());
        userProfile.setState(Status.ACTIVE);
        userProfile.setNsfDemographics(getNSFDemographics(userIdValue));
        return userProfile;
    }

    private static NSFDemographics getNSFDemographics(int userIdValue) {
        // construct nsfdemographics object
        NSFDemographics nsfDemographics = new NSFDemographics();
        nsfDemographics.setCustosInternalUserId("test-user-internal-" + userIdValue);
        nsfDemographics.setGender("male");
        nsfDemographics.setUsCitizenship(USCitizenship.US_CITIZEN);
        nsfDemographics.addToEthnicities(ethnicity.NOT_HISPANIC_LATINO);
        nsfDemographics.addToRaces(race.AMERICAN_INDIAN_OR_ALASKAN_NATIVE);
        return nsfDemographics;
    }

    private static AuthzToken initializeAuthzToken() throws Exception{
         authzToken = new AuthzToken();
        {
            HashMap<String, String> map_ = new HashMap<>();
            map_.put(Constants.GATEWAY_ID, testGatewayId);
            map_.put(Constants.USER_NAME, username);
            JSONObject json = getAccessToken();
            if(json.has("access_token")) {
                authzToken.setAccessToken(json.get("access_token").toString());
            }
            else{
                throw new Exception("Cannot find access token for the user. Check if the user exists in the IDP");
            }
            authzToken.setClaimsMap(map_);
            return authzToken;
        }
    }
    private static JSONObject getAccessToken() throws CustosSecurityException {
        new KeyCloakSecurityManager().initializeSecurityInfra();
        String password = "123456";
        String tokenEndPoint = "https://airavata.host:8443/auth/realms/default/protocol/openid-connect/token";
        String grant_type = "password";
        String client_id = "admin-cli";

        CloseableHttpClient httpClient = HttpClients.createSystem();
        HttpPost httpPost = new HttpPost(tokenEndPoint);
        ;
        List<NameValuePair> formParams = new ArrayList<>();
        formParams.add(new BasicNameValuePair("grant_type", "password"));
        formParams.add(new BasicNameValuePair("client_id", client_id));
        formParams.add(new BasicNameValuePair("username", username));
        formParams.add(new BasicNameValuePair("password", password));
        UrlEncodedFormEntity entity = new UrlEncodedFormEntity(formParams, Consts.UTF_8);
        httpPost.setEntity(entity);
        try {
            CloseableHttpResponse response = httpClient.execute(httpPost);
            try {
                String responseBody = EntityUtils.toString(response.getEntity());
                JSONObject tokenInfo = new JSONObject(responseBody);
                return tokenInfo;
            } finally {
                response.close();
            }

        } catch (IOException e) {
            throw new RuntimeException(e);
        } finally {
            try {
                httpClient.close();
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }
    }
}
