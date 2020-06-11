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

package org.apache.custos.resource.secret.manager.adaptor.outbound;

import com.google.protobuf.GeneratedMessageV3;

/**
 * This class creates PasswordCredentials from gRPC PasswordCredential
 */
public class PasswordCredential extends ResourceCredential {

    private String password;

    public PasswordCredential(GeneratedMessageV3 message) {
        super(message);
        if (message instanceof org.apache.custos.resource.secret.service.PasswordCredential) {
          this.password =   ((org.apache.custos.resource.secret.service.PasswordCredential) message).getPassword();
        }
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }
}
