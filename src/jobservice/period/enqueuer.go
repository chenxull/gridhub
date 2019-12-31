package period

import (
	"context"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/rds"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/lcm"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron"
	"math/rand"
	"time"
)

const (
	enqueuerSleep   = 2 * time.Minute
	enqueuerHorizon = 4 * time.Minute
	neverExecuted   = 365 * 24 * time.Hour

	// PeriodicExecutionMark marks the scheduled job to a periodic execution
	PeriodicExecutionMark = "_job_kind_periodic_"
)

type enqueuer struct {
	namespace   string
	context     context.Context
	pool        *redis.Pool
	policyStore *policyStore
	ctl         lcm.Controller
	// Diff with other nodes
	nodeID string
	// Track the error of enqueuing
	lastEnqueueErr error
	// For stop
	stopChan chan bool
}

func newEnqueuer(ctx context.Context, namespace string, pool *redis.Pool, ctl lcm.Controller) *enqueuer {
	nodeID := ctx.Value(utils.NodeID)
	if nodeID == nil {
		// Must be failed
		panic("missing node ID in the system context of periodic enqueuer")
	}

	return &enqueuer{
		context:     ctx,
		namespace:   namespace,
		pool:        pool,
		policyStore: newPolicyStore(ctx, namespace, pool),
		ctl:         ctl,
		stopChan:    make(chan bool, 1),
		nodeID:      nodeID.(string),
	}
}

//Blocking call
func (e *enqueuer) start() error {
	//Load policyStore when starting
	if err := e.policyStore.load(); err != nil {
		return err
	}

	go e.loop()
	logger.Info("Periodic enqueuer is started")

	return e.policyStore.serve()
}

func (e *enqueuer) loop() {
	defer func() {
		logger.Info("Periodic enqueuer is stopped")
	}()
	isHit := e.checkAndEnqueue()

	//Begin reaping periodically
	timer := time.NewTimer(e.nextTurn(isHit, e.lastEnqueueErr != nil))
	defer timer.Stop()

	for {
		select {
		case <-e.stopChan:
			e.policyStore.stopChan <- true
			return
		case <-timer.C:
			timer.Reset(neverExecuted)

			//Check and enqueue
			isHit = e.checkAndEnqueue()
			timer.Reset(e.nextTurn(isHit, e.lastEnqueueErr != nil))
		}
	}

}

// checkAndEnqueue checks if it should do enqueue and
func (e *enqueuer) checkAndEnqueue() (isHit bool) {
	if isHit = e.shouldEnqueue(); isHit {
		e.enqueue()
	}
	return
}

// nextTurn returns the next check time slot by applying
// priorities to balance the workloads across multiple nodes
func (e *enqueuer) nextTurn(isHit bool, enqErr bool) time.Duration {
	base := enqueuerSleep
	if isHit {
		// Down the hit priority by adding more waiting time
		base = base + time.Duration(3)*time.Second
		if enqErr {
			// Downgrade the priority if the node has occurred error when enqueuing
			base = base + time.Duration(5)*time.Second
		}
	} else {
		// Upgrade the priority of hitting in the next turn
		base = base - time.Duration(3)*time.Second
	}

	// Add random waiting time [0,8)
	base = base + time.Duration(rand.Intn(5))*time.Second

	return base
}

func (e *enqueuer) enqueue() {
	conn := e.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	//Reset error track
	e.policyStore.Iterate(func(id string, p *Policy) bool {
		e.scheduleNextJobs(p, conn)
		return true
	})
}

// scheduleNextJobs schedules job for next time slots based on the policy
func (e *enqueuer) scheduleNextJobs(p *Policy, conn redis.Conn) {
	nowTime := time.Unix(time.Now().Unix(), 0)
	horizon := nowTime.Add(enqueuerHorizon)
	schedule, err := cron.Parse(p.CronSpec)
	if err != nil {
		e.lastEnqueueErr = err
		logger.Errorf("Invalid corn spec in periodic policy %s %s: %s", p.JobName, p.ID, err)
	} else {
		for t := schedule.Next(nowTime); t.Before(horizon); t = schedule.Next(t) {
			epoch := t.Unix()

			// Clone parameters
			// Add extra argument for job running too.
			// Notes: Only for system using
			wJobParams := cloneParameters(p.JobParameters, epoch)
			// Create an execution (job) based on the periodic job template (policy)
			j := &work.Job{
				Name:       p.JobName,
				ID:         p.ID,
				EnqueuedAt: epoch,
				Args:       wJobParams,
			}
			rawJSON, err := utils.SerializeJob(j)
			if err != nil {
				e.lastEnqueueErr = err
				// Actually this error should not happen if the object struct is well defined
				logger.Errorf("Serialize job object for periodic job %s error: %s", p.ID, err)
				break
			}
			execution := e.createExecution(p, epoch)
			eTracker, err := e.ctl.New(execution)
			if err != nil {
				e.lastEnqueueErr = err
				logger.Errorf("Save stats data of job execution '%s' error: %s", execution.Info.JobID, err)
				break
			}

			//Put job to the scheduled job queue
			_, err = conn.Do("ZADD", rds.RedisKeyScheduled(e.namespace), epoch, rawJSON)
			if err != nil {
				e.lastEnqueueErr = err
				logger.Errorf("Put the execution of the periodic job '%s' to the scheduled job queue error: %s", p.ID, err)

				// Mark job status to be error
				// If this happened, the job stats is definitely becoming dirty data at job service side.
				// For the consumer side, the retrying of web hook may fix the problem.
				if err := eTracker.Fail(); err != nil {
					e.lastEnqueueErr = err
					logger.Errorf("Mark execution '%s' to failure status error: %s", execution.Info.JobID, err)
				}

				break // Probably redis connection is broken
			}

			logger.Debugf("Scheduled execution for periodic job %s:%s at %d", j.Name, p.ID, epoch)

		}
	}
}

// createExecution creates execution object
func (e *enqueuer) createExecution(p *Policy, runAt int64) *job.Stats {
	eID := fmt.Sprintf("%s@%d", p.ID, runAt)

	return &job.Stats{
		Info: &job.StatsInfo{
			JobID:         eID,
			JobName:       p.JobName,
			WebHookURL:    p.WebHookURL,
			CronSpec:      p.CronSpec,
			UpstreamJobID: p.ID,
			RunAt:         runAt,
			Status:        job.ScheduledStatus.String(),
			JobKind:       job.KindScheduled, // For periodic job execution, it should be set to 'scheduled'
			EnqueueTime:   time.Now().Unix(),
			RefLink:       fmt.Sprintf("/api/v1/jobs/%s", eID),
			Parameters:    p.JobParameters,
		},
	}
}
func (e *enqueuer) shouldEnqueue() bool {
	conn := e.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	// Acquired a lock before doing checking
	lockKey := rds.KeyPeriodicLock(e.namespace)
	if err := rds.AcquireLock(conn, lockKey, e.nodeID, 30); err != nil {
		logger.Errorf("acquire lock for periodic enqueuing error: %s", err)
		return false
	}
	//Acquired lock
	defer func() {
		if err := rds.ReleaseLock(conn, lockKey, e.nodeID); err != nil {
			logger.Errorf("release lock for periodic enqueuing error: %s", err)
		}
	}()
	shouldEnq := false
	lastEnqueue, err := redis.Int64(conn.Do("GET", rds.RedisKeyLastPeriodicEnqueue(e.namespace)))
	if err != nil {
		if err.Error() != redis.ErrNil.Error() {
			// Logged error
			logger.Errorf("get timestamp of last enqueue error: %s", err)
		}

		// Should enqueue
		shouldEnq = true
	} else {
		// Check further condition
		shouldEnq = lastEnqueue < (time.Now().Unix() - int64(enqueuerSleep/time.Minute)*60)
	}
	if shouldEnq {
		// Set last periodic enqueue timestamp
		if _, err := conn.Do("SET", rds.RedisKeyLastPeriodicEnqueue(e.namespace), time.Now().Unix()); err != nil {
			logger.Errorf("set last periodic enqueue timestamp error: %s", err)
		}

		// Anyway the action should be enforced
		// The negative effect of this failure is just more re-enqueues by other nodes
		return true
	}

	return false
}

func cloneParameters(params job.Parameters, epoch int64) job.Parameters {
	p := make(job.Parameters)

	for k, v := range params {
		p[k] = v
	}
	p[PeriodicExecutionMark] = fmt.Sprintf("%d", epoch)

	return p
}
