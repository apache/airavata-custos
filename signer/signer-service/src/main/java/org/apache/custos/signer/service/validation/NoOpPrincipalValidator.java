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
 * No-op principal validator.
 * <p>
 * This must only be used for local testing/dev.
 */
public class NoOpPrincipalValidator implements PrincipalValidator {

    private static final Logger LOGGER = LoggerFactory.getLogger(NoOpPrincipalValidator.class);

    public NoOpPrincipalValidator() {
        LOGGER.warn("NoOpPrincipalValidator is ENABLED. This bypasses principal authorization and must not be used in production.");
    }

    @Override
    public PrincipalValidationResult validate(String tenantId, String clientId, String requestedPrincipal,
                                              OidcTokenValidator.UserIdentity tokenIdentity, ClientSshConfigEntity clientConfig) {
        return PrincipalValidationResult.allow(requestedPrincipal);
    }
}


