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
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.integration.core.endpoint.TargetEndpoint;
import org.apache.custos.tenant.profile.service.Gateway;
import org.apache.custos.tenant.registration.model.Tenant;
import org.apache.custos.tenant.registration.tasks.AddIamAdminUserTask;
import org.apache.custos.tenant.registration.tasks.AddTenantTask;
import org.springframework.beans.factory.annotation.Value;
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

  //  @Value("${tenant.profile.core.service.dns.name}")
    private String tenantServiceName;

  //  @Value("${tenant.profile.core.service.port}")
    private int tenantServicePort;

    @PostMapping
    public void createTenant(@RequestBody Tenant tenant) {
        try {

            TargetEndpoint tenProfile = new TargetEndpoint("localhost",7000);
            TargetEndpoint iamCustos = new TargetEndpoint("localhost",7001);

            ServiceCallback callback = (msg, exception) -> System.out.println("Completing create tenant");
            AddTenantTask<Gateway, User> addTenantTask = new AddTenantTask<>(tenProfile);
            AddIamAdminUserTask<User,User>  addIamAdminUserTask = new AddIamAdminUserTask<>(iamCustos);

            ServiceChain chain = new ServiceChain.
                                     ServiceChainBuilder(addTenantTask,callback).nextTask(addIamAdminUserTask).build();
            Gateway gateway = Gateway.newBuilder().setGatewayId("qweqw").setInternalGatewayId("asdasd").build();
            chain.serve(gateway);

        } catch (Exception e) {
            e.printStackTrace();
        }

    }


}
