# Security Track Research — Dynamic Access Policy & Enforcement Engine

**Contributor:** Temitope Aderibigbe  
**Georgia Tech VIP Program | Spring 2026**  
**Related Issue:** [AIRAVATA-3978](https://issues.apache.org/jira/browse/AIRAVATA-3978)

## Overview

This directory contains research conducted on the security track for Apache Airavata Custos, focused on informing the design of a Dynamic Access Policy & Enforcement Engine (Project 2). The research covers identity and authentication platform analysis, authorization engine patterns, and a deep dive into Google's Zanzibar authorization system.

## Documents

- [`identity-platform-notes.md`](./identity-platform-notes.md) — Research notes on six identity/authentication platforms: Auth0, WorkOS, WSO2 Asgardeo, Amazon Cognito, Eggshell, and Keycloak. Covers core features, architecture, and user-facing authorization engine for each.
- [`zanzibar-paper-notes.md`](./zanzibar-paper-notes.md) — Notes on the Google Zanzibar paper (USENIX ATC 2019), covering the relation tuple data model, consistency model, API, system architecture, and direct relevance to Custos Project 2.

## Key Findings

- Custos currently handles authentication via Keycloak but has no general-purpose authorization engine for attribute or policy-based access decisions
- Among platforms reviewed, only Keycloak and WorkOS FGA offer true engine-side enforcement — the rest rely on application-side token claim checking
- Zanzibar's relation tuple model and per-namespace policy configs directly map to Project 2's requirements for per-tenant policies and a centralized policy decision service
- Tools like Open Policy Agent (OPA) and AWS Cedar are strong candidates as the policy evaluation engine for Project 2