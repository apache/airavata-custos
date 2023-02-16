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
import org.apache.custos.user.profile.persistance.model.UserProfile;
import org.apache.custos.user.profile.persistance.model.*;
import org.apache.custos.user.profile.persistance.repository.*;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.*;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;
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

    @Autowired
    private GroupMembershipRepository groupMembershipRepository;

    @Autowired
    private GroupToGroupMembershipRepository groupToGroupMembershipRepository;

    @Autowired
    private GroupMembershipTypeRepository groupMembershipTypeRepository;


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
                org.apache.custos.user.profile.service.UserProfile profile = UserProfileMapper.createUserProfileFromUserProfileEntity(profileEntity, null);

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

                org.apache.custos.user.profile.service.UserProfile prof = UserProfileMapper.createUserProfileFromUserProfileEntity(entity, null);

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
            int limit = request.getLimit();
            int offset = request.getOffset();

            List<UserProfile> profileList = new ArrayList<>();

            if (limit > 0) {
                profileList = repository.findByTenantIdWithPagination(tenantId, limit, offset);
            } else {
                profileList = repository.findByTenantId(tenantId);
            }


            List<org.apache.custos.user.profile.service.UserProfile> userProfileList = new ArrayList<>();

            if (profileList != null && profileList.size() > 0) {
                for (UserProfile entity : profileList) {
                    org.apache.custos.user.profile.service.UserProfile prof = UserProfileMapper
                            .createUserProfileFromUserProfileEntity(entity, null);
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
            LOGGER.error(msg, ex);
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

                List<String> values = atr.getValuesList();
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
                            UserProfileMapper.createUserProfileFromUserProfileEntity(userProfile, null);
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
            long tenantId = request.getTenantId();

            String effectiveId = groupId + "@" + tenantId;

            Optional<org.apache.custos.user.profile.persistance.model.Group> op = groupRepository.findById(effectiveId);

            String ownerId = request.getGroup().getOwnerId() + "@" + tenantId;

            Optional<UserProfile> userProfile = repository.findById(ownerId);

            if (userProfile.isEmpty()) {
                String msg = "Error occurred while creating  Group for " + request.getTenantId()
                        + " reason : Owner  not found";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                return;
            }

            org.apache.custos.user.profile.persistance.model.Group savedGroup = null;
            if (op.isEmpty()) {

                org.apache.custos.user.profile.persistance.model.Group entity =
                        GroupMapper.createGroupEntity(request.getGroup(), request.getTenantId());

                String parentId = entity.getParentId();
                if (parentId != null && !parentId.trim().equals("")) {

                    Optional<org.apache.custos.user.profile.persistance.model.Group> parent = groupRepository.findById(parentId);

                    if (parent.isEmpty()) {
                        String msg = "Error occurred while creating  Group for " + request.getTenantId()
                                + " reason : Parent group not found";
                        LOGGER.error(msg);
                        responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                        return;
                    }

                    entity = GroupMapper.setParentGroupMembership(parent.get(), entity);
                }

                savedGroup = groupRepository.save(entity);
            }

            Optional<org.apache.custos.user.profile.persistance.model.Group> exOP =
                    groupRepository.findById(effectiveId);

            if (exOP.isPresent()) {
                String type = DefaultGroupMembershipTypes.OWNER.name();

                Optional<UserGroupMembershipType> groupMembershipType = groupMembershipTypeRepository.findById(type);
                UserGroupMembershipType exist = null;

                if (groupMembershipType.isEmpty()) {
                    exist = new UserGroupMembershipType();
                    exist.setId(type);
                    groupMembershipTypeRepository.save(exist);
                }

                exist = groupMembershipType.get();


                UserGroupMembership userGroupMembership = new
                        UserGroupMembership();
                userGroupMembership.setGroup(savedGroup);
                userGroupMembership.setUserProfile(userProfile.get());
                userGroupMembership.setTenantId(tenantId);

                userGroupMembership.setUserGroupMembershipType(exist);
                groupMembershipRepository.save(userGroupMembership);

                Group exGroup = GroupMapper.createGroup(exOP.get(), userGroupMembership.getUserProfile().getUsername());
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

            long tenantId = request.getTenantId();

            String effectiveId = groupId + "@" + tenantId;


            Optional<org.apache.custos.user.profile.persistance.model.Group> exEntity =
                    groupRepository.findById(effectiveId);


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
                        groupRepository.findById(effectiveId);


                if (exOP.isPresent()) {
                    List<UserGroupMembership> userGroupMemberships = groupMembershipRepository.findAllByGroupId(effectiveId);

                    String ownerId = null;
                    for (UserGroupMembership userGroupMembership : userGroupMemberships) {
                        if (userGroupMembership.getUserGroupMembershipType().getId().equals(DefaultGroupMembershipTypes.OWNER.name())) {
                            ownerId = userGroupMembership.getUserProfile().getUsername();
                        }
                    }

                    Group exNewGroup = GroupMapper.createGroup(exOP.get(), ownerId);
                    responseObserver.onNext(exNewGroup);
                    responseObserver.onCompleted();
                } else {

                    String msg = "Error occurred while updating group  " + request.getTenantId()
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
            String msg = "Error occurred while updating group " + request.getGroup().getId() + "at "
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
            long tenantId = request.getTenantId();

            String effectiveId = userId + "@" + tenantId;

            Optional<org.apache.custos.user.profile.persistance.model.Group> op = groupRepository.findById(effectiveId);

            if (op.isPresent()) {
                org.apache.custos.user.profile.persistance.model.Group entity = op.get();

                List<UserGroupMembership> userGroupMemberships = groupMembershipRepository.findAllByGroupId(effectiveId);

                String ownerId = null;
                for (UserGroupMembership userGroupMembership : userGroupMemberships) {
                    if (userGroupMembership.getUserGroupMembershipType().getId().equals(DefaultGroupMembershipTypes.OWNER.name())) {
                        ownerId = userGroupMembership.getUserProfile().getUsername();
                    }
                }


                Group prof = GroupMapper.createGroup(entity, ownerId);

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
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getGroup(GroupRequest request, StreamObserver<Group> responseObserver) {
        try {
            LOGGER.debug("Request received to getGroup for group " + request.getGroup().getId() + "at " + request.getTenantId());

            String userId = request.getGroup().getId();

            long tenantId = request.getTenantId();

            String effectiveId = userId + "@" + tenantId;

            Optional<org.apache.custos.user.profile.persistance.model.Group> op = groupRepository.findById(effectiveId);

            if (op.isPresent()) {

                List<UserGroupMembership> userGroupMemberships = groupMembershipRepository.findAllByGroupId(effectiveId);

                String ownerId = null;
                for (UserGroupMembership userGroupMembership : userGroupMemberships) {
                    if (userGroupMembership.getUserGroupMembershipType().getId().equals(DefaultGroupMembershipTypes.OWNER.name())) {
                        ownerId = userGroupMembership.getUserProfile().getUsername();
                    }
                }
                Group entity = GroupMapper.createGroup(op.get(), ownerId);

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

            List<org.apache.custos.user.profile.persistance.model.Group> groups = groupRepository
                    .searchEntities(request.getTenantId(), request.getGroup(), request.getOffset(),request.getLimit());

            if (groups == null || groups.isEmpty()) {
                groups = groupRepository.
                        findAllByTenantId(request.getTenantId());
            }

            List<Group> groupList = new ArrayList<>();

            if (groups != null && !groups.isEmpty()) {
                for (org.apache.custos.user.profile.persistance.model.Group group : groups) {

                    List<UserGroupMembership> userGroupMemberships = groupMembershipRepository.findAllByGroupId(group.getId());

                    String ownerId = null;
                    for (UserGroupMembership userGroupMembership : userGroupMemberships) {
                        if (userGroupMembership.getUserGroupMembershipType().getId().equals(DefaultGroupMembershipTypes.OWNER.name())) {
                            ownerId = userGroupMembership.getUserProfile().getUsername();
                            break;
                        }
                    }

                    Group gr = GroupMapper.createGroup(group, ownerId);

                    groupList.add(gr);
                }
            }

            GetAllGroupsResponse response = GetAllGroupsResponse.newBuilder().addAllGroups(groupList).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching groups of client " + request.getClientId() +
                    " reason :" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void addUserToGroup(GroupMembership request,
                               StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to addUserToGroup for " + request.getTenantId());

            String group_id = request.getGroupId();
            String username = request.getUsername();
            long tenantId = request.getTenantId();
            String userId = username + "@" + tenantId;

            String effectiveGroupId = group_id + "@" + tenantId;

            Optional<org.apache.custos.user.profile.persistance.model.Group> group = groupRepository.findById(effectiveGroupId);
            Optional<UserProfile> userProfile = repository.findById(userId);

            if (group.isPresent() && userProfile.isPresent()) {

                List<UserGroupMembership> memberships =
                        groupMembershipRepository.findAllByGroupIdAndUserProfileId(effectiveGroupId, userId);

                if (memberships == null || memberships.isEmpty()) {


                    String type = request.getType();
                    if (type == null || type.trim().isEmpty()) {
                        type = DefaultGroupMembershipTypes.MEMBER.name();
                    } else {
                        type = type.toUpperCase();
                    }

                    Optional<UserGroupMembershipType> groupMembershipType = groupMembershipTypeRepository.findById(type);
                    UserGroupMembershipType exist = null;

                    if (groupMembershipType.isEmpty()) {
                        exist = new UserGroupMembershipType();
                        exist.setId(type);
                        groupMembershipTypeRepository.save(exist);
                    }

                    exist = groupMembershipType.get();

                    UserGroupMembership userGroupMembership = new
                            UserGroupMembership();
                    userGroupMembership.setGroup(group.get());
                    userGroupMembership.setUserProfile(userProfile.get());
                    userGroupMembership.setTenantId(tenantId);

                    userGroupMembership.setUserGroupMembershipType(exist);
                    groupMembershipRepository.save(userGroupMembership);
                }

                org.apache.custos.user.profile.service.Status status = org.apache.custos.user.profile.service.Status
                        .newBuilder()
                        .setStatus(true).build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            } else {
                String msg = "Group or user not available";
                LOGGER.error(msg);
                responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
            }

        } catch (Exception ex) {
            String msg = "Error occurred while add user to  group at  " + request.getTenantId() +
                    " reason :" + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void removeUserFromGroup(GroupMembership request,
                                    StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to removeUserFromGroup for " + request.getTenantId());

            String group_id = request.getGroupId();
            String username = request.getUsername();
            long tenantId = request.getTenantId();
            String userId = username + "@" + tenantId;

            String effectiveGroupId = group_id + "@" + tenantId;

            List<UserGroupMembership> memberships =
                    groupMembershipRepository.findAllByGroupIdAndUserProfileId(effectiveGroupId, userId);

            List<UserGroupMembership> userGroupMemberships = groupMembershipRepository
                    .findAllByGroupIdAndUserGroupMembershipTypeId(effectiveGroupId, DefaultGroupMembershipTypes.OWNER.name());

            if (userGroupMemberships != null && userGroupMemberships.size() == 1 &&
                    userGroupMemberships.get(0).getUserProfile().getUsername().equals(username)) {
                String msg = "Default owner " + username + " cannot be removed from group " + group_id;
                LOGGER.error(msg);
                responseObserver.onError(Status.ABORTED.withDescription(msg).asRuntimeException());
                return;
            }

            if (memberships != null && !memberships.isEmpty()) {

                memberships.forEach(membership -> {
                    groupMembershipRepository.delete(membership);
                });
            }
            org.apache.custos.user.profile.service.Status status = org.apache.custos.user.profile.service.Status
                    .newBuilder()
                    .setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while removing user from  in client " + request.getClientId() + " reason :"
                    + ex;
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void addChildGroupToParentGroup(GroupToGroupMembership request,
                                           StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to addChildGroupToParentGroup for " + request.getTenantId());

            long tenantId = request.getTenantId();

            String childId = request.getChildId();

            String parentId = request.getParentId();

            String effectiveChildId = childId + "@" + tenantId;
            String effectiveParentId = parentId + "@" + tenantId;


            Optional<org.apache.custos.user.profile.persistance.model.Group> childEntity = groupRepository.
                    findById(effectiveChildId);

            Optional<org.apache.custos.user.profile.persistance.model.Group> parentEntity = groupRepository.
                    findById(effectiveParentId);

            if (childEntity.isEmpty() || parentEntity.isEmpty()) {
                String msg = "Child or parent group not available";
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }

            List<org.apache.custos.user.profile.persistance.model.GroupToGroupMembership> groupToGroupMemberships =
                    groupToGroupMembershipRepository.findByChildIdAndParentId(effectiveChildId, effectiveParentId);
            if (groupToGroupMemberships == null || groupToGroupMemberships.isEmpty()) {

                org.apache.custos.user.profile.persistance.model.GroupToGroupMembership membership =
                        GroupMapper.groupToGroupMembership(childEntity.get(), parentEntity.get());

                org.apache.custos.user.profile.persistance.model.GroupToGroupMembership saved =
                        groupToGroupMembershipRepository.save(membership);

                if (saved != null && saved.getId() != null) {
                    org.apache.custos.user.profile.service.Status status =
                            org.apache.custos.user.profile.service.Status.newBuilder().setStatus(true).build();
                    responseObserver.onNext(status);
                    responseObserver.onCompleted();
                } else {
                    String msg = "Group membership creation failed";
                    LOGGER.error(msg);
                    responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
                }
            } else {
                org.apache.custos.user.profile.service.Status status =
                        org.apache.custos.user.profile.service.Status.newBuilder().setStatus(true).build();
                responseObserver.onNext(status);
                responseObserver.onCompleted();

            }


        } catch (Exception ex) {
            String msg = "Error occurred while adding child group to parent group for " + request.getTenantId() +
                    " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void removeChildGroupFromParentGroup(GroupToGroupMembership request,
                                                StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to removeChildGroupFromParentGroup for " + request.getTenantId());

            long tenantId = request.getTenantId();

            String childId = request.getChildId();

            String parentId = request.getParentId();

            String effectiveChildId = childId + "@" + tenantId;
            String effectiveParentId = parentId + "@" + tenantId;


            Optional<org.apache.custos.user.profile.persistance.model.Group> childEntity = groupRepository.
                    findById(effectiveChildId);

            Optional<org.apache.custos.user.profile.persistance.model.Group> parentEntity = groupRepository.
                    findById(effectiveParentId);

            if (childEntity.isEmpty() || parentEntity.isEmpty()) {
                String msg = "Child or parent group not available";
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
            }


            List<org.apache.custos.user.profile.persistance.model.GroupToGroupMembership> groupToGroupMemberships =
                    groupToGroupMembershipRepository.findByChildIdAndParentId(effectiveChildId, effectiveParentId);

            if (groupToGroupMemberships != null && !groupToGroupMemberships.isEmpty()) {
                groupToGroupMembershipRepository.delete(groupToGroupMemberships.get(0));

            }

            org.apache.custos.user.profile.service.Status status =
                    org.apache.custos.user.profile.service.Status.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while remove child group from parent group for " + request.getTenantId() +
                    " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllGroupsOfUser(UserProfileRequest request, StreamObserver<GetAllGroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllGroupsOfUser for " + request.getTenantId());

            long tenantId = request.getTenantId();
            String username = request.getProfile().getUsername();

            String userId = username + "@" + tenantId;

            List<UserGroupMembership> userGroupMemberships = groupMembershipRepository.findAllByUserProfileId(userId);

            List<org.apache.custos.user.profile.persistance.model.Group> groups = new ArrayList<>();

            if (userGroupMemberships != null && !userGroupMemberships.isEmpty()) {

                userGroupMemberships.forEach(userGroupMembership -> {

                    AtomicBoolean toBeAdded = new AtomicBoolean(true);
                    groups.forEach(le -> {
                        if (le.getId().equals(userGroupMembership.getGroup().getId())) {
                            toBeAdded.set(false);
                        }
                    });

                    if (toBeAdded.get()) {
                        groups.add(userGroupMembership.getGroup());
                    }
                });

            }

            Map<String, org.apache.custos.user.profile.persistance.model.Group> groupMap =
                    getAllUniqueGroups(groups, null);


            List<Group> groupList = new ArrayList<>();


            groupMap.keySet().forEach(gr -> {

                List<UserGroupMembership> memberships = groupMembershipRepository.findAllByGroupId(gr);

                String ownerId = null;
                for (UserGroupMembership userGroupMembership : memberships) {
                    if (userGroupMembership.getUserGroupMembershipType().getId().equals(DefaultGroupMembershipTypes.OWNER.name())) {
                        ownerId = userGroupMembership.getUserProfile().getUsername();
                        break;
                    }
                }
                groupList.add(GroupMapper.createGroup(groupMap.get(gr), ownerId));
            });


            GetAllGroupsResponse getAllGroupsResponse =
                    GetAllGroupsResponse.newBuilder().addAllGroups(groupList).build();
            responseObserver.onNext(getAllGroupsResponse);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while fetching user groups of user " + request.getProfile().getUsername() +
                    " in tenant " + request.getTenantId() +
                    " reason :" + ex;
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllParentGroupsOfGroup(GroupRequest request, StreamObserver<GetAllGroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllParentGroupsOfGroup for " + request.getTenantId());

            String groupId = request.getGroup().getId();
            long tenantId = request.getTenantId();

            String effectiveId = groupId + "@" + tenantId;

            Optional<org.apache.custos.user.profile.persistance.model.Group> groups = groupRepository.findById(effectiveId);

            if (groups.isEmpty()) {
                GetAllGroupsResponse getAllGroupsResponse = GetAllGroupsResponse.newBuilder().build();
                responseObserver.onNext(getAllGroupsResponse);
                responseObserver.onCompleted();
                return;
            } else {
                List<org.apache.custos.user.profile.persistance.model.Group> groupList = new ArrayList<>();
                groupList.add(groups.get());

                Map<String, org.apache.custos.user.profile.persistance.model.Group> groupMap =
                        getAllUniqueGroups(groupList, null);

                List<Group> serviceGroupList = new ArrayList<>();


                groupMap.keySet().forEach(gr -> {

                    List<UserGroupMembership> userGroupMemberships = groupMembershipRepository.findAllByGroupId(effectiveId);

                    String ownerId = null;
                    for (UserGroupMembership userGroupMembership : userGroupMemberships) {
                        if (userGroupMembership.getUserGroupMembershipType().getId().equals(DefaultGroupMembershipTypes.OWNER.name())) {
                            ownerId = userGroupMembership.getUserProfile().getUsername();
                        }
                    }

                    serviceGroupList.add(GroupMapper.createGroup(groupMap.get(gr), ownerId));
                });


                GetAllGroupsResponse getAllGroupsResponse =
                        GetAllGroupsResponse.newBuilder().addAllGroups(serviceGroupList).build();
                responseObserver.onNext(getAllGroupsResponse);
                responseObserver.onCompleted();


            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching all parent groups for group " + request.getGroup().getId() +
                    "in tenant " + request.getTenantId() +
                    " reason :" + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void addUserGroupMembershipType(UserGroupMembershipTypeRequest request,
                                           StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to addUserGroupMembershipType of type  " + request.getType());

            String type = request.getType().toUpperCase();

            Optional<UserGroupMembershipType> userGroupMembershipType = groupMembershipTypeRepository.findById(type);

            if (userGroupMembershipType.isEmpty()) {

                UserGroupMembershipType userGroupType = new UserGroupMembershipType();
                userGroupType.setId(type);
                groupMembershipTypeRepository.save(userGroupType);

            }

            org.apache.custos.user.profile.service.Status status = org.apache.custos.user.profile.service.Status
                    .newBuilder()
                    .setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while saving group membership type" + request.getType();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void removeUserGroupMembershipType(UserGroupMembershipTypeRequest request,
                                              StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to removeUserGroupMembershipType of type " + request.getType());

            String type = request.getType().toUpperCase();

            Optional<UserGroupMembershipType> userGroupMembershipType = groupMembershipTypeRepository.findById(type);

            if (userGroupMembershipType.isPresent()) {

                groupMembershipTypeRepository.delete(userGroupMembershipType.get());
            }

            org.apache.custos.user.profile.service.Status status = org.apache.custos.user.profile.service.Status
                    .newBuilder()
                    .setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while deleting removeUserGroupMembershipType of type  " + request.getType();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllChildUsers(GroupRequest request, StreamObserver<GetAllUserProfilesResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllChildUsers in tenant " + request.getTenantId() +
                    " for group with Id " + request.getGroup().getId());

            long tenantId = request.getTenantId();
            String username = request.getGroup().getId();

            String effectiveId = username + "@" + tenantId;

            Optional<org.apache.custos.user.profile.persistance.model.Group> groupOptional = groupRepository.findById(effectiveId);

            if (groupOptional.isEmpty()) {
                String msg = "group not found";
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }


            List<UserGroupMembership> memberships = groupMembershipRepository.findAllByGroupId(effectiveId);

            List<org.apache.custos.user.profile.service.UserProfile> userProfileList = new ArrayList<>();

            List<UserGroupMembership> selectedProfiles = new ArrayList<>();


            if (memberships != null && !memberships.isEmpty()) {

                memberships.forEach(mem -> {

                    AtomicBoolean addToList = new AtomicBoolean(true);

                    selectedProfiles.forEach(ex -> {
                        if (ex.getId().equals(mem.getUserProfile().getId())) {
                            addToList.set(false);
                        }
                    });

                    if (addToList.get()) {
                        selectedProfiles.add(mem);
                    }
                });
            }

            if (!selectedProfiles.isEmpty()) {
                selectedProfiles.forEach(gr -> {
                    userProfileList.add(UserProfileMapper.createUserProfileFromUserProfileEntity(gr.getUserProfile(),
                            gr.getUserGroupMembershipType().getId()));
                });
            }

            GetAllUserProfilesResponse getAllUserProfilesResponse = GetAllUserProfilesResponse
                    .newBuilder()
                    .addAllProfiles(userProfileList)
                    .build();
            responseObserver.onNext(getAllUserProfilesResponse);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while fetching all child users in tenant " + request.getTenantId() +
                    " for group with Id " + request.getGroup().getId();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getAllChildGroups(GroupRequest request, StreamObserver<GetAllGroupsResponse> responseObserver) {
        try {
            LOGGER.debug("Request received to getAllChildGroups in tenant " + request.getTenantId() +
                    " for group with Id " + request.getGroup().getId());

            long tenantId = request.getTenantId();
            String groupId = request.getGroup().getId();

            String effectiveParentId = groupId + "@" + tenantId;


            Optional<org.apache.custos.user.profile.persistance.model.Group> groupOptional = groupRepository.findById(effectiveParentId);

            if (groupOptional.isEmpty()) {
                String msg = "group not found";
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }


            List<org.apache.custos.user.profile.persistance.model.GroupToGroupMembership> memberships =
                    groupToGroupMembershipRepository.findAllByParentId(effectiveParentId);

            List<org.apache.custos.user.profile.service.Group> groupList = new ArrayList<>();

            HashMap<String, org.apache.custos.user.profile.persistance.model.Group> selectedGroupMap = new HashMap<>();


            if (memberships != null && !memberships.isEmpty()) {

                memberships.forEach(mem -> {
                    selectedGroupMap.put(mem.getChild().getId(), mem.getChild());

                });
            }


            selectedGroupMap.values().forEach(group -> {
                List<UserGroupMembership> groupMemberships = groupMembershipRepository.
                        findAllByGroupId(group.getId());
                AtomicReference<String> ownerId = new AtomicReference<>();
                groupMemberships.forEach(grm -> {
                    if (grm.getUserGroupMembershipType().getId().equals(DefaultGroupMembershipTypes.OWNER.name())) {
                        ownerId.set(grm.getUserProfile().getUsername());
                    }
                });

                groupList.add(GroupMapper.createGroup(group, ownerId.get()));


            });

            GetAllGroupsResponse getAllUserProfilesResponse = GetAllGroupsResponse
                    .newBuilder()
                    .addAllGroups(groupList)
                    .build();
            responseObserver.onNext(getAllUserProfilesResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching all child groups in tenant " + request.getTenantId() +
                    " for group with Id " + request.getGroup().getId();
            LOGGER.error(msg, ex);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void changeUserMembershipType(GroupMembership request,
                                         StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to changeUserMembershipType in   tenant"
                    + request.getTenantId() + " with id " + request.getUsername() + " to " + request.getType());

            long tenantId = request.getTenantId();
            String username = request.getUsername();
            String groupId = request.getGroupId();
            String type = request.getType();


            String userId = username + "@" + tenantId;
            String effectiveGroupId = groupId + "@" + tenantId;


            List<UserGroupMembership> userGroupMemberships =
                    groupMembershipRepository.findAllByGroupIdAndUserProfileId(effectiveGroupId, userId);

            if (userGroupMemberships == null || userGroupMemberships.isEmpty()) {
                String msg = "group membership not found";
                LOGGER.error(msg);
                responseObserver.onError(Status.NOT_FOUND.withDescription(msg).asRuntimeException());
                return;
            }

            List<UserGroupMembership> userMemberships = groupMembershipRepository
                    .findAllByGroupIdAndUserGroupMembershipTypeId(effectiveGroupId, DefaultGroupMembershipTypes.OWNER.name());

            if (userMemberships != null && userMemberships.size() == 1 &&
                    userMemberships.get(0).getUserProfile().getUsername().equals(username)) {
                String msg = "Default owner " + username + " cannot be changed for group " + groupId;
                LOGGER.error(msg);
                responseObserver.onError(Status.ABORTED.withDescription(msg).asRuntimeException());
                return;
            }

            UserGroupMembership groupMembership = userGroupMemberships.get(0);

            UserGroupMembershipType groupMembershipType = groupMembership.getUserGroupMembershipType();
            groupMembershipType.setId(type);

            groupMembership.setUserGroupMembershipType(groupMembershipType);

            groupMembershipRepository.save(groupMembership);

            if (type.equals(DefaultGroupMembershipTypes.OWNER.name())) {

                List<UserGroupMembership> memberships = groupMembershipRepository.findAllByGroupId(effectiveGroupId);

                if (memberships != null && !memberships.isEmpty()) {

                    for (UserGroupMembership membership : memberships) {

                        if (membership.getUserGroupMembershipType().getId().equals(DefaultGroupMembershipTypes.OWNER.name())
                                && !membership.getUserProfile().getUsername().equals(userId)) {

                            groupMembershipRepository.delete(membership);

                            UserGroupMembership userGroupMembership = new
                                    UserGroupMembership();
                            userGroupMembership.setGroup(membership.getGroup());
                            userGroupMembership.setUserProfile(membership.getUserProfile());
                            userGroupMembership.setTenantId(tenantId);

                            Optional<UserGroupMembershipType> membershipType = groupMembershipTypeRepository.
                                    findById(DefaultGroupMembershipTypes.MEMBER.name());
                            userGroupMembership.setUserGroupMembershipType(membershipType.get());
                            groupMembershipRepository.save(userGroupMembership);

                        }
                    }


                }

            }


            org.apache.custos.user.profile.service.Status status = org.apache.custos.user.profile.service.Status
                    .newBuilder()
                    .setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while changing membership type  in   tenant"
                    + request.getTenantId() + " with id " + request.getUsername() + " to " + request.getType();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void hasAccess(GroupMembership request,
                          StreamObserver<org.apache.custos.user.profile.service.Status> responseObserver) {
        try {
            LOGGER.debug("Request received to check access  in   tenant"
                    + request.getTenantId() + " with id " + request.getUsername() + " to " + request.getType());

            long tenantId = request.getTenantId();
            String username = request.getUsername();
            String groupId = request.getGroupId();
            String type = request.getType();


            String userId = username + "@" + tenantId;
            String effectiveGroupId = groupId + "@" + tenantId;


            List<UserGroupMembership> userGroupMemberships = groupMembershipRepository.
                    findAllByGroupIdAndUserProfileIdAndUserGroupMembershipTypeId(effectiveGroupId, userId, type);
            org.apache.custos.user.profile.service.Status status = null;

            if (userGroupMemberships == null || userGroupMemberships.isEmpty()) {
                status = org.apache.custos.user.profile.service.Status
                        .newBuilder()
                        .setStatus(false).build();
            } else {
                status = org.apache.custos.user.profile.service.Status
                        .newBuilder()
                        .setStatus(true).build();
            }


            responseObserver.onNext(status);
            responseObserver.onCompleted();


        } catch (Exception ex) {
            String msg = "Error occurred while checking access to   in   tenant"
                    + request.getTenantId() + " with id " + request.getUsername() + " to " + request.getType();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    public boolean initializeDBConfigs() {

        for (DefaultGroupMembershipTypes type : DefaultGroupMembershipTypes.values()) {

            Optional<UserGroupMembershipType> typeOptional = groupMembershipTypeRepository.
                    findById(type.name().toUpperCase());

            if (typeOptional.isEmpty()) {
                UserGroupMembershipType membershipType = new UserGroupMembershipType();
                membershipType.setId(type.name().toUpperCase());
                groupMembershipTypeRepository.save(membershipType);
            }
        }
        return true;

    }

    private Map<String, org.apache.custos.user.profile.persistance.model.Group>
    getAllUniqueGroups(List<org.apache.custos.user.profile.persistance.model.Group> leaveGroups,
                       Map<String, org.apache.custos.user.profile.persistance.model.Group> allParentGroups) {

        if (allParentGroups == null) {
            allParentGroups = new HashMap<>();

        }

        if (leaveGroups != null && !leaveGroups.isEmpty()) {

            for (org.apache.custos.user.profile.persistance.model.Group gr : leaveGroups) {
                List<org.apache.custos.user.profile.persistance.model.GroupToGroupMembership> memberships
                        = groupToGroupMembershipRepository.findAllByChildId(gr.getId());
                List<org.apache.custos.user.profile.persistance.model.Group> leaves = new ArrayList<>();
                allParentGroups.put(gr.getId(), gr);

                if (memberships != null && !memberships.isEmpty()) {
                    memberships.forEach(mem -> {
                        AtomicBoolean toBeAdded = new AtomicBoolean(true);
                        leaves.forEach(le -> {
                            if (le.getId().equals(mem.getParent().getId())) {
                                toBeAdded.set(false);
                            }
                        });

                        if (toBeAdded.get()) {
                            leaves.add(mem.getParent());
                        }
                    });

                    getAllUniqueGroups(leaves, allParentGroups);

                }

            }


        }

        return allParentGroups;

    }


}
