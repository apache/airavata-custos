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

package org.apache.custos.tenant.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.credential.store.service.CredentialMetadata;
import org.apache.custos.federated.authentication.service.CacheManipulationRequest;
import org.apache.custos.iam.service.*;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.integration.core.exceptions.UnAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.messaging.email.service.EmailDisablingRequest;
import org.apache.custos.messaging.email.service.EmailEnablingRequest;
import org.apache.custos.messaging.email.service.FetchEmailFriendlyEvents;
import org.apache.custos.messaging.email.service.FetchEmailTemplatesRequest;
import org.apache.custos.messaging.service.MessageEnablingRequest;
import org.apache.custos.tenant.management.service.Credentials;
import org.apache.custos.tenant.management.service.DeleteTenantRequest;
import org.apache.custos.tenant.management.service.GetTenantRequest;
import org.apache.custos.tenant.management.service.UpdateTenantRequest;
import org.apache.custos.tenant.management.utils.Constants;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.GetTenantsRequest;
import org.apache.custos.tenant.profile.service.Tenant;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.Optional;

/**
 * This validates custos credentials
 */
@Component
public class TenantManagementAuthInterceptorImpl extends AuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(TenantManagementAuthInterceptorImpl.class);

    private CredentialStoreServiceClient credentialStoreServiceClient;

    public TenantManagementAuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient,
                                               TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
        this.credentialStoreServiceClient = credentialStoreServiceClient;
    }


    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (method.equals("createTenant")) {

            String token = getToken(headers);
            if (token == null) {
                return msg;
            }
            Optional<AuthClaim> claim = authorize(headers);
            if (claim.isEmpty()) {
                return msg;
            } else {
                return (ReqT) ((Tenant) msg).toBuilder().setParentTenantId(claim.get().getTenantId()).build();
            }

        } else if (method.equals("getTenant")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            return claim.map(cl -> {
                GetTenantRequest tenantRequest = ((GetTenantRequest) msg);

                Credentials credentials = getCredentials(cl);

                return (ReqT) tenantRequest.toBuilder()
                        .setTenantId(cl.getTenantId()).setCredentials(credentials).build();
            }).orElseThrow(() ->
            {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });


        } else if (method.equals("updateTenant")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            return claim.map(cl -> {
                UpdateTenantRequest tenantRequest = ((UpdateTenantRequest) msg);
                Credentials credentials = getCredentials(cl);
                return (ReqT) tenantRequest.toBuilder()
                        .setTenantId(cl.getTenantId()).setCredentials(credentials).build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });

        } else if (method.equals("deleteTenant")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            return claim.map(cl -> {
                DeleteTenantRequest tenantRequest = ((DeleteTenantRequest) msg);
                Credentials credentials = getCredentials(cl);
                return (ReqT) tenantRequest.toBuilder()
                        .setTenantId(cl.getTenantId()).setCredentials(credentials).build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });

        } else if (method.equals("addTenantRoles")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);

            if (claim.isPresent()) {
                AddRolesRequest rolesRequest = ((AddRolesRequest) msg);
                String clientId = rolesRequest.getClientId();
                if (rolesRequest.getClientId() == null || rolesRequest.getClientId().trim().isEmpty()) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(claim.get().getTenantId()).
                            setClientId(claim.get().getCustosId()).build();
                }

                CredentialMetadata metadata = getCredentialsFromClientId(clientId);

                if (claim.get().isSuperTenant()) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
                }

                boolean validationStatus = validateParentChildTenantRelationShip(claim.get().getTenantId(),
                        metadata.getOwnerId());

                if (validationStatus) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
                } else {
                    String error = "Request is not authorized, user not authorized with requested clientId: "
                            + clientId;
                    throw new UnAuthorizedException(error, null);
                }
            } else {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            }
        } else if (method.equals("getTenantRoles")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            if (claim.isPresent()) {
                GetRolesRequest rolesRequest = ((GetRolesRequest) msg);
                String clientId = rolesRequest.getClientId();
                if (rolesRequest.getClientId() == null || rolesRequest.getClientId().trim().isEmpty()) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(claim.get().getTenantId()).
                            setClientId(claim.get().getCustosId()).build();
                }
                CredentialMetadata metadata = getCredentialsFromClientId(clientId);

                if (claim.get().isSuperTenant()) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
                }

                boolean validationStatus = validateParentChildTenantRelationShip(claim.get().getTenantId(),
                        metadata.getOwnerId());

                if (validationStatus) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
                } else {
                    String error = "Request is not authorized, user not authorized with requested clientId: " + clientId;
                    throw new UnAuthorizedException(error, null);
                }
            } else {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            }
        } else if (method.equals("deleteRole")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            if (claim.isPresent()) {

                DeleteRoleRequest rolesRequest = ((DeleteRoleRequest) msg);
                String clientId = rolesRequest.getClientId();
                if (rolesRequest.getClientId() == null || rolesRequest.getClientId().trim().isEmpty()) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(claim.get().getTenantId()).
                            setClientId(claim.get().getCustosId()).build();
                }
                CredentialMetadata metadata = getCredentialsFromClientId(clientId);

                if (claim.get().isSuperTenant()) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
                }

                boolean validationStatus = validateParentChildTenantRelationShip(claim.get().getTenantId(),
                        metadata.getOwnerId());

                if (validationStatus) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
                } else {
                    String error = "Request is not authorized, user not authorized with requested clientId: " + clientId;
                    throw new UnAuthorizedException(error, null);
                }
            } else {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            }
        } else if (method.equals("addProtocolMapper")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);

            if (claim.isPresent()) {
                AddProtocolMapperRequest rolesRequest = ((AddProtocolMapperRequest) msg);
                String clientId = rolesRequest.getClientId();

                if (rolesRequest.getClientId() == null || rolesRequest.getClientId().trim().isEmpty()) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(claim.get().getTenantId()).
                            setClientId(claim.get().getCustosId()).build();
                }
                CredentialMetadata metadata = getCredentialsFromClientId(clientId);

                if (claim.get().isSuperTenant()) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
                }

                boolean validationStatus = validateParentChildTenantRelationShip(claim.get().getTenantId(),
                        metadata.getOwnerId());

                if (validationStatus) {
                    return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
                } else {
                    String error = "Request is not authorized, user not authorized with requested clientId: "
                            + clientId;
                    throw new UnAuthorizedException(error, null);
                }
            } else {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            }

        } else if (method.equals("configureEventPersistence")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            EventPersistenceRequest rolesRequest = ((EventPersistenceRequest) msg);
            return claim.map(cl -> {
                return (ReqT) rolesRequest.toBuilder()
                        .setTenantId(cl.getTenantId()).setPerformedBy(cl.getUsername())
                        .build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });

        } else if (method.equals("enableMessaging")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            MessageEnablingRequest rolesRequest = ((MessageEnablingRequest) msg);
            return claim.map(cl -> {
                return (ReqT) rolesRequest.toBuilder()
                        .setTenantId(cl.getTenantId())
                        .setClientId(cl.getCustosId())
                        .build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });

        } else if (method.equals("enableEmail")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            EmailEnablingRequest rolesRequest = ((EmailEnablingRequest) msg);
            return claim.map(cl -> {
                return (ReqT) rolesRequest.toBuilder()
                        .setTenantId(cl.getTenantId())
                        .setClientId(cl.getCustosId())
                        .build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });

        } else if (method.equals("disableEmail")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            EmailDisablingRequest rolesRequest = ((EmailDisablingRequest) msg);
            return claim.map(cl -> {
                return (ReqT) rolesRequest.toBuilder()
                        .setTenantId(cl.getTenantId())
                        .setClientId(cl.getCustosId())
                        .build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });

        } else if (method.equals("getEmailTemplates")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            FetchEmailTemplatesRequest rolesRequest = ((FetchEmailTemplatesRequest) msg);
            return claim.map(cl -> {
                return (ReqT) rolesRequest.toBuilder()
                        .setTenantId(cl.getTenantId())
                        .setClientId(cl.getCustosId())
                        .build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });

        } else if (method.equals("getEmailFriendlyEvents")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);
            FetchEmailFriendlyEvents rolesRequest = ((FetchEmailFriendlyEvents) msg);
            return claim.map(cl -> {
                return (ReqT) rolesRequest.toBuilder()
                        .setTenantId(cl.getTenantId())
                        .setClientId(cl.getCustosId())
                        .build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });

        } else if (method.equals("getChildTenants")) {

            Optional<AuthClaim> claim = authorizeUsingUserToken(headers);

            if (claim.isPresent()) {
                GetTenantsRequest tenantsRequest = ((GetTenantsRequest) msg);
                String clientId = tenantsRequest.getParentClientId();

                if (tenantsRequest.getParentClientId() == null ||
                        tenantsRequest.getParentClientId().trim().isEmpty()) {
                    return (ReqT) tenantsRequest.toBuilder().setParentId(claim.get().getTenantId()).build();
                }
                CredentialMetadata metadata = getCredentialsFromClientId(clientId);

                if (claim.get().isSuperTenant()) {
                    return (ReqT) tenantsRequest.toBuilder().setParentId(metadata.getOwnerId()).build();
                }

                boolean validationStatus = validateParentChildTenantRelationShip(claim.get().getTenantId(),
                        metadata.getOwnerId());

                if (validationStatus) {
                    return (ReqT) tenantsRequest.toBuilder().setParentId(metadata.getOwnerId()).build();
                } else {
                    String error = "Request is not authorized, user not authorized with requested clientId: "
                            + clientId;
                    throw new UnAuthorizedException(error, null);
                }
            } else {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            }
        } else if (method.equals("getAllTenantsForUser")) {
            validateAuth(headers);
            return msg;
        } else if (method.equals("getFromCache") || method.equals("getInstitutions")) {
            Optional<AuthClaim> claim = validateAuth(headers);

            return claim.map(cl -> {
                CacheManipulationRequest request = ((CacheManipulationRequest) msg);
                return (ReqT) request.toBuilder().setTenantId(cl.getTenantId()).build();
            }).orElseThrow(() -> {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            });
        } else if (method.equals("addToCache") || method.equals("removeFromCache")) {
            Optional<AuthClaim> claim = validateAuth(headers);
            Optional<AuthClaim> userClaim = validateUserToken(headers);

            CacheManipulationRequest.Builder request = ((CacheManipulationRequest) msg).toBuilder();

            if (userClaim.isPresent()) {
                request = request.setPerformedBy(userClaim.get().getUsername());
            }
            if (claim.isPresent()) {
                return (ReqT) request.setTenantId(claim.get().getTenantId()).build();
            } else {
                String error = "Request is not authorized, token not found";
                throw new UnAuthorizedException(error, null);
            }
        }
        return msg;
    }


    private Optional<AuthClaim> validateAuth(Metadata headers) {
        try {
            return authorize(headers);
        } catch (Exception ex) {
            LOGGER.error(" Authorizing error " + ex.getMessage());
            throw new UnAuthorizedException("Request is not authorized", ex);
        }
    }


    private Optional<AuthClaim> validateUserToken(Metadata headers) {
        try {
            String usertoken = headers.get(Metadata.Key.of(Constants.USER_TOKEN, Metadata.ASCII_STRING_MARSHALLER));
            if (usertoken == null) {
                return Optional.empty();
            }

            return authorizeUsingUserToken(usertoken);
        } catch (Exception ex) {
            LOGGER.error(" Authorizing error " + ex.getMessage());
            throw new UnAuthorizedException("Request is not authorized", ex);
        }
    }

    private Credentials getCredentials(AuthClaim claim) {
        return Credentials.newBuilder()
                .setCustosClientId(claim.getCustosId())
                .setCustosClientSecret(claim.getCustosSecret())
                .setCustosClientIdIssuedAt(claim.getCustosIdIssuedAt())
                .setCustosClientSecretExpiredAt(claim.getCustosSecretExpiredAt())
                .setIamClientId(claim.getIamAuthId())
                .setIamClientSecret(claim.getIamAuthSecret())
                .setCiLogonClientId(claim.getCiLogonId())
                .setCiLogonClientSecret(claim.getCiLogonSecret()).build();

    }
}
