package models

import "time"

type Project struct {
	ID           string    `json:"id"`
	OriginatedID string    `json:"originated_id"` // The ID of the project in origination. For example: ACCESS Record ID.
	Title        string    `json:"title"`
	Origination  string    `json:"origination"` // ACCESS, NAIRR, XRASS, etc.
	ProjectPIID  string    `json:"project_pi_id"`
	CreatedTime  time.Time `json:"created_time"`
}

type Organization struct {
	ID           string `json:"id"`
	OriginatedID string `json:"originated_id"` // The ID of the organization in origination. For example: ACCESS Record ID.
	Name         string `json:"name"`
}

type User struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	MiddleName     string `json:"middle_name,omitempty"`
	Email          string `json:"email"`
}
