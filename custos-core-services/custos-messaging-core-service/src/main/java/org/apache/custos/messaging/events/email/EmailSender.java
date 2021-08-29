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

package org.apache.custos.messaging.events.email;

import org.slf4j.LoggerFactory;

import javax.mail.*;
import javax.mail.internet.InternetAddress;
import javax.mail.internet.MimeMessage;
import java.util.Properties;


public class EmailSender {
    private static final org.slf4j.Logger LOGGER = LoggerFactory.getLogger(EmailSender.class);

    public static void sendEmail(Properties prop, String senderEmail, String senderPassword, String subject,
                                 String body, String[] recipient) throws MessagingException {

        Session session = Session.getInstance(prop, new Authenticator() {
            @Override
            protected PasswordAuthentication getPasswordAuthentication() {
                return new PasswordAuthentication(senderEmail, senderPassword);
            }
        });

//            Session session = Session.getDefaultInstance(prop);
        Message message = new MimeMessage(session);
        message.setFrom(new InternetAddress(senderEmail));

        Address[] addresses = new Address[recipient.length];

        for (int i = 0; i < recipient.length; i++) {
            addresses[i] = new InternetAddress(recipient[i]);
        }
        message.setRecipients(
                Message.RecipientType.TO, addresses);
        message.setSubject(subject);

        message.setContent(body, "text/html");

        Transport.send(message);
    }


}
