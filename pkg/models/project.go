package models

import "time"

type Project struct {
	ID           string           `json:"id"            db:"id"`
	OriginatedID string           `json:"originated_id" db:"originated_id"` // The ID of the project in origination. For example: ACCESS Record ID.
	Title        string           `json:"title"         db:"title"`
	Origination  string           `json:"origination"   db:"origination"` // ACCESS, NAIRR, XRASS, etc.
	ProjectPIID  string           `json:"project_pi_id" db:"project_pi_id"`
	Status       AllocationStatus `json:"status"        db:"status"` // ACTIVE, INACTIVE, DELETED
	CreatedTime  time.Time        `json:"created_time"  db:"created_time"`
}

type Organization struct {
	ID           string `json:"id"            db:"id"`
	OriginatedID string `json:"originated_id" db:"originated_id"` // The ID of the organization in origination. For example: ACCESS Record ID.
	Name         string `json:"name"          db:"name"`
}

type UserStatus string

const (
	UserActive   UserStatus = "ACTIVE"
	UserInactive UserStatus = "INACTIVE"
	// UserMerged marks the retiring user after a Service.MergeUsers call.
	// The row is kept as a mapping so historical references stay resolvable, the
	// linkage to the surviving user lives in the user_merges table.
	UserMerged UserStatus = "MERGED"
)

type User struct {
	ID             string     `json:"id"                    db:"id"`
	OrganizationID string     `json:"organization_id"       db:"organization_id"`
	FirstName      string     `json:"first_name"            db:"first_name"`
	LastName       string     `json:"last_name"             db:"last_name"`
	MiddleName     string     `json:"middle_name,omitempty" db:"middle_name"`
	Email          string     `json:"email"                 db:"email"`
	Status         UserStatus `json:"status"                db:"status"`
}

// UserMerge records the linkage produced by Service.MergeUsers. The retiring
// user's row is kept (with status=MERGED), so historical references remain
// resolvable.
type UserMerge struct {
	ID              int64     `json:"id"                 db:"id"`
	RetiringUserID  string    `json:"retiring_user_id"   db:"retiring_user_id"`
	SurvivingUserID string    `json:"surviving_user_id"  db:"surviving_user_id"`
	Reason          string    `json:"reason,omitempty"   db:"reason"`
	MergedAt        time.Time `json:"merged_at"          db:"merged_at"`
}
