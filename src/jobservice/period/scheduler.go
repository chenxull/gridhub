package period

// Scheduler defines operations the periodic scheduler should have.
type Scheduler interface {
	// Start to serve periodic job scheduling process
	Start() error
	Stop() error
	// Schedule the specified cron job policy.
	Schedule(policy *Policy) (int64, error)

	UnSchedule(policyID string) error
}
