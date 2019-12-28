package runner

import (
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/env"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job/impl"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/lcm"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/period"
	"github.com/gocraft/work"
	"github.com/pkg/errors"
	"runtime"
	"time"
)

// RedisJob is a job wrapper to wrap the job.Interface to the style which can be recognized by the redis worker.
type RedisJob struct {
	job     interface{}
	context *env.Context
	ctl     lcm.Controller
}

func NewRedisJob(job interface{}, ctx *env.Context, ctl lcm.Controller) *RedisJob {
	return &RedisJob{
		job:     job,
		context: ctx,
		ctl:     ctl,
	}
}

// Run the job
func (rj *RedisJob) Run(j *work.Job) (err error) {
	var (
		runningJob  job.Interface
		execContext job.Context
		tracker     job.Tracker
		markStopped = bp(false)
	)

	// 根据标记情况和错误情况来记录
	defer func() {
		if !*markStopped {
			if err == nil {
				logger.Infof("|^_^| Job '%s:%s' exit with success", j.Name, j.ID)
			} else {
				// log error
				logger.Errorf("|@_@| Job '%s:%s' exit with error: %s", j.Name, j.ID, err)
			}
		}
	}()

	//Tracker the running job now 在任务执行的时候才开始追踪
	jID := j.ID

	if eID, yes := isPeriodicJobExecution(j); yes {
		jID = eID
	}
	if tracker, err = rj.ctl.Track(jID); err != nil {
		now := time.Now().Unix()
		if j.FailedAt == 0 || now-j.FailedAt < 2*24*3600 {
			j.Fails--
		}
		return
	}

	//Do operation based on the job status
	jStatus := job.Status(tracker.Job().Info.Status)
	switch jStatus {
	case job.PendingStatus, job.ScheduledStatus:
		//do nothing now
		break
	case job.StoppedStatus:
		markStopped = bp(true)
		return nil
	case job.ErrorStatus:
		if j.FailedAt > 0 && j.Fails > 0 {
			// Retry job
			// Reset job info
			if er := tracker.Reset(); er != nil {
				// Log error and return the original error if existing
				er = errors.Wrap(er, fmt.Sprintf("retrying job %s:%s failed", j.Name, j.ID))
				logger.Error(er)

				if len(j.LastErr) > 0 {
					return errors.New(j.LastErr)
				}
				return err
			}

			logger.Infof("|*_*| Retrying job %s:%s, revision: %d", j.Name, j.ID, tracker.Job().Info.Revision)
		}
		break
	default:
		return errors.Errorf("mismatch status for running job: expected <%s <> got %s", job.RunningStatus.String(), jStatus.String())
	}

	//Defer to switch status
	defer func() {
		// switch job status based on the returned error
		// The err happened here should not override the job run error, just log it.
		if err != nil {
			if er := tracker.Fail(); er != nil {
				logger.Errorf("Mark job status to fuliure error :%s", err)
			}
		}
		// Nil error might be returned by the stopped job. Check the latest status here.
		// If refresh latest status failed, let the process to go on to void missing status updating.
		if latest, er := tracker.Status(); er == nil {
			if latest == job.StoppedStatus {
				// Logged
				logger.Infof("Job %s:%s is stopped", tracker.Job().Info.JobName, tracker.Job().Info.JobID)
				// Stopped job, no exit message printing.
				markStopped = bp(true)
				return
			}
		}
		if er := tracker.Succeed(); er != nil {
			logger.Errorf("Mark job status to success error:%s", er)
		}
	}()

	// defer to handle runtime error
	defer func() {
		if r := recover(); r != nil {
			// Log the stack
			buf := make([]byte, 1<<10)
			size := runtime.Stack(buf, false)
			err = errors.Errorf("runtime error: %s; stack: %s", r, buf[0:size])
			logger.Errorf("Run job %s:%s error: %s", j.Name, j.ID, err)
		}
	}()

	//Build job Context
	if rj.context.JobContext == nil {
		rj.context.JobContext = impl.NewDefaultContext(rj.context.SystemContext)
	}
	if execContext, err = rj.context.JobContext.Build(tracker); err != nil {
		return
	}

	// Wrap job
	runningJob = Wrap(rj.job)
	//Set status to run
	if err = tracker.Run(); err != nil {
		return
	}
	//Run the job
	if err = runningJob.Run(execContext, j.Args); err != nil {
		return
	}
	//Handle retry
	rj.retry(runningJob, j)
	// Handle periodic job execution
	if _, yes := isPeriodicJobExecution(j); yes {
		if er := tracker.PeriodicExecutionDone(); er != nil {
			// Just log it
			logger.Error(er)
		}
	}
	return
}

func (rj *RedisJob) retry(j job.Interface, wj *work.Job) {
	if !j.ShouldRetry() {
		// Cancel retry immediately
		// Make it big enough to avoid retrying
		wj.Fails = 10000000000
		return
	}
}

func isPeriodicJobExecution(j *work.Job) (string, bool) {
	epoch, ok := j.Args[period.PeriodicExecutionMark]
	return fmt.Sprintf("%s@%s", j.ID, epoch), ok
}

func bp(b bool) *bool {
	return &b
}
