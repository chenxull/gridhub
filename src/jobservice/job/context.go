package job

import "context"

// Context is combination of BaseContext and other job specified resources.
// Context will be the real execution context for one job.
type Context interface {
	// Build the context based on the parent context
	// A new job context will be generated based on the current context
	// for the provided job.
	Build(tracker Tracker) (Context, error)

	// Get property from the context
	Get(prop string) (interface{}, bool)

	// SystemContext returns the system context
	SystemContext() context.Context

	// Checkin is bridge func for reporting detailed status
	Checkin(status string) error

	// OPCommand return the control operational command like stop if have
	OPCommand() (OPCommand, bool)
	// Return the logger
	//todo GetLogger() logger.Interface

	// Get tracker
	Tracker() Tracker
}

// 这里只是定义了，具体初始化做了什么事不管，只需要返回值符合要求即可
type ContextInitializer func(ctx context.Context) (Context, error)
