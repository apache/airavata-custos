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

package org.apache.custos.service.management;

import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;
import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.iam.api.AddExternalIDPLinksRequest;
import org.apache.custos.core.iam.api.AddUserAttributesRequest;
import org.apache.custos.core.iam.api.AddUserRolesRequest;
import org.apache.custos.core.iam.api.CheckingResponse;
import org.apache.custos.core.iam.api.DeleteExternalIDPsRequest;
import org.apache.custos.core.iam.api.DeleteUserAttributeRequest;
import org.apache.custos.core.iam.api.DeleteUserRolesRequest;
import org.apache.custos.core.iam.api.FindUsersRequest;
import org.apache.custos.core.iam.api.FindUsersResponse;
import org.apache.custos.core.iam.api.GetAllResources;
import org.apache.custos.core.iam.api.GetAllResourcesResponse;
import org.apache.custos.core.iam.api.GetExternalIDPsRequest;
import org.apache.custos.core.iam.api.GetExternalIDPsResponse;
import org.apache.custos.core.iam.api.OperationStatus;
import org.apache.custos.core.iam.api.RegisterUserRequest;
import org.apache.custos.core.iam.api.RegisterUserResponse;
import org.apache.custos.core.iam.api.RegisterUsersRequest;
import org.apache.custos.core.iam.api.RegisterUsersResponse;
import org.apache.custos.core.iam.api.ResetUserPassword;
import org.apache.custos.core.iam.api.ResourceTypes;
import org.apache.custos.core.iam.api.UpdateUserProfileRequest;
import org.apache.custos.core.iam.api.UserAttribute;
import org.apache.custos.core.iam.api.UserRepresentation;
import org.apache.custos.core.iam.api.UserSearchMetadata;
import org.apache.custos.core.iam.api.UserSearchRequest;
import org.apache.custos.core.identity.api.AuthToken;
import org.apache.custos.core.identity.api.GetUserManagementSATokenRequest;
import org.apache.custos.core.user.management.api.LinkUserProfileRequest;
import org.apache.custos.core.user.management.api.SynchronizeUserDBRequest;
import org.apache.custos.core.user.management.api.UserProfileRequest;
import org.apache.custos.core.user.profile.api.GetAllUserProfilesResponse;
import org.apache.custos.core.user.profile.api.GetUpdateAuditTrailRequest;
import org.apache.custos.core.user.profile.api.GetUpdateAuditTrailResponse;
import org.apache.custos.core.user.profile.api.UserProfile;
import org.apache.custos.core.user.profile.api.UserStatus;
import org.apache.custos.service.exceptions.AuthenticationException;
import org.apache.custos.service.exceptions.InternalServerException;
import org.apache.custos.service.iam.IamAdminService;
import org.apache.custos.service.identity.IdentityService;
import org.apache.custos.service.profile.UserProfileService;
import io.grpc.Context;
import jakarta.persistence.EntityNotFoundException;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.text.ParseException;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * The UserManagementService class provides methods for managing user registration, enabling and disabling users, and adding and deleting user attributes.
 */
@Service
public class UserManagementService {

    private static final Logger LOGGER = LoggerFactory.getLogger(UserManagementService.class);

    private final UserProfileService userProfileService;
    private final IdentityService identityService;
    private final IamAdminService iamAdminService;

    public UserManagementService(UserProfileService userProfileService, IdentityService identityService, IamAdminService iamAdminService) {
        this.userProfileService = userProfileService;
        this.identityService = identityService;
        this.iamAdminService = iamAdminService;
    }

    /**
     * Registers a user using the given request.
     *
     * @param request the request object used for user registration
     * @return the response object containing the result of the user registration
     * @throws AuthenticationException if an authentication exception occurs
     * @throws InternalServerException if an internal server exception occurs
     */
    public RegisterUserResponse registerUser(RegisterUserRequest request) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest.newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                RegisterUserRequest registerUserRequest = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                return iamAdminService.registerUser(registerUserRequest);

            } else {
                LOGGER.error("Cannot find service token");
                throw new RuntimeException("Cannot find service token when registering the user");
            }

        } catch (Exception ex) {
            String msg = "Error occurred while registering users,  " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("CredentialGenerationException")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Registers and enables users using the given request.
     *
     * @param request the request object used for user registration and enabling
     * @return the response object containing the result of the user registration and enabling
     * @throws AuthenticationException if an authentication exception occurs
     * @throws InternalServerException if an internal server exception occurs
     */
    public RegisterUsersResponse registerAndEnableUsers(RegisterUsersRequest request) {
        try {

            RegisterUsersResponse registerUsersResponse = iamAdminService.registerAndEnableUsers(request);

            if (!request.getUsersList().isEmpty() && registerUsersResponse.getAllUseresRegistered()) {
                try {
                    request.getUsersList().forEach(user -> {
                        List<org.apache.custos.core.user.profile.api.UserAttribute> userAtrList = new ArrayList<>();
                        if (!user.getAttributesList().isEmpty()) {

                            user.getAttributesList().forEach(atr -> {
                                org.apache.custos.core.user.profile.api.UserAttribute userAttribute = org.apache.custos.core.user.profile.api.UserAttribute.newBuilder()
                                        .setKey(atr.getKey())
                                        .addAllValues(atr.getValuesList())
                                        .build();

                                userAtrList.add(userAttribute);
                            });
                        }


                        UserProfile profile = UserProfile.newBuilder()
                                .setFirstName(user.getFirstName())
                                .setLastName(user.getLastName())
                                .setEmail(user.getEmail())
                                .setStatus(UserStatus.ACTIVE)
                                .addAllAttributes(userAtrList)
                                .addAllRealmRoles(user.getRealmRolesList())
                                .addAllClientRoles(user.getClientRolesList())
                                .setUsername(user.getUsername().toLowerCase())
                                .build();
                        org.apache.custos.core.user.profile.api.UserProfileRequest profileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                                .setProfile(profile)
                                .setTenantId(request.getTenantId())
                                .build();

                        userProfileService.createUserProfile(profileRequest);
                    });

                } catch (Exception ex) {
                    request.getUsersList().forEach(user -> {
                        UserSearchMetadata metadata = UserSearchMetadata.newBuilder()
                                .setUsername(user.getUsername())
                                .build();
                        UserSearchRequest searchRequest = UserSearchRequest
                                .newBuilder()
                                .setTenantId(request.getTenantId())
                                .setClientId(request.getClientId())
                                .setAccessToken(request.getAccessToken())
                                .setUser(metadata)
                                .build();
                        iamAdminService.deleteUser(searchRequest);
                    });
                }
            }

            return registerUsersResponse;

        } catch (Exception ex) {
            String msg = "Error occurred while registering and enabling  users,  " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Adds attributes to the user profiles based on the provided request.
     *
     * @param request the request object containing the information to add attributes to user profiles
     * @return the operation status indicating the result of adding user attributes
     * @throws AuthenticationException if an authentication exception occurs
     * @throws InternalServerException if an internal server exception occurs
     */
    public OperationStatus addUserAttributes(AddUserAttributesRequest request) {
        try {
            OperationStatus status = iamAdminService.addUserAttributes(request);

            for (String user : request.getUsersList()) {
                UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setUsername(user).build();

                UserSearchRequest searchRequest = UserSearchRequest
                        .newBuilder()
                        .setClientId(request.getClientId())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(request.getAccessToken())
                        .setUser(metadata)
                        .build();

                UserRepresentation representation = iamAdminService.getUser(searchRequest);


                if (representation != null) {
                    UserProfile profile = this.convertToProfile(representation);

                    org.apache.custos.core.user.profile.api.UserProfileRequest req = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                            .setTenantId(request.getTenantId())
                            .setProfile(profile)
                            .build();


                    UserProfile existingProfile = userProfileService.getUserProfile(req);

                    if (existingProfile == null || StringUtils.isBlank(existingProfile.getUsername())) {
                        userProfileService.createUserProfile(req);
                    } else {
                        userProfileService.updateUserProfile(req);
                    }
                }

            }
            return status;

        } catch (Exception ex) {
            String msg = "Error occurred while adding user attributes, " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Deletes user attributes based on the provided request.
     *
     * @param request the request object containing the information to delete user attributes
     * @return the operation status indicating the result of deleting user attributes
     * @throws AuthenticationException if an authentication exception occurs
     * @throws InternalServerException if an internal server exception occurs
     */
    public OperationStatus deleteUserAttributes(DeleteUserAttributeRequest request) {
        try {
            OperationStatus status = iamAdminService.deleteUserAttributes(request);

            for (String user : request.getUsersList()) {
                UserSearchMetadata metadata = UserSearchMetadata.newBuilder()
                        .setUsername(user).build();

                UserSearchRequest searchRequest = UserSearchRequest.newBuilder()
                        .setClientId(request.getClientId())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(request.getAccessToken())
                        .setUser(metadata)
                        .build();

                UserRepresentation representation = iamAdminService.getUser(searchRequest);

                if (representation != null) {
                    UserProfile profile = this.convertToProfile(representation);

                    org.apache.custos.core.user.profile.api.UserProfileRequest req = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                            .setTenantId(request.getTenantId())
                            .setProfile(profile)
                            .build();

                    userProfileService.updateUserProfile(req);
                }
            }
            return status;

        } catch (Exception ex) {
            String msg = "Error occurred while deleting user attributes " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Enables a user based on the provided search request.
     *
     * @param request the request object containing the user search criteria
     * @return the user representation object of the enabled user
     * @throws RuntimeException        if user enabling fails at the IDP server or if service token cannot be found
     * @throws InternalServerException if an error occurs while enabling the user
     */
    public UserRepresentation enableUser(UserSearchRequest request) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest.newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                UserRepresentation user = iamAdminService.enableUser(request);

                if (user != null) {
                    UserProfile profile = this.convertToProfile(user);

                    org.apache.custos.core.user.profile.api.UserProfileRequest profileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                            .setProfile(profile)
                            .setTenantId(request.getTenantId())
                            .build();

                    UserProfile exProfile = userProfileService.getUserProfile(profileRequest);

                    if (StringUtils.isBlank(exProfile.getUsername())) {
                        userProfileService.createUserProfile(profileRequest);
                    } else {
                        userProfileService.updateUserProfile(profileRequest);
                    }
                    return user;

                } else {
                    String msg = "User enabling failed at IDP server";
                    LOGGER.error(msg);
                    throw new RuntimeException(msg);
                }

            } else {
                LOGGER.error("Cannot find service token");
                throw new RuntimeException("Cannot find service token");
            }


        } catch (Exception ex) {
            String msg = "Error occurred while enabling user, " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Disables a user based on the provided search request.
     *
     * @param request the request object containing the user search criteria
     * @return the user representation object of the disabled user
     * @throws RuntimeException        if user disabling fails at the IDP server or if service token cannot be found
     * @throws InternalServerException if an error occurs while disabling the user
     */
    public UserRepresentation disableUser(UserSearchRequest request) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();

                UserRepresentation user = iamAdminService.disableUser(request);

                if (user != null) {
                    UserProfile profile = this.convertToProfile(user);
                    org.apache.custos.core.user.profile.api.UserProfileRequest profileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                            .setProfile(profile)
                            .setTenantId(request.getTenantId())
                            .build();

                    UserProfile exProfile = userProfileService.getUserProfile(profileRequest);

                    if (StringUtils.isBlank(exProfile.getUsername())) {
                        userProfileService.createUserProfile(profileRequest);
                    } else {
                        userProfileService.updateUserProfile(profileRequest);
                    }
                    return user;

                } else {
                    String msg = "User enabling failed at IDP server";
                    LOGGER.error(msg);
                    throw new RuntimeException(msg);
                }

            } else {
                LOGGER.error("Cannot find service token");
                throw new RuntimeException("Cannot find service token");
            }


        } catch (Exception ex) {
            String msg = "Error occurred while disabling user, " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Deletes a user based on the provided search request.
     *
     * @param request the request object containing the user search criteria
     * @return the operation status indicating the result of deleting the user
     * @throws RuntimeException if there is an error while deleting the user
     */
    public OperationStatus deleteUser(UserSearchRequest request) {
        try {
            UserProfile profileReq = UserProfile.newBuilder().setUsername(request.getUser().getUsername().toLowerCase()).build();
            org.apache.custos.core.user.profile.api.UserProfileRequest req = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                    .setTenantId(request.getTenantId())
                    .setProfile(profileReq)
                    .build();

            UserProfile profile = userProfileService.getUserProfile(req);

            if (profile != null && StringUtils.isNotBlank(profile.getUsername())) {

                UserProfile deletedProfile = userProfileService.deleteUserProfile(req);

                if (deletedProfile != null) {
                    return iamAdminService.deleteUser(request);

                } else {
                    String msg = "User profile deletion failed for " + request.getUser().getUsername();
                    LOGGER.error(msg);
                    throw new RuntimeException(msg);
                }

            } else {
                return iamAdminService.deleteUser(request);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while  deleting user " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Retrieves a user based on the provided search request.
     *
     * @param request the request object containing the user search criteria
     * @return the user representation object of the retrieved user
     * @throws AuthenticationException if an authentication exception occurs
     * @throws EntityNotFoundException if the user is not found
     * @throws InternalServerException if an internal server exception occurs
     */
    public UserRepresentation getUser(UserSearchRequest request) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();

            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                return iamAdminService.getUser(request);

            } else {
                LOGGER.error("Cannot find service token");
                throw new RuntimeException("Cannot find service token");
            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching user, " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else if (ex.getMessage().contains("NOT_FOUND")) {
                throw new EntityNotFoundException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Finds users based on the provided search request.
     *
     * @param request the request object containing the user search criteria
     * @return the response object containing the found users
     * @throws AuthenticationException if an authentication exception occurs
     * @throws InternalServerException if an internal server exception occurs
     */
    public FindUsersResponse findUsers(FindUsersRequest request) {
        try {
            long initiationTime = System.currentTimeMillis();
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest.newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                FindUsersResponse user = iamAdminService.findUsers(request);
                long endTime = System.currentTimeMillis();
                long total = endTime - initiationTime;
                LOGGER.debug("request received: " + initiationTime + " request end time" + endTime + " difference " + total);
                return user;

            } else {
                LOGGER.error("Cannot find service token");
                throw new RuntimeException("Cannot find service token");
            }

        } catch (Exception ex) {
            String msg = "Error occurred while pulling users, " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Resets the password for a user.
     *
     * @param request the request object containing the information to reset the password
     * @return an OperationStatus object indicating the result of the password reset
     * @throws InternalServerException if an error occurs while resetting the password
     */
    public OperationStatus resetPassword(ResetUserPassword request) {
        try {

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest.newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                return iamAdminService.resetPassword(request);

            } else {
                LOGGER.error("Cannot find service token");
                throw new RuntimeException("Cannot find service token");
            }

        } catch (Exception ex) {
            String msg = "Error occurred  while resetting password " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Adds roles to users based on the provided request.
     *
     * @param request the request object containing the information to add roles to users
     * @return the operation status indicating the result of adding roles to users
     * @throws AuthenticationException if an authentication exception occurs
     * @throws InternalServerException if an internal server exception occurs
     */
    public OperationStatus addRolesToUsers(AddUserRolesRequest request) {
        try {
            OperationStatus response = iamAdminService.addRolesToUsers(request);

            for (String user : request.getUsernamesList()) {
                UserSearchMetadata metadata = UserSearchMetadata
                        .newBuilder()
                        .setUsername(user).build();

                UserSearchRequest searchRequest = UserSearchRequest
                        .newBuilder()
                        .setClientId(request.getClientId())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(request.getAccessToken())
                        .setUser(metadata)
                        .build();

                UserRepresentation representation = iamAdminService.getUser(searchRequest);

                if (representation != null) {
                    UserProfile profile = this.convertToProfile(representation);

                    org.apache.custos.core.user.profile.api.UserProfileRequest req = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                            .setTenantId(request.getTenantId())
                            .setProfile(profile)
                            .build();


                    UserProfile existingUser = userProfileService.getUserProfile(req);

                    if (existingUser == null || StringUtils.isBlank(existingUser.getUsername())) {
                        userProfileService.createUserProfile(req);
                    } else {
                        userProfileService.updateUserProfile(req);
                    }
                }
            }
            return response;

        } catch (Exception ex) {
            String msg = "Error occurred while adding roles to users, " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Deletes user roles based on the provided request.
     *
     * @param request the request object containing the user search criteria
     * @return the operation status indicating the result of deleting user roles
     * @throws AuthenticationException if an authentication exception occurs
     * @throws InternalServerException if an internal server exception occurs
     */
    public OperationStatus deleteUserRoles(DeleteUserRolesRequest request) {
        try {
            OperationStatus response = iamAdminService.deleteRolesFromUser(request);
            UserSearchMetadata metadata = UserSearchMetadata.newBuilder()
                    .setUsername(request.getUsername()).build();

            UserSearchRequest searchRequest = UserSearchRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setTenantId(request.getTenantId())
                    .setAccessToken(request.getAccessToken())
                    .setUser(metadata)
                    .build();

            UserRepresentation representation = iamAdminService.getUser(searchRequest);

            if (representation != null) {
                UserProfile profile = this.convertToProfile(representation);

                org.apache.custos.core.user.profile.api.UserProfileRequest req = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                        .setTenantId(request.getTenantId())
                        .setProfile(profile)
                        .build();
                userProfileService.updateUserProfile(req);
            }
            return response;

        } catch (Exception ex) {
            String msg = "Error occurred while delete user roles,  " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Checks whether a user is enabled based on the provided search request.
     *
     * @param request the request object containing the user search criteria
     * @return the operation status indicating whether the user is enabled or not
     * @throws RuntimeException        if user enabling fails at the IDP server or if service token cannot be found
     * @throws InternalServerException if an error occurs while enabling the user
     */
    public OperationStatus isUserEnabled(UserSearchRequest request) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                return iamAdminService.isUserEnabled(request);

            } else {
                LOGGER.error("Cannot find service token");
                throw new RuntimeException("Cannot find service token");
            }
        } catch (Exception ex) {
            String msg = "Error occurred while enabling user " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Checks whether a username is available for registration.
     *
     * @param request the request object containing the user search criteria
     * @return the operation status indicating whether the username is available or not
     * @throws InternalServerException if an error occurs while checking the username
     */
    public OperationStatus isUsernameAvailable(UserSearchRequest request) {
        try {
            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(request.getClientId())
                    .setClientSecret(request.getClientSec())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                request = request.toBuilder().setAccessToken(token.getAccessToken()).build();
                return iamAdminService.isUsernameAvailable(request);

            } else {
                LOGGER.error("Cannot find service token");
                throw new RuntimeException("Cannot find service token");
            }

        } catch (Exception ex) {
            String msg = "Error occurred while checking username, " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Updates the user profile for the given user.
     *
     * @param request The user profile request containing the updated user profile information.
     * @return The updated user profile.
     * @throws InternalServerException if an error occurs while updating the user profile.
     * @throws AuthenticationException if the request is unauthenticated.
     */
    public UserProfile updateUserProfile(UserProfileRequest request) {
        try {
            LOGGER.debug("Request received to updateUserProfile for " + request.getUserProfile().getUsername() + " in " + request.getTenantId());

            UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setUsername(request.getUserProfile().getUsername()).build();

            UserSearchRequest info = UserSearchRequest
                    .newBuilder()
                    .setAccessToken(request.getAccessToken())
                    .setTenantId(request.getTenantId())
                    .setUser(metadata)
                    .build();

            CheckingResponse response = iamAdminService.isUserExist(info);


            if (!response.getIsExist()) {
                String msg = "User not found with username " + request.getUserProfile().getUsername();
                LOGGER.error(msg);
                throw new InternalServerException(msg);
            }

            UserRepresentation userRepresentation = iamAdminService.getUser(info);

            userRepresentation = userRepresentation
                    .toBuilder()
                    .setFirstName(request.getUserProfile().getFirstName())
                    .setLastName(request.getUserProfile().getLastName())
                    .setEmail(request.getUserProfile().getEmail())
                    .build();

            UpdateUserProfileRequest updateUserProfileRequest = UpdateUserProfileRequest
                    .newBuilder()
                    .setUser(userRepresentation)
                    .setAccessToken(request.getAccessToken())
                    .setTenantId(request.getTenantId())
                    .build();

            OperationStatus operationStatus = iamAdminService.updateUserProfile(updateUserProfileRequest);

            if (operationStatus != null && operationStatus.getStatus()) {
                try {
                    org.apache.custos.core.user.profile.api.UserProfileRequest userProfileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest.
                            newBuilder()
                            .setProfile(request.getUserProfile())
                            .setTenantId(request.getTenantId())
                            .build();

                    UserProfile profile = userProfileService.getUserProfile(userProfileRequest);

                    if (profile != null && StringUtils.isNotBlank(profile.getUsername())) {
                        profile = profile.toBuilder()
                                .setFirstName(request.getUserProfile().getFirstName())
                                .setLastName(request.getUserProfile().getLastName())
                                .build();
                        userProfileRequest = userProfileRequest.toBuilder().setProfile(profile).build();
                        userProfileService.updateUserProfile(userProfileRequest);
                        return profile;

                    } else {
                        UserProfile userProfile = UserProfile.newBuilder()
                                .setEmail(request.getUserProfile().getEmail())
                                .setFirstName(request.getUserProfile().getFirstName())
                                .setLastName(request.getUserProfile().getLastName())
                                .setUsername(request.getUserProfile().getUsername())
                                .build();
                        userProfileRequest = userProfileRequest.toBuilder().setProfile(userProfile).build();
                        userProfileService.createUserProfile(userProfileRequest);
                        return profile;
                    }

                } catch (Exception ex) {
                    String msg = "Error occurred while saving user profile in local DB, rolling back IAM service" + ex.getMessage();
                    LOGGER.error(msg);
                    UpdateUserProfileRequest rollingRequest = UpdateUserProfileRequest.newBuilder()
                            .setUser(userRepresentation)
                            .setAccessToken(request.getAccessToken())
                            .setTenantId(request.getTenantId())
                            .build();
                    iamAdminService.updateUserProfile(rollingRequest);
                    throw new RuntimeException(msg, ex);
                }
            } else {
                String msg = "Cannot update user profile in keycloak for user " + request.getUserProfile().getUsername();
                LOGGER.error(msg);
                throw new InternalServerException(msg);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while updating user profile " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Deletes the user profile based on the provided request.
     *
     * @param request the request object containing the user profile to be deleted
     * @return the deleted user profile
     * @throws AuthenticationException if an authentication exception occurs
     * @throws InternalServerException if an internal server exception occurs
     */
    public UserProfile deleteUserProfile(UserProfileRequest request) {
        try {
            LOGGER.debug("Request received to deleteUserProfile " + request.getUserProfile().getUsername() + " at " + request.getTenantId());

            UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setUsername(request.getUserProfile().getUsername()).build();
            UserSearchRequest info = UserSearchRequest.newBuilder()
                    .setAccessToken(request.getAccessToken())
                    .setTenantId(request.getTenantId())
                    .setUser(metadata)
                    .build();

            org.apache.custos.core.user.profile.api.UserProfileRequest userProfileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                    .setProfile(request.getUserProfile())
                    .setTenantId(request.getTenantId())
                    .build();
            UserProfile userProfile = userProfileService.deleteUserProfile(userProfileRequest);
            iamAdminService.deleteUser(info);

            return userProfile;

        } catch (Exception ex) {
            String msg = "Error occurred while deleting user profile " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Retrieves the user profile based on the provided request.
     *
     * @param request the request object containing the user profile search criteria
     * @return the user profile object corresponding to the retrieved profile
     * @throws InternalServerException if an internal server exception occurs
     */
    public UserProfile getUserProfile(UserProfileRequest request) {
        try {
            LOGGER.debug("Request received to getUserProfile " + request.getUserProfile().getUsername() + " at " + request.getTenantId());

            org.apache.custos.core.user.profile.api.UserProfileRequest userProfileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest
                    .newBuilder()
                    .setProfile(request.getUserProfile())
                    .setTenantId(request.getTenantId())
                    .build();

            return userProfileService.getUserProfile(userProfileRequest);

        } catch (Exception ex) {
            String msg = "Error occurred while pulling  user profile " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Retrieves all user profiles within a specific tenant.
     *
     * @param request the request object containing the search criteria and the variables needed for the retrieval
     * @return the response object containing the retrieved user profiles
     * @throws InternalServerException if an internal server exception occurs during the retrieval process
     */
    public GetAllUserProfilesResponse getAllUserProfilesInTenant(UserProfileRequest request) {
        try {
            LOGGER.debug("Request received to getAllUserProfilesInTenant " + request.getTenantId() + " at " + request.getTenantId());

            org.apache.custos.core.user.profile.api.UserProfileRequest userProfileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest
                    .newBuilder()
//                    .setProfile(request.getUserProfile())
                    .setTenantId(request.getTenantId())
                    .setOffset(request.getOffset())
                    .setLimit(request.getLimit())
                    .build();

            return userProfileService.getAllUserProfilesInTenant(userProfileRequest);

//            return request.getUserProfile().getAttributesList().isEmpty()
//                    ? userProfileService.getAllUserProfilesInTenant(userProfileRequest)
//                    : userProfileService.findUserProfilesByAttributes(userProfileRequest);

        } catch (Exception ex) {
            String msg = "Error occurred while pulling  all  user profiles in tenant " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Retrieves the audit trails for a user profile based on the provided request.
     *
     * @param request the request object containing the search criteria
     * @return the response object containing the audit trails for the user profile
     * @throws InternalServerException if an internal server exception occurs during the retrieval process
     */
    public GetUpdateAuditTrailResponse getUserProfileAuditTrails(GetUpdateAuditTrailRequest request) {
        try {
            LOGGER.debug("Request received to getUserProfileAuditTrails " + request.getUsername() + " at " + request.getTenantId());
            return userProfileService.getUserProfileAuditTrails(request);

        } catch (Exception ex) {
            String msg = "Error occurred while pulling user profile audit trails " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Links the user profile by updating the attributes and creating/updating the profile in the user profile service.
     *
     * @param request The request object containing the information required to link the user profile.
     * @return The operation status indicating the result of the linking process.
     * @throws EntityNotFoundException If the existing user to be linked cannot be found.
     * @throws AuthenticationException If authentication fails while searching for the access token.
     * @throws InternalServerException If an error occurs during the linking process.
     */
    public OperationStatus linkUserProfile(LinkUserProfileRequest request) {
        try {
            LOGGER.debug("Request received to linkUserProfile   at " + request.getTenantId());

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest.newBuilder()
                    .setClientId(request.getIamClientId())
                    .setClientSecret(request.getIamClientSecret())
                    .setTenantId(request.getTenantId())
                    .build();
            AuthToken token = identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);

            if (token != null && StringUtils.isNotBlank(token.getAccessToken())) {
                UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setUsername(request.getCurrentUsername()).build();
                UserSearchRequest searchRequest = UserSearchRequest.newBuilder()
                        .setClientId(request.getIamClientId())
                        .setTenantId(request.getTenantId())
                        .setAccessToken(token.getAccessToken())
                        .setUser(metadata)
                        .build();

                UserRepresentation userTobeLinked = iamAdminService.getUser(searchRequest);

                if (userTobeLinked != null && StringUtils.isNotBlank(userTobeLinked.getUsername())) {

                    UserSearchMetadata exMetadata = UserSearchMetadata.newBuilder().setUsername(request.getPreviousUsername()).build();
                    UserSearchRequest exSearchRequest = UserSearchRequest.newBuilder()
                            .setClientId(request.getIamClientId())
                            .setTenantId(request.getTenantId())
                            .setAccessToken(token.getAccessToken())
                            .setUser(exMetadata)
                            .build();

                    UserRepresentation exRep = iamAdminService.getUser(exSearchRequest);

                    if (exRep != null && StringUtils.isNotBlank(exRep.getUsername())) {
                        boolean profileUpdate = false;
                        List<UserAttribute> userAttributeList = new ArrayList<>();
                        for (String attribute : request.getLinkingAttributesList()) {
                            if ("name".equals(attribute)) {
                                profileUpdate = true;
                                userTobeLinked = userTobeLinked.toBuilder()
                                        .setFirstName(exRep.getFirstName())
                                        .setLastName(exRep.getLastName())
                                        .build();

                            } else if (("email").equals(attribute)) {
                                profileUpdate = true;
                                userTobeLinked = userTobeLinked.toBuilder().setEmail(exRep.getEmail()).build();

                            } else {
                                List<UserAttribute> userAttributes = exRep.getAttributesList().stream().
                                        filter(atr -> atr.getKey().equals(attribute)).toList();

                                if (!userAttributes.isEmpty()) {
                                    UserAttribute userAttribute = userAttributes.get(0);
                                    userAttributeList.add(userAttribute);
                                }
                            }
                        }

                        if (profileUpdate) {
                            UpdateUserProfileRequest updateUserProfileRequest = UpdateUserProfileRequest
                                    .newBuilder()
                                    .setUser(userTobeLinked)
                                    .setAccessToken(token.getAccessToken())
                                    .setTenantId(request.getTenantId())
                                    .build();
                            iamAdminService.updateUserProfile(updateUserProfileRequest);
                        }

                        if (!userAttributeList.isEmpty()) {
                            AddUserAttributesRequest addUserAttributesRequest = AddUserAttributesRequest
                                    .newBuilder()
                                    .addUsers(request.getCurrentUsername())
                                    .addAllAttributes(userAttributeList)
                                    .setTenantId(request.getTenantId())
                                    .setAccessToken(token.getAccessToken())
                                    .setClientId(request.getIamClientId())
                                    .setPerformedBy(request.getPerformedBy())
                                    .build();
                            iamAdminService.addUserAttributes(addUserAttributesRequest);

                        }

                        UserRepresentation updatedUser = iamAdminService.getUser(searchRequest);

                        if (updatedUser != null) {
                            UserProfile profile = this.convertToProfile(updatedUser);
                            org.apache.custos.core.user.profile.api.UserProfileRequest req = org.apache.custos.core.user.profile.api.UserProfileRequest
                                    .newBuilder()
                                    .setTenantId(request.getTenantId())
                                    .setProfile(profile)
                                    .build();

                            UserProfile existingProfile = userProfileService.getUserProfile(req);

                            if (existingProfile == null || StringUtils.isBlank(existingProfile.getUsername())) {
                                userProfileService.createUserProfile(req);
                            } else {
                                userProfileService.updateUserProfile(req);
                            }
                        }

                        CheckingResponse response = CheckingResponse.newBuilder().setIsExist(true).build();
                        return OperationStatus.newBuilder().setStatus(response.getIsExist()).build();

                    } else {
                        String msg = "Cannot found existing user ";
                        LOGGER.error(msg);
                        throw new EntityNotFoundException(msg);
                    }
                }
                LOGGER.error("Cannot find an existing user to be linked. User name: " + request.getCurrentUsername());
                throw new EntityNotFoundException("Cannot find an existing user to be linked. User name: " + request.getCurrentUsername());
            }
            LOGGER.error("Cannot find the access token for User search request. User name: " + request.getCurrentUsername());
            throw new RuntimeException("Cannot find the access token for User search request. User name: " + request.getCurrentUsername());

        } catch (Exception ex) {
            String msg = "Error occurred while linking user profile in tenant " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                throw new AuthenticationException(msg, ex);
            } else {
                throw new InternalServerException(msg, ex);
            }
        }
    }

    /**
     * Grants admin privileges to a user and updates their user profile.
     *
     * @param request the UserSearchRequest object containing the necessary information for granting admin privileges
     * @return an OperationStatus object indicating the status of the operation
     * @throws EntityNotFoundException if the user is not found
     * @throws InternalServerException if an error occurs while pulling user profile audit trails
     */
    public OperationStatus grantAdminPrivileges(UserSearchRequest request) {
        try {

            LOGGER.debug("Request received to grantAdminPrivileges " + request.getUser().getUsername() + " at " + request.getTenantId());

            iamAdminService.grantAdminPrivilege(request);
            UserRepresentation representation = iamAdminService.getUser(request);

            if (representation != null) {
                UserProfile profile = convertToProfile(representation);
                org.apache.custos.core.user.profile.api.UserProfileRequest profileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                        .setTenantId(request.getTenantId())
                        .setPerformedBy(request.getPerformedBy())
                        .setProfile(profile)
                        .build();

                UserProfile exProfile = userProfileService.getUserProfile(profileRequest);

                if (exProfile == null || StringUtils.isBlank(exProfile.getUsername())) {
                    userProfileService.createUserProfile(profileRequest);
                } else {
                    userProfileService.updateUserProfile(profileRequest);
                }

                return OperationStatus.newBuilder().setStatus(true).build();

            } else {
                String msg = "User not found. User Id: " + request.getUser().getId();
                LOGGER.error(msg);
                throw new EntityNotFoundException(msg);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while puling user profile audit trails " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Removes admin privileges for a user.
     *
     * @param request the UserSearchRequest object containing the user details and tenant ID
     * @return an OperationStatus object indicating the status of the operation
     * @throws EntityNotFoundException if the user is not found
     * @throws InternalServerException if an error occurs while removing the admin privileges
     */
    public OperationStatus removeAdminPrivileges(UserSearchRequest request) {

        try {
            LOGGER.debug("Request received to removeAdminPrivileges " + request.getUser().getUsername() + " at " + request.getTenantId());

            iamAdminService.removeAdminPrivilege(request);
            UserRepresentation representation = iamAdminService.getUser(request);
            if (representation != null) {

                UserProfile profile = convertToProfile(representation);
                org.apache.custos.core.user.profile.api.UserProfileRequest profileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest
                        .newBuilder()
                        .setTenantId(request.getTenantId())
                        .setPerformedBy(request.getPerformedBy())
                        .setProfile(profile)
                        .build();

                UserProfile exProfile = userProfileService.getUserProfile(profileRequest);

                if (exProfile == null || StringUtils.isBlank(exProfile.getUsername())) {
                    userProfileService.createUserProfile(profileRequest);
                } else {
                    userProfileService.updateUserProfile(profileRequest);
                }
                return OperationStatus.newBuilder().setStatus(true).build();

            } else {
                String msg = "User not found. User Id: " + request.getUser().getId();
                LOGGER.error(msg);
                throw new EntityNotFoundException(msg);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while  removing admin privileges " + ex.getMessage();
            LOGGER.error(msg);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Deletes the external IDPs (Identity Providers) of the users based on the provided request.
     *
     * @param request the request object containing the information to delete the external IDPs of users
     * @return the operation status indicating the result of deleting the external IDPs of users
     **/
    public OperationStatus deleteExternalIDPsOfUsers(DeleteExternalIDPsRequest request) {
        try {
            LOGGER.debug("Request received to deleteExternalIDPsOfUsers for " + request.getTenantId());
            return iamAdminService.deleteExternalIDPLinksOfUsers(request);

        } catch (Exception ex) {
            String msg = "Error occurred while  deleting external IDPs of Users " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * This method retrieves the external IDPs (Identity Providers) of users based on the provided request.
     *
     * @param request The request object containing the necessary details for retrieving the external IDPs of users.
     * @return A GetExternalIDPsResponse object containing the retrieved external IDPs of users.
     * @throws InternalServerException If an error occurs while fetching the external IDPs
     */
    public GetExternalIDPsResponse getExternalIDPsOfUsers(GetExternalIDPsRequest request) {
        try {
            LOGGER.debug("Request received to getExternalIDPs of users in " + request.getTenantId());
            return iamAdminService.getExternalIDPLinksOfUsers(request);

        } catch (Exception ex) {
            String msg = "Error occurred while  fetching external IDPs of Users " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Adds external IDPs (Identity Provider) links of users.
     *
     * @param request the request containing the external IDPs links to be added
     * @return the operation status indicating the result of the operation
     * @throws InternalServerException if an error occurs while adding the external IDPs links
     */
    public OperationStatus addExternalIDPsOfUsers(AddExternalIDPLinksRequest request) {
        try {
            LOGGER.debug("Request received to addExternalIDPsOfUsers of users in " + request.getTenantId());
            return iamAdminService.addExternalIDPLinksOfUsers(request);

        } catch (Exception ex) {
            String msg = "Error occurred while  adding external IDPs of Users " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    /**
     * Synchronizes user databases.
     *
     * @param request the SynchronizeUserDBRequest object containing the necessary parameters for synchronization
     * @return the OperationStatus object indicating the status of the synchronization operation
     * @throws InternalServerException if an error occurs during the synchronization process
     */
    public OperationStatus synchronizeUserDBs(SynchronizeUserDBRequest request) {
        OperationStatus status = null;
        try {
            Context ctx = Context.current().fork();
            ctx.run(() -> {
                GetAllResources resources = GetAllResources
                        .newBuilder()
                        .setClientId(request.getClientId())
                        .setTenantId(request.getTenantId())
                        .setResourceType(ResourceTypes.USER)
                        .build();
                GetAllResourcesResponse response = iamAdminService.getAllResources(resources);

                if (response != null && !response.getUsersList().isEmpty()) {

                    for (org.apache.custos.core.iam.api.UserRepresentation userRepresentation : response.getUsersList()) {

                        LOGGER.debug("User Name " + userRepresentation.getUsername());
                        UserProfile profile = convertToProfile(userRepresentation);

                        org.apache.custos.core.user.profile.api.UserProfileRequest profileRequest = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                                .setTenantId(request.getTenantId())
                                .setPerformedBy(Constants.SYSTEM)
                                .setProfile(profile)
                                .build();

                        UserProfile exProfile = userProfileService.getUserProfile(profileRequest);

                        if (exProfile == null || StringUtils.isBlank(exProfile.getUsername())) {
                            userProfileService.createUserProfile(profileRequest);
                        } else {
                            userProfileService.updateUserProfile(profileRequest);
                        }
                    }
                } else {
                    LOGGER.debug("Empty");
                }
            });
            return OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = "Error occurred at synchronizeAgentDBs " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new InternalServerException(msg, ex);
        }
    }

    public Map<String, Object> getUserInfo(String accessToken, long tenantId) {
        try {
            Map<String, Object> userInfo = iamAdminService.getUserInfo(accessToken, tenantId);

            SignedJWT signedJWT = SignedJWT.parse(accessToken);
            JWTClaimsSet claims = signedJWT.getJWTClaimsSet();
            userInfo.put("groups", claims.getClaim("groups"));
            userInfo.put("scopes", claims.getClaim("scopes"));

            return userInfo;
        } catch (ParseException e) {
            throw new IllegalArgumentException("Error while extracting userinfo");
        }
    }

    private UserProfile convertToProfile(UserRepresentation representation) {
        UserProfile.Builder profileBuilder = UserProfile.newBuilder();
        if (representation.getRealmRolesCount() > 0) {
            profileBuilder.addAllRealmRoles(representation.getRealmRolesList());
        }

        if (representation.getClientRolesCount() > 0) {
            profileBuilder.addAllClientRoles(representation.getClientRolesList());
        }

        if (representation.getAttributesCount() > 0) {
            List<UserAttribute> attributeList = representation.getAttributesList();

            List<org.apache.custos.core.user.profile.api.UserAttribute> userAtrList = new ArrayList<>();
            attributeList.forEach(atr -> {
                org.apache.custos.core.user.profile.api.UserAttribute userAttribute = org.apache.custos.core.user.profile.api.UserAttribute
                        .newBuilder()
                        .setKey(atr.getKey())
                        .addAllValues(atr.getValuesList())
                        .build();
                userAtrList.add(userAttribute);
            });

            profileBuilder.addAllAttributes(userAtrList);
        }

        profileBuilder.setUsername(representation.getUsername().toLowerCase());
        profileBuilder.setFirstName(representation.getFirstName());
        profileBuilder.setLastName(representation.getLastName());
        profileBuilder.setEmail(representation.getEmail());

        return profileBuilder.build();
    }

}
