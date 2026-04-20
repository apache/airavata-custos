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

package store

import (
	"context"
	"database/sql"

	"github.com/apache/airavata-custos/allocations/domain/model"
)

type PersonStore interface {
	FindByID(ctx context.Context, id string) (*model.Person, error)
	FindByAccessGlobalID(ctx context.Context, globalID string) (*model.Person, error)
	FindActiveByEmail(ctx context.Context, email string) (*model.Person, error)
	Save(ctx context.Context, tx *sql.Tx, p *model.Person) error
	Update(ctx context.Context, tx *sql.Tx, p *model.Person) error
	Deactivate(ctx context.Context, tx *sql.Tx, id string) error
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

type PersonGlobalIDStore interface {
	FindPersonByGlobalID(ctx context.Context, globalID string) (*model.Person, error)
	Save(ctx context.Context, tx *sql.Tx, g *model.PersonGlobalID) error
	UpdatePersonID(ctx context.Context, tx *sql.Tx, oldPersonID, newPersonID string) error
}

type PersonDNStore interface {
	ExistsByPersonAndDN(ctx context.Context, personID, dn string) (bool, error)
	Save(ctx context.Context, tx *sql.Tx, d *model.PersonDN) error
	DeleteByPersonID(ctx context.Context, tx *sql.Tx, personID string) error
	DeleteByPersonIDNotIn(ctx context.Context, tx *sql.Tx, personID string, dnsToKeep []string) error
	FindByPersonID(ctx context.Context, personID string) ([]model.PersonDN, error)
}

type ClusterAccountStore interface {
	FindByUsername(ctx context.Context, username string) (*model.ClusterAccount, error)
	FindByPerson(ctx context.Context, personID string) ([]model.ClusterAccount, error)
	Save(ctx context.Context, tx *sql.Tx, a *model.ClusterAccount) error
	UpdatePersonID(ctx context.Context, tx *sql.Tx, accountID, newPersonID string) error
}

type ProjectStore interface {
	FindByID(ctx context.Context, id string) (*model.Project, error)
	Save(ctx context.Context, tx *sql.Tx, p *model.Project) error
	Update(ctx context.Context, tx *sql.Tx, p *model.Project) error
}

type MembershipStore interface {
	FindByProjectAndAccount(ctx context.Context, projectID, accountID string) (*model.ProjectMembership, error)
	FindByProject(ctx context.Context, projectID string) ([]model.ProjectMembership, error)
	FindByProjectAndRole(ctx context.Context, projectID, role string) ([]model.ProjectMembership, error)
	FindByProjectAndPerson(ctx context.Context, projectID, personID string) ([]model.ProjectMembership, error)
	Save(ctx context.Context, tx *sql.Tx, m *model.ProjectMembership) error
	Update(ctx context.Context, tx *sql.Tx, m *model.ProjectMembership) error
}
