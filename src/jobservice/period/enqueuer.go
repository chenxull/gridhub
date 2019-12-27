package period

import (
	"context"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/lcm"
	"github.com/gomodule/redigo/redis"
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
