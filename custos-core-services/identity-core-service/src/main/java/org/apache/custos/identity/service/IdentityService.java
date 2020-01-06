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
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;

import java.util.Map;


/**
 * A service responsible for authenticate and authorize users
 */
@GRpcService
public class IdentityService extends IdentityServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(IdentityService.class);

    private static final int cacheLifetime = 1000 * 60 * 60;


    @Autowired
    private KeycloakAuthClient keycloakAuthClient;

    @Autowired
    private DefaultAuthzCacheManager authzCacheManager;

    @Value("${custos.identity.auth.cache.enabled:true}")
    private boolean isAuthzCacheEnabled;


    @Override
    public void authenticate(AuthenticationRequest request,
                             StreamObserver<AuthzToken> responseObserver) {
        try {


        } catch (Exception ex) {
            String msg = "Error occurred while authenticating  user " + request.getUsername() + " " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }


    @Override
    public void getUser(AuthzToken request,
                        StreamObserver< org.apache.custos.identity.service.User> responseObserver) {
        try {

            String username = request.getClaimsMap().get("username");
            org.apache.custos.federated.services.clients.keycloak.auth.User user = keycloakAuthClient.
                                                     getUser(request.getAccessToken(),username);
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
            String msg = "Error occurred while fetching  user " + request.getClaimsMap().get("username") + " " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void isAuthenticate(AuthzToken request,
                               StreamObserver<IsAuthenticateResponse> responseObserver) {

        String accessToken = request.getAccessToken();

        Map<String, String> claimsMap = request.getClaimsMapMap();

        String userName = claimsMap.get("username");
        String tenantId = claimsMap.get("tenantId");

        boolean isAuthenticated = false;

        try {
            if (isAuthzCacheEnabled) {

                //check in the cache
                AuthzCachedStatus authzCachedStatus = authzCacheManager.getAuthzCachedStatus(
                        new AuthzCacheIndex(userName, tenantId, accessToken));

                if (AuthzCachedStatus.AUTHORIZED.equals(authzCachedStatus)) {
                    LOGGER.debug("Authz decision for: (" + userName + ", " + accessToken + ") is retrieved from cache.");
                    isAuthenticated = true;
                } else if (AuthzCachedStatus.NOT_AUTHORIZED.equals(authzCachedStatus)) {
                    LOGGER.debug("Authz decision for: (" + userName + ", " + accessToken + ") is retrieved from cache.");
                    isAuthenticated = false;
                } else if (AuthzCachedStatus.NOT_CACHED.equals(authzCachedStatus)) {
                    LOGGER.debug("Authz decision for: (" + userName + ", " + accessToken + ") is not in the cache. " +
                            "Generating decision based on group membership.");
                    isAuthenticated = keycloakAuthClient.isUserAuthenticated(userName, accessToken, tenantId);
                    //cache the authorization decision
                    long currentTime = System.currentTimeMillis();
                    authzCacheManager.addToAuthzCache(new AuthzCacheIndex(userName, tenantId, accessToken),
                            new AuthzCacheEntry(isAuthenticated, currentTime + cacheLifetime, currentTime));

                } else {
                    //undefined status returned from the authz cache manager
                    throw new CustosSecurityException("Error in reading from the authorization cache.");
                }
            }

            if (isAuthenticated) {
                LOGGER.debug("User" + userName + "in gateway" + tenantId + "is authenticated");

            } else {
                LOGGER.debug("User" + userName + "in gateway" + tenantId + "is not authenticated");
            }
            IsAuthenticateResponse isAuthenticateResponse = IsAuthenticateResponse
                    .newBuilder()
                    .setAuthenticated(isAuthenticated)
                    .build();
            responseObserver.onNext(isAuthenticateResponse);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while checking authentication for  user " + request.getClaimsMap().get("username") + " " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void getUserManagementServiceAccountAccessToken(GetUserManagementSATokenRequest request,
                                                           StreamObserver<AuthzToken> responseObserver) {
        try {

            String accessToken = keycloakAuthClient.getUserManagementServiceAccountAccessToken(request.getClientId(),
                    request.getClientSecret(), request.getTenantId());

            AuthzToken.Builder builder = AuthzToken.newBuilder()
                    .setAccessToken(accessToken);
            Map<String, String> map = builder.getMutableClaimsMap();
            map.put("username", request.getClientId());
            map.put("tenantId", request.getTenantId());
            AuthzToken token = builder.build();

            responseObserver.onNext(token);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = "Error occurred while fetching access token for  user " + request.getTenantId() + " " + ex.getMessage();
            LOGGER.error(msg);
            responseObserver.onError(Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }
}
