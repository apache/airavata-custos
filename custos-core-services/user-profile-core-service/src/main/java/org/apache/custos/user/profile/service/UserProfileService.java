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

package org.apache.custos.user.profile.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.user.profile.mapper.AttributeUpdateMetadataMapper;
import org.apache.custos.user.profile.mapper.StatusUpdateMetadataMapper;
import org.apache.custos.user.profile.mapper.UserProfileMapper;
import org.apache.custos.user.profile.persistance.model.AttributeUpdateMetadata;
import org.apache.custos.user.profile.persistance.model.StatusUpdateMetadata;
import org.apache.custos.user.profile.persistance.model.UserProfileEntity;
import org.apache.custos.user.profile.persistance.repository.AttributeUpdateMetadataRepository;
import org.apache.custos.user.profile.persistance.repository.StatusUpdateMetadataRepository;
import org.apache.custos.user.profile.persistance.repository.UserRepository;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.Set;

@GRpcService
public class UserProfileService extends UserProfileServiceGrpc.UserProfileServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(UserProfileService.class);

    @Autowired
    private UserRepository repository;

    @Autowired
    private StatusUpdateMetadataRepository statusUpdaterRepository;

    @Autowired
    private AttributeUpdateMetadataRepository attributeUpdateMetadataRepository;


    @Override
    public void createUserProfile(UserProfile request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to createUserProfile for " + request.getUsername() + "at " + request.getTenantId());

            String userId = request.getUsername() + "@" + request.getTenantId();

            UserProfile profile = request.toBuilder().setUserId(userId).build();

            Optional<UserProfileEntity> op = repository.findById(userId);

            if (op.isEmpty()) {

                UserProfileEntity entity = UserProfileMapper.createUserProfileEntityFromUserProfile(profile);

                repository.save(entity);
            }

            responseObserver.onNext(profile);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while creating user profile for " + request.getUsername() + "at " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void updateUserProfile(UserProfile request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to updateUserProfile for " + request.getUsername() + "at " + request.getTenantId());

          Optional<UserProfileEntity> exEntity  =   repository.findById(request.getUserId());


          if (exEntity.isPresent()) {


              UserProfileEntity entity = UserProfileMapper.createUserProfileEntityFromUserProfile(request);


              Set<AttributeUpdateMetadata> metadata = AttributeUpdateMetadataMapper.
                      createAttributeUpdateMetadataEntity(exEntity.get(), entity, request.getUpdatedBy());

              entity.setAttributeUpdateMetadata(metadata);

              repository.save(entity);

              responseObserver.onNext(request);
              responseObserver.onCompleted();
          } else  {


          }

        } catch (Exception ex) {
            String msg = "Error occurred while updating user profile for " + request.getUsername() + "at " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUserProfile(GetUserProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to getUserProfile for " + request.getUsername() + "at " + request.getTenantId());

            String userId = request.getUsername() + "@" + request.getTenantId();

            Optional<UserProfileEntity> entity = repository.findById(userId);

            if (entity.isPresent()) {
                UserProfileEntity profileEntity = entity.get();

                UserProfile profile = UserProfileMapper.createUserProfileFromUserProfileEntity(profileEntity);

                responseObserver.onNext(profile);
                responseObserver.onCompleted();

            } else {

                responseObserver.onError(Status.NOT_FOUND.withDescription("User not found").asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred while fetching user profile for " + request.getUsername() + "at " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteUserProfile(DeleteUserProfileRequest request, StreamObserver<UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteUserProfile for " + request.getUsername() + "at " + request.getTenantId());
            long tenantId = request.getTenantId();

            String username = request.getUsername();

            String userId = username + "@" + tenantId;

            Optional<UserProfileEntity> profileEntity = repository.findById(userId);

            if (profileEntity.isPresent()) {
                UserProfileEntity entity = profileEntity.get();

                UserProfile prof = UserProfileMapper.createUserProfileFromUserProfileEntity(entity);

                repository.delete(profileEntity.get());
                responseObserver.onNext(prof);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND.withDescription("User profile not found")
                        .asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while deleting user profile for " + request.getUsername() + "at " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllUserProfilesInTenant(GetAllUserProfilesRequest request,
                                           StreamObserver<GetAllUserProfilesResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllUserProfilesInTenant for " + request.getTenantId());
            long tenantId = request.getTenantId();

            List<UserProfileEntity> profileList = repository.findByTenantId(tenantId);

            List<UserProfile> userProfileList = new ArrayList<>();

            if (profileList != null && profileList.size() > 0) {
                for (UserProfileEntity entity : profileList) {
                    UserProfile prof = UserProfileMapper.createUserProfileFromUserProfileEntity(entity);
                    userProfileList.add(prof);
                }
            }

            GetAllUserProfilesResponse response = GetAllUserProfilesResponse
                    .newBuilder()
                    .addAllProfiles(userProfileList)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while fetching  user profile for tenant " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getUserProfileAuditTrails(GetUpdateAuditTrailRequest request,
                                          StreamObserver<GetUpdateAuditTrailResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getUserProfileAuditTrails for " + request.getUsername() + "at " + request.getTenantId());

            String username = request.getUsername();

            long tenantId = request.getTenantId();

            String userId = username + "@" + tenantId;

            List<StatusUpdateMetadata> statusUpdateMetadata = statusUpdaterRepository.findAllByUserProfileEntityId(userId);

            List<AttributeUpdateMetadata> attributeUpdateMetadata = attributeUpdateMetadataRepository.findAllByUserProfileEntityId(userId);

            List<UserProfileStatusUpdateMetadata> userProfileStatusUpdateMetadata = new ArrayList<>();
            List<UserProfileAttributeUpdateMetadata> userProfileAttributeUpdateMetadata = new ArrayList<>();

            if (statusUpdateMetadata != null && statusUpdateMetadata.size() > 0) {

                for (StatusUpdateMetadata metadata : statusUpdateMetadata) {

                    UserProfileStatusUpdateMetadata met = StatusUpdateMetadataMapper.createUserProfileStatusMetadataFrom(metadata);
                    userProfileStatusUpdateMetadata.add(met);
                }

            }


            if (attributeUpdateMetadata != null && attributeUpdateMetadata.size() > 0) {

                for (AttributeUpdateMetadata metadata : attributeUpdateMetadata) {

                    UserProfileAttributeUpdateMetadata met = AttributeUpdateMetadataMapper.createAttributeUpdateMetadataFromEntity(metadata);
                    userProfileAttributeUpdateMetadata.add(met);
                }

            }

            GetUpdateAuditTrailResponse response = GetUpdateAuditTrailResponse
                    .newBuilder()
                    .addAllAttributeAudit(userProfileAttributeUpdateMetadata)
                    .addAllStatusAudit(userProfileStatusUpdateMetadata)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching  audit trials " + request.getUsername() + "at " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }
}
