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

package org.apache.custos.messaging.client;

import io.grpc.ClientInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.custos.messaging.email.service.*;
import org.apache.custos.messaging.service.Status;
import org.apache.custos.messaging.service.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class MessagingClient {

    private ManagedChannel managedChannel;
    private MessagingServiceGrpc.MessagingServiceBlockingStub messagingServiceBlockingStub;
    private MessagingServiceGrpc.MessagingServiceStub messagingServiceFutureStub;

    private List<ClientInterceptor> clientInterceptorList;


    private EmailServiceGrpc.EmailServiceBlockingStub emailServiceBlockingStub;
    private EmailServiceGrpc.EmailServiceStub emailServiceStub;


    public MessagingClient(List<ClientInterceptor> clientInterceptorList,
                           @Value("${messaging.core.service.dns.name}") String serviceHost,
                           @Value("${messaging.core.service.port}") int servicePort) {
        this.clientInterceptorList = clientInterceptorList;
        managedChannel = ManagedChannelBuilder.forAddress(
                serviceHost, servicePort).usePlaintext().intercept(clientInterceptorList).build();
        messagingServiceBlockingStub = MessagingServiceGrpc.newBlockingStub(managedChannel);
        messagingServiceFutureStub = MessagingServiceGrpc.newStub(managedChannel);
        emailServiceStub = EmailServiceGrpc.newStub(managedChannel);
        emailServiceBlockingStub = EmailServiceGrpc.newBlockingStub(managedChannel);
    }


    public MessageEnablingResponse enableMessaging(MessageEnablingRequest messageEnablingRequest) {
        return this.messagingServiceBlockingStub.enable(messageEnablingRequest);
    }

    public Status publish(Message request) {
        return this.messagingServiceBlockingStub.publish(request);
    }

    public void sendEmailAsync(EmailMessageSendingRequest request,
                               StreamObserver<org.apache.custos.messaging.email.service.Status> streamObserver) {
        this.emailServiceStub.send(request, streamObserver);
    }


    public void publishAsync(Message request, StreamObserver<Status> streamObserver) {
        this.messagingServiceFutureStub.publish(request, streamObserver);
    }

    public EmailTemplate enableEmail(EmailEnablingRequest emailEnablingRequest) {
        return this.emailServiceBlockingStub.enable(emailEnablingRequest);
    }

    public org.apache.custos.messaging.email.service.Status disableEmail(EmailDisablingRequest emailDisablingRequest) {
        return this.emailServiceBlockingStub.disable(emailDisablingRequest);
    }

    public FetchEmailTemplatesResponse getEmailTemplates(FetchEmailTemplatesRequest fetchEmailTemplatesRequest) {
        return this.emailServiceBlockingStub.getTemplates(fetchEmailTemplatesRequest);
    }

    public FetchEmailFriendlyEventsResponse fetchEmailFriendlyEvents(FetchEmailFriendlyEvents fetchEmailFriendlyEvents) {
        return this.emailServiceBlockingStub.getEmailFriendlyEvents(fetchEmailFriendlyEvents);
    }


}
