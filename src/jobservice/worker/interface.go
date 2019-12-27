package worker

import "github.com/chenxull/goGridhub/gridhub/src/jobservice/job"

//Interface for worker
// like a driver to transparent the lower queue
type Interface interface {
	// start to serve
	Start() error

	//RegisterJobs multiple jobs
	RegisterJobs(jobs map[string]interface{}) error

	//Enqueue
	Enqueue(jobName string, params job.Parameters, isUnique bool, webHook string) (*job.Stats, error)
	Schedule(jobName string, params job.Parameters, runAfterSeconds uint64, isUnique bool, webHook string) (*job.Stats, error)
	PeriodicallyEnqueue(jobName string, params job.Parameters, cronSetting string, isUnique bool, webHook string) (*job.Stats, error)

	// Return the status info of the worker.
	Stats() (*Stats, error)

	// Check if the job has been already registered.
	IsKnownJob(name string) (interface{}, bool)

	// Validate the parameters of the known job
	ValidateJobParameters(jobType interface{}, params job.Parameters) error

	// Stop the job
	StopJob(jobID string) error

	RetryJob(jobID string) error
}
