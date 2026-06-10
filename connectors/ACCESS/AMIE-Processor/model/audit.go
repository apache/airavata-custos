// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package model

// AuditAction identifies the kind of mutation recorded in an audit log entry.
type AuditAction string

const (
	AuditPacketReceived       AuditAction = "PACKET_RECEIVED"
	AuditCreatePerson         AuditAction = "CREATE_PERSON"
	AuditUpdatePerson         AuditAction = "UPDATE_PERSON"
	AuditDeletePerson         AuditAction = "DELETE_PERSON"
	AuditMergePersons         AuditAction = "MERGE_PERSONS"
	AuditCreateAccount        AuditAction = "CREATE_ACCOUNT"
	AuditCreateProject        AuditAction = "CREATE_PROJECT"
	AuditCreateAllocation     AuditAction = "CREATE_ALLOCATION"
	AuditInactivateProject    AuditAction = "INACTIVATE_PROJECT"
	AuditReactivateProject    AuditAction = "REACTIVATE_PROJECT"
	AuditCreateMembership     AuditAction = "CREATE_MEMBERSHIP"
	AuditInactivateMembership AuditAction = "INACTIVATE_MEMBERSHIP"
	AuditReactivateMembership AuditAction = "REACTIVATE_MEMBERSHIP"
	AuditPersistDNs           AuditAction = "PERSIST_DNS"
	AuditReplySent            AuditAction = "REPLY_SENT"
	AuditTransactionComplete  AuditAction = "TRANSACTION_COMPLETE"
)

// AmieAuditExtras carries the AMIE-specific references for a row written
// to the core audit_events table. One extras row is inserted per audit_events row.
type AmieAuditExtras struct {
	AuditEventID string  `db:"audit_event_id"`
	PacketID     string  `db:"packet_id"`
	EventID      *string `db:"event_id"`
}
