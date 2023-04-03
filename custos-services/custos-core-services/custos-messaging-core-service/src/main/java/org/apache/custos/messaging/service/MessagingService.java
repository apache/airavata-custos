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
import org.apache.custos.messaging.events.publisher.MessageProducer;
import org.apache.custos.messaging.mapper.MessagingMapper;
import org.apache.custos.messaging.persistance.model.MessagingMetadata;
import org.apache.custos.messaging.persistance.repository.MessagingMetadataRepository;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.Optional;

@GRpcService
public class MessagingService extends MessagingServiceGrpc.MessagingServiceImplBase {
    private static final Logger LOGGER = LoggerFactory.getLogger(MessagingService.class);

    @Autowired
    private MessagingMetadataRepository messagingMetadataRepository;

    @Autowired(required=false)
    private MessageProducer messageProducer;

    @Override
    public void publish(Message request, StreamObserver<Status> responseObserver) {
        try {
            Optional<MessagingMetadata> metadata = messagingMetadataRepository.findByTenantId(request.getTenantId());
            if (metadata.isPresent()) {
                messageProducer.publish(metadata.get().getTopic(), request);
            }
            Status status = Status.newBuilder().setStatus(true).build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            String msg = "Error occurred while publishing Messaging service" + ex.getMessage();
            LOGGER.error(msg, ex);
        }
    }

    @Override
    public void enable(MessageEnablingRequest request, StreamObserver<MessageEnablingResponse> responseObserver) {
        try {
            Optional<MessagingMetadata> metadata = messagingMetadataRepository.findByTenantId(request.getTenantId());
            if (metadata.isEmpty()) {
                MessagingMetadata meta = MessagingMapper.transform(request);
                messagingMetadataRepository.save(meta);
                MessageEnablingResponse response = MessageEnablingResponse
                        .newBuilder()
                        .setTopic(meta.getTopic())
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            } else {
                MessageEnablingResponse response = MessageEnablingResponse
                        .newBuilder()
                        .setTopic(metadata.get().getTopic())
                        .build();
                responseObserver.onNext(response);
                responseObserver.onCompleted();
            }

        } catch (Exception ex) {
            String msg = "Enabling messaging for client " + request.getClientId()
                    + " failed, reason: " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }
}
