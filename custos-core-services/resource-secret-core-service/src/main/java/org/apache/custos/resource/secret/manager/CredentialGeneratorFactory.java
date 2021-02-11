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

package org.apache.custos.resource.secret.manager;

import com.google.protobuf.GeneratedMessageV3;
import org.apache.custos.resource.secret.service.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

/**
 * This class is responsible for generate secrets
 */
@Component
public class CredentialGeneratorFactory {

    private static final Logger LOGGER = LoggerFactory.getLogger(CredentialGeneratorFactory.class);


    public Credential getCredential(GeneratedMessageV3 message) throws Exception {

        if (message instanceof SSHCredential) {
            return new org.apache.custos.resource.secret.manager.adaptor.outbound.SSHCredential(message);
        } else if (message instanceof CertificateCredential) {
            return new org.apache.custos.resource.secret.manager.adaptor.outbound.CertificateCredential(message);
        } else if (message instanceof PasswordCredential) {
            return new org.apache.custos.resource.secret.manager.adaptor.outbound.PasswordCredential(message);
        } else if (message instanceof  KVCredential){
            return new org.apache.custos.resource.secret.manager.adaptor.outbound.KVCredential(message);
        }else if (message instanceof CredentialMap){
            return new org.apache.custos.resource.secret.manager.adaptor.outbound.CredentialMap(message);
        }

        return null;
    }



}
