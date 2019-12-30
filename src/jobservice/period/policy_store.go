package period

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/rds"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron"
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

// Serialize the policy to raw data.
func (p *Policy) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// DeSerialize the raw json to policy.
func (p *Policy) DeSerialize(rawJSON []byte) error {
	return json.Unmarshal(rawJSON, p)
}

// Validate the policy
func (p *Policy) Validate() error {
	if utils.IsEmptyStr(p.ID) {
		return errors.New("missing ID in the periodic job policy object")
	}

	if utils.IsEmptyStr(p.JobName) {
		return errors.New("missing job name in the periodic job policy object")
	}

	if !utils.IsEmptyStr(p.WebHookURL) {
		if !utils.IsValidURL(p.WebHookURL) {
			return fmt.Errorf("bad web hook URL: %s", p.WebHookURL)
		}
	}

	if _, err := cron.Parse(p.CronSpec); err != nil {
		return err
	}

	return nil
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

// message is designed for sub/pub messages
type message struct {
	Event string  `json:"event"`
	Data  *Policy `json:"data"`
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

//Blocking call
func (ps *policyStore) serve() (err error) {
	defer func() {
		logger.Info("Periodical job policy store is stopped")
	}()
	conn := ps.pool.Get()
	psc := redis.PubSubConn{Conn: conn}

	defer func() {
		_ = psc.Close()
	}()

	//Subscribe channel
	err = psc.Subscribe(redis.Args{}.AddFlat(rds.KeyPeriodicNotification(ps.namespace))...)
	if err != nil {
		return
	}

	//Channels for sub/pub ctl
	errChan := make(chan error, 1)
	done := make(chan bool, 1)

	go func() {
		for {
			switch res := psc.Receive().(type) {
			case error:
				errChan <- fmt.Errorf("redis sub/pub chan error: %s", res.(error).Error())
				break
			case redis.Message:
				m := &message{}
				if err := json.Unmarshal(res.Data, m); err != nil {
					// logged
					logger.Errorf("Read invalid message: %s\n", res.Data)
					break
				}
				if err := ps.sync(m); err != nil {
					logger.Error(err)
				}
				break
			case redis.Subscription:
				switch res.Kind {
				case "subscribe":
					logger.Infof("Subscribe redis channel %s", res.Channel)
					break
				case "unsubscribe":
					// Unsubscribe all, means main goroutine is exiting
					logger.Infof("Unsubscribe redis channel %s", res.Channel)
					done <- true
					return
				}
			}
		}
	}()
	logger.Info("Periodical job policy store is serving with policy auto sync enabled")

}

//将redis中的数据同步到本机内存中
func (ps *policyStore) sync(m *message) error {
	if m == nil {
		return errors.New("nil message")
	}
	if m.Data == nil {
		return errors.New("missing data in the policy sync message")
	}
	switch m.Event {
	case changeEventSchedule:
		if err := ps.add(m.Data); err != nil {
			return fmt.Errorf("failed to sync scheduled policy %s: %s", m.Data.ID, err)
		}
	case changeEventUnSchedule:
		removed := ps.remove(m.Data.ID)
		if removed == nil {
			return fmt.Errorf("failed to sync unscheduled policy %s", m.Data.ID)
		}
	default:
		return fmt.Errorf("message %s is not supported", m.Event)
	}
	return nil
}

// Load all the policies from the backend to store
func (ps *policyStore) load() error {
	conn := ps.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	bytes, err := redis.Values(conn.Do("ZRANGE", rds.KeyPeriodicPolicy(ps.namespace), 0, -1))
	if err != nil {
		return err
	}

	count := 0
	for i, l := 0, len(bytes); i < l; i++ {
		rawPolicy := bytes[i].([]byte)
		p := &Policy{}

		if err := p.DeSerialize(rawPolicy); err != nil {
			// Ignore error which means the policy data is not valid
			// Only logged
			logger.Errorf("malform policy: %s; error: %s\n", rawPolicy, err)
			continue
		}

		// Add to cache store
		if err := ps.add(p); err != nil {
			// Only logged
			logger.Errorf("cache periodic policies error: %s", err)
			continue
		}

		count++

		logger.Debugf("Load periodic job policy: %s", string(rawPolicy))
	}

	logger.Infof("Load %d periodic job policies", count)

	return nil
}

func (ps *policyStore) add(item *Policy) error {
	if item == nil {
		return errors.New("nil policy to add")
	}
	if utils.IsEmptyStr(item.ID) {
		return errors.New("malform policy to add")
	}
	// 使用sync.Map 来存储
	v, _ := ps.hash.LoadOrStore(item.ID, item)
	if v == nil {
		return fmt.Errorf("failed to add policy: %s", item.ID)
	}
	return nil
}

func (ps *policyStore) remove(policyID string) *Policy {
	if utils.IsEmptyStr(policyID) {
		return nil
	}
	if v, ok := ps.hash.Load(policyID); ok {
		ps.hash.Delete(policyID)
		return v.(*Policy)
	}

	return nil
}

// Iterate all the policies in the store
func (ps *policyStore) Iterate(f func(id string, p *Policy) bool) {
	ps.hash.Range(func(k, v interface{}) bool {
		return f(k.(string), v.(*Policy))
	})
}
