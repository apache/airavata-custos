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

package org.apache.custos.integration.services.commons.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.credential.store.service.*;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.service.AuthToken;
import org.apache.custos.identity.service.Claim;
import org.apache.custos.identity.service.GetUserManagementSATokenRequest;
import org.apache.custos.identity.service.IsAuthenticatedResponse;
import org.apache.custos.integration.core.exceptions.UnAuthorizedException;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.GetTenantRequest;
import org.apache.custos.tenant.profile.service.GetTenantResponse;
import org.apache.custos.tenant.profile.service.Tenant;
import org.apache.custos.tenant.profile.service.TenantStatus;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicLong;
import java.util.concurrent.atomic.AtomicReference;

/**
 * Responsible for managing auth flow
 */
public abstract class AuthInterceptor implements IntegrationServiceInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(AuthInterceptor.class);


    private CredentialStoreServiceClient credentialStoreServiceClient;

    private TenantProfileClient tenantProfileClient;

    private IdentityClient identityClient;


    public AuthInterceptor(CredentialStoreServiceClient credentialStoreServiceClient,
                           TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        this.credentialStoreServiceClient = credentialStoreServiceClient;
        this.tenantProfileClient = tenantProfileClient;
        this.identityClient = identityClient;
    }

    public Optional<AuthClaim> authorize(Metadata headers) {
        try {
            String formattedToken = getToken(headers);
             LOGGER.info("Token "+ formattedToken);
            if (formattedToken == null) {
                throw new UnAuthorizedException(" token not found ", null);
            }
            return authorize(formattedToken);
        } catch (Exception ex) {
            throw new UnAuthorizedException(" invalid token " + ex.getMessage(), ex);
        }

    }

    public Optional<AuthClaim> authorize(String formattedToken) {
        try {
            TokenRequest request = TokenRequest
                    .newBuilder()
                    .setToken(formattedToken)
                    .build();
            GetAllCredentialsResponse response = credentialStoreServiceClient.getAllCredentialFromToken(request);
            return getAuthClaim(response);
        } catch (Exception ex) {
            throw new UnAuthorizedException(" invalid token " + ex.getMessage(), ex);
        }

    }


    public Optional<AuthClaim> authorizeUsingUserToken(Metadata headers) {

        try {
            String formattedToken = getToken(headers);

            if (formattedToken == null) {
                throw new UnAuthorizedException(" token not found ", null);
            }

            return authorizeUsingUserToken(formattedToken);
        } catch (Exception ex) {
            throw new UnAuthorizedException(" invalid token " + ex.getMessage(), ex);
        }

    }

    public Optional<AuthClaim> authorizeUsingUserToken(String formattedToken) {

        try {

            TokenRequest request = TokenRequest
                    .newBuilder()
                    .setToken(formattedToken)
                    .build();

            GetAllCredentialsResponse response = credentialStoreServiceClient.getAllCredentialsFromJWTToken(request);

            Optional<AuthClaim> claim = getAuthClaim(response);

            if (claim.isPresent()) {

                Claim userNameClaim = Claim.newBuilder()
                        .setKey("username")
                        .setValue(claim.get().getUsername()).build();

                Claim tenantClaim = Claim.newBuilder()
                        .setKey("tenantId")
                        .setValue(String.valueOf(claim.get().getTenantId()))
                        .build();


                List<Claim> claimList = new ArrayList<>();
                claimList.add(userNameClaim);
                claimList.add(tenantClaim);


                AuthToken token = AuthToken.newBuilder().
                        setAccessToken(formattedToken)
                        .addAllClaims(claimList)
                        .build();

                IsAuthenticatedResponse isAuthenticateResponse = identityClient.isAuthenticated(token);
                if (isAuthenticateResponse.getAuthenticated()) {
                    return claim;
                } else {
                    throw new UnAuthorizedException(" expired user token ", null);
                }
            } else {
                throw new UnAuthorizedException(" expired user token ", null);
            }
        } catch (Exception ex) {
            throw new UnAuthorizedException(ex.getMessage(), ex);
        }

    }


    public Optional<AuthClaim> authorizeUsingAgentBasicToken(Metadata headers, String parentId) {

        try {
            String formattedToken = getToken(headers);

            if (formattedToken == null) {
                throw new UnAuthorizedException(" token not found ", null);
            }

            TokenRequest request = TokenRequest
                    .newBuilder()
                    .setToken(formattedToken)
                    .setParentClientId(parentId)
                    .build();

            GetAllCredentialsResponse response = credentialStoreServiceClient.getCredentialByAgentBasicAuth(request);

            return getAuthClaim(response);

        } catch (Exception ex) {
            throw new UnAuthorizedException(" invalid token " + ex.getMessage(), ex);
        }

    }


    public Optional<AuthClaim> authorizeUsingAgentAndUserJWTTokens(Metadata headers) {
        try {

            String agentToken = getToken(headers);
            TokenRequest request = TokenRequest
                    .newBuilder()
                    .setToken(agentToken)
                    .build();

            GetAllCredentialsResponse response = credentialStoreServiceClient.getCredentialByAgentJWTToken(request);

            AtomicLong agentTenantId = new AtomicLong();
            AtomicReference<String> agentId = new AtomicReference<>();
            response.getSecretListList().forEach(sec -> {
                if (sec.getType() == Type.AGENT) {
                    agentTenantId.set(sec.getOwnerId());
                    agentId.set(sec.getId());
                }
            });

            Claim userNameClaim = Claim.newBuilder()
                    .setKey("username")
                    .setValue(agentId.get().toLowerCase()).build();

            Claim tenantClaim = Claim.newBuilder()
                    .setKey("tenantId")
                    .setValue(String.valueOf(agentTenantId.get()))
                    .build();


            List<Claim> claimList = new ArrayList<>();
            claimList.add(userNameClaim);
            claimList.add(tenantClaim);


            AuthToken token = AuthToken.newBuilder().
                    setAccessToken(agentToken)
                    .addAllClaims(claimList)
                    .build();


            IsAuthenticatedResponse isAuthenticateResponse = identityClient.isAuthenticated(token);
            if (isAuthenticateResponse.getAuthenticated()) {

                Optional<String> userToken = getUserTokenFromUserTokenHeader(headers);
                if (userToken.isPresent()) {
                    Optional<AuthClaim> claim = authorizeUsingUserToken(userToken.get());
                    if (claim.isPresent() && validateParentChildTenantRelationShip(agentTenantId.get(),
                            claim.get().getTenantId())) {
                        return claim;
                    } else {
                        throw new UnAuthorizedException("Agent" + agentId.get() + "" +
                                "  is not authorized to access resources on behalf of tenant ", null);
                    }
                } else {
                    throw new UnAuthorizedException("Agent" + agentId.get() + "" +
                            "  is not authorized to access resources on behalf of  user: user token not found ", null);
                }
            } else {
                throw new UnAuthorizedException("Agent" + agentId.get() + "" +
                        "  is not authorized ", null);
            }

        } catch (Exception ex) {
            throw new UnAuthorizedException(" invalid token " + ex.getMessage(), ex);
        }

    }


    public boolean isValidAgentJWTToken(Metadata headers) {

        try {
            String formattedToken = getToken(headers);

            if (formattedToken == null) {
                return false;
            }
            TokenRequest request = TokenRequest
                    .newBuilder()
                    .setToken(formattedToken)
                    .build();

            OperationStatus status = credentialStoreServiceClient.validateAgentJWTToken(request);

            return status.getState();

        } catch (Exception ex) {
            return false;
        }

    }

    public String getToken(Metadata headers) {
        String tokenWithBearer = headers.get(Metadata.Key.of("authorization", Metadata.ASCII_STRING_MARSHALLER));
        if (tokenWithBearer == null) {
            tokenWithBearer = headers.get(Metadata.Key.of("Authorization", Metadata.ASCII_STRING_MARSHALLER));
        }
        if (tokenWithBearer == null) {
            return null;
        }
        String prefix = "Bearer";
        String token = tokenWithBearer.substring(prefix.length());
        return token.trim();
    }


    /**
     * Authorize tenant request by checking validity of calling tenant and its child tenant given by clientId
     *
     * @param headers       parentTenant Headers
     * @param childClientId childTenant Headers
     * @return AuthClaim of child tenant
     */
    public Optional<AuthClaim> authorizeParentChildTenantValidationWithBasicAuth(Metadata headers, String childClientId) {
        Optional<AuthClaim> authClaim = authorize(headers);

        if (authClaim.isEmpty() || childClientId == null || childClientId.isEmpty()) {
            return authClaim;
        }

        GetCredentialRequest request = GetCredentialRequest
                .newBuilder()
                .setId(childClientId).build();


        CredentialMetadata metadata = credentialStoreServiceClient.getCustosCredentialFromClientId(request);


        GetAllCredentialsRequest credentialRequest = GetAllCredentialsRequest
                .newBuilder()
                .setOwnerId(metadata.getOwnerId())
                .build();

        GetAllCredentialsResponse allCredentials = credentialStoreServiceClient.getAllCredentials(credentialRequest);

        Optional<AuthClaim> childClaim = getAuthClaim(allCredentials);

        if (childClaim.isPresent() && (!validateTenantStatus(childClaim.get().getTenantId()) ||
                !validateParentChildTenantRelationShip(authClaim.get().getTenantId(), childClaim.get().getTenantId()))) {
            return Optional.empty();
        }

        return childClaim;

    }


    public Optional<AuthClaim> authorizeParentChildTenantWithBasicAuthAndUserTokenValidation(Metadata headers,
                                                                                             String childClientId,
                                                                                             String userToken) {
        Optional<AuthClaim> authClaim = authorize(headers);

        if (authClaim == null || childClientId == null || childClientId.isEmpty()) {
            return Optional.empty();
        }

        GetCredentialRequest request = GetCredentialRequest
                .newBuilder()
                .setId(childClientId).build();

        CredentialMetadata metadata = credentialStoreServiceClient
                .getCustosCredentialFromClientId(request);

        Optional<AuthClaim> optionalAuthClaim = getAuthClaim(metadata);

        if (optionalAuthClaim.isPresent() &&
                validateTenantStatus(optionalAuthClaim.get().getTenantId()) &&
                validateParentChildTenantRelationShip(authClaim.get().getTenantId(),
                        optionalAuthClaim.get().getTenantId())) {
            return authorizeUsingUserToken(userToken);
        }
        return Optional.empty();
    }

    public Optional<AuthClaim> authorizeParentChildTenantWithUserTokenValidation(Metadata headers,
                                                                                 String childClientId) {

        if (childClientId == null || childClientId.trim().isEmpty()) {
            return Optional.empty();
        }

        Optional<AuthClaim> authClaim = authorizeUsingUserToken(headers);

        GetCredentialRequest request = GetCredentialRequest
                .newBuilder()
                .setId(childClientId).build();

        CredentialMetadata metadata = credentialStoreServiceClient.getCustosCredentialFromClientId(request);


        if (!validateTenantStatus(metadata.getOwnerId()) ||
                (authClaim.isPresent() && !authClaim.get().isSuperTenant() &&
                        !validateParentChildTenantRelationShip(authClaim.get().getTenantId(),
                                metadata.getOwnerId()))) {
            return Optional.empty();
        }

        GetAllCredentialsRequest allReq = GetAllCredentialsRequest
                .newBuilder()
                .setOwnerId(metadata.getOwnerId())
                .build();
        GetAllCredentialsResponse response = credentialStoreServiceClient.getAllCredentials(allReq);

        return getAuthClaim(response);

    }


    public AuthToken getSAToken(String clientId, String clientSec, long tenantId) {
        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest
                .newBuilder()
                .setClientId(clientId)
                .setClientSecret(clientSec)
                .setTenantId(tenantId)
                .build();
        return identityClient.getUserManagementSATokenRequest(userManagementSATokenRequest);

    }


    public CredentialMetadata getCredentialsFromClientId(String clientId) {
        GetCredentialRequest request = GetCredentialRequest.newBuilder()
                .setId(clientId)
                .build();
        CredentialMetadata metadata = credentialStoreServiceClient.getCustosCredentialFromClientId(request);

        if (metadata == null || metadata.getOwnerId() == 0) {
            throw new UnAuthorizedException("Invalid client_id", null);
        }

        return metadata;
    }


    private Optional<AuthClaim> getAuthClaim(GetAllCredentialsResponse response) {
        if (response == null || response.getSecretListList().isEmpty()) {
            return Optional.empty();
        }

        AuthClaim authClaim = new AuthClaim();

        authClaim.setPerformedBy(response.getRequesterUserEmail());
        authClaim.setUsername(response.getRequesterUsername());
        response.getSecretListList().forEach(metadata -> {

                    if (metadata.getType() == Type.CUSTOS) {
                        authClaim.setTenantId(metadata.getOwnerId());
                        authClaim.setCustosId(metadata.getId());
                        authClaim.setCustosSecret(metadata.getSecret());
                        authClaim.setCustosIdIssuedAt(metadata.getClientIdIssuedAt());
                        authClaim.setCustosSecretExpiredAt(metadata.getClientSecretExpiredAt());
                        authClaim.setAdmin(metadata.getSuperAdmin());
                        authClaim.setSuperTenant(metadata.getSuperTenant());
                    } else if (metadata.getType() == Type.IAM) {
                        authClaim.setIamAuthId(metadata.getId());
                        authClaim.setIamAuthSecret(metadata.getSecret());

                    } else if (metadata.getType() == Type.CILOGON) {
                        authClaim.setCiLogonId(metadata.getId());
                        authClaim.setCiLogonSecret(metadata.getSecret());
                    } else if (metadata.getType() == Type.AGENT_CLIENT) {
                        authClaim.setAgentClientId(metadata.getId());
                        authClaim.setAgentClientSecret(metadata.getSecret());
                    } else if (metadata.getType() == Type.AGENT) {
                        authClaim.setAgentId(metadata.getId());
                        authClaim.setAgentPassword(metadata.getInternalSec());
                    }

                }

        );


        GetTenantRequest tenantRequest = GetTenantRequest
                .newBuilder()
                .setTenantId(authClaim.getTenantId())
                .build();

        GetTenantResponse tentResp = tenantProfileClient.getTenant(tenantRequest);

        if (tentResp.getTenant() != null && tentResp.getTenant().getTenantStatus().equals(TenantStatus.ACTIVE)) {
            return Optional.of(authClaim);
        }
        return Optional.empty();
    }


    private Optional<AuthClaim> getAuthClaim(CredentialMetadata metadata) {
        AuthClaim authClaim = new AuthClaim();
        if (metadata.getType() == Type.CUSTOS) {
            authClaim.setTenantId(metadata.getOwnerId());
            authClaim.setCustosId(metadata.getId());
            authClaim.setCustosSecret(metadata.getSecret());
            authClaim.setCustosIdIssuedAt(metadata.getClientIdIssuedAt());
            authClaim.setCustosSecretExpiredAt(metadata.getClientSecretExpiredAt());
            authClaim.setAdmin(metadata.getSuperAdmin());
            authClaim.setSuperTenant(metadata.getSuperTenant());
        } else if (metadata.getType() == Type.IAM) {
            authClaim.setIamAuthId(metadata.getId());
            authClaim.setIamAuthSecret(metadata.getSecret());

        } else if (metadata.getType() == Type.CILOGON) {
            authClaim.setCiLogonId(metadata.getId());
            authClaim.setCiLogonSecret(metadata.getSecret());
        } else if (metadata.getType() == Type.AGENT_CLIENT) {
            authClaim.setAgentClientId(metadata.getId());
            authClaim.setAgentClientSecret(metadata.getSecret());
        } else if (metadata.getType() == Type.AGENT) {
            authClaim.setAgentId(metadata.getId());
            authClaim.setAgentPassword(metadata.getInternalSec());
        }
        return Optional.of(authClaim);

    }


    private boolean validateTenantStatus(long tenantId) {
        GetTenantRequest tenantRequest = GetTenantRequest
                .newBuilder()
                .setTenantId(tenantId)
                .build();

        GetTenantResponse tentResp = tenantProfileClient.getTenant(tenantRequest);

        if (tentResp.getTenant() != null && tentResp.getTenant().getTenantStatus().equals(TenantStatus.ACTIVE)) {
            return true;
        }
        return false;
    }


    public boolean validateParentChildTenantRelationShip(long parentId, long childTenantId) {

        GetTenantRequest childTenantReq = GetTenantRequest
                .newBuilder()
                .setTenantId(childTenantId)
                .build();

        GetTenantResponse childTenantRes = tenantProfileClient.getTenant(childTenantReq);

        Tenant childTenant = childTenantRes.getTenant();


        // referring to same tenant
        if (childTenant != null && childTenant.getTenantId() == parentId) {
            return true;
        }

        //referring to child tenant
        if (childTenant != null && childTenant.getTenantId() != parentId && childTenant.getParentTenantId() == parentId) {
            return true;
        }

        return false;
    }


    private void attachTenantId(String tenantId, Metadata headers) {
        headers.put(Metadata.Key.of(Constants.TenantId, Metadata.ASCII_STRING_MARSHALLER), tenantId);
    }

    public Optional<String> getUserTokenFromUserTokenHeader(Metadata headers) {
        String header = headers.get(Metadata.Key.of(Constants.USER_TOKEN, Metadata.ASCII_STRING_MARSHALLER));
        if (header != null && !header.trim().isEmpty()) {
            return Optional.of(header);
        } else {
            return Optional.empty();
        }

    }

    public boolean clearUserTokenFromHeader(Metadata headers) {
        String obj = headers.get(Metadata.Key.of(Constants.USER_TOKEN, Metadata.ASCII_STRING_MARSHALLER));
        headers.remove(Metadata.Key.of(Constants.USER_TOKEN, Metadata.ASCII_STRING_MARSHALLER), obj);
        return true;
    }

    public boolean isAgentAuthenticationEnabled(Metadata headers) {
        String header = headers.get(Metadata.Key.of(Constants.AUTHENTICATE_AGENT, Metadata.ASCII_STRING_MARSHALLER));
        if (header != null && Boolean.parseBoolean(header)) {
            return true;
        } else {
            return false;
        }
    }

    public boolean isBasicAuth(Metadata headers) {
        try {
            this.authorize(headers);
            return true;
        } catch (Exception ex) {
            return false;
        }
    }

    public boolean isUserToken(Metadata headers) {
        try {
            this.authorizeUsingUserToken(headers);
            return true;
        } catch (Exception ex) {
            return false;
        }
    }

}
