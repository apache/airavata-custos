/*
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
 *
*/
package org.apache.custos.profile.handlers;

import org.apache.custos.commons.exceptions.ApplicationSettingsException;
import org.apache.custos.commons.exceptions.CustosSecurityException;
import org.apache.custos.commons.model.error.AuthorizationException;
import org.apache.custos.commons.model.security.AuthzToken;
import org.apache.custos.commons.model.security.UserInfo;
import org.apache.custos.commons.utils.Constants;
import org.apache.custos.commons.utils.CustosUtils;
import org.apache.custos.commons.utils.ServerSettings;
import org.apache.custos.profile.client.ProfileServiceClientFactory;
import org.apache.custos.profile.iam.admin.services.cpi.IamAdminServices;
import org.apache.custos.profile.iam.admin.services.cpi.exception.IamAdminServicesException;
import org.apache.custos.profile.model.user.Status;
import org.apache.custos.profile.model.user.UserProfile;
import org.apache.custos.profile.model.user.user_profile_modelConstants;
import org.apache.custos.profile.user.core.repositories.UserProfileRepository;
import org.apache.custos.profile.user.cpi.UserProfileService;
import org.apache.custos.profile.user.cpi.exception.UserProfileServiceException;
import org.apache.custos.security.interceptor.SecurityCheck;
import org.apache.custos.security.manager.CustosSecurityManager;
import org.apache.custos.security.manager.SecurityManagerFactory;
import org.apache.thrift.TException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;


public class UserProfileServiceHandler implements UserProfileService.Iface {

    private final static Logger logger = LoggerFactory.getLogger(UserProfileServiceHandler.class);

    private UserProfileRepository userProfileRepository;

    public UserProfileServiceHandler() {

        userProfileRepository = new UserProfileRepository();
    }

    @Override
    public String getAPIVersion() {
        return user_profile_modelConstants.USER_PROFILE_VERSION;
    }

    @Override
    //@SecurityCheck
    public UserProfile initializeUserProfile(AuthzToken authzToken) throws UserProfileServiceException, AuthorizationException {
        String gatewayId = authzToken.getClaimsMap().get(Constants.GATEWAY_ID);
        try {
            // Load UserInfo for the access token and create an initial UserProfile from it
            UserInfo userInfo = SecurityManagerFactory.getSecurityManager().getUserInfoFromAuthzToken(authzToken);
            final UserProfile existingProfile = userProfileRepository.getUserProfileByIdAndGateWay(userInfo.getUsername(), gatewayId);
            // If a user profile already exists, just return the existing profile
            if (existingProfile != null) {
                return existingProfile;
            }
            UserProfile userProfile = new UserProfile();
            userProfile.setUserId(userInfo.getUsername().toLowerCase());
            userProfile.setGatewayId(gatewayId);
            userProfile.setCustosInternalUserId(userProfile.getUserId() + "@" + gatewayId);
            userProfile.addToEmails(userInfo.getEmailAddress());
            userProfile.setFirstName(userInfo.getFirstName());
            userProfile.setLastName(userInfo.getLastName());
            userProfile.setCreationTime(CustosUtils.getCurrentTimestamp().getTime());
            userProfile.setLastAccessTime(CustosUtils.getCurrentTimestamp().getTime());
            userProfile.setValidUntil(-1);
            userProfile.setState(Status.ACTIVE);
            userProfile = userProfileRepository.createUserProfile(userProfile);
            if (null != userProfile) {
                logger.info("Added UserProfile with userId: " + userProfile.getUserId());
                return userProfile;
            } else {
                throw new Exception("User creation failed. Please try again.");
            }
        } catch (Exception e) {
            logger.error("Error while initializing user profile", e);
            UserProfileServiceException exception = new UserProfileServiceException();
            exception.setMessage("Error while initializing user profile. More info : " + e.getMessage());
            throw exception;
        }
    }

    @Override
    //@SecurityCheck
    public UserProfile addUserProfile(AuthzToken authzToken, UserProfile userProfile) throws UserProfileServiceException, AuthorizationException {
        try{
            // Lowercase user id and internal id
            userProfile.setUserId(userProfile.getUserId().toLowerCase());
            userProfile.setCustosInternalUserId(userProfile.getUserId() + "@" + userProfile.getGatewayId());
            userProfile = userProfileRepository.updateUserProfile(userProfile, getIAMUserProfileUpdater(authzToken, userProfile));
            if (null != userProfile) {
                logger.info("Added UserProfile with userId: " + userProfile.getUserId());
                return userProfile;
            } else {
                throw new Exception("User creation failed. Please try again.");
            }
        } catch (Exception e) {
            logger.error("Error while creating user profile", e);
            UserProfileServiceException exception = new UserProfileServiceException();
            exception.setMessage("Error while creating user profile. More info : " + e.getMessage());
            throw exception;
        }
    }

    @Override
    //@SecurityCheck
    public UserProfile updateUserProfile(AuthzToken authzToken, UserProfile userProfile) throws UserProfileServiceException, AuthorizationException {
        try {
            // After updating the user profile in the database but before committing the transaction, the
            // following will update the user profile in the IAM service also. If the update in the IAM service
            // fails then the transaction will be rolled back.
            Runnable iamUserProfileUpdater = getIAMUserProfileUpdater(authzToken, userProfile);
            if(userProfileRepository.updateUserProfile(userProfile, iamUserProfileUpdater) != null) {
                logger.info("Updated UserProfile with userId: " + userProfile.getUserId());
               return userProfile;
            }
        } catch (Exception e) {
            logger.error("Error while Updating user profile", e);
            UserProfileServiceException exception = new UserProfileServiceException();
            exception.setMessage("Error while Updating user profile. More info : " + e.getMessage());
            throw exception;
        }
        throw new UserProfileServiceException("User update failed. Please try again.");
    }

    private Runnable getIAMUserProfileUpdater(AuthzToken authzToken, UserProfile userProfile) throws UserProfileServiceException {
        String gatewayId = authzToken.getClaimsMap().get(Constants.GATEWAY_ID);
        return () -> {
            try {
                CustosSecurityManager securityManager = SecurityManagerFactory.getSecurityManager();
                AuthzToken serviceAccountAuthzToken = securityManager.getUserManagementServiceAccountAuthzToken(authzToken,gatewayId);
                IamAdminServices.Client iamAdminServicesClient = getIamAdminServicesClient();
                iamAdminServicesClient.updateUserProfile(serviceAccountAuthzToken, userProfile);
            } catch (CustosSecurityException |TException e) {
                throw new RuntimeException("Failed to update user profile in IAM service", e);
            }
        };
    }

    @Override
    //@SecurityCheck
    public UserProfile getUserProfileById(AuthzToken authzToken, String userId, String gatewayId) throws UserProfileServiceException, AuthorizationException, TException {
        try{
            UserProfile userProfile = userProfileRepository.getUserProfileByIdAndGateWay(userId, gatewayId);
            if(userProfile != null)
                return userProfile;
            else
                throw new Exception("User with userId: " + userId + ", in Gateway: " + gatewayId + ", does not exist.");
        } catch (Exception e) {
            logger.error("Error retrieving user profile by ID", e);
            UserProfileServiceException exception = new UserProfileServiceException();
            exception.setMessage("Error retrieving user profile by ID. More info : " + e.getMessage());
            throw exception;
        }
    }

    @Override
    //@SecurityCheck
    public UserProfile deleteUserProfile(AuthzToken authzToken, String userId, String gatewayId) throws UserProfileServiceException, AuthorizationException, TException {
        UserProfile userProfile = null;
        try{
            // find user-profile
            userProfile = userProfileRepository.getUserProfileByIdAndGateWay(userId, gatewayId);

            // delete user
            boolean deleteSuccess = userProfileRepository.delete(userId);
            logger.info("Delete UserProfile with userId: " + userId + ", " + (deleteSuccess? "Success!" : "Failed!"));
        } catch (Exception e) {
            logger.error("Error while deleting user profile", e);
            UserProfileServiceException exception = new UserProfileServiceException();
            exception.setMessage("Error while deleting user profile. More info : " + e.getMessage());
            throw exception;
        }
        return userProfile;
    }

    @Override
    //@SecurityCheck
    public List<UserProfile> getAllUserProfilesInGateway(AuthzToken authzToken, String gatewayId, int offset, int limit) throws UserProfileServiceException, AuthorizationException {
        try{
            List<UserProfile> usersInGateway = userProfileRepository.getAllUserProfilesInGateway(gatewayId, offset, limit);
            if(usersInGateway != null)
                return usersInGateway;
            else
                throw new Exception("There are no users for the requested gatewayId: " + gatewayId);
        } catch (Exception e) {
            logger.error("Error while retrieving user profile List", e);
            UserProfileServiceException exception = new UserProfileServiceException();
            exception.setMessage("Error while retrieving user profile List. More info : " + e.getMessage());
            throw exception;
        }
    }

    @Override
    public boolean doesUserExist(AuthzToken authzToken, String userId, String gatewayId) throws UserProfileServiceException, AuthorizationException {
        try{
            UserProfile userProfile = userProfileRepository.getUserProfileByIdAndGateWay(userId, gatewayId);
            return null != userProfile;
        } catch (Exception e) {
            logger.error("Error while finding user profile", e);
            UserProfileServiceException exception = new UserProfileServiceException();
            exception.setMessage("Error while finding user profile. More info : " + e.getMessage());
            throw exception;
        }
    }

    private IamAdminServices.Client getIamAdminServicesClient() throws UserProfileServiceException {
        try {
            final int serverPort = Integer.parseInt(ServerSettings.getProfileServiceServerPort());
            final String serverHost = ServerSettings.getProfileServiceServerHost();
            return ProfileServiceClientFactory.createIamAdminServiceClient(serverHost, serverPort);
        } catch (IamAdminServicesException | ApplicationSettingsException e) {
            logger.error("Failed to create IAM Admin Services client", e);
            UserProfileServiceException ex = new UserProfileServiceException("Failed to create IAM Admin Services client");
            throw ex;
        }
    }
}
