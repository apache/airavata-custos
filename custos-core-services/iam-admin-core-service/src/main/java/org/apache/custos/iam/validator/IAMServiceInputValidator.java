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

package org.apache.custos.iam.validator;


import org.apache.custos.iam.exceptions.MissingParameterException;
import org.apache.custos.iam.service.*;

/**
 * This class validates the  requests
 */
public class IAMServiceInputValidator {

    /**
     * Input parameter validater
     * @param methodName
     * @param obj
     * @return
     */
    public static void validate(String methodName, Object obj) {

        switch (methodName) {
            case "setUPTenant":
                validateSetUPTenant(obj);
                break;
            case "isUsernameAvailable":
                validateIsUsernameAvailable(obj);
                break;
            case "registerUser":
                validateRegisterUser(obj);
                break;
            case "enableUser":
            case "isUserEnabled":
            case "isUserExist":
            case "getUser":
            case "deleteUser":
                validateUserAccess(obj);
                break;
            case "getUsers":
                validateGetUsers(obj);
                break;
            case "resetPassword":
                validateResetPassword(obj);
                break;
            case "findUsers":
                validateFindUsers(obj);
            case "updateUserProfile":
                validateUpdateUserProfile(obj);
            case "addRoleToUser":
            case "deleteRoleFromUser":
                validateRoleOperationsRequest(obj);
            default:

        }

    }

    private static boolean validateSetUPTenant(Object obj) {
        if (obj instanceof SetUpTenantRequest) {
            SetUpTenantRequest request = (SetUpTenantRequest) obj;
            if (request.getTenantId() == null || request.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getAdminFirstname() == null || request.getAdminFirstname().trim().equals("")) {
                throw new MissingParameterException("Admin firstname should not be null", null);
            }
            if (request.getAdminLastname() == null || request.getAdminLastname().trim().equals("")) {
                throw new MissingParameterException("Admin lastname should not be null", null);
            }
            if (request.getAdminEmail() == null || request.getAdminEmail().trim().equals("")) {
                throw new MissingParameterException("Admin email should not be null", null);
            }
            if (request.getAdminPassword() == null || request.getAdminPassword().trim().equals("")) {
                throw new MissingParameterException("Admin password should not be null", null);
            }
            if (request.getAdminUsername() == null || request.getAdminUsername().trim().equals("")) {
                throw new MissingParameterException("Admin username should not be null", null);
            }
            if (request.getTenantName()== null || request.getTenantName().trim().equals("")) {
                throw new MissingParameterException("Tenant name should not be null", null);
            }
            if (request.getTenantURL()== null || request.getTenantURL().trim().equals("")) {
                throw new MissingParameterException("Tenant URL should not be null", null);
            }
            if (request.getRequesterEmail()== null || request.getRequesterEmail().trim().equals("")) {
                throw new MissingParameterException("Tenant requester email should not be null", null);
            }



        } else {
            throw new RuntimeException("Unexpected input type for method setUPTenant");
        }
        return true;
    }

    private static boolean validateIsUsernameAvailable(Object obj) {
        if (obj instanceof IsUsernameAvailableRequest) {
            IsUsernameAvailableRequest request = (IsUsernameAvailableRequest)obj;
            if (request.getTenantId() == null || request.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getAccessToken() == null || request.getAccessToken().trim().equals("")) {
                throw new MissingParameterException("Access token should not be null", null);
            }

            if (request.getUserName() == null || request.getUserName().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method isUsernameAvailable");
        }
        return true;
    }

    private static boolean validateRegisterUser(Object obj) {
        if (obj instanceof RegisterUserRequest) {
            RegisterUserRequest request = (RegisterUserRequest)obj;
            if (request.getTenantId() == null || request.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getAccessToken() == null || request.getAccessToken().trim().equals("")) {
                throw new MissingParameterException("Access token should not be null", null);
            }

            if (request.getUsername() == null || request.getUsername().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

            if (request.getFirstName() == null || request.getFirstName().trim().equals("")) {
                throw new MissingParameterException("Firstname should not be null", null);
            }
            if (request.getLastName() == null || request.getLastName().trim().equals("")) {
                throw new MissingParameterException("Lastname should not be null", null);
            }
            if (request.getEmail() == null || request.getEmail().trim().equals("")) {
                throw new MissingParameterException("Email should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method registerUser");
        }
        return true;
    }

    private static boolean validateUserAccess(Object obj) {
        if (obj instanceof UserAccessInfo) {
            UserAccessInfo request = (UserAccessInfo)obj;
            if (request.getTenantId() == null || request.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getAccessToken() == null || request.getAccessToken().trim().equals("")) {
                throw new MissingParameterException("Access token should not be null", null);
            }

            if (request.getUsername() == null || request.getUsername().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method userAccess");
        }
        return true;
    }


    private static boolean validateGetUsers(Object obj) {
        if (obj instanceof GetUsersRequest) {
        GetUsersRequest request = (GetUsersRequest)obj;
        UserAccessInfo info =  request.getInfo();
            if (info.getTenantId() == null || info.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (info.getAccessToken() == null || info.getAccessToken().trim().equals("")) {
                throw new MissingParameterException("Access token should not be null", null);
            }

            if (info.getUsername() == null || info.getUsername().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

            if (request.getSearch() == null || request.getSearch().trim().equals("")) {
                throw new MissingParameterException("Search parameter should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method getUsers");
        }
        return true;
    }

    private static boolean validateResetPassword(Object obj) {
        if (obj instanceof ResetUserPassword) {
            ResetUserPassword request = (ResetUserPassword) obj;
            UserAccessInfo info = request.getInfo();
            if (info.getTenantId() == null || info.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (info.getAccessToken() == null || info.getAccessToken().trim().equals("")) {
                throw new MissingParameterException("Access token should not be null", null);
            }

            if (info.getUsername() == null || info.getUsername().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

            if (request.getPassword() == null || request.getPassword().trim().equals("")) {
                throw new MissingParameterException("Password should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method resetPassword");
        }
        return true;
    }

    private static boolean validateFindUsers(Object obj) {
        if (obj instanceof FindUsersRequest) {
            FindUsersRequest request = (FindUsersRequest) obj;
            UserAccessInfo info = request.getInfo();
            if (info.getTenantId() == null || info.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (info.getAccessToken() == null || info.getAccessToken().trim().equals("")) {
                throw new MissingParameterException("Access token should not be null", null);
            }

            if (info.getUsername() == null || info.getUsername().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

            if (request.getEmail() == null || request.getEmail().trim().equals("")) {
                throw new MissingParameterException("Email should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method findUsers");
        }
        return true;
    }

    private static boolean validateUpdateUserProfile(Object obj) {
        if (obj instanceof UpdateUserProfileRequest) {
            UpdateUserProfileRequest re = (UpdateUserProfileRequest)obj;
            User request = re.getUser();
            if (request.getTenantId() == null || request.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (re.getAccessToken() == null || re.getAccessToken().trim().equals("")) {
                throw new MissingParameterException("Access token should not be null", null);
            }

            if (request.getUsername() == null || request.getUsername().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

            if (request.getFirstName() == null || request.getFirstName().trim().equals("")) {
                throw new MissingParameterException("Firstname should not be null", null);
            }
            if (request.getLastName() == null || request.getLastName().trim().equals("")) {
                throw new MissingParameterException("Lastname should not be null", null);
            }
            if (request.getEmail() == null || request.getEmail().trim().equals("")) {
                throw new MissingParameterException("Email should not be null", null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method updateUserProfile");
        }
        return true;
    }



    private static boolean validateRoleOperationsRequest(Object obj) {
        if (obj instanceof RoleOperationsUserRequest) {
            RoleOperationsUserRequest request = (RoleOperationsUserRequest)obj;
            if (request.getTenantId() == null || request.getTenantId().trim().equals("")) {
                throw new MissingParameterException("Tenant Id should not be null", null);
            }

            if (request.getAdminUsername() == null || request.getAdminUsername().trim().equals("")) {
                throw new MissingParameterException("Admin username should not be null", null);
            }

            if (request.getUsername() == null || request.getUsername().trim().equals("")) {
                throw new MissingParameterException("Username should not be null", null);
            }

            if (request.getPassword() == null || request.getPassword().trim().equals("")) {
                throw new MissingParameterException("Password should not be null", null);
            }
            if (request.getRole() == null || request.getRole().trim().equals("")) {
                throw new MissingParameterException("Role should not be null", null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method roleOperationsRequest");
        }
        return true;
    }

}