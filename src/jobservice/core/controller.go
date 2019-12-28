package core

import (
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/query"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/errs"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/mgt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/worker"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
)

// basicController implement the core interface and provides related job handle methods.
// basicController will coordinate the lower components to complete the process as a commander role.
type basicController struct {
	//Refer the backend worker
	backendWorker worker.Interface
	//Refer the job stats manager
	manager mgt.Manager
}

//NewController is constructor of basic
func NewController(backendWroker worker.Interface, mgr mgt.Manager) Interface {
	return &basicController{
		backendWorker: backendWroker,
		manager:       mgr,
	}
}

//LaunchJob is implementation of same method in core interface
func (bc *basicController) LaunchJob(req *job.Request) (res *job.Stats, err error) {
	if err := validJobReq(req); err != nil {
		return nil, errs.BadRequestError(err)
	}

	//Validate job name
	jobType, isKnowJob := bc.backendWorker.IsKnownJob(req.Job.Name)
	if !isKnowJob {
		return nil, errs.BadRequestError(errors.Errorf("job with name '%s' is unknown", req.Job.Name))
	}

	// Validate parameters
	if err := bc.backendWorker.ValidateJobParameters(jobType, req.Job.Parameters); err != nil {
		return nil, errs.BadRequestError(err)
	}
	//Enqueue job regarding of the kind
	switch req.Job.Metadata.JobKind {
	case job.KindScheduled:
		res, err = bc.backendWorker.Schedule(
			req.Job.Name,
			req.Job.Parameters,
			req.Job.Metadata.ScheduleDelay,
			req.Job.Metadata.IsUnique,
			req.Job.StatusHook,
		)
	case job.KindPeriodic:
		res, err = bc.backendWorker.PeriodicallyEnqueue(
			req.Job.Name,
			req.Job.Parameters,
			req.Job.Metadata.Cron,
			req.Job.Metadata.IsUnique,
			req.Job.StatusHook,
		)
	default:
		res, err = bc.backendWorker.Enqueue(
			req.Job.Name,
			req.Job.Parameters,
			req.Job.Metadata.IsUnique,
			req.Job.StatusHook,
		)
	}

	//Save job stats
	if err == nil {
		if err := bc.manager.SaveJob(res); err != nil {
			return nil, err
		}
	}
	return
}

func (bc *basicController) GetJob(jobID string) (*job.Stats, error) {
	if utils.IsEmptyStr(jobID) {
		return nil, errs.BadRequestError(errors.New("empty job ID"))
	}
	return bc.manager.GetJob(jobID)
}

func (bc *basicController) StopJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errs.BadRequestError(errors.New("empty job ID"))
	}

	return bc.backendWorker.StopJob(jobID)
}

func (bc *basicController) RetryJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errs.BadRequestError(errors.New("empty job ID"))
	}

	return bc.backendWorker.RetryJob(jobID)
}

// CheckStatus is implementation of same method in core interface.
func (bc *basicController) CheckStatus() (*worker.Stats, error) {
	return bc.backendWorker.Stats()
}

func (bc *basicController) GetJobLogData(jobID string) ([]byte, error) {
	if utils.IsEmptyStr(jobID) {
		return nil, errs.BadRequestError(errors.New("empty job ID"))
	}
	//
	//logData, err := logger.Retrieve(jobID)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return logData, nil
	panic("implement me")

}

func (bc *basicController) GetPeriodicExecutions(periodicJobID string, query *query.Parameter) ([]*job.Stats, int64, error) {
	if utils.IsEmptyStr(periodicJobID) {
		return nil, 0, errs.BadRequestError(errors.New("nil periodic job ID"))
	}

	return bc.manager.GetPeriodicExecution(periodicJobID, query)
}

func (bc *basicController) GetJobs(q *query.Parameter) ([]*job.Stats, int64, error) {
	onlyScheduledJobs := false
	if q != nil && q.Extras != nil {
		if v, ok := q.Extras.Get(query.ExtraParamKeyKind); ok {
			if job.KindScheduled == v.(string) {
				onlyScheduledJobs = true
			}
		}
	}

	if onlyScheduledJobs {
		return bc.manager.GetScheduledJobs(q)
	}

	return bc.manager.GetJobs(q)
}

func validJobReq(req *job.Request) error {
	if req == nil || req.Job == nil {
		return errors.New("empty job request is not allowed")
	}
	if utils.IsEmptyStr(req.Job.Name) {
		return errors.New("name of job must be specified")
	}

	if req.Job.Metadata == nil {
		return errors.New("metadata of job is missing")
	}

	if req.Job.Metadata.JobKind != job.KindGeneric &&
		req.Job.Metadata.JobKind != job.KindPeriodic &&
		req.Job.Metadata.JobKind != job.KindScheduled {
		return errors.Errorf(
			"job kind '%s' is not supported, only support '%s','%s','%s'",
			req.Job.Metadata.JobKind,
			job.KindGeneric,
			job.KindScheduled,
			job.KindPeriodic)
	}
	if req.Job.Metadata.JobKind == job.KindScheduled &&
		req.Job.Metadata.ScheduleDelay == 0 {
		return errors.Errorf("'schedule_delay' must be specified for %s job", job.KindScheduled)
	}

	if req.Job.Metadata.JobKind == job.KindPeriodic {
		if utils.IsEmptyStr(req.Job.Metadata.Cron) {
			return fmt.Errorf("'cron_spec' must be specified for the %s job", job.KindPeriodic)
		}

		if _, err := cron.Parse(req.Job.Metadata.Cron); err != nil {
			return fmt.Errorf("'cron_spec' is not correctly set: %s: %s", req.Job.Metadata.Cron, err)
		}
	}

	return nil
}
