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

import org.apache.custos.identity.management.client.IdentityManagementClient;
import org.apache.custos.tenant.manamgement.client.SuperAdminOperationsClient;
import org.apache.custos.tenant.manamgement.client.TenantManagementClient;

import javax.net.ssl.SSLException;
import java.io.IOException;

/**
 * The class responsible for provides the Custos clients
 */
public class CustosClientProvider {


    private String serverHost;

    private int serverPort;

    private String certFilePath;


    private CustosClientProvider(String serverHost, int serverPort, String certFilePath) {
        this.serverHost = serverHost;
        this.serverPort = serverPort;
        this.certFilePath = certFilePath;
    }


    public IdentityManagementClient getIdentityManagementClient(String clientId, String clientSecret) throws SSLException {
        return new IdentityManagementClient(this.serverHost, this.serverPort, clientId, clientSecret, this.certFilePath);
    }


    public TenantManagementClient getTenantManagementClient(String clientId, String clientSecret) throws IOException {
        return new TenantManagementClient(this.serverHost, this.serverPort, clientId, clientSecret);
    }


//    public SuperAdminOperationsClient getAdminTenantRegistrationClient() throws IOException {
//        return new SuperAdminOperationsClient(this.serverHost, this.serverPort);
//    }


    public static class Builder {

        private String serverHost;
        private int serverPort;
        private String certFilePath;

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

        public Builder setCertFilePath(String certFilePath) {

            this.certFilePath = certFilePath;
            return this;
        }

        public CustosClientProvider build() {
            if (serverHost == null || serverPort == 0 || certFilePath == null) {
                throw new NullPointerException("Server Host, Server Port and Cert File Path should not be null");
            }

            return new CustosClientProvider(this.serverHost, this.serverPort, this.certFilePath);

        }

    }


}
