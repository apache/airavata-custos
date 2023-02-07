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

package org.apache.custos.scim.utils;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.json.JSONObject;
import org.springframework.stereotype.Component;
import org.springframework.web.client.HttpStatusCodeException;

import javax.swing.text.html.Option;
import java.util.Map;
import java.util.Optional;

@Component
public class AuthHandler extends AuthInterceptor {


    public AuthHandler(CredentialStoreServiceClient credentialStoreServiceClient, TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }


    public String getToken(String headerValue) {
        String prefix = "Bearer";
        String token = headerValue.substring(prefix.length());
        return token.trim();
    }

    public Optional<AuthClaim> validateAndConfigure(String header, boolean userTokenValidation) throws HttpStatusCodeException {
        String token = this.getToken(header);
        Optional<AuthClaim> claim = null;
        if (userTokenValidation) {
            claim = authorizeUsingUserToken(token);
        } else {
            claim = authorize(token);
        }

        if (claim.isEmpty()) {
            throw new NotAuthorizedException();
        }

       return claim;
    }

    @Override
    public <ReqT> ReqT intercept(String s, Metadata metadata, ReqT reqT) {
        return null;
    }
}
