package org.apache.custos.agent.management.client;

import io.grpc.ManagedChannel;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.grpc.stub.MetadataUtils;
import org.apache.custos.agent.management.service.AgentManagementServiceGrpc;
import org.apache.custos.agent.management.service.AgentRegistrationResponse;
import org.apache.custos.agent.management.service.AgentSearchRequest;
import org.apache.custos.clients.core.ClientUtils;
import org.apache.custos.iam.service.*;

import java.io.IOException;
import java.util.Arrays;

/**
 * This class contains methods of Agent Management
 */
public class AgentManagementClient {

    private ManagedChannel managedChannel;

    private AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub;


    public AgentManagementClient(String serviceHost, int servicePort, String clientId,
                                 String clientSecret) throws IOException {

        managedChannel = NettyChannelBuilder.forAddress(serviceHost, servicePort)
                .sslContext(GrpcSslContexts
                        .forClient()
                        .trustManager(ClientUtils.getServerCertificate(serviceHost, clientId, clientSecret)) // public key
                        .build())
                .build();

        blockingStub = AgentManagementServiceGrpc.newBlockingStub(managedChannel);
        blockingStub = MetadataUtils.attachHeaders(blockingStub, ClientUtils.getAuthorizationHeader(clientId, clientSecret));
    }


    /**
     * Enable agents, this is should be the first operation to enable agent functions
     * @param adminToken
     * @return
     */
    public OperationStatus enableAgents(String adminToken) {

        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AgentClientMetadata metadata = AgentClientMetadata
                .newBuilder()
                .build();

        return blockingStub.enableAgents(metadata);
    }


    /**
     * Sets access token life time
     * @param adminToken
     * @param accessTokenLifeTime
     * @return
     */
    public OperationStatus configureAgentClient(String adminToken, long accessTokenLifeTime) {

        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AgentClientMetadata metadata = AgentClientMetadata
                .newBuilder()
                .setAccessTokenLifeTime(accessTokenLifeTime)
                .build();

        return blockingStub.configureAgentClient(metadata);
    }

    /**
     * This will register agent and activate agent
     * @param adminToken
     * @param id
     * @param realmRoles
     * @param agentAttributes
     * @return
     */
    public AgentRegistrationResponse registerAndEnableAgent(String adminToken, String id, String[] realmRoles, UserAttribute[] agentAttributes) {

        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        UserRepresentation userRepresentation = UserRepresentation
                .newBuilder()
                .setId(id)
                .addAllRealmRoles(Arrays.asList(realmRoles))
                .addAllAttributes(Arrays.asList(agentAttributes))
                .build();

        RegisterUserRequest userRequest = RegisterUserRequest
                .newBuilder()
                .setUser(userRepresentation)
                .build();

        return blockingStub.registerAndEnableAgent(userRequest);

    }


    /**
     * Get agent
     * @param adminToken
     * @param agentId
     * @return
     */
    public Agent getAgent(String adminToken, String agentId) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AgentSearchRequest request = AgentSearchRequest
                .newBuilder()
                .setId(agentId)
                .build();

        return blockingStub.getAgent(request);
    }

    /**
     * Delete agent
     * @param adminToken
     * @param agentId
     * @return
     */
    public OperationStatus deleteAgent(String adminToken, String agentId) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AgentSearchRequest request = AgentSearchRequest
                .newBuilder()
                .setId(agentId)
                .build();

        return blockingStub.deleteAgent(request);
    }


    /**
     * Disable agent
     * @param adminToken
     * @param agentId
     * @return
     */
    public OperationStatus disableAgent(String adminToken, String agentId) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AgentSearchRequest request = AgentSearchRequest
                .newBuilder()
                .setId(agentId)
                .build();

        return blockingStub.disableAgent(request);

    }


    /**
     * Enable agent
     * @param adminToken
     * @param agentId
     * @return
     */
    public OperationStatus enableAgent(String adminToken, String agentId) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AgentSearchRequest request = AgentSearchRequest
                .newBuilder()
                .setId(agentId)
                .build();

        return blockingStub.enableAgent(request);
    }


    /**
     * Add agent attributes
     * @param adminToken
     * @param agents
     * @param attributes
     * @return
     */
    public OperationStatus addAgentAttributes(String adminToken, String[] agents, UserAttribute[] attributes) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AddUserAttributesRequest addUserAttributesRequest = AddUserAttributesRequest
                .newBuilder()
                .addAllAttributes(Arrays.asList(attributes))
                .addAllAgents(Arrays.asList(agents))
                .build();
        return blockingStub.addAgentAttributes(addUserAttributesRequest);

    }


    /**
     * Delete agent attributes
     * @param adminToken
     * @param agents
     * @param attributes
     * @return
     */
    public OperationStatus deleteAgentAttributes(String adminToken, String[] agents, UserAttribute[] attributes) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        DeleteUserAttributeRequest addUserAttributesRequest = DeleteUserAttributeRequest
                .newBuilder()
                .addAllAttributes(Arrays.asList(attributes))
                .addAllAgents(Arrays.asList(agents))
                .build();
        return blockingStub.deleteAgentAttributes(addUserAttributesRequest);
    }


    /**
     * Add roles to agents
     * @param adminToken
     * @param agents
     * @param roles
     * @return
     */
    public OperationStatus addRolesToAgents(String adminToken, String[] agents, String[] roles) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AddUserRolesRequest addUserRolesRequest = AddUserRolesRequest
                .newBuilder()
                .addAllAgents(Arrays.asList(agents))
                .addAllRoles(Arrays.asList(roles))
                .build();
        return blockingStub.addRolesToAgent(addUserRolesRequest);

    }


    /**
     * Delete roles from agents
     * @param adminToken
     * @param id
     * @param roles
     * @return
     */
    public OperationStatus deleteRolesFromAgents(String adminToken, String id, String[] roles) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        DeleteUserRolesRequest deleteUserRolesRequest = DeleteUserRolesRequest
                .newBuilder()
                .addAllRoles(Arrays.asList(roles))
                .setId(id)
                .build();
        return blockingStub.deleteRolesFromAgent(deleteUserRolesRequest);

    }

    /**
     * Add protocol mapper, this will add attributes and roles to token
     * @param adminToken
     * @param name
     * @param attributeName
     * @param claimName
     * @param claimType
     * @param mapperType
     * @param addToIdToken
     * @param addToAccessToken
     * @param addToUserInfo
     * @param multiValued
     * @param aggregrateAttributeValues
     * @return
     */
    public OperationStatus addProtocolMapper(String adminToken, String name, String attributeName, String claimName, String claimType, String mapperType,
                                             boolean addToIdToken, boolean addToAccessToken, boolean addToUserInfo, boolean multiValued,
                                             boolean aggregrateAttributeValues) {
        AgentManagementServiceGrpc.AgentManagementServiceBlockingStub blockingStub =
                MetadataUtils.attachHeaders(this.blockingStub, ClientUtils.getAuthorizationHeader(adminToken));

        AddProtocolMapperRequest mapperRequest = AddProtocolMapperRequest
                .newBuilder()
                .setName(name)
                .setAttributeName(attributeName)
                .setClaimName(claimName)
                .setClaimType(ClaimJSONTypes.valueOf(claimType))
                .setMapperType(MapperTypes.valueOf(mapperType))
                .setAddToIdToken(addToIdToken)
                .setAddToAccessToken(addToAccessToken)
                .setAddToUserInfo(addToUserInfo)
                .setMultiValued(multiValued)
                .setAggregateAttributeValues(aggregrateAttributeValues)
                .build();
        return blockingStub.addProtocolMapper(mapperRequest);


    }


}
