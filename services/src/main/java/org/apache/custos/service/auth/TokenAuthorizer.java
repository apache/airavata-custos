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

package org.apache.custos.service.auth;

import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.credential.store.api.CredentialMetadata;
import org.apache.custos.core.credential.store.api.GetAllCredentialsRequest;
import org.apache.custos.core.credential.store.api.GetAllCredentialsResponse;
import org.apache.custos.core.credential.store.api.GetCredentialRequest;
import org.apache.custos.core.credential.store.api.TokenRequest;
import org.apache.custos.core.credential.store.api.Type;
import org.apache.custos.core.exception.UnauthorizedException;
import org.apache.custos.core.identity.api.AuthToken;
import org.apache.custos.core.identity.api.Claim;
import org.apache.custos.core.identity.api.GetUserManagementSATokenRequest;
import org.apache.custos.core.identity.api.IsAuthenticatedResponse;
import org.apache.custos.core.tenant.profile.api.GetTenantRequest;
import org.apache.custos.core.tenant.profile.api.GetTenantResponse;
import org.apache.custos.core.tenant.profile.api.Tenant;
import org.apache.custos.core.tenant.profile.api.TenantStatus;
import org.apache.custos.service.credential.store.CredentialStoreService;
import org.apache.custos.service.identity.IdentityService;
import org.apache.custos.service.profile.TenantProfileService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@Service
public class TokenAuthorizer {

    private static final Logger LOGGER = LoggerFactory.getLogger(TokenAuthorizer.class);

    private final CredentialStoreService credentialStoreService;
    private final TenantProfileService tenantProfileService;
    private final IdentityService identityService;


    public TokenAuthorizer(CredentialStoreService credentialStoreService, TenantProfileService tenantProfileService, IdentityService identityService) {
        this.credentialStoreService = credentialStoreService;
        this.tenantProfileService = tenantProfileService;
        this.identityService = identityService;
    }

    public Optional<AuthClaim> authorize(HttpHeaders headers) {
        try {
            String formattedToken = getToken(headers);

            if (formattedToken == null) {
                throw new UnauthorizedException("Token not found", null);
            }
            return authorize(formattedToken);
        } catch (Exception ex) {
            throw new UnauthorizedException("Invalid token: " + ex.getMessage(), ex);
        }
    }

    public String getToken(HttpHeaders headers) {
        String tokenWithBearer = headers.getFirst("Authorization");
        if (tokenWithBearer == null) {
            return null;
        }
        String prefix = "Bearer ";
        if (tokenWithBearer.startsWith(prefix)) {
            String token = tokenWithBearer.substring(prefix.length());
            return token.trim();
        }
        return null;
    }

    public Optional<AuthClaim> authorize(String formattedToken) {
        try {
            TokenRequest request = TokenRequest.newBuilder()
                    .setToken(formattedToken)
                    .build();
            GetAllCredentialsResponse response = credentialStoreService.getAllCredentialsFromToken(request);
            return getAuthClaim(response);
        } catch (Exception ex) {
            throw new UnauthorizedException("Invalid token: " + ex.getMessage(), ex);
        }
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

        GetTenantRequest tenantRequest = GetTenantRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .build();

        GetTenantResponse tentResp = tenantProfileService.getTenant(tenantRequest);

        if (tentResp.getTenant().getTenantStatus().equals(TenantStatus.ACTIVE)) {
            return Optional.of(authClaim);
        }
        return Optional.empty();
    }

    public Optional<AuthClaim> authorizeUsingUserToken(String formattedToken) {
        try {
            TokenRequest request = TokenRequest.newBuilder()
                    .setToken(formattedToken)
                    .build();

            GetAllCredentialsResponse response = credentialStoreService.getAllCredentialsFromJWTToken(request);
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

                IsAuthenticatedResponse isAuthenticateResponse = identityService.isAuthenticated(token);
                if (isAuthenticateResponse.getAuthenticated()) {
                    return claim;
                } else {
                    throw new UnauthorizedException("expired user token ", null);
                }

            } else {
                throw new UnauthorizedException("expired user token ", null);
            }

        } catch (Exception ex) {
            throw new UnauthorizedException(ex.getMessage(), ex);
        }
    }

    public Optional<AuthClaim> authorize(HttpHeaders headers, String clientId) {
        try {
            if (clientId != null && clientId.trim().isEmpty()) {
                clientId = null;
            }

            Optional<String> userToken = getUserTokenFromUserTokenHeader(headers);
            boolean isBasicAuth = isBasicAuth(headers);

            if (clientId == null && userToken.isEmpty() && isBasicAuth) {
                return authorize(headers);

            } else if (clientId != null && userToken.isEmpty() && isBasicAuth) {
                return authorizeParentChildTenantValidationWithBasicAuth(headers, clientId);

            } else if (clientId != null && userToken.isPresent()) {
                return authorizeParentChildTenantWithBasicAuthAndUserTokenValidation(headers, clientId, userToken.get());

            } else if (clientId != null && isUserToken(headers)) {
                return authorizeParentChildTenantWithUserTokenValidation(headers, clientId);

            } else {
                return authorizeUsingUserToken(headers);
            }

        } catch (Exception ex) {
            LOGGER.error("Error while generating AuthClaims for authorize", ex);
            throw ex;
        }
    }

    public Optional<String> getUserTokenFromUserTokenHeader(HttpHeaders headers) {
        String header = headers.getFirst(Constants.USER_TOKEN);
        if (header != null && !header.trim().isEmpty()) {
            return Optional.of(header);
        } else {
            return Optional.empty();
        }
    }

    public boolean isBasicAuth(HttpHeaders headers) {
        try {
            this.authorize(headers);
            return true;
        } catch (Exception ex) {
            return false;
        }
    }

    public Optional<AuthClaim> authorizeParentChildTenantValidationWithBasicAuth(HttpHeaders headers, String childClientId) {
        Optional<AuthClaim> authClaim = authorize(headers);

        if (authClaim.isEmpty() || childClientId == null || childClientId.isEmpty()) {
            return authClaim;
        }

        GetCredentialRequest request = GetCredentialRequest.newBuilder().setId(childClientId).build();
        CredentialMetadata metadata = credentialStoreService.getCustosCredentialFromClientId(request);
        GetAllCredentialsRequest credentialRequest = GetAllCredentialsRequest.newBuilder().setOwnerId(metadata.getOwnerId()).build();
        GetAllCredentialsResponse allCredentials = credentialStoreService.getAllCredentials(credentialRequest);

        Optional<AuthClaim> childClaim = getAuthClaim(allCredentials);

        if (childClaim.isPresent() && !authClaim.get().isSuperTenant() && (!validateTenantStatus(childClaim.get().getTenantId()) ||
                !validateParentChildTenantRelationShip(authClaim.get().getTenantId(), childClaim.get().getTenantId()))) {
            return Optional.empty();
        }

        return childClaim;
    }

    public boolean validateParentChildTenantRelationShip(long parentId, long childTenantId) {
        GetTenantRequest childTenantReq = GetTenantRequest.newBuilder().setTenantId(childTenantId).build();
        GetTenantResponse childTenantRes = tenantProfileService.getTenant(childTenantReq);
        Tenant childTenant = childTenantRes.getTenant();

        // referring to same tenant
        if (childTenant.getTenantId() == parentId) {
            return true;
        }

        //referring to child tenant
        return childTenant.getTenantId() != parentId && childTenant.getParentTenantId() == parentId;
    }

    public Optional<AuthClaim> authorizeParentChildTenantWithBasicAuthAndUserTokenValidation(HttpHeaders headers, String childClientId, String userToken) {
        Optional<AuthClaim> authClaim = authorize(headers);

        if (authClaim.isEmpty() || childClientId == null || childClientId.isEmpty()) {
            return Optional.empty();
        }

        GetCredentialRequest request = GetCredentialRequest.newBuilder().setId(childClientId).build();
        CredentialMetadata metadata = credentialStoreService.getCustosCredentialFromClientId(request);
        Optional<AuthClaim> optionalAuthClaim = getAuthClaim(metadata);

        if (optionalAuthClaim.isPresent() &&
                validateTenantStatus(optionalAuthClaim.get().getTenantId()) &&
                validateParentChildTenantRelationShip(authClaim.get().getTenantId(), optionalAuthClaim.get().getTenantId())) {
            return authorizeUsingUserToken(userToken);
        }
        return Optional.empty();
    }

    public Optional<AuthClaim> authorizeParentChildTenantWithUserTokenValidation(HttpHeaders headers, String childClientId) {
        if (childClientId == null || childClientId.trim().isEmpty()) {
            return Optional.empty();
        }

        Optional<AuthClaim> authClaim = authorizeUsingUserToken(headers);
        GetCredentialRequest request = GetCredentialRequest.newBuilder().setId(childClientId).build();
        CredentialMetadata metadata = credentialStoreService.getCustosCredentialFromClientId(request);

        if (!validateTenantStatus(metadata.getOwnerId()) ||
                (authClaim.isPresent() && !authClaim.get().isSuperTenant() &&
                        !validateParentChildTenantRelationShip(authClaim.get().getTenantId(), metadata.getOwnerId()))) {
            return Optional.empty();
        }

        GetAllCredentialsRequest allReq = GetAllCredentialsRequest.newBuilder().setOwnerId(metadata.getOwnerId()).build();
        GetAllCredentialsResponse response = credentialStoreService.getAllCredentials(allReq);

        return getAuthClaim(response);

    }

    public Optional<AuthClaim> authorizeUsingUserToken(HttpHeaders headers) {
        try {
            String formattedToken = getToken(headers);

            if (formattedToken == null) {
                throw new UnauthorizedException("token not found ", null);
            }
            return authorizeUsingUserToken(formattedToken);

        } catch (Exception ex) {
            throw new UnauthorizedException("invalid token " + ex.getMessage(), ex);
        }
    }

    public boolean isUserToken(HttpHeaders headers) {
        try {
            this.authorizeUsingUserToken(headers);
            return true;
        } catch (Exception ex) {
            return false;
        }
    }

    public AuthToken getSAToken(String clientId, String clientSec, long tenantId) {
        GetUserManagementSATokenRequest userManagementSATokenRequest = GetUserManagementSATokenRequest.newBuilder()
                .setClientId(clientId)
                .setClientSecret(clientSec)
                .setTenantId(tenantId)
                .build();
        return identityService.getUserManagementServiceAccountAccessToken(userManagementSATokenRequest);
    }

    public CredentialMetadata getCredentialsFromClientId(String clientId) {
        GetCredentialRequest request = GetCredentialRequest.newBuilder()
                .setId(clientId)
                .build();
        CredentialMetadata metadata = credentialStoreService.getCustosCredentialFromClientId(request);

        if (metadata == null || metadata.getOwnerId() == 0) {
            throw new UnauthorizedException("Invalid client_id", null);
        }

        return metadata;
    }

    public Optional<AuthClaim> validateUserToken(HttpHeaders headers) {
        try {
            String userToken = headers.getFirst(Constants.USER_TOKEN);
            if (userToken == null) {
                return Optional.empty();
            }
            return authorizeUsingUserToken(userToken);

        } catch (Exception ex) {
            LOGGER.error("Authorizing error " + ex.getMessage());
            throw new UnauthorizedException("Request is not authorized", ex);
        }
    }

    private boolean validateTenantStatus(long tenantId) {
        GetTenantRequest tenantRequest = GetTenantRequest.newBuilder().setTenantId(tenantId).build();
        GetTenantResponse tentResp = tenantProfileService.getTenant(tenantRequest);

        return tentResp.getTenant().getTenantStatus().equals(TenantStatus.ACTIVE);
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
}
