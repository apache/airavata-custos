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

package org.apache.custos.scim.resource.manager;

import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.iam.admin.client.IamAdminServiceClient;
import org.apache.custos.iam.service.*;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.identity.service.GetUserManagementSATokenRequest;
import org.apache.custos.scim.exception.CustosSCIMException;
import org.apache.custos.scim.utils.Constants;
import org.apache.custos.user.profile.client.UserProfileClient;
import org.apache.custos.user.profile.service.GetAllGroupsResponse;
import org.apache.custos.user.profile.service.GetAllUserProfilesResponse;
import org.apache.custos.user.profile.service.UserProfile;
import org.apache.custos.user.profile.service.UserStatus;
import org.json.JSONArray;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.wso2.charon3.core.attributes.Attribute;
import org.wso2.charon3.core.attributes.SimpleAttribute;
import org.wso2.charon3.core.config.SCIMUserSchemaExtensionBuilder;
import org.wso2.charon3.core.encoder.JSONDecoder;
import org.wso2.charon3.core.encoder.JSONEncoder;
import org.wso2.charon3.core.exceptions.*;
import org.wso2.charon3.core.extensions.UserManager;
import org.wso2.charon3.core.objects.Group;
import org.wso2.charon3.core.objects.User;
import org.wso2.charon3.core.protocol.endpoints.AbstractResourceManager;
import org.wso2.charon3.core.schema.SCIMConstants;
import org.wso2.charon3.core.schema.SCIMResourceSchemaManager;
import org.wso2.charon3.core.schema.SCIMResourceTypeSchema;
import org.wso2.charon3.core.utils.codeutils.SearchRequest;

import java.time.Instant;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static org.wso2.charon3.core.protocol.endpoints.AbstractResourceManager.getDecoder;
import static org.wso2.charon3.core.protocol.endpoints.AbstractResourceManager.getEncoder;

/**
 * Class responsible for manage Users. Responsible for request response formatting and
 * interact with core services
 */
@Component
public class ResourceManager implements UserManager {

    private static final Logger LOGGER = LoggerFactory.getLogger(ResourceManager.class);


    @Autowired
    private UserProfileClient userProfileClient;

    @Autowired
    private IamAdminServiceClient iamAdminServiceClient;

    @Autowired
    private IdentityClient identityClient;

    @Autowired
    private CredentialStoreServiceClient credentialStoreServiceClient;


    public ResourceManager(@Value("${scim.resource.user.endpoint:/v2/Users}") String userEndpoint,
                           @Value("${scim.resource.group.endpoint:/v2/Groups}") String groupEndpoint,
                           @Value("${scim.user.schema.location}") String location) {

        Map<String, String> endpointMap = new HashMap();
        endpointMap.put(SCIMConstants.USER_ENDPOINT, userEndpoint);
        endpointMap.put(SCIMConstants.GROUP_ENDPOINT, groupEndpoint);
        AbstractResourceManager.setEndpointURLMap(endpointMap);

        try {
            SCIMUserSchemaExtensionBuilder.getInstance()
                    .buildUserSchemaExtension(location);
        } catch (Exception e) {
            String msg = "User schema building error";
            LOGGER.error(msg, e);
            throw new RuntimeException(msg, e);
        }
    }


    @Override
    public User createUser(User user, Map<String, Boolean> map) throws CharonException, ConflictException, BadRequestException {
        Attribute attribute = user.getAttribute(Constants.CUSTOS_EXTENSION);
        Attribute clientId = attribute.getSubAttribute(Constants.CLIENT_ID);
        Attribute clientSec = attribute.getSubAttribute(Constants.CLIENT_SEC);
        Attribute tenantId = attribute.getSubAttribute(Constants.TENANT_ID);
        String clId = ((SimpleAttribute) clientId).getStringValue();
        String clSec = ((SimpleAttribute) clientSec).getStringValue();
        String tenant = ((SimpleAttribute) tenantId).getStringValue();

        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                .newBuilder()
                .setClientId(clId)
                .setClientSecret(clSec)
                .setTenantId(Long.valueOf(tenant))
                .build();
        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

        if (token != null && token.getAccessToken() != null) {

            UserRepresentation userRepresentation = UserRepresentation
                    .newBuilder()
                    .setFirstName(user.getName().getGivenName())
                    .setLastName(user.getName().getFamilyName())
                    .setEmail(user.getEmails().get(0).getValue())
                    .setUsername(user.getExternalId())
                    .setPassword(user.getPassword())
                    .setTemporaryPassword(false)
                    .build();

            RegisterUserRequest registerUserRequest = RegisterUserRequest.newBuilder()
                    .setTenantId(Long.valueOf(tenant))
                    .setAccessToken(token.getAccessToken())
                    .setUser(userRepresentation)
                    .build();

            RegisterUserResponse registerUserResponse = iamAdminServiceClient.registerUser(registerUserRequest);

            if (registerUserResponse.getIsRegistered()) {

                if (user.getActive()) {

                    UserSearchMetadata metada = UserSearchMetadata.newBuilder().setUsername(user.getExternalId()).build();

                    UserSearchRequest request = UserSearchRequest
                            .newBuilder()
                            .setTenantId(Long.valueOf(tenant))
                            .setAccessToken(token.getAccessToken())
                            .setUser(metada)
                            .build();


                    userRepresentation = iamAdminServiceClient.enableUser(request);

                    if (userRepresentation != null) {

                        UserProfile profile = this.convertToProfile(userRepresentation);

                        org.apache.custos.user.profile.service.UserProfileRequest profileRequest =
                                org.apache.custos.user.profile.service.UserProfileRequest.newBuilder()
                                        .setProfile(profile)
                                        .setTenantId(request.getTenantId())
                                        .build();

                        userProfileClient.createUserProfile(profileRequest);
                    }
                }

                try {
                    return this.convert(userRepresentation);
                } catch (InternalErrorException | NotFoundException e) {
                    String msg = "Error occurred while converting user representation";
                    throw new CustosSCIMException(msg, e);
                }

            } else {
                String msg = "User not successfully registered";
                LOGGER.error(msg);
                throw new CustosSCIMException(msg);
            }

        } else {
            String msg = "Token not found ";
            LOGGER.error(msg);
            throw new CustosSCIMException(msg);
        }

    }


    @Override
    public User getUser(String id, Map<String, Boolean> map) throws CharonException, BadRequestException, NotFoundException {
        JSONObject object = new JSONObject(id);
        Object obj = object.get(Constants.CUSTOS_EXTENSION);
        String clientId = ((String) ((JSONObject) obj).get(Constants.CLIENT_ID));
        String clientSec = ((String) ((JSONObject) obj).get(Constants.CLIENT_SEC));
        String decodedId = ((String) ((JSONObject) obj).get(Constants.ID));
        String tenantId = ((String) ((JSONObject) obj).get(Constants.TENANT_ID));

        long tenant = Long.valueOf(tenantId);

        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setClientSecret(clientSec)
                .setTenantId(tenant)
                .build();
        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

        if (token != null && token.getAccessToken() != null) {

            UserSearchMetadata metada = UserSearchMetadata.newBuilder().setUsername(decodedId).build();

            UserSearchRequest request = UserSearchRequest
                    .newBuilder()
                    .setTenantId(tenant)
                    .setAccessToken(token.getAccessToken())
                    .setUser(metada)
                    .build();

            UserRepresentation userRep = iamAdminServiceClient.getUser(request);

            if (userRep == null || userRep.getUsername().equals("")) {
                throw new NotFoundException("User not found");
            }

            try {
                return convert(userRep);
            } catch (InternalErrorException e) {
                throw new CharonException(SCIMConstants.USER);
            }

        } else {
            String msg = "Token not found ";
            LOGGER.error(msg);
            throw new NotFoundException(msg);
        }
    }

    @Override
    public void deleteUser(String id) throws NotFoundException, CharonException, NotImplementedException, BadRequestException {
        JSONObject object = new JSONObject(id);
        Object obj = object.get(Constants.CUSTOS_EXTENSION);
        String clientId = ((String) ((JSONObject) obj).get(Constants.CLIENT_ID));
        String clientSec = ((String) ((JSONObject) obj).get(Constants.CLIENT_SEC));
        String decodedId = ((String) ((JSONObject) obj).get(Constants.ID));
        String tenantId = ((String) ((JSONObject) obj).get(Constants.TENANT_ID));

        long tenant = Long.valueOf(tenantId);

        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setClientSecret(clientSec)
                .setTenantId(tenant)
                .build();
        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

        if (token != null && token.getAccessToken() != null) {

            UserProfile profileReq = UserProfile.newBuilder().setUsername(decodedId).build();

            org.apache.custos.user.profile.service.UserProfileRequest req =
                    org.apache.custos.user.profile.service.UserProfileRequest
                            .newBuilder()
                            .setTenantId(tenant)
                            .setProfile(profileReq)
                            .build();

            UserProfile profile = userProfileClient.getUser(req);
            UserSearchMetadata metada = UserSearchMetadata.newBuilder().setUsername(decodedId).build();

            UserSearchRequest request = UserSearchRequest
                    .newBuilder()
                    .setTenantId(tenant)
                    .setAccessToken(token.getAccessToken())
                    .setUser(metada)
                    .build();

            if (profile != null && !profile.getUsername().trim().equals("")) {


                UserProfile deletedProfile = userProfileClient.deleteUser(req);

                if (deletedProfile != null) {
                    iamAdminServiceClient.deleteUser(request);
                } else {
                    String msg = "User profile deletion failed for " + decodedId;
                    LOGGER.error(msg);
                    throw new CharonException(msg);
                }

            } else {
                iamAdminServiceClient.deleteUser(request);

            }
        }
    }

    @Override
    public List<Object> listUsersWithPost(SearchRequest searchRequest, Map<String, Boolean> map) throws CharonException, NotImplementedException, BadRequestException {

        try {
            JSONObject obj = new JSONObject(searchRequest.getDomainName());
            String clientId = ((String) ((JSONObject) obj).get(Constants.CLIENT_ID));
            String clientSec = ((String) ((JSONObject) obj).get(Constants.CLIENT_SEC));
            String tenantId = ((String) ((JSONObject) obj).get(Constants.TENANT_ID));

            long tenant = Long.valueOf(tenantId);

            GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                    .newBuilder()
                    .setClientId(clientId)
                    .setClientSecret(clientSec)
                    .setTenantId(tenant)
                    .build();
            AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

            if (token != null && token.getAccessToken() != null) {

                FindUsersRequest findUsersRequest = FindUsersRequest.newBuilder()
                        .setAccessToken(token.getAccessToken())
                        .setOffset(searchRequest.getStartIndex() - 1)
                        .setTenantId(tenant)
                        .setLimit(searchRequest.getCount())
                        .build();

                FindUsersResponse userRep = iamAdminServiceClient.getUsers(findUsersRequest);

                List<Object> userList = userRep.getUsersList().stream().map(user -> {
                    try {
                        return convert(user);
                    } catch (Exception e) {
                        throw new RuntimeException(e);
                    }
                }).collect(Collectors.toList());
                List<Object> users = new ArrayList<>();

                users.add(userList.size());
                users.addAll(userList);
                return users;
            } else {
                throw new CustosSCIMException("Invalid Credentials", new UnauthorizedException());
            }
        } catch (Exception ex) {
            throw new CustosSCIMException(" Error occurred while fetching users ", ex);
        }
    }

    @Override
    public User updateUser(User user, Map<String, Boolean> map) throws NotImplementedException, CharonException, BadRequestException, NotFoundException {
        Attribute attribute = user.getAttribute(Constants.CUSTOS_EXTENSION);
        Attribute clientId = attribute.getSubAttribute(Constants.CLIENT_ID);
        Attribute clientSec = attribute.getSubAttribute(Constants.CLIENT_SEC);
        Attribute tenantId = attribute.getSubAttribute(Constants.TENANT_ID);
        String clId = ((SimpleAttribute) clientId).getStringValue();
        String clSec = ((SimpleAttribute) clientSec).getStringValue();
        String tenant = ((SimpleAttribute) tenantId).getStringValue();

        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                .newBuilder()
                .setClientId(clId)
                .setClientSecret(clSec)
                .setTenantId(Long.valueOf(tenant))
                .build();
        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

        if (token != null && token.getAccessToken() != null) {

            UserRepresentation userRepresentation = UserRepresentation
                    .newBuilder()
                    .setFirstName(user.getName().getGivenName())
                    .setLastName(user.getName().getFamilyName())
                    .setEmail(user.getEmails().get(0).getValue())
                    .setUsername(user.getUsername())
                    .build();

            UpdateUserProfileRequest registerUserRequest = UpdateUserProfileRequest.newBuilder()
                    .setTenantId(Long.valueOf(tenant))
                    .setAccessToken(token.getAccessToken())
                    .setUser(userRepresentation)
                    .build();

            OperationStatus status = iamAdminServiceClient.updateUserProfile(registerUserRequest);

            if (status.getStatus()) {
                UserSearchMetadata metada = UserSearchMetadata.newBuilder().setUsername(user.getUsername()).build();

                UserSearchRequest request = UserSearchRequest
                        .newBuilder()
                        .setTenantId(Long.valueOf(tenant))
                        .setAccessToken(token.getAccessToken())
                        .setUser(metada)
                        .build();
                UserRepresentation enabledUser = null;
                if (user.getActive()) {
                    enabledUser = iamAdminServiceClient.enableUser(request);
                } else {
                    enabledUser = iamAdminServiceClient.disableUser(request);
                }
                if (enabledUser != null) {

                    UserProfile profile = this.convertToProfile(enabledUser);

                    org.apache.custos.user.profile.service.UserProfileRequest profileRequest =
                            org.apache.custos.user.profile.service.UserProfileRequest.newBuilder()
                                    .setProfile(profile)
                                    .setTenantId(request.getTenantId())
                                    .build();

                    UserProfile exProfile = userProfileClient.getUser(profileRequest);

                    if (exProfile.getUsername().equals("")) {
                        userProfileClient.createUserProfile(profileRequest);
                    } else {
                        userProfileClient.updateUserProfile(profileRequest);
                    }
                }

                try {
                    return this.convert(enabledUser);
                } catch (InternalErrorException e) {
                    String msg = "Error occurred while converting user representation to charon";
                    throw new CharonException(msg, e);
                }

            } else {
                String msg = "User not successfully registered";
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

        } else {
            String msg = "Token not found ";
            LOGGER.error(msg);
            throw new RuntimeException(msg);
        }
    }


    @Override
    public Group createGroup(Group group, Map<String, Boolean> map) throws CharonException, ConflictException, NotImplementedException, BadRequestException {

//        List<Object> members = group.getMembers();
//
//        String str = (String) members.get(0);
//
//        JSONObject obj = new JSONObject(str);
//
//        String clientId = ((String) ((JSONObject) obj).get(Constants.CLIENT_ID));
//        String clientSec = ((String) ((JSONObject) obj).get(Constants.CLIENT_SEC));
//        String tenantId = ((String) ((JSONObject) obj).get(Constants.TENANT_ID));
//
//        long tenant = Long.valueOf(tenantId);
//
//        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
//                .newBuilder()
//                .setClientId(clientId)
//                .setClientSecret(clientSec)
//                .setTenantId(tenant)
//                .build();
//
//        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);
//
//        if (token != null && token.getAccessToken() != null) {
//
//            GroupRepresentation representation = GroupRepresentation
//                    .newBuilder()
//                    .setName(group.getDisplayName())
//                    .build();
//
//            GroupsRequest request = GroupsRequest
//                    .newBuilder()
//                    .setAccessToken(token.getAccessToken())
//                    .setTenantId(tenant)
//                    .setClientId(clientId)
//                    .addGroups(representation)
//                    .build();
//
//            GroupsResponse response = iamAdminServiceClient.createGroups(request);
//
//            org.apache.custos.user.profile.service.Group groupR =
//                    org.apache.custos.user.profile.service.Group
//                            .newBuilder()
//                            .setId(response.getGroups(0).getId())
//                            .setName(group.getDisplayName())
//                            .build();
//
//            org.apache.custos.user.profile.service.GroupRequest groupRequest =
//                    org.apache.custos.user.profile.service.GroupRequest
//                            .newBuilder()
//                            .setTenantId(tenant)
//                            .setGroup(groupR)
//                            .build();
//            userProfileClient.createGroup(groupRequest);
//
//            try {
//                return convert(response.getGroups(0));
//            } catch (InternalErrorException e) {
//                String msg = "Error occurred while converting group representation to charon";
//                throw new CharonException(msg, e);
//            }
//
//        } else {
//            String msg = "Token not found ";
//            LOGGER.error(msg);
//            throw new RuntimeException(msg);
//        }
        throw new NotImplementedException();

    }

    @Override
    public Group getGroup(String id, Map<String, Boolean> map) throws NotImplementedException, BadRequestException, CharonException, NotFoundException {
//        JSONObject object = new JSONObject(id);
//        Object obj = object.get(Constants.CUSTOS_EXTENSION);
//        String clientId = ((String) ((JSONObject) obj).get(Constants.CLIENT_ID));
//        String clientSec = ((String) ((JSONObject) obj).get(Constants.CLIENT_SEC));
//        String decodedId = ((String) ((JSONObject) obj).get(Constants.ID));
//        String tenantId = ((String) ((JSONObject) obj).get(Constants.TENANT_ID));
//        String accessToken = ((String) ((JSONObject) obj).get(Constants.ACCESS_TOKEN));
//
//        long tenant = Long.valueOf(tenantId);
//
//        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
//                .newBuilder()
//                .setClientId(clientId)
//                .setClientSecret(clientSec)
//                .setTenantId(tenant)
//                .build();
//        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);
//
//        if (token != null && token.getAccessToken() != null) {
//
//            GroupRepresentation groupRepresentation = GroupRepresentation.newBuilder().setId(decodedId).build();
//
//            GroupRequest groupsRequest = GroupRequest
//                    .newBuilder()
//                    .setAccessToken(token.getAccessToken())
//                    .setClientId(clientId)
//                    .setClientSec(clientSec)
//                    .setTenantId(tenant)
//                    .setGroup(groupRepresentation)
//                    .build();
//
//            GroupRepresentation representation = iamAdminServiceClient.findGroup(groupsRequest);
//
//            try {
//                return convert(representation);
//            } catch (InternalErrorException e) {
//                throw new CharonException(SCIMConstants.GROUP);
//            }
//
//        } else {
//            String msg = "Token not found ";
//            LOGGER.error(msg);
//            throw new RuntimeException(msg);
//        }
        throw new NotImplementedException();

    }

    @Override
    public void deleteGroup(String id) throws NotFoundException, CharonException, NotImplementedException, BadRequestException {
//        JSONObject object = new JSONObject(id);
//        Object obj = object.get(Constants.CUSTOS_EXTENSION);
//        String clientId = ((String) ((JSONObject) obj).get(Constants.CLIENT_ID));
//        String clientSec = ((String) ((JSONObject) obj).get(Constants.CLIENT_SEC));
//        String decodedId = ((String) ((JSONObject) obj).get(Constants.ID));
//        String tenantId = ((String) ((JSONObject) obj).get(Constants.TENANT_ID));
//        String accessToken = ((String) ((JSONObject) obj).get(Constants.ACCESS_TOKEN));
//
//        long tenant = Long.valueOf(tenantId);
//
//        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
//                .newBuilder()
//                .setClientId(clientId)
//                .setClientSecret(clientSec)
//                .setTenantId(tenant)
//                .build();
//        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);
//
//        if (token != null && token.getAccessToken() != null) {
//
//            GroupRepresentation groupRepresentation = GroupRepresentation.newBuilder().setId(decodedId).build();
//
//            GroupRequest groupsRequest = GroupRequest
//                    .newBuilder()
//                    .setAccessToken(token.getAccessToken())
//                    .setClientId(clientId)
//                    .setClientSec(clientSec)
//                    .setTenantId(tenant)
//                    .setGroup(groupRepresentation)
//                    .build();
//
//            OperationStatus response = iamAdminServiceClient.deleteGroup(groupsRequest);
//
//
//            if (response.getStatus()) {
//                org.apache.custos.user.profile.service.Group group = org.apache.custos.user.profile.service.Group.newBuilder()
//                        .setId(decodedId)
//                        .build();
//
//                org.apache.custos.user.profile.service.GroupRequest groupRequest = org.apache.custos.user.profile.service.GroupRequest
//                        .newBuilder().
//                                setTenantId(tenant).
//                                setGroup(group).build();
//
//                userProfileClient.deleteGroup(groupRequest);
//            }
//
//        } else {
//            String msg = "Token not found ";
//            LOGGER.error(msg);
//            throw new RuntimeException(msg);
//        }
         throw new NotImplementedException();

    }

    @Override
    public Group updateGroup(Group group, Group group1, Map<String, Boolean> map) throws NotImplementedException, BadRequestException, CharonException, NotFoundException {
//        List<Object> members = group1.getMembers();
//
//        String str = (String) members.get(0);
//
//        JSONObject obj = new JSONObject(str);
//
//        String clientId = ((String) ((JSONObject) obj).get(Constants.CLIENT_ID));
//        String clientSec = ((String) ((JSONObject) obj).get(Constants.CLIENT_SEC));
//        String tenantId = ((String) ((JSONObject) obj).get(Constants.TENANT_ID));
//
//        long tenant = Long.valueOf(tenantId);
//
//
//        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
//                .newBuilder()
//                .setClientId(clientId)
//                .setClientSecret(clientSec)
//                .setTenantId(tenant)
//                .build();
//        AuthToken token = identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);
//
//        if (token != null && token.getAccessToken() != null) {
//
//            GroupRepresentation representation = GroupRepresentation
//                    .newBuilder()
//                    .setId(group1.getId())
//                    .setName(group1.getDisplayName())
//                    .build();
//
//            GroupRequest request = GroupRequest
//                    .newBuilder()
//                    .setAccessToken(token.getAccessToken())
//                    .setTenantId(tenant)
//                    .setClientId(clientId)
//                    .setClientSec(clientSec)
//                    .setGroup(representation)
//                    .build();
//
//            GroupRepresentation response = iamAdminServiceClient.updateGroup(request);
//
//            org.apache.custos.user.profile.service.Group groupR =
//                    org.apache.custos.user.profile.service.Group
//                            .newBuilder()
//                            .setId(group1.getId())
//                            .setName(group1.getDisplayName())
//                            .build();
//
//            org.apache.custos.user.profile.service.GroupRequest groupRequest =
//                    org.apache.custos.user.profile.service.GroupRequest
//                            .newBuilder()
//                            .setTenantId(tenant)
//                            .setGroup(groupR)
//                            .build();
//            userProfileClient.updateGroup(groupRequest);
//
//            try {
//                return convert(response);
//            } catch (InternalErrorException e) {
//                String msg = "Error occurred while converting group representation to charon";
//                throw new CharonException(msg, e);
//            }
//
//        } else {
//            String msg = "Token not found ";
//            LOGGER.error(msg);
//            throw new RuntimeException(msg);
//        }
        throw new NotImplementedException();

    }

    @Override
    public List<Object> listGroupsWithPost(SearchRequest searchRequest, Map<String, Boolean> map)
            throws NotImplementedException, BadRequestException, CharonException {
        try {
            JSONObject obj = new JSONObject(searchRequest.getDomainName());
            String tenantId = ((String) ((JSONObject) obj).get(Constants.TENANT_ID));

            long tenant = Long.valueOf(tenantId);

            org.apache.custos.user.profile.service.GroupRequest groupsRequest =
                    org.apache.custos.user.profile.service.GroupRequest
                            .newBuilder()
                            .setTenantId(tenant)
                            .setLimit(searchRequest.getCount())
                            .setOffset(searchRequest.getStartIndex() - 1)
                            .build();

            GetAllGroupsResponse getAllGroupsResponse = userProfileClient.getAllGroups(groupsRequest);

            List<org.apache.custos.user.profile.service.Group> groupsList = getAllGroupsResponse.getGroupsList();

            List<Object> groups = new ArrayList<>();
            groups.add(groupsList.size());

            for (org.apache.custos.user.profile.service.Group group : groupsList) {
                org.apache.custos.user.profile.service.GroupRequest groupRequest =
                        org.apache.custos.user.profile.service.GroupRequest
                                .newBuilder()
                                .setTenantId(tenant)
                                .setGroup(group)
                                .build();
                GetAllUserProfilesResponse response = userProfileClient.getAllChildUsers(groupRequest);
                List<UserProfile> userProfileList = response.getProfilesList();
                Group gr = convert(group, userProfileList);
                groups.add(gr);
            }
            return groups;
        } catch (Exception ex) {
            throw new CustosSCIMException("Invalid Credentials", new UnauthorizedException());
        }
    }


    @Override
    public User getMe(String s, Map<String, Boolean> map) throws CharonException, BadRequestException, NotFoundException {
        throw new BadRequestException("Method not implemented");
    }

    @Override
    public User createMe(User user, Map<String, Boolean> map) throws CharonException, ConflictException, BadRequestException {
        throw new BadRequestException("Method not implemented");
    }

    @Override
    public void deleteMe(String s) throws NotFoundException, CharonException, NotImplementedException, BadRequestException {
        throw new NotImplementedException("Method not implemented");
    }

    @Override
    public User updateMe(User user, Map<String, Boolean> map) throws NotImplementedException, CharonException, BadRequestException, NotFoundException {
        throw new NotImplementedException("Method not implemented");
    }


    private User convert(UserRepresentation representation) throws BadRequestException, CharonException, InternalErrorException, NotFoundException {

        //obtain the json encoder
        JSONEncoder encoder = getEncoder();

        //obtain the json decoder
        JSONDecoder decoder = getDecoder();

        //obtain the schema corresponding to user
        // unless configured returns core-user schema or else returns extended user schema)
        SCIMResourceTypeSchema schema = SCIMResourceSchemaManager.getInstance().getUserResourceSchema();

        String scimObjectString = getUser(representation);

        //decode the SCIM User object, encoded in the submitted payload.
        User user = (User) decoder.decodeResource(scimObjectString, schema, new User());
        return user;

    }


    private Group convert(org.apache.custos.user.profile.service.Group group, List<UserProfile> members) throws
            BadRequestException, CharonException, InternalErrorException, NotFoundException {

        //obtain the json encoder
        JSONEncoder encoder = getEncoder();

        //obtain the json decoder
        JSONDecoder decoder = getDecoder();

        //obtain the schema corresponding to user
        // unless configured returns core-user schema or else returns extended user schema)
        SCIMResourceTypeSchema schema = SCIMResourceSchemaManager.getInstance().getGroupResourceSchema();

        String scimObjectString = getGroup(group, members);

        //decode the SCIM User object, encoded in the submitted payload.
        Group user = (Group) decoder.decodeResource(scimObjectString, schema, new Group());
        return user;

    }


    private String getUser(UserRepresentation representation) throws NotFoundException {
        JSONObject object = new JSONObject();
        object.put("id", representation.getUsername());
        object.put("externalId", representation.getUsername());
        object.put("userName", representation.getUsername());
        boolean active = representation.getState().equals("ACTIVE") ? true : false;
        object.put("active", active);
        JSONObject name = new JSONObject();
        name.put("familyName", representation.getLastName());
        name.put("givenName", representation.getFirstName());
        object.put("name", name);
        Instant instant = Instant.ofEpochMilli(Double.doubleToLongBits(representation.getCreationTime()));
        JSONObject meta = new JSONObject();
        meta.put("created", instant.toString());
        String location = AbstractResourceManager.getResourceEndpointURL(SCIMConstants.USER_ENDPOINT) + representation.getUsername();
        meta.put("location", location);
        meta.put("resourceType", SCIMConstants.USER);

        object.put("meta", meta);
        JSONArray array = new JSONArray();
        array.put(SCIMConstants.CORE_SCHEMA_URI);
        object.put("schemas", array);


        JSONArray emails = new JSONArray();
        emails.put(representation.getEmail());
        object.put("emails", emails);
        return object.toString();

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

            List<org.apache.custos.user.profile.service.UserAttribute> userAtrList = new ArrayList<>();
            attributeList.forEach(atr -> {
                org.apache.custos.user.profile.service.UserAttribute userAttribute =
                        org.apache.custos.user.profile.service.UserAttribute
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
        UserStatus userStatus = representation.getState().equals("ACTIVE") ? UserStatus.ACTIVE : UserStatus.SUSPENDED;
        profileBuilder.setStatus(userStatus);

        return profileBuilder.build();

    }

    private String getGroup(org.apache.custos.user.profile.service.Group group, List<UserProfile> members) throws NotFoundException {
        JSONObject object = new JSONObject();
        object.put("id", group.getId());
        object.put("displayName", group.getName());
        JSONObject meta = new JSONObject();
        String location = AbstractResourceManager.getResourceEndpointURL(SCIMConstants.GROUP_ENDPOINT) + group.getId();
        meta.put("location", location);
        meta.put("resourceType", SCIMConstants.GROUP);
        meta.put("created", group.getCreatedTime());
        meta.put("lastModified", group.getLastModifiedTime());

        object.put("meta", meta);
        JSONArray array = new JSONArray();
        array.put(SCIMConstants.CORE_SCHEMA_URI);
        object.put("schemas", array);


        JSONArray userArr = new JSONArray();
        for (UserProfile userProfile : members) {
            JSONObject user = new JSONObject();
            user.put("value", userProfile.getUsername());
            String userLoc = AbstractResourceManager.getResourceEndpointURL(SCIMConstants.USER_ENDPOINT)
                    + userProfile.getUsername();
            user.put("location", userLoc);
            user.put("display", userProfile.getFirstName());
            userArr.put(user);
        }
        object.put("members", userArr);
        return object.toString();
    }


}
