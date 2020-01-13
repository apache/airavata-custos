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

package org.apache.custos.identity.service;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import org.apache.custos.federated.services.clients.keycloak.auth.KeycloakAuthClient;
import org.apache.custos.identity.authzcache.AuthzCacheEntry;
import org.apache.custos.identity.authzcache.AuthzCacheIndex;
import org.apache.custos.identity.authzcache.AuthzCachedStatus;
import org.apache.custos.identity.authzcache.DefaultAuthzCacheManager;
import org.apache.custos.identity.exceptions.CustosSecurityException;
import org.apache.custos.identity.service.IdentityServiceGrpc.IdentityServiceImplBase;
import org.apache.http.HttpStatus;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;


/**
 * A service responsible for authenticate and authorize users
 */
@GRpcService
public class IdentityService extends IdentityServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(IdentityService.class);

    private static final int CACHE_LIFE_TIME = 1000 * 60 * 60;


    @Autowired
    private KeycloakAuthClient keycloakAuthClient;

    @Autowired
    private DefaultAuthzCacheManager authzCacheManager;

    @Value("${custos.identity.auth.cache.enabled:true}")
    private boolean isAuthzCacheEnabled;


    @Override
    public void authenticate(AuthenticationRequest request,
                             StreamObserver<AuthToken> responseObserver) {
        try {
            LOGGER.debug("Authentication request received for " + request.getUsername());
            String accessToken = keycloakAuthClient.authenticate(
                    request.getClientId(), request.getClientSecret(),
                    String.valueOf(request.getTenantId()), request.getUsername(), request.getPassword());

            if (accessToken != null) {
                AuthToken.Builder authzBuilder = AuthToken.newBuilder()
                        .setAccessToken(accessToken);

                Claim userClaim = Claim.newBuilder().setKey("username").setValue(request.getUsername()).build();

                Claim tenantClaim = Claim.newBuilder().setKey("tenantId").setValue(String.valueOf(request.getTenantId())).build();

                authzBuilder.addClaims(userClaim);
                authzBuilder.addClaims(tenantClaim);

                AuthToken token = authzBuilder.build();

                responseObserver.onNext(token);
                responseObserver.onCompleted();

            } else {
                responseObserver.onError(Status.UNAUTHENTICATED.asRuntimeException());
            }

        } catch (org.keycloak.authorization.client.util.HttpResponseException ex) {
            String msg = "Error occurred while authenticating  user " + request.getUsername() + " " + ex.getReasonPhrase();
            LOGGER.error(msg);
            if (ex.getStatusCode() == HttpStatus.SC_UNAUTHORIZED) {
                responseObserver.onError(Status.UNAUTHENTICATED.withDescription(ex.getReasonPhrase()).asRuntimeException());
            }
        } catch (Exception ex) {
            String msg = "Error occurred while authenticating  user " + request.getUsername() + " " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }


    @Override
    public void getUser(AuthToken request,
                        StreamObserver<org.apache.custos.identity.service.User> responseObserver) {
        String username = null;
        String tenantId = null;
        try {
            for (Claim claim : request.getClaimsList()) {
                if (claim.getKey().equals("username")) {
                    username = claim.getValue();
                } else if (claim.getKey().equals("tenantId")) {
                    tenantId = claim.getValue();
                }
            }

            LOGGER.debug("Get user request received for " + username);

            org.apache.custos.federated.services.clients.keycloak.auth.User user = keycloakAuthClient.
                    getUser(request.getAccessToken(), tenantId);
            org.apache.custos.identity.service.User user1 =
                    org.apache.custos.identity.service.User.newBuilder()
                            .setEmailAddress(user.getEmailAddress())
                            .setFirstName(user.getFirstName())
                            .setLastName(user.getLastName())
                            .setFullName(user.getFullName())
                            .setSub(user.getSub())
                            .setUsername(user.getUsername())
                            .build();
            responseObserver.onNext(user1);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching  user " + username + " " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isAuthenticate(AuthToken request,
                               StreamObserver<IsAuthenticateResponse> responseObserver) {
        String username = null;
        String tenantId = null;

        for (Claim claim : request.getClaimsList()) {
            if (claim.getKey().equals("username")) {
                username = claim.getValue();
            } else if (claim.getKey().equals("tenantId")) {
                tenantId = claim.getValue();
            }
        }
        LOGGER.debug("Authentication status checking for  " + username
        );
        String accessToken = request.getAccessToken();


        boolean isAuthenticated = false;

        try {
            if (isAuthzCacheEnabled) {

                //check in the cache
                AuthzCachedStatus authzCachedStatus = authzCacheManager.getAuthzCachedStatus(
                        new AuthzCacheIndex(username, tenantId, accessToken));

                if (AuthzCachedStatus.AUTHORIZED.equals(authzCachedStatus)) {
                    LOGGER.debug("Authz decision for: (" + username + ", " + accessToken + ") is retrieved from cache.");
                    isAuthenticated = true;
                } else if (AuthzCachedStatus.NOT_AUTHORIZED.equals(authzCachedStatus)) {
                    LOGGER.debug("Authz decision for: (" + username + ", " + accessToken + ") is retrieved from cache.");
                    isAuthenticated = false;
                } else if (AuthzCachedStatus.NOT_CACHED.equals(authzCachedStatus)) {
                    LOGGER.debug("Authz decision for: (" + username + ", " + accessToken + ") is not in the cache. " +
                            "Generating decision based on group membership.");
                    isAuthenticated = keycloakAuthClient.isUserAuthenticated(username, tenantId, accessToken);
                    //cache the authorization decision
                    long currentTime = System.currentTimeMillis();
                    authzCacheManager.addToAuthzCache(new AuthzCacheIndex(username, tenantId, accessToken),
                            new AuthzCacheEntry(isAuthenticated, currentTime + CACHE_LIFE_TIME, currentTime));

                } else {
                    //undefined status returned from the authz cache manager
                    throw new CustosSecurityException("Error in reading from the authorization cache.");
                }
            } else {
                isAuthenticated = keycloakAuthClient.isUserAuthenticated(username, tenantId, accessToken);
            }

            if (isAuthenticated) {
                LOGGER.debug("User" + username + "in gateway" + tenantId + "is authenticated");

            } else {
                LOGGER.debug("User" + username + "in gateway" + tenantId + "is not authenticated");
            }
            IsAuthenticateResponse isAuthenticateResponse = IsAuthenticateResponse
                    .newBuilder()
                    .setAuthenticated(isAuthenticated)
                    .build();
            responseObserver.onNext(isAuthenticateResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while checking authentication for  user " + username + " " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.UNAUTHENTICATED.withDescription(msg).asRuntimeException());
        }
    }


    @Override
    public void getUserManagementServiceAccountAccessToken(GetUserManagementSATokenRequest request,
                                                           StreamObserver<AuthToken> responseObserver) {
        try {
            LOGGER.debug("Retreiving service account access token for " + request.getClientId());
            String accessToken = keycloakAuthClient.getUserManagementServiceAccountAccessToken(request.getClientId(),
                    request.getClientSecret(), String.valueOf(request.getTenantId()));

            AuthToken.Builder builder = AuthToken.newBuilder()
                    .setAccessToken(accessToken);


            Claim userClaim = Claim.newBuilder().setKey("username").setValue(request.getClientId()).build();
            Claim tenantId = Claim.newBuilder().setKey("tenantId").setValue(String.valueOf(request.getTenantId())).build();

            builder.addClaims(userClaim);
            builder.addClaims(tenantId);

            AuthToken token = builder.build();


            responseObserver.onNext(token);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching access token for  user " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }
}
