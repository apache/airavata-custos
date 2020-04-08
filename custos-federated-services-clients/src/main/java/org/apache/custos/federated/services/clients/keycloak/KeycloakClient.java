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

package org.apache.custos.federated.services.clients.keycloak;

import org.apache.catalina.security.SecurityUtil;
import org.apache.custos.core.services.commons.util.Constants;
import org.apache.http.HttpStatus;
import org.jboss.resteasy.client.jaxrs.ResteasyClient;
import org.jboss.resteasy.client.jaxrs.ResteasyClientBuilder;
import org.keycloak.admin.client.Keycloak;
import org.keycloak.admin.client.resource.*;
import org.keycloak.representations.idm.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import javax.validation.constraints.NotNull;
import javax.ws.rs.NotFoundException;
import javax.ws.rs.core.Response;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.security.KeyStore;
import java.util.*;
import java.util.stream.Collectors;

/**
 * This class acts as a rest client for keycloak server
 */
@Component
public class KeycloakClient {
    private final static Logger LOGGER = LoggerFactory.getLogger(KeycloakClient.class);


    private final static int POOL_SIZE = 10;

    private final static int ACCESS_TOKEN_LIFE_SPAN = 1800;

    private final static int SESSION_IDLE_TIMEOUT = 3600;

    @Value("${iam.server.client.id:admin-cli}")
    private String clientId;

    @Value("${iam.server.truststore.path:/home/ubuntu/keystore/keycloak-client-truststore.pkcs12}")
    private String trustStorePath;

    @Value("${iam.server.truststore.password:keycloak}")
    private String truststorePassword;

    @Value("${iam.server.url:https://keycloak.custos.scigap.org:31000/auth/}")
    private String iamServerURL;

    @Value("${iam.server.admin.username:keycloak}")
    private String superAdminUserName;

    @Value("${iam.server.admin.password:5SocwZ78TEUyyj0R}")
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

    public void createRealm(String realmId, String displayName) {
        Keycloak client = null;
        try {
            // get client
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);
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
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public void updateRealm(String realmId, String displayName) {
        Keycloak client = null;
        try {
            // get client
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);
            // create realm

            RealmResource realmResource = client.realm(realmId);

            if (realmResource != null) {

                RealmRepresentation newRealmDetails = realmResource.toRepresentation();
                newRealmDetails.setId(realmId);
                newRealmDetails.setDisplayName(displayName);
                newRealmDetails.setRealm(realmId);
                realmResource.update(newRealmDetails);
            } else {
                String msg = "Realm not found, reason: ";
                LOGGER.error(msg);
                throw new RuntimeException(msg, null);
            }

        } catch (Exception ex) {
            String msg = "Error creating Realm in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }

    public boolean createRealmAdminAccount(String realmId, String adminUsername, String adminFirstname, String adminLastname, String adminEmail, String adminPassword) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);
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
                retrievedUser.roles().realmLevel().add(Arrays.asList(adminRoleResource.toRepresentation()));

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
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean updateRealmAdminAccount(String realmId, String adminUsername, String adminFirstname, String adminLastname, String adminEmail, String adminPassword) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);
            UserRepresentation representation = getUserByUsername(client, realmId, adminUsername);
            if (representation != null) {
                UserRepresentation user = representation;
                user.setUsername(adminUsername);
                user.setFirstName(adminFirstname);
                user.setLastName(adminLastname);
                user.setEmail(adminEmail);
                user.setEmailVerified(true);
                user.setEnabled(true);
                client.realm(realmId).users().get(representation.getId()).update(representation);
                return true;
            } else {
                return createRealmAdminAccount(realmId, adminUsername, adminFirstname, adminLastname, adminEmail, adminPassword);
            }
        } catch (Exception ex) {
            String msg = "Error updating Realm Admin Account in keycloak server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean grantAdminPrivilege(String realmId, String username) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);
            UserRepresentation representation = getUserByUsername(client, realmId, username);
            if (representation != null) {

                UserResource retrievedUser = client.realm(realmId).users().get(representation.getId());
                RoleResource adminRoleResource = client.realm(realmId).roles().get("admin");
                retrievedUser.roles().realmLevel().add(Arrays.asList(adminRoleResource.toRepresentation()));

                String realmManagementClientId = getRealmManagementClientId(client, realmId);

                retrievedUser.roles().clientLevel(realmManagementClientId).add(retrievedUser.roles().clientLevel(realmManagementClientId).listAvailable());
                return true;

            } else {
                String msg = "Cannot find existing user with username " + username;
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }
        } catch (Exception ex) {
            String msg = "Error granting admin privilege, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }

    public boolean removeAdminPrivilege(String realmId, String username) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);
            UserRepresentation representation = getUserByUsername(client, realmId, username);
            if (representation != null) {

                UserResource retrievedUser = client.realm(realmId).users().get(representation.getId());
                RoleResource adminRoleResource = client.realm(realmId).roles().get("admin");
                retrievedUser.roles().realmLevel().remove(Arrays.asList(adminRoleResource.toRepresentation()));
                String realmManagementClientId = getRealmManagementClientId(client, realmId);
                List<RoleRepresentation> representations = retrievedUser.roles().clientLevel(realmManagementClientId).listEffective();

                retrievedUser.roles().clientLevel(realmManagementClientId).remove(retrievedUser.roles().clientLevel(realmManagementClientId).listEffective());
                return true;

            } else {
                String msg = "Cannot find existing user with username " + username;
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }
        } catch (Exception ex) {
            String msg = "Error removing admin privilege, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public KeycloakClientSecret configureClient(String realmId, String clientName, @NotNull String tenantURL, List<String> redirectUris) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);
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


            // Remove trailing slash from gatewayURL
            if (tenantURL.endsWith("/")) {
                tenantURL = tenantURL.substring(0, tenantURL.length() - 1);
            }
            // Add redirect URL after login
            // redirectUris.add(tenantURL + "/callback-url"); // PGA
            // redirectUris.add(tenantURL + "/auth/callback*"); // Django
            // Add redirect URL after logout

            List<String> newList = new ArrayList<>();
            newList.addAll(redirectUris);
            newList.add(tenantURL);


            pgaClient.setRedirectUris(newList);
            pgaClient.setPublicClient(false);
            Response httpResponse = client.realms().realm(realmId).clients().create(pgaClient);
            LOGGER.debug("Realm client configuration exited with code : " + httpResponse.getStatus() + " : " + httpResponse.getStatusInfo());

            // Add the manage-users role to the web client
            UserRepresentation serviceAccountUserRepresentation = getUserByUsername(client, realmId, "service-account-" + pgaClient.getClientId());
            UserResource serviceAccountUser = client.realms().realm(realmId).users().get(serviceAccountUserRepresentation.getId());
            String realmManagementClientId = getRealmManagementClientId(client, realmId);
            List<RoleRepresentation> manageUsersRole = serviceAccountUser.roles().clientLevel(realmManagementClientId).listAvailable()
                    .stream()
                    .filter(r -> r.getName().equals("manage-users"))
                    .collect(Collectors.toList());
            serviceAccountUser.roles().clientLevel(realmManagementClientId).add(manageUsersRole);

            if (httpResponse.getStatus() == HttpStatus.SC_CREATED) {
                String ClientUUID = client.realms().realm(realmId).clients().findByClientId(pgaClient.getClientId()).get(0).getId();
                CredentialRepresentation clientSecret = client.realms().realm(realmId).clients().get(ClientUUID).getSecret();
                KeycloakClientSecret keycloakClientSecret = new KeycloakClientSecret(pgaClient.getClientId(), clientSecret.getValue());
                return keycloakClientSecret;
            } else {
                LOGGER.error("Request for realm client creation failed with HTTP code : " + httpResponse.getStatus());
                LOGGER.error("Reason for realm client creation failure : " + httpResponse.getStatusInfo());
                throw new RuntimeException("Reason for realm client creation failure :" + httpResponse.getStatusInfo(), null);
            }
        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);

            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public KeycloakClientSecret updateClient(String realmId, String clientName, @NotNull String tenantURL, List<String> redirectUris) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);

            List<ClientRepresentation> clientRepresentations = client.realm(realmId).clients().findByClientId(clientName);

            if (clientRepresentations == null || clientRepresentations.isEmpty()) {
                String msg = "Cannot find a client with name " + clientName;
                LOGGER.error(msg);
                throw new RuntimeException(msg);
            }

            ClientRepresentation pgaClient = clientRepresentations.get(0);

            pgaClient.setBaseUrl(tenantURL);


            // Remove trailing slash from gatewayURL
            if (tenantURL.endsWith("/")) {
                tenantURL = tenantURL.substring(0, tenantURL.length() - 1);
            }
            // Add redirect URL after login
            // redirectUris.add(tenantURL + "/callback-url"); // PGA
            // redirectUris.add(tenantURL + "/auth/callback*"); // Django
            // Add redirect URL after logout

            List<String> newList = new ArrayList<>();
            newList.addAll(redirectUris);
            newList.add(tenantURL);


            pgaClient.setRedirectUris(newList);
            pgaClient.setPublicClient(false);
            client.realms().realm(realmId).clients().get(pgaClient.getId()).update(pgaClient);

            String ClientUUID = client.realms().realm(realmId).clients().findByClientId(pgaClient.getClientId()).get(0).getId();
            CredentialRepresentation clientSecret = client.realms().realm(realmId).clients().get(ClientUUID).getSecret();
            KeycloakClientSecret keycloakClientSecret = new KeycloakClientSecret(pgaClient.getClientId(), clientSecret.getValue());
            return keycloakClientSecret;

        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);

            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean isUsernameAvailable(String realmId, String username, String accessToken) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            return userRepresentation == null;
        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean createUser(String realmId, String username, String newPassword, String firstName,
                              String lastName, String emailAddress, boolean tempPassowrd, String accessToken) throws UnauthorizedException {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
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
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean enableUserAccount(String realmId, String accessToken, String username) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);

            UserResource userResource = client.realm(realmId).users().get(userRepresentation.getId());
            UserRepresentation profile = userResource.toRepresentation();
            profile.setEnabled(true);
            // We require that a user verify their email before enabling the account
            // profile.setEmailVerified(true);
            userResource.update(profile);
            return true;
        } catch (Exception ex) {
            String msg = "Error occurred enableUserAccount, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean disableUserAccount(String realmId, String accessToken, String username) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);

            if (userRepresentation != null) {

                UserResource userResource = client.realm(realmId).users().get(userRepresentation.getId());
                UserRepresentation profile = userResource.toRepresentation();
                profile.setEnabled(false);
                // We require that a user verify their email before enabling the account
                // profile.setEmailVerified(true);
                userResource.update(profile);
            }
            return true;
        } catch (Exception ex) {
            String msg = "Error in disableUserAccount at keycloak, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }

    public boolean isUserAccountEnabled(String realmId, String accessToken, String username) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            return userRepresentation != null && userRepresentation.isEnabled();
        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }

    public boolean isUserExist(String realmId, String accessToken, String username) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);
            return userRepresentation != null;
        } catch (Exception ex) {
            String msg = "Error getting values from property file, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }

    public UserRepresentation getUser(String realmId, String accessToken, String username) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
            return getUserByUsername(client, realmId, username);
        } catch (Exception ex) {
            String msg = "Error retrieving user, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public List<UserRepresentation> getUsers(String accessToken, String realmId, int offset, int limit,
                                             String username, String firstName, String lastName, String email) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
            return searchUsers(client, realmId, username, firstName, lastName, email, offset, limit);

        } catch (Exception ex) {
            String msg = "Error occurred while searching for user, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean resetUserPassword(String accessToken, String realmId, String username, String newPassword) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
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
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public List<UserRepresentation> findUser(String accessToken, String realmId, String email, String userName) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
            return client.realm(realmId).users().search(userName,
                    null,
                    null,
                    email,
                    0, 1);
        } catch (Exception ex) {
            String msg = "Error finding user in keycloak server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }

    public void updateUserRepresentation(String accessToken, String realmId, String username,
                                         String firstname, String lastName, String email) {

        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
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
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean deleteUser(String accessToken, String realmId, String username) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
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
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean addRolesToUsers(String accessToken, String realmId, List<String> users, List<String> roles, String clientId, boolean clientLevel) {

        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

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
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean removeRoleFromUser(String accessToken, String realmId, String username, List<String> roles, String clientId, boolean clientLevel) {

        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);
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
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean deleteRealm(String realmId) {
        Keycloak client = null;
        try {
            // get client
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);

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
        } finally {
            if (client != null) {
                client.close();
            }
        }
        return true;

    }


    public boolean configureOIDCFederatedIDP(String realmId, String displayName, String scopes, KeycloakClientSecret secret, Map<String, String> configs) {
        Keycloak client = null;
        try {

            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);
            RealmResource realmResource = client.realm(realmId);

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

            realmResource.identityProviders().create(idp);


        } catch (Exception ex) {
            String msg = "Error occurred while configuring  IDP in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
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
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

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

        } finally {
            if (client != null) {
                client.close();
            }
        }
        return true;

    }


    /**
     * This deletes user attributes of users
     *
     * @param realmId
     * @param attributeMap
     * @param users
     * @return
     */
    public boolean deleteUserAttributes(String realmId, String accessToken, Map<String, List<String>> attributeMap, List<String> users) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

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

        } finally {
            if (client != null) {
                client.close();
            }
        }
        return true;

    }

    /**
     * Create protocol mapper representation in given client
     *
     * @param protocolMapperRepresentations
     * @param realmId
     * @param clientId
     * @return
     */
    public boolean addProtocolMapper(ProtocolMapperRepresentation protocolMapperRepresentations,
                                     String realmId, String clientId) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);

            RealmResource realmResource = client.realm(realmId);

            ClientRepresentation representation = realmResource.clients().findByClientId(clientId).get(0);


            ProtocolMappersResource resource = realmResource.clients().get(representation.getId()).getProtocolMappers();
            resource.createMapper(protocolMapperRepresentations);

        } catch (Exception ex) {
            String msg = "Error occurred while adding protocol mappers in Keycloak Server, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);

        } finally {
            if (client != null) {
                client.close();
            }
        }
        return true;
    }


    /**
     * Configure Roles in keycloak Realm or Client
     *
     * @param roleRepresentations
     * @param realmId
     * @param clientScope         if true add roles to client else to realm
     * @return
     */
    public boolean addRoles(List<RoleRepresentation> roleRepresentations, String realmId, String clientId, boolean clientScope) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);

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

        } finally {
            if (client != null) {
                client.close();
            }
        }
        return true;
    }


    /**
     * Provides all Roles belongs to client, if clientId not present, provides all
     * Roles related to Realm
     *
     * @param realmId
     * @param clientId
     */
    public List<RoleRepresentation> getAllRoles(String realmId, String clientId) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);

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

        } finally {
            if (client != null) {
                client.close();
            }
        }

    }

    /**
     * Configure event persistance for Keycloak Realms.
     *
     * @param realmId
     * @param eventType
     * @param time
     * @param enabelEvents
     * @param isAdminEvent
     * @return
     */
    public boolean configureEventPersistence(String realmId, String eventType, long time, boolean enabelEvents, boolean isAdminEvent) {

        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);

            RealmEventsConfigRepresentation representation = client.realm(realmId).getRealmEventsConfig();

            if (isAdminEvent) {
                representation.setAdminEventsEnabled(true);
            } else {
                representation.setEventsEnabled(enabelEvents);
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

        } finally {
            if (client != null) {
                client.close();
            }
        }

    }


    /**
     * Get Last login event of given user
     *
     * @param realmId
     * @param clientId
     * @return
     */
    public EventRepresentation getLastLoginEvent(String realmId, String clientId, String username) {

        Keycloak client = null;
        try {
            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);

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
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);

        } finally {
            if (client != null) {
                client.close();
            }
        }

    }

    /**
     * provides last active session of given user
     *
     * @param realmId
     * @param clientId
     * @param accessToken
     * @param username
     * @return
     */
    public UserSessionRepresentation getLatestSession(String realmId, String clientId, String accessToken, String username) {

        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

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

        } finally {
            if (client != null) {
                client.close();
            }
        }

    }

    /**
     * creates groups and child groups in Keycloak
     *
     * @param realmId
     * @param clientId
     * @param accessToken
     * @param groupRepresentations
     * @return
     */
    public List<GroupRepresentation> createGroups(String realmId, String clientId, String accessToken, List<GroupRepresentation> groupRepresentations) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

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
                    GroupRepresentation savedRep =
                            client.realm(realmId).groups().group(representation.getId()).toRepresentation();
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

        } finally {
            if (client != null) {
                client.close();
            }
        }
        return null;
    }


    /**
     * Update given group
     *
     * @param realmId
     * @param accessToken
     * @param groupRepresentation
     * @return
     */
    public GroupRepresentation updateGroup(String realmId, String clientId, String accessToken, GroupRepresentation groupRepresentation) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

            client.realm(realmId).groups().
                    group(groupRepresentation.getId()).update(groupRepresentation);

            List<RoleRepresentation> exRoles =
                    client.realm(realmId).groups().group(groupRepresentation.getId()).roles().realmLevel().listAll();

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

            ClientRepresentation clientRepresentation =
                    client.realm(realmId).clients().findByClientId(clientId).get(0);

            List<RoleRepresentation> exClientRoles =
                    client.realm(realmId).groups().group(groupRepresentation.getId())
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

        } finally {
            if (client != null) {
                client.close();
            }
        }

    }


    /**
     * Delete given group
     *
     * @param realmId
     * @param accessToken
     * @param groupId
     * @return
     */
    public boolean deleteGroup(String realmId, String accessToken, String groupId) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

            client.realm(realmId).groups().
                    group(groupId).remove();
            return true;
        } catch (Exception ex) {
            String msg = "Error occurred while deleting group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);

        } finally {
            if (client != null) {
                client.close();
            }
        }

    }

    /**
     * find group by group Id or group name
     *
     * @param realmId
     * @param accessToken
     * @return
     */
    public GroupRepresentation findGroup(String realmId, String accessToken, String id, String name) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

            if (id != null && !id.trim().equals("")) {
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
        } finally {
            if (client != null) {
                client.close();
            }
        }
        return null;
    }


    /**
     * pull all groups related to given realm
     *
     * @param realmId
     * @param accessToken
     * @return
     */
    public List<GroupRepresentation> getAllGroups(String realmId, String accessToken) {
        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

            return client.realm(realmId).groups().groups();


        } catch (Exception ex) {
            if (ex.getMessage().contains("HTTP 404")) {
                return null;
            } else {
                String msg = "Error occurred finding groups, reason: " + ex.getMessage();
                LOGGER.error(msg, ex);
                throw new RuntimeException(msg, ex);
            }
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean addUserToGroup(String realmId, String username, String groupId, String accessToken) {

        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);


            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);

            client.realm(realmId).users().get(userRepresentation.getId()).joinGroup(groupId);
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred while adding user to group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean removeUserFromGroup(String realmId, String username, String groupId, String accessToken) {

        Keycloak client = null;
        try {
            client = getClient(iamServerURL, realmId, accessToken);

            UserRepresentation userRepresentation = getUserByUsername(client, realmId, username);

            client.realm(realmId).users().get(userRepresentation.getId()).leaveGroup(groupId);
            return true;

        } catch (Exception ex) {
            String msg = "Error occurred while remove user from group, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }
    }


    public boolean configureAgentClient(String realmId, String clientId, long accessTokenLifeTime) {
        Keycloak client = null;
        try {

            client = getClient(iamServerURL, superAdminRealmID, superAdminUserName, superAdminPassword);

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
        } finally {
            if (client != null) {
                client.close();
            }
        }

    }


    public boolean isValidEndUser(String realmId, String username, String accessToken) {
        Keycloak client = null;
        try {

            client = getClient(iamServerURL, realmId, accessToken);

            UserRepresentation representation = getUserByUsername(client, realmId, username);

            if (representation == null) {
                return false;
            }

            Map<String, List<String>> attributes = representation.getAttributes();

            if (attributes != null && !attributes.isEmpty()) {

                for (String key : attributes.keySet()) {
                    if (key.equals(Constants.CUSTOS_REALM_AGENT)) {
                        return false;
                    }

                }
            }
            return true;
        } catch (Exception ex) {
            String msg = "Error occurred end user validity: " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        } finally {
            if (client != null) {
                client.close();
            }
        }


    }


    private ResteasyClient getRestClient() {
        return new ResteasyClientBuilder()
                .connectionPoolSize(POOL_SIZE)
                .trustStore(loadKeyStore())
                .build();
    }

    private Keycloak getClient(String adminUrl, String realm, String loginUsername, String password) {

        return KeycloakUtils.getClient(adminUrl, realm, loginUsername,
                password, clientId, trustStorePath, truststorePassword);
    }

    private Keycloak getClient(String adminUrl, String realm, String accessToken) {

        return KeycloakUtils.getClient(adminUrl, realm, accessToken, trustStorePath, truststorePassword);
    }

    private KeyStore loadKeyStore() {

        InputStream is = null;
        try {


            File trustStoreFile = new File(trustStorePath);

            if (trustStoreFile.exists()) {
                LOGGER.debug("Loading trust store file from path " + trustStorePath);
                is = new FileInputStream(trustStorePath);
            } else {
                LOGGER.debug("Trying to load trust store file form class path " + trustStorePath);
                is = SecurityUtil.class.getClassLoader().getResourceAsStream(trustStorePath);
                if (is != null) {
                    LOGGER.debug("Trust store file was loaded form class path " + trustStorePath);
                }
            }

            if (is == null) {
                throw new RuntimeException("Could not find a trust store file in path " + trustStorePath);
            }

            KeyStore ks = KeyStore.getInstance(KeyStore.getDefaultType());
            ks.load(is, truststorePassword.toCharArray());
            return ks;
        } catch (Exception e) {
            throw new RuntimeException("Failed to load trust store KeyStore instance", e);
        } finally {
            if (is != null) {
                try {
                    is.close();
                } catch (IOException e) {
                    LOGGER.error("Failed to close trust store FileInputStream", e);
                }
            }
        }
    }


    private RealmRepresentation createDefaultRoles(RealmRepresentation realmDetails) {
        List<RoleRepresentation> defaultRoles = new ArrayList<RoleRepresentation>();
        RoleRepresentation adminRole = new RoleRepresentation();
        adminRole.setName("admin");
        adminRole.setDescription("Admin role for PGA users");
        defaultRoles.add(adminRole);
        RoleRepresentation adminReadOnlyRole = new RoleRepresentation();
        adminReadOnlyRole.setName("admin-read-only");
        adminReadOnlyRole.setDescription("Read only role for PGA Admin users");
        defaultRoles.add(adminReadOnlyRole);
        RoleRepresentation gatewayUserRole = new RoleRepresentation();
        gatewayUserRole.setName("gateway-user");
        gatewayUserRole.setDescription("default role for PGA users");
        defaultRoles.add(gatewayUserRole);
        RoleRepresentation pendingUserRole = new RoleRepresentation();
        pendingUserRole.setName("user-pending");
        pendingUserRole.setDescription("role for newly registered PGA users");
        defaultRoles.add(pendingUserRole);
        RoleRepresentation gatewayProviderRole = new RoleRepresentation();
        gatewayProviderRole.setName("gateway-provider");
        gatewayProviderRole.setDescription("role for gateway providers in the super-admin PGA");
        defaultRoles.add(gatewayProviderRole);
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
                                                 String firstName, String lastName, String email, int offset, int limit) {

        // Searching for users by username returns also partial matches, so need to filter down to an exact match if it exists
        List<UserRepresentation> userResourceList = client.realm(tenantId).users().search(
                username.toLowerCase(), firstName, lastName, email, offset, limit);

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
