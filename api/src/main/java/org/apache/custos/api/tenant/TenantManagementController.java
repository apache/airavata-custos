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

package org.apache.custos.api.tenant;

import org.apache.custos.core.credential.store.api.CredentialMetadata;
import org.apache.custos.core.exception.UnauthorizedException;
import org.apache.custos.core.federated.authentication.api.CacheManipulationRequest;
import org.apache.custos.core.federated.authentication.api.GetInstitutionsResponse;
import org.apache.custos.core.federated.authentication.api.Status;
import org.apache.custos.core.iam.api.AddProtocolMapperRequest;
import org.apache.custos.core.iam.api.AddRolesRequest;
import org.apache.custos.core.iam.api.AllRoles;
import org.apache.custos.core.iam.api.DeleteRoleRequest;
import org.apache.custos.core.iam.api.EventPersistenceRequest;
import org.apache.custos.core.iam.api.GetRolesRequest;
import org.apache.custos.core.iam.api.OperationStatus;
import org.apache.custos.core.tenant.management.api.CreateTenantResponse;
import org.apache.custos.core.tenant.management.api.Credentials;
import org.apache.custos.core.tenant.management.api.DeleteTenantRequest;
import org.apache.custos.core.tenant.management.api.GetTenantRequest;
import org.apache.custos.core.tenant.management.api.TenantValidationRequest;
import org.apache.custos.core.tenant.management.api.UpdateTenantRequest;
import org.apache.custos.core.tenant.profile.api.GetAllTenantsForUserRequest;
import org.apache.custos.core.tenant.profile.api.GetAllTenantsForUserResponse;
import org.apache.custos.core.tenant.profile.api.GetAllTenantsResponse;
import org.apache.custos.core.tenant.profile.api.GetAttributeUpdateAuditTrailResponse;
import org.apache.custos.core.tenant.profile.api.GetAuditTrailRequest;
import org.apache.custos.core.tenant.profile.api.GetStatusUpdateAuditTrailResponse;
import org.apache.custos.core.tenant.profile.api.GetTenantsRequest;
import org.apache.custos.core.tenant.profile.api.Tenant;
import org.apache.custos.core.tenant.profile.api.UpdateStatusRequest;
import org.apache.custos.core.tenant.profile.api.UpdateStatusResponse;
import org.apache.custos.service.auth.AuthClaim;
import org.apache.custos.service.auth.TokenAuthorizer;
import org.apache.custos.service.management.TenantManagementService;
import io.swagger.v3.oas.annotations.Hidden;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import org.apache.commons.lang3.StringUtils;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import java.util.List;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/api/v1/tenant-management")
@Tag(name = "Tenant Management")
public class TenantManagementController {

    private final TenantManagementService tenantManagementService;
    private final TokenAuthorizer tokenAuthorizer;


    public TenantManagementController(TenantManagementService tenantManagementService, TokenAuthorizer tokenAuthorizer) {
        this.tenantManagementService = tenantManagementService;
        this.tokenAuthorizer = tokenAuthorizer;
    }

    @PostMapping("/oauth2/tenant")
    public ResponseEntity<CreateTenantResponse> createTenant(@Valid @RequestBody Tenant request, @RequestHeader HttpHeaders headers) {
        String token = tokenAuthorizer.getToken(headers);

        if (StringUtils.isBlank(token)) {
            Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);
            if (claim.isPresent()) {
                AuthClaim authClaim = claim.get();
                request = request.toBuilder().setParentTenantId(authClaim.getTenantId()).build();
            }
        }

        CreateTenantResponse response = tenantManagementService.createTenant(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/oauth2/tenant")
    public ResponseEntity<Tenant> getTenant(@Valid @RequestBody GetTenantRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            Credentials credentials = getCredentials(authClaim);

            request = request.toBuilder()
                    .setTenantId(authClaim.getTenantId())
                    .setCredentials(credentials)
                    .build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }

        Tenant response = tenantManagementService.getTenant(request);
        return ResponseEntity.ok(response);
    }

    @PutMapping("/user/profile")
    public ResponseEntity<Tenant> updateTenant(@Valid @RequestBody UpdateTenantRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            Credentials credentials = getCredentials(authClaim);

            request = request.toBuilder()
                    .setTenantId(authClaim.getTenantId())
                    .setCredentials(credentials)
                    .build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }

        Tenant response = tenantManagementService.updateTenant(request);
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/user")
    public ResponseEntity<?> deleteTenant(@Valid @RequestBody DeleteTenantRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            Credentials credentials = getCredentials(authClaim);

            request = request.toBuilder()
                    .setTenantId(authClaim.getTenantId())
                    .setCredentials(credentials)
                    .build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }

        tenantManagementService.deleteTenant(request);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/tenant/credentials/status")
    public ResponseEntity<OperationStatus> validateTenant(@Valid @RequestBody TenantValidationRequest request) {
        OperationStatus response = tenantManagementService.validateTenant(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/roles")
    public ResponseEntity<AllRoles> addTenantRoles(@Valid @RequestBody AddRolesRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            if (StringUtils.isBlank(request.getClientId())) {
                request = request.toBuilder()
                        .setTenantId(authClaim.getTenantId())
                        .setClientId(authClaim.getCustosId())
                        .build();
            } else {
                CredentialMetadata metadata = tokenAuthorizer.getCredentialsFromClientId(request.getClientId());
                if (authClaim.isSuperTenant() || tokenAuthorizer.validateParentChildTenantRelationShip(authClaim.getTenantId(), metadata.getOwnerId())) {
                    request = request.toBuilder()
                            .setTenantId(metadata.getOwnerId())
                            .build();
                } else {
                    throw new UnauthorizedException("Request is not authorized, user not authorized with the requested clientId: " + request.getClientId());
                }
            }
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }

        AllRoles response = tenantManagementService.addTenantRoles(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/roles")
    public ResponseEntity<AllRoles> getTenantRoles(@Valid @RequestBody GetRolesRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            if (StringUtils.isBlank(request.getClientId())) {
                request = request.toBuilder()
                        .setTenantId(authClaim.getTenantId())
                        .setClientId(authClaim.getCustosId())
                        .build();
            } else {
                CredentialMetadata metadata = tokenAuthorizer.getCredentialsFromClientId(request.getClientId());
                if (authClaim.isSuperTenant() || tokenAuthorizer.validateParentChildTenantRelationShip(authClaim.getTenantId(), metadata.getOwnerId())) {
                    request = request.toBuilder()
                            .setTenantId(metadata.getOwnerId())
                            .build();
                } else {
                    throw new UnauthorizedException("Request is not authorized, user not authorized with the requested clientId: " + request.getClientId());
                }
            }
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }

        AllRoles response = tenantManagementService.getTenantRoles(request);
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/role")
    public ResponseEntity<OperationStatus> deleteRole(@Valid @RequestBody DeleteRoleRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            if (StringUtils.isBlank(request.getClientId())) {
                request = request.toBuilder()
                        .setTenantId(authClaim.getTenantId())
                        .setClientId(authClaim.getCustosId())
                        .build();
            } else {
                CredentialMetadata metadata = tokenAuthorizer.getCredentialsFromClientId(request.getClientId());
                if (authClaim.isSuperTenant() || tokenAuthorizer.validateParentChildTenantRelationShip(authClaim.getTenantId(), metadata.getOwnerId())) {
                    request = request.toBuilder()
                            .setTenantId(metadata.getOwnerId())
                            .build();
                } else {
                    throw new UnauthorizedException("Request is not authorized, user not authorized with the requested clientId: " + request.getClientId());
                }
            }
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }

        OperationStatus response = tenantManagementService.deleteRole(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/protocol/mapper")
    public ResponseEntity<OperationStatus> addProtocolMapper(@Valid @RequestBody AddProtocolMapperRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            if (StringUtils.isBlank(request.getClientId())) {
                request = request.toBuilder()
                        .setTenantId(authClaim.getTenantId())
                        .setClientId(authClaim.getCustosId())
                        .build();
            } else {
                CredentialMetadata metadata = tokenAuthorizer.getCredentialsFromClientId(request.getClientId());
                if (authClaim.isSuperTenant() || tokenAuthorizer.validateParentChildTenantRelationShip(authClaim.getTenantId(), metadata.getOwnerId())) {
                    request = request.toBuilder()
                            .setTenantId(metadata.getOwnerId())
                            .build();
                } else {
                    throw new UnauthorizedException("Request is not authorized, user not authorized with the requested clientId: " + request.getClientId());
                }
            }
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }
        OperationStatus response = tenantManagementService.addProtocolMapper(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/events")
    public ResponseEntity<OperationStatus> configureEventPersistence(@Valid @RequestBody EventPersistenceRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            request = request.toBuilder()
                    .setTenantId(claim.get().getTenantId())
                    .setPerformedBy(claim.get().getPerformedBy())
                    .build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }
        OperationStatus response = tenantManagementService.configureEventPersistence(request);
        return ResponseEntity.ok(response);
    }


    @PostMapping("/status")
    public ResponseEntity<UpdateStatusResponse> updateTenantStatus(@Valid @RequestBody UpdateStatusRequest request) {
        UpdateStatusResponse response = tenantManagementService.updateTenantStatus(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/tenants")
    public ResponseEntity<GetAllTenantsResponse> getAllTenants(@Valid @RequestBody GetTenantsRequest request) {
        GetAllTenantsResponse response = tenantManagementService.getAllTenants(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/child/tenants")
    public ResponseEntity<GetAllTenantsResponse> getChildTenants(@Valid @RequestBody GetTenantsRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorizeUsingUserToken(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();

            if (StringUtils.isBlank(request.getParentClientId())) {
                request = request.toBuilder()
                        .setParentId(authClaim.getTenantId())
                        .build();
            } else {
                CredentialMetadata metadata = tokenAuthorizer.getCredentialsFromClientId(request.getParentClientId());
                if (authClaim.isSuperTenant() || tokenAuthorizer.validateParentChildTenantRelationShip(authClaim.getTenantId(), metadata.getOwnerId())) {
                    request = request.toBuilder()
                            .setParentId(metadata.getOwnerId())
                            .build();
                } else {
                    throw new UnauthorizedException("Request is not authorized, user not authorized with the requested clientId: " + request.getParentClientId());
                }
            }
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }
        GetAllTenantsResponse response = tenantManagementService.getChildTenants(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/tenants/{requesterEmail}")
    public ResponseEntity<GetAllTenantsForUserResponse> getAllTenantsForUser(@PathVariable("requesterEmail") String requesterEmail, @RequestHeader HttpHeaders headers) {
        tokenAuthorizer.authorize(headers);
        GetAllTenantsForUserResponse response = tenantManagementService.getAllTenantsForUser(GetAllTenantsForUserRequest.newBuilder().setRequesterEmail(requesterEmail).build());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/audit/status/{tenantId}")
    public ResponseEntity<GetStatusUpdateAuditTrailResponse> getTenantStatusUpdateAuditTrail(@PathVariable("tenantId") long tenantId) {
        GetStatusUpdateAuditTrailResponse response = tenantManagementService.getTenantStatusUpdateAuditTrail(GetAuditTrailRequest.newBuilder().setTenantId(tenantId).build());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/cache/institutions/CILogon")
    public ResponseEntity<GetInstitutionsResponse> getFromCache(@Valid @RequestBody CacheManipulationRequest request, @RequestHeader HttpHeaders headers) {
        GetInstitutionsResponse response = tenantManagementService.getFromCache(generateCacheManipulationRequest(request, headers));
        return ResponseEntity.ok(response);
    }

    @GetMapping("/institutions/CILogon")
    public ResponseEntity<GetInstitutionsResponse> getInstitutions(@Valid @RequestBody CacheManipulationRequest request, @RequestHeader HttpHeaders headers) {
        GetInstitutionsResponse response = tenantManagementService.getInstitutions(generateCacheManipulationRequest(request, headers));
        return ResponseEntity.ok(response);
    }

    @PostMapping("/cache/institutions/CILogon")
    public ResponseEntity<Status> addToCache(@Valid @RequestBody CacheManipulationRequest request, @RequestHeader HttpHeaders headers) {
        request = generateCacheManipulationRequest(request, headers);
        Optional<AuthClaim> userClaim = tokenAuthorizer.validateUserToken(headers);
        if (userClaim.isPresent()) {
            request = request.toBuilder()
                    .setPerformedBy(userClaim.get().getUsername())
                    .build();
        }
        Status response = tenantManagementService.addToCache(request);
        return ResponseEntity.ok(response);
    }

    @DeleteMapping("/cache/institutions/CILogon")
    public ResponseEntity<Status> removeFromCache(@Valid @RequestBody CacheManipulationRequest request, @RequestHeader HttpHeaders headers) {
        request = generateCacheManipulationRequest(request, headers);
        Optional<AuthClaim> userClaim = tokenAuthorizer.validateUserToken(headers);
        if (userClaim.isPresent()) {
            request = request.toBuilder()
                    .setPerformedBy(userClaim.get().getUsername())
                    .build();
        }
        Status response = tenantManagementService.removeFromCache(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/audit/attributes/{tenantId}")
    public ResponseEntity<GetAttributeUpdateAuditTrailResponse> getTenantAttributeUpdateAuditTrail(@PathVariable("tenantId") int tenantId) {
        GetAttributeUpdateAuditTrailResponse response = tenantManagementService.getTenantAttributeUpdateAuditTrail(GetAuditTrailRequest.newBuilder()
                .setTenantId(tenantId).build());
        return ResponseEntity.ok(response);
    }

    @Hidden
    @PostMapping("/oauth2/tenant/{tenantId}/client")
    public ResponseEntity<?> configureClient(@PathVariable("tenantId") int tenantId, @RequestBody Map<String, Object> body, @RequestHeader HttpHeaders headers) {
//        tokenAuthorizer.authorize(headers);
        Map<String, String> response = tenantManagementService.addClient(tenantId, (String) body.get("tenantUrl"), (List<String>) body.get("redirectUris"));
        return ResponseEntity.ok(response);
    }


    private Credentials getCredentials(AuthClaim claim) {
        return Credentials.newBuilder()
                .setCustosClientId(claim.getCustosId())
                .setCustosClientSecret(claim.getCustosSecret())
                .setCustosClientIdIssuedAt(claim.getCustosIdIssuedAt())
                .setCustosClientSecretExpiredAt(claim.getCustosSecretExpiredAt())
                .setIamClientId(claim.getIamAuthId())
                .setIamClientSecret(claim.getIamAuthSecret())
                .setCiLogonClientId(claim.getCiLogonId())
                .setCiLogonClientSecret(claim.getCiLogonSecret()).build();
    }

    private CacheManipulationRequest generateCacheManipulationRequest(CacheManipulationRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            return request.toBuilder()
                    .setTenantId(authClaim.getTenantId())
                    .build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, token not found");
        }
    }
}
