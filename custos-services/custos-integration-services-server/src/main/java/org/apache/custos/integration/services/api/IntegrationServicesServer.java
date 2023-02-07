/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.integration.services.api;

import io.grpc.ServerInterceptor;
import org.apache.custos.agent.management.interceptors.AgentManagementClientAuthInterceptorImpl;
import org.apache.custos.agent.management.interceptors.AgentManagementInputValidator;
import org.apache.custos.agent.management.interceptors.AgentManagementSuperTenantRestrictedOperationsInterceptorImpl;
import org.apache.custos.agent.management.interceptors.AgentManagementUserAuthInterceptorImpl;
import org.apache.custos.group.management.interceptors.GroupManagementClientAuthInterceptorImpl;
import org.apache.custos.group.management.interceptors.GroupManagementInputValidator;
import org.apache.custos.identity.management.interceptors.IdentityManagementAgentAuthInterceptor;
import org.apache.custos.identity.management.interceptors.IdentityManagementAuthInterceptorImpl;
import org.apache.custos.identity.management.interceptors.IdentityManagementInputValidator;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.apache.custos.integration.core.interceptor.ServiceInterceptor;
import org.apache.custos.integration.services.commons.interceptors.LoggingInterceptor;
import org.apache.custos.log.management.interceptors.LogManagementAuthInterceptorImpl;
import org.apache.custos.log.management.interceptors.LogManagementInputValidator;
import org.apache.custos.resource.secret.management.interceptors.ResourceSecretManagementAuthInterceptorImpl;
import org.apache.custos.resource.secret.management.interceptors.ResourceSecretManagementInputValidator;
import org.apache.custos.sharing.management.interceptors.SharingManagementAuthInterceptorImpl;
import org.apache.custos.sharing.management.interceptors.SharingManagementInputValidator;
import org.apache.custos.tenant.management.interceptors.TenantManagementAuthInterceptorImpl;
import org.apache.custos.tenant.management.interceptors.TenantManagementDynamicRegistrationValidator;
import org.apache.custos.tenant.management.interceptors.TenantManagementInputValidator;
import org.apache.custos.tenant.management.interceptors.TenantManagementSuperTenantRestrictedOperationsInterceptorImpl;
import org.apache.custos.user.management.interceptors.UserManagementAuthInterceptorImpl;
import org.apache.custos.user.management.interceptors.UserManagementInputValidator;
import org.apache.custos.user.management.interceptors.UserManagementSuperTenantRestrictedOperationsInterceptorImpl;
import org.lognet.springboot.grpc.GRpcGlobalInterceptor;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;

import java.util.Stack;


@SpringBootApplication
@ComponentScan(basePackages = "org.apache.custos")
public class IntegrationServicesServer {
    public static void main(String[] args) {
        SpringApplication.run(IntegrationServicesServer.class, args);
    }



    @Bean
    public Stack<IntegrationServiceInterceptor> getInterceptorSet(AgentManagementClientAuthInterceptorImpl agentManagementClientAuthInterceptor,
                                                                  AgentManagementInputValidator agentManagementInputValidator,
                                                                  AgentManagementSuperTenantRestrictedOperationsInterceptorImpl agentManagementSuperTenantRestrictedOperationsInterceptor,
                                                                  AgentManagementUserAuthInterceptorImpl agentManagementUserAuthInterceptor,
                                                                  GroupManagementClientAuthInterceptorImpl groupManagementClientAuthInterceptor,
                                                                  GroupManagementInputValidator groupManagementInputValidator,
                                                                  IdentityManagementAgentAuthInterceptor identityManagementAgentAuthInterceptor,
                                                                  IdentityManagementAuthInterceptorImpl identityManagementAuthInterceptor,
                                                                  IdentityManagementInputValidator identityManagementInputValidator,
                                                                  LogManagementInputValidator logManagementInputValidator,
                                                                  LogManagementAuthInterceptorImpl logManagementAuthInterceptor,
                                                                  ResourceSecretManagementAuthInterceptorImpl resourceSecretManagementAuthInterceptor,
                                                                  ResourceSecretManagementInputValidator resourceSecretManagementInputValidator,
                                                                  SharingManagementAuthInterceptorImpl sharingManagementAuthInterceptor,
                                                                  SharingManagementInputValidator sharingManagementInputValidator,
                                                                  TenantManagementAuthInterceptorImpl tenantManagementAuthInterceptor,
                                                                  TenantManagementDynamicRegistrationValidator tenantManagementDynamicRegistrationValidator,
                                                                  TenantManagementInputValidator tenantManagementInputValidator,
                                                                  TenantManagementSuperTenantRestrictedOperationsInterceptorImpl tenantManagementSuperTenantRestrictedOperationsInterceptor,
                                                                  UserManagementAuthInterceptorImpl userManagementAuthInterceptor,
                                                                  UserManagementInputValidator userManagementInputValidator,
                                                                  UserManagementSuperTenantRestrictedOperationsInterceptorImpl userManagementSuperTenantRestrictedOperationsInterceptor,
                                                                  LoggingInterceptor loggingInterceptor) {
        Stack<IntegrationServiceInterceptor> interceptors = new Stack<>();
        interceptors.add(agentManagementClientAuthInterceptor);
        interceptors.add(agentManagementInputValidator);
        interceptors.add(agentManagementSuperTenantRestrictedOperationsInterceptor);
        interceptors.add(agentManagementUserAuthInterceptor);
        interceptors.add(groupManagementClientAuthInterceptor);
        interceptors.add(groupManagementInputValidator);
        interceptors.add(identityManagementAgentAuthInterceptor);
        interceptors.add(identityManagementAuthInterceptor);
        interceptors.add(logManagementInputValidator);
        interceptors.add(logManagementAuthInterceptor);
        interceptors.add(identityManagementInputValidator);
        interceptors.add(resourceSecretManagementAuthInterceptor);
        interceptors.add(resourceSecretManagementInputValidator);
        interceptors.add(sharingManagementAuthInterceptor);
        interceptors.add(sharingManagementInputValidator);
        interceptors.add(tenantManagementAuthInterceptor);
        interceptors.add(tenantManagementDynamicRegistrationValidator);
        interceptors.add(tenantManagementInputValidator);
        interceptors.add(tenantManagementSuperTenantRestrictedOperationsInterceptor);
        interceptors.add(userManagementAuthInterceptor);
        interceptors.add(userManagementInputValidator);
        interceptors.add(userManagementSuperTenantRestrictedOperationsInterceptor);
        interceptors.add(loggingInterceptor);


        return interceptors;
    }

    @Bean
    @GRpcGlobalInterceptor
    ServerInterceptor validationInterceptor(Stack<IntegrationServiceInterceptor> integrationServiceInterceptors) {
        return new ServiceInterceptor(integrationServiceInterceptors);
    }


}
