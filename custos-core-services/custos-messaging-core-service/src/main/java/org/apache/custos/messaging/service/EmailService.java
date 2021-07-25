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

package org.apache.custos.messaging.service;

import io.grpc.stub.StreamObserver;
import org.apache.custos.messaging.email.service.Status;
import org.apache.custos.messaging.email.service.*;
import org.lognet.springboot.grpc.GRpcService;

@GRpcService
public class EmailService extends EmailServiceGrpc.EmailServiceImplBase {


    @Override
    public void send(EmailMessageSendingRequest request, StreamObserver<Status> responseObserver) {
        super.send(request, responseObserver);
    }

    @Override
    public void enable(EmailEnablingRequest request, StreamObserver<EmailTemplate> responseObserver) {
        super.enable(request, responseObserver);
    }

    @Override
    public void getTemplates(FetchEmailTemplatesRequest request, StreamObserver<FetchEmailTemplatesResponse> responseObserver) {
        super.getTemplates(request, responseObserver);
    }

    @Override
    public void disable(EmailDisablingRequest request, StreamObserver<Status> responseObserver) {
        super.disable(request, responseObserver);
    }
}
