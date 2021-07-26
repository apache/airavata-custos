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
import org.apache.custos.messaging.mapper.EmailMapper;
import org.apache.custos.messaging.persistance.repository.EmailBodyParamsRepository;
import org.apache.custos.messaging.persistance.repository.EmailReceiversRepository;
import org.apache.custos.messaging.persistance.repository.EmailTemplateRepository;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@GRpcService
public class EmailService extends EmailServiceGrpc.EmailServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(EmailService.class);

    @Autowired
    private EmailTemplateRepository emailTemplateRepository;

    @Autowired
    private EmailBodyParamsRepository emailBodyParamsRepository;

    @Autowired
    private EmailReceiversRepository emailReceiversRepository;


    @Override
    public void send(EmailMessageSendingRequest request, StreamObserver<Status> responseObserver) {
        super.send(request, responseObserver);
    }

    @Override
    public void enable(EmailEnablingRequest request, StreamObserver<EmailTemplate> responseObserver) {
        try {

            org.apache.custos.messaging.persistance.model.EmailTemplate emailTemplate = EmailMapper.transform(request);
            emailTemplate.setStatus(true);

            if (request.getEmailTemplate().getTemplateId() > 0) {
                Optional<org.apache.custos.messaging.persistance.model.EmailTemplate> optionalEmailTemplate =
                        emailTemplateRepository.findById(request.getEmailTemplate().getTemplateId());
                if (optionalEmailTemplate.isEmpty()) {
                    String msg = " Cannot find EmailTemplate with id " + optionalEmailTemplate.get().getId();
                    LOGGER.error(msg);
                    responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
                    return;
                }

                org.apache.custos.messaging.persistance.model.EmailTemplate exTemplate = optionalEmailTemplate.get();
                exTemplate.getBodyParams().forEach(param -> {
                    emailBodyParamsRepository.delete(param);
                });

                exTemplate.getEmailReceivers().forEach(param -> {
                    emailReceiversRepository.delete(param);
                });
                emailTemplate.setId(exTemplate.getId());
                emailTemplate.setCreatedAt(exTemplate.getCreatedAt());
            }

            org.apache.custos.messaging.persistance.model.EmailTemplate saved = emailTemplateRepository.save(emailTemplate);

            EmailTemplate emailTemp = EmailMapper.transform(saved);
            responseObserver.onNext(emailTemp);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Error occurred while creating email template reason : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }

    @Override
    public void getTemplates(FetchEmailTemplatesRequest request, StreamObserver<FetchEmailTemplatesResponse> responseObserver) {
        try {
            List<org.apache.custos.messaging.persistance.model.EmailTemplate> emailTemplateList =
                    emailTemplateRepository.findByTenantId(request.getTenantId());
            List<EmailTemplate> emailTemplates = new ArrayList<>();
            if (emailTemplateList != null && !emailTemplateList.isEmpty()) {
                emailTemplateList.forEach(temp -> {
                    emailTemplates.add(EmailMapper.transform(temp));
                });
            }
            FetchEmailTemplatesResponse response = FetchEmailTemplatesResponse
                    .newBuilder()
                    .addAllTemplates(emailTemplates)
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Error occurred while get templates reason : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
    }

    @Override
    public void disable(EmailDisablingRequest request, StreamObserver<Status> responseObserver) {
        try {
            long templateId = request.getEmailTemplate().getTemplateId();
            Optional<org.apache.custos.messaging.persistance.model.EmailTemplate> optionalEmailTemplate =
                    emailTemplateRepository.findById(templateId);
            if (optionalEmailTemplate.isPresent()) {
                org.apache.custos.messaging.persistance.model.EmailTemplate emailTemplate = optionalEmailTemplate.get();
                emailTemplate.setStatus(false);
                emailTemplateRepository.save(emailTemplate);
            }
            responseObserver.onNext(Status.newBuilder().setStatus(true).build());
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Error occurred while disabling email template reason : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }

    }
}
