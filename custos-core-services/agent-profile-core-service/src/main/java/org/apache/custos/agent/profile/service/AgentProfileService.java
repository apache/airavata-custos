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

package org.apache.custos.agent.profile.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.agent.profile.mapper.AgentMapper;
import org.apache.custos.agent.profile.persistance.repository.AgentAttributeRepository;
import org.apache.custos.agent.profile.persistance.repository.AgentRepository;
import org.apache.custos.agent.profile.persistance.repository.AgentRoleRepository;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.Optional;

@GRpcService
public class AgentProfileService extends AgentProfileServiceGrpc.AgentProfileServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(AgentProfileService.class);


    @Autowired
    private AgentRepository agentRepository;

    @Autowired
    private AgentRoleRepository agentRoleRepository;

    @Autowired
    private AgentAttributeRepository agentAttributeRepository;


    @Override
    public void createAgent(AgentRequest request, StreamObserver<Agent> responseObserver) {
        try {
            LOGGER.debug("Request received to createAgent for " + request.getAgent().getId() + "at " + request.getTenantId());

            String agentId = request.getAgent().getId() + "@" + request.getTenantId();

            Optional<org.apache.custos.agent.profile.persistance.model.Agent> op = agentRepository.findById(agentId);

            if (op.isEmpty()) {

                org.apache.custos.agent.profile.persistance.model.Agent entity =
                        AgentMapper.createAgent(request.getAgent(), request.getTenantId());
                entity.setId(agentId);
                agentRepository.save(entity);

                Optional<org.apache.custos.agent.profile.persistance.model.Agent> optionalAgent =
                        agentRepository.findById(agentId);

                if (optionalAgent.isPresent()) {

                    Agent ex = AgentMapper.createAgent(optionalAgent.get());

                    responseObserver.onNext(ex);
                    responseObserver.onCompleted();

                } else {
                    String msg = "Error occurred while creating agent, agent is not saved";
                    LOGGER.error(msg);
                    responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

                }

            } else {
                String msg = "Error occurred while creating agent already exists";
                LOGGER.error(msg);
                responseObserver.onError(Status.ALREADY_EXISTS.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while creating agent for " + request.getAgent().getId() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void updateAgent(AgentRequest request, StreamObserver<Agent> responseObserver) {
        try {
            LOGGER.debug("Request received to updateAgent for " + request.getAgent().getId() + "at " + request.getTenantId());

            String userId = request.getAgent().getId() + "@" + request.getTenantId();

            Optional<org.apache.custos.agent.profile.persistance.model.Agent>
                    exEntity = agentRepository.findById(userId);

            Optional<org.apache.custos.agent.profile.persistance.model.Agent>
                    preEntity = agentRepository.findById(request.getAgent().getId());
            if (exEntity.isEmpty()) {
                exEntity = preEntity;
                userId = request.getAgent().getId();
            }

            if (exEntity.isPresent()) {

                org.apache.custos.agent.profile.persistance.model.Agent entity =
                        AgentMapper.createAgent(request.getAgent(), request.getTenantId());
                entity.setId(exEntity.get().getId());
                entity.setCreatedAt(exEntity.get().getCreatedAt());

                org.apache.custos.agent.profile.persistance.model.Agent exProfile = exEntity.get();

                if (exProfile.getAgentAttribute() != null) {
                    exProfile.getAgentAttribute().forEach(atr -> {
                        agentAttributeRepository.delete(atr);

                    });
                }

                if (exProfile.getAgentRole() != null) {
                    exProfile.getAgentRole().forEach(role -> {
                        agentRoleRepository.delete(role);
                    });
                }

                agentRepository.save(entity);

                Optional<org.apache.custos.agent.profile.persistance.model.Agent> optionalAgent =
                        agentRepository.findById(userId);

                if (optionalAgent.isPresent()) {

                    Agent ex = AgentMapper.createAgent(optionalAgent.get());

                    responseObserver.onNext(ex);
                    responseObserver.onCompleted();

                } else {
                    String msg = "Error occurred while updating agent, agent is not saved";
                    LOGGER.error(msg);
                    responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());

                }
            } else {
                String msg = "Cannot find a agent for " + userId;
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while updating agent for " + request.getAgent().getId() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteAgent(AgentRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteAgent for " + request.getAgent().getId() + "at " + request.getTenantId());

            String id = request.getAgent().getId() + "@" + request.getTenantId();

            Optional<org.apache.custos.agent.profile.persistance.model.Agent> agentOptional = agentRepository.findById(id);
            Optional<org.apache.custos.agent.profile.persistance.model.Agent>
                    preEntity = agentRepository.findById(request.getAgent().getId());
            if (agentOptional.isEmpty()) {
                agentOptional = preEntity;
            }


            if (agentOptional.isPresent()) {

                agentRepository.delete(agentOptional.get());
                OperationStatus status = OperationStatus.newBuilder().setStatus(true).build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND.withDescription("Agent not found")
                        .asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while deleting agent for " + request.getAgent().getId() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAgent(AgentRequest request, StreamObserver<Agent> responseObserver) {
        try {
            LOGGER.debug("Request received to getAgent for " + request.getAgent().getId() + "at " + request.getTenantId());

            String id = request.getAgent().getId() + "@" + request.getTenantId();

            Optional<org.apache.custos.agent.profile.persistance.model.Agent> agentOptional = agentRepository.findById(id);

            Optional<org.apache.custos.agent.profile.persistance.model.Agent>
                    preEntity = agentRepository.findById(request.getAgent().getId());
            if (agentOptional.isEmpty()) {
                agentOptional = preEntity;
            }

            if (agentOptional.isPresent()) {


                Agent agent = AgentMapper.createAgent(agentOptional.get());

                responseObserver.onNext(agent);
                responseObserver.onCompleted();
            } else {

                responseObserver.onNext(null);
                responseObserver.onCompleted();
            }


        } catch (Exception ex) {
            String msg = "Error occurred while fetching agent for " + request.getAgent().getId() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


}
