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

/**
 * Validates whether an authenticated token identity is authorized to request a given Unix principal.
 *
 * <ul>
 *   <li><b>token identity</b>: issuer + subject (and other claims), NOT necessarily a Unix account</li>
 *   <li><b>requested principal</b>: Unix account name requested by the client app</li>
 *   <li><b>validated principal</b>: Unix account name approved by an external source of truth</li>
 * </ul>
 */
public interface PrincipalValidator {

    // Even with the pluggable source of truth this check will be expensive
    // TODO - Keep an account mapping within the Custos-Signer
    //  Options to sync account name changes
    //   1. invoke a background thread to look up the account name requested if there's an issue can deal with the callback
    //   2. Periodic lookup --> this is risky
    //  OR keep the expensive external call?
    PrincipalValidationResult validate(String tenantId, String clientId, String requestedPrincipal,
                                       OidcTokenValidator.UserIdentity tokenIdentity, ClientSshConfigEntity clientConfig);
}


