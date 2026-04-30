# Identity & Authentication Platform Notes

**Contributor:** Temitope Aderibigbe  
**Georgia Tech VIP Program | Spring 2026**  

---

## Overview

Research notes on six identity and authentication platforms relevant to the Custos security stack. For each platform, this document covers core features, architectural design, and how user-facing authorization is handled — specifically where enforcement occurs (application-side token claim checking vs. dedicated policy engine).

---

## 1. Auth0

**Type:** Identity-as-a-Service (IDaaS) — Cloud-Native Authentication Platform

### Overview
Auth0 is a cloud-based Identity-as-a-Service platform developed by Okta. It provides developers with drop-in authentication and authorization solutions, abstracting the complexity of identity management behind simple APIs and SDKs. It is designed for ease of integration across web, mobile, and API-based applications.

### Core Features & Function
- Universal Login: Centralized, customizable login page hosted by Auth0
- Social & Enterprise SSO: Pre-built connections to 30+ identity providers (Google, GitHub, SAML, LDAP, AD)
- Multi-Factor Authentication (MFA): Built-in support for TOTP, SMS, push, and WebAuthn
- Role-Based Access Control (RBAC): Fine-grained permissions management via roles and scopes
- Machine-to-Machine (M2M) Auth: Client Credentials flow for API-to-API communication
- Anomaly Detection: Automatic blocking of suspicious login attempts and credential stuffing
- Rules & Actions: JavaScript hooks injected into the authentication pipeline for custom logic
- Auth0 Management API: Programmatic management of users, connections, and applications

### How It Works (Architecture)
- Auth0 operates as a central Authorization Server following OAuth 2.0 and OpenID Connect (OIDC) standards
- Tenant model: Each organization gets an isolated tenant — a dedicated Auth0 environment with its own user store, settings, and connections
- Authentication flow: Users are redirected to Auth0's Universal Login page → Auth0 verifies credentials → issues JWT access tokens and ID tokens → tokens are returned to the client app
- Extensibility pipeline: Actions are Node.js functions that execute at specific points in the token issuance flow, allowing injection of custom claims or external API calls
- Token format: Issues JWT-based access tokens signed with RS256 or HS256. JWKS endpoints allow downstream services to validate tokens
- Deployment: Primarily SaaS with a Private Cloud option for enterprises needing data residency

### Authorization Engine (User-Facing)
- Authorization is built around roles and permissions embedded into the JWT access token at login time
- Enforcement is entirely application-side — Auth0 does not sit in the request path after token issuance
- Actions allow custom authorization behavior at token issuance time (e.g., blocking a token if a user lacks a required attribute)
- No dedicated runtime authorization engine — all authorization context is baked into the token at login

---

## 2. WorkOS

**Type:** Enterprise SSO & Directory Sync Infrastructure for B2B SaaS

### Overview
WorkOS is a developer-focused platform providing enterprise-grade authentication features as building blocks — primarily SAML SSO, SCIM Directory Sync, and Magic Auth. Purpose-built for B2B SaaS companies that need to integrate with enterprise customers' existing identity infrastructure without managing those integrations themselves.

### Core Features & Function
- Single Sign-On (SSO): SAML 2.0 and OIDC connections to enterprise IdPs like Okta, Azure AD, Ping
- Directory Sync (SCIM): Automated user provisioning/deprovisioning from enterprise directories
- Magic Auth: Passwordless login via one-time email links
- Admin Portal: Self-service configuration UI embeddable in your app for enterprise customers
- Fine-Grained Authorization (FGA): ReBAC engine for complex permission modeling inspired by Google Zanzibar
- Audit Logs: Structured event log stream for compliance and monitoring

### How It Works (Architecture)
- WorkOS acts as an intermediary broker between your application and enterprise identity providers
- SSO flow: App redirects to WorkOS → WorkOS initiates SAML/OIDC flow with customer IdP → normalized profile returned to app → backend exchanges code for user profile via WorkOS API
- Directory Sync: Receives SCIM 2.0 push events from enterprise directories and exposes a unified events API
- FGA: Permissions modeled as relationships between objects (e.g., user → viewer → document). Engine evaluates relationship graphs to answer authorization queries
- Stateless architecture: Issues short-lived tokens/codes and leaves session management to the application

### Authorization Engine (User-Facing)
- Two distinct layers: basic role system (application-side) and Fine-Grained Authorization (engine-side)
- FGA is a dedicated server-side Policy Decision Point — the application sends a check request and the engine returns allow/deny based on a relationship graph
- One of the few platforms reviewed with a true runtime authorization service operating independently of token claims
- Directory group memberships from enterprise IdPs feed into FGA relationship definitions

---

## 3. WSO2 Asgardeo

**Type:** Cloud-Native CIAM (Customer Identity & Access Management) by WSO2

### Overview
Asgardeo is a SaaS CIAM platform built by WSO2, the same company behind WSO2 Identity Server. Designed for organizations building consumer-facing or B2B applications that need identity features without hosting an identity server themselves.

### Core Features & Function
- Standard Protocol Support: OAuth 2.0, OIDC, SAML 2.0, WS-Federation
- Social Login: Google, GitHub, Facebook, Microsoft, Apple and more via federation
- Multi-Factor Authentication: TOTP, SMS OTP, email OTP, FIDO2/WebAuthn (passkeys)
- Adaptive Authentication: JavaScript-based policy engine to dynamically step up authentication based on context
- SCIM 2.0 User Management: Standard API for provisioning and managing users programmatically
- B2B Organization Management: Hierarchical org model for managing sub-tenants with delegated admin

### How It Works (Architecture)
- Multi-tenant SaaS deployment of WSO2 Identity Server — each customer organization gets an isolated root organization
- Visual, flow-based authentication designer where admins configure login steps graphically
- Adaptive authentication: JavaScript sandbox executes policies per-login evaluating IP, device, or risk score
- Token issuance: Follows OIDC/OAuth 2.0 flows, issuing JWTs signed with per-organization keys
- B2B model: Root organizations can create child organizations with their own IdP configs and user pools

### Authorization Engine (User-Facing)
- Authorization handled through application roles and user groups embedded into JWT claims at token issuance
- No standalone runtime authorization engine — enforcement is fully application-side
- Supports OAuth 2.0 scopes for API-level access control validated by backend services or API gateways
- Adaptive authentication scripts run at login time only — not per-request resource-level authorization

---

## 4. Amazon Cognito

**Type:** AWS-Native Identity Service for Web and Mobile Applications

### Overview
Amazon Cognito is AWS's managed identity service providing two core constructs: User Pools (identity provider and user directory) and Identity Pools (for granting AWS resource access). Integrates deeply with the AWS ecosystem.

### Core Features & Function
- User Pools: Managed user directory with sign-up, sign-in, password policies, and MFA
- Identity Pools: Exchange tokens from any IdP for temporary AWS IAM credentials
- Social & SAML/OIDC Federation: Integrate external identity providers into User Pools
- Lambda Triggers: Hook into authentication lifecycle events via AWS Lambda
- Groups & RBAC: User groups with IAM role mappings for coarse-grained access control

### How It Works (Architecture)
- User Pools: OIDC-compliant identity provider issuing JWTs (ID token, access token, refresh token)
- Identity Pools: Accept tokens from any trusted provider and exchange them for temporary AWS IAM credentials via AWS STS
- Lambda Triggers: Called at lifecycle points (pre-signup, post-confirmation, pre/post authentication, token customization) as the primary extensibility mechanism
- AWS-native integration: Integrates with API Gateway, ALB, AppSync, and IAM

### Authorization Engine (User-Facing)
- Authorization handled through User Pool Groups — group names embedded in the ID token under the `cognito:groups` claim
- No dedicated authorization engine — enforcement is application-side or via AWS service integrations (API Gateway Cognito Authorizers, IAM policies)
- Identity Pools translate group memberships into temporary IAM credentials, making AWS IAM the enforcement layer for resource-level access
- Lambda Triggers are the primary escape hatch for custom authorization logic

---

## 5. Eggshell Authentication

**Type:** Lightweight, Open-Source Authentication Framework

### Overview
Eggshell Authentication is a lightweight, open-source authentication framework designed for simplicity and developer control. A self-hosted, code-first library for developers who want authentication primitives without the overhead of a full identity server.

### Core Features & Function
- Session & Token Management: JWT issuance and refresh token rotation
- Password Hashing: Secure credential storage using bcrypt/argon2
- OAuth 2.0 Client Support: Acts as an OAuth client for social login flows
- Middleware Integrations: Plugs into common web frameworks as authentication middleware
- Minimal Dependencies: Small footprint with explicit dependency management for auditability
- Customizable Storage Adapters: Supports custom backends (SQL, NoSQL)

### How It Works (Architecture)
- Library-first architecture — runs embedded within the application process, not as a separate server
- Provides discrete modules (token issuance, session management, password hashing, OAuth client) that developers compose into their own authentication flow
- Token flow: Issues JWTs locally within the app. App controls signing keys, token lifetime, and claim structure directly
- Adapter pattern for persistence — developers plug in their own database adapter

### Authorization Engine (User-Facing)
- No built-in authorization engine — purely an authentication library
- Developer decides what claims to include (roles, groups, permissions) and how the application enforces them
- No PEP, PDP, or policy layer — enforcement happens wherever the developer writes it in application code
- Must be paired with a separate authorization library or policy engine for any meaningful access control

---

## 6. Keycloak

**Type:** Open-Source Enterprise Identity & Access Management Server

### Overview
Keycloak is an open-source IAM solution developed by Red Hat providing a full-featured, self-hosted identity server. The most widely adopted open-source IAM platform and the identity provider used directly by Custos.

### Core Features & Function
- Single Sign-On & Single Logout: SSO across all applications within a realm
- Standard Protocols: OAuth 2.0, OIDC, SAML 2.0, and LDAP/Active Directory federation
- User Federation: Sync and authenticate users from external LDAP/AD directories
- Fine-Grained Authorization: Policy-based authorization server with resource, scope, and policy definitions (UMA 2.0)
- Multi-Tenancy via Realms: Isolated authentication domains within a single Keycloak instance
- Themeable UI: Customizable login and registration pages via Freemarker templates

### How It Works (Architecture)
- Realm: Top-level organizational unit — a self-contained authentication domain. Custos uses realms to implement multi-tenancy, each science gateway tenant gets its own Keycloak realm
- Clients: Applications requesting authentication are registered as clients with type (public, confidential, bearer-only), redirect URIs, and protocol settings
- Token lifecycle: OAuth 2.0 Authorization Server and OIDC Provider issuing JWTs signed with realm-specific RSA key pairs
- Authorization Services: Includes a Policy Enforcement Point (PEP) and Policy Decision Point (PDP) architecture for resource-level authorization
- Deployment: Quarkus-based Java server. Custos runs Keycloak on port 8080
- Custos relevance: Custos creates and manages Keycloak realms programmatically when tenants are initialized

### Authorization Engine (User-Facing)
- Most complete built-in authorization engine of all platforms reviewed — full PDP and PEP architecture via Authorization Services
- Application-side enforcement: Roles and group memberships embedded in JWT via `realm_access.roles` and `resource_access.<client>.roles` claims — no additional server calls needed
- Engine-side enforcement: Applications can query Keycloak at runtime, receiving an RPT (Requesting Party Token) reflecting what the user is authorized to access, evaluated server-side against defined policies
- Policies can be role-based, user-based, group-based, JavaScript-based, or time-based, combined with AND/OR/NOT logic
- Unique among platforms reviewed: supports both application-side and engine-side enforcement depending on complexity of the authorization requirement