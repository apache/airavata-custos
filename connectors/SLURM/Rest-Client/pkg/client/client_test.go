// cli/internal/client/client_test.go
package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoSetsHeaders(t *testing.T) {
	var gotName, gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotName = r.Header.Get("X-SLURM-USER-NAME")
		gotToken = r.Header.Get("X-SLURM-USER-TOKEN")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "root", "tok123", "41")
	if _, err := c.do("GET", "/slurm/v0.0.41/ping", nil, nil); err != nil {
		t.Fatal(err)
	}
	if gotName != "root" {
		t.Errorf("X-SLURM-USER-NAME = %q", gotName)
	}
	if gotToken != "tok123" {
		t.Errorf("X-SLURM-USER-TOKEN = %q", gotToken)
	}
}

func TestDoUnwrapsErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"errors":[{"description":"nope","error_number":42,"error":"Bad","source":"test"}]}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "root", "tok", "41")
	_, err := c.do("GET", "/x", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "42") || !strings.Contains(err.Error(), "nope") {
		t.Errorf("err = %v", err)
	}
}
