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

import com.nimbusds.jose.JOSEObjectType;
import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.JWSHeader;
import com.nimbusds.jose.JWSSigner;
import com.nimbusds.jose.crypto.RSASSASigner;
import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;
import org.apache.custos.core.model.user.Group;
import org.apache.custos.core.model.user.GroupRole;
import org.apache.custos.core.user.profile.api.UserProfile;
import org.apache.custos.core.user.profile.api.UserProfileRequest;
import org.apache.custos.service.profile.UserProfileService;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cache.Cache;
import org.springframework.cache.CacheManager;
import org.springframework.stereotype.Service;

import java.security.KeyPair;
import java.text.ParseException;
import java.util.Arrays;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

@Service
public class TokenService {

    private static final Logger LOGGER = LoggerFactory.getLogger(TokenService.class);

    private final KeyLoader keyLoader;
    private final UserProfileService userProfileService;
    private final CacheManager cacheManager;

    @Autowired
    public TokenService(KeyLoader keyLoader, UserProfileService userProfileService, CacheManager cacheManager) {
        this.keyLoader = keyLoader;
        this.userProfileService = userProfileService;
        this.cacheManager = cacheManager;
    }


    public String generateWithCustomClaims(String token, long tenantId) throws Exception {
        KeyPair keyPair = keyLoader.getKeyPair();
        String keyID = keyLoader.getKeyID();

        SignedJWT signedJWT = SignedJWT.parse(token);

        JWTClaimsSet oldClaims = signedJWT.getJWTClaimsSet();
        String email = String.valueOf(oldClaims.getClaim("email"));
        Set<String> existingScopes = new HashSet<>(Arrays.asList(((String) oldClaims.getClaim("scope")).split(" ")));

        JWTClaimsSet newClaims = null;
        try {
            if (StringUtils.isNotBlank(email)) {
                UserProfileRequest request = UserProfileRequest.newBuilder()
                        .setTenantId(tenantId)
                        .setProfile(UserProfile.newBuilder().setUsername(email).build())
                        .build();

                List<Group> groups = userProfileService.getGroupsOfUser(request);

                List<String> groupIds = groups.stream()
                        .map(Group::getExternalId)
                        .toList();

                List<String> scopes = groups.stream()
                        .flatMap(group -> group.getGroupRole().stream())
                        .map(GroupRole::getValue)
                        .distinct()
                        .toList();

                existingScopes.addAll(scopes);

                newClaims = new JWTClaimsSet.Builder(oldClaims)
                        .claim("groups", groupIds)
                        .claim("scope", String.join(" ", scopes))
                        .claim("scopes", scopes)
                        .claim("iss", "https://" + tenantId + ".usecustos.org")
                        .build();
            }

        } catch (Exception ex) {
            LOGGER.error("Error while adding custom claims to the token belongs to: {}", email);
        }

        JWSHeader header = new JWSHeader.Builder(JWSAlgorithm.RS256)
                .keyID(keyID)
                .type(JOSEObjectType.JWT)
                .build();

        SignedJWT newSignedJWT = new SignedJWT(header, (newClaims != null ? newClaims : oldClaims));

        JWSSigner signer = new RSASSASigner(keyPair.getPrivate());
        newSignedJWT.sign(signer);

        cacheToken(newClaims.getJWTID(), token);
        return newSignedJWT.serialize();
    }

    public String getKCToken(String customizedToken) throws ParseException {
        SignedJWT signedJWT = SignedJWT.parse(customizedToken);
        String jwtId = signedJWT.getJWTClaimsSet().getJWTID();
        Cache cache = cacheManager.getCache("KCTokenCache");
        if (cache != null) {
            Cache.ValueWrapper valueWrapper = cache.get(jwtId);
            if (valueWrapper != null) {
                return (String) valueWrapper.get();
            }
        }

        return null;
    }

    private void cacheToken(String jti, String token) {
        Cache cache = cacheManager.getCache("KCTokenCache");
        if (cache != null) {
            cache.put(jti, token);
        }
    }

}
