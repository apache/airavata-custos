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
import org.apache.custos.core.iam.api.GetExternalIDPsRequest;
import org.apache.custos.core.iam.api.GetExternalIDPsResponse;
import org.apache.custos.core.iam.api.OperationStatus;
import org.apache.custos.core.iam.api.RegisterUserResponse;
import org.apache.custos.core.iam.api.UserSearchRequest;
import org.apache.custos.core.identity.api.AuthToken;
import org.apache.custos.core.user.management.api.SynchronizeUserDBRequest;
import org.apache.custos.core.user.management.api.UserProfileRequest;
import org.apache.custos.core.user.profile.api.UsersRolesRequest;
import org.apache.custos.core.user.profile.api.UsersRolesFullRequest;
import org.apache.custos.core.user.profile.api.GetAllUserProfilesResponse;
import org.apache.custos.core.user.profile.api.UpdateUserProfileRequest;
import org.apache.custos.core.user.profile.api.GetUpdateAuditTrailRequest;
import org.apache.custos.core.user.profile.api.GetUpdateAuditTrailResponse;
import org.apache.custos.core.user.profile.api.UserAttributeRequest;
import org.apache.custos.core.user.profile.api.UserAttributeFullRequest;
import org.apache.custos.core.user.profile.api.UserProfile;
import org.apache.custos.core.user.profile.api.Status;
import org.apache.custos.service.auth.AuthClaim;
import org.apache.custos.service.auth.TokenAuthorizer;
import org.apache.custos.service.management.UserManagementService;
import org.apache.custos.service.profile.UserProfileService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.enums.ParameterIn;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
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
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.PathVariable;
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

    private final UserProfileService userProfileService;

    public UserManagementController(UserManagementService userManagementService, TokenAuthorizer tokenAuthorizer, UserProfileService userProfileService) {
        this.userManagementService = userManagementService;
        this.tokenAuthorizer = tokenAuthorizer;
        this.userProfileService = userProfileService;
    }

    @PostMapping("/user")
    @Operation(
            summary = "Register tenant user",
            description = "This operation registers a new tenant user in the system. If the user profile already exists, the existing user profile will be returned.",

            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = RegisterUserResponse.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<UserProfile> registerUser(@RequestBody UserProfile userProfile, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();
        org.apache.custos.core.user.profile.api.UserProfileRequest request = org.apache.custos.core.user.profile.api.UserProfileRequest.newBuilder()
                .setProfile(userProfile)
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .build();

        UserProfile createdProfile = userProfileService.createUserProfile(request);

        return ResponseEntity.ok(createdProfile);
    }

    @GetMapping("/users/{userId}/federatedIDPs")
    @Operation(
            summary = "Get external identity providers of a user",
            description = "This operation retrieves associated external Identity Providers (IDPs) of identified users. " +
                    "The GetExternalIDPsRequest should include user identifiers for whom the external IDPs need to be retrieved. " +
                    "Upon successful execution, the system returns a GetExternalIDPsResponse containing the associated external IDPs for each user."
    )
    public ResponseEntity<GetExternalIDPsResponse> getExternalIDPsOfUsers(@PathVariable(value = "userId") String userId, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();

        GetExternalIDPsRequest request = GetExternalIDPsRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setUserId(userId)
                .build();

        GetExternalIDPsResponse response = userManagementService.getExternalIDPsOfUsers(request);
        return ResponseEntity.ok(response);
    }


    @PostMapping("/users/{userId}/attributes")
    @Operation(
            summary = "Add attributes to a tenant user",
            description = "This operation adds specified attributes to a tenant user. The id of each attribute passed in is ignored."
    )
    public ResponseEntity<UserProfile> addTenantUserAttributes(@PathVariable("userId") String userId, @RequestBody UserAttributeRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();

        UserAttributeFullRequest userAttributeFullRequest = UserAttributeFullRequest.newBuilder()
                .setUserAttributeRequest(request)
                .setTenantId(authClaim.getTenantId())
                .setUsername(userId)
                .build();

        UserProfile profile = userProfileService.addUserAttributes(userAttributeFullRequest);

        return ResponseEntity.ok(profile);
    }

    @DeleteMapping("/users/{userId}/attributes")
    @Operation(
            summary = "Delete attributes from a tenant user",
            description = "This operation removes specified attributes to a tenant user. The id of each attribute passed in is ignored."
    )
    public ResponseEntity<UserProfile> deleteTenantUserAttributes(@PathVariable("userId") String userId, @RequestBody UserAttributeRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();

        UserAttributeFullRequest userAttributeFullRequest = UserAttributeFullRequest.newBuilder()
                .setUserAttributeRequest(request)
                .setTenantId(authClaim.getTenantId())
                .setUsername(userId)
                .build();

        UserProfile profile = userProfileService.deleteUserAttributes(userAttributeFullRequest);

        return ResponseEntity.ok(profile);
    }

    @GetMapping("/users/{userId}/metadata-trails")
    @Operation(
            summary = "Get update metadata of a user",
            description = "Shows when any attribute of a user is updated."
    )
    public ResponseEntity<GetUpdateAuditTrailResponse> getUpdateMetadata(@PathVariable("userId") String userId, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();
        GetUpdateAuditTrailRequest request = GetUpdateAuditTrailRequest.newBuilder()
                .setUsername(userId)
                .setTenantId(authClaim.getTenantId())
                .build();

        GetUpdateAuditTrailResponse response = userManagementService.getUserProfileAuditTrails(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/roles")
    @Operation(
            summary = "Add roles to tenant users",
            description = "This operation adds specified roles (client & realm) to identified users."
    )
    public ResponseEntity<Status> addUsersRoles(@Valid @RequestBody UsersRolesRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();

        UsersRolesFullRequest fullRequest = UsersRolesFullRequest.newBuilder()
                .setUsersRoles(request)
                .setTenantId(authClaim.getTenantId())
                .build();

        Status operationStatus = userManagementService.addRolesToUsers(fullRequest);

        return ResponseEntity.ok(operationStatus);
    }

    @DeleteMapping("/roles")
    @Operation(
            summary = "Delete roles from tenant users",
            description = "This operation deletes specified roles (client & realm) to identified users."
    )
    public ResponseEntity<Status> deleteUsersRoles(@Valid @RequestBody UsersRolesRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();

        UsersRolesFullRequest fullRequest = UsersRolesFullRequest.newBuilder()
                .setUsersRoles(request)
                .setTenantId(authClaim.getTenantId())
                .build();

        Status operationStatus = userManagementService.deleteRolesFromUsers(fullRequest);

        return ResponseEntity.ok(operationStatus);
    }

    @DeleteMapping("/users/{userId}")
    @Operation(
            summary = "Delete user",
            description = "Deletes both the tenant user and the keycloak user."
    )
    public ResponseEntity<OperationStatus> deleteUser(@PathVariable(value="userId") String userId, @RequestHeader HttpHeaders headers) {
        UserSearchRequest request = UserSearchRequest.newBuilder()
                .setUser(org.apache.custos.core.iam.api.UserSearchMetadata.newBuilder().setUsername(userId))
                .build();

        OperationStatus response = userManagementService.deleteUser(generateUserSearchRequest(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response);
    }

    @PatchMapping("/users/{userId}")
    @Operation(
            summary = "Update tenant user profile",
            description = "This operation updates profiles of existing users. The UserProfile should only specify fields that should be updated. Currently, only the first and last name can be updated." +
                    "user details. Upon successful profile update, the system sends back the updated UserProfile wrapped in a ResponseEntity."
    )
    public ResponseEntity<UserProfile> updateUserProfile(@RequestBody UpdateUserProfileRequest request, @PathVariable("userId") String username, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();
        AuthToken authToken = tokenAuthorizer.getSAToken(authClaim.getIamAuthId(), authClaim.getIamAuthSecret(), authClaim.getTenantId());

        if (authToken == null || StringUtils.isBlank(authToken.getAccessToken())) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized. Service Account token is invalid");
        }

        UserProfile userProfile = UserProfile.newBuilder()
                .setUsername(username)
                .setFirstName(request.getFirstName())
                .setLastName(request.getLastName())
                .build();

        UserProfileRequest userProfileRequest = UserProfileRequest.newBuilder()
                .setUserProfile(userProfile)
                .setClientId(authClaim.getIamAuthId())
                .setClientSecret(authClaim.getIamAuthSecret())
                .setTenantId(authClaim.getTenantId())
                .setAccessToken(authToken.getAccessToken())
                .setPerformedBy(Constants.SYSTEM)
                .build();

        UserProfile response = userManagementService.updateUserProfile(userProfileRequest);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/users/{userId}")
    @Operation(
            summary = "Get tenant user profile",
            description = "This operation retrieves the profile of a specified user. Parameters should be passed in the query string."
    )
    public ResponseEntity<UserProfile> getUserProfile(
            @PathVariable("userId") String username,
            @RequestHeader HttpHeaders headers
    ) {
        // Authorization check
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);

        // Build the UserProfile using the builder pattern
        UserProfile userProfile = UserProfile.newBuilder()
                .setUsername(username)
                .build();

        // Build the UserProfileRequest using the builder pattern
        UserProfileRequest request = UserProfileRequest.newBuilder()
                .setUserProfile(userProfile)
                .setLimit(1)
                .setOffset(0)
                .build();

        if (claim.isPresent()) {
            request = request.toBuilder().setTenantId(claim.get().getTenantId()).build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        // Retrieve user profile
        UserProfile response = userManagementService.getUserProfile(request);

        return ResponseEntity.ok(response);
    }

    @GetMapping("/users")
    @Operation(
            summary = "Get all user profiles in tenant (with search & pagination)",
            description = "Can only specify one search term. If more than one is specified, priority is given to username, first name, then last name."
    )
    public ResponseEntity<GetAllUserProfilesResponse> getAllUserProfilesInTenant(
            @RequestParam(value = "offset", defaultValue = "0") int offset,
            @RequestParam(value = "limit", defaultValue = "20") int limit,
            @RequestParam(value = "first_name", defaultValue = "", required = false) String firstName,
            @RequestParam(value = "last_name", defaultValue = "", required = false) String lastName,
            @RequestParam(value = "username", defaultValue = "", required=false) String username,
            @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        UserProfile searchProfile = UserProfile.newBuilder()
                .setFirstName(firstName)
                .setLastName(lastName)
                .setUsername(username)
                .build();

        UserProfileRequest request = UserProfileRequest.newBuilder()
                .setUserProfile(searchProfile)
                .setOffset(offset)
                .setLimit(limit)
                .setTenantId(claim.get().getTenantId())
                .build();

        GetAllUserProfilesResponse response = userManagementService.getAllUserProfilesInTenant(request);

        return ResponseEntity.ok(response);
    }

    @PostMapping("/db/synchronize")
    @Operation(
            summary = "Synchronize User Databases",
            description = "Adds all missing keycloak users to the tenant user database. ."
    )
    public ResponseEntity<OperationStatus> synchronizeUserDBs(@RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();

        SynchronizeUserDBRequest request = SynchronizeUserDBRequest.newBuilder()
                .setClientId(authClaim.getIamAuthId())
                .setTenantId(authClaim.getTenantId())
                .build();
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
