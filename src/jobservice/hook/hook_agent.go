package hook

import (
	"context"
	"encoding/json"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/rds"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/env"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/lcm"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

const (
	// Influenced by the worker number setting
	maxEventChanBuffer = 1024
	// Max concurrent client handlers
	maxHandlers = 5
	// The max time for expiring the retrying events
	// 180 days
	maxEventExpireTime = 3600 * 24 * 180
	// Waiting a short while if any errors occurred
	shortLoopInterval = 5 * time.Second
	// Waiting for long while if no retrying elements found
	longLoopInterval = 5 * time.Minute
)

type Agent interface {
	//Trigger hooks
	Trigger(evt *Event) error
	// Serves events now
	Serve() error
	// Attach a job life cycle controller
	Attach(ctl lcm.Controller)
}

// Event contains the hook URL and the data
type Event struct {
	URL       string            `json:"url"`
	Message   string            `json:"message"`   // meaningful text for event
	Data      *job.StatusChange `json:"data"`      // generic data
	Timestamp int64             `json:"timestamp"` // Use as time threshold of discarding the event (unit: second)
}

func (e *Event) Validate() error {
	_, err := url.Parse(e.URL)
	if err != nil {
		return err
	}
	if e.Data == nil {
		return errors.New("nil hook data")

	}
	return nil
}

func (e *Event) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// Deserialize the bytes to event
func (e *Event) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, e)
}

//Basic agent for usage
type basicAgent struct {
	context   context.Context
	namespace string
	client    Client
	ctl       lcm.Controller
	events    chan *Event
	tokens    chan bool
	redisPool *redis.Pool
	wg        *sync.WaitGroup
}

// NewAgent is constructor of basic agent
func NewAgent(ctx *env.Context, ns string, redisPool *redis.Pool) Agent {
	tks := make(chan bool, maxHandlers)
	// Put tokens
	for i := 0; i < maxHandlers; i++ {
		tks <- true
	}
	return &basicAgent{
		context:   ctx.SystemContext,
		namespace: ns,
		client:    NewClient(ctx.SystemContext),
		events:    make(chan *Event, maxEventChanBuffer),
		tokens:    tks,
		redisPool: redisPool,
		wg:        ctx.WG,
	}
}

//Trigger hooks
func (ba *basicAgent) Trigger(evt *Event) error {
	if evt == nil {
		return errors.New("nil event")
	}

	if err := evt.Validate(); err != nil {
		return err
	}

	// 验证完将时间发送到通道中
	ba.events <- evt

	return nil
}

// Attach a job life cycle controller
func (ba *basicAgent) Attach(ctl lcm.Controller) {
	ba.ctl = ctl
}

//Start the basic agent
func (ba *basicAgent) Serve() error {
	if ba.ctl == nil {
		return errors.New("nil life cycle controller of hook agent")
	}

	ba.wg.Add(1)
	// 用来发送 重试队列中的数据
	go ba.loopRetry()
	logger.Info("Hook event retrying loop is started")

	ba.wg.Add(1)
	go ba.serve()
	logger.Info("Basic hook agent is started")

	return nil
}

func (ba *basicAgent) serve() {
	defer func() {
		logger.Info("Basic hook agent is stopped")
		ba.wg.Done()
	}()

	for {
		select {
		case evt := <-ba.events:
			// 限流，确保在一定时间内只有固定数量的事件
			<-ba.tokens

			go func(evt *Event) {
				defer func() {
					ba.tokens <- true //return token
				}()

				// send message to endpoint
				if err := ba.client.SendEvent(evt); err != nil {
					logger.Errorf("Send hook event '%s' to '%s' failed with error: %s; push to the queue for retrying later", evt.Message, evt.URL, err)
					// 将任务放入重试队列
					if err := ba.pushForRetry(evt); err != nil {
						logger.Errorf("Failed to push hook event to the retry queue: %s", err)
						<-time.After(time.Duration(rand.Int31n(55)+5) * time.Second)
						ba.events <- evt
					}
				}
			}(evt)
		case <-ba.context.Done():
			return
		}
	}
}

// 对redis 队列
func (ba *basicAgent) pushForRetry(evt *Event) error {
	if evt == nil {
		return nil
	}
	rawJSON, err := evt.Serialize()
	if err != nil {
		return nil
	}

	now := time.Now().Unix()
	if evt.Timestamp > 0 && now-evt.Timestamp >= maxEventExpireTime {
		// Expired, do not need to push back to the retry queue
		logger.Warningf("Event is expired: %s", rawJSON)

		return nil
	}

	conn := ba.redisPool.Get()
	defer func() {
		_ = conn.Close()
	}()
	key := rds.KeyHookEventRetryQueue(ba.namespace)
	args := make([]interface{}, 0)
	score := time.Now().UnixNano()
	args = append(args, key, "NX", score, rawJSON)
	_, err = conn.Do("ZADD", args...)
	if err != nil {
		return err
	}

	return nil
}

func (ba *basicAgent) loopRetry() {
	defer func() {
		logger.Info("Hook event retrying loop exit")
		ba.wg.Done()
	}()
	token := make(chan bool, 1)
	token <- true

	for {
		<-token
		if err := ba.reSend(); err != nil {
			waitInterval := shortLoopInterval
			if err == rds.ErrNoElements {
				waitInterval = longLoopInterval
			} else {
				logger.Errorf("Resend hook event error: %s", err.Error())
			}

			// 定时器
			select {
			case <-time.After(waitInterval):
			// Just wait,do nothing
			case <-ba.context.Done():
				//terminated
				return
			}
		}
		// put token back
		token <- true
	}
}

//先从redis 中获取最近的一个，然后将其发送
func (ba *basicAgent) reSend() error {
	evt, err := ba.popMinOne()
	if err != nil {
		return err
	}

	jobID, status, err := extractJobID(evt.Data)
	if err != nil {
		return err
	}
	// 获取job 的信息
	t, err := ba.ctl.Track(jobID)
	if err != nil {
		return err
	}
	// 当前状态和tracker 中的状态对比
	diff := status.Compare(job.Status(t.Job().Info.Status))
	if diff > 0 ||
		(diff == 0 && t.Job().Info.CheckIn != evt.Data.CheckIn) {
		ba.events <- evt
		return nil
	}
	return errors.Errorf("outdated hook event: %s, latest job status: %s", evt.Message, t.Job().Info.Status)

}

// 从redis 中获取信息
func (ba *basicAgent) popMinOne() (*Event, error) {
	conn := ba.redisPool.Get()
	defer func() {
		_ = conn.Close()
	}()

	key := rds.KeyHookEventRetryQueue(ba.namespace)
	minOne, err := rds.ZPopMin(conn, key)
	if err != nil {
		return nil, err
	}

	rawEvent, ok := minOne.([]byte)
	if !ok {
		return nil, errors.New("bad request: non bytes slice for raw event")
	}

	evt := &Event{}
	if err := evt.Deserialize(rawEvent); err != nil {
		return nil, err
	}

	return evt, nil
}
func extractJobID(data *job.StatusChange) (string, job.Status, error) {
	if data != nil && len(data.JobID) > 0 {
		status := job.Status(data.Status)
		if status.Validate() == nil {
			return data.JobID, status, nil
		}
	}

	return "", "", errors.New("malform job status change data")
}
