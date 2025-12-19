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

import org.apache.custos.signer.service.auth.OidcTokenValidator;
import org.apache.custos.signer.service.model.ClientSshConfigEntity;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * COmanage principal validator.
 * <p>
 * TODO deny by default until COmanage integration is implemented.
 */
public class COmanagePrincipalValidator implements PrincipalValidator {

    private static final Logger LOGGER = LoggerFactory.getLogger(COmanagePrincipalValidator.class);

    public static final String REASON_COMANAGE_NOT_IMPLEMENTED = "COMANAGE_NOT_IMPLEMENTED";

    private final SignerValidationProperties.COmanageConfig config;

    public COmanagePrincipalValidator(SignerValidationProperties.COmanageConfig config) {
        this.config = config;
        if (config.getRegistryUrl() == null || config.getRegistryUrl().trim().isEmpty()) {
            LOGGER.error("COmanage registry URL not configured. All sign requests will be denied.");

        } else {
            LOGGER.warn("COmanage validator initialized with registry: {} (integration not yet implemented)", config.getRegistryUrl());
            throw new UnsupportedOperationException("COmanage integration not yet implemented");
        }
    }

    @Override
    public PrincipalValidationResult validate(String tenantId, String clientId, String requestedPrincipal,
                                              OidcTokenValidator.UserIdentity tokenIdentity, ClientSshConfigEntity clientConfig) {

        // TODO call COmanage, map token identity -> allowed unix principals
        return PrincipalValidationResult.deny(REASON_COMANAGE_NOT_IMPLEMENTED);
    }
}


