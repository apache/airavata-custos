package models

import "time"

type AllocationStatus string

const (
	ACTIVE   AllocationStatus = "ACTIVE"
	INACTIVE AllocationStatus = "INACTIVE"
	DELETED  AllocationStatus = "DELETED"
)

type ComputeCluster struct {
	ID   string `json:"id"   db:"id"`
	Name string `json:"name" db:"name"` // A human-readable name for the compute cluster, e.g., "Cluster A", "Cluster B", etc.
}

// ClusterAccount represents the provisioned posix account a user holds on the
// deployment's compute cluster (e.g. username "alice123"). Each Custos
// deployment is wired to a single downstream cluster, so the cluster scope is
// implicit. Username is globally unique within the deployment.
type ClusterAccount struct {
	ID       string           `json:"id"       db:"id"`
	UserID   string           `json:"user_id"  db:"user_id"`
	Username string           `json:"username" db:"username"`
	Status   AllocationStatus `json:"status"   db:"status"` // ACTIVE, INACTIVE, DELETED
}

type ComputeAllocation struct {
	ID               string           `json:"id"                 db:"id"`
	ProjectID        string           `json:"project_id"         db:"project_id"`
	Name             string           `json:"name"               db:"name"`
	Status           AllocationStatus `json:"status"             db:"status"`             // ACTIVE, INACTIVE, DELETED, etc.
	ComputeClusterID string           `json:"compute_cluster_id" db:"compute_cluster_id"` // The ID of the compute cluster where the allocation is provisioned.
	InitialSUAmount  int64            `json:"initial_su_amount"  db:"initial_su_amount"`  // SUs allocated at the time of allocation creation.
	StartTime        time.Time        `json:"start_time"         db:"start_time"`
	EndTime          time.Time        `json:"end_time"           db:"end_time"`
}

type ComputeAllocationResource struct {
	ID             string `json:"id"              db:"id"`
	Name           string `json:"name"            db:"name"`            // A human-readable name for the resource, e.g., "GPU B200", "CPU", "GPU RTX6000", etc.
	ResourceType   string `json:"resource_type"   db:"resource_type"`   // CPU, GPU, etc.
	ResourceAmount int64  `json:"resource_amount" db:"resource_amount"` // Number of CPUs, GPUs, etc. allocated.
}

type ComputeAllocationResourceMapping struct {
	ID                          string `json:"id"                             db:"id"`
	ComputeAllocationID         string `json:"compute_allocation_id"          db:"compute_allocation_id"`
	ComputeAllocationResourceID string `json:"compute_allocation_resource_id" db:"compute_allocation_resource_id"`
}

type ComputeAllocationResourceRate struct {
	ID                          string    `json:"id"                             db:"id"`
	ComputeAllocationResourceID string    `json:"compute_allocation_resource_id" db:"compute_allocation_resource_id"`
	Rate                        float64   `json:"rate"                           db:"rate"`       // The rate for the resource in SUs per unit, e.g., 0.5 SU per CPU hour, 2 SU per GPU hour, etc.
	StartTime                   time.Time `json:"start_time"                     db:"start_time"` // The time when this rate becomes effective.
	EndTime                     time.Time `json:"end_time"                       db:"end_time"`   // The time when this rate expires.
}

type ComputeAllocationDiff struct { // Diff will occur either through a change reqest or automated workflow like ACCESS AIME
	ID                  string           `json:"id"                    db:"id"`
	ComputeAllocationID string           `json:"compute_allocation_id" db:"compute_allocation_id"`
	DiffType            string           `json:"diff_type"             db:"diff_type"`     // "USAGE_UPDATE", "ALLOCATION_STATUS_CHANGE", etc.
	NewSUAmount         int64            `json:"new_su_amount"         db:"new_su_amount"` // New allocation amount in SUs, e.g., 900 SUs, etc.
	Status              AllocationStatus `json:"status"                db:"status"`        // ACTIVE, INACTIVE, DELETED, etc.
	Timestamp           time.Time        `json:"timestamp"             db:"timestamp"`     // The time when the diff was generated.
	Description         string           `json:"description,omitempty" db:"description"`   // Optional description of the diff, e.g., "SU usage updated based on job completion", "Allocation marked as INACTIVE due to end time reached", etc.
}

type ComputeAllocationChangeRequest struct { // Represents a request to change the allocation, e.g., requesting more SUs, requesting a reduction in SUs, etc from users or admins.
	ID                  string           `json:"id"                    db:"id"`
	ComputeAllocationID string           `json:"compute_allocation_id" db:"compute_allocation_id"`
	RequestedSUAmount   int64            `json:"requested_su_amount"   db:"requested_su_amount"` // The requested allocation amount in SUs, e.g., 1200 SUs, etc.
	RequestedStatus     AllocationStatus `json:"requested_status"      db:"requested_status"`    // ACTIVE, INACTIVE, DELETED, etc.
	Reason              string           `json:"reason"                db:"reason"`              // The reason for the change request, e.g., "Need more SUs for upcoming jobs", "Requesting reduction in SUs due to project completion", etc.
	ChangeStatus        string           `json:"change_status"         db:"change_status"`       // "PENDING", "APPROVED", "REJECTED", etc.
	RequesterID         string           `json:"requester_id"          db:"requester_id"`        // The ID of the user who made the change request.
	ApproverID          string           `json:"approver_id,omitempty" db:"approver_id"`         // The ID of the user who approved/rejected the change request, if applicable.
	Timestamp           time.Time        `json:"timestamp"             db:"timestamp"`           // The time when the change request was made.
}

type ComputeAllocationChangeRequestEvent struct {
	ID                               string    `json:"id"                                   db:"id"`
	ComputeAllocationChangeRequestID string    `json:"compute_allocation_change_request_id" db:"compute_allocation_change_request_id"`
	EventType                        string    `json:"event_type"                           db:"event_type"`  // "CREATED", "APPROVED", "REJECTED", etc.
	Description                      string    `json:"description,omitempty"                db:"description"` // Optional description of the event, e.g., "Change request created by user", "Change request approved by admin", etc.
	Timestamp                        time.Time `json:"timestamp"                            db:"timestamp"`   // The time when the event occurred.
}

type ComputeAllocationUsage struct { // Represents the usage of a compute allocation, e.g., when a job consumes some of the allocated SUs, etc.
	ID                          string    `json:"id"                             db:"id"`
	ComputeAllocationID         string    `json:"compute_allocation_id"          db:"compute_allocation_id"`
	UsedRawAmount               int64     `json:"used_raw_amount"                db:"used_raw_amount"`                // The raw amount of resource used, e.g., 20 CPU hours, 10 GPU hours, etc.
	UsedSUAmount                int64     `json:"used_su_amount"                 db:"used_su_amount"`                 // SUs used by the allocation, e.g., 200 SUs, etc.
	CalculatedTime              time.Time `json:"last_updated"                   db:"calculated_time"`                // The last time the usage was updated. SU should be calculated up to this point in time and charge rates should be applied based on the rates effective at this time.
	UserID                      string    `json:"user_id"                        db:"user_id"`                        // The ID of the user who used the allocation.
	JobID                       string    `json:"job_id"                         db:"job_id"`                         // The ID of the job that consumed the allocation.
	ComputeAllocationResourceID string    `json:"compute_allocation_resource_id" db:"compute_allocation_resource_id"` // The specific resource consumed, e.g., 20 CPU hours, 10 GPU hours, etc.
}

type ComputeAllocationMembership struct {
	ID                  string           `json:"id"                    db:"id"`
	ComputeAllocationID string           `json:"compute_allocation_id" db:"compute_allocation_id"`
	UserID              string           `json:"user_id"               db:"user_id"`
	AllocationAmount    int64            `json:"allocation_amount"     db:"allocation_amount"` // SUs allocated to the user, e.g., 100 CPU hours, 50 GPU hours, etc.
	StartTime           time.Time        `json:"start_time"            db:"start_time"`
	EndTime             time.Time        `json:"end_time"              db:"end_time"`
	MembershipStatus    AllocationStatus `json:"membership_status"     db:"membership_status"` // ACTIVE, INACTIVE, etc.
}
