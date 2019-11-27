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

package org.apache.custos.tenant.registration.controller;

import org.apache.custos.iam.service.User;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceChain;
import org.apache.custos.tenant.profile.service.Gateway;
import org.apache.custos.tenant.registration.model.Tenant;
import org.apache.custos.tenant.registration.tasks.AddIamAdminUserTask;
import org.apache.custos.tenant.registration.tasks.AddTenantTask;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * A service use to register a tenant
 */
@RestController
@RequestMapping(value = "/tenant")
public class TenantRegistrationController {

    private static final Logger logger = LoggerFactory.getLogger(TenantRegistrationController.class);

    @Autowired
    private AddTenantTask<Gateway, User> addTenantTask;

    @Autowired
    private AddIamAdminUserTask<User, User> addIamAdminUserTask;

    @PostMapping
    public void createTenant(@RequestBody Tenant tenant) {
        try {

            logger.info("Receiving tenant info for tenant " + tenant.getName());


            ServiceCallback callback = (msg, exception) -> System.out.println("Completing create tenant");


            ServiceChain chain = ServiceChain.newBuilder(addTenantTask, callback).
                    nextTask(addIamAdminUserTask).build();

            Gateway gateway = Gateway.newBuilder().setGatewayId("qweqw").setInternalGatewayId("asdasd").build();


            chain.serve(gateway);

        } catch (Exception e) {
            e.printStackTrace();
        }

    }


}