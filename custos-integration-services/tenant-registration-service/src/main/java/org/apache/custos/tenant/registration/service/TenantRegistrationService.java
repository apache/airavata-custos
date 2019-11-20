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

package org.apache.custos.tenant.registration.service;

import org.apache.custos.tenant.registration.model.Tenant;
import org.custos.tenant.profile.client.TenantProfileClient;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.UUID;

/**
 * A service use to register a tenant
 */
@RestController
@RequestMapping(value = "/tenant")
public class TenantRegistrationService {

    @Value("${tenant.profile.core.service.name}")
    private String tenantServiceName;

    @Value("${tenant.profile.core.service.port}")
    private int tenantServicePort;


    @PostMapping
    public void createTenant(@RequestBody Tenant tenant){

        TenantProfileClient client = new TenantProfileClient(tenantServiceName, tenantServicePort);
        client.addGateway(tenant.getName(),UUID.randomUUID().toString());

        System.out.println("Successfully added the gateway");

    }


}
