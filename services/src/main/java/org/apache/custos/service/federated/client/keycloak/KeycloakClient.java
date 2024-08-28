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

package org.apache.custos.service.federated.client.keycloak;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.custos.core.exception.UnauthorizedException;
import org.apache.custos.core.constants.Constants;
import org.apache.http.HttpStatus;
import org.keycloak.admin.client.Keycloak;
import org.keycloak.admin.client.resource.GroupResource;
import org.keycloak.admin.client.resource.ProtocolMappersResource;
import org.keycloak.admin.client.resource.RealmResource;
import org.keycloak.admin.client.resource.RoleMappingResource;
import org.keycloak.admin.client.resource.RoleResource;
import org.keycloak.admin.client.resource.UserResource;
import org.keycloak.representations.idm.ClientRepresentation;
import org.keycloak.representations.idm.CredentialRepresentation;
import org.keycloak.representations.idm.EventRepresentation;
import org.keycloak.representations.idm.FederatedIdentityRepresentation;
import org.keycloak.representations.idm.GroupRepresentation;
import org.keycloak.representations.idm.IdentityProviderRepresentation;
import org.keycloak.representations.idm.MappingsRepresentation;
import org.keycloak.representations.idm.ProtocolMapperRepresentation;
import org.keycloak.representations.idm.RealmEventsConfigRepresentation;
import org.keycloak.representations.idm.RealmRepresentation;
import org.keycloak.representations.idm.RoleRepresentation;
import org.keycloak.representations.idm.RolesRepresentation;
import org.keycloak.representations.idm.UserRepresentation;
import org.keycloak.representations.idm.UserSessionRepresentation;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import javax.validation.constraints.NotNull;
import javax.ws.rs.NotFoundException;
import javax.ws.rs.core.Response;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * This class acts as a rest client for keycloak server
 */
@Component
public class KeycloakClient {
    private static final Logger LOGGER = LoggerFactory.getLogger(KeycloakClient.class);

    private static final int ACCESS_TOKEN_LIFE_SPAN = 1800;
    private static final int SESSION_IDLE_TIMEOUT = 3600;

    private final ObjectMapper objectMapper;

    @Value("${iam.server.client.id:admin-cli}")
    private String clientId;

    @Value("${iam.server.url}")
    private String iamServerURL;

    @Value("${iam.server.admin.username}")
    private String superAdminUserName;

    @Value("${iam.server.admin.password}")
    private String superAdminPassword;

    @Value("${iam.server.super.admin.realm.id:master}")
    private String superAdminRealmID;

    @Value("${iam.federated.cilogon.authorization.endpoint:https://cilogon.org/authorize}")
    private String ciLogonAuthorizationEndpoint;

    @Value("${iam.federated.cilogon.token.endpoint:https://cilogon.org/oauth2/token}")
    private String ciLogonTokenEndpoint;

    @Value("${iam.federated.cilogon.token.userinfo.endpoint:https://cilogon.org/oauth2/userinfo}")
    private String ciLogonUserInfoEndpoint;

    @Value("${iam.federated.cilogon.issuer:https://cilogon.org}")
    private String ciLogonIssuerUri;

    @Value("${iam.federated.cilogon.jwksUri:https://cilogon.org/oauth2/certs}")
    private String jwksUri;

    public KeycloakClient(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    public void createRealm(String realmId, String displayName) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            // create realm
            RealmRepresentation newRealmDetails = new RealmRepresentation();
            newRealmDetails.setEnabled(true);
            newRealmDetails.setId(realmId);
            newRealmDetails.setDisplayName(displayName);
            newRealmDetails.setRealm(realmId);
            // Following two settings allow duplicate email addresses
            newRealmDetails.setLoginWithEmailAllowed(false);
            newRealmDetails.setDuplicateEmailsAllowed(true);
            // Default access token lifespan to 30 minutes, SSO session idle to 60 minutes
            newRealmDetails.setAccessTokenLifespan(ACCESS_TOKEN_LIFE_SPAN);
            newRealmDetails.setSsoSessionIdleTimeout(SESSION_IDLE_TIMEOUT);
            RealmRepresentation realmWithRoles = createDefaultRoles(newRealmDetails);
            client.realms().create(realmWithRoles);

        } catch (Exception ex) {
            String msg = "Error creating Realm in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public void updateRealm(String realmId, String displayName) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            // create realm
            RealmResource realmResource = client.realm(realmId);

            if (realmResource != null) {
                RealmRepresentation newRealmDetails = realmResource.toRepresentation();
                newRealmDetails.setId(realmId);
                newRealmDetails.setDisplayName(displayName);
                newRealmDetails.setRealm(realmId);
                realmResource.update(newRealmDetails);
            } else {
                LOGGER.error("Realm not found");
                throw new RuntimeException("Realm not found");
            }

        } catch (Exception ex) {
            String msg = "Error creating Realm in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public boolean createRealmAdminAccount(String realmId, String adminUsername, String adminFirstname,
                                           String adminLastname, String adminEmail, String adminPassword) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            UserRepresentation user = new UserRepresentation();
            user.setUsername(adminUsername);
            user.setFirstName(adminFirstname);
            user.setLastName(adminLastname);
            user.setEmail(adminEmail);
            user.setEmailVerified(true);
            user.setEnabled(true);
            Response httpResponse = client.realm(realmId).users().create(user);
            LOGGER.debug("Realm admin account creation exited with code : " + httpResponse.getStatus() + " : " + httpResponse.getStatusInfo());
            if (httpResponse.getStatus() == HttpStatus.SC_CREATED) { //HTTP code for record creation: HTTP 201
                List<UserRepresentation> retrieveCreatedUserList = client.realm(realmId).users().search(user.getUsername(),
                        user.getFirstName(),
                        user.getLastName(),
                        user.getEmail(),
                        0, 1);
                UserResource retrievedUser = client.realm(realmId).users().get(retrieveCreatedUserList.get(0).getId());

                // Add user to the "admin" role
                RoleResource adminRoleResource = client.realm(realmId).roles().get("admin");
                retrievedUser.roles().realmLevel().add(Collections.singletonList(adminRoleResource.toRepresentation()));

                CredentialRepresentation credential = new CredentialRepresentation();
                credential.setType(CredentialRepresentation.PASSWORD);
                credential.setValue(adminPassword);
                credential.setTemporary(false);
                retrievedUser.resetPassword(credential);
                List<ClientRepresentation> realmClients = client.realm(realmId).clients().findAll();
                String realmManagementClientId = getRealmManagementClientId(client, realmId);
                for (ClientRepresentation realmClient : realmClients) {
                    if (realmClient.getClientId().equals("realm-management")) {
                        realmManagementClientId = realmClient.getId();
                    }
                }
                retrievedUser.roles().clientLevel(realmManagementClientId).add(retrievedUser.roles().clientLevel(realmManagementClientId).listAvailable());
                return true;
            } else {
                LOGGER.error("Request for Tenant Admin Account Creation failed with HTTP code : " + httpResponse.getStatus());
                LOGGER.error("Reason for Tenant Admin account creation failure : " + httpResponse.getStatusInfo());
                throw new RuntimeException("Reason for Tenant Admin account creation failure : " + httpResponse.getStatusInfo(), null);
            }
        } catch (Exception ex) {
            String msg = "Error creating Realm Admin Account in keycloak server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean updateRealmAdminAccount(String realmId, String adminUsername, String adminFirstname,
                                           String adminLastname, String adminEmail, String adminPassword) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            UserRepresentation representation = getUserByUsername(client, realmId, adminUsername);
            if (representation != null) {
                representation.setUsername(adminUsername);
                representation.setFirstName(adminFirstname);
                representation.setLastName(adminLastname);
                representation.setEmail(adminEmail);
                representation.setEmailVerified(true);
                representation.setEnabled(true);
                client.realm(realmId).users().get(representation.getId()).update(representation);
                return true;
            } else {
                return createRealmAdminAccount(realmId, adminUsername, adminFirstname, adminLastname, adminEmail, adminPassword);
            }
        } catch (Exception ex) {
            String msg = "Error updating Realm Admin Account in keycloak server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean grantAdminPrivilege(String realmId, String username) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            UserRepresentation representation = getUserByUsername(client, realmId, username);
            if (representation != null) {
                UserResource retrievedUser = client.realm(realmId).users().get(representation.getId());
                RoleResource adminRoleResource = client.realm(realmId).roles().get("admin");
                retrievedUser.roles().realmLevel().add(Collections.singletonList(adminRoleResource.toRepresentation()));

                String realmManagementClientId = getRealmManagementClientId(client, realmId);

                retrievedUser.roles().clientLevel(realmManagementClientId).
                        add(retrievedUser.roles().clientLevel(realmManagementClientId).listAvailable());
                return true;

            } else {
                LOGGER.error("Cannot find existing user with username: " + username);
                throw new RuntimeException("Cannot find existing user with username: " + username);
            }

        } catch (Exception ex) {
            String msg = "Error granting admin privilege, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public boolean removeAdminPrivilege(String realmId, String username) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            UserRepresentation representation = getUserByUsername(client, realmId, username);
            if (representation != null) {

                UserResource retrievedUser = client.realm(realmId).users().get(representation.getId());
                RoleResource adminRoleResource = client.realm(realmId).roles().get("admin");
                retrievedUser.roles().realmLevel().remove(Collections.singletonList(adminRoleResource.toRepresentation()));
                String realmManagementClientId = getRealmManagementClientId(client, realmId);

                retrievedUser.roles().clientLevel(realmManagementClientId).
                        remove(retrievedUser.roles().clientLevel(realmManagementClientId).listEffective());
                return true;

            } else {
                LOGGER.error("Cannot find existing user with username: " + username);
                throw new RuntimeException("Cannot find existing user with username: " + username);
            }
        } catch (Exception ex) {
            String msg = "Error removing admin privilege, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public KeycloakClientSecret configureClient(String realmId, String clientName, @NotNull String tenantURL, List<String> redirectUris) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            ClientRepresentation pgaClient = new ClientRepresentation();
            pgaClient.setName(clientName);
            pgaClient.setClientId(clientName);
            pgaClient.setProtocol("openid-connect");
            pgaClient.setStandardFlowEnabled(true);
            pgaClient.setEnabled(true);
            pgaClient.setAuthorizationServicesEnabled(true);
            pgaClient.setDirectAccessGrantsEnabled(true);
            pgaClient.setServiceAccountsEnabled(true);
            pgaClient.setFullScopeAllowed(true);
            pgaClient.setClientAuthenticatorType("client-secret");

            pgaClient.setBaseUrl(tenantURL);

            // Remove trailing slash from URL
            if (tenantURL.endsWith("/")) {
                tenantURL = tenantURL.substring(0, tenantURL.length() - 1);
            }

            List<String> newList = new ArrayList<>(redirectUris);
            newList.add(tenantURL);

            pgaClient.setRedirectUris(newList);

            List<String> webOrigins = new ArrayList<>();
            webOrigins.add("+");
            pgaClient.setWebOrigins(webOrigins);

            pgaClient.setPublicClient(false);
            try (Response httpResponse = client.realms().realm(realmId).clients().create(pgaClient)) {
                LOGGER.debug("Realm client configuration exited with code : " + httpResponse.getStatus() + " : " + httpResponse.getStatusInfo());

                // Add the manage-users role to the web client
                UserRepresentation serviceAccountUserRepresentation =
                        getUserByUsername(client, realmId, "service-account-" + pgaClient.getClientId());
                UserResource serviceAccountUser = client.realms().realm(realmId).users().get(serviceAccountUserRepresentation.getId());
                String realmManagementClientId = getRealmManagementClientId(client, realmId);
                List<RoleRepresentation> manageUsersRole =
                        serviceAccountUser.roles().clientLevel(realmManagementClientId).listAvailable()
                                .stream()
                                .filter(r -> r.getName().equals("manage-users"))
                                .collect(Collectors.toList());
                serviceAccountUser.roles().clientLevel(realmManagementClientId).add(manageUsersRole);

                if (httpResponse.getStatus() == HttpStatus.SC_CREATED) {
                    String ClientUUID = client.realms().realm(realmId).clients().findByClientId(pgaClient.getClientId()).get(0).getId();
                    CredentialRepresentation clientSecret = client.realms().realm(realmId).clients().get(ClientUUID).getSecret();
                    return new KeycloakClientSecret(pgaClient.getClientId(), clientSecret.getValue());

                } else {
                    LOGGER.error("Request for realm client creation failed with HTTP code : " + httpResponse.getStatus());
                    LOGGER.error("Reason for realm client creation failure : " + httpResponse.getStatusInfo());
                    throw new RuntimeException("Reason for realm client creation failure :" + httpResponse.getStatusInfo(), null);
                }
            }
        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public KeycloakClientSecret updateClient(String realmId, String clientName, @NotNull String tenantURL, List<String> redirectUris) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            List<ClientRepresentation> clientRepresentations = client.realm(realmId).clients().findByClientId(clientName);

            if (clientRepresentations == null || clientRepresentations.isEmpty()) {
                LOGGER.error("Cannot find a client with name " + clientName);
                throw new RuntimeException("Cannot find a client with name " + clientName);
            }

            ClientRepresentation pgaClient = clientRepresentations.get(0);
            pgaClient.setBaseUrl(tenantURL);

            // Remove trailing slash from gatewayURL
            if (tenantURL.endsWith("/")) {
                tenantURL = tenantURL.substring(0, tenantURL.length() - 1);
            }

            List<String> newList = new ArrayList<>(redirectUris);
            newList.add(tenantURL);

            pgaClient.setRedirectUris(newList);
            pgaClient.setPublicClient(false);
            client.realms().realm(realmId).clients().get(pgaClient.getId()).update(pgaClient);

            String ClientUUID = client.realms().realm(realmId).clients().findByClientId(pgaClient.getClientId()).get(0).getId();
            CredentialRepresentation clientSecret = client.realms().realm(realmId).clients().get(ClientUUID).getSecret();
            return new KeycloakClientSecret(pgaClient.getClientId(), clientSecret.getValue());

        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean isUsernameAvailable(String realmId, String username, String accessToken) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            return userRepresentation == null;

        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean createUser(String realmId, String username, String newPassword, String firstName,
                              String lastName, String emailAddress, boolean tempPassowrd, String accessToken) throws UnauthorizedException {

        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation user = new UserRepresentation();
            user.setUsername(username);
            user.setFirstName(firstName);
            user.setLastName(lastName);
            user.setEmail(emailAddress);
            user.setEnabled(false);
            Response httpResponse = client.realm(realmId).users().create(user);

            if (httpResponse.getStatus() == HttpStatus.SC_CREATED) { //HTTP code for record creation: HTTP 201
                List<UserRepresentation> retrieveCreatedUserList = client.realm(realmId).users().search(user.getUsername(),
                        user.getFirstName(),
                        user.getLastName(),
                        user.getEmail(),
                        0, 1);
                UserResource retrievedUser = client.realm(realmId).users().get(retrieveCreatedUserList.get(0).getId());
                CredentialRepresentation credential = new CredentialRepresentation();
                credential.setType(CredentialRepresentation.PASSWORD);
                credential.setValue(newPassword);
                credential.setTemporary(tempPassowrd);
                retrievedUser.resetPassword(credential);
                return true;

            } else {
                String msg = "Reason for user account creation failure : " + httpResponse.getStatusInfo();
                LOGGER.error("Request for user Account Creation failed with HTTP code : " + httpResponse.getStatus());
                LOGGER.error(msg);
                throw new UnauthorizedException(msg, null);
            }
        }
    }


    public boolean enableUserAccount(String realmId, String accessToken, String username) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);

            UserResource userResource = client.realm(realmId).users().get(userRepresentation.getId());
            UserRepresentation profile = userResource.toRepresentation();
            profile.setEnabled(true);
            userResource.update(profile);
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred enableUserAccount, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean disableUserAccount(String realmId, String accessToken, String username) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);

            if (userRepresentation != null) {
                UserResource userResource = client.realm(realmId).users().get(userRepresentation.getId());
                UserRepresentation profile = userResource.toRepresentation();
                profile.setEnabled(false);
                userResource.update(profile);
            }
            return true;

        } catch (Exception ex) {
            String msg = "Error in disableUserAccount at keycloak, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public boolean isUserAccountEnabled(String realmId, String accessToken, String username) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            return userRepresentation != null && userRepresentation.isEnabled();

        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public boolean isUserExist(String realmId, String accessToken, String username) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            return userRepresentation != null;

        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public UserRepresentation getUser(String realmId, String accessToken, String username) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            return getUserByUsername(client, realmId, username);

        } catch (Exception ex) {
            String msg = "Error retrieving user, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public UserRepresentation getUser(String realmId, String username) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            return getUserByUsername(client, realmId, username);

        } catch (Exception ex) {
            String msg = "Error retrieving user, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public List<UserRepresentation> getUsers(String accessToken, String realmId, int offset, int limit,
                                             String username, String firstName, String lastName,
                                             String email, String search) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            return searchUsers(client, realmId, username, firstName, lastName, email, search, offset, limit);

        } catch (Exception ex) {
            String msg = "Error occurred while searching for user, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean resetUserPassword(String accessToken, String realmId, String username, String newPassword) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            if (userRepresentation != null) {
                UserResource retrievedUser = client.realm(realmId).users().get(userRepresentation.getId());
                CredentialRepresentation credential = new CredentialRepresentation();
                credential.setType(CredentialRepresentation.PASSWORD);
                credential.setValue(newPassword);
                credential.setTemporary(false);
                retrievedUser.resetPassword(credential);
                // Remove the UPDATE_PASSWORD required action
                userRepresentation = retrievedUser.toRepresentation();
                userRepresentation.getRequiredActions().remove("UPDATE_PASSWORD");
                retrievedUser.update(userRepresentation);
                return true;
            } else {
                String msg = "requested User not found";
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }
        } catch (Exception ex) {
            String msg = "Error resetting user password in keycloak server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public List<UserRepresentation> findUser(String accessToken, String realmId, String email, String userName) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            return client.realm(realmId).users().search(userName, null, null, email, 0, 1);
        } catch (Exception ex) {
            String msg = "Error finding user in keycloak server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public void updateUserRepresentation(String accessToken, String realmId, String username,
                                         String firstname, String lastName, String email) {

        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            if (userRepresentation != null) {
                userRepresentation.setFirstName(firstname);
                userRepresentation.setLastName(lastName);
                userRepresentation.setEmail(email);
                UserResource userResource = client.realm(realmId).users().get(userRepresentation.getId());

                userResource.update(userRepresentation);
            } else {
                throw new RuntimeException("User [" + username + "] wasn't found in Keycloak!");
            }
        } catch (Exception ex) {
            String msg = "Error updating user profile in keycloak server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean deleteUser(String accessToken, String realmId, String username) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            if (userRepresentation != null) {
                client.realm(realmId).users().delete(userRepresentation.getId());
                return true;
            } else {
                throw new RuntimeException("User [" + username + "] wasn't found in Keycloak!");
            }
        } catch (Exception ex) {
            String msg = "Error deleting user in keycloak server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean addRolesToUsers(String accessToken, String realmId, List<String> users,
                                   List<String> roles, String clientId, boolean clientLevel) {

        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            for (String username : users) {

                UserRepresentation representation = getUserByUsername(client, realmId, username.toLowerCase());
                ClientRepresentation clientRepresentation = client.realm(realmId).clients().findByClientId(clientId).get(0);
                if (representation != null) {
                    RealmResource realmResource = client.realm(realmId);
                    UserResource resource = client.realm(realmId).users().get(representation.getId());
                    List<RoleRepresentation> roleRepresentations = new ArrayList<>();
                    if (clientLevel) {
                        for (String role : roles) {
                            RoleResource roleResource = realmResource.clients().get(clientRepresentation.getId()).roles().get(role);
                            roleRepresentations.add(roleResource.toRepresentation());
                        }
                        resource.roles().clientLevel(clientRepresentation.getId()).add(roleRepresentations);

                    } else {

                        for (String role : roles) {
                            RoleResource roleResource = client.realm(realmId).roles().get(role);
                            roleRepresentations.add(roleResource.toRepresentation());
                        }
                        resource.roles().realmLevel().add(roleRepresentations);
                    }

                }
            }
            return true;
        } catch (Exception ex) {
            String msg = "Error while adding roles to user " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean removeRoleFromUser(String accessToken, String realmId, String username,
                                      List<String> roles, String clientId, boolean clientLevel) {

        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            UserRepresentation representation = getUserByUsername(client, realmId, username.toLowerCase());

            if (representation != null) {
                UserResource retrievedUser = client.realm(realmId).users().get(representation.getId());

                if (clientLevel) {
                    List<ClientRepresentation> clientRepresentationList =
                            client.realm(realmId).clients().findByClientId(clientId);

                    if (clientRepresentationList != null && !clientRepresentationList.isEmpty()) {
                        ClientRepresentation clientRep = clientRepresentationList.get(0);
                        List<RoleRepresentation> roleRepresentations = new ArrayList<>();
                        for (String roleName : roles) {
                            RoleResource roleResource = client.realm(realmId).
                                    clients().get(clientRep.getId()).roles().get(roleName);
                            if (roleResource != null) {
                                roleRepresentations.add(roleResource.toRepresentation());
                            }
                        }
                        if (!roleRepresentations.isEmpty()) {
                            retrievedUser.roles().clientLevel(clientRep.getId()).remove(roleRepresentations);
                        }


                    }
                } else {
                    List<RoleRepresentation> roleRepresentations = new ArrayList<>();
                    for (String roleName : roles) {
                        RoleResource roleResource = client.realm(realmId).roles().get(roleName);
                        if (roleResource != null) {
                            roleRepresentations.add(roleResource.toRepresentation());
                        }
                    }
                    if (!roleRepresentations.isEmpty()) {
                        retrievedUser.roles().realmLevel().remove(roleRepresentations);
                    }
                }
            }
            return true;
        } catch (Exception ex) {
            String msg = "Error removing roles from user , reason " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean deleteRealm(String realmId) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmResource realmResource = client.realm(realmId);

            if (realmResource != null) {
                realmResource.remove();
            }

        } catch (NotFoundException ex) {
            LOGGER.debug("Realm not found", ex);
        } catch (Exception ex) {
            String msg = "Error deleting Realm in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
        return true;

    }


    public boolean configureOIDCFederatedIDP(String realmId, String displayName, String scopes, KeycloakClientSecret secret, Map<String, String> configs) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmResource realmResource = client.realm(realmId);

            List<IdentityProviderRepresentation> representations = realmResource.identityProviders().findAll();

            for (IdentityProviderRepresentation representation : representations) {
                realmResource.identityProviders().get(representation.getInternalId()).remove();
            }

            IdentityProviderRepresentation idp = new IdentityProviderRepresentation();

            idp.setAlias("oidc");
            idp.setDisplayName(displayName);
            idp.setProviderId("oidc");
            idp.setEnabled(true);
            if (configs != null) {
                idp.setConfig(configs);
            }

            idp.getConfig().put("clientId", secret.getClientId());
            idp.getConfig().put("clientSecret", secret.getClientSecret());
            idp.getConfig().put("authorizationUrl", ciLogonAuthorizationEndpoint);
            idp.getConfig().put("tokenUrl", ciLogonTokenEndpoint);
            idp.getConfig().put("userInfoUrl", ciLogonUserInfoEndpoint);
            idp.getConfig().put("defaultScope", scopes);
            idp.getConfig().put("issuer", ciLogonIssuerUri);
            idp.getConfig().put("jwksUri", jwksUri);
            idp.getConfig().put("forwardParameters", "idphint");

            realmResource.identityProviders().create(idp);

        } catch (Exception ex) {
            String msg = "Error occurred while configuring  IDP in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
        return true;
    }


    /**
     * This adds user attributes to users
     *
     * @param realmId
     * @param attributeMap
     * @param users
     * @return
     */
    public boolean addUserAttributes(String realmId, String accessToken, Map<String, List<String>> attributeMap, List<String> users) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            RealmResource realmResource = client.realm(realmId);

            for (String user : users) {
                UserRepresentation userRepresentation = getUserByUsername(client, realmId, user.toLowerCase());

                if (userRepresentation != null) {
                    UserResource resource = realmResource.users().get(userRepresentation.getId());

                    Map<String, List<String>> exAtrMap = userRepresentation.getAttributes();

                    if (exAtrMap != null && !exAtrMap.isEmpty()) {
                        attributeMap.keySet().forEach(key -> {
                            exAtrMap.put(key, attributeMap.get(key));
                        });
                        userRepresentation.setAttributes(exAtrMap);

                    } else {
                        userRepresentation.setAttributes(attributeMap);
                    }

                    resource.update(userRepresentation);
                }
            }

        } catch (Exception ex) {
            String msg = "Error occurred while adding user attributes in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
        return true;
    }


    /**
     * Deletes the specified attributes for the given users in Keycloak Server.
     *
     * @param realmId      The ID of the realm in Keycloak.
     * @param accessToken  The access token for authentication.
     * @param attributeMap A map containing attribute names as keys and lists of attribute values to delete as values.
     * @param users        A list of usernames for the users whose attributes need to be deleted.
     * @return true if the attributes were successfully deleted, false otherwise.
     * @throws RuntimeException if an error occurs while deleting user attributes.
     */
    public boolean deleteUserAttributes(String realmId, String accessToken, Map<String, List<String>> attributeMap, List<String> users) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            RealmResource realmResource = client.realm(realmId);

            for (String user : users) {
                UserRepresentation userRepresentation = getUserByUsername(client, realmId, user.toLowerCase());
                UserResource resource = realmResource.users().get(userRepresentation.getId());

                Map<String, List<String>> exAtrMap = userRepresentation.getAttributes();

                if (exAtrMap != null && !exAtrMap.isEmpty()) {
                    attributeMap.keySet().forEach(key -> {
                        List<String> stringList = exAtrMap.get(key);
                        if (stringList != null && !stringList.isEmpty()) {
                            stringList.removeAll(attributeMap.get(key));
                            exAtrMap.put(key, stringList);
                        }
                    });
                    userRepresentation.setAttributes(exAtrMap);
                }

                resource.update(userRepresentation);
            }


        } catch (Exception ex) {
            String msg = "Error occurred while deleting user attributes in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
        return true;
    }

    /**
     * Adds a protocol mapper to a client in the specified realm
     *
     * @param protocolMapperRepresentation the representation of the protocol mapper to be added
     * @param realmId                      the ID of the realm
     * @param clientId                     the ID of the client
     * @return true if the protocol mapper was successfully added, false otherwise
     */
    public boolean addProtocolMapper(ProtocolMapperRepresentation protocolMapperRepresentation,
                                     String realmId, String clientId) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {

            RealmResource realmResource = client.realm(realmId);

            ClientRepresentation representation = realmResource.clients().findByClientId(clientId).get(0);

            ProtocolMappersResource resource = realmResource.clients().get(representation.getId()).getProtocolMappers();
            resource.createMapper(protocolMapperRepresentation);

        } catch (Exception ex) {
            String msg = "Error occurred while adding protocol mappers in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);

        }
        return true;
    }


    /**
     * Retrieves a list of UserRepresentation objects for all users in the specified realm.
     *
     * @param realmId The ID of the realm from which to retrieve the users.
     * @return A List of UserRepresentation objects containing the user details.
     * @throws RuntimeException if an error occurred while fetching the user details.
     */
    public List<UserRepresentation> getAllUsers(String realmId) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            List<UserRepresentation> representations = client.realm(realmId).users().list();
            List<UserRepresentation> representationList = new ArrayList<>();
            if (representations != null && !representations.isEmpty()) {
                for (UserRepresentation userRepresentation : representations) {
                    UserRepresentation userRep = getUserByUsername(client, realmId, userRepresentation.getUsername());
                    representationList.add(userRep);
                }
            }
            return representationList;
        } catch (Exception ex) {
            String msg = "Error occurred while adding protocol mappers in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    /**
     * Adds a list of roles to a realm or client in Keycloak Server.
     *
     * @param roleRepresentations The list of role representations to be added.
     * @param realmId             The ID of the realm.
     * @param clientId            The ID of the client (if clientScope is true).
     * @param clientScope         Flag indicating whether the roles should be added to a client or realm.
     * @return A boolean indicating whether the roles were successfully added.
     * @throws RuntimeException if an error occurs while adding roles in Keycloak Server.
     */
    public boolean addRoles(List<RoleRepresentation> roleRepresentations, String realmId, String clientId, boolean clientScope) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmResource realmResource = client.realm(realmId);

            if (clientScope) {
                ClientRepresentation representation = realmResource.clients().findByClientId(clientId).get(0);

                for (RoleRepresentation roleRepresentation : roleRepresentations) {
                    realmResource.clients().get(representation.getId()).roles().create(roleRepresentation);
                }

            } else {
                for (RoleRepresentation representation : roleRepresentations) {
                    realmResource.roles().create(representation);
                }
            }

        } catch (Exception ex) {
            String msg = "Error occurred while adding roles in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
        return true;
    }

    /**
     * Deletes a role in Keycloak server.
     *
     * @param id          the ID of the role to be deleted
     * @param realmId     the ID of the realm in which the role exists
     * @param clientId    the ID of the client for which the role is associated
     * @param clientScope if true add roles to client else to realm
     * @return true if the role is successfully deleted, false otherwise
     * @throws RuntimeException if an error occurs while deleting the role
     */
    public boolean deleteRole(String id, String realmId, String clientId, boolean clientScope) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmResource realmResource = client.realm(realmId);

            if (clientScope) {
                ClientRepresentation representation = realmResource.clients().findByClientId(clientId).get(0);
                realmResource.clients().get(representation.getId()).roles().deleteRole(id);

            } else {
                realmResource.roles().deleteRole(id);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while delete role" + id + " in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
        return true;
    }


    /**
     * Provides all Roles belongs to client, if clientId not present, provides all
     * Roles related to Realm
     *
     * @param realmId  The ID of the realm to retrieve the roles from.
     * @param clientId Optional parameter to filter roles for a specific client. Null to retrieve all roles in the realm.
     * @return A list of RoleRepresentation objects representing the retrieved roles.
     * @throws RuntimeException If an error occurs while accessing the Keycloak Server.
     */
    public List<RoleRepresentation> getAllRoles(String realmId, String clientId) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmResource realmResource = client.realm(realmId);

            if (clientId != null) {
                ClientRepresentation representation = realmResource.clients().findByClientId(clientId).get(0);
                return realmResource.clients().get(representation.getId()).roles().list();

            } else {
                return realmResource.roles().list();
            }

        } catch (Exception ex) {
            String msg = "Error occurred while adding roles in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    /**
     * Configures event persistence for a given realm and event type.
     *
     * @param realmId      the ID of the realm
     * @param eventType    the type of event to configure
     * @param time         the expiration time for the events (in milliseconds)
     * @param enableEvents whether to enable event persistence
     * @param isAdminEvent whether the event is an admin event
     * @return true if event persistence was successfully configured, false otherwise
     * @throws RuntimeException if an error occurred while configuring event persistence
     */
    public boolean configureEventPersistence(String realmId, String eventType, long time, boolean enableEvents, boolean isAdminEvent) {

        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmEventsConfigRepresentation representation = client.realm(realmId).getRealmEventsConfig();

            if (isAdminEvent) {
                representation.setAdminEventsEnabled(true);
            } else {
                representation.setEventsEnabled(enableEvents);
                representation.setEventsExpiration(time);
                List<String> eventTypes = representation.getEnabledEventTypes();
                if (eventTypes != null && !eventTypes.isEmpty() && !eventTypes.contains(eventType)) {
                    eventTypes.add(eventType);

                } else {
                    eventTypes = new ArrayList<>();
                    eventTypes.add(eventType);
                }

                representation.setEnabledEventTypes(eventTypes);
                client.realm(realmId).updateRealmEventsConfig(representation);
            }
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred while configuring event persistence events, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    /**
     * Retrieve the last login event for a given user.
     *
     * @param realmId  the ID of the realm in which the user is authenticated
     * @param clientId the ID of the client involved in the login
     * @param username the username of the user
     * @return the most recent EventRepresentation object representing the last login event for the user,
     * or null if no login event is found or an error occurred
     */
    public EventRepresentation getLastLoginEvent(String realmId, String clientId, String username) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            List<EventRepresentation> eventRepresentations = client.realm(realmId).getEvents();

            for (EventRepresentation representation : eventRepresentations) {
                Map<String, String> map = representation.getDetails();
                if (map != null && !map.isEmpty()) {
                    for (String key : map.keySet()) {
                        if (key.equals("username") && map.get(key).equals(username)) {
                            return representation;
                        }
                    }
                }
            }
            return null;

        } catch (Exception ex) {
            String msg = "Error occurred while pulling events, reason: " + ex.getMessage();
            LOGGER.warn(msg, ex);
            return null;
        }
    }

    /**
     * Retrieves the latest user session for the given realm, client, access token, and username.
     * Returns null if no user session is found.
     *
     * @param realmId     the ID of the realm
     * @param clientId    the ID of the client
     * @param accessToken the access token for authorization
     * @param username    the username of the user
     * @return the latest UserSessionRepresentation if found, otherwise null
     * @throws RuntimeException if an error occurs while retrieving the user session
     */
    public UserSessionRepresentation getLatestSession(String realmId, String clientId, String accessToken, String username) {

        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            List<UserRepresentation> userResourceList = client.realm(realmId).users().search(
                    username.toLowerCase(), null, null, null, null, null);

            if (!userResourceList.isEmpty() && userResourceList.get(0).getUsername().equals(username.toLowerCase())) {
                UserRepresentation userRepresentation = userResourceList.get(0);
                List<UserSessionRepresentation> userSessionRepresentations = client.realm(realmId).users().get(userRepresentation.getId()).getUserSessions();

                if (!userSessionRepresentations.isEmpty()) {
                    return userSessionRepresentations.get(userSessionRepresentations.size() - 1);
                }
            }
            return null;

        } catch (Exception ex) {
            String msg = "Error occurred while pulling active user sessions, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean deleteExternalIDPLinks(String realmId) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmResource realmResource = client.realm(realmId);
            List<UserRepresentation> userResourceList = client.realm(realmId).users().list();
            userResourceList.forEach(user -> {
                UserResource userResource = realmResource.users().get(user.getId());
                List<FederatedIdentityRepresentation> federatedIdentityRepresentations =
                        userResource.getFederatedIdentity();
                if (federatedIdentityRepresentations != null && !federatedIdentityRepresentations.isEmpty()) {
                    federatedIdentityRepresentations.forEach(fed -> {
                        userResource.removeFederatedIdentity(fed.getIdentityProvider());
                    });
                }
            });
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred while deleting external IDP links of realm " + realmId + ", reason " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public boolean deleteExternalIDPLinks(String realmId, List<String> users) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmResource realmResource = client.realm(realmId);
            List<UserRepresentation> userResourceList = client.realm(realmId).users().list();
            userResourceList.forEach(user -> {
                if (users.contains(user.getUsername())) {
                    UserResource userResource = realmResource.users().get(user.getId());
                    List<FederatedIdentityRepresentation> federatedIdentityRepresentations =
                            userResource.getFederatedIdentity();
                    if (federatedIdentityRepresentations != null && !federatedIdentityRepresentations.isEmpty()) {
                        federatedIdentityRepresentations.forEach(fed -> {
                            userResource.removeFederatedIdentity(fed.getIdentityProvider());
                        });
                    }
                }
            });
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred while deleting external IDP links of realm "
                    + realmId + ", reason " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public List<FederatedIdentityRepresentation> getExternalIDPLinks(String realmId, String requestedUser) {
        List<FederatedIdentityRepresentation> arrayList = new ArrayList<>();
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            RealmResource realmResource = client.realm(realmId);
            List<UserRepresentation> userResourceList = client.realm(realmId).users().list();
            userResourceList.forEach(user -> {
                if (requestedUser.equals(user.getUsername())) {
                    UserResource userResource = realmResource.users().get(user.getId());
                    List<FederatedIdentityRepresentation> federatedIdentityRepresentations =
                            userResource.getFederatedIdentity();
                    if (federatedIdentityRepresentations != null && !federatedIdentityRepresentations.isEmpty()) {
                        arrayList.addAll(federatedIdentityRepresentations);
                    }
                }
            });
            return arrayList;

        } catch (Exception ex) {
            String msg = "Error occurred while deleting external IDP links of realm " + realmId + ", reason " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public void addExternalIDPLinks(String realmId, List<FederatedIdentityRepresentation> representations) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            if (representations != null && !representations.isEmpty()) {
                representations.forEach(fed -> {
                    List<UserRepresentation> userRepresentationList = client.realm(realmId).users().search(fed.getUserName());
                    userRepresentationList.forEach(user -> {
                        UserResource userResource = client.realm(realmId).users().get(user.getId());
                        userResource.addFederatedIdentity(fed.getIdentityProvider(), fed);
                    });

                });
            }

        } catch (Exception ex) {
            String msg = "Error occurred while adding external IDP links " + realmId + ", reason " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);

        }

    }


    /**
     * Creates groups in Keycloak for a given realm.
     *
     * @param realmId              The ID of the realm.
     * @param clientId             The ID of the client.
     * @param clientSec            The client secret.
     * @param groupRepresentations The list of GroupRepresentation objects representing the groups to be created.
     * @return A list of GroupRepresentation objects representing the created groups in Keycloak.
     */
    public List<GroupRepresentation> createGroups(String realmId, String clientId, String clientSec, List<GroupRepresentation> groupRepresentations) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            List<GroupRepresentation> representationList = new ArrayList<>();

            for (GroupRepresentation representation : groupRepresentations) {
                Response response = client.realm(realmId).groups().add(representation);

                if (response.getStatus() == HttpStatus.SC_CREATED) {
                    String id = getCreatedId(response);

                    if (representation.getRealmRoles() != null && !representation.getRealmRoles().isEmpty()) {
                        List<RoleRepresentation> roleRepresentation = new ArrayList<>();
                        for (String role : representation.getRealmRoles()) {
                            RoleResource resource = client.realm(realmId).roles().get(role);
                            if (resource != null) {
                                roleRepresentation.add(resource.toRepresentation());
                            }
                        }
                        if (!roleRepresentation.isEmpty()) {
                            client.realm(realmId).groups().group(id).roles().realmLevel().add(roleRepresentation);
                        }

                    }

                    if (representation.getClientRoles() != null && !representation.getClientRoles().isEmpty()) {
                        List<RoleRepresentation> clientRepresentations = new ArrayList<>();
                        ClientRepresentation clientRepresentation =
                                client.realm(realmId).clients().findByClientId(clientId).get(0);
                        for (String role : representation.getClientRoles().get(clientId)) {

                            RoleResource resource = client.realm(realmId).clients().get(clientRepresentation.getId()).roles().get(role);

                            if (resource != null) {
                                clientRepresentations.add(resource.toRepresentation());
                            }
                        }
                        if (!clientRepresentations.isEmpty()) {
                            client.realm(realmId).groups().group(id).roles().
                                    clientLevel(clientRepresentation.getId()).add(clientRepresentations);
                        }

                    }

                    representation.setId(id);
                    this.createGroup(client, realmId, clientId, representation);
                    response.close();
                    GroupRepresentation savedRep = client.realm(realmId).groups().group(representation.getId()).toRepresentation();
                    representationList.add(savedRep);
                    return representationList;

                } else if (response.getStatus() == HttpStatus.SC_UNAUTHORIZED) {
                    String msg = "Error occurred while creating group, reason: HTTP " + response.getStatus() + " Unauthorized";
                    LOGGER.error(msg);
                    throw new RuntimeException(msg);

                } else {
                    String msg = "Error occurred while creating group, reason: HTTP  " + response.getStatus();
                    LOGGER.error(msg);
                    throw new RuntimeException(msg);
                }
            }
        } catch (Exception ex) {
            String msg = "Error occurred while creating group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
        return null;
    }


    /**
     * Updates the given group representation in the specified realm.
     *
     * @param realmId             The ID of the realm.
     * @param clientId            The ID of the client.
     * @param clientSec           The secret key of the client.
     * @param groupRepresentation The group representation to update.
     * @return The updated group representation.
     */
    public GroupRepresentation updateGroup(String realmId, String clientId, String clientSec, GroupRepresentation groupRepresentation) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            client.realm(realmId).groups().group(groupRepresentation.getId()).update(groupRepresentation);

            List<RoleRepresentation> exRoles = client.realm(realmId).groups().group(groupRepresentation.getId()).roles().realmLevel().listAll();

            if (exRoles != null && !exRoles.isEmpty()) {
                client.realm(realmId).groups().group(groupRepresentation.getId()).roles().realmLevel().remove(exRoles);
            }

            if (groupRepresentation.getRealmRoles() != null && !groupRepresentation.getRealmRoles().isEmpty()) {
                List<RoleRepresentation> roleRepresentation = new ArrayList<>();
                for (String role : groupRepresentation.getRealmRoles()) {
                    RoleResource resource = client.realm(realmId).roles().get(role);
                    if (resource != null) {
                        roleRepresentation.add(resource.toRepresentation());
                    }
                }
                if (!roleRepresentation.isEmpty()) {
                    client.realm(realmId).groups().group(groupRepresentation.getId()).roles().realmLevel().add(roleRepresentation);
                }

            }

            ClientRepresentation clientRepresentation = client.realm(realmId).clients().findByClientId(clientId).get(0);

            List<RoleRepresentation> exClientRoles = client.realm(realmId).groups().group(groupRepresentation.getId())
                    .roles().clientLevel(clientRepresentation.getId()).listAll();

            if (exClientRoles != null && !exClientRoles.isEmpty()) {
                client.realm(realmId).groups().group(groupRepresentation.getId())
                        .roles().clientLevel(clientRepresentation.getId()).remove(exClientRoles);
            }

            if (groupRepresentation.getClientRoles() != null && !groupRepresentation.getClientRoles().isEmpty()) {
                List<RoleRepresentation> clientRepresentations = new ArrayList<>();

                for (String role : groupRepresentation.getClientRoles().get(clientId)) {
                    RoleResource resource = client.realm(realmId).clients().get(clientRepresentation.getId()).roles().get(role);

                    if (resource != null) {
                        clientRepresentations.add(resource.toRepresentation());
                    }
                }
                if (!clientRepresentations.isEmpty()) {
                    client.realm(realmId).groups().group(groupRepresentation.getId()).roles().
                            clientLevel(clientRepresentation.getId()).add(clientRepresentations);
                }

            }

            return client.realm(realmId).groups().group(groupRepresentation.getId()).toRepresentation();

        } catch (Exception ex) {
            String msg = "Error occurred while updating group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    /**
     * Deletes a group from Keycloak.
     *
     * @param realmId   the ID of the realm containing the group
     * @param clientId  the ID of the client used for authentication
     * @param clientSec the secret key of the client used for authentication
     * @param groupId   the ID of the group to be deleted
     * @return true if the group is deleted successfully, false otherwise
     */
    public boolean deleteGroup(String realmId, String clientId, String clientSec, String groupId) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            String id = client.realm(realmId).groups().group(groupId).toRepresentation().getId();
            client.realm(realmId).groups().group(id).remove();
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred while deleting group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    /**
     * Finds a group in the Keycloak server.
     *
     * @param realmId     the ID of the realm where the group is located
     * @param accessToken the access token used for authentication
     * @param id          the ID of the group to find
     * @param name        the name of the group to find
     * @return the GroupRepresentation object representing the found group, or null if not found or an error occurred
     */
    public GroupRepresentation findGroup(String realmId, String accessToken, String id, String name) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            if (id != null && !id.trim().isEmpty()) {
                GroupResource resource = client.realm(realmId).groups().group(id);
                if (resource != null) {
                    return resource.toRepresentation();
                }
            } else {
                List<GroupRepresentation> groupRepresentations = client.
                        realm(realmId).groups().groups(name, 0, 1);
                if (groupRepresentations != null && !groupRepresentations.isEmpty()) {
                    return groupRepresentations.get(0);
                }
            }

        } catch (Exception ex) {
            if (ex.getMessage().contains("HTTP 404")) {
                return null;
            } else {
                String msg = "Error occurred finding groups, reason: " + ex.getMessage();
                LOGGER.error(msg, ex);
                throw new RuntimeException(msg, ex);
            }
        }
        return null;
    }


    /**
     * Retrieves all groups for a given realm.
     *
     * @param realmId     The ID of the realm.
     * @param accessToken The access token for authentication.
     * @return A list of GroupRepresentation objects representing the groups in the realm.
     * Returns null if an HTTP 404 error occurs.
     */
    public List<GroupRepresentation> getAllGroups(String realmId, String accessToken) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            List<GroupRepresentation> groupRepresentations = new ArrayList<>();

            for (GroupRepresentation representation : client.realm(realmId).groups().groups()) {
                groupRepresentations.
                        add(client.realm(realmId).groups().group(representation.getId()).toRepresentation());
            }

            return groupRepresentations;

        } catch (Exception ex) {
            if (ex.getMessage().contains("HTTP 404")) {
                return null;
            } else {
                String msg = "Error occurred finding groups, reason: " + ex.getMessage();
                LOGGER.error(msg, ex);
                throw new RuntimeException(msg, ex);
            }
        }
    }


    public boolean addUserToGroup(String realmId, String username, String groupId, String accessToken) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            client.realm(realmId).users().get(userRepresentation.getId()).joinGroup(groupId);
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred while adding user to group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean removeUserFromGroup(String realmId, String username, String groupId, String accessToken) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            client.realm(realmId).users().get(userRepresentation.getId()).leaveGroup(groupId);
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred while remove user from group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean configureAgentClient(String realmId, String clientId, long accessTokenLifeTime) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            ClientRepresentation representation = client.realm(realmId).clients().findByClientId(clientId).get(0);

            if (representation != null) {
                Map<String, String> attributes = representation.getAttributes();

                if (attributes == null || attributes.isEmpty()) {
                    attributes = new HashMap<>();
                }
                attributes.put("access.token.lifespan", String.valueOf(accessTokenLifeTime));

                client.realm(realmId).clients().get(representation.getId()).update(representation);
                return true;
            }
            return false;

        } catch (Exception ex) {
            String msg = "Error occurred while remove user from group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }


    public boolean isValidEndUser(String realmId, String username, String accessToken) {
        try (Keycloak client = getClient(iamServerURL, realmId, accessToken)) {
            return isValidEndUser(client, realmId, username);
        } catch (Exception ex) {
            String msg = "Error occurred end user validity: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public boolean isValidEndUser(String realmId, String username) {
        try (Keycloak client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword)) {
            return isValidEndUser(client, realmId, username);
        } catch (Exception ex) {
            String msg = "Error occurred end user validity: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public Map<String, Object> getUserInfo(String accessToken, long realmId) {
        String userInfoEndpoint = iamServerURL + "realms/" + realmId + "/protocol/openid-connect/userinfo";

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(userInfoEndpoint))
                .header("Authorization", "Bearer " + accessToken)
                .header("Accept", "application/json")
                .GET()
                .build();

        HttpClient client = HttpClient.newHttpClient();

        try {
            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
            if (response.statusCode() == 200) {
                return objectMapper.readValue(response.body(), Map.class);
            } else {
                throw new IllegalStateException("Failed to fetch user info: Status code " + response.statusCode());
            }
        } catch (Exception e) {
            throw new RuntimeException("Error fetching user info", e);
        }
    }


    private boolean isValidEndUser(Keycloak client, String realmId, String username) {
        UserRepresentation representation = getUserByUsername(client, realmId, username);

        if (representation == null) {
            return false;
        }

        Map<String, List<String>> attributes = representation.getAttributes();

        if (attributes != null && !attributes.isEmpty()) {

            for (String key : attributes.keySet()) {
                if (key.equals(Constants.REALM_AGENT)) {
                    return false;
                }

            }
        }
        return true;
    }


    private Keycloak getClient(String adminUrl, String realm, String loginUsername, String password) {
        return KeycloakUtils.getClient(adminUrl, realm, loginUsername, password, clientId);
    }

    private Keycloak getClient(String adminUrl, String realm, String accessToken) {
        return KeycloakUtils.getClient(adminUrl, realm, accessToken);
    }


    private RealmRepresentation createDefaultRoles(RealmRepresentation realmDetails) {
        List<RoleRepresentation> defaultRoles = new ArrayList<RoleRepresentation>();
        RoleRepresentation adminRole = new RoleRepresentation();
        adminRole.setName("admin");
        adminRole.setDescription("Admin role for PGA users");
        defaultRoles.add(adminRole);

        RolesRepresentation rolesRepresentation = new RolesRepresentation();
        rolesRepresentation.setRealm(defaultRoles);
        realmDetails.setRoles(rolesRepresentation);
        return realmDetails;
    }


    private String getRealmManagementClientId(Keycloak client, String realmId) {
        List<ClientRepresentation> realmClients = client.realm(realmId).clients().findAll();
        String realmManagementClientId = null;
        for (ClientRepresentation realmClient : realmClients) {
            if (realmClient.getClientId().equals("realm-management")) {
                realmManagementClientId = realmClient.getId();
            }
        }
        return realmManagementClientId;
    }

    private Optional<String> getRealmManagementClientSecret(Keycloak client, String realmId) {
        List<ClientRepresentation> realmClients = client.realm(realmId).clients().findAll();
        String realmManagementClientId = null;
        for (ClientRepresentation realmClient : realmClients) {
            if (realmClient.getClientId().equals("realm-management")) {
                realmManagementClientId = realmClient.getClientId();
                String ClientUUID = client.realms().realm(realmId).clients().findByClientId(realmManagementClientId).get(0).getId();
                return Optional.ofNullable(client.realms().realm(realmId).clients().get(ClientUUID).getSecret().getValue());
            }
        }
        return Optional.empty();
    }


    private UserRepresentation getUserByUsername(Keycloak client, String tenantId, String username) {

        // Searching for users by username returns also partial matches, so need to filter down to an exact match if it exists
        List<UserRepresentation> userResourceList = client.realm(tenantId).users().search(
                username.toLowerCase(), null, null, null, null, null);

        for (UserRepresentation userRepresentation : userResourceList) {
            if (userRepresentation.getUsername().equals(username.toLowerCase())) {
                RoleMappingResource resource = client.realm(tenantId).users().get(userRepresentation.getId()).roles();
                MappingsRepresentation representation = resource.getAll();
                if (representation != null && representation.getRealmMappings() != null) {
                    List<String> roleRepresentations = new ArrayList<>();
                    representation.getRealmMappings().forEach(t -> roleRepresentations.add(t.getName()));
                    userRepresentation.setRealmRoles(roleRepresentations);
                }
                if (representation != null && representation.getClientMappings() != null) {
                    Map<String, List<String>> roleRepresentations = new HashMap<>();
                    representation.getClientMappings().keySet().forEach(key -> {
                        if (representation.getClientMappings().get(key).getMappings() != null) {
                            List<String> roleList = new ArrayList<>();
                            representation.getClientMappings().get(key).getMappings().forEach(t -> roleList.add(t.getName()));
                            roleRepresentations.put(key, roleList);
                        }
                    });
                    userRepresentation.setClientRoles(roleRepresentations);
                }

                return userRepresentation;
            }
        }
        return null;
    }


    private boolean createGroup(Keycloak client, String realmId, String clientId, GroupRepresentation parentRepresentation) {

        if (parentRepresentation.getSubGroups() != null && !parentRepresentation.getSubGroups().isEmpty()) {
            List<GroupRepresentation> groupRepresentations = parentRepresentation.getSubGroups();
            if (groupRepresentations != null && !groupRepresentations.isEmpty()) {
                for (GroupRepresentation representation : groupRepresentations) {
                    Response createdRes = client.realm(realmId).groups().add(representation);
                    String id = getCreatedId(createdRes);
                    if (id != null) {
                        representation.setId(id);
                        Response response = client.realm(realmId).groups().group(parentRepresentation.getId()).subGroup(representation);
                        if (response.getStatus() == HttpStatus.SC_CREATED || response.getStatus() == HttpStatus.SC_NO_CONTENT) {
                            if (representation.getRealmRoles() != null && !representation.getRealmRoles().isEmpty()) {
                                List<RoleRepresentation> roleRepresentation = new ArrayList<>();
                                for (String role : representation.getRealmRoles()) {
                                    RoleResource resource = client.realm(realmId).roles().get(role);
                                    if (resource != null) {
                                        roleRepresentation.add(resource.toRepresentation());
                                    }
                                }
                                if (!roleRepresentation.isEmpty()) {
                                    client.realm(realmId).groups().group(id).roles().realmLevel().add(roleRepresentation);
                                }

                            }

                            if (representation.getClientRoles() != null && !representation.getClientRoles().isEmpty()) {
                                List<RoleRepresentation> clientRepresentations = new ArrayList<>();
                                ClientRepresentation clientRepresentation =
                                        client.realm(realmId).clients().findByClientId(clientId).get(0);
                                for (String role : representation.getClientRoles().get(clientId)) {
                                    RoleResource resource = client.realm(realmId).clients().get(clientRepresentation.getId()).roles().get(role);

                                    if (resource != null) {
                                        clientRepresentations.add(resource.toRepresentation());
                                    }
                                }
                                if (!clientRepresentations.isEmpty()) {
                                    client.realm(realmId).groups().group(id).roles().
                                            clientLevel(clientRepresentation.getId()).add(clientRepresentations);
                                }

                            }
                            createGroup(client, realmId, clientId, representation);
                        }
                        response.close();
                    }
                }
            }
        }
        return true;
    }


    private List<UserRepresentation> searchUsers(Keycloak client, String tenantId, String username,
                                                 String firstName, String lastName, String email, String search, int offset, int limit) {

        // Searching for users by username returns also partial matches, so need to filter down to an exact match if it exists
        List<UserRepresentation> userResourceList = null;
        if (search != null && !search.trim().equals("")) {
            userResourceList = client.realm(tenantId).users().search(search, offset, limit);
        } else {

            userResourceList = client.realm(tenantId).users().search(
                    username.toLowerCase(), firstName, lastName, email, offset, limit);
        }

        if (userResourceList != null && !userResourceList.isEmpty()) {
            List<UserRepresentation> newList = new ArrayList<>();
            for (UserRepresentation userRepresentation : userResourceList) {
                RoleMappingResource resource = client.realm(tenantId).users().get(userRepresentation.getId()).roles();
                MappingsRepresentation representation = resource.getAll();
                if (representation != null && representation.getRealmMappings() != null) {
                    List<String> roleRepresentations = new ArrayList<>();
                    representation.getRealmMappings().forEach(t -> roleRepresentations.add(t.getName()));
                    userRepresentation.setRealmRoles(roleRepresentations);
                }
                if (representation != null && representation.getClientMappings() != null) {
                    Map<String, List<String>> roleRepresentations = new HashMap<>();
                    representation.getClientMappings().keySet().forEach(key -> {
                        if (representation.getClientMappings().get(key).getMappings() != null) {
                            List<String> roleList = new ArrayList<>();
                            representation.getClientMappings().get(key).getMappings().forEach(t -> roleList.add(t.getName()));
                            roleRepresentations.put(key, roleList);
                        }
                    });
                    userRepresentation.setClientRoles(roleRepresentations);
                }
                newList.add(userRepresentation);
            }
            return userResourceList;
        }

        return userResourceList;
    }

    private String getCreatedId(Response response) {
        URI location = response.getLocation();
        if (location == null) {
            return null;
        }
        String path = location.getPath();
        return path.substring(path.lastIndexOf('/') + 1);
    }

}
