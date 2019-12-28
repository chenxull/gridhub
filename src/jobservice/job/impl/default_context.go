package impl

import (
	"context"
	"errors"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
)

// DefaultContext provides a basic job context
type DefaultContext struct {
	// System context
	sysContext context.Context
	// Logger for job
	logger logger.Interface
	// Other required information
	properties map[string]interface{}
	// Track the job attached with the context
	tracker job.Tracker
}

// NewDefaultContext is constructor of building DefaultContext
func NewDefaultContext(sysCtx context.Context) job.Context {
	return &DefaultContext{
		sysContext: sysCtx,
		properties: make(map[string]interface{}),
	}
}

// Build implements the same method in env.Context interface
// This func will build the job execution context before running
func (dc *DefaultContext) Build(t job.Tracker) (job.Context, error) {
	if t == nil {
		return nil, errors.New("nil job tracker")
	}

	jContext := &DefaultContext{
		sysContext: dc.sysContext,
		tracker:    t,
		properties: make(map[string]interface{}),
	}

	// Copy properties
	if len(dc.properties) > 0 {
		for k, v := range dc.properties {
			jContext.properties[k] = v
		}
	}

	// Set loggers for job
	//lg, err := createLoggers(t.Job().Info.JobID)
	//if err != nil {
	//	return nil, err
	//}

	//jContext.logger = lg

	return jContext, nil
}

// Get implements the same method in env.Context interface
func (dc *DefaultContext) Get(prop string) (interface{}, bool) {
	v, ok := dc.properties[prop]
	return v, ok
}

// SystemContext implements the same method in env.Context interface
func (dc *DefaultContext) SystemContext() context.Context {
	return dc.sysContext
}

// Checkin is bridge func for reporting detailed status
func (dc *DefaultContext) Checkin(status string) error {
	return dc.tracker.CheckIn(status)
}

// OPCommand return the control operational command like stop if have
func (dc *DefaultContext) OPCommand() (job.OPCommand, bool) {
	latest, err := dc.tracker.Status()
	if err != nil {
		return job.NilCommand, false
	}

	if job.StoppedStatus == latest {
		return job.StopCommand, true
	}

	return job.NilCommand, false
}

// GetLogger returns the logger
func (dc *DefaultContext) GetLogger() logger.Interface {
	return dc.logger
}

// Tracker returns the tracker tracking the job attached with the context
func (dc *DefaultContext) Tracker() job.Tracker {
	return dc.tracker
}
