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

package org.apache.custos.api.group;

import org.apache.custos.core.constants.Constants;
import org.apache.custos.core.user.profile.api.DefaultGroupMembershipTypes;
import org.apache.custos.core.user.profile.api.GetAllGroupsResponse;
import org.apache.custos.core.user.profile.api.GetAllUserProfilesResponse;
import org.apache.custos.core.user.profile.api.Group;
import org.apache.custos.core.user.profile.api.GroupAttribute;
import org.apache.custos.core.user.profile.api.GroupMembership;
import org.apache.custos.core.user.profile.api.GroupRequest;
import org.apache.custos.core.user.profile.api.GroupToGroupMembership;
import org.apache.custos.core.user.profile.api.Status;
import org.apache.custos.core.user.profile.api.UserGroupMembershipTypeRequest;
import org.apache.custos.core.user.profile.api.UserProfile;
import org.apache.custos.core.user.profile.api.UserProfileRequest;
import org.apache.custos.service.auth.AuthClaim;
import org.apache.custos.service.auth.TokenAuthorizer;
import org.apache.custos.service.management.GroupManagementService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.enums.ParameterIn;
import io.swagger.v3.oas.annotations.media.ArraySchema;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.media.SchemaProperty;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;
import org.springframework.web.servlet.support.ServletUriComponentsBuilder;

import java.net.URI;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/v1/group-management")
@Tag(name = "Group Management")
public class GroupManagementController {

    private final GroupManagementService groupManagementService;
    private final TokenAuthorizer tokenAuthorizer;

    public GroupManagementController(GroupManagementService groupManagementService, TokenAuthorizer tokenAuthorizer) {
        this.groupManagementService = groupManagementService;
        this.tokenAuthorizer = tokenAuthorizer;
    }

    @PostMapping("/groups")
    @Operation(
            summary = "Group Creation",
            description = "Creates a new group. If a group ID is provided and it already exists, an error will be thrown. " +
                    "If no ID is provided, the group name will be used as the ID along with a random string.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "id",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "The Group ID"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "name",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Group Name"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "description",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Parent Group ID"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "owner_id",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "User ID of the group owner"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "parent_id",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Parent Group ID"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "attributes",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = GroupAttribute.class),
                                                    arraySchema = @Schema(description = "List of Group Attributes")
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "realm_roles",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = String.class),
                                                    arraySchema = @Schema(description = "List of Realm Roles")
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "client_roles",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = String.class),
                                                    arraySchema = @Schema(description = "List of Client Roles")
                                            )
                                    )
                            }
                    )
            ),
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group creation request",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "201", description = "Successful operation", content = @Content(schema = @Schema(implementation = Group.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "400", description = "Bad Request, If the 'parent_id' doesn't exist", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<Group> createGroup(@RequestBody Group request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, headers.getFirst("client_id"));
        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthClaim authClaim = claim.get();
        GroupRequest groupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .setPerformedBy(authClaim.getPerformedBy() != null ? authClaim.getPerformedBy() : Constants.SYSTEM)
                .setGroup(request)
                .build();

        Group group = groupManagementService.createGroup(groupRequest);

        URI location = ServletUriComponentsBuilder.fromCurrentRequest()
                .path("/{id}")
                .buildAndExpand(group.getId())
                .encode()
                .toUri();

        return ResponseEntity.created(location).body(group);
    }

    @PutMapping("/groups/{groupId}")
    @Operation(
            summary = "Update Group",
            description = "Updates an existing group with the given ID",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "name",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Group Name"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "description",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Group Description"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "attributes",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = GroupAttribute.class),
                                                    arraySchema = @Schema(description = "List of Group Attributes")
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "realm_roles",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = String.class),
                                                    arraySchema = @Schema(description = "List of Realm Roles")
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "client_roles",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = String.class),
                                                    arraySchema = @Schema(description = "List of Client Roles")
                                            )
                                    )
                            }
                    )
            ),
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group update request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the group to update",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "200", description = "Group updated successfully", content = @Content(schema = @Schema(implementation = Group.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "404", description = "When the associated Group cannot be found", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<Group> updateGroup(@PathVariable("groupId") String groupId, @RequestBody Group request, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupRequest groupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .setPerformedBy(authClaim.getPerformedBy() != null ? authClaim.getPerformedBy() : Constants.SYSTEM)
                .setGroup(request.toBuilder().setId(groupId).build())
                .build();

        Group updatedGroup = groupManagementService.updateGroup(groupRequest);
        return ResponseEntity.ok(updatedGroup);
    }

    @PatchMapping("/groups/{groupId}")
    @Operation(
            summary = "Patch Group",
            description = "Patches group with given ID. To remove all realm_roles or client_roles, pass in a list with \"custos-remove-all\" as the first list element. To remove all group attributes, pass in a list with \"custos-remove-all\" as the key of the first list element.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "name",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Group Name"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "description",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Group Description"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "attributes",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = GroupAttribute.class),
                                                    arraySchema = @Schema(description = "List of Group Attributes")
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "realm_roles",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = String.class),
                                                    arraySchema = @Schema(description = "List of Realm Roles")
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "client_roles",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = String.class),
                                                    arraySchema = @Schema(description = "List of Client Roles")
                                            )
                                    )
                            }
                    )
            ),
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group update request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the group to update",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "200", description = "Group updated successfully", content = @Content(schema = @Schema(implementation = Group.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "404", description = "When the associated Group cannot be found", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<Group> patchGroup(@PathVariable("groupId") String groupId, @RequestBody Group request, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);

        GroupRequest exGroupReq = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .setPerformedBy(authClaim.getPerformedBy() != null ? authClaim.getPerformedBy() : Constants.SYSTEM)
                .setGroup(request.toBuilder().setId(groupId).build())
                .build();

        Group exGroup = groupManagementService.findGroup(exGroupReq);

        Group.Builder mergedGroupBuilder = exGroup.toBuilder()
                .setName(!request.getName().isEmpty() ? request.getName() : exGroup.getName())
                .setDescription(!request.getDescription().isEmpty() ? request.getDescription() : exGroup.getDescription());

        String REMOVE_ALL_TAG = "custos-remove-all";

        List<String> clientRolesLst = request.getClientRolesList();
         if (!clientRolesLst.isEmpty()) {
             mergedGroupBuilder.clearClientRoles();
             if (!clientRolesLst.get(0).equals(REMOVE_ALL_TAG)) {
                 mergedGroupBuilder.addAllClientRoles(request.getClientRolesList());
             }
        }

        List<String> realmRolesLst = request.getRealmRolesList();
        if (!realmRolesLst.isEmpty()) {
            mergedGroupBuilder.clearRealmRoles();
            if (!realmRolesLst.get(0).equals(REMOVE_ALL_TAG)) {
                mergedGroupBuilder.addAllRealmRoles(request.getRealmRolesList());
            }
        }

        List<GroupAttribute> groupAttributeList = request.getAttributesList();
        if (!groupAttributeList.isEmpty()) {
            mergedGroupBuilder.clearAttributes();
            if (!groupAttributeList.get(0).getKey().equals(REMOVE_ALL_TAG)) {
                mergedGroupBuilder.addAllAttributes(request.getAttributesList());
            }
        }

        Group mergedGroup = mergedGroupBuilder.build();

        GroupRequest updateGroupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .setPerformedBy(authClaim.getPerformedBy() != null ? authClaim.getPerformedBy() : Constants.SYSTEM)
                .setGroup(mergedGroup.toBuilder().setId(groupId).build())
                .build();

        Group updatedGroup = groupManagementService.updateGroup(updateGroupRequest);

        return ResponseEntity.ok(updatedGroup);
    }

    @DeleteMapping("/groups/{groupId}")
    @Operation(
            summary = "Delete Group",
            description = "Deletes an existing group with the given ID",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group deletion request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the group to delete",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "204", description = "Group deleted successfully"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "404", description = "When the associated Group cannot be found", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<?> deleteGroup(@PathVariable("groupId") String groupId, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupRequest groupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .setPerformedBy(authClaim.getPerformedBy() != null ? authClaim.getPerformedBy() : Constants.SYSTEM)
                .setGroup(Group.newBuilder().setId(groupId).build())
                .build();

        groupManagementService.deleteGroup(groupRequest);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/groups/{groupId}")
    @Operation(
            summary = "Get Group by ID",
            description = "Retrieves a group with the given ID",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group retrieval request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the group to retrieve",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "200", description = "Group found successfully", content = @Content(schema = @Schema(implementation = Group.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "404", description = "When the Group cannot be found", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<Group> findGroup(@PathVariable("groupId") String groupId, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupRequest groupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .setPerformedBy(authClaim.getPerformedBy() != null ? authClaim.getPerformedBy() : Constants.SYSTEM)
                .setGroup(Group.newBuilder().setId(groupId).build())
                .build();

        Group group = groupManagementService.findGroup(groupRequest);
        return ResponseEntity.ok(group);
    }

    @GetMapping("/groups")
    @Operation(
            summary = "Get All Groups (with Search)",
            description = "Retrieves a list of groups based on search criteria. Supports pagination.",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group retrieval request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "name",
                            in = ParameterIn.QUERY,
                            description = "Filter by group name (partial match)",
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "description",
                            in = ParameterIn.QUERY,
                            description = "Filter by group description (partial match)",
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "id",
                            in = ParameterIn.QUERY,
                            description = "Filter by group ID",
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "created_time",
                            in = ParameterIn.QUERY,
                            description = "Filter by groups created on or after this timestamp (milliseconds since epoch)",
                            schema = @Schema(type = "integer", format = "int64")
                    ),
                    @Parameter(
                            name = "last_modified_time",
                            in = ParameterIn.QUERY,
                            description = "Filter by groups modified on or after this timestamp (milliseconds since epoch)",
                            schema = @Schema(type = "integer", format = "int64")
                    ),
                    @Parameter(
                            name = "offset",
                            in = ParameterIn.QUERY,
                            description = "Pagination offset (start index)",
                            schema = @Schema(type = "integer", defaultValue = "0")
                    ),
                    @Parameter(
                            name = "limit",
                            in = ParameterIn.QUERY,
                            description = "Maximum number of groups to return",
                            schema = @Schema(type = "integer", defaultValue = "20")
                    )
            },
            responses = {
                    @ApiResponse(
                            responseCode = "200",
                            description = "Groups retrieved successfully",
                            content = @Content(schema = @Schema(implementation = GetAllGroupsResponse.class))
                    ),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<GetAllGroupsResponse> searchGroups(
            @RequestParam(name = "name", defaultValue = "", required = false) String name,
            @RequestParam(name = "description", defaultValue = "", required = false) String description,
            @RequestParam(name = "id", defaultValue = "", required = false) String id,
            @RequestParam(name = "created_time", defaultValue = "0", required = false) Long createdTime,
            @RequestParam(name = "last_modified_time", defaultValue = "0", required = false) Long lastModifiedTime,
            @RequestParam(name = "offset", defaultValue = "0") int offset,
            @RequestParam(name = "limit", defaultValue = "20") int limit,
            @RequestHeader HttpHeaders headers) {

        AuthClaim authClaim = authorize(headers);
        Group group = Group.newBuilder()
                .setId(id)
                .setName(name)
                .setDescription(description)
                .setCreatedTime(createdTime)
                .setLastModifiedTime(lastModifiedTime)
                .build();

        GroupRequest groupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .setPerformedBy(authClaim.getPerformedBy() != null ? authClaim.getPerformedBy() : Constants.SYSTEM)
                .setGroup(group)
                .setOffset(offset)
                .setLimit(limit)
                .build();

        GetAllGroupsResponse groups = groupManagementService.getAllGroups(groupRequest);
        return ResponseEntity.ok(groups);
    }

    @DeleteMapping("/groups/{groupId}/members/{userId}")
    @Operation(
            summary = "Remove User from Group",
            description = "Removes a user from an existing group. If the user is the group owner, an error will be thrown.",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the user removal request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the group where to remove the user",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "userId",
                            in = ParameterIn.PATH,
                            description = "ID of the User to be removed from the group",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "204", description = "User successfully removed from group"),
                    @ApiResponse(responseCode = "400", description = "Bad Request - Cannot remove the default owner"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Group or user not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<?> removeUserFromGroup(@PathVariable("groupId") String groupId, @PathVariable("userId") String userId, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupMembership request = GroupMembership.newBuilder()
                .setGroupId(groupId)
                .setUsername(userId)
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .build();
        Status status = groupManagementService.removeUserFromGroup(request);
        return status.getStatus() ? ResponseEntity.noContent().build() : ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("USER_REMOVAL_FAILED");

    }

    @PostMapping("/groups/{groupId}/members")
    @Operation(
            summary = "Add User to Group",
            description = "Adds an existing user to an existing group. If the user is already a member of the group, the request is considered successful.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    description = "Group membership details",
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "username",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Username"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "type",
                                            schema = @Schema(
                                                    implementation = DefaultGroupMembershipTypes.class,
                                                    description = "Type of the user"
                                            )
                                    )
                            }
                    )
            ),
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the user addition request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the group to add the user",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "200", description = "User successfully added to the group (or was already a member)"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Group or user not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<Status> addUserToGroup(@PathVariable("groupId") String groupId, @RequestBody GroupMembership request, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        request = request.toBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .setGroupId(groupId)
                .build();
        org.apache.custos.core.user.profile.api.Status status = groupManagementService.addUserToGroup(request);
        return ResponseEntity.ok(status);
    }

    @PostMapping("/groups/{parentId}/children")
    @Operation(
            summary = "Add Child Group to Parent Group",
            description = "Establishes a parent-child relationship between two existing groups. If the relationship already exists, the request is considered successful.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    description = "Group-to-group membership details",
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "child_id",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Child Group ID"
                                            )
                                    )
                            }
                    )
            ),
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group relationship creation request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "parentId",
                            in = ParameterIn.PATH,
                            description = "The ID of the parent group to which the child group will be added",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "200", description = "Child group successfully added to parent group (or relationship already existed)"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Child or parent group not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<Status> addChildGroupToParentGroup(@PathVariable("parentId") String parentGroupId, @RequestBody GroupToGroupMembership request, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        request = request.toBuilder()
                .setParentId(parentGroupId)
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .build();

        return ResponseEntity.ok(groupManagementService.addChildGroupToParentGroup(request));
    }

    @DeleteMapping("/groups/{parentId}/children/{childId}")
    @Operation(
            summary = "Remove Child Group from Parent Group",
            description = "Removes the parent-child relationship between two groups. If the relationship does not exist, the request is still considered successful.",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group relationship removal request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "parentId",
                            in = ParameterIn.PATH,
                            description = "The ID of the parent group",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "childId",
                            in = ParameterIn.PATH,
                            description = "The ID of the child group",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "204", description = "Child group successfully removed from parent group (or relationship did not exist)"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Child or parent group not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<?> removeChildGroupFromParentGroup(@PathVariable("parentId") String parentId, @PathVariable("childId") String childId, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupToGroupMembership groupRequest = GroupToGroupMembership.newBuilder()
                .setParentId(parentId)
                .setChildId(childId)
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .build();

        Status status = groupManagementService.removeChildGroupFromParentGroup(groupRequest);
        return status.getStatus() ? ResponseEntity.noContent().build() : ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("CHILD_GROUP_REMOVAL_FAILED");

    }

    @GetMapping("/users/{userId}/group-memberships")
    @Operation(
            summary = "Get All Groups of a User",
            description = "Retrieves all groups that a user is a member of",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group retrieval request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "userId",
                            in = ParameterIn.PATH,
                            description = "User ID",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(
                            responseCode = "200",
                            description = "Groups retrieved successfully",
                            content = @Content(schema = @Schema(implementation = GetAllGroupsResponse.class))
                    ),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "User not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<GetAllGroupsResponse> getAllGroupsOfUser(@PathVariable("userId") String username, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        UserProfileRequest request = UserProfileRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setProfile(UserProfile.newBuilder().setUsername(username).build())
                .build();
        GetAllGroupsResponse response = groupManagementService.getAllGroupsOfUser(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/groups/{groupId}/parents")
    @Operation(
            summary = "Get All Parent Groups of a Group",
            description = "Retrieves all parent groups of the specified group.",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group retrieval request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the group for which to retrieve parent groups",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(
                            responseCode = "200",
                            description = "Parent groups retrieved successfully",
                            content = @Content(schema = @Schema(implementation = GetAllGroupsResponse.class))
                    ),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Group not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<GetAllGroupsResponse> getAllParentGroupsOfGroup(@PathVariable("groupId") String groupId, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupRequest groupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setGroup(Group.newBuilder().setId(groupId).build())
                .build();

        GetAllGroupsResponse response = groupManagementService.getAllParentGroupsOfGroup(groupRequest);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/groups/{groupId}/members")
    @Operation(
            summary = "Get All Users of a Group",
            description = "Retrieves all users who are members of the specified group",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the user retrieval request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the group for which to retrieve child users",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(
                            responseCode = "200",
                            description = "Child users retrieved successfully",
                            content = @Content(schema = @Schema(implementation = GetAllUserProfilesResponse.class))
                    ),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Group not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<GetAllUserProfilesResponse> getAllChildUsers(@PathVariable("groupId") String groupId, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupRequest groupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setGroup(Group.newBuilder().setId(groupId).build())
                .build();

        GetAllUserProfilesResponse response = groupManagementService.getAllChildUsers(groupRequest);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/groups/{groupId}/children")
    @Operation(
            summary = "Get All Child Groups of a Group",
            description = "Retrieves all child groups of the specified group",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group retrieval request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "The ID of the parent group for which to retrieve child groups",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(
                            responseCode = "200",
                            description = "Child groups retrieved successfully",
                            content = @Content(schema = @Schema(implementation = GetAllGroupsResponse.class))
                    ),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Group not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<GetAllGroupsResponse> getAllChildGroups(@PathVariable("groupId") String groupId, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupRequest groupRequest = GroupRequest.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setGroup(Group.newBuilder().setId(groupId).build())
                .build();

        GetAllGroupsResponse response = groupManagementService.getAllChildGroups(groupRequest);
        return ResponseEntity.ok(response);
    }

    @PutMapping("/groups/{groupId}/members/{userId}")
    @Operation(
            summary = "Change User Membership Type",
            description = "Changes the membership type of a user within a group. Only one owner can exist within a group, " +
                    "and the default owner cannot have their membership type changed.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    description = "Group membership details for updating the membership type",
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "group_id",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "The ID of the group to which the user should be added"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "username",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Username"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "type",
                                            schema = @Schema(
                                                    implementation = DefaultGroupMembershipTypes.class,
                                                    description = "Type of the user"
                                            )
                                    )
                            }
                    )
            ),
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the membership type change request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "ID of the group where the user membership type should be changed",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "userId",
                            in = ParameterIn.PATH,
                            description = "User ID",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "200", description = "Membership type changed successfully"),
                    @ApiResponse(responseCode = "400", description = "Bad Request - Cannot change default owner's type or invalid type"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Group membership not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<?> changeUserMembershipType(@PathVariable("groupId") String groupId, @PathVariable("userId") String userId, @RequestBody GroupMembership request, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        request = request.toBuilder()
                .setGroupId(groupId)
                .setUsername(userId)
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .setClientSec(authClaim.getIamAuthSecret())
                .build();
        Status status = groupManagementService.changeUserMembershipType(request);
        return status.getStatus() ? ResponseEntity.ok(status) : ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("USER_MEMBERSHIP_TYPE_CHANGED_FAILED");
    }

    @GetMapping("/groups/{groupId}/members/{userId}/access")
    @Operation(
            summary = "Check User Group Access",
            description = "Checks if a user has a specific type of access within a group",
            parameters = {
                    @Parameter(
                            name = "userId",
                            in = ParameterIn.PATH,
                            description = "Username of the user",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "groupId",
                            in = ParameterIn.PATH,
                            description = "Group ID",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the access check request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "type",
                            in = ParameterIn.QUERY,
                            description = "Type of membership (e.g., MEMBER, OWNER, or ADMIN)",
                            required = true,
                            schema = @Schema(implementation = DefaultGroupMembershipTypes.class)
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "200", description = "Access check result", content = @Content(schema = @Schema(implementation = Status.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<Status> hasAccess(@PathVariable("groupId") String groupId, @PathVariable("userId") String username, @RequestParam String type, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        GroupMembership request = GroupMembership.newBuilder()
                .setTenantId(authClaim.getTenantId())
                .setUsername(username)
                .setGroupId(groupId)
                .setType(type)
                .build();
        Status status = groupManagementService.hasAccess(request);
        return ResponseEntity.ok(status);
    }

    @PostMapping("/group-membership-types")
    @Operation(
            summary = "Add Group Membership Type",
            description = "Adds a new group membership type. If the type already exists, the request is considered successful.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    description = "Group membership type",
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "type",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Type of the user"
                                            )
                                    )
                            }
                    )
            ),
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group membership type creation request",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "200", description = "Group membership type successfully added (or already existed)"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<Status> addGroupMembershipType(@RequestBody UserGroupMembershipTypeRequest request, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        request = request.toBuilder()
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .build();
        Status status = groupManagementService.addGroupMembershipType(request);
        return ResponseEntity.ok(status);
    }

    @DeleteMapping("/group-membership-types/{type}")
    @Operation(
            summary = "Remove Group Membership Type",
            description = "Removes an existing group membership type. If the type does not exist or if it is used by existing group memberships, the request will fail.",
            parameters = {
                    @Parameter(
                            name = "client_id",
                            in = ParameterIn.HEADER,
                            description = "The client ID initiating the group membership type removal request",
                            required = true,
                            schema = @Schema(type = "string")
                    ),
                    @Parameter(
                            name = "type",
                            in = ParameterIn.PATH,
                            description = "The type of group membership to remove",
                            required = true,
                            schema = @Schema(type = "string")
                    )
            },
            responses = {
                    @ApiResponse(responseCode = "204", description = "Group membership type successfully removed"),
                    @ApiResponse(responseCode = "400", description = "Bad Request - Cannot remove default membership types (MEMBER, OWNER, ADMIN) or if the type is in use"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request"),
                    @ApiResponse(responseCode = "404", description = "Group membership type not found"),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error")
            }
    )
    public ResponseEntity<?> removeUserGroupMembershipType(@PathVariable("type") String type, @RequestHeader HttpHeaders headers) {
        AuthClaim authClaim = authorize(headers);
        UserGroupMembershipTypeRequest request = UserGroupMembershipTypeRequest.newBuilder()
                .setType(type)
                .setTenantId(authClaim.getTenantId())
                .setClientId(authClaim.getIamAuthId())
                .build();
        Status status = groupManagementService.removeUserGroupMembershipType(request);
        return status.getStatus() ? ResponseEntity.noContent().build() : ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body("GROUP_MEMBERSHIP_REMOVAL_FAILED");
    }

    private AuthClaim authorize(@RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, headers.getFirst("client_id"));
        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        return claim.get();
    }

}
