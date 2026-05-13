// cli/internal/client/accounts_test.go
package operations

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListAccounts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/slurmdb/v0.0.41/accounts" || r.Method != "GET" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"accounts":[{"name":"root","description":"root account","organization":"artisan"}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "root", "t")
	accts, err := c.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(accts) != 1 || accts[0].Name != "root" {
		t.Errorf("accts = %+v", accts)
	}
}

func TestCreateAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/slurmdb/v0.0.41/accounts_association/" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			AssociationCondition struct {
				Accounts []string `json:"accounts"`
				Clusters []string `json:"clusters"`
			} `json:"association_condition"`
			Account struct {
				Description  string `json:"description"`
				Organization string `json:"organization"`
			} `json:"account"`
		}
		_ = json.Unmarshal(body, &payload)
		if len(payload.AssociationCondition.Accounts) != 1 || payload.AssociationCondition.Accounts[0] != "eng" {
			t.Errorf("accounts = %+v", payload.AssociationCondition.Accounts)
		}
		if len(payload.AssociationCondition.Clusters) != 1 || payload.AssociationCondition.Clusters[0] != "artisan" {
			t.Errorf("clusters = %+v", payload.AssociationCondition.Clusters)
		}
		if payload.Account.Description != "engineering" {
			t.Errorf("description = %q", payload.Account.Description)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "root", "t")
	if err := c.CreateAccount(Account{Name: "eng", Description: "engineering"}, "artisan"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/slurmdb/v0.0.41/account/eng" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "root", "t")
	if err := c.DeleteAccount("eng"); err != nil {
		t.Fatal(err)
	}
}

func TestGetAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/slurmdb/v0.0.41/account/eng" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"accounts":[{"name":"eng","description":"engineering","organization":"artisan"}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "root", "t")
	a, err := c.GetAccount("eng")
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "eng" || a.Description != "engineering" {
		t.Errorf("a = %+v", a)
	}
}
