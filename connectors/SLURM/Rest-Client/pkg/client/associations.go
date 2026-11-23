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

// cli/internal/client/associations.go
package client

import (
	"errors"
	"log"
	"log/slog"
	"net/url"
)

type AssocFilter struct {
	Account   string
	User      string
	Cluster   string
	Partition string
}

func (f AssocFilter) query() string {
	v := url.Values{}
	if f.Account != "" {
		v.Set("account", f.Account)
	}
	if f.User != "" {
		v.Set("user", f.User)
	}
	if f.Cluster != "" {
		v.Set("cluster", f.Cluster)
	}
	if f.Partition != "" {
		v.Set("partition", f.Partition)
	}
	return v.Encode()
}

type associationsResponse struct {
	Associations []Association `json:"associations"`
}

func (c *Client) ListAssociations(f AssocFilter) ([]Association, error) {
	path := "/slurmdb/v0.0." + c.apiVersion + "/associations"
	if q := f.query(); q != "" {
		path += "?" + q
	}
	var out associationsResponse
	if _, err := c.do("GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Associations, nil
}

// UpsertAssociation creates or updates an association. slurmrestd POST
// /slurmdb/v0.0.41/associations is an upsert: if the (cluster,account,user)
// triple exists, it's updated; otherwise created.
func (c *Client) UpsertAssociation(a Association) error {
	body := map[string]any{"associations": []Association{a}}
	_, err := c.do("POST", "/slurmdb/v0.0."+c.apiVersion+"/associations", body, nil)

	if err != nil {
		log.Printf("Failed to upsert association: %v", err)
		return err
	}

	filter := AssocFilter{
		Account:   a.Account,
		User:      a.User,
		Cluster:   a.Cluster,
		Partition: a.Partition,
	}
	assos, err := c.ListAssociations(filter)
	if err != nil {
		log.Printf("Failed to list associations after upsert: %v", err)
		return err
	}

	if len(assos) == 0 {
		log.Printf("No associations found after upsert")
		return errors.New("association not found after upsert")
	}

	log.Printf("Fetched upserted association: %+v", assos[0])
	fetchedAssociation := assos[0]
	if fetchedAssociation.Account != a.Account || fetchedAssociation.User != a.User ||
		fetchedAssociation.Cluster != a.Cluster || fetchedAssociation.Partition != a.Partition {
		log.Printf("Fetched association does not match upserted association: got %+v, want %+v", fetchedAssociation, a)
		return errors.New("fetched association does not match upserted association")
	}

	currentLimits := fetchedAssociation.Limits
	desiredLimits := a.Limits

	if NeedsTRESUpdate(currentLimits.GrpTRES, desiredLimits.GrpTRES) {
		slog.Error("Fetched association still needs a GrpTRES update", "current", currentLimits.GrpTRES, "desired", desiredLimits.GrpTRES)
		return errors.New("fetched association still needs a GrpTRES update")
	}

	if NeedsTRESUpdate(currentLimits.GrpTRESMins, desiredLimits.GrpTRESMins) {
		slog.Error("Fetched association still needs a GrpTRESMins update", "current", currentLimits.GrpTRESMins, "desired", desiredLimits.GrpTRESMins)
		return errors.New("fetched association still needs a GrpTRESMins update")
	}

	return nil
}

// NoLimit as a TRES count clears that limit on write.
const NoLimit = -1

// tresKey identifies a limit regardless of its count.
type tresKey struct{ typ, name string }

// NeedsTRESUpdate reports if current differs from any limit desired sets or clears.
// Limits only in current are ignored.
func NeedsTRESUpdate(current, desired []TRES) bool {
	inCurrent := make(map[tresKey]float64, len(current))
	for _, t := range current {
		inCurrent[tresKey{t.Type, t.Name}] = t.Count
	}
	for _, t := range desired {
		count, isSet := inCurrent[tresKey{t.Type, t.Name}]
		isAlreadyCleared := t.Count == NoLimit && !isSet
		isSame := isSet && count == t.Count
		if !isAlreadyCleared && !isSame {
			return true
		}
	}
	return false
}

// WithTRESSetOutside returns desired plus each current limit whose type and name desired does not set.
func WithTRESSetOutside(current, desired []TRES) []TRES {
	inDesired := make(map[tresKey]bool, len(desired))
	for _, t := range desired {
		inDesired[tresKey{t.Type, t.Name}] = true
	}
	merged := append([]TRES(nil), desired...)
	for _, t := range current {
		if !inDesired[tresKey{t.Type, t.Name}] {
			merged = append(merged, t)
		}
	}
	return merged
}

// DeleteAssociation deletes the user associations matching f one at a time, as slurmrestd requires.
func (c *Client) DeleteAssociation(f AssocFilter) error {
	matched, err := c.ListAssociations(f)
	if err != nil {
		return err
	}
	for _, assoc := range matched {
		if assoc.User == "" {
			continue
		}
		item := AssocFilter{Cluster: assoc.Cluster, Account: assoc.Account, User: assoc.User, Partition: assoc.Partition}
		if _, err := c.do("DELETE", "/slurmdb/v0.0."+c.apiVersion+"/association?"+item.query(), nil, nil); err != nil {
			return err
		}
	}
	return nil
}

// SetDefaultAccount gives user an association on account and makes it their default, creating the user if needed.
func (c *Client) SetDefaultAccount(cluster, user, account string) error {
	defaultAccount := map[string]string{"account": account}
	body := map[string]any{
		"association_condition": map[string]any{"accounts": []string{account}, "clusters": []string{cluster}, "users": []string{user}},
		"user":                  map[string]any{"default": defaultAccount},
	}
	if _, err := c.do("POST", "/slurmdb/v0.0."+c.apiVersion+"/users_association/", body, nil); err != nil {
		return err
	}
	// users_association leaves an existing user's default as it was.
	users := map[string]any{"users": []map[string]any{{"name": user, "default": defaultAccount}}}
	_, err := c.do("POST", "/slurmdb/v0.0."+c.apiVersion+"/users", users, nil)
	return err
}
