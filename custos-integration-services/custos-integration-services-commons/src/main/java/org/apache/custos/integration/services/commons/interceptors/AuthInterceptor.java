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
import org.apache.custos.identity.service.IsAuthenticateResponse;
import org.apache.custos.integration.core.exceptions.NotAuthorizedException;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.apache.custos.tenant.profile.service.GetTenantRequest;
import org.apache.custos.tenant.profile.service.GetTenantResponse;
import org.apache.custos.tenant.profile.service.Tenant;
import org.apache.custos.tenant.profile.service.TenantStatus;
import org.apache.tomcat.util.bcel.Const;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.List;

/**
 * Responsible for managing auth flow
 */
public abstract class AuthInterceptor implements IntegrationServiceInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(AuthInterceptor.class);


    private CredentialStoreServiceClient credentialStoreServiceClient;

    private TenantProfileClient tenantProfileClient;

    private IdentityClient identityClient;


    public AuthInterceptor(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        this.credentialStoreServiceClient = credentialStoreServiceClient;
        this.tenantProfileClient = tenantProfileClient;
        this.identityClient = identityClient;
    }

    public AuthClaim authorize(Metadata headers) {
        try {
            String formattedToken = getToken(headers);

            if (formattedToken == null) {
                return null;
            }
            return authorize(formattedToken);
        } catch (Exception ex) {
            throw new NotAuthorizedException("Wrong credentials " + ex.getMessage(), ex);
        }

    }

    public AuthClaim authorize(String formattedToken) {
        try {
            TokenRequest request = TokenRequest
                    .newBuilder()
                    .setToken(formattedToken)
                    .build();
            GetAllCredentialsResponse response = credentialStoreServiceClient.getAllCredentialFromToken(request);
            return getAuthClaim(response);
        } catch (Exception ex) {
            throw new NotAuthorizedException("Wrong credentials " + ex.getMessage(), ex);
        }

    }


    public AuthClaim authorizeUsingUserToken(Metadata headers) {

        try {
            String formattedToken = getToken(headers);

            if (formattedToken == null) {
                return null;
            }

            return authorizeUsingUserToken(formattedToken);
        } catch (Exception ex) {
            throw new NotAuthorizedException("Wrong credentials " + ex.getMessage(), ex);
        }

    }

    public AuthClaim authorizeUsingUserToken(String formattedToken) {

        try {

            TokenRequest request = TokenRequest
                    .newBuilder()
                    .setToken(formattedToken)
                    .build();

            GetAllCredentialsResponse response = credentialStoreServiceClient.getAllCredentialsFromJWTToken(request);

            AuthClaim claim = getAuthClaim(response);

            if (claim != null) {

                Claim userNameClaim = Claim.newBuilder()
                        .setKey("username")
                        .setValue(claim.getUsername()).build();

                Claim tenantClaim = Claim.newBuilder()
                        .setKey("tenantId")
                        .setValue(String.valueOf(claim.getTenantId()))
                        .build();


                List<Claim> claimList = new ArrayList<>();
                claimList.add(userNameClaim);
                claimList.add(tenantClaim);


                AuthToken token = AuthToken.newBuilder().
                        setAccessToken(formattedToken)
                        .addAllClaims(claimList)
                        .build();


                IsAuthenticateResponse isAuthenticateResponse = identityClient.isAuthenticated(token);
                if (isAuthenticateResponse.getAuthenticated()) {
                    return claim;
                }
            }

            return null;
        } catch (Exception ex) {
            throw new NotAuthorizedException("Wrong credentials " + ex.getMessage(), ex);
        }

    }


    public AuthClaim authorizeUsingAgentBasicToken(Metadata headers, String parentId) {

        try {
            String formattedToken = getToken(headers);

            if (formattedToken == null) {
                return null;
            }

            TokenRequest request = TokenRequest
                    .newBuilder()
                    .setToken(formattedToken)
                    .setParentClientId(parentId)
                    .build();

            GetAllCredentialsResponse response = credentialStoreServiceClient.getCredentialByAgentBasicAuth(request);

            return getAuthClaim(response);

        } catch (Exception ex) {
            throw new NotAuthorizedException("Wrong credentials " + ex.getMessage(), ex);
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
    public AuthClaim authorizeWithParentChildTenantValidationByBasicAuth(Metadata headers, String childClientId) {
        AuthClaim authClaim = authorize(headers);

        if (authClaim == null) {
            return null;
        }

        if (childClientId == null || childClientId.trim().equals("")) {
            return authClaim;
        }

        GetCredentialRequest request = GetCredentialRequest
                .newBuilder()
                .setId(childClientId).build();


        CredentialMetadata metadata = credentialStoreServiceClient.getCustosCredentialFromClientId(request);


        GetCredentialRequest credentialRequest = GetCredentialRequest
                .newBuilder()
                .setOwnerId(metadata.getOwnerId())
                .setType(Type.IAM).build();

        CredentialMetadata iamCredentials = credentialStoreServiceClient.getCredential(credentialRequest);

        AuthClaim childClaim = getAuthClaim(metadata);

        childClaim.setIamAuthSecret(iamCredentials.getSecret());
        childClaim.setIamAuthId(iamCredentials.getId());

        boolean statusValidation = validateTenantStatus(childClaim.getTenantId());

        if (!statusValidation) {
            return null;
        }

        boolean relationShipValidation = validateParentChildTenantRelationShip(authClaim.getTenantId(), childClaim.getTenantId());

        if (!relationShipValidation) {
            return null;
        }

        return childClaim;

    }


    public AuthClaim authorizeWithParentChildTenantValidationByBasicAuthAndUserTokenValidation(Metadata headers,
                                                                                               String childClientId,
                                                                                               String userToken) {
        AuthClaim authClaim = authorize(headers);

        if (authClaim == null) {
            return null;
        }

        if (childClientId == null || childClientId.trim().equals("")) {
            return authClaim;
        }

        GetCredentialRequest request = GetCredentialRequest
                .newBuilder()
                .setId(childClientId).build();

        CredentialMetadata metadata = credentialStoreServiceClient.getCustosCredentialFromClientId(request);

        AuthClaim childClaim = getAuthClaim(metadata);


        boolean statusValidation = validateTenantStatus(childClaim.getTenantId());

        if (!statusValidation) {
            return null;
        }

        boolean relationShipValidation = validateParentChildTenantRelationShip(authClaim.getTenantId(), childClaim.getTenantId());

        if (!relationShipValidation) {
            return null;
        }

        return authorizeUsingUserToken(userToken);

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


    private AuthClaim getAuthClaim(GetAllCredentialsResponse response) {
        if (response == null || response.getSecretListCount() == 0) {
            return null;
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
            return authClaim;
        }
        return null;
    }


    private AuthClaim getAuthClaim(CredentialMetadata metadata) {
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
        return authClaim;

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


    private boolean validateParentChildTenantRelationShip(long parentId, long childTenantId) {

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
        headers.put(Metadata.Key.of(Constants.TenantId,Metadata.ASCII_STRING_MARSHALLER), tenantId);
    }


}
