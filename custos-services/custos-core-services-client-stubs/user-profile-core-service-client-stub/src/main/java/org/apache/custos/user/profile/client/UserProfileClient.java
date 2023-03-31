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

package org.apache.custos.user.profile.client;

import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.user.profile.service.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * This class is used to connect with user profile service
 */
@Component
public class UserProfileClient {

    private ManagedChannel managedChannel;
    private UserProfileServiceGrpc.UserProfileServiceStub userProfileServiceStub;
    private UserProfileServiceGrpc.UserProfileServiceBlockingStub userProfileServiceBlockingStub;


    private final List<ClientInterceptor> clientInterceptorList;


    public UserProfileClient(List<ClientInterceptor> clientInterceptorList,
                             @Value("${core.services.server.hostname:localhost}") String serviceHost,
                             @Value("${core.services.server.port:7070}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext().intercept(clientInterceptorList).build();
        userProfileServiceStub = UserProfileServiceGrpc.newStub(managedChannel);
        userProfileServiceBlockingStub = UserProfileServiceGrpc.newBlockingStub(managedChannel);
    }


    public void createUserProfileAsync(UserProfileRequest profile, final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "Create user profile task failed");

        userProfileServiceStub.createUserProfile(profile, observer);
    }


    public void updateUserProfileAsync(UserProfileRequest profile, final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "Update user profile task failed");

        userProfileServiceStub.updateUserProfile(profile, observer);
    }


    public void getUserAsync(UserProfileRequest request, final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "get user profile task failed");

        userProfileServiceStub.getUserProfile(request, observer);
    }


    public void deleteUserAsync(UserProfileRequest request, final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "delete user profile task failed");

        userProfileServiceStub.deleteUserProfile(request, observer);
    }


    public void getUserProfileUpdateAuditTrailAsync(org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest request,
                                                    final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "get user profile update audit trail failed");

        userProfileServiceStub.getUserProfileAuditTrails(request, observer);
    }


    public void getAllUserProfilesInTenantAsync(UserProfileRequest request,
                                                final ServiceCallback callback) {

        StreamObserver observer = this.getObserver(callback, "get all user profiles in tenant async");

        userProfileServiceStub.getAllUserProfilesInTenant(request, observer);
    }

    public UserProfile createUserProfile(UserProfileRequest profile) {
        return userProfileServiceBlockingStub.createUserProfile(profile);
    }


    public UserProfile updateUserProfile(UserProfileRequest profile) {

        return userProfileServiceBlockingStub.updateUserProfile(profile);
    }


    public UserProfile getUser(UserProfileRequest request) {

        return userProfileServiceBlockingStub.getUserProfile(request);
    }


    public UserProfile deleteUser(UserProfileRequest request) {

        return userProfileServiceBlockingStub.deleteUserProfile(request);
    }


    public org.apache.custos.user.profile.service.GetUpdateAuditTrailResponse getUserProfileUpdateAuditTrail(
            org.apache.custos.user.profile.service.GetUpdateAuditTrailRequest request) {

        return userProfileServiceBlockingStub.getUserProfileAuditTrails(request);
    }


    public GetAllUserProfilesResponse getAllUserProfilesInTenant(UserProfileRequest request) {
        return userProfileServiceBlockingStub.getAllUserProfilesInTenant(request);
    }

    public GetAllUserProfilesResponse findUserProfilesByUserAttributes(UserProfileRequest request) {
        return userProfileServiceBlockingStub.findUserProfilesByAttributes(request);
    }

    public Group createGroup(GroupRequest request) {
        return userProfileServiceBlockingStub.createGroup(request);
    }

    public Group updateGroup(GroupRequest request) {
        return userProfileServiceBlockingStub.updateGroup(request);
    }

    public Group deleteGroup(GroupRequest request) {
        return userProfileServiceBlockingStub.deleteGroup(request);
    }

    public Group getGroup(GroupRequest request) {
        return userProfileServiceBlockingStub.getGroup(request);
    }


    public GetAllGroupsResponse getAllGroups(GroupRequest request) {
        return userProfileServiceBlockingStub.getAllGroups(request);
    }

    public Status addUserToGroup(GroupMembership request) {
        return userProfileServiceBlockingStub.addUserToGroup(request);
    }

    public Status removeUserFromGroup(GroupMembership request) {
        return userProfileServiceBlockingStub.removeUserFromGroup(request);
    }

    public Status addChildGroupToParentGroup(GroupToGroupMembership request) {
        return userProfileServiceBlockingStub.addChildGroupToParentGroup(request);
    }

    public Status removeChildGroupFromParentGroup(GroupToGroupMembership request) {
        return userProfileServiceBlockingStub.removeChildGroupFromParentGroup(request);
    }

    public GetAllGroupsResponse getAllGroupsOfUser(UserProfileRequest request) {
        return userProfileServiceBlockingStub.getAllGroupsOfUser(request);
    }

    public GetAllGroupsResponse getAllParentGroupsOfGroup(GroupRequest request) {

        return userProfileServiceBlockingStub.getAllParentGroupsOfGroup(request);
    }


    public GetAllUserProfilesResponse getAllChildUsers(GroupRequest request) {
        return userProfileServiceBlockingStub.getAllChildUsers(request);
    }

    public GetAllGroupsResponse getAllChildGroups(GroupRequest request) {
        return userProfileServiceBlockingStub.getAllChildGroups(request);
    }

    public Status changeUserMembershipType(GroupMembership request) {
        return userProfileServiceBlockingStub.changeUserMembershipType(request);
    }

    public Status addUserMembershipType(UserGroupMembershipTypeRequest request) {
        return userProfileServiceBlockingStub.addUserGroupMembershipType(request);
    }

    public Status removeUserMembershipType(UserGroupMembershipTypeRequest request) {
        return userProfileServiceBlockingStub.removeUserGroupMembershipType(request);
    }

    public Status hasAccess(GroupMembership request) {
        return userProfileServiceBlockingStub.hasAccess(request);
    }


    private StreamObserver getObserver(ServiceCallback callback, String failureMsg) {
        final Object[] response = new Object[1];
        StreamObserver observer = new StreamObserver() {
            @Override
            public void onNext(Object o) {
                response[0] = o;
            }

            @Override
            public void onError(Throwable throwable) {
                callback.onError(new ServiceException(failureMsg, throwable, null));
            }

            @Override
            public void onCompleted() {
                callback.onCompleted(response[0]);
            }
        };

        return observer;
    }

}
