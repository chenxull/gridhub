package period

import (
	"context"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/lcm"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

// basicScheduler manages the periodic scheduling policies.
type basicScheduler struct {
	context   context.Context
	pool      *redis.Pool
	namespace string
	enqueuer  *enqueuer
	client    *work.Client
	ctl       lcm.Controller
}

// NewScheduler is constructor of basicScheduler
func NewScheduler(ctx context.Context, namespace string, pool *redis.Pool, ctl lcm.Controller) Scheduler {
	return &basicScheduler{
		context:   ctx,
		pool:      pool,
		namespace: namespace,
		enqueuer:  newEnqueuer(ctx, namespace, pool, ctl),
		client:    work.NewClient(namespace, pool),
		ctl:       ctl,
	}
}

func (bs *basicScheduler) Start() error {
	panic("implement me")
}

func (bs *basicScheduler) Stop() error {
	panic("implement me")
}

func (bs *basicScheduler) Schedule(policy *Policy) (int64, error) {
	panic("implement me")
}

func (bs *basicScheduler) UnSchedule(policyID string) error {
	panic("implement me")
}
