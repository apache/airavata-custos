// cli/internal/client/client.go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	user       string
	token      string
	apiVersion string
	http       *http.Client
}

func New(baseURL, user, token, apiVersion string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		user:       user,
		token:      token,
		apiVersion: apiVersion,
		http:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) do(method, path string, body any, out any) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-SLURM-USER-NAME", c.user)
	req.Header.Set("X-SLURM-USER-TOKEN", c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		buf, _ := io.ReadAll(resp.Body)
		var er ErrorResponse
		if json.Unmarshal(buf, &er) == nil && len(er.Errors) > 0 {
			e := er.Errors[0]
			return resp, fmt.Errorf("slurmrestd %d: %s (code=%d source=%s)",
				resp.StatusCode, e.Description, e.ErrorNumber, e.Source)
		}
		return resp, fmt.Errorf("slurmrestd %d: %s", resp.StatusCode, string(buf))
	}
	if out != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp, fmt.Errorf("decode: %w", err)
		}
	}
	return resp, nil
}
