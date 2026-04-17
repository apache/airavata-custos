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

package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/google/uuid"
)

type personStore interface {
	FindByID(ctx context.Context, id string) (*model.Person, error)
	FindByAccessGlobalID(ctx context.Context, globalID string) (*model.Person, error)
	Save(ctx context.Context, tx *sql.Tx, p *model.Person) error
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

type personDNStore interface {
	ExistsByPersonAndDN(ctx context.Context, personID, dn string) (bool, error)
	Save(ctx context.Context, tx *sql.Tx, d *model.PersonDN) error
	DeleteByPersonID(ctx context.Context, tx *sql.Tx, personID string) error
	DeleteByPersonIDNotIn(ctx context.Context, tx *sql.Tx, personID string, dnsToKeep []string) error
	FindByPersonID(ctx context.Context, personID string) ([]model.PersonDN, error)
}

type personAccountStore interface {
	FindByPerson(ctx context.Context, personID string) ([]model.ClusterAccount, error)
	UpdatePersonID(ctx context.Context, tx *sql.Tx, accountID, newPersonID string) error
}

type PersonService struct {
	persons  personStore
	dns      personDNStore
	accounts personAccountStore
}

func NewPersonService(persons personStore, dns personDNStore, accounts personAccountStore) *PersonService {
	return &PersonService{
		persons:  persons,
		dns:      dns,
		accounts: accounts,
	}
}

// FindOrCreateFromPacket looks up a person by their ACCESS Global ID or
// creates a new person record from the supplied AMIE packet body.
func (s *PersonService) FindOrCreateFromPacket(ctx context.Context, tx *sql.Tx, body map[string]any) (*model.Person, error) {
	globalID, _ := body["UserGlobalID"].(string)
	if globalID == "" {
		return nil, fmt.Errorf("person_service: UserGlobalID is required")
	}

	existing, err := s.persons.FindByAccessGlobalID(ctx, globalID)
	if err != nil {
		return nil, fmt.Errorf("person_service: finding person by global ID %s: %w", globalID, err)
	}
	if existing != nil {
		return existing, nil
	}

	firstName, _ := body["UserFirstName"].(string)
	lastName, _ := body["UserLastName"].(string)
	email, _ := body["UserEmail"].(string)

	p := &model.Person{
		ID:             uuid.NewString(),
		AccessGlobalID: globalID,
		FirstName:      firstName,
		LastName:       lastName,
		Email:          email,
		Organization:   optionalString(body, "UserOrganization"),
		OrgCode:        optionalString(body, "UserOrgCode"),
		NsfStatusCode:  optionalString(body, "NsfStatusCode"),
	}

	if err := s.persons.Save(ctx, tx, p); err != nil {
		return nil, fmt.Errorf("person_service: saving new person %s: %w", p.ID, err)
	}

	slog.DebugContext(ctx, "created person from packet", "person_id", p.ID, "global_id", globalID)

	// Persist DN list if present.
	if dnList, ok := body["UserDnList"].([]any); ok {
		for _, raw := range dnList {
			dn, _ := raw.(string)
			if dn == "" {
				continue
			}
			d := &model.PersonDN{PersonID: p.ID, DN: dn}
			if err := s.dns.Save(ctx, tx, d); err != nil {
				return nil, fmt.Errorf("person_service: saving DN for person %s: %w", p.ID, err)
			}
		}
	}

	return p, nil
}

// ReplaceFromModifyPacket selectively updates person fields that are present
// in the AMIE modify-person packet body.
func (s *PersonService) ReplaceFromModifyPacket(ctx context.Context, tx *sql.Tx, body map[string]any) error {
	personID, _ := body["PersonID"].(string)
	if personID == "" {
		return fmt.Errorf("person_service: PersonID is required for modify")
	}

	p, err := s.persons.FindByID(ctx, personID)
	if err != nil {
		return fmt.Errorf("person_service: finding person %s: %w", personID, err)
	}
	if p == nil {
		return fmt.Errorf("person_service: person %s not found", personID)
	}

	// Update only fields present in the body.
	if v, ok := body["UserFirstName"]; ok {
		p.FirstName, _ = v.(string)
	}
	if v, ok := body["UserLastName"]; ok {
		p.LastName, _ = v.(string)
	}
	if v, ok := body["UserEmail"]; ok {
		p.Email, _ = v.(string)
	}
	if _, ok := body["UserOrganization"]; ok {
		p.Organization = optionalString(body, "UserOrganization")
	}
	if _, ok := body["UserOrgCode"]; ok {
		p.OrgCode = optionalString(body, "UserOrgCode")
	}
	if _, ok := body["NsfStatusCode"]; ok {
		p.NsfStatusCode = optionalString(body, "NsfStatusCode")
	}

	// Handle DN list updates.
	if rawDNs, ok := body["UserDnList"]; ok {
		dnList, isList := rawDNs.([]any)
		if isList && len(dnList) > 0 {
			var dnsToKeep []string
			for _, raw := range dnList {
				dn, _ := raw.(string)
				if dn != "" {
					dnsToKeep = append(dnsToKeep, dn)
				}
			}
			if err := s.dns.DeleteByPersonIDNotIn(ctx, tx, personID, dnsToKeep); err != nil {
				return fmt.Errorf("person_service: pruning DNs for person %s: %w", personID, err)
			}
			for _, dn := range dnsToKeep {
				exists, err := s.dns.ExistsByPersonAndDN(ctx, personID, dn)
				if err != nil {
					return fmt.Errorf("person_service: checking DN existence for person %s: %w", personID, err)
				}
				if !exists {
					d := &model.PersonDN{PersonID: personID, DN: dn}
					if err := s.dns.Save(ctx, tx, d); err != nil {
						return fmt.Errorf("person_service: saving DN for person %s: %w", personID, err)
					}
				}
			}
		} else if isList && len(dnList) == 0 {
			if err := s.dns.DeleteByPersonID(ctx, tx, personID); err != nil {
				return fmt.Errorf("person_service: clearing DNs for person %s: %w", personID, err)
			}
		}
	}

	if err := s.persons.Save(ctx, tx, p); err != nil {
		return fmt.Errorf("person_service: saving updated person %s: %w", personID, err)
	}

	slog.DebugContext(ctx, "updated person from modify packet", "person_id", personID)
	return nil
}

// DeleteFromModifyPacket removes a person identified by the packet body.
// Cascade rules in the database handle related cleanup.
func (s *PersonService) DeleteFromModifyPacket(ctx context.Context, tx *sql.Tx, body map[string]any) error {
	personID, _ := body["PersonID"].(string)
	if personID == "" {
		return fmt.Errorf("person_service: PersonID is required for delete")
	}

	if err := s.persons.Delete(ctx, tx, personID); err != nil {
		return fmt.Errorf("person_service: deleting person %s: %w", personID, err)
	}

	slog.DebugContext(ctx, "deleted person from modify packet", "person_id", personID)
	return nil
}

// MergePersons transfers all accounts and DNs from the retiring person to the
// surviving person, then deletes the retiring person.
func (s *PersonService) MergePersons(ctx context.Context, tx *sql.Tx, survivingID, retiringID string) error {
	surviving, err := s.persons.FindByID(ctx, survivingID)
	if err != nil {
		return fmt.Errorf("person_service: finding surviving person %s: %w", survivingID, err)
	}
	if surviving == nil {
		return fmt.Errorf("person_service: surviving person %s not found", survivingID)
	}

	retiring, err := s.persons.FindByID(ctx, retiringID)
	if err != nil {
		return fmt.Errorf("person_service: finding retiring person %s: %w", retiringID, err)
	}
	if retiring == nil {
		return fmt.Errorf("person_service: retiring person %s not found", retiringID)
	}

	// Move cluster accounts from the retiring person to the surviving person.
	retiringAccounts, err := s.accounts.FindByPerson(ctx, retiringID)
	if err != nil {
		return fmt.Errorf("person_service: finding accounts for retiring person %s: %w", retiringID, err)
	}
	for _, acct := range retiringAccounts {
		if err := s.accounts.UpdatePersonID(ctx, tx, acct.ID, survivingID); err != nil {
			return fmt.Errorf("person_service: moving account %s to surviving person %s: %w", acct.ID, survivingID, err)
		}
	}

	// Merge DNs: copy any DNs from the retiring person that the surviving
	// person does not already have.
	retiringDNs, err := s.dns.FindByPersonID(ctx, retiringID)
	if err != nil {
		return fmt.Errorf("person_service: finding DNs for retiring person %s: %w", retiringID, err)
	}
	for _, dn := range retiringDNs {
		exists, err := s.dns.ExistsByPersonAndDN(ctx, survivingID, dn.DN)
		if err != nil {
			return fmt.Errorf("person_service: checking DN existence for surviving person %s: %w", survivingID, err)
		}
		if !exists {
			d := &model.PersonDN{PersonID: survivingID, DN: dn.DN}
			if err := s.dns.Save(ctx, tx, d); err != nil {
				return fmt.Errorf("person_service: saving merged DN for surviving person %s: %w", survivingID, err)
			}
		}
	}

	// Delete the retiring person; cascade rules handle related records.
	if err := s.persons.Delete(ctx, tx, retiringID); err != nil {
		return fmt.Errorf("person_service: deleting retiring person %s: %w", retiringID, err)
	}

	slog.DebugContext(ctx, "merged persons", "surviving_id", survivingID, "retiring_id", retiringID)
	return nil
}

// PersistDNsForPerson saves any distinguished names that the person does not
// already have.
func (s *PersonService) PersistDNsForPerson(ctx context.Context, tx *sql.Tx, personID string, dnList []string) error {
	for _, dn := range dnList {
		exists, err := s.dns.ExistsByPersonAndDN(ctx, personID, dn)
		if err != nil {
			return fmt.Errorf("person_service: checking DN existence for person %s: %w", personID, err)
		}
		if exists {
			continue
		}
		d := &model.PersonDN{PersonID: personID, DN: dn}
		if err := s.dns.Save(ctx, tx, d); err != nil {
			return fmt.Errorf("person_service: saving DN for person %s: %w", personID, err)
		}
	}
	return nil
}

func optionalString(m map[string]any, key string) *string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	s, _ := v.(string)
	if s == "" {
		return nil
	}
	return &s
}
