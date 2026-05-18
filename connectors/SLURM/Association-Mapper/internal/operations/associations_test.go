// cli/internal/client/associations_test.go
package operations

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestListAssociationsByAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/slurmdb/v0.0.41/associations" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got, _ := url.QueryUnescape(r.URL.RawQuery); got != "account=eng" {
			t.Fatalf("query = %q", got)
		}
		_, _ = w.Write([]byte(`{"associations":[{"account":"eng","cluster":"artisan","user":"alice","id_association":5}]}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "root", "t", "41")
	assocs, err := c.ListAssociations(AssocFilter{Account: "eng"})
	if err != nil {
		t.Fatal(err)
	}
	if len(assocs) != 1 || assocs[0].User != "alice" || assocs[0].ID != 5 {
		t.Errorf("assocs = %+v", assocs)
	}
}

func TestCreateAssociation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/slurmdb/v0.0.41/associations" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		var payload struct {
			Associations []Association `json:"associations"`
		}
		_ = json.Unmarshal(b, &payload)
		if len(payload.Associations) != 1 || payload.Associations[0].Account != "eng" ||
			payload.Associations[0].User != "alice" {
			t.Errorf("payload = %+v", payload)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "root", "t", "41")
	err := c.UpsertAssociation(Association{Account: "eng", Cluster: "artisan", User: "alice"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAssociation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/slurmdb/v0.0.41/association" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("account") != "eng" || r.URL.Query().Get("user") != "alice" {
			t.Fatalf("query = %v", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "root", "t", "41")
	if err := c.DeleteAssociation(AssocFilter{Account: "eng", User: "alice"}); err != nil {
		t.Fatal(err)
	}
}
