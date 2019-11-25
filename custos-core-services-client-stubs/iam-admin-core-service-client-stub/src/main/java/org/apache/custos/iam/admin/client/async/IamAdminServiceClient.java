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

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.iam.service.IamAdminServiceGrpc;
import org.apache.custos.iam.service.User;
import org.apache.custos.integration.core.ServiceCallback;
import org.apache.custos.integration.core.ServiceException;
import org.apache.custos.integration.core.endpoint.TargetEndpoint;

/**
 * The client class used to connect IAM admin services
 */
public class IamAdminServiceClient {
    private ManagedChannel managedChannel;
    private IamAdminServiceGrpc.IamAdminServiceStub iamAdminServiceStub;

    private String iamAdminServiceAddress;
    private int port;


    public IamAdminServiceClient(TargetEndpoint targetEndpoint) {
        this.iamAdminServiceAddress = targetEndpoint.getDnsName();
        this.port = targetEndpoint.getPort();
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

        managedChannel = ManagedChannelBuilder.forAddress(
                this.iamAdminServiceAddress, this.port).usePlaintext(true).build();


        iamAdminServiceStub = IamAdminServiceGrpc.newStub(managedChannel);
        iamAdminServiceStub.addUser(user, observer);


    }


}
