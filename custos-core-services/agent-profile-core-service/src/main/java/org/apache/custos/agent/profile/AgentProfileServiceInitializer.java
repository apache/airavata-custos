///*
// * Licensed to the Apache Software Foundation (ASF) under one
// * or more contributor license agreements. See the NOTICE file
// * distributed with this work for additional information
// * regarding copyright ownership. The ASF licenses this file
// * to you under the Apache License, Version 2.0 (the
// * "License"); you may not use this file except in compliance
// * with the License. You may obtain a copy of the License at
// *
// * http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing,
// * software distributed under the License is distributed on an
// * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// * KIND, either express or implied. See the License for the
// *  specific language governing permissions and limitations
// *  under the License.
// */
//
//package org.apache.custos.agent.profile;
//
//import io.grpc.ServerInterceptor;
//import org.apache.custos.agent.profile.validator.AgentInputValidator;
//import org.apache.custos.core.services.commons.ServiceInterceptor;
//import org.lognet.springboot.grpc.GRpcGlobalInterceptor;
//import org.springframework.boot.SpringApplication;
//import org.springframework.boot.autoconfigure.SpringBootApplication;
//import org.springframework.context.annotation.Bean;
//import org.springframework.data.jpa.repository.config.EnableJpaAuditing;
//
//@SpringBootApplication
//@EnableJpaAuditing
//public class AgentProfileServiceInitializer {
//
//    public static void main(String[] args) {
//        SpringApplication.run(AgentProfileServiceInitializer.class, args);
//    }
//
//    @Bean
//    public AgentInputValidator getValidator() {
//        return new AgentInputValidator();
//    }
//
//    @Bean
//    @GRpcGlobalInterceptor
//    ServerInterceptor validationInterceptor(AgentInputValidator validator){
//        return new ServiceInterceptor(validator);
//    }
//}
