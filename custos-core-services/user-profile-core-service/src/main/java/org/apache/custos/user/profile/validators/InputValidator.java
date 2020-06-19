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

            case "createGroup":
                validateCreateGroup(obj, methodName);
                break;
            case "updateGroup":
                validateUpdateGroup(obj, methodName);
                break;
            case "deleteGroup":
                validateDeleteGroup(obj, methodName);
                break;
            case "getGroup":
                validateFindGroup(obj, methodName);
                break;
            case "addUserToGroup":
            case "removeUserFromGroup":
            case "hasAccess":
            case "changeUserMembershipType":
                validateGroupMembership(obj, methodName);
                break;

            case "addChildGroupToParentGroup":
            case "removeChildGroupFromParentGroup":
                validateGroupToGroupMembership(obj, methodName);

            case "getAllGroupsOfUser":
                validateGetAllGroupsOfUser(obj, methodName);
                break;
            case "getAllParentGroupsOfGroup":
                validateGetAllParentGroupsOfGroup(obj, methodName);
                break;
            case "addUserGroupMembershipType":
            case "removeUserGroupMembershipType":
                validateUserGroupMembershipTypeRequest(obj, methodName);
                break;
            case "getAllChildUsers":
            case "getAllChildGroups":
                validateGetUserAndChildGroups(obj, methodName);
                break;


            default:

        }
    }


    private boolean validateUserProfile(Object obj, String method) {
        if (obj instanceof UserProfileRequest) {
            UserProfileRequest profile = (UserProfileRequest) obj;

            if (profile.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profile.getProfile() == null) {
                throw new MissingParameterException("Profile should be valid ", null);
            }
            if (profile.getProfile().getUsername() == null || profile.getProfile().getUsername().equals("")) {
                throw new MissingParameterException("username should not be null", null);
            }

            if ((profile.getProfile().getFirstName() == null || profile.getProfile().getFirstName().equals("")) &&
                    !profile.getProfile().getType().equals(UserTypes.COMMUNITY) ) {
                throw new MissingParameterException("firstName should not be null", null);
            }
            if ((profile.getProfile().getLastName() == null || profile.getProfile().getLastName().equals("")) &&
                    !profile.getProfile().getType().equals(UserTypes.COMMUNITY)) {
                throw new MissingParameterException("lastName should not be null", null);
            }
            if ((profile.getProfile().getEmail() == null || profile.getProfile().getEmail().equals("")) &&
                    !profile.getProfile().getType().equals(UserTypes.COMMUNITY)) {
                throw new MissingParameterException("emailAddress should not be null", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + method);
        }
        return true;
    }


    private boolean validateUsernameAndTenantId(Object obj, String method) {
        if (method.equals("getUserProfile")) {
            UserProfileRequest profileRequest = (UserProfileRequest) obj;
            if (profileRequest.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileRequest.getProfile() == null) {
                throw new MissingParameterException("Profile should be valid ", null);
            }
            if (profileRequest.getProfile().getUsername() == null || profileRequest.getProfile().getUsername().equals("")) {
                throw new MissingParameterException("username should not be null", null);
            }
        } else if (method.equals("deleteUserProfile")) {

            UserProfileRequest profileRequest = (UserProfileRequest) obj;
            if (profileRequest.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileRequest.getProfile() == null) {
                throw new MissingParameterException("Profile should be valid ", null);
            }
            if (profileRequest.getProfile().getUsername() == null || profileRequest.getProfile().getUsername().equals("")) {
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


    private boolean validateCreateGroup(Object obj, String method) {
        if (obj instanceof GroupRequest) {
            GroupRequest profile = (GroupRequest) obj;

            if (profile.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profile.getGroup() == null) {
                throw new MissingParameterException("Group should be valid ", null);
            }
            if (profile.getGroup().getName() == null || profile.getGroup().getName().equals("")) {
                throw new MissingParameterException("Name should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method " + method);
        }
        return true;

    }

    private boolean validateUpdateGroup(Object obj, String method) {
        if (obj instanceof GroupRequest) {
            GroupRequest profile = (GroupRequest) obj;

            if (profile.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profile.getGroup() == null) {
                throw new MissingParameterException("Group should be valid ", null);
            }
            if (profile.getGroup().getName() == null || profile.getGroup().getName().equals("")) {
                throw new MissingParameterException("Name should not be null", null);
            }

            if (profile.getGroup().getId() == null || profile.getGroup().getId().equals("")) {
                throw new MissingParameterException("Id should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method " + method);
        }
        return true;

    }

    private boolean validateDeleteGroup(Object obj, String method) {

        if (obj instanceof GroupRequest) {
            GroupRequest profile = (GroupRequest) obj;

            if (profile.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profile.getGroup() == null) {
                throw new MissingParameterException("Group should be valid ", null);
            }
            if (profile.getGroup().getId() == null) {
                throw new MissingParameterException("Group Id be valid ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method " + method);
        }
        return true;
    }

    private boolean validateFindGroup(Object obj, String method) {
        if (obj instanceof GroupRequest) {
            GroupRequest profile = (GroupRequest) obj;

            if (profile.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profile.getGroup() == null) {
                throw new MissingParameterException("Group should be valid ", null);
            }

            if ((profile.getGroup().getName() == null || profile.getGroup().getName().equals("")) &&
                    (profile.getGroup().getId() == null || profile.getGroup().getId().equals(""))) {
                throw new MissingParameterException("Name or Id should not be null", null);
            }


        } else {
            throw new RuntimeException("Unexpected input type for method " + method);
        }
        return true;

    }

    private boolean validateGroupMembership(Object obj, String method) {
        if (obj instanceof GroupMembership) {
            GroupMembership profile = (GroupMembership) obj;

            if (profile.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }

            if (profile.getGroupId() == null || profile.getGroupId().equals("")) {
                throw new MissingParameterException("Group Id be valid ", null);
            }

            if (profile.getUsername() == null || profile.getUsername().equals("")) {


                throw new MissingParameterException("Username should not be null", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method " + method);
        }
        return true;

    }


    private boolean validateGetAllUserProfiles(Object obj, String method) {
        if (obj instanceof UserProfileRequest) {
            UserProfileRequest profileReq = (UserProfileRequest) obj;
            if (profileReq.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method" + method);
        }
        return true;

    }


    private boolean validateGroupToGroupMembership(Object obj, String method) {
        if (obj instanceof GroupToGroupMembership) {
            GroupToGroupMembership profileReq = (GroupToGroupMembership) obj;
            if (profileReq.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileReq.getChildId() == null || profileReq.getChildId().trim().equals("")) {
                throw new MissingParameterException("ChildId should not be null ", null);
            }

            if (profileReq.getParentId() == null || profileReq.getParentId().trim().equals("")) {
                throw new MissingParameterException("ParentId should not be null ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method" + method);
        }
        return true;

    }


    private boolean validateGetAllGroupsOfUser(Object obj, String method) {
        if (obj instanceof UserProfileRequest) {
            UserProfileRequest profileReq = (UserProfileRequest) obj;
            if (profileReq.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileReq.getProfile() == null || profileReq.getProfile().getUsername() == null
                    || profileReq.getProfile().getUsername().equals("")) {
                throw new MissingParameterException("Username should not be null ", null);
            }

        } else {
            throw new RuntimeException("Unexpected input type for method" + method);
        }
        return true;

    }

    private boolean validateGetAllParentGroupsOfGroup(Object obj, String method) {
        if (obj instanceof GroupRequest) {
            GroupRequest profileReq = (GroupRequest) obj;
            if (profileReq.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileReq.getGroup() == null || profileReq.getGroup().getId() == null
                    || profileReq.getGroup().getId().equals("")) {
                throw new MissingParameterException("groupId should not be null ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method" + method);
        }
        return true;
    }


    private boolean validateUserGroupMembershipTypeRequest(Object obj, String method) {
        if (obj instanceof UserGroupMembershipTypeRequest) {
            UserGroupMembershipTypeRequest profileReq = (UserGroupMembershipTypeRequest) obj;
            if (profileReq.getType() == null
                    || profileReq.getType().equals("")) {
                throw new MissingParameterException("Membership type is null ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method" + method);
        }
        return true;

    }


    private boolean validateGetUserAndChildGroups(Object obj, String method) {
        if (obj instanceof GroupRequest) {
            GroupRequest profileReq = (GroupRequest) obj;
            if (profileReq.getTenantId() == 0) {
                throw new MissingParameterException("tenantId should be valid ", null);
            }
            if (profileReq.getGroup() == null || profileReq.getGroup().getId() == null
                    || profileReq.getGroup().getId().equals("")) {
                throw new MissingParameterException("groupId should not be null ", null);
            }
        } else {
            throw new RuntimeException("Unexpected input type for method" + method);
        }
        return true;
    }


}
