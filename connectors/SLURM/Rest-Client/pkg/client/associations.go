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

	fetchedLimits := fetchedAssociation.Limits
	upsertedLimits := a.Limits

	if !tresSliceEqualUnordered(fetchedLimits.GrpTRES, upsertedLimits.GrpTRES) {
		slog.Error("Fetched association limits GrpTRES does not match upserted association limits GrpTRES", "got", fetchedLimits.GrpTRES, "want", upsertedLimits.GrpTRES)
		return errors.New("fetched association limits GrpTRES does not match upserted association limits GrpTRES")
	}

	if !tresSliceEqualUnordered(fetchedLimits.GrpTRESMins, upsertedLimits.GrpTRESMins) {
		slog.Error("Fetched association limits GrpTRESMins does not match upserted association limits GrpTRESMins", "got", fetchedLimits.GrpTRESMins, "want", upsertedLimits.GrpTRESMins)
		return errors.New("fetched association limits GrpTRESMins does not match upserted association limits GrpTRESMins")
	}

	return nil
}

// tresSliceEqualUnordered reports whether two TRES slices contain the same
// elements regardless of order. nil and empty slices are treated as equal.
func tresSliceEqualUnordered(a, b []TRES) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[TRES]int, len(a))
	for _, t := range a {
		counts[t]++
	}
	for _, t := range b {
		counts[t]--
		if counts[t] < 0 {
			return false
		}
	}
	return true
}

func (c *Client) DeleteAssociation(f AssocFilter) error {
	path := "/slurmdb/v0.0." + c.apiVersion + "/association"
	if q := f.query(); q != "" {
		path += "?" + q
	}
	_, err := c.do("DELETE", path, nil, nil)
	return err
}
