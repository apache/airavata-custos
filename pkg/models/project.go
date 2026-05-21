package models

import "time"

// UserStatus enumerates the lifecycle states a User may occupy.
type UserStatus string

const (
	UserActive    UserStatus = "ACTIVE"
	UserInactive  UserStatus = "INACTIVE"
	UserSuspended UserStatus = "SUSPENDED"
	UserMerged    UserStatus = "MERGED"
)

// ProjectStatus enumerates the lifecycle states a Project may occupy.
type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "ACTIVE"
	ProjectInactive ProjectStatus = "INACTIVE"
	ProjectDeleted  ProjectStatus = "DELETED"
)

type Project struct {
	ID           string        `json:"id"            db:"id"`
	OriginatedID string        `json:"originated_id" db:"originated_id"` // The ID of the project in origination. For example: ACCESS Record ID.
	Title        string        `json:"title"         db:"title"`
	Origination  string        `json:"origination"   db:"origination"` // ACCESS, NAIRR, XRASS, etc.
	ProjectPIID  string        `json:"project_pi_id" db:"project_pi_id"`
	Status       ProjectStatus `json:"status"        db:"status"`
	CreatedTime  time.Time     `json:"created_time"  db:"created_time"`
}

type Organization struct {
	ID           string `json:"id"            db:"id"`
	OriginatedID string `json:"originated_id" db:"originated_id"` // The ID of the organization in origination. For example: ACCESS Record ID.
	Name         string `json:"name"          db:"name"`
}

type User struct {
	ID             string     `json:"id"              db:"id"`
	OrganizationID string     `json:"organization_id" db:"organization_id"`
	FirstName      string     `json:"first_name"      db:"first_name"`
	LastName       string     `json:"last_name"       db:"last_name"`
	MiddleName     string     `json:"middle_name,omitempty" db:"middle_name"`
	Email          string     `json:"email"           db:"email"`
	Status         UserStatus `json:"status"          db:"status"`
}
