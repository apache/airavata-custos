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
import org.apache.custos.user.profile.mapper.GroupMapper;
import org.apache.custos.user.profile.mapper.StatusUpdateMetadataMapper;
import org.apache.custos.user.profile.mapper.UserProfileMapper;
import org.apache.custos.user.profile.persistance.model.AttributeUpdateMetadata;
import org.apache.custos.user.profile.persistance.model.StatusUpdateMetadata;
import org.apache.custos.user.profile.persistance.model.UserProfile;
import org.apache.custos.user.profile.persistance.repository.*;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.Set;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.stream.Collectors;

@GRpcService
public class UserProfileService extends UserProfileServiceGrpc.UserProfileServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(UserProfileService.class);

    @Autowired
    private UserRepository repository;

    @Autowired
    private StatusUpdateMetadataRepository statusUpdaterRepository;

    @Autowired
    private AttributeUpdateMetadataRepository attributeUpdateMetadataRepository;

    @Autowired
    private UserAttributeRepository userAttributeRepository;

    @Autowired
    private RoleRepository roleRepository;

    @Autowired
    private GroupRepository groupRepository;

    @Autowired
    private GroupRoleRepository groupRoleRepository;

    @Autowired
    private GroupAttributeRepository groupAttributeRepository;


    @Override
    public void createUserProfile(UserProfileRequest request, StreamObserver<org.apache.custos.user.profile.service.UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to createUserProfile for " + request.getProfile().getUsername() + "at " + request.getTenantId());

            String userId = request.getProfile().getUsername() + "@" + request.getTenantId();

            Optional<UserProfile> op = repository.findById(userId);

            if (op.isEmpty()) {

                UserProfile entity = UserProfileMapper.createUserProfileEntityFromUserProfile(request.getProfile());
                entity.setId(userId);
                entity.setTenantId(request.getTenantId());
                repository.save(entity);
            }

            responseObserver.onNext(request.getProfile());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while creating user profile for " + request.getProfile().getUsername() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void updateUserProfile(UserProfileRequest request, StreamObserver<org.apache.custos.user.profile.service.UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to updateUserProfile for " + request.getProfile().getUsername() + "at " + request.getTenantId());

            String userId = request.getProfile().getUsername() + "@" + request.getTenantId();

            Optional<UserProfile> exEntity = repository.findById(userId);


            if (exEntity.isPresent()) {


                UserProfile entity = UserProfileMapper.createUserProfileEntityFromUserProfile(request.getProfile());


                Set<AttributeUpdateMetadata> metadata = AttributeUpdateMetadataMapper.
                        createAttributeUpdateMetadataEntity(exEntity.get(), entity, request.getPerformedBy());


                entity.setAttributeUpdateMetadata(metadata);
                entity.setId(userId);
                entity.setTenantId(request.getTenantId());
                entity.setCreatedAt(exEntity.get().getCreatedAt());

                UserProfile exProfile = exEntity.get();

                if (exProfile.getUserAttribute() != null) {
                    exProfile.getUserAttribute().forEach(atr -> {
                        userAttributeRepository.delete(atr);

                    });
                }

                if (exProfile.getUserRole() != null) {
                    exProfile.getUserRole().forEach(role -> {
                        roleRepository.delete(role);
                    });
                }

                repository.save(entity);

                responseObserver.onNext(request.getProfile());
                responseObserver.onCompleted();
            } else {
                String msg = "Cannot find a user profile for " + userId;
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while updating user profile for " + request.getProfile().getUsername() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUserProfile(UserProfileRequest request, StreamObserver<org.apache.custos.user.profile.service.UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to getUserProfile for " + request.getProfile().getUsername() + "at " + request.getTenantId());

            String userId = request.getProfile().getUsername() + "@" + request.getTenantId();

            Optional<UserProfile> entity = repository.findById(userId);

            if (entity.isPresent()) {
                UserProfile profileEntity = entity.get();
                org.apache.custos.user.profile.service.UserProfile profile = UserProfileMapper.createUserProfileFromUserProfileEntity(profileEntity);

                responseObserver.onNext(profile);
                responseObserver.onCompleted();

            } else {

                responseObserver.onNext(null);
                responseObserver.onCompleted();
            }


        } catch (Exception ex) {
            String msg = "Error occurred while fetching user profile for " + request.getProfile().getUsername() + "at " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void deleteUserProfile(UserProfileRequest request, StreamObserver<org.apache.custos.user.profile.service.UserProfile> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteUserProfile for " + request.getProfile().getUsername() + "at " + request.getTenantId());
            long tenantId = request.getTenantId();

            String username = request.getProfile().getUsername();

            String userId = username + "@" + tenantId;

            Optional<UserProfile> profileEntity = repository.findById(userId);

            if (profileEntity.isPresent()) {
                UserProfile entity = profileEntity.get();

                org.apache.custos.user.profile.service.UserProfile prof = UserProfileMapper.createUserProfileFromUserProfileEntity(entity);

                repository.delete(profileEntity.get());
                responseObserver.onNext(prof);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND.withDescription("User profile not found")
                        .asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while deleting user profile for " + request.getProfile().getUsername() + "at " + request.getTenantId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllUserProfilesInTenant(UserProfileRequest request,
                                           StreamObserver<GetAllUserProfilesResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllUserProfilesInTenant for " + request.getTenantId());
            long tenantId = request.getTenantId();

            List<UserProfile> profileList = repository.findByTenantId(tenantId);

            List<org.apache.custos.user.profile.service.UserProfile> userProfileList = new ArrayList<>();

            if (profileList != null && profileList.size() > 0) {
                for (UserProfile entity : profileList) {
                    org.apache.custos.user.profile.service.UserProfile prof = UserProfileMapper.createUserProfileFromUserProfileEntity(entity);
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
    public void findUserProfilesByAttributes(UserProfileRequest request, StreamObserver<GetAllUserProfilesResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to findUserProfilesByAttributes at " + request.getTenantId());


            List<UserAttribute> attributeList = request.getProfile().getAttributesList();

            List<UserProfile> selectedProfiles = new ArrayList<>();
            List<org.apache.custos.user.profile.service.UserProfile> userProfileList = new ArrayList<>();
            attributeList.forEach(atr -> {

                List<String> values = atr.getValueList();
                values.forEach(val -> {
                    List<UserProfile>
                            userAttributes = userAttributeRepository.findFilteredUserProfiles(atr.getKey(), val);
                    if (userAttributes == null || userAttributes.isEmpty()) {
                        GetAllUserProfilesResponse response = GetAllUserProfilesResponse
                                .newBuilder()
                                .build();
                        responseObserver.onNext(response);
                        responseObserver.onCompleted();
                        return;
                    }
                    if (selectedProfiles.isEmpty()) {
                        selectedProfiles.addAll(userAttributes);
                    } else {
                        List<UserProfile> profiles = userAttributes.stream().filter(newProf -> {
                            AtomicBoolean matched = new AtomicBoolean(false);
                            selectedProfiles.forEach(selectedProfile -> {
                                if (selectedProfile.getId().equals(newProf.getId())) {
                                    matched.set(true);
                                }
                            });
                            return matched.get();
                        }).collect(Collectors.toList());
                        selectedProfiles.clear();
                        selectedProfiles.addAll(profiles);
                    }
                });
            });

            if (!selectedProfiles.isEmpty()) {

                selectedProfiles.forEach(userProfile -> {
                    org.apache.custos.user.profile.service.UserProfile prof =
                            UserProfileMapper.createUserProfileFromUserProfileEntity(userProfile);
                    userProfileList.add(prof);

                });
                GetAllUserProfilesResponse response = GetAllUserProfilesResponse
                        .newBuilder()
                        .addAllProfiles(userProfileList)
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();


            } else {
                GetAllUserProfilesResponse response = GetAllUserProfilesResponse
                        .newBuilder()
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching user profile for " + request.getProfile().getUsername() + "at " + request.getTenantId();
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

            List<StatusUpdateMetadata> statusUpdateMetadata = statusUpdaterRepository.findAllByUserProfileId(userId);

            List<AttributeUpdateMetadata> attributeUpdateMetadata = attributeUpdateMetadataRepository.findAllByUserProfileId(userId);

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


    @Override
    public void createGroup(GroupRequest request, StreamObserver<Group> responseObserver) {
        try {
            LOGGER.debug("Request received to createGroup from tenant" + request.getTenantId());

            String groupId = request.getGroup().getId();

            Optional<org.apache.custos.user.profile.persistance.model.Group> op = groupRepository.findById(groupId);

            if (op.isEmpty()) {

                org.apache.custos.user.profile.persistance.model.Group entity =
                        GroupMapper.createGroupEntity(request.getGroup(), request.getTenantId());
                groupRepository.save(entity);
            }

            Optional<org.apache.custos.user.profile.persistance.model.Group> exOP =
                    groupRepository.findById(request.getGroup().getId());

            if (exOP.isPresent()) {
                Group exGroup = GroupMapper.createGroup(exOP.get());
                responseObserver.onNext(exGroup);
                responseObserver.onCompleted();
            } else {

                String msg = "Error occurred while creating Group for " + request.getTenantId()
                        + " reason : DB error";
                LOGGER.error(msg);

                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }


        } catch (Exception ex) {
            String msg = "Error occurred while creating Group for " + request.getTenantId() +
                    " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void updateGroup(GroupRequest request, StreamObserver<Group> responseObserver) {
        try {
            LOGGER.debug("Request received to updateGroup for group with id  " + request.getGroup().getId() +
                    "at tenant " + request.getTenantId());

            String groupId = request.getGroup().getId();

            Optional<org.apache.custos.user.profile.persistance.model.Group> exEntity =
                    groupRepository.findById(groupId);


            if (exEntity.isPresent()) {


                org.apache.custos.user.profile.persistance.model.Group entity = GroupMapper.
                        createGroupEntity(request.getGroup(), request.getTenantId());


                entity.setCreatedAt(exEntity.get().getCreatedAt());

                org.apache.custos.user.profile.persistance.model.Group exGroup = exEntity.get();

                if (exGroup.getGroupAttribute() != null) {
                    exGroup.getGroupAttribute().forEach(atr -> {
                        groupAttributeRepository.delete(atr);

                    });
                }

                if (exGroup.getGroupRole() != null) {
                    exGroup.getGroupRole().forEach(role -> {
                        groupRoleRepository.delete(role);
                    });
                }

                groupRepository.save(entity);

                Optional<org.apache.custos.user.profile.persistance.model.Group> exOP =
                        groupRepository.findById(request.getGroup().getId());

                if (exOP.isPresent()) {
                    Group exNewGroup = GroupMapper.createGroup(exOP.get());
                    responseObserver.onNext(exNewGroup);
                    responseObserver.onCompleted();
                } else {

                    String msg = "Error occurred while creating Group for " + request.getTenantId()
                            + " reason : DB error";
                    LOGGER.error(msg);

                    responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                }

            } else {
                String msg = "Cannot find a group for " + groupId;
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while creating user profile for " + request.getGroup().getId() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void deleteGroup(GroupRequest request, StreamObserver<Group> responseObserver) {
        try {
            LOGGER.debug("Request received to deleteGroup for " + request.getGroup().getId() + "at " + request.getTenantId());

            String userId = request.getGroup().getId();

            Optional<org.apache.custos.user.profile.persistance.model.Group> op = groupRepository.findById(userId);

            if (op.isPresent()) {
                org.apache.custos.user.profile.persistance.model.Group entity = op.get();

                Group prof = GroupMapper.createGroup(entity);

                groupRepository.delete(op.get());

                List<org.apache.custos.user.profile.persistance.model.Group> groupList = groupRepository.findByParentId(entity.getId());

                if (groupList != null && !groupList.isEmpty()) {
                    for (org.apache.custos.user.profile.persistance.model.Group group : groupList) {
                        groupRepository.delete(group);
                    }

                }

                responseObserver.onNext(prof);
                responseObserver.onCompleted();
            } else {
                responseObserver.onError(Status.NOT_FOUND.withDescription("Group not found")
                        .asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while creating group for " + request.getGroup() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getGroup(GroupRequest request, StreamObserver<Group> responseObserver) {
        try {
            LOGGER.debug("Request received to getGroup for group " + request.getGroup().getId() + "at " + request.getTenantId());

            String userId = request.getGroup().getId();

            Optional<org.apache.custos.user.profile.persistance.model.Group> op = groupRepository.findById(userId);

            if (op.isPresent()) {

                Group entity = GroupMapper.createGroup(op.get());
                responseObserver.onNext(entity);
                responseObserver.onCompleted();

            } else {
                responseObserver.onNext(null);
                responseObserver.onCompleted();

            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching group " + request.getGroup().getId() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getAllGroups(GroupRequest request, StreamObserver<GetAllGroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllGroups for " + request.getTenantId());

            List<org.apache.custos.user.profile.persistance.model.Group> groups = groupRepository.findAll();

            List<Group> groupList = new ArrayList<>();

            if (groups != null && groups.isEmpty()) {

                for (org.apache.custos.user.profile.persistance.model.Group group : groups) {

                    groupList.add(GroupMapper.createGroup(group));
                }

            }
            GetAllGroupsResponse response = GetAllGroupsResponse.newBuilder().addAllGroups(groupList).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching groups for " + request.getTenantId() + "at "
                    + request.getTenantId() + " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }


}
