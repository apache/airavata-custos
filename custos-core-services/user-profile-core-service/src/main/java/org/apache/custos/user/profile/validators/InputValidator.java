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

package org.apache.custos.user.profile.validators;

import org.apache.custos.core.services.commons.Validator;
import org.apache.custos.core.services.commons.exceptions.MissingParameterException;
import org.apache.custos.user.profile.service.*;

/**
 * Validate inputs
 */
public class InputValidator implements Validator {
    @Override
    public void validate(String methodName, Object obj) {
        switch (methodName) {
            case "createUserProfile":
            case "updateUserProfile":
                validateUserProfile(obj, methodName);
                break;
            case "getUserProfile":
            case "deleteUserProfile":
            case "getUserProfileAuditTrails":
                validateUsernameAndTenantId(obj, methodName);
                break;
            case "validateGetAllUserProfiles":
                validateGetAllUserProfiles(obj, methodName);
                break;
            default:

        }
    }


    private boolean validateUserProfile(Object obj, String method) {
        if (obj instanceof UserProfile) {
            UserProfile profile = (UserProfile) obj;
            if (method.equals("updateUserProfile")) {
                if (profile.getUserId() == null || profile.getUserId().equals("")) {
                    throw new MissingParameterException("userId should not be null", null);
                }
            }
            if (profile.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profile.getUsername() == null || profile.getUsername().equals("")) {
                throw new MissingParameterException("username should not be null", null);
            }

            if (profile.getFirstName() == null || profile.getFirstName().equals("")) {
                throw new MissingParameterException("firstName should not be null", null);
            }
            if (profile.getLastName() == null || profile.getLastName().equals("")) {
                throw new MissingParameterException("lastName should not be null", null);
            }
            if (profile.getEmail() == null || profile.getEmail().equals("")) {
                throw new MissingParameterException("emailAddress should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method" + method);
        }
        return true;
    }


    private boolean validateUsernameAndTenantId(Object obj, String method) {
        if (method.equals("getUserProfile")) {
            GetUserProfileRequest profileRequest = (GetUserProfileRequest) obj;
            if (profileRequest.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileRequest.getUsername() == null || profileRequest.getUsername().equals("")) {
                throw new MissingParameterException("username should not be null", null);
            }
        } else if (method.equals("deleteUserProfile")) {

            DeleteUserProfileRequest profileRequest = (DeleteUserProfileRequest) obj;
            if (profileRequest.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileRequest.getUsername() == null || profileRequest.getUsername().equals("")) {
                throw new MissingParameterException("username should not be null", null);
            }

        } else if (method.equals("getUserProfileAuditTrails")) {

            GetUpdateAuditTrailRequest profileRequest = (GetUpdateAuditTrailRequest) obj;
            if (profileRequest.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileRequest.getUsername() == null || profileRequest.getUsername().equals("")) {
                throw new MissingParameterException("username should not be null", null);
            }

        }

        return true;

    }

    private boolean validateGetAllUserProfiles(Object obj, String method) {
        if (obj instanceof GetAllUserProfilesRequest) {
            GetAllUserProfilesRequest profileReq = (GetAllUserProfilesRequest) obj;
            if (profileReq.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method" + method);
        }
        return true;

    }
}
