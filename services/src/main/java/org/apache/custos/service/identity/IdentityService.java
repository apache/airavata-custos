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

package org.apache.custos.service.identity;


import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.exception.UnauthorizedException;
import org.apache.custos.core.identity.api.AuthToken;
import org.apache.custos.core.identity.api.AuthenticationRequest;
import org.apache.custos.core.identity.api.AuthorizationResponse;
import org.apache.custos.core.identity.api.Claim;
import org.apache.custos.core.identity.api.EndSessionRequest;
import org.apache.custos.core.identity.api.GetAuthorizationEndpointRequest;
import org.apache.custos.core.identity.api.GetJWKSRequest;
import org.apache.custos.core.identity.api.GetOIDCConfiguration;
import org.apache.custos.core.identity.api.GetTokenRequest;
import org.apache.custos.core.identity.api.GetUserManagementSATokenRequest;
import org.apache.custos.core.identity.api.IsAuthenticatedResponse;
import org.apache.custos.core.identity.api.OIDCConfiguration;
import org.apache.custos.core.identity.api.OperationStatus;
import org.apache.custos.core.identity.api.TokenResponse;
import org.apache.custos.core.identity.authzcache.AuthzCacheEntry;
import org.apache.custos.core.identity.authzcache.AuthzCacheIndex;
import org.apache.custos.core.identity.authzcache.AuthzCachedStatus;
import org.apache.custos.core.identity.authzcache.DefaultAuthzCacheManager;
import org.apache.custos.core.identity.exceptions.AuthSecurityException;
import org.apache.custos.service.auth.TokenService;
import org.apache.custos.service.federated.client.keycloak.auth.KeycloakAuthClient;
import org.apache.http.HttpStatus;
import org.json.JSONArray;
import org.json.JSONObject;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.Collections;
import java.util.List;
import java.util.stream.IntStream;


/**
 * A service responsible for authenticate and authorize users
 */
@Service
public class IdentityService {

    private static final Logger LOGGER = LoggerFactory.getLogger(IdentityService.class);

    private static final int CACHE_LIFE_TIME = 1000 * 60 * 60;

    @Autowired
    private KeycloakAuthClient keycloakAuthClient;

    @Autowired
    private DefaultAuthzCacheManager authzCacheManager;

    @Autowired
    private TokenService tokenService;

    @Value("${custos.identity.auth.cache.enabled:false}")
    private boolean isAuthzCacheEnabled;

    @Value("${custos.api.domain}")
    private String apiDomain;

    public AuthToken authenticate(AuthenticationRequest request) {
        try {
            LOGGER.debug("Authentication request received for " + request.getUsername());
            String accessToken = keycloakAuthClient.authenticate(request.getClientId(), request.getClientSecret(),
                    String.valueOf(request.getTenantId()), request.getUsername(), request.getPassword());

            if (accessToken != null) {
                AuthToken.Builder authzBuilder = AuthToken.newBuilder().setAccessToken(accessToken);
                return authzBuilder.build();
            } else {
                throw new RuntimeException("Could not obtain an access token from the Keycloak");
            }

        } catch (org.keycloak.authorization.client.util.HttpResponseException ex) {
            String msg = "Error occurred while authenticating  user " + request.getUsername() + " " + ex.getReasonPhrase();
            LOGGER.error(msg);

            if (ex.getStatusCode() == HttpStatus.SC_UNAUTHORIZED) {
                throw new UnauthorizedException(msg);
            }
            throw new RuntimeException(msg);

        } catch (Exception ex) {
            String msg = "Error occurred while authenticating  user " + request.getUsername() + " " + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }


    public org.apache.custos.core.identity.api.User getUser(AuthToken request) {
        String username = null;
        String tenantId = null;
        String custosId = null;
        try {

            for (Claim claim : request.getClaimsList()) {
                switch (claim.getKey()) {
                    case "username" -> username = claim.getValue();
                    case "tenantId" -> tenantId = claim.getValue();
                    case "clientId" -> custosId = claim.getValue();
                }
            }

            LOGGER.debug("Get user request received for " + username);

            org.apache.custos.service.federated.client.keycloak.auth.User user = keycloakAuthClient.getUser(request.getAccessToken(), tenantId);
            return org.apache.custos.core.identity.api.User.newBuilder()
                    .setEmailAddress(user.getEmailAddress())
                    .setFirstName(user.getFirstName())
                    .setLastName(user.getLastName())
                    .setFullName(user.getFullName())
                    .setSub(user.getSub())
                    .setUsername(user.getUsername())
                    .setClientId(custosId != null ? custosId : "")
                    .build();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching  user " + username + " " + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }

    public IsAuthenticatedResponse isAuthenticated(AuthToken request) {
        String username = null;
        String tenantId = null;

        for (Claim claim : request.getClaimsList()) {
            switch (claim.getKey()) {
                case "username" -> username = claim.getValue();
                case "tenantId" -> tenantId = claim.getValue();
            }
        }

        LOGGER.debug("Authentication status checking for  " + username);
        LOGGER.debug("Authentication status checking for  " + username + " token " + request.getAccessToken());

        String accessToken = request.getAccessToken();

        boolean isAuthenticated;

        try {
            if (isAuthzCacheEnabled) {
                //check in the cache
                AuthzCachedStatus authzCachedStatus = authzCacheManager.getAuthzCachedStatus(new AuthzCacheIndex(username, tenantId, accessToken));

                String authzDecisionCacheLog = "Authz decision for: ({}, {}) {} cache.";
                switch (authzCachedStatus) {
                    case AUTHORIZED -> {
                        LOGGER.debug(authzDecisionCacheLog, username, accessToken, "is retrieved from");
                        isAuthenticated = true;
                    }
                    case NOT_AUTHORIZED -> {
                        LOGGER.debug(authzDecisionCacheLog, username, accessToken, "is retrieved from");
                        isAuthenticated = false;
                    }
                    case NOT_CACHED -> {
                        LOGGER.debug(authzDecisionCacheLog, username, accessToken, "is not in the");
                        LOGGER.info("Executing is User Authenticated");
                        isAuthenticated = keycloakAuthClient.isUserAuthenticated(username, tenantId, accessToken);
                        // cache the authorization decision
                        long currentTime = System.currentTimeMillis();
                        authzCacheManager.addToAuthzCache(
                                new AuthzCacheIndex(username, tenantId, accessToken),
                                new AuthzCacheEntry(isAuthenticated, currentTime + CACHE_LIFE_TIME, currentTime));
                    }
                    default -> throw new AuthSecurityException("Error in reading from the authorization cache.");
                }

            } else {
                isAuthenticated = keycloakAuthClient.isUserAuthenticated(username, tenantId, tokenService.getKCToken(accessToken));
            }

            if (isAuthenticated) {
                LOGGER.debug("User" + username + "in gateway" + tenantId + "is authenticated");

            } else {
                LOGGER.debug("User" + username + "in gateway" + tenantId + "is not authenticated");
            }

            return IsAuthenticatedResponse
                    .newBuilder()
                    .setAuthenticated(isAuthenticated)
                    .build();

        } catch (Exception ex) {
            String msg = "Error occurred while validating authentication status of  user " + username + " " + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg);
        }
    }

    public AuthToken getUserManagementServiceAccountAccessToken(GetUserManagementSATokenRequest request) {
        try {
            LOGGER.debug("Retrieving service account access token for " + request.getClientId());
            String accessToken = keycloakAuthClient.getUserManagementServiceAccountAccessToken(request.getClientId(),
                    request.getClientSecret(), String.valueOf(request.getTenantId()));

            AuthToken.Builder builder = AuthToken.newBuilder().setAccessToken(accessToken);

            Claim userClaim = Claim.newBuilder().setKey("username").setValue(request.getClientId()).build();
            builder.addClaims(userClaim);

            return builder.build();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching access token for  user " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }

    public TokenResponse getToken(GetTokenRequest request) {
        try {
            LOGGER.debug("Request token for " + request.getTenantId());

            JSONObject object;
            String grantType = request.getGrantType();
            String clientId = request.getClientId();
            String clientSecret = request.getClientSecret();
            String tenantId = String.valueOf(request.getTenantId());

            object = switch (grantType) {
                case Constants.PASSWORD_GRANT_TYPE ->
                        keycloakAuthClient.getAccessTokenFromPasswordGrantType(clientId, clientSecret, tenantId, request.getUsername(), request.getPassword());
                case Constants.REFRESH_TOKEN ->
                        keycloakAuthClient.getAccessTokenFromRefreshTokenGrantType(clientId, clientSecret, tenantId, request.getRefreshToken());
                case Constants.CLIENT_CREDENTIALS ->
                        keycloakAuthClient.getAccessTokenFromClientCredentialsGrantType(clientId, clientSecret, tenantId);
                default ->
                        keycloakAuthClient.getAccessToken(clientId, clientSecret, tenantId, request.getCode(), request.getRedirectUri(), request.getCodeVerifier());
            };

            try {
                return generateTokenResponse(object);

            } catch (Exception ex) {
                String msg = object != null
                        ? object.getString("error") + " " + object.getString("error_description")
                        : "Error while extracting the token from the Keycloak";
                LOGGER.error(msg, ex);
                throw new RuntimeException(msg, ex);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching access token for  user " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public AuthorizationResponse getAuthorizeEndpoint(GetAuthorizationEndpointRequest request) {
        try {
            LOGGER.debug("Request authorization endpoint for " + request.getTenantId());

            String authEndpoint = keycloakAuthClient.getAuthorizationEndpoint(String.valueOf(request.getTenantId()));
            return AuthorizationResponse.newBuilder().setAuthorizationEndpoint(authEndpoint).build();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching access token for  user " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }

    public OIDCConfiguration getOIDCConfiguration(GetOIDCConfiguration request) {
        try {
            LOGGER.debug("Request for fetch OIDC configuration " + request.getTenantId());

            JSONObject object = keycloakAuthClient.getOIDCConfiguration(String.valueOf(request.getTenantId()), request.getClientId());

            return OIDCConfiguration.newBuilder()
                    .setIssuer("https://" + request.getTenantId() + ".usecustos.org")
                    .setAuthorizationEndpoint(apiDomain + "/api/v1/identity-management/authorize")
                    .setTokenEndpoint(apiDomain + "/api/v1/identity-management/token")
                    .setUserinfoEndpoint(apiDomain + "/api/v1/user-management/userinfo")
                    .setJwksUri(apiDomain + "/api/v1/identity-management/.well-known/jwks.json")
                    .addAllResponseTypesSupported(jsonArrayToList(object.getJSONArray("response_types_supported")))
                    .addAllSubjectTypesSupported(jsonArrayToList(object.getJSONArray("subject_types_supported")))
                    .addAllIdTokenSigningAlgValuesSupported(jsonArrayToList(object.getJSONArray("id_token_signing_alg_values_supported")))
                    .addAllScopesSupported(jsonArrayToList(object.getJSONArray("scopes_supported")))
                    .addAllTokenEndpointAuthMethodsSupported(jsonArrayToList(object.getJSONArray("token_endpoint_auth_methods_supported")))
                    .addAllClaimsSupported(jsonArrayToList(object.getJSONArray("claims_supported")))
                    .addAllIntrospectionEndpointAuthSigningAlgValuesSupported(jsonArrayToList(object.getJSONArray("introspection_endpoint_auth_signing_alg_values_supported")))
                    .setRequestParameterSupported(object.optBoolean("request_parameter_supported"))
                    .setPushedAuthorizationRequestEndpoint(object.optString("pushed_authorization_request_endpoint"))
                    .setIntrospectionEndpoint(apiDomain + "/api/v1/identity-management/token/introspect")
                    .setClaimsParameterSupported(object.optBoolean("claims_parameter_supported"))
                    .addAllIdTokenEncryptionEncValuesSupported(jsonArrayToList(object.getJSONArray("id_token_encryption_enc_values_supported")))
                    .addAllUserinfoEncryptionEncValuesSupported(jsonArrayToList(object.getJSONArray("userinfo_encryption_enc_values_supported")))
                    .addAllIntrospectionEndpointAuthMethodsSupported(jsonArrayToList(object.getJSONArray("introspection_endpoint_auth_methods_supported")))
                    .addAllAuthorizationEncryptionAlgValuesSupported(jsonArrayToList(object.getJSONArray("authorization_encryption_alg_values_supported")))
                    .setTlsClientCertificateBoundAccessTokens(object.optBoolean("tls_client_certificate_bound_access_tokens"))
                    .addAllResponseModesSupported(jsonArrayToList(object.getJSONArray("response_modes_supported")))
                    .setBackchannelLogoutSessionSupported(object.optBoolean("backchannel_logout_session_supported"))
                    .addAllBackchannelAuthenticationRequestSigningAlgValuesSupported(jsonArrayToList(object.getJSONArray("backchannel_authentication_request_signing_alg_values_supported")))
                    .addAllAuthorizationEncryptionEncValuesSupported(jsonArrayToList(object.getJSONArray("authorization_encryption_enc_values_supported")))
                    .addAllRevocationEndpointAuthSigningAlgValuesSupported(jsonArrayToList(object.getJSONArray("revocation_endpoint_auth_signing_alg_values_supported")))
                    .addAllBackchannelTokenDeliveryModesSupported(jsonArrayToList(object.getJSONArray("backchannel_token_delivery_modes_supported")))
                    .addAllRevocationEndpointAuthMethodsSupported(jsonArrayToList(object.getJSONArray("revocation_endpoint_auth_methods_supported")))
                    .setRequestUriParameterSupported(object.optBoolean("request_uri_parameter_supported"))
                    .addAllGrantTypesSupported(jsonArrayToList(object.getJSONArray("grant_types_supported")))
                    .setRequireRequestUriRegistration(object.optBoolean("require_request_uri_registration"))
                    .addAllCodeChallengeMethodsSupported(jsonArrayToList(object.getJSONArray("code_challenge_methods_supported")))
                    .addAllIdTokenEncryptionAlgValuesSupported(jsonArrayToList(object.getJSONArray("id_token_encryption_alg_values_supported")))
                    .setFrontchannelLogoutSessionSupported(object.optBoolean("frontchannel_logout_session_supported"))
                    .addAllAuthorizationSigningAlgValuesSupported(jsonArrayToList(object.getJSONArray("authorization_signing_alg_values_supported")))
                    .addAllRequestObjectSigningAlgValuesSupported(jsonArrayToList(object.getJSONArray("request_object_signing_alg_values_supported")))
                    .addAllRequestObjectEncryptionAlgValuesSupported(jsonArrayToList(object.getJSONArray("request_object_encryption_alg_values_supported")))
                    .setCheckSessionIframe(object.optString("check_session_iframe"))
                    .setBackchannelLogoutSupported(object.optBoolean("backchannel_logout_supported"))
                    .addAllAcrValuesSupported(jsonArrayToList(object.getJSONArray("acr_values_supported")))
                    .addAllRequestObjectEncryptionEncValuesSupported(jsonArrayToList(object.getJSONArray("request_object_encryption_enc_values_supported")))
                    .setDeviceAuthorizationEndpoint(object.optString("device_authorization_endpoint"))
                    .addAllUserinfoSigningAlgValuesSupported(jsonArrayToList(object.getJSONArray("userinfo_signing_alg_values_supported")))
                    .setRequirePushedAuthorizationRequests(object.optBoolean("require_pushed_authorization_requests"))
                    .addAllClaimTypesSupported(jsonArrayToList(object.getJSONArray("claim_types_supported")))
                    .addAllUserinfoEncryptionAlgValuesSupported(jsonArrayToList(object.getJSONArray("userinfo_encryption_alg_values_supported")))
                    .setEndSessionEndpoint(object.optString("end_session_endpoint"))
                    .setRevocationEndpoint(object.optString("revocation_endpoint"))
                    .setBackchannelAuthenticationEndpoint(object.optString("backchannel_authentication_endpoint"))
                    .setFrontchannelLogoutSupported(object.optBoolean("frontchannel_logout_supported"))
                    .addAllTokenEndpointAuthSigningAlgValuesSupported(jsonArrayToList(object.getJSONArray("token_endpoint_auth_signing_alg_values_supported")))
                    .setRegistrationEndpoint(object.optString("registration_endpoint"))
                    .build();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching access token for  user " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }

    public TokenResponse getTokenByPasswordGrantType(GetTokenRequest request) {
        try {
            LOGGER.debug("Request token for " + request.getUsername() + " at " + request.getTenantId());

            JSONObject object = keycloakAuthClient.getAccessTokenFromPasswordGrantType(request.getClientId(),
                    request.getClientSecret(),
                    String.valueOf(request.getTenantId()),
                    request.getUsername(),
                    request.getPassword());

            try {
                return generateTokenResponse(object);

            } catch (Exception ex) {
                String msg = object.getString("error") + " " + object.getString("error_description");
                LOGGER.error(msg, ex);
                throw new RuntimeException(msg, ex);
            }
        } catch (Exception ex) {
            String msg = "Error occurred while fetching access token for  user " + request.getUsername() + " " + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }

    public TokenResponse getTokenByRefreshTokenGrantType(GetTokenRequest request) {
        try {
            LOGGER.debug("Request token for " + request.getUsername() + " at " + request.getTenantId());

            JSONObject object = keycloakAuthClient.getAccessTokenFromRefreshTokenGrantType(request.getClientId(),
                    request.getClientSecret(),
                    String.valueOf(request.getTenantId()),
                    request.getRefreshToken());

            try {
                return generateTokenResponse(object);

            } catch (Exception ex) {
                String msg = object.getString("error") + " " + object.getString("error_description");
                LOGGER.error(msg);
                throw new RuntimeException(msg, ex);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while fetching access token for  user " + request.getUsername() + " " + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }

    public JSONObject getJWKS(GetJWKSRequest request) {
        try {
            LOGGER.debug("Request JWT certificates for " + request.getTenantId());

            JSONObject object = keycloakAuthClient.getJWTVerificationCerts(request.getClientId(), request.getClientSecret(), String.valueOf(request.getTenantId()));

            try {
                if (object != null && object.getJSONArray("keys") != null) {
                    return object;
                }

                LOGGER.error("Missing 'keys' field in jwks_uri");
                throw new Exception("Missing 'keys' field in jwks_uri");

            } catch (Exception ex) {
                String msg = object != null
                        ? object.getString("error") + " " + object.getString("error_description")
                        : "JWKS format error";
                LOGGER.error(msg, ex);
                throw new RuntimeException(msg, ex);
            }

        } catch (Exception ex) {
            String msg = "Error occurred while pulling certs" + ex.getMessage();
            LOGGER.error(msg);
            throw new RuntimeException(msg, ex);
        }
    }

    public OperationStatus endSession(EndSessionRequest request) {
        try {
            LOGGER.debug("Request to end session for " + request.getTenantId());

            boolean status = keycloakAuthClient.revokeRefreshToken(request.getClientId(), request.getClientSecret(),
                    String.valueOf(request.getTenantId()), request.getRefreshToken());

            return OperationStatus.newBuilder().setStatus(status).build();

        } catch (Exception ex) {
            String msg = "Error occurred while revoking refresh token" + ex.getMessage();
            LOGGER.error(msg, ex);
            throw new RuntimeException(msg, ex);
        }
    }

    public JSONObject tokenIntrospection(String clientId, String clientSecret, String tenantId, String token) {
        JSONObject object = keycloakAuthClient.tokenIntrospection(clientId, clientSecret, tenantId, token);
        object.remove("realm_access");
        object.remove("resource_access");
        object.put("iss", "https://" + tenantId + ".usecustos.org");
        return object;
    }

    private TokenResponse generateTokenResponse(JSONObject object) throws Exception {
        if (object != null && object.has("access_token")) {
            return TokenResponse.newBuilder()
                    .setAccessToken(object.getString("access_token"))
                    .setExpiresIn(object.optInt("expires_in"))
                    .setRefreshToken(object.optString("refresh_token"))
                    .setRefreshExpiresIn(object.optInt("refresh_expires_in"))
                    .setTokenType(object.optString("token_type"))
                    .setScope(object.optString("scope"))
                    .setIdToken(object.optString("id_token"))
                    .setSessionState(object.optString("session_state"))
                    .setNotBeforePolicy(object.optInt("not-before-policy"))
                    .build();
        }
        LOGGER.error("Keycloak authentication client does not have the 'access_token");
        throw new Exception("Keycloak authentication client does not have the 'access_token");
    }

    private List<String> jsonArrayToList(JSONArray jsonArray) {
        return jsonArray == null ? Collections.emptyList() :
                IntStream.range(0, jsonArray.length())
                        .mapToObj(jsonArray::getString)
                        .toList();
    }
}
