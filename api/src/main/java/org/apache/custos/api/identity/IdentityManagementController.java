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

package org.apache.custos.api.identity;

import com.nimbusds.jose.jwk.JWK;
import com.nimbusds.jose.jwk.KeyUse;
import com.nimbusds.jose.jwk.RSAKey;
import org.apache.custos.core.credential.store.api.Credentials;
import org.apache.custos.core.identity.api.AuthToken;
import org.apache.custos.core.identity.api.AuthenticationRequest;
import org.apache.custos.core.identity.api.Claim;
import org.apache.custos.core.identity.api.GetOIDCConfiguration;
import org.apache.custos.core.identity.api.GetTokenRequest;
import org.apache.custos.core.identity.api.GetUserManagementSATokenRequest;
import org.apache.custos.core.identity.api.IsAuthenticatedResponse;
import org.apache.custos.core.identity.api.OIDCConfiguration;
import org.apache.custos.core.identity.api.OperationStatus;
import org.apache.custos.core.identity.api.TokenResponse;
import org.apache.custos.core.identity.api.User;
import org.apache.custos.core.identity.management.api.AuthorizationRequest;
import org.apache.custos.core.identity.management.api.AuthorizationResponse;
import org.apache.custos.core.identity.management.api.EndSessionRequest;
import org.apache.custos.core.identity.management.api.GetCredentialsRequest;
import org.apache.custos.service.auth.AuthClaim;
import org.apache.custos.service.auth.KeyLoader;
import org.apache.custos.service.auth.TokenAuthorizer;
import org.apache.custos.service.credential.store.Credential;
import org.apache.custos.service.credential.store.CredentialManager;
import org.apache.custos.service.management.IdentityManagementService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.media.ArraySchema;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.media.SchemaProperty;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.apache.commons.lang3.StringUtils;
import org.json.JSONObject;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

import java.net.URI;
import java.security.KeyPair;
import java.security.interfaces.RSAPublicKey;
import java.util.Base64;
import java.util.Collections;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/api/v1/identity-management")
@Tag(name = "Identity Management")
public class IdentityManagementController {

    private final IdentityManagementService identityManagementService;
    private final TokenAuthorizer tokenAuthorizer;
    private final KeyLoader keyLoader;

    public IdentityManagementController(IdentityManagementService identityManagementService, TokenAuthorizer tokenAuthorizer, KeyLoader keyLoader) {
        this.identityManagementService = identityManagementService;
        this.tokenAuthorizer = tokenAuthorizer;
        this.keyLoader = keyLoader;
    }

    @PostMapping("/authenticate")
    @Operation(
            summary = "User Authentication",
            description = "Authenticates the user by verifying the provided request credentials. If authenticated successfully, " +
                    "returns an AuthToken which includes authentication details and associated claims.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "username",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "The user's username"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "password",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "The user's password"
                                            )
                                    )
                            }
                    )
            ),
            responses = {
                    @io.swagger.v3.oas.annotations.responses.ApiResponse(
                            responseCode = "200",
                            description = "Successful operation",
                            content = @Content(
                                    schemaProperties = {
                                            @SchemaProperty(
                                                    name = "access_token",
                                                    schema = @Schema(
                                                            type = "string",
                                                            description = "Access Token"
                                                    )
                                            ),
                                            @SchemaProperty(
                                                    name = "claims",
                                                    array = @ArraySchema(
                                                            schema = @Schema(implementation = Claim.class),
                                                            arraySchema = @Schema(description = "List of Claims")
                                                    )
                                            )
                                    }
                            )
                    ),
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = AuthToken.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<AuthToken> authenticate(@RequestBody AuthenticationRequest request, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            request.toBuilder().setTenantId(authClaim.getTenantId())
                    .setClientId(authClaim.getIamAuthId())
                    .setClientSecret(authClaim.getIamAuthSecret());
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthToken response = identityManagementService.authenticate(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/authenticate/status")
    @Operation(
            summary = "Authentication Status Check",
            description = "Checks the authentication status based on the provided AuthToken. Returns an IsAuthenticatedResponse portraying the status.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "access_token",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Access Token"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "claims",
                                            array = @ArraySchema(
                                                    schema = @Schema(implementation = Claim.class),
                                                    arraySchema = @Schema(description = "List of Claims")
                                            )
                                    )
                            }
                    )
            ),
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = Boolean.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<Boolean> isAuthenticated(@RequestBody AuthToken request, @RequestHeader HttpHeaders headers) {
        IsAuthenticatedResponse response = identityManagementService.isAuthenticated(generateAuthTokenRequest(request.toBuilder(), headers).build());
        return ResponseEntity.ok(response.getAuthenticated());
    }

    @GetMapping("/user")
    @Operation(
            summary = "Retrieve User Information",
            description = "Retrieves User Information using the provided access token. Returns a User object containing user details.",
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = User.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<User> getUser(@Parameter(description = "Access Token used for authentication", required = true)
                                        @RequestParam("access_token") String accessToken, @RequestHeader HttpHeaders headers) {
        AuthToken.Builder builder = AuthToken.newBuilder().setAccessToken(accessToken);
        User response = identityManagementService.getUser(generateAuthTokenRequest(builder, headers).build());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/account/token")
    @Operation(
            summary = "Get User Management Service Account Access Token",
            description = "Retrieves the User Management Service Account Access Token using the provided GetUserManagementSATokenRequest. " +
                    "Returns an AuthToken for the user management service account.",
            responses = {
                    @io.swagger.v3.oas.annotations.responses.ApiResponse(
                            responseCode = "200",
                            description = "Successful operation",
                            content = @Content(
                                    schemaProperties = {
                                            @SchemaProperty(
                                                    name = "access_token",
                                                    schema = @Schema(
                                                            type = "string",
                                                            description = "Access Token"
                                                    )
                                            ),
                                            @SchemaProperty(
                                                    name = "claims",
                                                    array = @ArraySchema(
                                                            schema = @Schema(implementation = Claim.class),
                                                            arraySchema = @Schema(description = "List of Claims")
                                                    )
                                            )
                                    }
                            )
                    ),
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = AuthToken.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<AuthToken> getUserManagementServiceAccountAccessToken(
            @RequestParam(value = "client_id", required = false) String clientId,
            @RequestParam(value = "client_secret", required = false) String clientSecret,
            @RequestParam(value = "tenant_id", required = false) String tenantId,
            @RequestHeader HttpHeaders headers) {

        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);
        GetUserManagementSATokenRequest.Builder builder = GetUserManagementSATokenRequest.newBuilder();
        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            builder.setTenantId(authClaim.getTenantId())
                    .setClientId(authClaim.getIamAuthId())
                    .setClientSecret(authClaim.getIamAuthSecret());
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        AuthToken response = identityManagementService.getUserManagementServiceAccountAccessToken(builder.build());
        return ResponseEntity.ok(response);
    }

    @PostMapping("/user/logout")
    @Operation(
            summary = "End User Session",
            description = "Ends the user session based on the provided EndSessionRequest. " +
                    "Returns an OperationStatus response confirming the action.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            mediaType = "application/json",
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "refresh_token",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Refresh Token"
                                            )
                                    )
                            }
                    )
            ),
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = Boolean.class))),
                    @ApiResponse(responseCode = "400", description = "Invalid request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<Boolean> endUserSession(@RequestBody Map<String, String> requestData, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);
        String refreshToken = requestData.get("refresh_token");
        if (refreshToken == null || refreshToken.isBlank()) {
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "Missing or empty refresh_token");
        }

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            org.apache.custos.core.identity.api.EndSessionRequest endSessionRequest = org.apache.custos.core.identity.api.EndSessionRequest.newBuilder()
                    .setTenantId(authClaim.getTenantId())
                    .setClientId(authClaim.getIamAuthId())
                    .setClientSecret(authClaim.getIamAuthSecret())
                    .setRefreshToken(refreshToken)
                    .build();
            EndSessionRequest request = EndSessionRequest.newBuilder().setBody(endSessionRequest).build();
            OperationStatus response = identityManagementService.endUserSession(request);

            return ResponseEntity.ok(response.getStatus());

        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
    }

    @GetMapping("/authorize")
    @Operation(
            summary = "Authorize User",
            description = "Authorizes the user by verifying the provided AuthorizationRequest. If authorized, an AuthorizationResponse is returned.",
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = AuthorizationResponse.class))),
                    @ApiResponse(responseCode = "400", description = "Invalid request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<?> authorize(
            @RequestParam(value = "client_id") String clientId,
            @RequestParam(value = "redirect_uri") String redirectUri,
            @RequestParam(value = "scope") String scope,
            @RequestParam(value = "state") String state,
            @RequestParam(value = "response_type") String responseType,
            @RequestParam(value = "code_challenge", required = false) String codeChallenge,
            @RequestParam(value = "code_challenge_method", required = false) String codeChallengeMethod) {

        AuthorizationRequest request = AuthorizationRequest.newBuilder()
                .setClientId(clientId)
                .setRedirectUri(redirectUri)
                .setScope(scope)
                .setState(state)
                .setResponseType(responseType)
                .setCodeChallenge(StringUtils.isNotBlank(codeChallenge) ? codeChallenge : "")
                .setCodeChallengeMethod(StringUtils.isNotBlank(codeChallengeMethod) ? codeChallengeMethod : "")
                .build();
        AuthorizationResponse response = identityManagementService.authorize(request);
        return ResponseEntity.status(HttpStatus.FOUND).location(URI.create(response.getRedirectUri())).build();
    }

    @PostMapping(value = "/token", consumes = MediaType.APPLICATION_FORM_URLENCODED_VALUE)
    @Operation(
            summary = "Get Token",
            description = "Retrieves a token based on the provided request. For basic authentication, use 'user_name' and 'password'; for authorization code grant flow, use 'code'. If successful, returns a TokenResponse.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            mediaType = "application/json",
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "code",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Authorization Code",
                                                    example = "wsxdcfvgbg"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "redirect_uri",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Redirect URI",
                                                    example = "https://domain/callback"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "grant_type",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "Grant Type",
                                                    example = "authorization_code"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "username",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "User Name"
                                            )
                                    ),
                                    @SchemaProperty(
                                            name = "password",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "User's password"
                                            )
                                    )
                            }
                    )
            ),
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = TokenResponse.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "404", description = "When the associated Tenant cannot be found", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<?> token(@RequestParam Map<String, String> params, @RequestHeader HttpHeaders headers) {
        // Expects the Base64 encoded value 'clientId:clientSecret' for Authorization Header
        //Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);

        /*if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            request = request.toBuilder().setTenantId(authClaim.getTenantId())
                    .setClientId(authClaim.getIamAuthId())
                    .setClientSecret(authClaim.getIamAuthSecret())
                    .build();
        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }*/

        boolean isPKCEFlow = params.containsKey("code_verifier");
        if (!isPKCEFlow && !(headers.containsKey("authorization") || params.containsKey("client_secret"))) {
            return ResponseEntity.badRequest().body("No client id and secret or code_verifier is provided in the request");
        }

        Credential credential;
        if (isPKCEFlow) {
            // If PKCE flow, only get the client ID
            credential = new Credential(params.get("client_id"), "");
        } else {
            if (headers.containsKey("authorization") && headers.getFirst("authorization") != null) {
                String authorizationHeader = headers.get("authorization").get(0);
                if (!authorizationHeader.startsWith("Basic")) {
                    return ResponseEntity.badRequest().body("Expecting a basic auth type auth header");
                }

                credential = CredentialManager.decodeToken(authorizationHeader.substring("Basic ".length()));
            } else {
                credential = new Credential(params.get("client_id"), params.get("client_secret"));
            }

            if (credential == null) {
                return ResponseEntity.badRequest().body("Credentials were provided in incorrect format. Is should be 'Basic base64(client_id:client_secret)'");
            }
        }

        // TODO : Validate key availability in params redirect_uri, code, grant_type
        GetTokenRequest request = GetTokenRequest.newBuilder()
                .setRedirectUri(params.get("redirect_uri"))
                .setCode(params.get("code"))
                .setCodeVerifier(params.getOrDefault("code_verifier", ""))
                .setGrantType(params.get("grant_type"))
                .setClientId(credential.getId())
                .setTenantId(Long.parseLong(credential.getId().split("-")[2])) // TODO This is very risky. Find a different way to figure out tenant id
                .setClientSecret(credential.getSecret()).build();

        TokenResponse response = identityManagementService.token(request);
        System.out.println("customized token \n");
        System.out.println(response.getAccessToken());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/credentials")
    @Operation(
            summary = "Get Credentials",
            description = "Retrieves credentials based on the provided client_id. Returns a Credentials object.",
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = Credentials.class))),
                    @ApiResponse(responseCode = "400", description = "Invalid request", content = @Content()),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<Credentials> getCredentials(@RequestParam(value = "client_id") String clientId, @RequestHeader HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers, clientId);

        if (claim.isPresent()) {
            AuthClaim authClaim = claim.get();
            Credentials credentials = Credentials.newBuilder()
                    .setCustosClientId(authClaim.getCustosId())
                    .setCustosClientSecret(authClaim.getCustosSecret())
                    .setCustosClientIdIssuedAt(authClaim.getCustosIdIssuedAt())
                    .setCustosClientSecretExpiredAt(authClaim.getCustosSecretExpiredAt())
                    .setCiLogonClientId(authClaim.getCiLogonId())
                    .setCiLogonClientSecret(authClaim.getCiLogonSecret())
                    .setIamClientId(authClaim.getIamAuthId())
                    .setIamClientSecret(authClaim.getIamAuthSecret())
                    .build();
            GetCredentialsRequest request = GetCredentialsRequest.newBuilder().setCredentials(credentials).build();
            Credentials response = identityManagementService.getCredentials(request);
            return ResponseEntity.ok(response);

        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }
    }

    @GetMapping("/.well-known/openid-configuration")
    @Operation(
            summary = "Get OIDC Configuration",
            description = "Retrieves the OpenID Connect (OIDC) configuration using the provided GetOIDCConfiguration request. " +
                    "Returns an OIDCConfiguration object.",
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful operation", content = @Content(schema = @Schema(implementation = OIDCConfiguration.class))),
                    @ApiResponse(responseCode = "401", description = "Unauthorized Request", content = @Content()),
                    @ApiResponse(responseCode = "404", description = "When the associated Tenant or Credentials cannot be found", content = @Content()),
                    @ApiResponse(responseCode = "500", description = "Internal Server Error", content = @Content())
            }
    )
    public ResponseEntity<OIDCConfiguration> getOIDCConfiguration(@RequestParam(value = "client_id") String clientId) {
        GetOIDCConfiguration request = GetOIDCConfiguration.newBuilder()
                .setClientId(clientId)
                .build();
        OIDCConfiguration response = identityManagementService.getOIDCConfiguration(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/.well-known/jwks.json")
    public ResponseEntity<?> keys() {
        KeyPair keyPair = keyLoader.getKeyPair();
        RSAPublicKey rsaPublicKey = (RSAPublicKey) keyPair.getPublic();

        JWK jwk = new RSAKey.Builder(rsaPublicKey)
                .keyID(keyLoader.getKeyID())
                .keyUse(KeyUse.SIGNATURE)
                .build();

        Map<String, Object> jwkMap = Map.of(
                "kty", "RSA",
                "alg", "RS256",
                "use", "sig",
                "kid", jwk.getKeyID(),
                "n", Base64.getUrlEncoder().withoutPadding().encodeToString(rsaPublicKey.getModulus().toByteArray()),
                "e", Base64.getUrlEncoder().withoutPadding().encodeToString(rsaPublicKey.getPublicExponent().toByteArray()),
                "x5c", Collections.singletonList(Base64.getUrlEncoder().withoutPadding().encodeToString(rsaPublicKey.getEncoded()))
        );

        return ResponseEntity.ok(Map.of("keys", Collections.singletonList(jwkMap)));
    }

    @PostMapping("/token/introspect")
    @Operation(
            summary = "Introspect Token",
            description = "Validates and provides information about the provided access token.",
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(
                    required = true,
                    content = @Content(
                            mediaType = "application/x-www-form-urlencoded",
                            schemaProperties = {
                                    @SchemaProperty(
                                            name = "token",
                                            schema = @Schema(
                                                    type = "string",
                                                    description = "The access token to introspect"
                                            )
                                    )
                            }
                    )
            ),
            responses = {
                    @ApiResponse(responseCode = "200", description = "Successful introspection"),
                    @ApiResponse(responseCode = "400", description = "Invalid request"),
                    @ApiResponse(responseCode = "401", description = "Unauthorized client")
            }
    )
    public ResponseEntity<?> introspectToken(@RequestParam Map<String, String> params) {
        if (params.containsKey("client_id") && params.containsKey("client_secret") && params.containsKey("token")) {
            JSONObject response = identityManagementService.introspectToken(params.get("client_id"), params.get("client_secret"),
                    String.valueOf(params.get("client_id").split("-")[2]), params.get("token"));

            return ResponseEntity.ok(response.toMap());

        } else {
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "Missing required client_id, client_secret, or token params");
        }
    }


    private AuthToken.Builder generateAuthTokenRequest(AuthToken.Builder builder, HttpHeaders headers) {
        Optional<AuthClaim> claim = tokenAuthorizer.authorize(headers);
        if (claim.isEmpty()) {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized");
        }

        Optional<AuthClaim> opAuthClaim = tokenAuthorizer.authorizeUsingUserToken(builder.getAccessToken());

        if (opAuthClaim.isPresent()) {
            AuthClaim authClaim = claim.get();
            Claim userClaim = Claim.newBuilder().setKey("username").setValue(authClaim.getUsername()).build();
            Claim tenantClaim = Claim.newBuilder().setKey("tenantId").setValue(String.valueOf(authClaim.getTenantId())).build();
            Claim clientClaim = Claim.newBuilder().setKey("clientId").setValue(String.valueOf(authClaim.getCustosId())).build();

            builder.addClaims(userClaim);
            builder.addClaims(tenantClaim);
            builder.addClaims(clientClaim);

            return builder;

        } else {
            throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "Request is not authorized, User token not found");
        }
    }
}
