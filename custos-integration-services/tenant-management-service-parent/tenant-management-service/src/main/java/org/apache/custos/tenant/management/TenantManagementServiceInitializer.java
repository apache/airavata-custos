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

package org.apache.custos.tenant.management;

import brave.Tracing;
import brave.grpc.GrpcTracing;
import io.grpc.ClientInterceptor;
import io.grpc.ServerInterceptor;
import org.apache.custos.integration.core.interceptor.IntegrationServiceInterceptor;
import org.apache.custos.integration.core.interceptor.ServiceInterceptor;
import org.apache.custos.tenant.management.interceptors.AuthInterceptorImpl;
import org.apache.custos.tenant.management.interceptors.DynamicRegistrationValidator;
import org.apache.custos.tenant.management.interceptors.InputValidator;
import org.lognet.springboot.grpc.GRpcGlobalInterceptor;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;


@SpringBootApplication
@ComponentScan(basePackages = "org.apache.custos")
public class TenantManagementServiceInitializer {
    public static void main(String[] args) {
        SpringApplication.run(TenantManagementServiceInitializer.class, args);
    }

    @Bean
    public GrpcTracing grpcTracing(Tracing tracing) {
        //   Tracing tracing1 =  Tracing.newBuilder().build();
        return GrpcTracing.create(tracing);
    }

    //We also create a client-side interceptor and put that in the context, this interceptor can then be injected into gRPC clients and
    //then applied to the managed channel.
    @Bean
    ClientInterceptor grpcClientSleuthInterceptor(GrpcTracing grpcTracing) {
        return grpcTracing.newClientInterceptor();
    }

    @Bean
    @GRpcGlobalInterceptor
    ServerInterceptor grpcServerSleuthInterceptor(GrpcTracing grpcTracing) {
        return grpcTracing.newServerInterceptor();
    }

    @Bean
    public List<IntegrationServiceInterceptor> getInterceptorSet(AuthInterceptorImpl authInterceptor,
                                                                DynamicRegistrationValidator registrationValidator, InputValidator validator) {
        List<IntegrationServiceInterceptor> interceptors = new ArrayList<>();
        interceptors.add(validator);
        interceptors.add(authInterceptor);
        interceptors.add(registrationValidator);

        return interceptors;
    }

    @Bean
    @GRpcGlobalInterceptor
    ServerInterceptor validationInterceptor(List<IntegrationServiceInterceptor> integrationServiceInterceptors) {
        return new ServiceInterceptor(integrationServiceInterceptors);
    }

}
