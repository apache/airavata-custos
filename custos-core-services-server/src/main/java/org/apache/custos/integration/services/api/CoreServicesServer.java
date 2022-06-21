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
import org.apache.custos.agent.profile.validator.AgentInputValidator;
import org.apache.custos.cluster.management.validator.ClusterManagementInputValidator;
import org.apache.custos.core.services.commons.ServiceInterceptor;
import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.credential.store.validator.CredentialStoreInputValidator;
import org.apache.custos.federated.authentication.validator.FederatedAuthenticationInputValidator;
import org.apache.custos.iam.validator.IAMInputValidator;
import org.apache.custos.identity.validator.IdentityInputValidator;
import org.apache.custos.messaging.events.publisher.MessageProducer;
import org.apache.custos.resource.secret.validator.ResourceSecretInputValidator;
import org.apache.custos.sharing.validator.SharingInputValidator;
import org.apache.custos.tenant.profile.validator.TenantProfileInputValidator;
import org.apache.custos.user.profile.validators.UserProfileInputValidator;
import org.lognet.springboot.grpc.GRpcGlobalInterceptor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.domain.EntityScan;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.data.jpa.repository.config.EnableJpaAuditing;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;
import org.apache.custos.logging.validator.CustosLoggingInputValidator;

import java.util.Stack;


@SpringBootApplication
@EnableJpaAuditing
@EnableJpaRepositories(basePackages = "org.apache.custos")
@ComponentScan(basePackages = "org.apache.custos")
@EntityScan(basePackages = "org.apache.custos")
public class CoreServicesServer {
    public static void main(String[] args) {
        SpringApplication.run(CoreServicesServer.class, args);
    }



    @Bean
    public Stack<Validator> getInterceptorSet(AgentInputValidator inputValidator,
                                                       ClusterManagementInputValidator clusterManagementInputValidator,
                                                       CredentialStoreInputValidator credentialStoreInputValidator,
                                                       CustosLoggingInputValidator custosLoggingInputValidator,
                                                       FederatedAuthenticationInputValidator federatedAuthenticationInputValidator,
                                                       IAMInputValidator iamInputValidator,
                                                       IdentityInputValidator identityInputValidator,
                                                       ResourceSecretInputValidator resourceSecretInputValidator,
                                                       SharingInputValidator sharingInputValidator,
                                                       TenantProfileInputValidator tenantProfileInputValidator,
                                                       UserProfileInputValidator userProfileInputValidator) {
        Stack<Validator> interceptors = new Stack<>();
        interceptors.add(inputValidator);
        interceptors.add(clusterManagementInputValidator);
        interceptors.add(credentialStoreInputValidator);
        interceptors.add(custosLoggingInputValidator);
        interceptors.add(federatedAuthenticationInputValidator);
        interceptors.add(iamInputValidator);
        interceptors.add(identityInputValidator);
        interceptors.add(resourceSecretInputValidator);
        interceptors.add(sharingInputValidator);
        interceptors.add(tenantProfileInputValidator);
        interceptors.add(userProfileInputValidator);

        return interceptors;
    }




    @Bean
    @GRpcGlobalInterceptor
    ServerInterceptor validationInterceptor(Stack<Validator> validators){
        return new ServiceInterceptor(validators);
    }

    @Bean
    public MessageProducer registerMessageProducer(@Value("${core.messaging.service.broker.url}") String borkerURL,
                                                   @Value("${core.messaging.service.publisher.id}") String publisherId) {
        return new MessageProducer(borkerURL, publisherId);
    }
}
