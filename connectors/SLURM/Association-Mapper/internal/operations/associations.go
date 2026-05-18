// cli/internal/client/associations.go
package operations

import "net/url"

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
	return err
}

func (c *Client) DeleteAssociation(f AssocFilter) error {
	path := "/slurmdb/v0.0." + c.apiVersion + "/association"
	if q := f.query(); q != "" {
		path += "?" + q
	}
	_, err := c.do("DELETE", path, nil, nil)
	return err
}
