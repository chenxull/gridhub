package lcm

import (
	"context"
	"encoding/json"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/rds"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/env"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	// Waiting a short while if any errors occurred
	shortLoopInterval = 5 * time.Second
	// Waiting for long while if no retrying elements found
	longLoopInterval = 5 * time.Minute
)

// Controller is designed to control the life cycle of the job
type Controller interface {
	//Run daemon process if needed
	Serve() error
	// New tracker from the new provided stats
	New(stats *job.Stats) (job.Tracker, error)

	//Tracker the life cycle of the specified existing job
	Track(jobId string) (job.Tracker, error)
}

// basicController is default implementation of Controller based on redis
type basicController struct {
	context   context.Context
	namespace string
	pool      *redis.Pool
	callback  job.HookCallback
	wg        *sync.WaitGroup
}

func NewController(ctx *env.Context, ns string, pool *redis.Pool, callback job.HookCallback) Controller {
	return &basicController{
		context:   ctx.SystemContext,
		namespace: ns,
		pool:      pool,
		callback:  callback,
		wg:        ctx.WG,
	}
}

func (bc *basicController) Serve() error {
	bc.wg.Add(1)
	go bc.loopForRestoreDeadStatus()

	logger.Info("Status restoring loop is started")

	return nil
}

func (bc *basicController) New(stats *job.Stats) (job.Tracker, error) {
	if stats == nil {
		return nil, errors.New("nil stats when creating job tracker")
	}

	if err := stats.Validate(); err != nil {
		return nil, errors.Errorf("error occurred when creating job tracker: %s", err)
	}
	bt := job.NewBasicTrackerWithStats(bc.context, stats, bc.namespace, bc.pool, bc.callback)
	// 将 job 的数据存储到 redis 中
	if err := bt.Save(); err != nil {
		return nil, err
	}
	return bt, nil

}

// Track and attache with the job,获取当前job的stats
func (bc *basicController) Track(jobID string) (job.Tracker, error) {
	bt := job.NewBasicTrackerWithID(bc.context, jobID, bc.namespace, bc.pool, bc.callback)
	if err := bt.Load(); err != nil {
		return nil, err
	}

	return bt, nil
}

// loopForRestoreDeadStatus is a loop to restore the dead states of jobs
func (bc *basicController) loopForRestoreDeadStatus() {
	defer func() {
		logger.Info("Status restoring loop is stopped")
		bc.wg.Done()
	}()

	token := make(chan bool, 1)
	token <- true
	for {
		<-token
		if err := bc.restoreDeadStatus(); err != nil {
			waitInterval := shortLoopInterval
			if err == rds.ErrNoElements {
				// No elements
				waitInterval = longLoopInterval
			} else {
				logger.Errorf("restore dead status error: %s, put it back to the retrying Q later again", err)
			}
			// wait for a while or be terminated
			// wait for a while or be terminated
			select {
			case <-time.After(waitInterval):
			case <-bc.context.Done():
				return
			}
		}
		token <- true
	}
}

// restoreDeadStatus try to restore the dead status
// 恢复死去 job 的状态，并对其进行标记
func (bc *basicController) restoreDeadStatus() error {
	//Get one
	deadOne, err := bc.popOneDead()
	if err != nil {
		return err
	}
	// Try to update status,t 中保存着job信息
	t, err := bc.Track(deadOne.JobID)
	if err != nil {
		return err
	}
	return t.UpdateStatusWithRetry(job.Status(deadOne.TargetStatus))

}

// popOneDead retrieves one dead status from the backend Q from lowest to highest
func (bc *basicController) popOneDead() (*job.SimpleStatusChange, error) {
	//从 redis中获取 dead job 的数据
	conn := bc.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	key := rds.KeyStatusUpdateRetryQueue(bc.namespace)
	v, err := rds.ZPopMin(conn, key)
	if err != nil {
		return nil, err
	}
	if bytes, ok := v.([]byte); ok {
		ssc := &job.SimpleStatusChange{}
		if err := json.Unmarshal(bytes, ssc); err == nil {
			return ssc, nil
		}
	}
	return nil, errors.New("pop one dead error: bad result reply")

}
