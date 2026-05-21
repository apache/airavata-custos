// cli/internal/client/associations.go
package operations

import (
	"errors"
	"log"
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

	return nil
}

func (c *Client) DeleteAssociation(f AssocFilter) error {
	path := "/slurmdb/v0.0." + c.apiVersion + "/association"
	if q := f.query(); q != "" {
		path += "?" + q
	}
	_, err := c.do("DELETE", path, nil, nil)
	return err
}
