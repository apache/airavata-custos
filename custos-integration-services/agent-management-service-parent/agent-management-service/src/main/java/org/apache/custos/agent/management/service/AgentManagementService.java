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

package org.apache.custos.agent.management.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.agent.profile.client.AgentProfileClient;
import org.apache.custos.agent.profile.service.Agent;
import org.apache.custos.agent.profile.service.AgentAttribute;
import org.apache.custos.agent.profile.service.AgentRequest;
import org.apache.custos.agent.profile.service.AgentStatus;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.credential.store.service.CredentialMetadata;
import org.apache.custos.credential.store.service.GetCredentialRequest;
import org.apache.custos.credential.store.service.Type;
import org.apache.custos.iam.admin.client.IamAdminServiceClient;
import org.apache.custos.iam.service.*;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.GetTenantRequest;
import org.apache.custos.tenant.profile.service.GetTenantResponse;
import org.apache.custos.tenant.profile.service.Tenant;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.List;

@GRpcService
public class AgentManagementService extends org.apache.custos.agent.management.service.AgentManagementServiceGrpc.AgentManagementServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(AgentManagementService.class);

    @Autowired
    private AgentProfileClient agentProfileClient;

    @Autowired
    private IamAdminServiceClient iamAdminServiceClient;

    @Autowired
    private IdentityClient identityClient;

    @Autowired
    private TenantProfileClient tenantProfileClient;

    @Autowired
    private CredentialStoreServiceClient credentialStoreServiceClient;


    private static final String AGENT_CLIENT = "agent-client";

    private static final String CUSTOS_REALM_AGENT = "custos-realm-agent";


    @Override
    public void enableAgents(AgentClientMetadata request, StreamObserver<OperationStatus> responseObserver) {
        try {

            LOGGER.debug("Request received to enable agent " + request.getTenantId());


            GetTenantRequest tenantRequest = GetTenantRequest
                    .newBuilder()
                    .setTenantId(request.getTenantId()).build();

            GetTenantResponse response = tenantProfileClient.getTenant(tenantRequest);

            Tenant tenant = response.getTenant();

            if (tenant == null || tenant.getTenantId() == 0) {
                String msg = "Tenant not found  for " + request.getTenantId();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            List<String> redirectURIs = new ArrayList<>();
            String redirectURI = tenant.getClientUri() + "/agent/callback";
            redirectURIs.add(redirectURI);

            request = request.toBuilder()
                    .setTenantURL(tenant.getClientUri())
                    .addAllRedirectURIs(redirectURIs)
                    .setClientName(AGENT_CLIENT)
                    .build();

            GetCredentialRequest credentialRequest = GetCredentialRequest
                    .newBuilder()
                    .setType(Type.AGENT_CLIENT)
                    .setOwnerId(request.getTenantId())
                    .build();

            CredentialMetadata credentialMetadata = this.credentialStoreServiceClient.getCredential(credentialRequest);
            if (credentialMetadata != null && !credentialMetadata.getId().equals("")) {
                OperationStatus operationStatus = OperationStatus.newBuilder().setStatus(true).build();
                responseObserver.onNext(operationStatus);
                responseObserver.onCompleted();
                return;
            }


            SetUpTenantResponse setUpTenantResponse = iamAdminServiceClient.createAgentClient(request);


            if (setUpTenantResponse == null || setUpTenantResponse.getClientId().equals("") ||
                    setUpTenantResponse.getClientSecret().equals("")) {
                String msg = "Agent activation failed for tenant " + request.getTenantId();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            CredentialMetadata metadata = CredentialMetadata.newBuilder().
                    setId(setUpTenantResponse.getClientId())
                    .setSecret(setUpTenantResponse.getClientSecret())
                    .setOwnerId(request.getTenantId())
                    .setType(Type.AGENT_CLIENT)
                    .build();
            org.apache.custos.credential.store.service.OperationStatus status = credentialStoreServiceClient.putCredential(metadata);

            OperationStatus operationStatus = OperationStatus.newBuilder().setStatus(status.getState()).build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at enableAgents " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void registerAndEnableAgent(RegisterUserRequest request, StreamObserver<org.apache.custos.agent.management.service.AgentRegistrationResponse> responseObserver) {
        try {

            LOGGER.debug("Request received to registerAndEnableAgent for tenant " + request.getTenantId() +
                    " for agent " + request.getUser().getId());

            UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(request.getUser().getId()).build();

            UserSearchRequest userSearchRequest = UserSearchRequest.
                    newBuilder().setUser(metadata).
                    setTenantId(request.getTenantId())
                    .setAccessToken(request.getAccessToken()).build();

            OperationStatus status = iamAdminServiceClient.isAgentNameAvailable(userSearchRequest);
            if (status.getStatus()) {


                CredentialMetadata credentialMetadata = CredentialMetadata.newBuilder()
                        .setOwnerId(request.getTenantId())
                        .setId(request.getUser().getId())
                        .build();

                CredentialMetadata agentCredential = credentialStoreServiceClient.createAgentCredential(credentialMetadata);

                UserRepresentation representation = request.getUser();
                representation = representation.toBuilder().setPassword(agentCredential.getInternalSec()).build();


                List<UserAttribute> attributeList = request.toBuilder().getUser().getAttributesList();
                if (attributeList == null || attributeList.isEmpty()) {
                    attributeList = new ArrayList<>();

                }

                UserAttribute attribute = UserAttribute.newBuilder()
                        .setKey(CUSTOS_REALM_AGENT).addValues("true").build();
                attributeList.add(attribute);
                representation = representation.toBuilder().addAllAttributes(attributeList).build();

                request = request.toBuilder().setUser(representation).build();

                RegisterUserResponse response = iamAdminServiceClient.registerAndEnableAgent(request);

                if (response.getIsRegistered()) {

                    Agent agent = Agent.newBuilder()
                            .setId(representation.getId())
                            .setStatus(AgentStatus.ENABLED)
                            .build();

                    if (representation.getRealmRolesList() != null && !representation.getRealmRolesList().isEmpty()) {
                        agent = agent.toBuilder().addAllRoles(representation.getRealmRolesList()).build();

                    }

                    if (representation.getAttributesList() != null && !representation.getAttributesList().isEmpty()) {
                        List<AgentAttribute> agentAttributes = new ArrayList<>();
                        representation.getAttributesList().forEach(atr -> {
                            AgentAttribute agentAttribute = AgentAttribute
                                    .newBuilder()
                                    .setKey(atr.getKey())
                                    .addAllValue(atr.getValuesList())
                                    .build();
                            agentAttributes.add(agentAttribute);

                        });

                        agent = agent.toBuilder().addAllAttributes(agentAttributes).build();
                    }


                    AgentRequest agentRequest = AgentRequest.newBuilder()
                            .setTenantId(request.getTenantId())
                            .setAgent(agent)
                            .build();
                    agentProfileClient.createAgent(agentRequest);

                    org.apache.custos.agent.management.service.AgentRegistrationResponse registrationResponse =
                            org.apache.custos.agent.management.service.AgentRegistrationResponse
                                    .newBuilder()
                                    .setId(request.getUser().getId())
                                    .setSecret(agentCredential.getSecret())
                                    .build();

                    responseObserver.onNext(registrationResponse);
                    responseObserver.onCompleted();

                } else {
                    String msg = "Agent name not registered ";
                    LOGGER.error(msg);
                    responseObserver.onError(Status.INTERNAL.
                            withDescription(msg).asRuntimeException());
                }

            } else {
                String msg = "Agent name not valid ";
                LOGGER.error(msg);
                responseObserver.onError(Status.ALREADY_EXISTS.
                        withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at registerAndEnableAgent " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void configureAgentClient(AgentClientMetadata request, StreamObserver<OperationStatus> responseObserver) {
        try {

            LOGGER.debug("Request received to configure agent client" + request.getTenantId());

            GetTenantRequest tenantRequest = GetTenantRequest
                    .newBuilder()
                    .setTenantId(request.getTenantId()).build();

            GetTenantResponse response = tenantProfileClient.getTenant(tenantRequest);

            Tenant tenant = response.getTenant();

            if (tenant == null || tenant.getTenantId() == 0) {
                String msg = "Tenant not found  for " + request.getTenantId();
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            request = request.toBuilder().setClientName(AGENT_CLIENT).build();

            OperationStatus status = iamAdminServiceClient.configureAgentClient(request);

            OperationStatus operationStatus = OperationStatus.newBuilder().setStatus(status.getStatus()).build();
            responseObserver.onNext(operationStatus);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred at configureAgentClient " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void getAgent(AgentSearchRequest request, StreamObserver<org.apache.custos.iam.service.Agent> responseObserver) {
        try {
            LOGGER.debug("Request received to getAgent " + request.getTenantId());

            UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(request.getId()).build();

            UserSearchRequest searchRequest = UserSearchRequest
                    .newBuilder()
                    .setUser(metadata)
                    .setTenantId(request.getTenantId())
                    .setAccessToken(request.getAccessToken())
                    .build();

            org.apache.custos.iam.service.Agent agent = iamAdminServiceClient.getAgent(searchRequest);
            List<UserAttribute> attributeList = agent.getAttributesList();
            List<UserAttribute> userAttributes = new ArrayList<>();
            for (UserAttribute attribute : attributeList) {
                if (!attribute.getKey().trim().equals(CUSTOS_REALM_AGENT)) {
                    userAttributes.add(attribute);
                }
            }
            agent = agent.toBuilder().clearAttributes().addAllAttributes(userAttributes).build();

            responseObserver.onNext(agent);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred at getAgent " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void deleteAgent(AgentSearchRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {

            LOGGER.debug("Request received to deleteAgent " + request.getTenantId());

            UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(request.getId()).build();

            UserSearchRequest searchRequest = UserSearchRequest
                    .newBuilder()
                    .setUser(metadata)
                    .setTenantId(request.getTenantId())
                    .setAccessToken(request.getAccessToken())
                    .build();

            OperationStatus status = iamAdminServiceClient.deleteAgent(searchRequest);

            if (status.getStatus()) {

                Agent agent = Agent.newBuilder().setId(request.getId()).build();

                AgentRequest agentRequest = AgentRequest.newBuilder().setTenantId(request.getTenantId())
                        .setAgent(agent).build();
                agentProfileClient.deleteAgent(agentRequest);
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            } else {
                String msg = "Error occurred at delete Agent at IAM Server ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at getAgent " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void disableAgent(AgentSearchRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {

            LOGGER.debug("Request received to disableAgent " + request.getTenantId());

            UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(request.getId()).build();

            UserSearchRequest searchRequest = UserSearchRequest
                    .newBuilder()
                    .setUser(metadata)
                    .setTenantId(request.getTenantId())
                    .setAccessToken(request.getAccessToken())
                    .build();

            OperationStatus status = iamAdminServiceClient.disableAgent(searchRequest);

            if (status.getStatus()) {

                org.apache.custos.iam.service.Agent agent = iamAdminServiceClient.getAgent(searchRequest);


                Agent agentProfile = Agent.newBuilder()
                        .setId(agent.getId())
                        .setStatus(AgentStatus.DISABLED)
                        .addAllRoles(agent.getRealmRolesList())
                        .build();

                for (UserAttribute atr : agent.getAttributesList()) {

                    AgentAttribute agentAttribute = AgentAttribute.newBuilder()
                            .setKey(atr.getKey())
                            .addAllValue(atr.getValuesList())
                            .build();

                    agentProfile = agentProfile.toBuilder().addAttributes(agentAttribute).build();

                }

                AgentRequest agentRequest = AgentRequest.newBuilder().setTenantId(request.getTenantId())
                        .setAgent(agentProfile).build();
                agentProfileClient.updateAgent(agentRequest);
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            } else {
                String msg = "Error occurred at disable Agent at IAM Server ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at disableAgent " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void addAgentAttributes(AddUserAttributesRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {

            LOGGER.debug("Request received to addAgentAttributes " + request.getTenantId());

            OperationStatus status = iamAdminServiceClient.addAgentAttributes(request);

            if (status.getStatus()) {

                for (String agent : request.getAgentsList()) {

                    UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(agent).build();

                    UserSearchRequest searchRequest = UserSearchRequest
                            .newBuilder()
                            .setUser(metadata)
                            .setTenantId(request.getTenantId())
                            .setAccessToken(request.getAccessToken())
                            .build();

                    org.apache.custos.iam.service.Agent iamAgent = iamAdminServiceClient.getAgent(searchRequest);

                    Agent agentProfile = Agent.newBuilder()
                            .setId(iamAgent.getId())
                            .setStatus(iamAgent.getIsEnabled() ? AgentStatus.ENABLED : AgentStatus.DISABLED)
                            .addAllRoles(iamAgent.getRealmRolesList())
                            .build();

                    for (UserAttribute atr : iamAgent.getAttributesList()) {

                        AgentAttribute agentAttribute = AgentAttribute.newBuilder()
                                .setKey(atr.getKey())
                                .addAllValue(atr.getValuesList())
                                .build();

                        agentProfile = agentProfile.toBuilder().addAttributes(agentAttribute).build();

                    }

                    AgentRequest agentRequest = AgentRequest.newBuilder().setTenantId(request.getTenantId())
                            .setAgent(agentProfile).build();
                    agentProfileClient.updateAgent(agentRequest);
                }
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            } else {
                String msg = "Error occurred at addAgentAttributes at IAM Server ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at addAgentAttributes " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }

    }


    @Override
    public void deleteAgentAttributes(DeleteUserAttributeRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteAgentAttributes " + request.getTenantId());

            OperationStatus status = iamAdminServiceClient.deleteAgentAttributes(request);

            if (status.getStatus()) {

                for (String agent : request.getAgentsList()) {

                    UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(agent).build();

                    UserSearchRequest searchRequest = UserSearchRequest
                            .newBuilder()
                            .setUser(metadata)
                            .setTenantId(request.getTenantId())
                            .setAccessToken(request.getAccessToken())
                            .build();

                    org.apache.custos.iam.service.Agent iamAgent = iamAdminServiceClient.getAgent(searchRequest);

                    Agent agentProfile = Agent.newBuilder()
                            .setId(iamAgent.getId())
                            .setStatus(iamAgent.getIsEnabled() ? AgentStatus.ENABLED : AgentStatus.DISABLED)
                            .addAllRoles(iamAgent.getRealmRolesList())
                            .build();

                    for (UserAttribute atr : iamAgent.getAttributesList()) {

                        AgentAttribute agentAttribute = AgentAttribute.newBuilder()
                                .setKey(atr.getKey())
                                .addAllValue(atr.getValuesList())
                                .build();

                        agentProfile = agentProfile.toBuilder().addAttributes(agentAttribute).build();

                    }

                    AgentRequest agentRequest = AgentRequest.newBuilder().setTenantId(request.getTenantId())
                            .setAgent(agentProfile).build();
                    agentProfileClient.updateAgent(agentRequest);
                }
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            } else {
                String msg = "Error occurred at deleteAgentAttributes at IAM Server ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at deleteAgentAttributes " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }


    @Override
    public void addRolesToAgent(AddUserRolesRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteAgentAttributes " + request.getTenantId());

            OperationStatus status = iamAdminServiceClient.addRolesToAgent(request);

            if (status.getStatus()) {

                for (String agent : request.getAgentsList()) {

                    UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(agent).build();

                    UserSearchRequest searchRequest = UserSearchRequest
                            .newBuilder()
                            .setUser(metadata)
                            .setTenantId(request.getTenantId())
                            .setAccessToken(request.getAccessToken())
                            .build();

                    org.apache.custos.iam.service.Agent iamAgent = iamAdminServiceClient.getAgent(searchRequest);

                    Agent agentProfile = Agent.newBuilder()
                            .setId(iamAgent.getId())
                            .setStatus(iamAgent.getIsEnabled() ? AgentStatus.ENABLED : AgentStatus.DISABLED)
                            .addAllRoles(iamAgent.getRealmRolesList())
                            .build();

                    for (UserAttribute atr : iamAgent.getAttributesList()) {

                        AgentAttribute agentAttribute = AgentAttribute.newBuilder()
                                .setKey(atr.getKey())
                                .addAllValue(atr.getValuesList())
                                .build();

                        agentProfile = agentProfile.toBuilder().addAttributes(agentAttribute).build();

                    }

                    AgentRequest agentRequest = AgentRequest.newBuilder().setTenantId(request.getTenantId())
                            .setAgent(agentProfile).build();
                    agentProfileClient.updateAgent(agentRequest);
                }
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            } else {
                String msg = "Error occurred at addRolesToAgent at IAM Server ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at addRolesToAgent " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void deleteRolesFromAgent(DeleteUserRolesRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteAgentAttributes " + request.getTenantId());

            OperationStatus status = iamAdminServiceClient.deleteAgentRoles(request);

            if (status.getStatus()) {


                UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(request.getId()).build();

                UserSearchRequest searchRequest = UserSearchRequest
                        .newBuilder()
                        .setUser(metadata)
                        .setTenantId(request.getTenantId())
                        .setAccessToken(request.getAccessToken())
                        .build();

                org.apache.custos.iam.service.Agent iamAgent = iamAdminServiceClient.getAgent(searchRequest);

                Agent agentProfile = Agent.newBuilder()
                        .setId(iamAgent.getId())
                        .setStatus(iamAgent.getIsEnabled() ? AgentStatus.ENABLED : AgentStatus.DISABLED)
                        .addAllRoles(iamAgent.getRealmRolesList())
                        .build();

                for (UserAttribute atr : iamAgent.getAttributesList()) {

                    AgentAttribute agentAttribute = AgentAttribute.newBuilder()
                            .setKey(atr.getKey())
                            .addAllValue(atr.getValuesList())
                            .build();

                    agentProfile = agentProfile.toBuilder().addAttributes(agentAttribute).build();

                }

                AgentRequest agentRequest = AgentRequest.newBuilder().setTenantId(request.getTenantId())
                        .setAgent(agentProfile).build();
                agentProfileClient.updateAgent(agentRequest);
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            } else {
                String msg = "Error occurred at deleteRolesFromAgent at IAM Server ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at deleteRolesFromAgent " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void enableAgent(AgentSearchRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to enableAgent " + request.getTenantId());

            UserSearchMetadata metadata = UserSearchMetadata.newBuilder().setId(request.getId()).build();

            UserSearchRequest searchRequest = UserSearchRequest
                    .newBuilder()
                    .setUser(metadata)
                    .setTenantId(request.getTenantId())
                    .setAccessToken(request.getAccessToken())
                    .build();

            OperationStatus status = iamAdminServiceClient.enableAgent(searchRequest);

            if (status.getStatus()) {

                org.apache.custos.iam.service.Agent agent = iamAdminServiceClient.getAgent(searchRequest);


                Agent agentProfile = Agent.newBuilder()
                        .setId(agent.getId())
                        .setStatus(AgentStatus.ENABLED)
                        .addAllRoles(agent.getRealmRolesList())
                        .build();

                for (UserAttribute atr : agent.getAttributesList()) {

                    AgentAttribute agentAttribute = AgentAttribute.newBuilder()
                            .setKey(atr.getKey())
                            .addAllValue(atr.getValuesList())
                            .build();

                    agentProfile = agentProfile.toBuilder().addAttributes(agentAttribute).build();

                }

                AgentRequest agentRequest = AgentRequest.newBuilder().setTenantId(request.getTenantId())
                        .setAgent(agentProfile).build();
                agentProfileClient.updateAgent(agentRequest);
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            } else {
                String msg = "Error occurred at disable Agent at IAM Server ";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred at disableAgent " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("UNAUTHENTICATED")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }
}
