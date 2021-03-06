package job

import (
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/pkg/errors"
)

// Parameters for job execution.
type Parameters map[string]interface{}

type Request struct {
	Job *RequestBody `json:"job"`
}

type RequestBody struct {
	Name       string     `json:"name"`
	Parameters Parameters `json:"parameters"`
	Metadata   *Metadata  `json:"metadata"`
	StatusHook string     `json:"status_hook"`
}

// Metadata stores the metadata of job.
type Metadata struct {
	JobKind       string `json:"kind"`
	ScheduleDelay uint64 `json:"schedule_delay,omitempty"`
	Cron          string `json:"cron_spec,omitempty"`
	IsUnique      bool   `json:"unique"`
}

// Stats keeps the result of job launching.
type Stats struct {
	Info *StatsInfo `json:"job"`
}

type StatsInfo struct {
	JobID         string     `json:"id"`
	Status        string     `json:"status"`
	JobName       string     `json:"name"`
	JobKind       string     `json:"kind"`
	IsUnique      bool       `json:"unique"`
	RefLink       string     `json:"ref_link,omitempty"`
	CronSpec      string     `json:"cron_spec,omitempty"`
	EnqueueTime   int64      `json:"enqueue_time"`
	UpdateTime    int64      `json:"update_time"`
	RunAt         int64      `json:"run_at,omitempty"`
	CheckIn       string     `json:"check_in,omitempty"`
	CheckInAt     int64      `json:"check_in_at,omitempty"`
	DieAt         int64      `json:"die_at,omitempty"`
	WebHookURL    string     `json:"web_hook_url,omitempty"`
	UpstreamJobID string     `json:"upstream_job_id,omitempty"`   // Ref the upstream job if existing
	NumericPID    int64      `json:"numeric_policy_id,omitempty"` // The numeric policy ID of the periodic job
	Parameters    Parameters `json:"parameters,omitempty"`
	Revision      int64      `json:"revision,omitempty"` // For differentiating the each retry of the same job
}

// ActionRequest defines for triggering job action like stop/cancel.
type ActionRequest struct {
	Action string `json:"action"`
}

// StatusChange is designed for reporting the status change via hook.
type StatusChange struct {
	JobID    string     `json:"job_id"`
	Status   string     `json:"status"`
	CheckIn  string     `json:"check_in,omitempty"`
	Metadata *StatsInfo `json:"metadata,omitempty"`
}

// SimpleStatusChange only keeps job ID and the target status
type SimpleStatusChange struct {
	JobID        string `json:"job_id"`
	TargetStatus string `json:"target_status"`
}

func (st *Stats) Validate() error {
	if st.Info == nil {
		return errors.New("nil stats body")
	}

	if utils.IsEmptyStr(st.Info.JobID) {
		return errors.New("missing job ID in job stats")
	}

	if utils.IsEmptyStr(st.Info.JobName) {
		return errors.New("missing job name in job stats")
	}

	if utils.IsEmptyStr(st.Info.JobKind) {
		return errors.New("missing job name in job stats")
	}

	if st.Info.JobKind != KindGeneric &&
		st.Info.JobKind != KindPeriodic &&
		st.Info.JobKind != KindScheduled {
		return errors.Errorf("job kind is not supported: %s", st.Info.JobKind)
	}

	status := Status(st.Info.Status)
	if err := status.Validate(); err != nil {
		return err
	}

	if st.Info.JobKind == KindPeriodic {
		if utils.IsEmptyStr(st.Info.CronSpec) {
			return errors.New("missing cron spec for periodic job")
		}
	}

	if st.Info.JobKind == KindScheduled {
		if st.Info.RunAt == 0 {
			return errors.New("enqueue timestamp missing for scheduled job")
		}
	}

	return nil
}
