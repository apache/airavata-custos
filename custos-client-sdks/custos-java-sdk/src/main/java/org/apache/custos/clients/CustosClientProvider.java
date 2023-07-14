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

package org.apache.custos.clients;

import org.apache.custos.agent.management.client.AgentManagementClient;
import org.apache.custos.group.management.client.GroupManagementClient;
import org.apache.custos.identity.management.client.IdentityManagementClient;
import org.apache.custos.resource.secret.management.client.ResourceSecretManagementAgentClient;
import org.apache.custos.resource.secret.management.client.ResourceSecretManagementClient;
import org.apache.custos.sharing.management.client.SharingManagementClient;
import org.apache.custos.tenant.manamgement.client.TenantManagementClient;
import org.apache.custos.user.management.client.UserManagementClient;


import java.io.IOException;

/**
 * The class responsible for provides the Custos clients
 */
public class CustosClientProvider {


    private String serverHost;

    private int serverPort;

    private String clientId;

    private String clientSec;


    private boolean plainText;


    private CustosClientProvider(String serverHost, int serverPort, String clientId, String clientSec, boolean plainText) {
        this.serverHost = serverHost;
        this.serverPort = serverPort;
        this.clientId = clientId;
        this.clientSec = clientSec;
        this.plainText = plainText;
    }


    public IdentityManagementClient getIdentityManagementClient() throws IOException {
        return new IdentityManagementClient(this.serverHost, this.serverPort, this.clientId, this.clientSec);

    }


    public TenantManagementClient getTenantManagementClient() throws IOException {
        return new TenantManagementClient(this.serverHost, this.serverPort, this.clientId, this.clientSec);
    }

    public AgentManagementClient getAgentManagementClient() throws IOException {
        return new AgentManagementClient(this.serverHost, this.serverPort, this.clientId, this.clientSec);
    }

    public GroupManagementClient getGroupManagementClient() throws IOException {
        return new GroupManagementClient(this.serverHost, this.serverPort, this.clientId, this.clientSec);
    }

    public ResourceSecretManagementClient getResourceSecretManagementClient() throws IOException {
        return new ResourceSecretManagementClient(this.serverHost, this.serverPort, this.clientId, this.clientSec);
    }

    public SharingManagementClient getSharingManagementClient() throws IOException {
        return new SharingManagementClient(this.serverHost, this.serverPort, this.clientId, this.clientSec);
    }

    public UserManagementClient getUserManagementClient() throws IOException {
        return new UserManagementClient(this.serverHost, this.serverPort, this.clientId, this.clientSec, this.plainText);
    }

    public ResourceSecretManagementClient getResourceSecretManagementClientForAgents() throws IOException {
        return  new ResourceSecretManagementAgentClient(this.serverHost, this.serverPort, this.clientId, this.clientSec);
    }


    public static class Builder {

        private String serverHost;
        private int serverPort;
        private String clientId;

        private String clientSec;

        private boolean plainText;


        public Builder() {
        }

        public Builder setServerHost(String serverHost) {
            this.serverHost = serverHost;
            return this;
        }

        public Builder setServerPort(int serverPort) {
            this.serverPort = serverPort;
            return this;
        }

        public Builder setClientId(String clientId) {
            this.clientId = clientId;
            return this;
        }

        public Builder setClientSec(String clientSec) {
            this.clientSec = clientSec;
            return this;
        }

        public Builder usePlainText(boolean plainText) {
            this.plainText = plainText;
            return this;
        }


        public CustosClientProvider build() {
            if (serverHost == null || serverPort == 0 || clientId == null || clientSec == null) {
                throw new NullPointerException("Server Host, Server Port, clientId, clientSec   should not be null");
            }

            return new CustosClientProvider(this.serverHost, this.serverPort, this.clientId, this.clientSec, this.plainText);

        }

    }


}
