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

package org.apache.custos.messaging.mapper;

import org.apache.custos.messaging.email.service.CustosEvent;
import org.apache.custos.messaging.email.service.EmailEnablingRequest;
import org.apache.custos.messaging.persistance.model.EmailBodyParams;
import org.apache.custos.messaging.persistance.model.EmailReceivers;
import org.apache.custos.messaging.persistance.model.EmailTemplate;

import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/**
 * Mapping messages between gRPC and data entity.
 */
public class EmailMapper {

    public static EmailTemplate transform(EmailEnablingRequest emailEnablingRequest) {
        EmailTemplate emailTemplate = new EmailTemplate();
        emailTemplate.setTenantId(emailEnablingRequest.getTenantId());
        emailTemplate.setSubject(emailEnablingRequest.getEmailTemplate().getSubject());
        emailTemplate.setCustosEvent(emailEnablingRequest.getEmailTemplate().getCustosEvent().name());
        List<String> bodyPrams = emailEnablingRequest.getEmailTemplate().getBodyParamsList();
        List<String> userEmailReceivers = emailEnablingRequest.getEmailTemplate().getReceivingUsersList();
        List<String> groupEmailReceivers = emailEnablingRequest.getEmailTemplate().getReceivingGroupsList();
        Set<EmailBodyParams> paramSet = new HashSet<>();
        Set<EmailReceivers> emailReceivers = new HashSet<>();
        bodyPrams.forEach(parm -> {
            EmailBodyParams emailBodyParams = new EmailBodyParams();
            emailBodyParams.setValue(parm);
            emailBodyParams.setEmailTemplate(emailTemplate);
            paramSet.add(emailBodyParams);
        });
        userEmailReceivers.forEach(user -> {
            EmailReceivers receivers = new EmailReceivers();
            receivers.setUserId(user);
            receivers.setUserType("USER");
            receivers.setEmailTemplate(emailTemplate);
            emailReceivers.add(receivers);
        });

        groupEmailReceivers.forEach(grp -> {
            EmailReceivers receivers = new EmailReceivers();
            receivers.setUserId(grp);
            receivers.setUserType("GROUP");
            receivers.setEmailTemplate(emailTemplate);
            emailReceivers.add(receivers);
        });

        emailTemplate.setBodyParams(paramSet);
        emailTemplate.setEmailReceivers(emailReceivers);
        return emailTemplate;
    }

    public static org.apache.custos.messaging.email.service.EmailTemplate transform(EmailTemplate emailTemplate) {

        Set<EmailReceivers> emailReceivers = emailTemplate.getEmailReceivers();
        Set<EmailBodyParams> emailBodyParams = emailTemplate.getBodyParams();

        List<String> receivingUsers = new ArrayList<>();
        List<String> receivingGroups = new ArrayList<>();
        List<String> bodyParams = new ArrayList<>();
        emailReceivers.forEach(email -> {
            if (email.getUserType().equals("USER")) {
                receivingUsers.add(email.getUserId());
            } else {
                receivingGroups.add(email.getUserId());
            }
        });

        emailBodyParams.forEach(parm -> {
            bodyParams.add(parm.getValue());

        });

        org.apache.custos.messaging.email.service.EmailTemplate template = org.apache.custos.messaging.email.service
                .EmailTemplate
                .newBuilder()
                .setTemplateId(emailTemplate.getId())
                .setCustosEvent(CustosEvent.valueOf(emailTemplate.getCustosEvent()))
                .setSubject(emailTemplate.getSubject())
                .addAllReceivingUsers(receivingUsers)
                .addAllReceivingGroups(receivingGroups)
                .addAllBodyParams(bodyParams)
                .build();
        return template;

    }


}
