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

package org.apache.custos.tenant.registration.tasks;

import org.apache.custos.iam.service.User;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceTaskImpl;
import org.apache.custos.integration.core.endpoint.TargetEndpoint;
import org.apache.custos.tenant.profile.service.Gateway;
import org.custos.tenant.profile.client.async.TenantProfileClient;

public class AddTenantTask<T, U> extends ServiceTaskImpl<T, U> {

    private TargetEndpoint endpoint;


    public AddTenantTask(TargetEndpoint targetEndpoint) {
        this.endpoint = targetEndpoint;
    }

    @Override
    public void invokeService(T data) {
        Gateway gateway = (Gateway) data;
        TenantProfileClient tenantProfileClient = new TenantProfileClient(this.endpoint);

        ServiceCallback myCallback = (msg, exception) -> {
            if (exception != null) {
                System.out.println(exception);
                getServiceCallback().onCompleted(msg, exception);
            } else {
                System.out.println("Calling next task");
                U user = (U) User.newBuilder().setUserId("XXXXX").build();
                invokeNextTask(user);
            }
        };

        tenantProfileClient.addGateway(gateway, myCallback);
    }
}
