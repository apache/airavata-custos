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

package org.apache.custos.iam.admin.client.async;

import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.service.IamAdminServiceGrpc;
import org.apache.custos.iam.service.User;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * The client class used to connect IAM admin services
 */
@Component
public class IamAdminServiceClient {
    private ManagedChannel managedChannel;
    private IamAdminServiceGrpc.IamAdminServiceStub iamAdminServiceStub;


    private final List<ClientInterceptor> clientInterceptorList;


    public IamAdminServiceClient(List<ClientInterceptor> clientInterceptorList,
                                 @Value("${iam.admin.service.dns.name}") String serviceHost,
                                 @Value("${iam.admin.service.port}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext(true).intercept(clientInterceptorList).build();
        iamAdminServiceStub = IamAdminServiceGrpc.newStub(managedChannel);
    }


    public void addUser(User user, ServiceCallback callback) {
        StreamObserver observer = new StreamObserver() {
            @Override
            public void onNext(Object o) {
                //TODO implement this
            }

            @Override
            public void onError(Throwable throwable) {
                callback.onCompleted(null, new ServiceException("Add admin user task failed", throwable, null));
            }

            @Override
            public void onCompleted() {
                callback.onCompleted("Completed", null);
            }
        };


        iamAdminServiceStub.addUser(user, observer);


    }


}
