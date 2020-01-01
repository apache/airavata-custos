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

//package org.apache.custos.tenant.management.tasks;
//
//import org.apache.custos.iam.admin.client.IamAdminServiceClient;
//import org.apache.custos.iam.service.SetUpTenantRequest;
//import org.apache.custos.iam.service.SetUpTenantResponse;
//import org.apache.custos.integration.core.ServiceCallback;
//import org.apache.custos.integration.core.ServiceTaskImpl;
//import org.slf4j.Logger;
//import org.slf4j.LoggerFactory;
//import org.springframework.beans.factory.annotation.Autowired;
//import org.springframework.stereotype.Component;
//
//@Component
//public class AddIamAdminUserTask<T, U> extends ServiceTaskImpl<T, U> {
//
//    private static final Logger LOGGER = LoggerFactory.getLogger(AddIamAdminUserTask.class);
//
//    @Autowired
//    private IamAdminServiceClient iamAdminServiceClient;
//
//    @Override
//    public void invokeService(T data) {
//        LOGGER.info("Invoking Add Iam Admin User Task");
//        SetUpTenantRequest request = (SetUpTenantRequest) data;
//
//        ServiceCallback myCallback = (msg, exception) -> {
//            if (exception != null) {
//                LOGGER.info("IAM admin setup tenant  task completed and return to parent");
//                getServiceCallback().onCompleted(msg, exception);
//            } else {
//                SetUpTenantResponse response = (SetUpTenantResponse)msg;
//
//
//                invokeNextTask();
//            }
//        };
//
//        iamAdminServiceClient.setUPTenantAsync(request, myCallback);
//    }
//}