///*
// * Licensed to the Apache Software Foundation (ASF) under one
// * or more contributor license agreements. See the NOTICE file
// * distributed with this work for additional information
// * regarding copyright ownership. The ASF licenses this file
// * to you under the Apache License, Version 2.0 (the
// * "License"); you may not use this file except in compliance
// * with the License. You may obtain a copy of the License at
// *
// *  http://www.apache.org/licenses/LICENSE-2.0
// *
// *  Unless required by applicable law or agreed to in writing,
// *  software distributed under the License is distributed on an
// *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// *  KIND, either express or implied. See the License for the
// *  specific language governing permissions and limitations
// *  under the License.
// */
//
//package org.apache.custos.messaging;
//
//import org.apache.custos.messaging.events.publisher.MessageProducer;
//import org.springframework.beans.factory.annotation.Value;
//import org.springframework.boot.SpringApplication;
//import org.springframework.boot.autoconfigure.SpringBootApplication;
//import org.springframework.boot.autoconfigure.domain.EntityScan;
//import org.springframework.context.annotation.Bean;
//import org.springframework.context.annotation.ComponentScan;
//import org.springframework.data.jpa.repository.config.EnableJpaAuditing;
//import org.springframework.data.jpa.repository.config.EnableJpaRepositories;
//
//@SpringBootApplication
//@EnableJpaAuditing
//@EnableJpaRepositories(basePackages = "org.apache.custos")
//@ComponentScan(basePackages = "org.apache.custos")
//@EntityScan(basePackages = "org.apache.custos")
//public class MessagingServiceInitializer {
//
//    public static void main(String[] args) {
//        SpringApplication.run(MessagingServiceInitializer.class, args);
//    }
//
//
//    @Bean
//    public MessageProducer registerMessageProducer(@Value("${core.messaging.service.broker.url}") String borkerURL,
//                                                   @Value("${core.messaging.service.publisher.id}") String publisherId) {
//        return new MessageProducer(borkerURL, publisherId);
//
//    }
//}
//
