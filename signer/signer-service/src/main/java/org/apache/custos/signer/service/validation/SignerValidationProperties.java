/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */
package org.apache.custos.signer.service.validation;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

/**
 * Principal validation configuration.
 */
@Component
@ConfigurationProperties(prefix = "signer.validation")
public class SignerValidationProperties {

    /**
     * Validator selection
     * <p>
     * Expected values: {@code comanage}, {@code no-op}.
     */
    private String principalValidator = "comanage";

    /**
     * COmanage registry configuration.
     */
    private COmanageConfig comanage = new COmanageConfig();

    public String getPrincipalValidator() {
        return principalValidator;
    }

    public void setPrincipalValidator(String principalValidator) {
        this.principalValidator = principalValidator;
    }

    public COmanageConfig getComanage() {
        return comanage;
    }

    public void setComanage(COmanageConfig comanage) {
        this.comanage = comanage;
    }

    public static class COmanageConfig {

        private String registryUrl;

        /**
         * API endpoint path (relative to registryUrl)
         */
        private String apiPath = "/registry/co_people.json";

        private String apiToken;

        /**
         * Request timeout in seconds.
         */
        private int timeoutSeconds = 10;

        private boolean verifySsl = true;

        public String getRegistryUrl() {
            return registryUrl;
        }

        public void setRegistryUrl(String registryUrl) {
            this.registryUrl = registryUrl;
        }

        public String getApiPath() {
            return apiPath;
        }

        public void setApiPath(String apiPath) {
            this.apiPath = apiPath;
        }

        public String getApiToken() {
            return apiToken;
        }

        public void setApiToken(String apiToken) {
            this.apiToken = apiToken;
        }

        public int getTimeoutSeconds() {
            return timeoutSeconds;
        }

        public void setTimeoutSeconds(int timeoutSeconds) {
            this.timeoutSeconds = timeoutSeconds;
        }

        public boolean isVerifySsl() {
            return verifySsl;
        }

        public void setVerifySsl(boolean verifySsl) {
            this.verifySsl = verifySsl;
        }
    }
}


