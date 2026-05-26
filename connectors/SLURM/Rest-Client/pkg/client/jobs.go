package client

/*
curl -s -X GET \
  "http://localhost:6820/slurmdb/v0.0.41/jobs" \
  -H "X-SLURM-USER-NAME: root" \
  -H "X-SLURM-USER-TOKEN: $SLURM_JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "users": ["root"],
    "start_time": {
      "set": true,
      "infinite": false,
      "number": 1746057600
    }
  }'
*/

import (
	"fmt"
)

type jobsResponse struct {
	Jobs []JobInfo `json:"jobs"`
}

type JobFilter struct {
	Users     []string     `json:"users,omitempty"`
	StartTime *slurmNumber `json:"start_time,omitempty"`
	EndTime   *slurmNumber `json:"end_time,omitempty"`
}

func (c *Client) ListJobs(filter JobFilter) ([]JobInfo, error) {
	var out jobsResponse
	if _, err := c.do("GET", "/slurmdb/v0.0."+c.apiVersion+"/jobs", filter, &out); err != nil {
		return nil, err
	}
	return out.Jobs, nil
}

func (c *Client) GetJob(id int64) (*JobInfo, error) {
	var out jobsResponse
	if _, err := c.do("GET", fmt.Sprintf("/slurmdb/v0.0.%s/job/%d", c.apiVersion, id), nil, &out); err != nil {
		return nil, err
	}
	if len(out.Jobs) == 0 {
		return nil, fmt.Errorf("job %d not found", id)
	}
	return &out.Jobs[0], nil
}

func (c *Client) SubmitJob(request JobSubmitRequest) (*JobSubmitResponse, error) {
	var out JobSubmitResponse
	if _, err := c.do("POST", fmt.Sprintf("/slurm/v0.0.%s/job/submit", c.apiVersion), request, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
