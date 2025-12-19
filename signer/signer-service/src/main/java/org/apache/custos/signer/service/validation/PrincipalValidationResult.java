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

/**
 * Result of validating/authorizing a requested Unix principal for certificate issuance.
 *
 * @param allowed            whether the request is allowed
 * @param validatedPrincipal Unix principal to sign (only meaningful when allowed=true)
 * @param reasonCode         stable reason code for audit
 */
public record PrincipalValidationResult(boolean allowed, String validatedPrincipal, String reasonCode) {

    public static final String REASON_ALLOWED = "ALLOWED";
    public static final String REASON_NOT_ALLOWED = "NOT_ALLOWED";
    public static final String REASON_NOT_IMPLEMENTED = "NOT_IMPLEMENTED";
    public static final String REASON_INVALID_PRINCIPAL_FORMAT = "INVALID_PRINCIPAL_FORMAT";

    public static PrincipalValidationResult allow(String validatedPrincipal) {
        return new PrincipalValidationResult(true, validatedPrincipal, REASON_ALLOWED);
    }

    public static PrincipalValidationResult deny(String reasonCode) {
        return new PrincipalValidationResult(false, null, reasonCode != null ? reasonCode : REASON_NOT_ALLOWED);
    }
}


