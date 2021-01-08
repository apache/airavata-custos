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
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
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

/**
 * This validates custos credentials
 */
@Component
public class AuthInterceptorImpl extends AuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(AuthInterceptorImpl.class);

    private CredentialStoreServiceClient credentialStoreServiceClient;

    public AuthInterceptorImpl(CredentialStoreServiceClient credentialStoreServiceClient,
                               TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
        this.credentialStoreServiceClient = credentialStoreServiceClient;
    }


    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {


        if (method.equals("createTenant")) {

            AuthClaim claim = authorize(headers);
            if (claim == null) {
                return msg;
            } else {

                return (ReqT) ((Tenant) msg).toBuilder().setParentTenantId(claim.getTenantId()).build();
            }

        } else if (method.equals("getTenant")) {

            AuthClaim claim = authorizeUsingUserToken(headers);

            GetTenantRequest tenantRequest = ((GetTenantRequest) msg);

            Credentials credentials = getCredentials(claim);


            return (ReqT) tenantRequest.toBuilder()
                    .setTenantId(claim.getTenantId()).setCredentials(credentials).build();
        } else if (method.equals("updateTenant")) {


            AuthClaim claim = authorizeUsingUserToken(headers);

            UpdateTenantRequest tenantRequest = ((UpdateTenantRequest) msg);

            Credentials credentials = getCredentials(claim);

            return (ReqT) tenantRequest.toBuilder()
                    .setTenantId(claim.getTenantId()).setCredentials(credentials).build();
        } else if (method.equals("deleteTenant")) {

            AuthClaim claim = authorizeUsingUserToken(headers);

            DeleteTenantRequest tenantRequest = ((DeleteTenantRequest) msg);

            Credentials credentials = getCredentials(claim);


            return (ReqT) tenantRequest.toBuilder()
                    .setTenantId(claim.getTenantId()).setCredentials(credentials).build();
        } else if (method.equals("addTenantRoles")) {

            AuthClaim claim = authorizeUsingUserToken(headers);

            AddRolesRequest rolesRequest = ((AddRolesRequest) msg);
            String clientId = rolesRequest.getClientId();
            if (rolesRequest.getClientId() == null || rolesRequest.getClientId().trim().equals("")) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(claim.getTenantId()).
                        setClientId(claim.getCustosId()).build();
            }

            CredentialMetadata metadata = getCredentialsFromClientId(clientId);

            if (claim.isSuperTenant()) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
            }

            boolean validationStatus = validateParentChildTenantRelationShip(claim.getTenantId(), metadata.getOwnerId());

            if (validationStatus) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
            } else {
                String error = "Request is not authorized, user not authorized with requested clientId: " + clientId;
                throw new NotAuthorizedException(error, null);
            }
        } else if (method.equals("getTenantRoles")) {

            AuthClaim claim = authorizeUsingUserToken(headers);

            GetRolesRequest rolesRequest = ((GetRolesRequest) msg);
            String clientId = rolesRequest.getClientId();
            if (rolesRequest.getClientId() == null || rolesRequest.getClientId().trim().equals("")) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(claim.getTenantId()).
                        setClientId(claim.getCustosId()).build();
            }
            CredentialMetadata metadata = getCredentialsFromClientId(clientId);

            if (claim.isSuperTenant()) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
            }

            boolean validationStatus = validateParentChildTenantRelationShip(claim.getTenantId(), metadata.getOwnerId());

            if (validationStatus) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
            } else {
                String error = "Request is not authorized, user not authorized with requested clientId: " + clientId;
                throw new NotAuthorizedException(error, null);
            }
        } else if (method.equals("deleteRole")) {

            AuthClaim claim = authorizeUsingUserToken(headers);

            DeleteRoleRequest rolesRequest = ((DeleteRoleRequest) msg);
            String clientId = rolesRequest.getClientId();
            if (rolesRequest.getClientId() == null || rolesRequest.getClientId().trim().equals("")) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(claim.getTenantId()).
                        setClientId(claim.getCustosId()).build();
            }
            CredentialMetadata metadata = getCredentialsFromClientId(clientId);

            if (claim.isSuperTenant()) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
            }

            boolean validationStatus = validateParentChildTenantRelationShip(claim.getTenantId(), metadata.getOwnerId());

            if (validationStatus) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
            } else {
                String error = "Request is not authorized, user not authorized with requested clientId: " + clientId;
                throw new NotAuthorizedException(error, null);
            }
        } else if (method.equals("addProtocolMapper")) {

            AuthClaim claim = authorizeUsingUserToken(headers);

            AddProtocolMapperRequest rolesRequest = ((AddProtocolMapperRequest) msg);
            String clientId = rolesRequest.getClientId();

            if (rolesRequest.getClientId() == null || rolesRequest.getClientId().trim().equals("")) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(claim.getTenantId()).
                        setClientId(claim.getCustosId()).build();
            }
            CredentialMetadata metadata = getCredentialsFromClientId(clientId);

            if (claim.isSuperTenant()) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
            }

            boolean validationStatus = validateParentChildTenantRelationShip(claim.getTenantId(), metadata.getOwnerId());

            if (validationStatus) {
                return (ReqT) rolesRequest.toBuilder().setTenantId(metadata.getOwnerId()).build();
            } else {
                String error = "Request is not authorized, user not authorized with requested clientId: " + clientId;
                throw new NotAuthorizedException(error, null);
            }

        } else if (method.equals("configureEventPersistence")) {

            AuthClaim claim = authorizeUsingUserToken(headers);
            EventPersistenceRequest rolesRequest = ((EventPersistenceRequest) msg);

            return (ReqT) rolesRequest.toBuilder()
                    .setTenantId(claim.getTenantId()).setPerformedBy(claim.getUsername())
                    .build();
        } else if (method.equals("getChildTenants")) {

            AuthClaim claim = authorizeUsingUserToken(headers);

            GetTenantsRequest tenantsRequest = ((GetTenantsRequest) msg);
            String clientId = tenantsRequest.getParentClientId();

            if (tenantsRequest.getParentClientId() == null || tenantsRequest.getParentClientId().trim().equals("")) {
                return (ReqT) tenantsRequest.toBuilder().setParentId(claim.getTenantId()).build();
            }
            CredentialMetadata metadata = getCredentialsFromClientId(clientId);

            if (claim.isSuperTenant()) {
                return (ReqT) tenantsRequest.toBuilder().setParentId(metadata.getOwnerId()).build();
            }

            boolean validationStatus = validateParentChildTenantRelationShip(claim.getTenantId(), metadata.getOwnerId());

            if (validationStatus) {
                return (ReqT) tenantsRequest.toBuilder().setParentId(metadata.getOwnerId()).build();
            } else {
                String error = "Request is not authorized, user not authorized with requested clientId: " + clientId;
                throw new NotAuthorizedException(error, null);
            }
        } else if (method.equals("getAllTenantsForUser")) {
            validateAuth(headers);
            return msg;
        } else if (method.equals("getFromCache") || method.equals("getInstitutions")) {
            AuthClaim claim = validateAuth(headers);

            CacheManipulationRequest request = ((CacheManipulationRequest) msg);
            return (ReqT) request.toBuilder().setTenantId(claim.getTenantId()).build();

        } else if (method.equals("addToCache") || method.equals("removeFromCache")) {
            AuthClaim claim = validateAuth(headers);
            AuthClaim userClaim = validateUserToken(headers);

            CacheManipulationRequest.Builder request = ((CacheManipulationRequest) msg).toBuilder();

            if (userClaim != null) {
                request = request.setPerformedBy(userClaim.getUsername());
            }

            return (ReqT) request.setTenantId(claim.getTenantId()).build();

        }
        return msg;
    }


    private AuthClaim validateAuth(Metadata headers) {
        AuthClaim claim = null;
        try {
            claim = authorize(headers);
        } catch (Exception ex) {
            LOGGER.error(" Authorizing error " + ex.getMessage());
            throw new NotAuthorizedException("Request is not authorized", ex);
        }
        if (claim == null) {
            throw new NotAuthorizedException("Request is not authorized", null);
        }
        return claim;
    }


    private AuthClaim validateUserToken(Metadata headers) {
        AuthClaim claim = null;
        try {
            String usertoken = headers.get(Metadata.Key.of(Constants.USER_TOKEN, Metadata.ASCII_STRING_MARSHALLER));
            if (usertoken == null) {
                return null;
            }

            claim = authorizeUsingUserToken(usertoken);
        } catch (Exception ex) {
            LOGGER.error(" Authorizing error " + ex.getMessage());
            throw new NotAuthorizedException("Request is not authorized", ex);
        }
        if (claim == null) {
            throw new NotAuthorizedException("Request is not authorized", null);
        }
        return claim;


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
