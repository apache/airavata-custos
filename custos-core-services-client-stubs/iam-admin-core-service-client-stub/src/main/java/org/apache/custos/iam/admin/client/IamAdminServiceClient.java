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

package org.apache.custos.iam.admin.client;

import com.google.protobuf.Empty;
import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.service.*;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.bind.annotation.DeleteMapping;

import java.util.List;

/**
 * The client class used to connect IAM admin services
 */
@Component
public class IamAdminServiceClient {
    private ManagedChannel managedChannel;
    private IamAdminServiceGrpc.IamAdminServiceStub iamAdminServiceStub;
    private IamAdminServiceGrpc.IamAdminServiceBlockingStub iamAdminServiceBlockingStub;

    private final List<ClientInterceptor> clientInterceptorList;

    private String iamServerURL;


    public IamAdminServiceClient(List<ClientInterceptor> clientInterceptorList,
                                 @Value("${iam.admin.service.dns.name}") String serviceHost,
                                 @Value("${iam.admin.service.port}") int servicePort,
                                 @Value("${iam.server.url}") String url) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext(true).intercept(clientInterceptorList).build();
        iamAdminServiceStub = IamAdminServiceGrpc.newStub(managedChannel);
        iamAdminServiceBlockingStub = IamAdminServiceGrpc.newBlockingStub(managedChannel);
        this.iamServerURL = url;
    }


    public void setUPTenantAsync(SetUpTenantRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "Setup  tenant task failed");
        iamAdminServiceStub.setUPTenant(request, observer);
    }


    public void isUsernameAvailableAsync(UserSearchRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "isUsernameAvailable task failed");
        iamAdminServiceStub.isUsernameAvailable(request, observer);
    }


    public void isUserEnabledAsync(UserSearchRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "isUserEnabled task failed");
        iamAdminServiceStub.isUserEnabled(request, observer);
    }


    public void registerUserAsync(RegisterUserRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "registerUser task failed");
        iamAdminServiceStub.registerUser(request, observer);
    }


    public void enableUserAsync(UserSearchRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "enableUser task failed");
        iamAdminServiceStub.enableUser(request, observer);
    }


    public void isUserExistAsync(UserSearchRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "isUserExist task failed");
        iamAdminServiceStub.isUserExist(request, observer);
    }


    public void getUserAsync(UserSearchRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "getUser task failed");
        iamAdminServiceStub.getUser(request, observer);
    }


    public void getUsersAsync(FindUsersRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "getUsers task failed");
        iamAdminServiceStub.findUsers(request, observer);
    }


    public void resetPasswordAsync(ResetUserPassword request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "resetPassword task failed");
        iamAdminServiceStub.resetPassword(request, observer);
    }


    public void findUsersAsync(FindUsersRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "findUsers task failed");
        iamAdminServiceStub.findUsers(request, observer);
    }


    public void updateUserProfileAsync(UpdateUserProfileRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "updateUserProfile task failed");
        iamAdminServiceStub.updateUserProfile(request, observer);
    }

    public void deleteUserAsync(UserSearchRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "deleteUser task failed");
        iamAdminServiceStub.deleteUser(request, observer);
    }


    public void deleteRoleFromUserAsync(DeleteUserRolesRequest request, final ServiceCallback callback) {
        StreamObserver observer = this.getObserver(callback, "deleteRoleFromUser task failed");
        iamAdminServiceStub.deleteRolesFromUser(request, observer);
    }

    public void getOperationsMetadataAsync(GetOperationsMetadataRequest request, final ServiceCallback callback) {
        StreamObserver observer = getObserver(callback, "get operations metadata");
        iamAdminServiceStub.getOperationMetadata(request, observer);
    }

    public SetUpTenantResponse setUPTenant(SetUpTenantRequest request) {
        return iamAdminServiceBlockingStub.setUPTenant(request);
    }


    public OperationStatus isUsernameAvailable(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.isUsernameAvailable(request);
    }


    public OperationStatus isUserEnabled(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.isUserEnabled(request);
    }


    public OperationStatus addRolesToUsers(AddUserRolesRequest request) {
        return iamAdminServiceBlockingStub.addRolesToUsers(request);
    }


    public RegisterUserResponse registerUser(RegisterUserRequest request) {
        return iamAdminServiceBlockingStub.registerUser(request);

    }


    public UserRepresentation enableUser(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.enableUser(request);

    }

    public UserRepresentation disableUser(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.disableUser(request);

    }


    public CheckingResponse isUserExist(UserSearchRequest request) {

        return iamAdminServiceBlockingStub.isUserExist(request);
    }


    public UserRepresentation getUser(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.getUser(request);

    }


    public FindUsersResponse getUsers(FindUsersRequest request) {
        return iamAdminServiceBlockingStub.findUsers(request);

    }


    public OperationStatus resetPassword(ResetUserPassword request) {
        return iamAdminServiceBlockingStub.resetPassword(request);

    }


    public OperationStatus updateUserProfile(UpdateUserProfileRequest request) {
        return iamAdminServiceBlockingStub.updateUserProfile(request);
    }

    public OperationStatus deleteUser(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.deleteUser(request);
    }


    public OperationStatus deleteUserRoles(DeleteUserRolesRequest request) {
        return iamAdminServiceBlockingStub.deleteRolesFromUser(request);
    }

    public GetOperationsMetadataResponse getOperationsMetadata(GetOperationsMetadataRequest request) {
        return iamAdminServiceBlockingStub.getOperationMetadata(request);
    }

    public FederateIDPResponse configureFederatedIDP(ConfigureFederateIDPRequest request) {
        return iamAdminServiceBlockingStub.configureFederatedIDP(request);
    }


    public RegisterUsersResponse registerAndEnableUsers(RegisterUsersRequest registerUsersRequest) {
        return iamAdminServiceBlockingStub.registerAndEnableUsers(registerUsersRequest);
    }

    public AllRoles addRolesToTenant(AddRolesRequest rolesRequest) {
        return iamAdminServiceBlockingStub.addRolesToTenant(rolesRequest);
    }

    public OperationStatus addProtocolMapper(AddProtocolMapperRequest addProtocolMapper) {
        return iamAdminServiceBlockingStub.addProtocolMapper(addProtocolMapper);
    }

    public OperationStatus addUserAttributes(AddUserAttributesRequest addUserAttributesRequest) {
        return iamAdminServiceBlockingStub.addUserAttributes(addUserAttributesRequest);
    }

    public OperationStatus deleteUserAttributes(DeleteUserAttributeRequest deleteUserAttributeRequest) {
        return iamAdminServiceBlockingStub.deleteUserAttributes(deleteUserAttributeRequest);
    }

    public OperationStatus configureEventPersistence(EventPersistenceRequest persistenceRequest) {
        return iamAdminServiceBlockingStub.configureEventPersistence(persistenceRequest);
    }

    public GroupsResponse createGroups(GroupsRequest groupsRequest) {
        return iamAdminServiceBlockingStub.createGroups(groupsRequest);
    }

    public GroupRepresentation updateGroup(GroupRequest groupsRequest) {
        return iamAdminServiceBlockingStub.updateGroup(groupsRequest);
    }

    public OperationStatus deleteGroup(GroupRequest groupsRequest) {
        return iamAdminServiceBlockingStub.deleteGroup(groupsRequest);
    }

    public GroupRepresentation findGroup(GroupRequest groupsRequest) {
        return iamAdminServiceBlockingStub.findGroup(groupsRequest);
    }


    public GroupsResponse getAllGroups(GroupRequest request) {
        return iamAdminServiceBlockingStub.getAllGroups(request);
    }


    public OperationStatus addUserToGroup(UserGroupMappingRequest request) {
        return iamAdminServiceBlockingStub.addUserToGroup(request);
    }

    public OperationStatus removeUserFromGroup(UserGroupMappingRequest request) {
        return iamAdminServiceBlockingStub.removeUserFromGroup(request);
    }

    public SetUpTenantResponse createAgentClient(AgentClientMetadata request) {
        return iamAdminServiceBlockingStub.createAgentClient(request);
    }
    public OperationStatus configureAgentClient(AgentClientMetadata request) {
        return iamAdminServiceBlockingStub.configureAgentClient(request);
    }
    public OperationStatus isAgentNameAvailable(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.isAgentNameAvailable(request);
    }
    public RegisterUserResponse registerAndEnableAgent(RegisterUserRequest request) {
        return iamAdminServiceBlockingStub.registerAndEnableAgent(request);
    }
    public OperationStatus disableAgent(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.disableAgent(request);
    }
    public OperationStatus deleteAgent(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.deleteAgent(request);
    }
    public OperationStatus addAgentAttributes(AddUserAttributesRequest request) {
        return iamAdminServiceBlockingStub.addAgentAttributes(request);
    }
    public OperationStatus deleteAgentAttributes(DeleteUserAttributeRequest request) {
        return iamAdminServiceBlockingStub.deleteAgentAttributes(request);
    }
    public OperationStatus addRolesToAgent(AddUserRolesRequest request) {
        return iamAdminServiceBlockingStub.addRolesToAgent(request);
    }
    public OperationStatus deleteAgentRoles(DeleteUserRolesRequest request) {
        return iamAdminServiceBlockingStub.deleteAgentRoles(request);
    }

    public OperationStatus enableAgent(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.enableAgent(request);
    }

    public SetUpTenantResponse updateTenant(SetUpTenantRequest request) {
        return iamAdminServiceBlockingStub.updateTenant(request);
    }

    public Agent getAgent(UserSearchRequest request) {
        return iamAdminServiceBlockingStub.getAgent(request);
    }

    public String getIamServerURL() {
        return iamServerURL;
    }


    public Empty deleteTenant(DeleteTenantRequest request) {
        return iamAdminServiceBlockingStub.deleteTenant(request);
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
