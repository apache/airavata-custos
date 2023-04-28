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

import org.apache.custos.integration.core.utils.Constants;
import org.apache.custos.tenant.management.service.CreateTenantResponse;
import org.apache.custos.tenant.manamgement.client.SuperAdminOperationsClient;
import org.apache.custos.tenant.manamgement.client.TenantManagementClient;
import org.apache.custos.tenant.profile.service.TenantStatus;
import org.springframework.boot.autoconfigure.SpringBootApplication;

import java.io.IOException;

@SpringBootApplication
public class CustosBootstrapService {
    public static void main(String[] args) throws IOException {

        TenantManagementClient tenantManagementClient = new TenantManagementClient("localhost", 7000);
        CreateTenantResponse response = tenantManagementClient.registerTenant("Custos Super Tenant",
                "xxxx@gmail.com",
                "Custos",
                "Admin",
                "xxxx@gmail.com",
                "custosadmin",
                "1234",
                new String[]{"xxxx@gmail.com"},
                new String[]{"http://localhost:8080/callback"},
                "http://localhost:8080/",
                "openid email profile cilogon",
                "localhost",
                "http://localhost:8080/",
                "This is custos bootstrapping client");

        SuperAdminOperationsClient adminOperationsClient = new SuperAdminOperationsClient("localhost", 7000);
        adminOperationsClient.updateTenantStatus(response.getClientId(), TenantStatus.ACTIVE, true,
                Constants.SYSTEM);
        System.out.println("Super Tenant Activate Successfully");
        System.out.println("Client Id :" + response.getClientId() + " Client Secret :" + response.getClientSecret());

    }
}
