package models

import "time"

// ExternalIdentity links a User to its identifier in an external system
// (ACCESS, NAIRR, CILogon, etc.). One user may have many external identities.
//
// Source-specific attributes (e.g. NSF status code, ACCESS org code) belong in
// Metadata as a JSON-encoded blob. 'core' does not validate its shape.
type ExternalIdentity struct {
	ID         string    `json:"id"                  db:"id"`
	UserID     string    `json:"user_id"             db:"user_id"`
	Source     string    `json:"source"              db:"source"`      // e.g. access, nairr, cilogon, internal
	ExternalID string    `json:"external_id"         db:"external_id"` // the source's native identifier
	OIDCSub    string    `json:"oidc_sub,omitempty"  db:"oidc_sub"`    // OIDC subject when the source issues one
	Metadata   string    `json:"metadata,omitempty"  db:"metadata"`    // JSON-encoded source-specific fields
	CreatedAt  time.Time `json:"created_at"          db:"created_at"`
}

// UserDN binds an X.509 distinguished name (e.g. mTLS client cert subject) to a
// User. Append-only: DNs are credentials and are added or removed, never edited.
type UserDN struct {
	ID        string    `json:"id"         db:"id"`
	UserID    string    `json:"user_id"    db:"user_id"`
	DN        string    `json:"dn"         db:"dn"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
