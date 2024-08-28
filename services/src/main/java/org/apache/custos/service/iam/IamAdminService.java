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

package org.apache.custos.service.iam;

import jakarta.persistence.EntityNotFoundException;
import org.apache.commons.lang3.StringUtils;
import org.apache.custos.core.commons.StatusUpdater;
import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.exception.UnauthorizedException;
import org.apache.custos.core.iam.api.AddExternalIDPLinksRequest;
import org.apache.custos.core.iam.api.AddProtocolMapperRequest;
import org.apache.custos.core.iam.api.AddRolesRequest;
import org.apache.custos.core.iam.api.AddUserAttributesRequest;
import org.apache.custos.core.iam.api.AddUserRolesRequest;
import org.apache.custos.core.iam.api.Agent;
import org.apache.custos.core.iam.api.AgentClientMetadata;
import org.apache.custos.core.iam.api.AllRoles;
import org.apache.custos.core.iam.api.CheckingResponse;
import org.apache.custos.core.iam.api.ConfigureFederateIDPRequest;
import org.apache.custos.core.iam.api.DeleteExternalIDPsRequest;
import org.apache.custos.core.iam.api.DeleteRoleRequest;
import org.apache.custos.core.iam.api.DeleteTenantRequest;
import org.apache.custos.core.iam.api.DeleteUserAttributeRequest;
import org.apache.custos.core.iam.api.DeleteUserRolesRequest;
import org.apache.custos.core.iam.api.EventPersistenceRequest;
import org.apache.custos.core.iam.api.ExternalIDPLink;
import org.apache.custos.core.iam.api.FederateIDPResponse;
import org.apache.custos.core.iam.api.FindUsersRequest;
import org.apache.custos.core.iam.api.FindUsersResponse;
import org.apache.custos.core.iam.api.GetAllResources;
import org.apache.custos.core.iam.api.GetAllResourcesResponse;
import org.apache.custos.core.iam.api.GetExternalIDPsRequest;
import org.apache.custos.core.iam.api.GetExternalIDPsResponse;
import org.apache.custos.core.iam.api.GetOperationsMetadataRequest;
import org.apache.custos.core.iam.api.GetOperationsMetadataResponse;
import org.apache.custos.core.iam.api.GetRolesRequest;
import org.apache.custos.core.iam.api.GroupRepresentation;
import org.apache.custos.core.iam.api.GroupRequest;
import org.apache.custos.core.iam.api.GroupsRequest;
import org.apache.custos.core.iam.api.GroupsResponse;
import org.apache.custos.core.iam.api.MapperTypes;
import org.apache.custos.core.iam.api.OperationMetadata;
import org.apache.custos.core.iam.api.RegisterUserRequest;
import org.apache.custos.core.iam.api.RegisterUserResponse;
import org.apache.custos.core.iam.api.RegisterUsersRequest;
import org.apache.custos.core.iam.api.RegisterUsersResponse;
import org.apache.custos.core.iam.api.ResetUserPassword;
import org.apache.custos.core.iam.api.ResourceTypes;
import org.apache.custos.core.iam.api.RoleRepresentation;
import org.apache.custos.core.iam.api.SetUpTenantRequest;
import org.apache.custos.core.iam.api.SetUpTenantResponse;
import org.apache.custos.core.iam.api.UpdateUserProfileRequest;
import org.apache.custos.core.iam.api.UserAttribute;
import org.apache.custos.core.iam.api.UserGroupMappingRequest;
import org.apache.custos.core.iam.api.UserSearchRequest;
import org.apache.custos.core.model.commons.OperationStatus;
import org.apache.custos.core.model.commons.StatusEntity;
import org.apache.custos.service.auth.TokenService;
import org.apache.custos.service.federated.client.keycloak.KeycloakClient;
import org.apache.custos.service.federated.client.keycloak.KeycloakClientSecret;
import org.keycloak.representations.idm.EventRepresentation;
import org.keycloak.representations.idm.FederatedIdentityRepresentation;
import org.keycloak.representations.idm.ProtocolMapperRepresentation;
import org.keycloak.representations.idm.UserRepresentation;
import org.keycloak.representations.idm.UserSessionRepresentation;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.text.ParseException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
public class IamAdminService {

    private static final Logger LOGGER = LoggerFactory.getLogger(IamAdminService.class);

    private final KeycloakClient keycloakClient;

    private final StatusUpdater statusUpdater;

    private final TokenService tokenService;

    @Value("${iam.server.url}")
    private String iamServerURL;

    public IamAdminService(KeycloakClient keycloakClient, StatusUpdater statusUpdater, TokenService tokenService) {
        this.keycloakClient = keycloakClient;
        this.statusUpdater = statusUpdater;
        this.tokenService = tokenService;
    }

    public SetUpTenantResponse setUPTenant(SetUpTenantRequest request) {
        try {
            LOGGER.debug("Request received to setUPTenant  " + request.getTenantId());

            keycloakClient.deleteRealm(String.valueOf(request.getTenantId()));
            keycloakClient.createRealm(String.valueOf(request.getTenantId()), request.getTenantName());
            keycloakClient.createRealmAdminAccount(String.valueOf(request.getTenantId()), request.getAdminUsername(),
                    request.getAdminFirstname(), request.getAdminLastname(), request.getAdminEmail(), request.getAdminPassword());

            KeycloakClientSecret clientSecret = keycloakClient.configureClient(String.valueOf(request.getTenantId()),
                    request.getCustosClientId(), request.getTenantURL(), request.getRedirectURIsList());

            SetUpTenantResponse response = SetUpTenantResponse.newBuilder()
                    .setClientId(clientSecret.getClientId())
                    .setClientSecret(clientSecret.getClientSecret())
                    .build();

            statusUpdater.updateStatus(IAMOperations.SET_UP_TENANT.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getRequesterEmail());
            return response;

        } catch (Exception ex) {
            String msg = "Error occurred during setUPTenant" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.SET_UP_TENANT.name(), OperationStatus.FAILED, request.getTenantId(), request.getRequesterEmail());
            throw new RuntimeException(msg, ex);
        }
    }

    public SetUpTenantResponse updateTenant(SetUpTenantRequest request) {
        try {
            LOGGER.debug("Request received to updateTenant  " + request.getTenantId());

            keycloakClient.updateRealm(String.valueOf(request.getTenantId()), request.getTenantName());
            keycloakClient.updateRealmAdminAccount(String.valueOf(request.getTenantId()), request.getAdminUsername(),
                    request.getAdminFirstname(), request.getAdminLastname(), request.getAdminEmail(), request.getAdminPassword());

            KeycloakClientSecret clientSecret = keycloakClient.updateClient(String.valueOf(request.getTenantId()),
                    request.getCustosClientId(), request.getTenantURL(), request.getRedirectURIsList());

            SetUpTenantResponse response = SetUpTenantResponse.newBuilder()
                    .setClientId(clientSecret.getClientId())
                    .setClientSecret(clientSecret.getClientSecret())
                    .build();

            statusUpdater.updateStatus(IAMOperations.UPDATE_TENANT.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getRequesterEmail());
            return response;

        } catch (Exception ex) {
            String msg = "Error occurred during updateTenant" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.UPDATE_TENANT.name(), OperationStatus.FAILED, request.getTenantId(), request.getRequesterEmail());
            throw new RuntimeException(msg, ex);
        }
    }

    public void deleteTenant(DeleteTenantRequest request) {
        try {
            LOGGER.debug("Request received to delete tenant  " + request.getTenantId());
            keycloakClient.deleteRealm(String.valueOf(request.getTenantId()));

        } catch (Exception ex) {
            String msg = "Error occurred during delete tenant" + ex;
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus isUsernameAvailable(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to isUsernameAvailable at " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUsernameAvailable(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(isAvailable).build();

        } catch (Exception ex) {
            String msg = "Error occurred during isUsernameAvailable" + ex;
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus isUserEnabled(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to isUserEnabled at " + request.getTenantId());

            boolean isAvailable = keycloakClient.isUserAccountEnabled(String.valueOf(request.getTenantId()),
                    request.getAccessToken(),
                    request.getUser().getUsername());
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(isAvailable).build();

        } catch (Exception ex) {
            String msg = "Error occurred during isUserEnabled" + ex;
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public RegisterUserResponse registerUser(RegisterUserRequest request) {

        try {
            LOGGER.debug("Request received to registerUser for " + request.getTenantId());

            boolean registered = keycloakClient.createUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(),
                    request.getUser().getPassword(),
                    request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail(),
                    request.getUser().getTemporaryPassword(),
                    request.getAccessToken());

            RegisterUserResponse registerUserResponse = RegisterUserResponse.newBuilder().setIsRegistered(registered).build();
            statusUpdater.updateStatus(IAMOperations.REGISTER_USER.name(), OperationStatus.SUCCESS,
                    request.getTenantId(), String.valueOf(request.getTenantId()));

            return registerUserResponse;

        } catch (UnauthorizedException ex) {
            String msg = "Error occurred during registerUser" + ex;
            statusUpdater.updateStatus(IAMOperations.REGISTER_USER.name(), OperationStatus.FAILED, request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            throw new RuntimeException(msg, ex);

        } catch (Exception ex) {
            String msg = "Error occurred during registerUser" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.REGISTER_USER.name(), OperationStatus.FAILED, request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.UserRepresentation enableUser(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to enableUser for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());

            if (!status) {
                statusUpdater.updateStatus(IAMOperations.ENABLE_USER.name(), OperationStatus.FAILED, request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                String msg = "User not valid " + request.getUser().getId();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

            boolean accountEnabled = keycloakClient.enableUserAccount(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUser().getUsername());

            if (accountEnabled) {
                UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()),
                        request.getAccessToken(), request.getUser().getUsername());

                org.apache.custos.core.iam.api.UserRepresentation user = getUser(representation, request.getClientId());

                statusUpdater.updateStatus(IAMOperations.ENABLE_USER.name(), OperationStatus.SUCCESS, request.getTenantId(),
                        String.valueOf(request.getTenantId()));
                return user;

            } else {
                statusUpdater.updateStatus(IAMOperations.ENABLE_USER.name(), OperationStatus.FAILED, request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                String msg = "Account enabling failed for user: " + request.getUser().getId() + " in tenant: " + request.getTenantId();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }


        } catch (Exception ex) {
            String msg = "Error occurred during enableUser" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.ENABLE_USER.name(), OperationStatus.FAILED, request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.UserRepresentation disableUser(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to disable for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());

            if (!status) {
                statusUpdater.updateStatus(IAMOperations.DISABLE_USER.name(), OperationStatus.FAILED, request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                String msg = "User not valid, user ID: " + request.getUser().getId();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

            boolean accountDisabled = keycloakClient.disableUserAccount(String.valueOf(request.getTenantId()),
                    request.getAccessToken(), request.getUser().getUsername());

            if (accountDisabled) {
                UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()),
                        request.getAccessToken(), request.getUser().getUsername());

                org.apache.custos.core.iam.api.UserRepresentation user = getUser(representation, request.getClientId());
                statusUpdater.updateStatus(IAMOperations.DISABLE_USER.name(), OperationStatus.SUCCESS, request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                return user;

            } else {
                statusUpdater.updateStatus(IAMOperations.DISABLE_USER.name(), OperationStatus.FAILED, request.getTenantId(),
                        String.valueOf(request.getTenantId()));

                String msg = "Account disabling failed for user: " + request.getUser().getId();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

        } catch (Exception ex) {
            String msg = "Error occurred during disabling user" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DISABLE_USER.name(), OperationStatus.FAILED, request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            throw new RuntimeException(msg, ex);
        }
    }

    public CheckingResponse isUserExist(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to isUserExist for " + request.getTenantId());

            boolean isUserExist = keycloakClient.isUserExist(String.valueOf(request.getTenantId()), request.getAccessToken(), request.getUser().getUsername());
            return CheckingResponse.newBuilder().setIsExist(isUserExist).build();

        } catch (Exception ex) {
            String msg = "Error occurred during isUserExist" + ex;
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.UserRepresentation getUser(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to getUser for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), request.getUser().getUsername());

            if (!status) {
                String msg = "User " + request.getUser().getUsername() + "not found at " + request.getTenantId();
                throw new EntityNotFoundException(msg);
            }

            UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()), request.getUser().getUsername());

            if (representation != null) {
                org.apache.custos.core.iam.api.UserRepresentation user = getUser(representation, request.getClientId());

                UserSessionRepresentation sessionRepresentation = keycloakClient.getLatestSession(String.valueOf(request.getTenantId()),
                        request.getClientId(), request.getAccessToken(), request.getUser().getUsername());

                if (sessionRepresentation != null) {
                    user = user.toBuilder().setLastLoginAt(sessionRepresentation.getLastAccess()).build();

                } else {
                    EventRepresentation eventRepresentation = keycloakClient.getLastLoginEvent(String.valueOf(request.getTenantId()),
                            request.getClientId(), request.getUser().getUsername());

                    if (eventRepresentation != null) {
                        user = user.toBuilder().setLastLoginAt(eventRepresentation.getTime()).build();
                    }
                }
                return user;

            } else {
                String msg = "User " + request.getUser().getUsername() + " not found in " + request.getTenantId();
                throw new EntityNotFoundException(msg);
            }

        } catch (Exception ex) {
            String msg = "Error occurred during getUser" + ex;
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("Unauthorized")) {
                throw new RuntimeException("Unauthorized request", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public FindUsersResponse findUsers(FindUsersRequest request) {
        try {
            LOGGER.debug("Request received to getUsers for " + request.getUser().getUsername());
            long initiationTime = System.currentTimeMillis();
            List<UserRepresentation> representation = keycloakClient.getUsers(request.getAccessToken(),
                    String.valueOf(request.getTenantId()), request.getOffset(), request.getLimit(),
                    request.getUser().getUsername(), request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail(),
                    request.getUser().getId());
            List<org.apache.custos.core.iam.api.UserRepresentation> users = new ArrayList<>();
            representation.forEach(r -> {
                boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), r.getUsername(), request.getAccessToken());

                if (status) {

                    org.apache.custos.core.iam.api.UserRepresentation user = this.getUser(r, request.getClientId());

                    UserSessionRepresentation sessionRepresentation = keycloakClient.getLatestSession(String.valueOf(request.getTenantId()),
                            request.getClientId(), null, r.getUsername());

                    if (sessionRepresentation != null) {
                        user = user.toBuilder().setLastLoginAt(sessionRepresentation.getLastAccess()).build();
                    } else {
                        EventRepresentation eventRepresentation = keycloakClient.getLastLoginEvent(String.valueOf(request.getTenantId()), request.getClientId(), r.getUsername());

                        if (eventRepresentation != null) {
                            user = user.toBuilder().setLastLoginAt(eventRepresentation.getTime()).build();
                        }
                    }
                    users.add(user);
                }
            });

            long endTime = System.currentTimeMillis();
            long total = endTime - initiationTime;
            LOGGER.info("request received: " + initiationTime + " request end time" + endTime + " difference " + total);
            return FindUsersResponse.newBuilder().addAllUsers(users).build();

        } catch (Exception ex) {
            String msg = "Error occurred during getUsers" + ex;
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("Unauthorized")) {
                throw new RuntimeException("Unauthorized request", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus resetPassword(ResetUserPassword request) {
        String userId = request.getUsername() + "@" + request.getTenantId();
        try {
            LOGGER.debug("Request received to resetPassword for " + request.getUsername());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), request.getUsername(), request.getAccessToken());

            if (!status) {
                statusUpdater.updateStatus(IAMOperations.RESET_PASSWORD.name(), OperationStatus.FAILED, request.getTenantId(), userId);
                String msg = "User not valid, user name: " + request.getUsername();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

            boolean isChanged = keycloakClient.resetUserPassword(request.getAccessToken(),
                    String.valueOf(request.getTenantId()), request.getUsername(), request.getPassword());
            statusUpdater.updateStatus(IAMOperations.RESET_PASSWORD.name(), OperationStatus.SUCCESS, request.getTenantId(), userId);

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(isChanged).build();

        } catch (Exception ex) {
            String msg = "Error occurred during resetPassword" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.RESET_PASSWORD.name(), OperationStatus.FAILED, request.getTenantId(), userId);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteExternalIDPLinksOfUsers(DeleteExternalIDPsRequest request) {
        try {
            long tenantId = request.getTenantId();
            boolean status = request.getUserIdList().isEmpty()
                    ? keycloakClient.deleteExternalIDPLinks(String.valueOf(tenantId))
                    : keycloakClient.deleteExternalIDPLinks(String.valueOf(tenantId), request.getUserIdList());
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = "Error occurred while deletingExternalIDPLinksOfUsers" + ex;
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public GetExternalIDPsResponse getExternalIDPLinksOfUsers(GetExternalIDPsRequest request) {
        try {
            long tenantId = request.getTenantId();
            List<FederatedIdentityRepresentation> identityRepresentations = keycloakClient.getExternalIDPLinks(String.valueOf(tenantId), request.getUserId());
            GetExternalIDPsResponse.Builder response = GetExternalIDPsResponse.newBuilder();
            identityRepresentations.forEach(rep -> response.addIdpLinks(ExternalIDPLink.newBuilder()
                    .setProviderAlias(rep.getIdentityProvider())
                    .setProviderUsername(rep.getUserName())
                    .setProviderUserId(rep.getUserId())));

            return response.build();

        } catch (Exception ex) {
            String msg = "Error occurred while getExternalIDPLinksOfUsers" + ex;
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus addExternalIDPLinksOfUsers(AddExternalIDPLinksRequest request) {
        try {
            long tenantId = request.getTenantId();
            List<ExternalIDPLink> externalIDPLinkList = request.getIdpLinksList();
            List<FederatedIdentityRepresentation> federatedIdentityRepresentations = new ArrayList<>();
            externalIDPLinkList.forEach(link -> {
                FederatedIdentityRepresentation representation = new FederatedIdentityRepresentation();
                representation.setUserId(link.getProviderUserId());
                representation.setUserName(link.getProviderUsername());
                representation.setIdentityProvider(link.getProviderAlias());
                federatedIdentityRepresentations.add(representation);

            });
            keycloakClient.addExternalIDPLinks(String.valueOf(tenantId), federatedIdentityRepresentations);
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = "Error occurred while getExternalIDPLinksOfUsers" + ex;
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus updateUserProfile(UpdateUserProfileRequest request) {
        String userId = request.getUser().getUsername() + "@" + request.getTenantId();
        try {
            LOGGER.debug("Request received to updateUserProfile for " + request.getUser().getUsername());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), request.getUser().getUsername(), request.getAccessToken());
            if (!status) {
                statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(), OperationStatus.FAILED, request.getTenantId(), userId);
                String msg = "User not valid, user Id: " + request.getUser().getId();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

            keycloakClient.updateUserRepresentation(request.getAccessToken(),
                    String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(),
                    request.getUser().getFirstName(),
                    request.getUser().getLastName(),
                    request.getUser().getEmail());

            statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(), OperationStatus.SUCCESS, request.getTenantId(), userId);
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = "Error occurred during updateUserProfile" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.UPDATE_USER_PROFILE.name(), OperationStatus.FAILED, request.getTenantId(), userId);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteUser(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to deleteUser for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), request.getUser().getUsername(), request.getAccessToken());

            if (!status) {
                statusUpdater.updateStatus(IAMOperations.DELETE_USER.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
                String msg = "User not valid, user Id: " + request.getUser().getId();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

            boolean isUpdated = keycloakClient.deleteUser(request.getAccessToken(), String.valueOf(request.getTenantId()), request.getUser().getUsername());
            statusUpdater.updateStatus(IAMOperations.DELETE_USER.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(isUpdated).build();

        } catch (Exception ex) {
            String msg = "Error occurred during deleteUser" + ex;
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_USER.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());

            if (ex.getMessage().contains("Unauthorized")) {
                throw new RuntimeException("Unauthorized request", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteRolesFromUser(DeleteUserRolesRequest request) {
        try {
            LOGGER.debug("Request received to deleteRoleFromUser for " + request.getTenantId());

            boolean status = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), request.getUsername());

            if (!status) {
                statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_USER.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), request.getPerformedBy());
                String msg = "User not valid ";
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

            if (!request.getRolesList().isEmpty()) {
                keycloakClient.removeRoleFromUser(request.getAccessToken(), String.valueOf(request.getTenantId()), request.getUsername(),
                        request.getRolesList(), request.getClientId(), false);
            }

            if (!request.getClientRolesList().isEmpty()) {
                keycloakClient.removeRoleFromUser(request.getAccessToken(), String.valueOf(request.getTenantId()), request.getUsername(),
                        request.getClientRolesList(), request.getClientId(), true);

            }

            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_USER.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = "Error occurred during deleteRoleFromUser" + ex;
            LOGGER.error(msg, ex);

            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_USER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());

            if (ex.getMessage().contains("Unauthorized")) {
                throw new RuntimeException("Unauthorized request", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public GetOperationsMetadataResponse getOperationMetadata(GetOperationsMetadataRequest request) {
        try {
            LOGGER.debug("Calling getOperationMetadata API for traceId " + request.getTraceId());

            List<OperationMetadata> metadata = new ArrayList<>();
            List<StatusEntity> entities = statusUpdater.getOperationStatus(request.getTraceId());
            if (entities != null && !entities.isEmpty()) {
                metadata = entities.stream()
                        .map(this::convertFromEntity)
                        .collect(Collectors.toList());
            }

            return GetOperationsMetadataResponse.newBuilder().addAllMetadata(metadata).build();

        } catch (Exception ex) {
            String msg = " operation failed for " + request.getTraceId();
            LOGGER.error(msg);
            throw new RuntimeException(msg);
        }
    }

    public FederateIDPResponse configureFederatedIDP(ConfigureFederateIDPRequest request) {
        try {
            LOGGER.debug("Request received to configureFederatedIDP for " + request.getTenantId());

            keycloakClient.configureOIDCFederatedIDP(String.valueOf(request.getTenantId()), "CILogon", request.getScope(),
                    new KeycloakClientSecret(request.getClientID(), request.getClientSec()), null);

            statusUpdater.updateStatus(IAMOperations.CONFIGURE_IDP.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getRequesterEmail());

            return FederateIDPResponse.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = " Configure Federated IDP failed for " + request.getTenantId();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }

    public RegisterUsersResponse registerAndEnableUsers(RegisterUsersRequest request) {
        try {
            LOGGER.debug("Request received to registerMultipleUsers for " + request.getTenantId());

            List<org.apache.custos.core.iam.api.UserRepresentation> userRepresentations = request.getUsersList();
            List<org.apache.custos.core.iam.api.UserRepresentation> failedList = new ArrayList<>();

            for (org.apache.custos.core.iam.api.UserRepresentation userRepresentation : userRepresentations) {
                try {
                    keycloakClient.createUser(String.valueOf(request.getTenantId()),
                            userRepresentation.getUsername(),
                            userRepresentation.getPassword(),
                            userRepresentation.getFirstName(),
                            userRepresentation.getLastName(),
                            userRepresentation.getEmail(),
                            userRepresentation.getTemporaryPassword(),
                            request.getAccessToken());

                    keycloakClient.enableUserAccount(String.valueOf(request.getTenantId()),
                            request.getAccessToken(), userRepresentation.getUsername().toLowerCase());
                    List<String> userList = new ArrayList<>();
                    userList.add(userRepresentation.getUsername());

                    if (!userRepresentation.getRealmRolesList().isEmpty()) {
                        keycloakClient.addRolesToUsers(request.getAccessToken(), String.valueOf(request.getTenantId()), userList, userRepresentation.getRealmRolesList(),
                                request.getClientId(), false);
                    }
                    if (!userRepresentation.getClientRolesList().isEmpty()) {
                        keycloakClient.addRolesToUsers(request.getAccessToken(), String.valueOf(request.getTenantId()), userList, userRepresentation.getClientRolesList(),
                                request.getClientId(), true);
                    }

                    if (!userRepresentation.getAttributesList().isEmpty()) {

                        Map<String, List<String>> map = new HashMap<>();
                        for (UserAttribute attribute : userRepresentation.getAttributesList()) {

                            if (attribute.getKey().equals(Constants.REALM_AGENT)) {
                                // Constants.REALM_AGENT cannot be used as a valid attribute
                                continue;
                            }
                            map.put(attribute.getKey(), attribute.getValuesList());
                        }

                        keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), map, userList);
                    }

                } catch (UnauthorizedException ex) {
                    String msg = " Error occurred while adding user " + userRepresentation.getUsername() +
                            " to realm" + request.getTenantId();
                    LOGGER.error(msg);
                    throw new RuntimeException(msg, ex);

                } catch (Exception ex) {
                    if (ex.getMessage().contains("Unauthorized")) {
                        throw new RuntimeException("Unauthorized request", ex);
                    }

                    LOGGER.error(" Error occurred while adding user " + userRepresentation.getUsername() + " to realm" + request.getTenantId());
                    failedList.add(userRepresentation);
                }
            }

            if (failedList.isEmpty()) {
                statusUpdater.updateStatus(IAMOperations.REGISTER_ENABLE_USERS.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
            }

            return RegisterUsersResponse.newBuilder().setAllUseresRegistered(failedList.isEmpty()).addAllFailedUsers(failedList).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.REGISTER_ENABLE_USERS.name(), OperationStatus.FAILED, request.getTenantId(), String.valueOf(request.getTenantId()));
            String msg = " Register  multiple users  failed for " + request.getTenantId();
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus addRolesToUsers(AddUserRolesRequest request) {
        try {
            LOGGER.debug("Request received to addRolesToUsers for " + request.getTenantId());

            List<String> validUserNames = new ArrayList<>();

            for (String username : request.getUsernamesList()) {
                if (keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), username)) {
                    validUserNames.add(username);
                }
            }

            keycloakClient.addRolesToUsers(request.getAccessToken(), String.valueOf(request.getTenantId()),
                    validUserNames, request.getRolesList(), request.getClientId(), request.getClientLevel());
            statusUpdater.updateStatus(IAMOperations.ADD_ROLES_TO_USERS.name(), OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_ROLES_TO_USERS.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = "Add multiple users failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public AllRoles addRolesToTenant(AddRolesRequest request) {
        try {
            LOGGER.debug("Request received to add roles to tenant for " + request.getTenantId());

            List<RoleRepresentation> rolesRepresentations = request.getRolesList();
            List<org.keycloak.representations.idm.RoleRepresentation> keycloakRolesList = new ArrayList<>();

            for (RoleRepresentation roleRepresentation : rolesRepresentations) {
                org.keycloak.representations.idm.RoleRepresentation role = new org.keycloak.representations.idm.RoleRepresentation();
                role.setName(roleRepresentation.getName());
                role.setDescription(roleRepresentation.getDescription());
                role.setComposite(roleRepresentation.getComposite());
                keycloakRolesList.add(role);
            }

            keycloakClient.addRoles(keycloakRolesList, String.valueOf(request.getTenantId()), request.getClientId(), request.getClientLevel());

            statusUpdater.updateStatus(IAMOperations.ADD_ROLES_TO_TENANT.name(), OperationStatus.SUCCESS,
                    request.getTenantId(), String.valueOf(request.getTenantId()));

            List<org.keycloak.representations.idm.RoleRepresentation> allKeycloakRoles = keycloakClient.getAllRoles(String.valueOf(request.getTenantId()),
                    (request.getClientLevel()) ? request.getClientId() : null);
            AllRoles.Builder builder = AllRoles.newBuilder();

            if (allKeycloakRoles != null && !allKeycloakRoles.isEmpty()) {
                List<RoleRepresentation> roleRepresentations = allKeycloakRoles.stream()
                        .map(role -> RoleRepresentation
                                .newBuilder()
                                .setName(role.getName())
                                .setComposite(role.isComposite())
                                .setId(role.getId())
                                .setDescription(Optional.ofNullable(role.getDescription()).orElse(""))
                                .build())
                        .collect(Collectors.toList());

                builder.addAllRoles(roleRepresentations);
                builder.setScope(request.getClientLevel() ? "client_level" : "realm_level");
            }
            return builder.build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_ROLES_TO_TENANT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            String msg = " Add roles   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteRole(DeleteRoleRequest request) {
        try {
            LOGGER.debug("Request received to add roles to tenant for " + request.getTenantId());

            keycloakClient.deleteRole(request.getRole().getId(), String.valueOf(request.getTenantId()),
                    request.getClientId(), request.getClientLevel());
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = "Deleting role" + request.getRole().getName() + "  " + "failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public AllRoles getRolesOfTenant(GetRolesRequest request) {
        try {
            LOGGER.debug("Request received to add roles to tenant for " + request.getTenantId());

            List<org.keycloak.representations.idm.RoleRepresentation> allKeycloakRoles = keycloakClient.getAllRoles(String.valueOf(request.getTenantId()), (request.getClientLevel()) ? request.getClientId() : null);
            AllRoles.Builder builder = AllRoles.newBuilder();
            if (allKeycloakRoles != null && !allKeycloakRoles.isEmpty()) {
                List<RoleRepresentation> roleRepresentations = new ArrayList<>();
                for (org.keycloak.representations.idm.RoleRepresentation role : allKeycloakRoles) {
                    RoleRepresentation roleRepresentation = RoleRepresentation.
                            newBuilder().setName(role.getName())
                            .setComposite(role.isComposite())
                            .setId(role.getId())
                            .build();
                    if (role.getDescription() != null) {
                        roleRepresentation = roleRepresentation.toBuilder().setDescription(role.getDescription()).build();
                    }
                    roleRepresentations.add(roleRepresentation);

                }

                builder.addAllRoles(roleRepresentations);
                if (request.getClientLevel()) {
                    builder.setScope("client_level");
                } else {
                    builder.setScope("realm_level");
                }
            }
            return builder.build();

        } catch (Exception ex) {
            String msg = "Get roles failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus addProtocolMapper(AddProtocolMapperRequest request) {
        try {
            LOGGER.debug("Request received to add protocol mapper " + request.getTenantId());

            String mapperModel;
            Map<String, String> configMap = new HashMap<>();
            if (request.getMapperType().equals(MapperTypes.USER_ATTRIBUTE)) {
                mapperModel = "oidc-usermodel-attribute-mapper";
                configMap.put("user.attribute", request.getAttributeName());
            } else if (request.getMapperType().equals(MapperTypes.USER_REALM_ROLE)) {
                mapperModel = "oidc-usermodel-realm-role-mapper";
            } else if (request.getMapperType().equals(MapperTypes.USER_CLIENT_ROLE)) {
                mapperModel = "oidc-usermodel-client-role-mapper";
                configMap.put("usermodel.clientRoleMapping.clientId", request.getClientId());
            } else {
                throw new RuntimeException("Mapping type not supported");
            }

            ProtocolMapperRepresentation protocolMapperRepresentation = new ProtocolMapperRepresentation();
            protocolMapperRepresentation.setName(request.getName());
            protocolMapperRepresentation.setProtocol("openid-connect");
            protocolMapperRepresentation.setProtocolMapper(mapperModel);

            configMap.put("user.session.note", request.getClaimName());
            configMap.put("id.token.claim", String.valueOf(request.getAddToIdToken()));
            configMap.put("access.token.claim", String.valueOf(request.getAddToAccessToken()));
            configMap.put("claim.name", request.getClaimName());
            switch (request.getClaimType()) {
                case JSON:
                    configMap.put("jsonType.label", "JSON");
                    break;
                case LONG:
                    configMap.put("jsonType.label", "long");
                    break;
                case STRING:
                    configMap.put("jsonType.label", "String");
                    break;
                case BOOLEAN:
                    configMap.put("jsonType.label", "boolean");
                    break;
                case INTEGER:
                    configMap.put("jsonType.label", "int");
                    break;
                default: {
                    throw new RuntimeException("Unknown claim type");
                }
            }

            configMap.put("aggregate.attrs", String.valueOf(request.getAggregateAttributeValues()));
            configMap.put("userinfo.token.claim", String.valueOf(request.getAddToUserInfo()));
            configMap.put("multivalued", String.valueOf(request.getMultiValued()));


            protocolMapperRepresentation.setConfig(configMap);

            keycloakClient.addProtocolMapper(protocolMapperRepresentation, String.valueOf(request.getTenantId()),
                    request.getClientId());

            statusUpdater.updateStatus(IAMOperations.ADD_PROTOCOL_MAPPER.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_PROTOCOL_MAPPER.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    String.valueOf(request.getTenantId()));
            String msg = " Add protocol mapper   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus addUserAttributes(AddUserAttributesRequest request) {
        try {
            LOGGER.debug("Request received to addUserAttributes " + request.getTenantId());

            List<UserAttribute> attributes = request.getAttributesList();
            List<String> validUserNames = new ArrayList<>();

            for (String username : request.getUsersList()) {
                if (keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), username, request.getAccessToken())) {
                    validUserNames.add(username);
                }
            }

            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                if (attribute.getKey().equals(Constants.REALM_AGENT)) {
                    String msg = Constants.REALM_AGENT + " cannot be used as a valid attribute";
                    LOGGER.error(msg);
                    throw new RuntimeException(msg);
                }
                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, validUserNames);

            statusUpdater.updateStatus(IAMOperations.ADD_USER_ATTRIBUTE.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_USER_ATTRIBUTE.name(), OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            String msg = " Add attributes   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteUserAttributes(DeleteUserAttributeRequest request) {
        try {
            LOGGER.debug("Request received to delete user attributes " + request.getTenantId());

            List<String> validUserNames = new ArrayList<>();
            for (String username : request.getUsersList()) {
                if (keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), username, request.getAccessToken())) {
                    validUserNames.add(username);
                }
            }

            List<UserAttribute> attributes = request.getAttributesList();

            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            keycloakClient.deleteUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, validUserNames);
            statusUpdater.updateStatus(IAMOperations.DELETE_USER_ATTRIBUTES.name(), OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.DELETE_USER_ATTRIBUTES.name(), OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            String msg = " Add attributes   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus configureEventPersistence(EventPersistenceRequest request) {
        try {
            LOGGER.debug("Request received to configureEventPersistence " + request.getTenantId());

            keycloakClient.configureEventPersistence(
                    String.valueOf(request.getTenantId()),
                    request.getEvent(),
                    request.getPersistenceTime(),
                    request.getEnable(),
                    request.getAdminEvent()
            );
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.CONFIGURE_PERSISTANCE.name(), OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            String msg = " Configure Event Persistence   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public GroupsResponse createGroups(GroupsRequest request) {
        try {
            LOGGER.debug("Request received to createGroup " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();

            List<org.keycloak.representations.idm.GroupRepresentation> groupRepresentations =
                    transformToKeycloakGroups(request.getClientId(), request.getGroupsList());

            List<org.keycloak.representations.idm.GroupRepresentation> representations =
                    keycloakClient.createGroups(String.valueOf(tenantId), request.getClientId(), request.getClientSec(), groupRepresentations);

            List<GroupRepresentation> groups = transformKeycloakGroupsToGroups(request.getClientId(), representations);

            statusUpdater.updateStatus(IAMOperations.CREATE_GROUP.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            return GroupsResponse.newBuilder().addAllGroups(groups).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.CREATE_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Create Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public GroupRepresentation updateGroup(GroupRequest request) {
        try {
            LOGGER.debug("Request received to updateGroup " + request.getTenantId());

            long tenantId = request.getTenantId();

            List<GroupRepresentation> representations = new ArrayList<>();
            representations.add(request.getGroup());
            List<org.keycloak.representations.idm.GroupRepresentation> groupRepresentations = transformToKeycloakGroups(request.getClientId(), representations);

            org.keycloak.representations.idm.GroupRepresentation groupRepresentation = keycloakClient.updateGroup(String.valueOf(tenantId),
                    request.getClientId(), request.getClientSec(), groupRepresentations.get(0));

            statusUpdater.updateStatus(IAMOperations.UPDATE_GROUP.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());

            return transformKeycloakGroupToGroup(request.getClientId(), groupRepresentation, null);

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.UPDATE_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Update Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteGroup(GroupRequest request) {
        try {
            LOGGER.debug("Request received to deleteGroup " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();

            keycloakClient.deleteGroup(String.valueOf(tenantId), request.getClientId(), request.getClientSec(), request.getGroup().getId());
            statusUpdater.updateStatus(IAMOperations.DELETE_GROUP.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(),
                    request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.DELETE_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Delete Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public GroupRepresentation findGroup(GroupRequest request) {
        try {
            LOGGER.debug("Request received to findGroup " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();

            org.keycloak.representations.idm.GroupRepresentation groupRepresentation = keycloakClient.findGroup(String.valueOf(tenantId),
                    accessToken, request.getGroup().getId(), request.getGroup().getName());

            return groupRepresentation != null
                    ? transformKeycloakGroupToGroup(request.getClientId(), groupRepresentation, null)
                    : null;

        } catch (Exception ex) {
            String msg = " find Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public GroupsResponse getAllGroups(GroupRequest request) {
        try {
            LOGGER.debug("Request received to getAllGroups " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();
            List<org.keycloak.representations.idm.GroupRepresentation> groupRepresentation = keycloakClient.getAllGroups(String.valueOf(tenantId), accessToken);
            List<GroupRepresentation> groups = transformKeycloakGroupsToGroups(request.getClientId(), groupRepresentation);

            return !groups.isEmpty()
                    ? GroupsResponse.newBuilder().addAllGroups(groups).build()
                    : GroupsResponse.newBuilder().build();

        } catch (Exception ex) {
            String msg = " Get Groups   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus addUserToGroup(UserGroupMappingRequest request) {
        try {
            LOGGER.debug("Request received to getAllGroups " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();

            boolean status = keycloakClient.addUserToGroup(String.valueOf(tenantId), request.getUsername(), request.getGroupId(), accessToken);
            statusUpdater.updateStatus(IAMOperations.ADD_USER_TO_GROUP.name(),
                    status ? OperationStatus.SUCCESS : OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.ADD_USER_TO_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = "  Groups   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus removeUserFromGroup(UserGroupMappingRequest request) {
        try {
            LOGGER.debug("Request received to getAllGroups " + request.getTenantId());

            long tenantId = request.getTenantId();
            String accessToken = request.getAccessToken();

            boolean status = keycloakClient.removeUserFromGroup(String.valueOf(tenantId), request.getUsername(), request.getGroupId(), accessToken);
            statusUpdater.updateStatus(IAMOperations.REMOVE_USER_FROM_GROUP.name(),
                    status ? OperationStatus.SUCCESS : OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.REMOVE_USER_FROM_GROUP.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());

            String msg = "  Remove user from Group   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public SetUpTenantResponse createAgentClient(AgentClientMetadata request) {
        try {
            LOGGER.debug("Request received to configureAgentClient " + request.getTenantId());

            KeycloakClientSecret secret = keycloakClient.configureClient(
                    String.valueOf(request.getTenantId()),
                    request.getClientName(),
                    request.getTenantURL(),
                    request.getRedirectURIsList());

            if (secret != null) {
                statusUpdater.updateStatus(IAMOperations.CONFIGURE_AGENT_CLIENT.name(), OperationStatus.SUCCESS,
                        request.getTenantId(), request.getPerformedBy());

                return SetUpTenantResponse.newBuilder()
                        .setClientId(secret.getClientId())
                        .setClientSecret(secret.getClientSecret())
                        .build();
            } else {
                String msg = " Configure agent client  failed for " + request.getTenantId();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }
        } catch (Exception ex) {
            statusUpdater.updateStatus(IAMOperations.CONFIGURE_AGENT_CLIENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(),
                    request.getPerformedBy());
            String msg = " Configure agent client  failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus configureAgentClient(AgentClientMetadata request) {
        try {
            LOGGER.debug("Request received to configureAgentClient " + request.getTenantId());

            boolean status = keycloakClient.configureAgentClient(String.valueOf(request.getTenantId()), request.getClientName(),
                    request.getAccessTokenLifeTime());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = " Register and configureAgentClient user   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.REGISTER_AGENT.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public RegisterUserResponse registerAndEnableAgent(RegisterUserRequest request) {
        try {
            LOGGER.debug("Request received to registerAndEnableAgent " + request.getTenantId());

            boolean status = keycloakClient.createUser(String.valueOf(request.getTenantId()),
                    request.getUser().getId(),
                    request.getUser().getPassword(),
                    null,
                    null,
                    null,
                    false,
                    request.getAccessToken());

            if (status) {
                List<String> userList = new ArrayList<>();
                userList.add(request.getUser().getId());
                if (!request.getUser().getRealmRolesList().isEmpty()) {
                    keycloakClient.addRolesToUsers(request.getAccessToken(),
                            String.valueOf(request.getTenantId()), userList, request.getUser().getRealmRolesList(),
                            request.getClientId(), false);
                }
                if (!request.getUser().getClientRolesList().isEmpty()) {
                    keycloakClient.addRolesToUsers(request.getAccessToken(),
                            String.valueOf(request.getTenantId()), userList, request.getUser().getClientRolesList(),
                            request.getClientId(), true);
                }

                if (!request.getUser().getAttributesList().isEmpty()) {
                    Map<String, List<String>> map = new HashMap<>();
                    for (UserAttribute attribute : request.getUser().getAttributesList()) {
                        map.put(attribute.getKey(), attribute.getValuesList());
                    }
                    keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), map, userList);
                }

                status = keycloakClient.enableUserAccount(String.valueOf(request.getTenantId()), request.getAccessToken(), request.getUser().getId());
                statusUpdater.updateStatus(IAMOperations.REGISTER_AGENT.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());

                return RegisterUserResponse.newBuilder().setIsRegistered(status).build();

            } else {
                String msg = " Register and enable user   failed for  user " + request.getUser().getId() + "of tenant"
                        + request.getTenantId();
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }


        } catch (Exception ex) {
            String msg = " Register and enable user   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.REGISTER_AGENT.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteAgent(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to deleteAgent " + request.getTenantId());

            boolean status = keycloakClient.deleteUser(request.getAccessToken(), String.valueOf(request.getTenantId()), request.getUser().getId());
            statusUpdater.updateStatus(IAMOperations.DELETE_AGENT.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = " Delete agent  failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_AGENT.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus disableAgent(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to disableAgent " + request.getTenantId());
            boolean status = keycloakClient.disableUserAccount(String.valueOf(request.getTenantId()), request.getAccessToken(), request.getUser().getId());

            statusUpdater.updateStatus(IAMOperations.DISABLE_AGENT.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = " Disable agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_USER.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus isAgentNameAvailable(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to isAgentNameAvailable " + request.getTenantId());
            boolean status = keycloakClient.isUsernameAvailable(String.valueOf(request.getTenantId()), request.getUser().getId(), request.getAccessToken());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = " Is agent name available   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus addAgentAttributes(AddUserAttributesRequest request) {
        try {
            LOGGER.debug("Request received to addAgentAttributes " + request.getTenantId());

            List<UserAttribute> attributes = request.getAttributesList();
            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            boolean status = keycloakClient.addUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, request.getAgentsList());

            statusUpdater.updateStatus(IAMOperations.ADD_AGENT_ATTRIBUTES.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = " Add agent attributes   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.ADD_AGENT_ATTRIBUTES.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteAgentAttributes(DeleteUserAttributeRequest request) {
        try {
            LOGGER.debug("Request received to deleteAgentAttributes " + request.getTenantId());
            List<UserAttribute> attributes = request.getAttributesList();

            Map<String, List<String>> attributeMap = new HashMap<>();
            for (UserAttribute attribute : attributes) {
                attributeMap.put(attribute.getKey(), attribute.getValuesList());
            }

            boolean status = keycloakClient.deleteUserAttributes(String.valueOf(request.getTenantId()), request.getAccessToken(), attributeMap, request.getAgentsList());
            statusUpdater.updateStatus(IAMOperations.DELETE_AGENT_ATTRIBUTES.name(), OperationStatus.SUCCESS, request.getTenantId(), request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = " Delete agent attributes   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_AGENT_ATTRIBUTES.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus addRolesToAgent(AddUserRolesRequest request) {
        try {
            LOGGER.debug("Request received to addRolesToAgent " + request.getTenantId());
            boolean status = keycloakClient.addRolesToUsers(request.getAccessToken(), String.valueOf(request.getTenantId()),
                    request.getAgentsList(), request.getRolesList(), request.getClientId(), request.getClientLevel());
            statusUpdater.updateStatus(IAMOperations.ADD_ROLE_TO_AGENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = " Add roles to agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.ADD_ROLE_TO_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus deleteAgentRoles(DeleteUserRolesRequest request) {
        try {
            LOGGER.debug("Request received to deleteRolesFromAgent " + request.getTenantId());

            if (!request.getRolesList().isEmpty()) {
                keycloakClient.removeRoleFromUser(request.getAccessToken(),
                        String.valueOf(request.getTenantId()), request.getId(), request.getRolesList(), request.getClientId(), false);
            }

            if (!request.getClientRolesList().isEmpty()) {
                keycloakClient.removeRoleFromUser(request.getAccessToken(), String.valueOf(request.getTenantId()), request.getId(),
                        request.getClientRolesList(), request.getClientId(), true);

            }
            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_AGENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(true).build();

        } catch (Exception ex) {
            String msg = " Delete roles from agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_AGENT.name(), OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public Agent getAgent(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to getAgent " + request.getTenantId());

            UserRepresentation representation = keycloakClient.getUser(String.valueOf(request.getTenantId()), request.getAccessToken(), request.getUser().getId());

            if (representation != null) {
                if (representation.getAttributes() == null || representation.getAttributes().isEmpty() ||
                        representation.getAttributes().get(Constants.REALM_AGENT).get(0) == null ||
                        !representation.getAttributes().get(Constants.REALM_AGENT).get(0).equals("true")) {
                    throw new EntityNotFoundException("Agent not found for User search request, user Id: " + request.getUser().getId());

                } else {
                    return getAgent(representation);
                }
            } else {
                throw new EntityNotFoundException("Agent not found for User search request, user Id: " + request.getUser().getId());
            }


        } catch (Exception ex) {
            String msg = " Delete roles from agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.DELETE_ROLE_FROM_AGENT.name(),
                    OperationStatus.FAILED,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus enableAgent(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to enableAgent " + request.getTenantId());
            boolean status = keycloakClient.enableUserAccount(String.valueOf(request.getTenantId()), request.getAccessToken(), request.getUser().getId());

            statusUpdater.updateStatus(IAMOperations.ENABLE_AGENT.name(), OperationStatus.FAILED, request.getTenantId(), request.getPerformedBy());

            return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = " Enable agent   failed for " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.ENABLE_AGENT.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus grantAdminPrivilege(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to grantAdminPrivilege " + request.getTenantId());

            boolean validationStatus = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), request.getUser().getUsername(), request.getAccessToken());

            if (validationStatus) {
                boolean status = keycloakClient.grantAdminPrivilege(String.valueOf(request.getTenantId()), request.getUser().getUsername());
                statusUpdater.updateStatus(IAMOperations.GRANT_ADMIN_PRIVILEGE.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), request.getPerformedBy());

                return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();
            } else {
                String msg = " Not a valid user: " + request.getUser().getId();
                LOGGER.error(msg);
                throw new EntityNotFoundException(msg);
            }

        } catch (Exception ex) {
            String msg = " Grant admin privilege " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.GRANT_ADMIN_PRIVILEGE.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public org.apache.custos.core.iam.api.OperationStatus removeAdminPrivilege(UserSearchRequest request) {
        try {
            LOGGER.debug("Request received to removeAdminPrivilege " + request.getTenantId());

            boolean validationStatus = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                    request.getUser().getUsername(), request.getAccessToken());

            if (validationStatus) {
                boolean status = keycloakClient.removeAdminPrivilege(String.valueOf(request.getTenantId()), request.getUser().getUsername());
                statusUpdater.updateStatus(IAMOperations.REMOVE_ADMIN_PRIVILEGE.name(),
                        OperationStatus.FAILED,
                        request.getTenantId(), request.getPerformedBy());

                return org.apache.custos.core.iam.api.OperationStatus.newBuilder().setStatus(status).build();

            } else {
                String msg = " Not a valid user: " + request.getUser().getId();
                LOGGER.error(msg);
                throw new EntityNotFoundException(msg);
            }

        } catch (Exception ex) {
            String msg = " Remove admin privilege " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            statusUpdater.updateStatus(IAMOperations.REMOVE_ADMIN_PRIVILEGE.name(),
                    OperationStatus.SUCCESS,
                    request.getTenantId(), request.getPerformedBy());
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public GetAllResourcesResponse getAllResources(GetAllResources request) {
        try {
            LOGGER.debug("Request received to getAllResources for tenant " + request.getTenantId());

            List<UserRepresentation> representations = keycloakClient.getAllUsers(String.valueOf(request.getTenantId()));
            GetAllResourcesResponse resourcesResponse = GetAllResourcesResponse.newBuilder().build();
            if (!representations.isEmpty()) {
                if (request.getResourceType().name().equals(ResourceTypes.USER.name())) {
                    List<org.apache.custos.core.iam.api.UserRepresentation> users = new ArrayList<>();
                    for (UserRepresentation userRepresentation : representations) {
                        boolean validationStatus = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()), userRepresentation.getUsername());
                        if (validationStatus) {
                            users.add(getUser(userRepresentation, request.getClientId()));
                        }
                    }

                    return resourcesResponse.toBuilder().addAllUsers(users).build();

                } else {
                    List<Agent> agents = new ArrayList<>();
                    for (UserRepresentation userRepresentation : representations) {
                        boolean validationStatus = keycloakClient.isValidEndUser(String.valueOf(request.getTenantId()),
                                userRepresentation.getUsername());
                        if (!validationStatus) {
                            agents.add(getAgent(userRepresentation));
                        }
                    }
                    return resourcesResponse.toBuilder().addAllAgents(agents).build();
                }

            } else {
                String msg = " Empty resources";
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }


        } catch (Exception ex) {
            String msg = " Get all resources failed";
            LOGGER.error(msg, ex);
            if (ex.getMessage().contains("HTTP 401 Unauthorized")) {
                throw new RuntimeException("HTTP 401 Unauthorized", ex);
            } else {
                throw new RuntimeException(msg, ex);
            }
        }
    }

    public String getIamServerURL() {
        return iamServerURL;
    }

    public Map<String, Object> getUserInfo(String accessToken, long tenantId) throws ParseException {
        return keycloakClient.getUserInfo(tokenService.getKCToken(accessToken), tenantId);
    }

    public Map<String, String> configureClient(long tenantId, String clientId, String tenantUrl, List<String> redirectUris) {
        KeycloakClientSecret clientSecret = keycloakClient.configureClient(String.valueOf(tenantId),
                clientId, tenantUrl, redirectUris);
        return Map.of("clientId", clientSecret.getClientId(), "clientSecret", clientSecret.getClientSecret());
    }

    private org.apache.custos.core.iam.api.UserRepresentation getUser(UserRepresentation representation, String clientId) {
        String state = Status.PENDING_CONFIRMATION;

        if (representation.isEnabled()) {
            state = Status.ACTIVE;
        } else if (representation.isEmailVerified()) {
            state = Status.CONFIRMED;
        }

        Map<String, List<String>> attributes = representation.getAttributes();

        List<UserAttribute> attributeList = new ArrayList<>();

        if (attributes != null && !attributes.isEmpty()) {
            attributeList = attributes.entrySet().stream()
                    .map(entry -> UserAttribute.newBuilder()
                            .setKey(entry.getKey())
                            .addAllValues(entry.getValue())
                            .build())
                    .collect(Collectors.toList());
        }

        org.apache.custos.core.iam.api.UserRepresentation.Builder builder = org.apache.custos.core.iam.api.UserRepresentation.newBuilder()
                .setUsername(representation.getUsername())
                .setFirstName(representation.getFirstName())
                .setLastName(representation.getLastName())
                .setState(state)
                .setCreationTime(representation.getCreatedTimestamp())
                .setEmail(representation.getEmail());

        if (representation.getAttributes() != null && !representation.getAttributes().isEmpty()) {
            builder.addAllAttributes(attributeList);
        }

        if (representation.getRealmRoles() != null && !representation.getRealmRoles().isEmpty()) {
            builder.addAllRealmRoles(representation.getRealmRoles());
        }

        if (representation.getClientRoles() != null && representation.getClientRoles().get(clientId) != null &&
                !representation.getClientRoles().get(clientId).isEmpty()) {
            builder.addAllClientRoles(representation.getClientRoles().get(clientId));
        }

        return builder.build();
    }

    private OperationMetadata convertFromEntity(StatusEntity entity) {
        return OperationMetadata.newBuilder()
                .setEvent(entity.getEvent())
                .setStatus(entity.getState())
                .setPerformedBy(entity.getPerformedBy())
                .setTimeStamp(entity.getTime().toString()).build();
    }

    private List<org.keycloak.representations.idm.GroupRepresentation> transformToKeycloakGroups(String clientId, List<GroupRepresentation> groupRepresentations) {
        List<org.keycloak.representations.idm.GroupRepresentation> groupsList = new ArrayList<>();
        for (GroupRepresentation groupRepresentation : groupRepresentations) {
            groupsList.add(transformSingleGroupToKeycloakGroup(clientId, groupRepresentation, null));
        }
        return groupsList;
    }

    private org.keycloak.representations.idm.GroupRepresentation transformSingleGroupToKeycloakGroup(String clientId, GroupRepresentation groupRepresentation,
                                                                                                     org.keycloak.representations.idm.GroupRepresentation parentGroup) {
        String name = groupRepresentation.getName();
        String id = groupRepresentation.getId();

        if (StringUtils.isNotBlank(groupRepresentation.getOwnerId())) {
            groupRepresentation = groupRepresentation.toBuilder().addAttributes(UserAttribute.newBuilder()
                    .setKey(Constants.OWNER_ID).addValues(groupRepresentation.getOwnerId()).build()).build();
        }

        if (StringUtils.isNotBlank(groupRepresentation.getDescription())) {
            groupRepresentation = groupRepresentation.toBuilder().addAttributes(UserAttribute.newBuilder()
                    .setKey(Constants.DESCRIPTION).addValues(groupRepresentation.getDescription()).build()).build();
        }

        List<UserAttribute> attributeList = groupRepresentation.getAttributesList();
        List<String> realmRoles = groupRepresentation.getRealmRolesList();
        List<String> clientRoles = groupRepresentation.getClientRolesList();

        org.keycloak.representations.idm.GroupRepresentation keycloakGroup = new org.keycloak.representations.idm.GroupRepresentation();

        keycloakGroup.setName(name);

        if (StringUtils.isNotBlank(id)) {
            keycloakGroup.setId(id);
        }

        Map<String, List<String>> map = new HashMap<>();
        if (!attributeList.isEmpty()) {
            for (UserAttribute attribute : attributeList) {
                map.put(attribute.getKey(), attribute.getValuesList());
            }
            keycloakGroup.setAttributes(map);
        }

        if (!realmRoles.isEmpty()) {
            keycloakGroup.setRealmRoles(realmRoles);
        }

        Map<String, List<String>> clientMap = new HashMap<>();
        if (!clientRoles.isEmpty()) {
            clientMap.put(clientId, clientRoles);
            keycloakGroup.setClientRoles(clientMap);
        }

        if (!groupRepresentation.getSubGroupsList().isEmpty()) {
            for (GroupRepresentation representation : groupRepresentation.getSubGroupsList()) {
                transformSingleGroupToKeycloakGroup(clientId, representation, keycloakGroup);
            }
        }

        if (parentGroup != null) {
            List<org.keycloak.representations.idm.GroupRepresentation> groupRepList = parentGroup.getSubGroups();
            if (groupRepList == null) {
                groupRepList = new ArrayList<>();
            }
            String path = parentGroup.getPath() + "/" + keycloakGroup.getName();
            keycloakGroup.setPath(path);
            groupRepList.add(keycloakGroup);
            parentGroup.setSubGroups(groupRepList);
            return parentGroup;
        }

        String path = "/" + keycloakGroup.getName();
        keycloakGroup.setPath(path);

        return keycloakGroup;
    }

    private List<GroupRepresentation> transformKeycloakGroupsToGroups(String clientId,
                                                                      List<org.keycloak.representations.idm.GroupRepresentation> groupRepresentations) {
        List<GroupRepresentation> groupsList = new ArrayList<>();
        for (org.keycloak.representations.idm.GroupRepresentation groupRepresentation : groupRepresentations) {
            groupsList.add(transformKeycloakGroupToGroup(clientId, groupRepresentation, null));
        }
        return groupsList;
    }

    private GroupRepresentation transformKeycloakGroupToGroup(String clientId, org.keycloak.representations.idm.GroupRepresentation group,
                                                              GroupRepresentation parent) {
        String name = group.getName();
        String id = group.getId();
        List<String> realmRoles = group.getRealmRoles();
        Map<String, List<String>> clientRoles = group.getClientRoles();
        Map<String, List<String>> atrs = group.getAttributes();

        GroupRepresentation representation = GroupRepresentation.newBuilder()
                .setName(name)
                .setId(id)
                .build();

        if (realmRoles != null && !realmRoles.isEmpty()) {
            representation = representation.toBuilder().addAllRealmRoles(realmRoles).build();
        }

        if (clientRoles != null && !clientRoles.isEmpty() && !clientRoles.get(clientId).isEmpty()) {
            representation = representation.toBuilder().addAllClientRoles(clientRoles.get(clientId)).build();
        }

        if (atrs != null && !atrs.isEmpty()) {
            List<UserAttribute> attributeList = new ArrayList<>();
            for (String key : atrs.keySet()) {
                if (key.equals(Constants.OWNER_ID)) {
                    representation = representation.toBuilder().setOwnerId(atrs.get(key).get(0)).build();
                } else if (key.equals(Constants.DESCRIPTION)) {
                    representation = representation.toBuilder().setDescription(atrs.get(key).get(0)).build();
                } else {
                    UserAttribute attribute = UserAttribute.newBuilder().setKey(key).addAllValues(atrs.get(key)).build();
                    attributeList.add(attribute);
                }
            }
            representation = representation.toBuilder().addAllAttributes(attributeList).build();
        }

        if (group.getSubGroups() != null && !group.getSubGroups().isEmpty()) {
            for (org.keycloak.representations.idm.GroupRepresentation subGroup : group.getSubGroups()) {
                representation = transformKeycloakGroupToGroup(clientId, subGroup, representation);
            }
        }

        if (parent != null) {
            parent = parent.toBuilder().addSubGroups(representation).build();
        }

        if (parent != null) {
            return parent;
        } else {
            return representation;
        }
    }

    private Agent getAgent(UserRepresentation representation) {

        Agent.Builder builder = Agent.newBuilder().setId(representation.getUsername())
                .setIsEnabled(representation.isEnabled())
                .setCreationTime(representation.getCreatedTimestamp());

        for (String key : representation.getAttributes().keySet()) {
            UserAttribute attribute = UserAttribute.newBuilder().setKey(key).addAllValues(representation.getAttributes().get(key)).build();
            builder.addAttributes(attribute);
        }

        if (representation.getRealmRoles() != null && !representation.getRealmRoles().isEmpty()) {
            builder.addAllRealmRoles(representation.getRealmRoles());
        }

        if (representation.getClientRoles() != null && !representation.getClientRoles().isEmpty() &&
                representation.getClientRoles().get(Constants.AGENT_CLIENT) != null &&
                !representation.getClientRoles().get(Constants.AGENT_CLIENT).isEmpty()) {
            builder.addAllClientRoles(representation.getClientRoles().get(Constants.AGENT_CLIENT));
        }

        return builder.build();
    }
}
