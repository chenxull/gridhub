package core

import (
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/query"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/worker"
)

type Interface interface {
	LaunchJob(req *job.Request) (*job.Stats, error)
	GetJob(jobID string) (*job.Stats, error)
	StopJob(jobID string) error
	RetryJob(jobID string) error
	CheckStatus(stats *worker.Stats, err error)
	GetJobLogData(jobID string) ([]byte, error)
	// Get the periodic executions for the specified periodic job.
	GetPeriodicExecutions(periodicJobID string, query *query.Parameter) ([]*job.Stats, int64, error)
	GetJobs(query *query.Parameter) ([]*job.Stats, int64, error)
}
