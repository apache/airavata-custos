package org.apache.custos.group.management.client;

import io.grpc.ManagedChannel;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.group.management.service.GroupManagementServiceGrpc;
import org.apache.custos.iam.service.*;

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
     * @param adminToken
     * @param groupRepresentation
     * @return
     */
    public GroupsResponse createGroup(String adminToken, GroupRepresentation[] groupRepresentation) {

        GroupManagementServiceGrpc.GroupManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        GroupsRequest request = GroupsRequest
                .newBuilder()
                .addAllGroups(Arrays.asList(groupRepresentation))
                .build();
        return blockingStub.createGroups(request);

    }


    /**
     * update group
     * @param adminToken
     * @param groupRepresentation
     * @return
     */
    public GroupRepresentation updateGroup(String adminToken, GroupRepresentation groupRepresentation) {
        GroupManagementServiceGrpc.GroupManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .build();
        return blockingStub.updateGroup(request);
    }

    /**
     * delete group
     * @param adminToken
     * @param groupRepresentation
     * @return
     */
    public OperationStatus deleteGroup(String adminToken, GroupRepresentation groupRepresentation) {
        GroupManagementServiceGrpc.GroupManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .build();
        return blockingStub.deleteGroup(request);
    }


    /**
     * find group
     * @param adminToken
     * @param groupName
     * @param groupId
     * @return
     */
    public OperationStatus findGroup(String adminToken, String groupName, String groupId) {
        GroupManagementServiceGrpc.GroupManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        GroupRepresentation groupRepresentation =
                GroupRepresentation.newBuilder().setId(groupId).setName(groupName).build();
        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .build();
        return blockingStub.deleteGroup(request);
    }


    /**
     * Get all groups
     * @param adminToken
     * @return
     */
    public GroupsResponse getAllGroups(String adminToken) {
        GroupManagementServiceGrpc.GroupManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        GroupRequest request = GroupRequest
                .newBuilder()
                .build();
        return blockingStub.getAllGroups(request);
    }


    /**
     * Add user to group
     * @param adminToken
     * @param username
     * @param groupId
     * @return
     */
    public OperationStatus addUserToGroup(String adminToken, String username, String groupId) {
        GroupManagementServiceGrpc.GroupManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        UserGroupMappingRequest request = UserGroupMappingRequest
                .newBuilder()
                .setUsername(username)
                .setGroupId(groupId).build();
        return blockingStub.addUserToGroup(request);

    }

    /**
     * Remove user from group
     * @param adminToken
     * @param username
     * @param groupId
     * @return
     */
    public OperationStatus removeUserFromGroup(String adminToken, String username, String groupId) {
        GroupManagementServiceGrpc.GroupManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        UserGroupMappingRequest request = UserGroupMappingRequest
                .newBuilder()
                .setUsername(username)
                .setGroupId(groupId).build();
        return blockingStub.removeUserFromGroup(request);

    }

}
