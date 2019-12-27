package period

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"sync"
)

const (
	// changeEventSchedule : Schedule periodic job policy event
	changeEventSchedule = "Schedule"
	// changeEventUnSchedule : UnSchedule periodic job policy event
	changeEventUnSchedule = "UnSchedule"
)

// Policy ...
type Policy struct {
	// Policy can be treated as job template of periodic job.
	// The info of policy will be copied into the scheduled job executions for the periodic job.
	ID            string                 `json:"id"`
	JobName       string                 `json:"job_name"`
	CronSpec      string                 `json:"cron_spec"`
	JobParameters map[string]interface{} `json:"job_params,omitempty"`
	WebHookURL    string                 `json:"web_hook_url,omitempty"`
}

// policyStore is in-memory cache for the periodic job policies.
type policyStore struct {
	// k-v pair and key is the policy ID
	hash      *sync.Map
	namespace string
	context   context.Context
	pool      *redis.Pool
	// For stop
	stopChan chan bool
}

// newPolicyStore is constructor of policyStore
func newPolicyStore(ctx context.Context, ns string, pool *redis.Pool) *policyStore {
	return &policyStore{
		hash:      new(sync.Map),
		context:   ctx,
		namespace: ns,
		pool:      pool,
		stopChan:  make(chan bool, 1),
	}
}

// message is designed for sub/pub messages
type message struct {
	Event string  `json:"event"`
	Data  *Policy `json:"data"`
}
