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

package org.apache.custos.identity.management.interceptors;

import io.grpc.Metadata;
import org.apache.custos.credential.store.client.CredentialStoreServiceClient;
import org.apache.custos.identity.client.IdentityClient;
import org.apache.custos.identity.management.service.EndSessionRequest;
import org.apache.custos.identity.management.service.GetAgentTokenRequest;
import org.apache.custos.integration.core.exceptions.UnAuthorizedException;
import org.apache.custos.integration.services.commons.interceptors.AuthInterceptor;
import org.apache.custos.integration.services.commons.model.AuthClaim;
import org.apache.custos.tenant.profile.client.async.TenantProfileClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * Responsible for authorize agent specific methods
 */
@Component
public class AgentAuthInterceptor extends AuthInterceptor {

    private static final Logger LOGGER = LoggerFactory.getLogger(AgentAuthInterceptor.class);

    public AgentAuthInterceptor(CredentialStoreServiceClient credentialStoreServiceClient,
                                TenantProfileClient tenantProfileClient, IdentityClient identityClient) {
        super(credentialStoreServiceClient, tenantProfileClient, identityClient);
    }

    @Override
    public <ReqT> ReqT intercept(String method, Metadata headers, ReqT msg) {
        if (method.equals("getAgentToken")) {
            AuthClaim claim = authorizeUsingAgentBasicToken(headers, ((GetAgentTokenRequest) msg).getClientId());
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }
            GetAgentTokenRequest reqCore =
                    ((GetAgentTokenRequest) msg).toBuilder()
                            .setTenantId(claim.getTenantId())
                            .setAgentClientId(claim.getAgentClientId())
                            .setAgentClientSecret(claim.getAgentClientSecret())
                            .setAgentId(claim.getAgentId())
                            .setAgentPassword(claim.getAgentPassword())
                            .build();

            return (ReqT) reqCore;
        } else if (method.equals("endAgentSession")) {
            LOGGER.info("ClientId " + ((EndSessionRequest) msg).getClientId());
            AuthClaim claim = authorizeUsingAgentBasicToken(headers, ((EndSessionRequest) msg).getClientId());
            if (claim == null) {
                throw new UnAuthorizedException("Request is not authorized", null);
            }


            org.apache.custos.identity.service.EndSessionRequest endSessionRequest =
                    ((EndSessionRequest) msg).getBody().toBuilder()
                            .setClientId(claim.getAgentClientId())
                            .setClientSecret(claim.getAgentClientSecret())
                            .setTenantId(claim.getTenantId())
                            .build();
            return (ReqT) ((EndSessionRequest) msg).toBuilder().setBody(endSessionRequest).build();

        }
        return msg;

    }
}
