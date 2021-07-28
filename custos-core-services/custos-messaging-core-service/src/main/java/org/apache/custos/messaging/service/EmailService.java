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
import org.apache.custos.messaging.events.email.EmailSender;
import org.apache.custos.messaging.mapper.EmailMapper;
import org.apache.custos.messaging.persistance.model.EmailBodyParams;
import org.apache.custos.messaging.persistance.repository.EmailBodyParamsRepository;
import org.apache.custos.messaging.persistance.repository.EmailReceiversRepository;
import org.apache.custos.messaging.persistance.repository.EmailTemplateRepository;
import org.lognet.springboot.grpc.GRpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;

import java.util.*;

@GRpcService
public class EmailService extends EmailServiceGrpc.EmailServiceImplBase {

    private static final Logger LOGGER = LoggerFactory.getLogger(EmailService.class);

    @Autowired
    private EmailTemplateRepository emailTemplateRepository;

    @Autowired
    private EmailBodyParamsRepository emailBodyParamsRepository;

    @Autowired
    private EmailReceiversRepository emailReceiversRepository;

    @Value("${mail.smtp.auth:true}")
    private String mailSmtpAuth;

    @Value("${mail.smtp.starttls.enable:true}")
    private String mailSmtpStarttlsEnable;

    @Value("${mail.smtp.host:smtp.gmail.com}")
    private String mailSmtpHost;

    @Value("${mail.smtp.port:587}")
    private String mailSmtpPort;

    @Value("${mail.smtp.ssl.trust:smtp.gmail.com}")
    private String mailSmtpSslTrust;

    @Value("${mail.sender.username:custosemailagent@gmail.com}")
    private String senderUserName;
    @Value("${mail.sender.password}")
    private String senderPassword;


    @Override
    public void send(EmailMessageSendingRequest request, StreamObserver<Status> responseObserver) {
        try {
            long tenantId = request.getTenantId();
            CustosEvent event = request.getMessage().getCustosEvent();

            Optional<org.apache.custos.messaging.persistance.model.EmailTemplate> emailTemplate = emailTemplateRepository
                    .findByTenantIdAndCustosEvent(tenantId, event.name());
            if (emailTemplate.isPresent()) {
                org.apache.custos.messaging.persistance.model.EmailTemplate template = emailTemplate.get();

                String subject = template.getSubject();
                String body = template.getBody();
                Set<EmailBodyParams> emailBodyParams = template.getBodyParams();
                Map<String, String> bodyValues = request.getMessage().getParametersMap();
                LOGGER.info(body);
                for (EmailBodyParams val : emailBodyParams) {
                    if (bodyValues.containsKey(val.getValue())) {
                        body = body.replace(val.getValue(), bodyValues.get(val.getValue()));
                    }
                }
                LOGGER.info(body);

                Properties properties = new Properties();
                properties.put("mail.smtp.auth", mailSmtpAuth);
                properties.put("mail.smtp.starttls.enable", mailSmtpStarttlsEnable);
                properties.put("mail.smtp.host", mailSmtpHost);
                properties.put("mail.smtp.port", mailSmtpPort);
                properties.put("mail.smtp.ssl.trust", mailSmtpSslTrust);

                List<String> emails = request.getMessage().getReceiverEmailList();
                EmailSender.sendEmail(properties, senderUserName, senderPassword, subject, body, (String[]) emails.toArray());

            }

            org.apache.custos.messaging.email.service.Status status = org.apache.custos.messaging.email.service.Status
                    .newBuilder()
                    .setStatus(true)
                    .build();
            responseObserver.onNext(status);
            responseObserver.onCompleted();

        } catch (Exception ex) {
            String msg = " Error occurred while sending email reason : " + ex.getMessage();
            LOGGER.error(msg, ex);
            responseObserver.onError(io.grpc.Status.INTERNAL.withDescription(msg).asRuntimeException());
        }
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
