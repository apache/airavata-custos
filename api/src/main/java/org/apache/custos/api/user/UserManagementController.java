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

package org.apache.custos.api.user;

import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.iam.api.AddExternalIDPLinksRequest;
import org.apache.custos.core.iam.api.AddUserAttributesRequest;
import org.apache.custos.core.iam.api.AddUserRolesRequest;
import org.apache.custos.core.iam.api.DeleteExternalIDPsRequest;
import org.apache.custos.core.iam.api.DeleteUserAttributeRequest;
import org.apache.custos.core.iam.api.DeleteUserRolesRequest;
import org.apache.custos.core.iam.api.FindUsersRequest;
import org.apache.custos.core.iam.api.FindUsersResponse;
import org.apache.custos.core.iam.api.GetExternalIDPsRequest;
import org.apache.custos.core.iam.api.GetExternalIDPsResponse;
import org.apache.custos.core.iam.api.OperationStatus;
import org.apache.custos.core.iam.api.RegisterUserRequest;
import org.apache.custos.core.iam.api.RegisterUserResponse;
import org.apache.custos.core.iam.api.RegisterUsersRequest;
import org.apache.custos.core.iam.api.RegisterUsersResponse;
import org.apache.custos.core.iam.api.ResetUserPassword;
import org.apache.custos.core.iam.api.UserAttribute;
import org.apache.custos.core.iam.api.UserRepresentation;
import org.apache.custos.core.iam.api.UserSearchMetadata;
import org.apache.custos.core.iam.api.UserSearchRequest;
import org.apache.custos.core.identity.api.AuthToken;
import org.apache.custos.core.user.management.api.LinkUserProfileRequest;
import org.apache.custos.core.user.management.api.SynchronizeUserDBRequest;
import org.apache.custos.core.user.management.api.UserProfileRequest;
import org.apache.custos.core.user.profile.api.GetAllUserProfilesResponse;
import org.apache.custos.core.user.profile.api.GetUpdateAuditTrailRequest;
import org.apache.custos.core.user.profile.api.GetUpdateAuditTrailResponse;
import org.apache.custos.core.user.profile.api.UserProfile;
import org.apache.custos.service.auth.AuthClaim;
import org.apache.custos.service.auth.TokenAuthorizer;
import org.apache.custos.service.management.UserManagementService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.enums.ParameterIn;
import io.swagger.v3.oas.annotations.media.ArraySchema;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.media.SchemaProperty;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import org.apache.commons.lang3.StringUtils;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/api/v1/user-management")
@Tag(name = "User Management")
public class UserManagementController {

    private final UserManagementService userManagementService;
    private final TokenAuthorizer tokenAuthorizer;

    public UserManagementController(UserManagementService userManagementService, TokenAuthorizer tokenAuthorizer) {
        this.userManagementService = userManagementService;
        this.tokenAuthorizer = tokenAuthorizer;
    }

    @PostMapping("/user")
    @Operation(
            summary = "Register User",
            description = "This operation registers a new user in the system. The supplied RegisterUserRequest should " +
                    "include all the necessary user information such as username, email, and other user attributes. " +
                    "Upon successful registration, the system generates a unique identifier for the user and returns a " +
                    "RegisterUserResponse enriched with this identifier and other details. Any violation of the user " +
                    "registration policy (e.g., registering with an already existing username) will be handled according " +
                    "to the application's error handling protocol.",

            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = RegisterUserResponse.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<RegisterUserResponse> registerUser(@RequestParam(value = "client_id", required = false) String clientId, @RequestBody UserRepresentation requestData, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, clientId);

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            RegisterUserRequest request = RegisterUserRequest.newBuilder()
                    .setTenantId(authClaim.getTenantId())
                    .setClientId(authClaim.getIamAuthId())
                    .setClientSec(authClaim.getIamAuthSecret())
                    .setUser(requestData)
                    .build();
            RegisterUserResponse response = userManagementService.registerUser(request);
            return ResponseEntity.ok(response);

        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
    }

    @PostMapping("/users")
    @Operation(
            summary = "Register and Enable Multiple Users",
            description = "This operation registers and enables multiple users in the system simultaneously. " +
                    "The RegisterUsersRequest should include a list of user entities to be registered. " +
                    "Each entity should have the necessary user information such as username, email, etc. " +
                    "Upon successful registration and enablement, the system generates unique identifiers for the users and returns a " +
                    "RegisterUsersResponse with a boolean indicating all the users got registered. " +
                    "If it false then the response will contain the user representations of the failed users.",

            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "users",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = UserRepresentation.class),
                                                    arraySchema = @Schema(description = "List of Users")
                                            )
                                    )
                            }
                    )
            ),
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = RegisterUsersResponse.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<RegisterUsersResponse> registerAndEnableUsers(@RequestParam(value = "client_id") String clientId, @RequestBody RegisterUsersRequest request, @RequestHeader HttpHeaders headers) {
        headers = attachUserToken(headers, clientId);
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, clientId);

        if (claim.isPresent()) {
            Optional<String> userTokenOp = tokenAuthorizer.getUserTokenFromUserTokenHeader(headers);
            String userToken = userTokenOp.isEmpty()
                    ? tokenAuthorizer.getToken(headers)
                    : userTokenOp.get();

            AuthClaim authClaim = claim.get();

            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setTenantId(authClaim.getTenantId())
                    .setAccessToken(userToken)
                    .setPerformedBy(authClaim.getPerformedBy()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        RegisterUsersResponse response = userManagementService.registerAndEnableUsers(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/attributes")
    @Operation(
            summary = "Add User Attributes",
            description = "This operation allows adding custom attributes to an existing user. " +
                    "The AddUserAttributesRequest should specify the user and the attributes to add. " +
                    "Upon successful execution, the system adds the new attributes to the user profile and returns an " +
                    "OperationStatus representing the result.",

            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "users",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = String.class),
                                                    arraySchema = @Schema(description = "List of User Names")
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "attributes",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = UserAttribute.class),
                                                    arraySchema = @Schema(description = "List of User Attributes")
                                            )
                                    )
                            }
                    )
            ),

            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = OperationStatus.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<OperationStatus> addUserAttributes(@RequestParam(value = "client_id") String clientId, @RequestBody AddUserAttributesRequest request, @RequestHeader HttpHeaders headers) {
        headers = attachUserToken(headers, clientId);
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, clientId);

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            AuthToken authToken = tokenAuthorizer.getSAToken(authClaim.getIamAuthId(), authClaim.getIamAuthSecret(), authClaim.getTenantId());

            if (authToken == null || StringUtils.isBlank(authToken.getAccessToken())) {
                throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized. Service Account token is invalid");
            }

            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setTenantId(authClaim.getTenantId())
                    .setAccessToken(authToken.getAccessToken())
                    .setPerformedBy(authClaim.getPerformedBy()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        OperationStatus response = userManagementService.addUserAttributes(request);
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/attributes")
    @Operation(
            summary = "Delete User Attributes",
            description = "This operation allows removing specific attributes from an existing user. " +
                    "The DeleteUserAttributeRequest should specify the user and the attributes to delete. " +
                    "Upon successful execution, the system removes the specified attributes from the user profile and returns " +
                    "an OperationStatus representing the result."
    )
    public ResponseEntity<OperationStatus> deleteUserAttributes(@RequestBody DeleteUserAttributeRequest request, @RequestHeader HttpHeaders headers) {
        headers = attachUserToken(headers, request.getClientId());
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            AuthToken authToken = tokenAuthorizer.getSAToken(authClaim.getIamAuthId(), authClaim.getIamAuthSecret(), authClaim.getTenantId());

            if (authToken == null || StringUtils.isBlank(authToken.getAccessToken())) {
                throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized. Service Account token is invalid");
            }

            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setTenantId(authClaim.getTenantId())
                    .setAccessToken(authToken.getAccessToken())
                    .setPerformedBy(authClaim.getPerformedBy()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        OperationStatus response = userManagementService.deleteUserAttributes(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/user/activation")
    @Operation(
            summary = "Enable User",
            description = "This operation enables a previously disabled user. The UserSearchRequest should specify " +
                    "the criteria to identify the particular user. Upon successful execution, the system changes the user's " +
                    "status to 'enabled' and returns an updated UserRepresentation reflecting this new status."
    )
    public ResponseEntity<UserRepresentation> enableUser(@RequestBody UserSearchRequest request, @RequestHeader HttpHeaders headers) {
        UserRepresentation response = userManagementService.enableUser(generateUserSearchRequestWithoutAdditionalHeader(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response);
    }

    @PostMapping("/user/deactivation")
    @Operation(
            summary = "Disable User",
            description = "This operation disables a previously enabled user. The UserSearchRequest should specify " +
                    "the criteria to identify the particular user. Upon successful execution, the system changes " +
                    "the user's status to 'disabled' and returns an updated UserRepresentation reflecting this new status."
    )
    public ResponseEntity<UserRepresentation> disableUser(@RequestBody UserSearchRequest request, @RequestHeader HttpHeaders headers) {
        UserRepresentation response = userManagementService.disableUser(generateUserSearchRequestWithoutAdditionalHeader(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response);
    }

    @PostMapping("/user/admin")
    @Operation(
            summary = "Grant Admin Privileges",
            description = "This operation grants admin privileges to a specified existing user. " +
                    "The UserSearchRequest should specify the criteria to identify the particular user. " +
                    "After successful execution, the system updates the user's profile to include admin privileges " +
                    "and returns an OperationStatus reflecting the result of this operation."
    )
    public ResponseEntity<OperationStatus> grantAdminPrivileges(@RequestBody UserSearchRequest request, @RequestHeader HttpHeaders headers) {
        OperationStatus response = userManagementService.grantAdminPrivileges(generateUserSearchRequest(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/user/admin")
    @Operation(
            summary = "Remove Admin Privileges",
            description = "This operation removes admin privileges from a specified existing user who previously had them. " +
                    "The UserSearchRequest should specify the criteria to identify the particular user. " +
                    "Upon successful execution, the system updates the user's profile to remove admin privileges and returns " +
                    "an OperationStatus reflecting the result."
    )
    public ResponseEntity<OperationStatus> removeAdminPrivileges(@RequestBody UserSearchRequest request, @RequestHeader HttpHeaders headers) {
        OperationStatus response = userManagementService.removeAdminPrivileges(generateUserSearchRequest(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/users/federatedIDPs")
    @Operation(
            summary = "Delete External IDPs Of Users",
            description = "This operation deletes specified external Identity Providers (IDPs) from identified users. " +
                    "The DeleteExternalIDPsRequest should include user identifiers and the list of external IDPs to be removed. " +
                    "Upon successful execution, the system deletes the associated external IDPs from the user profiles " +
                    "and returns an OperationStatus reflecting the result."
    )
    public ResponseEntity<OperationStatus> deleteExternalIDPsOfUsers(@RequestBody DeleteExternalIDPsRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            request = request.toBuilder().setTenantId(authClaim.getTenantId())
                    .setClientId(authClaim.getIamAuthId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
        OperationStatus response = userManagementService.deleteExternalIDPsOfUsers(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/users/federatedIDPs")
    @Operation(
            summary = "Add External IDPs Of Users",
            description = "This operation associates specified external Identity Providers (IDPs) with identified users. " +
                    "The AddExternalIDPLinksRequest should include user identifiers and the list of external IDPs to be added. " +
                    "Upon successful execution, the system associates the specified external IDPs with the user profiles and returns " +
                    "an OperationStatus reflecting the result."
    )
    public ResponseEntity<OperationStatus> addExternalIDPsOfUsers(@RequestBody AddExternalIDPLinksRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            request = request.toBuilder().setTenantId(authClaim.getTenantId())
                    .setClientId(authClaim.getIamAuthId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        OperationStatus response = userManagementService.addExternalIDPsOfUsers(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/users/federatedIDPs")
    @Operation(
            summary = "Get External IDPs Of Users",
            description = "This operation retrieves associated external Identity Providers (IDPs) of identified users. " +
                    "The GetExternalIDPsRequest should include user identifiers for whom the external IDPs need to be retrieved. " +
                    "Upon successful execution, the system returns a GetExternalIDPsResponse containing the associated external IDPs for each user."
    )
    public ResponseEntity<GetExternalIDPsResponse> getExternalIDPsOfUsers(@RequestBody GetExternalIDPsRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            request = request.toBuilder().setTenantId(authClaim.getTenantId())
                    .setClientId(authClaim.getIamAuthId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
        GetExternalIDPsResponse response = userManagementService.getExternalIDPsOfUsers(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/users/roles")
    @Operation(
            summary = "Add Roles To Users",
            description = "This operation adds specified roles to identified users. The AddUserRolesRequest " +
                    "should include user identifiers and the list of roles to be added. Upon successful execution, " +
                    "the system associates the specified roles with the user profiles and returns an OperationStatus reflecting the result."
    )
    public ResponseEntity<OperationStatus> addRolesToUsers(@Valid @RequestBody AddUserRolesRequest request, @RequestHeader HttpHeaders headers) {
        headers = attachUserToken(headers, request.getClientId());
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            AuthToken authToken = tokenAuthorizer.getSAToken(authClaim.getIamAuthId(), authClaim.getIamAuthSecret(), authClaim.getTenantId());

            if (authToken == null || StringUtils.isBlank(authToken.getAccessToken())) {
                throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized. Service Account token is invalid");
            }

            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setTenantId(authClaim.getTenantId())
                    .setAccessToken(authToken.getAccessToken())
                    .setPerformedBy(authClaim.getPerformedBy()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        OperationStatus response = userManagementService.addRolesToUsers(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/user/activation/status")
    @Operation(
            summary = "Check If User Is Enabled",
            description = "This operation checks whether a specified user is enabled. The UserSearchRequest " +
                    "should specify the criteria to identify the particular user. Upon successful execution, " +
                    "it returns an OperationStatus that reflects whether the user is enabled."
    )
    public ResponseEntity<OperationStatus> isUserEnabled(@RequestBody UserSearchRequest request, @RequestHeader HttpHeaders headers) {
        OperationStatus response = userManagementService.isUserEnabled(generateUserSearchRequestWithoutAdditionalHeader(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/user/availability")
    @Operation(
            summary = "Check If Username Is Available",
            description = "This operation checks whether a given username is available. The UserSearchRequest " +
                    "should specify the username to check. Upon successful execution, it returns an OperationStatus " +
                    "that reflects whether the username is available for registration."
    )
    public ResponseEntity<OperationStatus> isUsernameAvailable(@RequestBody UserSearchRequest request, @RequestHeader HttpHeaders headers) {
        OperationStatus response = userManagementService.isUsernameAvailable(generateUserSearchRequestWithoutAdditionalHeader(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/user")
    @Operation(
            summary = "Retrieve User",
            description = "This operation retrieves a specified user's profile. The UserSearchRequest should specify " +
                    "the criteria to identify the particular user. It returns a UserRepresentation that includes " +
                    "detailed information about the user."
    )
    public ResponseEntity<UserRepresentation> getUser(@Valid @RequestBody UserSearchRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setClientSec(authClaim.getIamAuthSecret())
                    .setTenantId(claim.get().getTenantId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        UserRepresentation response = userManagementService.getUser(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/users")
    @Operation(
            summary = "Find Users",
            description = "This operation searches for users that match the criteria provided in the FindUsersRequest, " +
                    "which can include attributes like username, email, roles, etc. It returns a FindUsersResponse " +
                    "containing the matching users' profiles."
    )
    public ResponseEntity<FindUsersResponse> findUsers(@RequestParam(value = "client_id") String clientId,
                                                       @RequestParam(value = "offset") int offset,
                                                       @RequestParam(value = "limit") int limit,
                                                       @RequestParam("user.id") String userId,
                                                       @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, clientId);

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            UserSearchMetadata userSearchMetadata = UserSearchMetadata.newBuilder()
                    .setId(userId)
                    .build();
            FindUsersRequest request = FindUsersRequest.newBuilder().setClientId(authClaim.getIamAuthId())
                    .setClientSec(authClaim.getIamAuthSecret())
                    .setTenantId(claim.get().getTenantId())
                    .setOffset(offset)
                    .setLimit(limit)
                    .setUser(userSearchMetadata)
                    .build();

            FindUsersResponse response = userManagementService.findUsers(request);
            return ResponseEntity.ok(response);
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
    }

    @PutMapping("/user/password")
    @Operation(
            summary = "Update User Password",
            description = "This operation updates a specified user's password. An UpdatePasswordRequest should include " +
                    "user identifier and new password. Upon successful execution, it returns an OperationStatus " +
                    "reflecting the result of the password update process."
    )
    public ResponseEntity<OperationStatus> resetPassword(@RequestBody ResetUserPassword request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setClientSec(authClaim.getIamAuthSecret())
                    .setTenantId(claim.get().getTenantId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
        OperationStatus response = userManagementService.resetPassword(request);
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/user")
    @Operation(
            summary = "Delete User",
            description = "This operation deletes an existing user from the system. The UserSearchRequest should specify " +
                    "the criteria to identify the particular user. Upon successful execution, the system removes " +
                    "the user profile and returns an OperationStatus reflecting the result of the delete operation."
    )
    public ResponseEntity<OperationStatus> deleteUser(@RequestBody UserSearchRequest request, @RequestHeader HttpHeaders headers) {
        OperationStatus response = userManagementService.deleteUser(generateUserSearchRequest(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/user/roles")
    @Operation(
            summary = "Delete User Roles",
            description = "This operation removes specified roles from a identified user. The DeleteUserRolesRequest " +
                    "should include user identifier and a list of roles to be removed. Upon successful execution, the system " +
                    "disassociates the specified roles from the user profile and returns an OperationStatus reflecting the result."
    )
    public ResponseEntity<OperationStatus> deleteUserRoles(@Valid @RequestBody DeleteUserRolesRequest request, @RequestHeader HttpHeaders headers) {
        headers = attachUserToken(headers, request.getClientId());
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            AuthToken authToken = tokenAuthorizer.getSAToken(authClaim.getIamAuthId(), authClaim.getIamAuthSecret(), authClaim.getTenantId());

            if (authToken == null || StringUtils.isBlank(authToken.getAccessToken())) {
                throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized. Service Account token is invalid");
            }

            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setTenantId(authClaim.getTenantId())
                    .setAccessToken(authToken.getAccessToken())
                    .setPerformedBy(authClaim.getPerformedBy().isEmpty() ? Constants.SYSTEM : authClaim.getPerformedBy()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        OperationStatus response = userManagementService.deleteUserRoles(request);
        return ResponseEntity.ok(response);
    }

    @PutMapping("/user/profile")
    @Operation(
            summary = "Update User Profile",
            description = "This operation updates profiles of existing users. The UserProfileRequest should specify the updated " +
                    "user details. Upon successful profile update, the system sends back the updated UserProfile wrapped in a ResponseEntity."
    )
    public ResponseEntity<UserProfile> updateUserProfile(@RequestBody UserProfileRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            AuthToken authToken = tokenAuthorizer.getSAToken(authClaim.getIamAuthId(), authClaim.getIamAuthSecret(), authClaim.getTenantId());

            if (authToken == null || StringUtils.isBlank(authToken.getAccessToken())) {
                throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized. Service Account token is invalid");
            }

            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setClientSecret(authClaim.getIamAuthSecret())
                    .setTenantId(authClaim.getTenantId())
                    .setAccessToken(authToken.getAccessToken())
                    .setPerformedBy(Constants.SYSTEM).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }


        UserProfile response = userManagementService.updateUserProfile(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/user/profile")
    @Operation(
            summary = "Get User Profile",
            description = "This operation retrieves the profile of a specified user. The UserProfileRequest should specify which user's " +
                    "profile is to be retrieved. The system would return a ResponseEntity containing the UserProfile for the specified user."
    )
    public ResponseEntity<UserProfile> getUserProfile(@Valid @RequestBody UserProfileRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);

        if (claim.isPresent()) {
            request = request.toBuilder().setTenantId(claim.get().getTenantId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        UserProfile response = userManagementService.getUserProfile(request);
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/user/profile")
    @Operation(
            summary = "Delete User Profile",
            description = "This operation deletes the profile of a specified user. The UserProfileRequest should specify which user's " +
                    "profile is to be deleted. Upon successful profile deletion, the system would send back a ResponseEntity."
    )
    public ResponseEntity<UserProfile> deleteUserProfile(@RequestBody UserProfileRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            request = request.toBuilder().setClientId(authClaim.getIamAuthId())
                    .setClientSecret(authClaim.getIamAuthSecret())
                    .setTenantId(authClaim.getTenantId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        UserProfile response = userManagementService.deleteUserProfile(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/users/profile")
    @Operation(
            summary = "Get All User Profiles In Tenant",
            description = "This operation retrieves the profiles of all users in the tenant. A UserProfileRequest is used to get user profiles. " +
                    "Upon successful execution, the system sends back a ResponseEntity containing GetAllUserProfilesResponse, " +
                    "wrapping all user profiles in the tenant."
    )
    public ResponseEntity<GetAllUserProfilesResponse> getAllUserProfilesInTenant(@RequestBody UserProfileRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, request.getClientId());

        if (claim.isPresent()) {
            request = request.toBuilder().setTenantId(claim.get().getTenantId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        GetAllUserProfilesResponse response = userManagementService.getAllUserProfilesInTenant(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/user/profile/mapper")
    @Operation(
            summary = "Link User Profile",
            description = "This operation associates or links profiles of different users. The LinkUserProfileRequest should specify " +
                    "which user profiles are to be linked. Upon successful execution, the system gives back an OperationStatus wrapped in a " +
                    "ResponseEntity reflecting the result."
    )
    public ResponseEntity<OperationStatus> linkUserProfile(@RequestBody LinkUserProfileRequest request, @RequestHeader HttpHeaders headers) {
        String token = tokenAuthorizer.getToken(headers);
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            request = request.toBuilder().setIamClientId(authClaim.getIamAuthId())
                    .setIamClientSecret(authClaim.getIamAuthSecret())
                    .setTenantId(authClaim.getTenantId())
                    .setAccessToken(token)
                    .setPerformedBy(authClaim.getPerformedBy()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        OperationStatus response = userManagementService.linkUserProfile(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/user/profile/audit")
    @Operation(
            summary = "Get User Profile Audit Trails",
            description = "This operation retrieves the audit trails of updates to a user's profile. The GetUpdateAuditTrailRequest should specify " +
                    "the user whose audit trails need to be retrieved. Upon successful execution, the system returns a ResponseEntity containing " +
                    "a GetUpdateAuditTrailResponse wrapping the retrieved audit details."
    )
    public ResponseEntity<GetUpdateAuditTrailResponse> getUserProfileAuditTrails(@RequestBody GetUpdateAuditTrailRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isPresent()) {
            request = request.toBuilder().setTenantId(claim.get().getTenantId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        GetUpdateAuditTrailResponse response = userManagementService.getUserProfileAuditTrails(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/db/synchronize")
    @Operation(
            summary = "Synchronize User Databases",
            description = "This operation synchronizes user databases. The SynchronizeUserDBRequest should contain the necessary " +
                    "information to perform the synchronization. Upon successful execution, the system provides the synchronization " +
                    "status within an OperationStatus object wrapped in a ResponseEntity."
    )
    public ResponseEntity<OperationStatus> synchronizeUserDBs(@Valid @RequestBody SynchronizeUserDBRequest request) {
        OperationStatus response = userManagementService.synchronizeUserDBs(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/userinfo")
    @Operation(
            summary = "Retrieve User Info",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group membership type removal request",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            }
    )
    public ResponseEntity<Object> userInfo(@RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, headers.getFirst("client_id"));

        if (claim.isPresent()) {
            Map<String, Object> userInfo = userManagementService.getUserInfo(tokenAuthorizer.getToken(headers), claim.get().getTenantId());
            return ResponseEntity.ok(userInfo);

        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
    }

    private HttpHeaders attachUserToken(HttpHeaders headers, String clientId) {
        if (StringUtils.isBlank(clientId)) {
            String formattedUserToken = tokenAuthorizer.getToken(headers);
            headers.add(Constants.USER_TOKEN, formattedUserToken);
            return headers;
        }

        return headers;
    }

    private UserSearchRequest.Builder generateUserSearchRequest(UserSearchRequest.Builder builder, HttpHeaders headers) {
        headers = attachUserToken(headers, builder.getClientId());
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, builder.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            AuthToken authToken = tokenAuthorizer.getSAToken(authClaim.getIamAuthId(), authClaim.getIamAuthSecret(), authClaim.getTenantId());

            if (authToken == null || StringUtils.isBlank(authToken.getAccessToken())) {
                throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized. Service Account token is invalid");
            }

            builder.setClientId(authClaim.getIamAuthId())
                    .setClientSec(authClaim.getIamAuthSecret())
                    .setTenantId(authClaim.getTenantId())
                    .setAccessToken(authToken.getAccessToken())
                    .setPerformedBy(Constants.SYSTEM);

            return builder;

        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
    }

    private UserSearchRequest.Builder generateUserSearchRequestWithoutAdditionalHeader(UserSearchRequest.Builder builder, HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, builder.getClientId());

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            builder.setClientId(authClaim.getIamAuthId())
                    .setClientSec(authClaim.getIamAuthSecret())
                    .setTenantId(authClaim.getTenantId());

            return builder;

        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
    }
}
