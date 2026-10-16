/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { enumLabel } from "@/shared/format";

// Plain names of audit event types, shared by the trace view and the AMIE
// packet drawer (a sanctioned cross-feature export, like ViewTraceLink).
export const EVENT_LABELS: Record<string, string> = {
  CLUSTER_ACCOUNT_APPROVED: "Cluster account approved",
  CLUSTER_ACCOUNT_DENIED: "Cluster account denied",
  CLUSTER_ADMIN_GRANTED: "Cluster admin access granted",
  PRIVILEGE_GRANTED: "Privilege granted",
  PRIVILEGE_REVOKED: "Privilege revoked",
  ROLE_CREATED: "Role created",
  ROLE_UPDATED: "Role updated",
  ROLE_DELETED: "Role deleted",
  ROLE_PRIVILEGE_ADDED: "Privilege added to role",
  ROLE_PRIVILEGE_REMOVED: "Privilege removed from role",
  ROLE_GRANTED: "Role granted",
  ROLE_REVOKED: "Role revoked",
  ROLE_BOOTSTRAPPED: "Role bootstrapped",
  USER_BOOTSTRAPPED: "Admin user bootstrapped",
  IDENTITY_LINKED: "Sign-in identity linked",
  USER_MERGED: "Users merged",
  NOTIFICATION_SENT: "Email sent",
  EVENT_DELIVERY_SUCCEEDED: "Delivery succeeded",
  EVENT_DELIVERY_FAILED: "Delivery failed",
  EVENT_DELIVERY_RETRIED: "Delivery retried by an admin",
  PACKET_RECEIVED: "Packet received",
  PACKET_FAILED: "Packet failed",
  CREATE_PERSON: "Person created",
  UPDATE_PERSON: "Person updated",
  DELETE_PERSON: "Person deleted",
  MERGE_PERSONS: "Persons merged",
  CREATE_ACCOUNT: "Cluster account created",
  CREATE_PROJECT: "Project created",
  CREATE_ALLOCATION: "Allocation created",
  ATTACH_RESOURCE: "Resource attached",
  INACTIVATE_PROJECT: "Project inactivated",
  REACTIVATE_PROJECT: "Project reactivated",
  CREATE_MEMBERSHIP: "Membership created",
  INACTIVATE_MEMBERSHIP: "Membership inactivated",
  REACTIVATE_MEMBERSHIP: "Membership reactivated",
  PERSIST_DNS: "DNs updated",
  REPLY_SENT: "Reply sent",
  REPLY_HELD: "Reply held for approval",
  TRANSACTION_COMPLETE: "Transaction complete",
  PosixUsernameBuildFailed: "Username could not be built",
  PosixUsernameTruncated: "Username truncated",
  PosixUsernameAllocatorExhausted: "No free username",
  ComanageProvisioningStarted: "Provisioning started",
  ComanageProvisioningFailed: "Provisioning failed",
  ComanageClusterAccountAttached: "Cluster account attached",
  ComanageClusterAccountRemoved: "Cluster account removed",
  ComanageCoPersonCreated: "COmanage person created",
  ComanageOIDCLinkageCreated: "Sign-in identity linked in COmanage",
  ComanageClusterAdminGroupJoined: "Added to the cluster admin group",
  ComanageAllocationGroupCreated: "Allocation group created",
  ComanageAllocationGroupDeleted: "Allocation group deleted",
  ComanageAllocationGroupMemberAdded: "Added to the allocation group",
  ComanageAllocationGroupMemberRemoved: "Removed from the allocation group",
  ComputeAllocationCreationSucceeded: "SLURM account created",
  ComputeAllocationCreationFailed: "SLURM account could not be created",
  ComputeAllocationUpdateFailed: "SLURM associations could not be updated",
  ComputeAllocationDeletionFailed: "SLURM associations could not be removed",
  ComputeAllocationProvisioned: "SLURM associations restored",
  ComputeAllocationDeprovisioned: "SLURM associations removed",
  ComputeAllocationMembershipCreationSucceeded: "SLURM user association created",
  ComputeAllocationMembershipCreationFailed: "SLURM user association could not be created",
  ComputeAllocationMembershipUpdateFailed: "SLURM user association could not be updated",
  ComputeAllocationMembershipDeletionFailed: "SLURM user association could not be removed",
  ComputeAllocationMembershipDeprovisioned: "SLURM user association removed",
  ComputeAllocationMembershipResourceOverrideCreationSucceeded: "SLURM user limit set",
  ComputeAllocationMembershipResourceOverrideCreationFailed: "SLURM user limit could not be set",
  ComputeAllocationResourceMappingCreationSucceeded: "SLURM resource limit set",
  ComputeAllocationResourceMappingCreationFailed: "SLURM resource limit could not be set",
};

// Plain names of allocation diff types; AMIE writes its AllocationType upper-cased.
export const DIFF_LABELS: Record<string, string> = {
  ALLOCATION_STATUS_CHANGE: "Status changed",
  USAGE_UPDATE: "Usage updated",
  CHANGE_REQUEST_APPROVED: "Change request approved",
  MANUAL_ADJUSTMENT: "Manual adjustment",
  ADMIN: "Admin change",
  NEW: "New grant",
  RENEWAL: "Renewal",
  SUPPLEMENT: "Supplement",
  EXTENSION: "Extension",
  TRANSFER: "Transfer",
  ADJUSTMENT: "Adjustment",
  ADVANCE: "Advance",
};

const ROLE_TOKENS: Record<string, string> = { PI: "PI", CO_PI: "Co-PI" };

// A description that is a bare enum token (DECODED, CO_PI) reads as its label; other text as written.
export const tokenLabel = (text: string) =>
  /^[A-Z][A-Z0-9_]*$/.test(text) ? (ROLE_TOKENS[text] ?? enumLabel(text)) : text;
