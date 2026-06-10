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
	StartTime int64        `json:"start_time,omitempty"`
	EndTime   int64 		`json:"end_time,omitempty"`
}

type internalJobFilter struct {
	Users     []string `json:"users,omitempty"`
	StartTime *SlurmNumber `json:"start_time,omitempty"`
	EndTime   *SlurmNumber `json:"end_time,omitempty"`
}

func (c *Client) ListJobs(filter JobFilter) ([]JobInfo, error) {
	var out jobsResponse
	internalFilter := internalJobFilter{
		Users:     filter.Users,
	}
	if filter.StartTime > 0 {
		internalFilter.StartTime = &SlurmNumber{Set: true, Infinite: false, Number: filter.StartTime}
	}
	if filter.EndTime > 0 {
		internalFilter.EndTime = &SlurmNumber{Set: true, Infinite: false, Number: filter.EndTime}
	}

	if _, err := c.do("GET", "/slurmdb/v0.0."+c.apiVersion+"/jobs", internalFilter, &out); err != nil {
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
