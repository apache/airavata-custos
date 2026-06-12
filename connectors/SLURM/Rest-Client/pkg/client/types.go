// cli/internal/client/types.go
package client

import "encoding/json"

type ErrorResponse struct {
	Errors []struct {
		Description string `json:"description"`
		ErrorNumber int    `json:"error_number"`
		Error       string `json:"error"`
		Source      string `json:"source"`
	} `json:"errors"`
	Warnings []struct {
		Description string `json:"description"`
		Source      string `json:"source"`
	} `json:"warnings"`
}

type TRES struct {
	Type  string `json:"type"`
	Name  string `json:"name,omitempty"`
	Count int64  `json:"count"`
}

type Account struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Organization string `json:"organization,omitempty"`
}

type AssocLimits struct {
	GrpJobs     *int64 `json:"grp_jobs,omitempty"`
	GrpTRES     []TRES `json:"grp_tres,omitempty"`
	GrpTRESMins []TRES `json:"grp_tres_mins,omitempty"`
	MaxWallPM   *int64 `json:"max_wall_pj,omitempty"`
}

type Association struct {
	Account       string `json:"account"`
	Cluster       string `json:"cluster"`
	User          string `json:"user"`
	Partition     string `json:"partition,omitempty"`
	ParentAccount string `json:"parent_account,omitempty"`
	IsDefault     *bool  `json:"is_default,omitempty"`
	ID            int64  `json:"id_association,omitempty"`
	// QoS is the list of QoS names permitted for this association
	// (e.g. ["normal"]). Marshaled to the top-level `qos` field on the wire.
	QoS []string `json:"-"`
	// Limits is a logical grouping — slurmrestd v0.0.41 actually encodes limits
	// in a nested `max` object per-association. We translate between the two
	// shapes in Marshal/UnmarshalJSON below.
	Limits AssocLimits `json:"-"`
}

// SlurmNumber matches slurmrestd's {set, infinite, number} triple used for
// all scalar limit values in the v0.0.41 accounting schema.
type SlurmNumber struct {
	Set      bool  `json:"set"`
	Infinite bool  `json:"infinite"`
	Number   int64 `json:"number"`
}

func numPtr(n *int64) *SlurmNumber {
	if n == nil {
		return nil
	}
	return &SlurmNumber{Set: true, Number: *n}
}

func ptrNum(n *SlurmNumber) *int64 {
	if n == nil || !n.Set || n.Infinite {
		return nil
	}
	v := n.Number
	return &v
}

// assocMax is the v0.0.41 "max" sub-object inside each association. We only
// populate the fields we actually manage; slurmrestd ignores unset sub-objects.
type assocMax struct {
	Jobs *assocMaxJobs `json:"jobs,omitempty"`
	TRES *assocMaxTRES `json:"tres,omitempty"`
}

type assocMaxJobs struct {
	Per *assocMaxJobsPer `json:"per,omitempty"`
}

type assocMaxJobsPer struct {
	Count     *SlurmNumber `json:"count,omitempty"`      // GrpJobs
	WallClock *SlurmNumber `json:"wall_clock,omitempty"` // MaxWallDurationPerJob (seconds)
}

type assocMaxTRES struct {
	Total []TRES          `json:"total,omitempty"` // GrpTRES
	Group *assocMaxTRESGp `json:"group,omitempty"`
}

type assocMaxTRESGp struct {
	Minutes []TRES `json:"minutes,omitempty"` // GrpTRESMins
}

// assocWire is the on-the-wire association record used for both request
// marshaling and response unmarshaling. `user` is emitted even when empty
// because slurmrestd rejects payloads without it (error 9200).
type assocWire struct {
	Account       string    `json:"account"`
	Cluster       string    `json:"cluster"`
	User          string    `json:"user"`
	Partition     string    `json:"partition,omitempty"`
	ParentAccount string    `json:"parent_account,omitempty"`
	IsDefault     *bool     `json:"is_default,omitempty"`
	ID            int64     `json:"id_association,omitempty"`
	QoS           []string  `json:"qos,omitempty"`
	Max           *assocMax `json:"max,omitempty"`
}

func (a Association) MarshalJSON() ([]byte, error) {
	w := assocWire{
		Account:       a.Account,
		Cluster:       a.Cluster,
		User:          a.User,
		Partition:     a.Partition,
		ParentAccount: a.ParentAccount,
		IsDefault:     a.IsDefault,
		ID:            a.ID,
		QoS:           a.QoS,
	}
	m := &assocMax{}
	touched := false
	if a.Limits.GrpJobs != nil || a.Limits.MaxWallPM != nil {
		per := &assocMaxJobsPer{}
		if a.Limits.GrpJobs != nil {
			per.Count = numPtr(a.Limits.GrpJobs)
		}
		if a.Limits.MaxWallPM != nil {
			per.WallClock = numPtr(a.Limits.MaxWallPM)
		}
		m.Jobs = &assocMaxJobs{Per: per}
		touched = true
	}
	if len(a.Limits.GrpTRES) > 0 || len(a.Limits.GrpTRESMins) > 0 {
		t := &assocMaxTRES{}
		if len(a.Limits.GrpTRES) > 0 {
			t.Total = a.Limits.GrpTRES
		}
		if len(a.Limits.GrpTRESMins) > 0 {
			t.Group = &assocMaxTRESGp{Minutes: a.Limits.GrpTRESMins}
		}
		m.TRES = t
		touched = true
	}
	if touched {
		w.Max = m
	}
	return json.Marshal(w)
}

func (a *Association) UnmarshalJSON(data []byte) error {
	var w assocWire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	a.Account = w.Account
	a.Cluster = w.Cluster
	a.User = w.User
	a.Partition = w.Partition
	a.ParentAccount = w.ParentAccount
	a.IsDefault = w.IsDefault
	a.ID = w.ID
	a.QoS = w.QoS
	a.Limits = AssocLimits{}
	if w.Max != nil {
		if w.Max.Jobs != nil && w.Max.Jobs.Per != nil {
			a.Limits.GrpJobs = ptrNum(w.Max.Jobs.Per.Count)
			a.Limits.MaxWallPM = ptrNum(w.Max.Jobs.Per.WallClock)
		}
		if w.Max.TRES != nil {
			a.Limits.GrpTRES = w.Max.TRES.Total
			if w.Max.TRES.Group != nil {
				a.Limits.GrpTRESMins = w.Max.TRES.Group.Minutes
			}
		}
	}
	return nil
}

type JobTime struct {
	Elapsed    int64 `json:"elapsed"`
	Eligible   int64 `json:"eligible"`
	End        int64 `json:"end"`
	Start      int64 `json:"start"`
	Submission int64 `json:"submission"`
	Suspended  int64 `json:"suspended"`
}

type JobTresInfo struct {
	Allocated []TRES `json:"allocated,omitempty"`
	Requested []TRES `json:"requested,omitempty"`
}

type JobExitInfo struct {
	Status     []string    `json:"status"`
	ReturnCode SlurmNumber `json:"return_code"`
}

type JobInfo struct {
	Account         string      `json:"account"`
	Cluster         string      `json:"cluster"`
	Time            JobTime     `json:"time"`
	JobID           int64       `json:"job_id"`
	Name            string      `json:"name"`
	Partition       string      `json:"partition"`
	QoS             string      `json:"qos"`
	User            string      `json:"user"`
	Nodes           string      `json:"nodes"`
	Tres            JobTresInfo `json:"tres"`
	ExitCode        JobExitInfo `json:"exit_code"`
	DerivedExitCode JobExitInfo `json:"derived_exit_code"`
}

type JobSubmitParam struct {
	Account           string      `json:"account"`
	Partition         string      `json:"partition,omitempty"`
	QoS               string      `json:"qos,omitempty"`
	Name              string      `json:"name,omitempty"`
	Tasks             int64       `json:"tasks,omitempty"`
	CurrentWorkingDir string      `json:"current_working_directory,omitempty"`
	Environment       []string    `json:"environment,omitempty"`
	CpusPerTask       int64       `json:"cpus_per_task,omitempty"`
	Memory            int64       `json:"memory,omitempty"`
	TimeLimit         SlurmNumber `json:"time_limit,omitempty"` // seconds
}

type JobSubmitRequest struct {
	JobSubmitParam JobSubmitParam `json:"job"`
	Script         string         `json:"script"`
}

type JobSubmitResponse struct {
	JobID            int64  `json:"job_id"`
	StepID           string `json:"step_id"`
	JobSubmitUserMsg string `json:"job_submit_user_msg"`
}
