package org.apache.custos.user.management.client;/*
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


import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.AbstractClient;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.iam.service.*;
import org.apache.custos.user.management.service.UserManagementServiceGrpc;
import org.apache.custos.user.management.service.UserProfileRequest;
import org.apache.custos.user.profile.service.GetAllUserProfilesResponse;
import org.apache.custos.user.profile.service.UserProfile;

import java.io.IOException;
import java.util.Arrays;

/**
 * The class containes operations permitted for user management client
 */
public class UserManagementClient extends AbstractClient {

    private UserManagementServiceGrpc.UserManagementServiceBlockingStub blockingStub;

    private String clientId;

    private String clientSec;


    public UserManagementClient(String serviceHost, int servicePort, String clientId,
                                String clientSecret) throws IOException {
        super(serviceHost, servicePort, clientId, clientSecret);
        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getServerCertificate(serviceHost, clientId, clientSecret)) // public key
                        .build())
                .build();

        blockingStub = UserManagementServiceGrpc.newBlockingStub(managedChannel);


        blockingStub = MetadataUtils.attachHeaders(blockingStub,
                ClientUtils.getAuthorizationHeader(clientId, clientSecret));

        this.clientId = clientId;

        this.clientSec = clientSecret;

    }

    public RegisterUserResponse registerUser(String username, String firstName, String lastName,
                                             String password, String email, boolean isTempPassword) {

        UserRepresentation userRepresentation = UserRepresentation
                .newBuilder()
                .setUsername(username)
                .setFirstName(firstName)
                .setLastName(lastName)
                .setPassword(password)
                .setEmail(email)
                .setTemporaryPassword(isTempPassword)
                .build();

        RegisterUserRequest registerUserRequest = RegisterUserRequest
                .newBuilder()
                .setUser(userRepresentation)
                .build();

        return blockingStub.registerUser(registerUserRequest);

    }


    public UserRepresentation enableUser(String userName) {

        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(userName)
                .build();


        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata).build();

        return blockingStub.enableUser(request);
    }

    public OperationStatus addUserAttributes(String adminToken, UserAttribute[] attributes, String[] users) {

        UserManagementServiceGrpc.UserManagementServiceBlockingStub unAuthorizedStub =
                UserManagementServiceGrpc.newBlockingStub(managedChannel);
        unAuthorizedStub =
                MetadataUtils.attachHeaders(unAuthorizedStub, ClientUtils.getAuthorizationHeader(adminToken));

        AddUserAttributesRequest request = AddUserAttributesRequest
                .newBuilder()
                .addAllAttributes(Arrays.asList(attributes))
                .addAllUsers(Arrays.asList(users))
                .build();
        return unAuthorizedStub.addUserAttributes(request);

    }

    public OperationStatus deleteUserAttributes(String adminToken, UserAttribute[] attributes, String[] users) {

        UserManagementServiceGrpc.UserManagementServiceBlockingStub unAuthorizedStub =
                UserManagementServiceGrpc.newBlockingStub(managedChannel);
        unAuthorizedStub =
                MetadataUtils.attachHeaders(unAuthorizedStub, ClientUtils.getAuthorizationHeader(adminToken));


        DeleteUserAttributeRequest request = DeleteUserAttributeRequest
                .newBuilder()
                .addAllAttributes(Arrays.asList(attributes))
                .addAllUsers(Arrays.asList(users))
                .build();
        return unAuthorizedStub.deleteUserAttributes(request);

    }


    public OperationStatus addRolesToUsers(String adminToken, String[] roles, String[] username, boolean isClientLevel) {
        UserManagementServiceGrpc.UserManagementServiceBlockingStub unAuthorizedStub =
                UserManagementServiceGrpc.newBlockingStub(managedChannel);
        unAuthorizedStub =
                MetadataUtils.attachHeaders(unAuthorizedStub, ClientUtils.getAuthorizationHeader(adminToken));

        AddUserRolesRequest request = AddUserRolesRequest
                .newBuilder()
                .addAllRoles(Arrays.asList(roles))
                .addAllUsernames(Arrays.asList(username))
                .setClientLevel(isClientLevel)
                .build();
        return unAuthorizedStub.addRolesToUsers(request);

    }


    public OperationStatus deleteUserRoles(String adminToken, String[] clientRoles, String[] realmRoles, String username) {
        UserManagementServiceGrpc.UserManagementServiceBlockingStub unAuthorizedStub =
                UserManagementServiceGrpc.newBlockingStub(managedChannel);
        unAuthorizedStub =
                MetadataUtils.attachHeaders(unAuthorizedStub, ClientUtils.getAuthorizationHeader(adminToken));

        DeleteUserRolesRequest request = DeleteUserRolesRequest
                .newBuilder()
                .addAllClientRoles(Arrays.asList(clientRoles))
                .addAllRoles(Arrays.asList(realmRoles))
                .setUsername(username)
                .build();
        return unAuthorizedStub.deleteUserRoles(request);

    }


    public OperationStatus isUserEnabled(String username) {

        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(username)
                .build();

        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata)
                .build();

        return blockingStub.isUserEnabled(request);
    }


    public OperationStatus isUsernameAvailable(String username) {

        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(username)
                .build();

        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata)
                .build();

        return blockingStub.isUsernameAvailable(request);
    }


    public UserRepresentation getUser(String username) {
        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(username)
                .build();

        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata)
                .build();

        return blockingStub.getUser(request);

    }


    public FindUsersResponse findUser(String username, String firstName, String lastName, String email, int offset, int limit) {

        UserSearchMetadata.Builder builder = UserSearchMetadata
                .newBuilder();

        if (username != null) {
            builder = builder.setUsername(username);
        }

        if (firstName != null) {
            builder = builder.setFirstName(firstName);
        }

        if (lastName != null) {
            builder = builder.setLastName(lastName);
        }

        if (email != null) {
            builder = builder.setEmail(email);
        }
        UserSearchMetadata metadata = builder.build();

        FindUsersRequest request = FindUsersRequest
                .newBuilder()
                .setUser(metadata)
                .setLimit(limit)
                .setOffset(offset)
                .build();

        return blockingStub.findUsers(request);

    }


    public OperationStatus resetUserPassword(String username, String password) {

        ResetUserPassword userPassword = ResetUserPassword
                .newBuilder()
                .setPassword(password)
                .setUsername(username)
                .build();

        return blockingStub.resetPassword(userPassword);

    }


    public OperationStatus deleteUser(String adminToken, String username) {

        UserManagementServiceGrpc.UserManagementServiceBlockingStub unAuthorizedStub =
                UserManagementServiceGrpc.newBlockingStub(managedChannel);
        unAuthorizedStub =
                MetadataUtils.attachHeaders(unAuthorizedStub, ClientUtils.getAuthorizationHeader(adminToken));

        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(username)
                .build();

        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata)
                .build();

        return unAuthorizedStub.deleteUser(request);

    }


    public RegisterUserResponse registerUser(String username, String firstName, String lastName,
                                             String password, String email, boolean isTempPassword,
                                             String clientId) {

        UserRepresentation userRepresentation = UserRepresentation
                .newBuilder()
                .setUsername(username)
                .setFirstName(firstName)
                .setLastName(lastName)
                .setPassword(password)
                .setEmail(email)
                .setTemporaryPassword(isTempPassword)
                .build();

        RegisterUserRequest registerUserRequest = RegisterUserRequest
                .newBuilder()
                .setClientId(clientId)
                .setUser(userRepresentation)
                .build();

        return blockingStub.registerUser(registerUserRequest);

    }


    public UserRepresentation enableUser(String userName, String clientId) {

        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(userName)
                .build();


        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setClientId(clientId)
                .setUser(metadata).build();

        return blockingStub.enableUser(request);
    }

    public OperationStatus addUserAttributes(UserAttribute[] attributes, String[] users, String clientId) {


        AddUserAttributesRequest request = AddUserAttributesRequest
                .newBuilder()
                .addAllAttributes(Arrays.asList(attributes))
                .addAllUsers(Arrays.asList(users))
                .setClientId(clientId)
                .build();
        return blockingStub.addUserAttributes(request);

    }

    public OperationStatus deleteUserAttributes(UserAttribute[] attributes, String[] users, String clientId) {


        DeleteUserAttributeRequest request = DeleteUserAttributeRequest
                .newBuilder()
                .addAllAttributes(Arrays.asList(attributes))
                .addAllUsers(Arrays.asList(users))
                .setClientId(clientId)
                .build();
        return blockingStub.deleteUserAttributes(request);

    }


    public OperationStatus addRolesToUsers(String[] roles, String[] username,
                                           boolean isClientLevel, String clientId, String adminToken) {
        UserManagementServiceGrpc.UserManagementServiceBlockingStub unAuthorizedStub =
                UserManagementServiceGrpc.newBlockingStub(managedChannel);
        unAuthorizedStub =
                MetadataUtils.attachHeaders(unAuthorizedStub, ClientUtils.getUserTokenHeader(adminToken));
        unAuthorizedStub = MetadataUtils.attachHeaders(unAuthorizedStub,
                ClientUtils.getAuthorizationHeader(this.clientId, this.clientSec));

        AddUserRolesRequest request = AddUserRolesRequest
                .newBuilder()
                .addAllRoles(Arrays.asList(roles))
                .addAllUsernames(Arrays.asList(username))
                .setClientLevel(isClientLevel)
                .setClientId(clientId)
                .build();
        return unAuthorizedStub.addRolesToUsers(request);

    }


    public OperationStatus deleteUserRoles(String[] clientRoles,
                                           String[] realmRoles, String username, String clientId, String adminToken) {
        UserManagementServiceGrpc.UserManagementServiceBlockingStub unAuthorizedStub =
                UserManagementServiceGrpc.newBlockingStub(managedChannel);
        unAuthorizedStub =
                MetadataUtils.attachHeaders(unAuthorizedStub, ClientUtils.getUserTokenHeader(adminToken));
        unAuthorizedStub = MetadataUtils.attachHeaders(unAuthorizedStub,
                ClientUtils.getAuthorizationHeader(this.clientId, this.clientSec));

        DeleteUserRolesRequest request = DeleteUserRolesRequest
                .newBuilder()
                .addAllClientRoles(Arrays.asList(clientRoles))
                .addAllRoles(Arrays.asList(realmRoles))
                .setUsername(username)
                .setClientId(clientId)
                .build();
        return unAuthorizedStub.deleteUserRoles(request);

    }


    public OperationStatus isUserEnabled(String username, String clientId) {

        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(username)
                .build();

        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata)
                .setClientId(clientId)
                .build();

        return blockingStub.isUserEnabled(request);
    }


    public OperationStatus isUsernameAvailable(String username, String clientId) {

        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(username)
                .build();

        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata)
                .setClientId(clientId)
                .build();

        return blockingStub.isUsernameAvailable(request);
    }


    public UserRepresentation getUser(String username, String clientId) {
        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(username)
                .build();

        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata)
                .setClientId(clientId)
                .build();

        return blockingStub.getUser(request);

    }


    public GetAllUserProfilesResponse getAllUserProfiles(String clientId) {
        UserProfileRequest request = UserProfileRequest
                .newBuilder()
                .setClientId(clientId)
                .build();

        return blockingStub.getAllUserProfilesInTenant(request);

    }


    public FindUsersResponse findUsers(String searchString, String username, String firstName,
                                       String lastName, String email, int offset, int limit, String clientId) {

        UserSearchMetadata.Builder builder = UserSearchMetadata
                .newBuilder();

        if (username != null) {
            builder = builder.setUsername(username);
        }

        if (firstName != null) {
            builder = builder.setFirstName(firstName);
        }

        if (lastName != null) {
            builder = builder.setLastName(lastName);
        }

        if (email != null) {
            builder = builder.setEmail(email);
        }

        if (searchString != null) {
            builder = builder.setId(searchString);
        }
        UserSearchMetadata metadata = builder.build();

        FindUsersRequest request = FindUsersRequest
                .newBuilder()
                .setUser(metadata)
                .setLimit(limit)
                .setOffset(offset)
                .setClientId(clientId)
                .build();

        return blockingStub.findUsers(request);

    }


    public OperationStatus resetUserPassword(String username, String password, String clientId) {

        ResetUserPassword userPassword = ResetUserPassword
                .newBuilder()
                .setPassword(password)
                .setUsername(username)
                .setClientId(clientId)
                .build();

        return blockingStub.resetPassword(userPassword);

    }


    public OperationStatus deleteUser(String username, String clientId, String adminToken) {

        UserManagementServiceGrpc.UserManagementServiceBlockingStub stub =
                UserManagementServiceGrpc.newBlockingStub(managedChannel);
        stub = MetadataUtils.attachHeaders(stub,
                ClientUtils.getAuthorizationHeader(this.clientId, this.clientSec));
        MetadataUtils.attachHeaders(stub, ClientUtils.getUserTokenHeader(adminToken));

        UserSearchMetadata metadata = UserSearchMetadata
                .newBuilder()
                .setUsername(username)
                .build();

        UserSearchRequest request = UserSearchRequest
                .newBuilder()
                .setUser(metadata)
                .setClientId(clientId)
                .build();

        return stub.deleteUser(request);

    }


    public UserProfile updateUserProfile(String username, String firstName, String lastName, String email,
                                         String clientId) {


        UserProfile userProfile = UserProfile.newBuilder()
                .setUsername(username)
                .setFirstName(firstName)
                .setLastName(lastName)
                .setEmail(email)
                .build();

        UserProfileRequest userProfileRequest = UserProfileRequest
                .newBuilder()
                .setClientId(clientId)
                .setUserProfile(userProfile)
                .build();

        return blockingStub.updateUserProfile(userProfileRequest);

    }

}
