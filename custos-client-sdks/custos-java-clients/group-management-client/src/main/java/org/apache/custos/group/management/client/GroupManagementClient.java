package org.apache.custos.group.management.client;

import io.grpc.ManagedChannel;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.group.management.service.GroupManagementServiceGrpc;
import org.apache.custos.iam.service.GroupRequest;
import org.apache.custos.iam.service.*;
import org.apache.custos.user.profile.service.*;

import java.io.IOException;
import java.util.Arrays;

/**
 * This contains group management related functions
 */
public class GroupManagementClient {

    private ManagedChannel managedChannel;

    private GroupManagementServiceGrpc.GroupManagementServiceBlockingStub blockingStub;


    public GroupManagementClient(String serviceHost, int servicePort, String clientId,
                                 String clientSecret) throws IOException {

        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getServerCertificate(serviceHost, clientId, clientSecret)) // public key
                        .build())
                .build();

        blockingStub = GroupManagementServiceGrpc.newBlockingStub(managedChannel);
        blockingStub = MetadataUtils.attachHeaders(blockingStub, ClientUtils.getAuthorizationHeader(clientId, clientSecret));
    }


    /**
     * Create groups
     *
     * @param clientId
     * @param groupRepresentation
     * @return
     */
    public GroupsResponse createKeycloakGroup(String clientId, GroupRepresentation[] groupRepresentation) {

        GroupsRequest request = GroupsRequest
                .newBuilder()
                .addAllGroups(Arrays.asList(groupRepresentation))
                .setClientId(clientId)
                .build();
        return blockingStub.createKeycloakGroups(request);

    }


    /**
     * update group
     *
     * @param clientId
     * @param groupRepresentation
     * @return
     */
    public GroupRepresentation updateKeycloakGroup(String clientId, GroupRepresentation groupRepresentation) {

        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .setClientId(clientId)
                .build();
        return blockingStub.updateKeycloakGroup(request);
    }

    /**
     * delete group
     *
     * @param clientId
     * @param groupRepresentation
     * @return
     */
    public OperationStatus deleteKeycloakGroup(String clientId, GroupRepresentation groupRepresentation) {

        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .setClientId(clientId)
                .build();
        return blockingStub.deleteKeycloakGroup(request);
    }


    /**
     * find group
     *
     * @param clientId
     * @param groupName
     * @param groupId
     * @return
     */
    public GroupRepresentation findKeycloakGroup(String clientId, String groupName, String groupId) {

        GroupRepresentation groupRepresentation =
                GroupRepresentation.newBuilder().build();
        if (groupName != null) {
            groupRepresentation = groupRepresentation.toBuilder().setName(groupName).build();
        }

        if (groupId != null) {
            groupRepresentation = groupRepresentation.toBuilder().setId(groupId).build();
        }

        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .setClientId(clientId)
                .build();
        return blockingStub.findKeycloakGroup(request);
    }


    /**
     * Get all groups
     *
     * @param clientId
     * @return
     */
    public GroupsResponse getAllKeycloakGroups(String clientId) {
        GroupRequest request = GroupRequest
                .newBuilder()
                .setClientId(clientId)
                .build();
        return blockingStub.getAllKeycloakGroups(request);
    }


    /**
     * Add user to group
     *
     * @param clientId
     * @param username
     * @param groupId
     * @return
     */
    public OperationStatus addUserToKeycloakGroup(String clientId, String username, String groupId, String type) {
        UserGroupMappingRequest request = UserGroupMappingRequest
                .newBuilder()
                .setUsername(username)
                .setClientId(clientId)
                .setMembershipType(type)
                .setGroupId(groupId).build();
        return blockingStub.addUserToKeycloakGroup(request);

    }

    /**
     * Remove user from group
     *
     * @param clientId
     * @param username
     * @param groupId
     * @return
     */
    public OperationStatus removeUserFromKeycloakGroup(String clientId, String username, String groupId) {

        UserGroupMappingRequest request = UserGroupMappingRequest
                .newBuilder()
                .setUsername(username)
                .setClientId(clientId)
                .setGroupId(groupId).build();
        return blockingStub.removeUserFromKeycloakGroup(request);

    }


    public Status addChildGroupToParentGroup(String clientId, String parentId, String childId) {
        GroupToGroupMembership membership = GroupToGroupMembership
                .newBuilder()
                .setChildId(childId)
                .setParentId(parentId)
                .setClientId(clientId)
                .build();

        return blockingStub.addChildGroupToParentGroup(membership);
    }


    public Status removeChildGroupFromParentGroup(String clientId, String parentId, String childId) {
        GroupToGroupMembership membership = GroupToGroupMembership
                .newBuilder()
                .setChildId(childId)
                .setParentId(parentId)
                .setClientId(clientId)
                .build();

        return blockingStub.removeChildGroupFromParentGroup(membership);
    }


    public GetAllGroupsResponse getAllGroupsOfUser(String clientId, String username) {
        UserProfile userProfile = UserProfile
                .newBuilder()
                .setUsername(username)
                .build();
        UserProfileRequest userProfileRequest = UserProfileRequest
                .newBuilder()
                .setProfile(userProfile)
                .setClientId(clientId)
                .build();

        return blockingStub.getAllGroupsOfUser(userProfileRequest);
    }


    public GetAllGroupsResponse getAllParentGroupsOfGroup(String clientId, String id) {
        Group group = Group
                .newBuilder()
                .setId(id)
                .build();

        org.apache.custos.user.profile.service.GroupRequest request =
                org.apache.custos.user.profile.service.GroupRequest
                        .newBuilder()
                        .setGroup(group)
                        .setClientId(clientId)
                        .build();

        return blockingStub.getAllParentGroupsOfGroup(request);
    }


    public GetAllUserProfilesResponse getAllChildUsers(String clientId, String id) {
        Group group = Group
                .newBuilder()
                .setId(id)
                .build();

        org.apache.custos.user.profile.service.GroupRequest request =
                org.apache.custos.user.profile.service.GroupRequest
                        .newBuilder()
                        .setGroup(group)
                        .setClientId(clientId)
                        .build();

        return blockingStub.getAllChildUsers(request);
    }


    public GetAllGroupsResponse getAllChildGroups(String clientId, String id) {
        Group group = Group
                .newBuilder()
                .setId(id)
                .build();

        org.apache.custos.user.profile.service.GroupRequest request =
                org.apache.custos.user.profile.service.GroupRequest
                        .newBuilder()
                        .setGroup(group)
                        .setClientId(clientId)
                        .build();

        return blockingStub.getAllChildGroups(request);
    }


    public Status changeUserMembershipType(String clientId, String username, String groupId, String type) {

        GroupMembership membership = GroupMembership
                .newBuilder()
                .setUsername(username)
                .setGroupId(groupId)
                .setType(type)
                .setClientId(clientId)
                .build();

        return blockingStub.changeUserMembershipType(membership);
    }


    public Status hasAccess(String clientId, String groupId, String userId, String type) {
        GroupMembership membership = GroupMembership
                .newBuilder()
                .setUsername(userId)
                .setGroupId(groupId)
                .setType(type)
                .setClientId(clientId)
                .build();

        return blockingStub.hasAccess(membership);

    }

}
