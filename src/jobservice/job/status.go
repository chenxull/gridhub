package job

import "fmt"

const (
	// PendingStatus   : job status pending
	PendingStatus Status = "Pending"
	// RunningStatus   : job status running
	RunningStatus Status = "Running"
	// StoppedStatus   : job status stopped
	StoppedStatus Status = "Stopped"
	// ErrorStatus     : job status error
	ErrorStatus Status = "Error"
	// SuccessStatus   : job status success
	SuccessStatus Status = "Success"
	// ScheduledStatus : job status scheduled
	ScheduledStatus Status = "Scheduled"
)

// Status of job
type Status string

func (s Status) Validate() error {
	if s.Code() == -1 {
		return fmt.Errorf("%s is not valid job status", s)
	}

	return nil
}

//Code of job status
func (s Status) Code() int {
	switch s {
	case "Pending":
		return 0
	case "Scheduled":
		return 1
	case "Running":
		return 2
	case "Stopped":
		return 3
	case "Error":
		return 3
	case "Success":
		return 3
	default:
	}
	return -1
}

// Compare the two job status
// if < 0, s before another status
// if == 0, same status
// if > 0, s after another status
func (s Status) Compare(another Status) int {
	return s.Code() - another.Code()
}

func (s Status) String() string {
	return string(s)
}

// Final returns if the status is final status
// e.g: "Stopped", "Error" or "Success"
func (s Status) Final() bool {
	return s.Code() == 3
}
