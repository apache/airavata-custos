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
     *
     * @param clientId
     * @param groupRepresentation
     * @return
     */
    public GroupsResponse createGroup(String clientId, GroupRepresentation[] groupRepresentation) {

        GroupsRequest request = GroupsRequest
                .newBuilder()
                .addAllGroups(Arrays.asList(groupRepresentation))
                .setClientId(clientId)
                .build();
        return blockingStub.createGroups(request);

    }


    /**
     * update group
     *
     * @param clientId
     * @param groupRepresentation
     * @return
     */
    public GroupRepresentation updateGroup(String clientId, GroupRepresentation groupRepresentation) {

        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .setClientId(clientId)
                .build();
        return blockingStub.updateGroup(request);
    }

    /**
     * delete group
     *
     * @param clientId
     * @param groupRepresentation
     * @return
     */
    public OperationStatus deleteGroup(String clientId, GroupRepresentation groupRepresentation) {

        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .setClientId(clientId)
                .build();
        return blockingStub.deleteGroup(request);
    }


    /**
     * find group
     *
     * @param clientId
     * @param groupName
     * @param groupId
     * @return
     */
    public OperationStatus findGroup(String clientId, String groupName, String groupId) {
        GroupRepresentation groupRepresentation =
                GroupRepresentation.newBuilder().setId(groupId).setName(groupName).build();
        GroupRequest request = GroupRequest
                .newBuilder()
                .setGroup(groupRepresentation)
                .setClientId(clientId)
                .build();
        return blockingStub.deleteGroup(request);
    }


    /**
     * Get all groups
     *
     * @param clientId
     * @return
     */
    public GroupsResponse getAllGroups(String clientId) {
        GroupRequest request = GroupRequest
                .newBuilder()
                .setClientId(clientId)
                .build();
        return blockingStub.getAllGroups(request);
    }


    /**
     * Add user to group
     *
     * @param clientId
     * @param username
     * @param groupId
     * @return
     */
    public OperationStatus addUserToGroup(String clientId, String username, String groupId) {
        UserGroupMappingRequest request = UserGroupMappingRequest
                .newBuilder()
                .setUsername(username)
                .setClientId(clientId)
                .setGroupId(groupId).build();
        return blockingStub.addUserToGroup(request);

    }

    /**
     * Remove user from group
     *
     * @param clientId
     * @param username
     * @param groupId
     * @return
     */
    public OperationStatus removeUserFromGroup(String clientId, String username, String groupId) {

        UserGroupMappingRequest request = UserGroupMappingRequest
                .newBuilder()
                .setUsername(username)
                .setClientId(clientId)
                .setGroupId(groupId).build();
        return blockingStub.removeUserFromGroup(request);

    }

}
