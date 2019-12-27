package cworker

import (
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/env"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/lcm"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/period"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/worker"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var (
	workerPoolDeadTime = 10 * time.Second
)

const (
	workerPoolStatusHealthy      = "Healthy"
	workerPoolStatusDead         = "Dead"
	pingRedisMaxTimes            = 10
	defaultWorkerCount      uint = 10
)

// basicWorker is the worker implementation based on gocraft/work powered by redis.
type basicWorker struct {
	namespace string
	redisPool *redis.Pool
	pool      *work.WorkerPool
	enqueuer  *work.Enqueuer
	client    *work.Client
	context   *env.Context

	scheduler period.Scheduler
	ctl       lcm.Controller
	// key is name of known job
	// value is the type of known job
	knownJobs *sync.Map
}

// workerContext ...
// We did not use this context to pass context info so far, just a placeholder.
type workerContext struct{}

// log the job
func (rpc *workerContext) logJob(job *work.Job, next work.NextMiddlewareFunc) error {
	jobCopy := *job
	// as the args may contain sensitive information, ignore them when logging the detail
	jobCopy.Args = nil
	jobInfo, _ := utils.SerializeJob(&jobCopy)
	logger.Infof("Job incoming: %s", jobInfo)

	return next()
}

// NewWorker is constructor of worker
func NewWorker(ctx *env.Context, namespace string, workerCount uint, redisPool *redis.Pool, ctl lcm.Controller) worker.Interface {
	wc := defaultWorkerCount
	if workerCount > 0 {
		wc = workerCount
	}

	return &basicWorker{
		namespace: namespace,
		redisPool: redisPool,
		pool:      work.NewWorkerPool(workerContext{}, wc, namespace, redisPool),
		enqueuer:  work.NewEnqueuer(namespace, redisPool),
		client:    work.NewClient(namespace, redisPool),
		scheduler: period.NewScheduler(ctx.SystemContext, namespace, redisPool, ctl),
		ctl:       ctl,
		context:   ctx,
		knownJobs: new(sync.Map),
	}
}

// Start to serve
func (w *basicWorker) Start() error {
	if w.redisPool == nil {
		return errors.New("missing redis pool")
	}

	if utils.IsEmptyStr(w.namespace) {
		return errors.New("missing namespace")
	}

	if w.context == nil || w.context.SystemContext == nil {
		// report and exit
		return errors.New("missing context")
	}

	if w.ctl == nil {
		return errors.New("missing job life cycle controller")
	}

	// Test the redis connection
	if err := w.ping(); err != nil {
		return err
	}

	// Start the periodic scheduler
	w.context.WG.Add(1)
	go func() {
		defer func() {
			w.context.WG.Done()
		}()
		//Blocking call
		if err := w.scheduler.Start(); err != nil {
			w.context.ErrorChan <- err
		}
	}()
	// Listen to the system signal
	w.context.WG.Add(1)
	go func() {
		defer func() {
			w.context.WG.Done()
			logger.Infof("Basic worker is stopped")
		}()
		<-w.context.SystemContext.Done()
		if err := w.scheduler.Stop(); err != nil {
			logger.Errorf("stop scheduler error: %s", err)
		}
		w.pool.Stop()
	}()
	// Start the backend worker pool
	w.pool.Middleware((*workerContext).logJob)
	w.pool.Start()
	logger.Infof("redis worker is started")
	return nil
}

// RegisterJobs is used to register multiple jobs to worker.
func (w *basicWorker) RegisterJobs(jobs map[string]interface{}) error {
	if jobs == nil || len(jobs) == 0 {
		return nil
	}

	for name, j := range jobs {
		if err := w(name, j); err != nil {
			return err
		}
	}
}

func (w *basicWorker) Enqueue(jobName string, params job.Parameters, isUnique bool, webHook string) (*job.Stats, error) {
	panic("implement me")
}

func (w *basicWorker) Schedule(jobName string, params job.Parameters, runAfterSeconds uint64, isUnique bool, webHook string) (*job.Stats, error) {
	panic("implement me")
}

func (w *basicWorker) PeriodicallyEnqueue(jobName string, params job.Parameters, cronSetting string, isUnique bool, webHook string) (*job.Stats, error) {
	panic("implement me")
}

func (w *basicWorker) Stats() (*worker.Stats, error) {
	panic("implement me")
}

func (w *basicWorker) IsKnownJob(name string) (interface{}, bool) {
	panic("implement me")
}

func (w *basicWorker) ValidateJobParameters(jobType interface{}, params job.Parameters) error {
	panic("implement me")
}

func (w *basicWorker) StopJob(jobID string) error {
	panic("implement me")
}

func (w *basicWorker) RetryJob(jobID string) error {
	panic("implement me")
}

func (w *basicWorker) ping() error {
	conn := w.redisPool.Get()
	defer func() {
		_ = conn.Close()
	}()

	var err error
	for count := 1; count <= pingRedisMaxTimes; count++ {
		if _, err = conn.Do("ping"); err == nil {
			return nil
		}
		time.Sleep(time.Duration(count+4) * time.Second)
	}
	return fmt.Errorf("connect to redis server timeout: %s", err.Error())

}

// RegisterJob is used to register the job to the worker.
// j is the type of the job
func (w *basicWorker) registerJob(name string, j interface{}) (err error) {

}
