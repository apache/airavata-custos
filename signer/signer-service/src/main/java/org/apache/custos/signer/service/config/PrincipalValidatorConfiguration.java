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
 */
package org.apache.custos.signer.service.config;

import org.apache.custos.signer.service.validation.COmanagePrincipalValidator;
import org.apache.custos.signer.service.validation.NoOpPrincipalValidator;
import org.apache.custos.signer.service.validation.PrincipalValidator;
import org.apache.custos.signer.service.validation.SignerValidationProperties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.env.Environment;
import org.springframework.core.env.Profiles;

@Configuration
public class PrincipalValidatorConfiguration {

    private static final Logger LOGGER = LoggerFactory.getLogger(PrincipalValidatorConfiguration.class);

    @Bean
    public PrincipalValidator principalValidator(SignerValidationProperties props, Environment env) {
        String selected = props.getPrincipalValidator() != null ? props.getPrincipalValidator().trim().toLowerCase() : "comanage";
        String[] activeProfiles = env.getActiveProfiles();

        switch (selected) {
            case "comanage":
                LOGGER.info("Using COmanagePrincipalValidator");
                return new COmanagePrincipalValidator(props.getComanage());

            case "no-op":
                if (!env.acceptsProfiles(Profiles.of("test"))) {
                    throw new IllegalStateException("NoOp principal validator can only be enabled when the 'test' Spring profile is active. " +
                            "Current active profiles: " + (activeProfiles.length > 0 ? String.join(",", activeProfiles) : "none"));
                }
                LOGGER.warn("Principal validation is configured as NoOp under 'test' profile. This must not be used in production.");
                return new NoOpPrincipalValidator();

            default:
                throw new IllegalArgumentException("Unknown principal validator: " + selected);
        }
    }
}

