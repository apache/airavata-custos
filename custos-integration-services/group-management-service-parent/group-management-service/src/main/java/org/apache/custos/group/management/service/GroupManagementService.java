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

package org.apache.custos.group.management.service;


import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.admin.client.IamAdminServiceClient;
import org.apache.custos.iam.service.*;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.user.profile.client.UserProfileClient;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import  org.apache.custos.group.management.service.GroupManagementServiceGrpc;

@GRpcService
public class GroupManagementService extends GroupManagementServiceGrpc.GroupManagementServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(GroupManagementService.class);

    @Autowired
    private UserProfileClient userProfileClient;

    @Autowired
    private IamAdminServiceClient iamAdminServiceClient;

    @Autowired
    private IdentityClient identityClient;


    @Override
    public void createGroups(GroupsRequest request, StreamObserver<GroupsResponse> responseObserver) {
        try {
            LOGGER.info("Request received create Groups");


        } catch (Exception ex) {
            String msg = "Error occurred at createGroups " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("CredentialGenerationException")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void updateGroup(GroupRequest request, StreamObserver<GroupRepresentation> responseObserver) {
        try {
            LOGGER.info("Request received update");
            LOGGER.info(request.getGroup().getId());
        } catch (Exception ex) {
            String msg = "Error occurred at updateGroup " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("CredentialGenerationException")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void deleteGroup(GroupRequest request, StreamObserver<OperationStatus> responseObserver) {
        try {
            LOGGER.info("Request received delete");
            LOGGER.info(request.getGroup().getId());

        } catch (Exception ex) {
            String msg = "Error occurred at deleteGroup " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("CredentialGenerationException")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void findGroup(GroupRequest request, StreamObserver<GroupRepresentation> responseObserver) {
        try {
            LOGGER.info("Request received findGroup");

        } catch (Exception ex) {
            String msg = "Error occurred at findGroup " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("CredentialGenerationException")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }

    @Override
    public void getAllGroups(GroupRequest request, StreamObserver<GroupsResponse> responseObserver) {
        try {
            LOGGER.info("Request received getAllGroups");

        } catch (Exception ex) {
            String msg = "Error occurred at getAllGroups " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("CredentialGenerationException")) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
            } else {
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }
        }
    }
}
