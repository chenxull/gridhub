package impl

import (
	"context"
	"errors"
	"fmt"
	comcfg "github.com/chenxull/goGridhub/gridhub/src/common/config"
	"github.com/chenxull/goGridhub/gridhub/src/common/dao"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"math"
	"sync"
	"time"
)

const (
	maxRetryTimes = 5
)

//Context ...
type Context struct {
	sysContext context.Context
	logger     logger.Interface
	properties map[string]interface{}
	// admin server client
	cfgMgr comcfg.CfgManager
	// job life cycle tracker
	tracker job.Tracker
	// job logger configs settings map lock
	lock sync.Mutex
}

func NewContext(sysCtx context.Context, cfgMgr *comcfg.CfgManager) *Context {
	return &Context{
		sysContext: sysCtx,
		cfgMgr:     *cfgMgr,
		properties: make(map[string]interface{}),
	}
}

// Init ...
func (c *Context) Init() error {
	var (
		counter = 0
		err     error
	)

	for counter == 0 || err != nil {
		counter++
		err = c.cfgMgr.Load()
		if err != nil {
			logger.Errorf("Job context initialization error: %s\n", err.Error())
			if counter < maxRetryTimes {
				backoff := (int)(math.Pow(2, (float64)(counter))) + 2*counter + 5
				logger.Infof("Retry in %d seconds", backoff)
				time.Sleep(time.Duration(backoff) * time.Second)
			} else {
				return fmt.Errorf("job context initialization error: %s (%d times tried)", err.Error(), counter)
			}
		}
	}
	db := c.cfgMgr.GetDatabaseCfg()

	err = dao.InitDatabase(db)
	if err != nil {
		return err
	}
	return nil
}

// Build implements the same method in env.JobContext interface
// This func will build the job execution context before running
func (c *Context) Build(tracker job.Tracker) (job.Context, error) {
	if tracker == nil || tracker.Job() == nil {
		return nil, errors.New("nil job tracker")
	}
	jContext := &Context{
		sysContext: c.sysContext,
		cfgMgr:     c.cfgMgr,
		properties: make(map[string]interface{}),
		tracker:    tracker,
	}

	// Copy properties
	if len(c.properties) > 0 {
		for k, v := range c.properties {
			jContext.properties[k] = v
		}
	}

	// 更新配置信息
	err := c.cfgMgr.Load()
	if err != nil {
		return nil, err
	}

	props := c.cfgMgr.GetAll()
	for k, v := range props {
		jContext.properties[k] = v
	}

	//todo set loggers for job

	return jContext, nil

}

func (c *Context) Get(prop string) (interface{}, bool) {
	v, ok := c.properties[prop]
	return v, ok
}

func (c *Context) SystemContext() context.Context {
	return c.sysContext
}

func (c *Context) Checkin(status string) error {
	return c.tracker.CheckIn(status)
}

func (c *Context) OPCommand() (job.OPCommand, bool) {
	latest, err := c.tracker.Status()
	if err != nil {
		return job.NilCommand, false
	}

	if job.StoppedStatus == latest {
		return job.StopCommand, true
	}

	return job.NilCommand, false
}

func (c *Context) Tracker() job.Tracker {
	return c.tracker
}
