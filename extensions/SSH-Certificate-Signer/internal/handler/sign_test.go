// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func signError(t *testing.T, body string) (int, string) {
	t.Helper()
	h := NewSignHandler(nil, nil, nil, nil, SignSettings{ReservedUsernames: []string{"root"}}, slog.Default())
	rec := httptest.NewRecorder()
	h.Handle(rec, httptest.NewRequest(http.MethodPost, "/api/v1/sign", strings.NewReader(body)))
	var got struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	return rec.Code, got.Error
}

// Make sure a reserved username is refused before any validator or key is used.
func TestSign_RefusesReservedUsername(t *testing.T) {
	code, errorCode := signError(t, `{"principal":"root","public_key":"k","user_email":"a@example.edu","user_sub":"s","requested_by":"user/1"}`)
	if code != http.StatusForbidden || errorCode != "reserved_username" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure a value that could forge a login log line or make two key IDs the same is refused.
func TestSign_RefusesUnsafeAttributes(t *testing.T) {
	for name, body := range map[string]string{
		"newline in email":       `{"principal":"jdoe","public_key":"k","user_email":"a@example.edu\nfake","user_sub":"s","requested_by":"user/1"}`,
		"separator in requester": `{"principal":"jdoe","public_key":"k","user_email":"a@example.edu","user_sub":"s","requested_by":"b:c"}`,
	} {
		if code, errorCode := signError(t, body); code != http.StatusBadRequest || errorCode != "invalid_request" {
			t.Errorf("%s: got %d %s", name, code, errorCode)
		}
	}
}

// Make sure an address list sshd would reject or never match is refused at signing.
func TestSign_RefusesUnreadableSourceAddress(t *testing.T) {
	for _, address := range []string{"10.0.0.5/24", "10.0.0.1, 10.0.0.2", "10.0.0.1,", "login.example.edu", "fe80::1%eth0", "::ffff:10.0.0.1", "::ffff:10.0.0.0/104"} {
		body := `{"principal":"jdoe","public_key":"k","user_email":"a@example.edu","user_sub":"s","requested_by":"user/1","source_address":"` + address + `"}`
		if code, errorCode := signError(t, body); code != http.StatusBadRequest || errorCode != "invalid_request" {
			t.Errorf("%s: got %d %s", address, code, errorCode)
		}
	}
}

// Make sure a certificate cannot be issued without saying who it is for and who asked.
func TestSign_RequiresUserAndRequester(t *testing.T) {
	for _, body := range []string{
		`{"principal":"jdoe","public_key":"k","user_sub":"s","requested_by":"user/1"}`,
		`{"principal":"jdoe","public_key":"k","user_email":"a@example.edu","requested_by":"user/1"}`,
		`{"principal":"jdoe","public_key":"k","user_email":"a@example.edu","user_sub":"s"}`,
	} {
		if code, errorCode := signError(t, body); code != http.StatusBadRequest || errorCode != "invalid_request" {
			t.Errorf("%s: got %d %s", body, code, errorCode)
		}
	}
}

func TestPrincipalRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		// Valid cases
		{"simple lowercase", "jdoe", true},
		{"underscore start", "_user", true},
		{"with hyphen", "j-doe", true},
		{"with underscore", "j_doe", true},
		{"with numbers", "user123", true},
		{"single char", "a", true},
		{"32 chars", "abcdefghijklmnopqrstuvwxyz012345", true},

		// Invalid cases
		{"uppercase", "JDoe", false},
		{"dot", "j.doe", false},
		{"at sign", "jdoe@example", false},
		{"digit start", "1user", false},
		{"33 chars", "abcdefghijklmnopqrstuvwxyz0123456", false},
		{"empty", "", false},
		{"space", "j doe", false},
		{"special chars", "user$", false},
		{"John.Doe", "John.Doe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := principalRegex.MatchString(tt.input)
			if got != tt.valid {
				t.Errorf("principalRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"generic error", errors.New("connection refused"), false},
		{"unique violation", &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint \"serial_number\""}, true},
		{"wrapped unique violation", fmt.Errorf("failed to insert: %w", &pgconn.PgError{Code: "23505"}), true},
		{"other pg error", &pgconn.PgError{Code: "23503", Message: "foreign key violation"}, false},
		{"plain message no match", errors.New("duplicate key value violates unique constraint"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDuplicateKeyError(tt.err)
			if got != tt.want {
				t.Errorf("isDuplicateKeyError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
