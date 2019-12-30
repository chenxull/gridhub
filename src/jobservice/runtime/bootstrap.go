package runtime

import (
	"context"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/api"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/config"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/core"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/env"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/hook"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/lcm"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/mgt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/worker"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	dialConnectionTimeout = 30 * time.Second
	healthCheckPeriod     = time.Minute
	dialReadTimeout       = healthCheckPeriod + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
)

// JobService ...
var JobService = &Bootstrap{}

type Bootstrap struct {
	//jobContextInitializer is  func(ctx context.Context) (Context, error)
	jobConextInitializer job.ContextInitializer
}

// SetJobContextInitializer set the job context initializer
func (bs *Bootstrap) SetJobContextInitializer(initializer job.ContextInitializer) {
	if initializer != nil {
		bs.jobConextInitializer = initializer
	}
}

// LoadAndRun will load configurations, initialize components and then start the related process to serve requests.
// Return error if meet any problems.
func (bs *Bootstrap) LoadAndRun(ctx context.Context, cancel context.CancelFunc) (err error) {
	//rootContext 作为上下文的根控制整个应用中的数据
	rootContext := &env.Context{
		SystemContext: ctx,
		WG:            &sync.WaitGroup{},
		ErrorChan:     make(chan error, 5),
	}

	// Build specified job context。设计的很巧妙,使用在main 函数中注册的函数来初始化ctx 使其携带有配置信息
	if bs.jobConextInitializer != nil {
		rootContext.JobContext, err = bs.jobConextInitializer(ctx)
	}

	// 通过这个引用，就可用访问其方法
	cfg := config.DefaultConfig

	var (
		// worker 工作框架对象
		backendWorker worker.Interface
		// 获取 job 先关信息
		manager mgt.Manager
	)
	// 启动redis
	if cfg.PoolConfig.Backend == config.JobServicePoolBackendRedis {
		// Number of workers
		workerNum := cfg.PoolConfig.WorkerCount

		// Add {} to namespace to void slot issue
		namespace := fmt.Sprintf("{%s}", cfg.PoolConfig.RedisPoolCfg.Namespace)
		// Get redis connection pool
		redisPool := bs.getRedisPool(cfg.PoolConfig.RedisPoolCfg)

		manager = mgt.NewManager(ctx, namespace, redisPool)
		//todo create hook agent ,it's a singleton object

		hookAgent := hook.NewAgent(rootContext, namespace, redisPool)
		hookCallback := func(URL string, change *job.StatusChange) error {
			msg := fmt.Sprintf("status change: job=%s, status=%s", change.JobID, change.Status)
			if !utils.IsEmptyStr(change.CheckIn) {
				msg = fmt.Sprintf("%s, check_in=%s", msg, change.CheckIn)
			}
			evt := &hook.Event{
				URL:       URL,
				Message:   msg,
				Data:      change,
				Timestamp: time.Now().Unix(),
			}
			return hookAgent.Trigger(evt)
		}
		// Create job life cycle management controller
		lcmCtl := lcm.NewController(rootContext, namespace, redisPool, hookCallback)

		// Start the backend worker
		backendWorker, err = bs.loadAndRunRedisWorkerPool(
			rootContext,
			namespace,
			workerNum,
			redisPool,
			lcmCtl,
		)
		if err != nil {
			return errors.Errorf("load and run worker error: %s", err)

		}

		//Run daemon process of life cycle controller
		if err = lcmCtl.Serve(); err != nil {
			return errors.Errorf("start life cycle controller error: %s", err)
		}

		// Start agent
		// Non blocking call
		hookAgent.Attach(lcmCtl)
		if err = hookAgent.Serve(); err != nil {
			return errors.Errorf("start hook agent error: %s", err)
		}
	} else {
		return errors.Errorf("worker backend '%s' is not supported", cfg.PoolConfig.Backend)
	}

	// Initialize controller
	ctl := core.NewController(backendWorker, manager)
	apiServer := bs.createAPIServer(ctx, cfg, ctl)

	//Listen to the system signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, os.Kill)
	terminated := false
	go func(errChan chan error) {
		defer func() {
			//GraceFully shutdown
			if er := apiServer.Stop(); er != nil {
				logger.Error(er)
			}
			// Notify others who're listening to the system context
			cancel()
		}()

		select {
		case <-sig:
			terminated = true
		case err = <-errChan:
			return
		}
	}(rootContext.ErrorChan)
	node := ctx.Value(utils.NodeID)
	logger.Infof("API server is serving at %d with [%s] mode at node [%s]", cfg.Port, cfg.Protocol, node)
	if er := apiServer.Start(); er != nil {
		if !terminated {
			// Tell the listening goroutine
			rootContext.ErrorChan <- er
		}
	} else {
		// In case
		sig <- os.Interrupt
	}

	// Wait everyone exit
	rootContext.WG.Wait()

	return
}

//Load and run the API server
func (bs *Bootstrap) createAPIServer(ctx context.Context, cfg *config.Configuration, ctl core.Interface) *api.Server {

	authProvider := &api.SecretAuthenticator{}
	handler := api.NewDefaultHandler(ctl)
	router := api.NewBaseRouter(handler, authProvider)
	serverConfig := api.ServerConfig{
		Protocol: cfg.Protocol,
		Port:     cfg.Port,
	}
	if cfg.HTTPSConfig != nil {
		serverConfig.Cert = cfg.HTTPSConfig.Cert
		serverConfig.Key = cfg.HTTPSConfig.Key
	}
	return api.NewServer(ctx, router, serverConfig)
}

// Load and run the worker worker
func (bs *Bootstrap) loadAndRunRedisWorkerPool(
	ctx *env.Context,
	ns string,
	workers uint,
	redisPool *redis.Pool,
	lcmCtl lcm.Controller,
) (worker.Interface, error) {

}

// Get a redis connection pool
func (bs *Bootstrap) getRedisPool(redisPoolConfig *config.RedisPoolConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     6,
		Wait:        true,
		IdleTimeout: time.Duration(redisPoolConfig.IdleTimeoutSecond) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(
				redisPoolConfig.RedisURL,
				redis.DialConnectTimeout(dialConnectionTimeout),
				redis.DialReadTimeout(dialReadTimeout),
				redis.DialWriteTimeout(dialWriteTimeout),
			)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}

			_, err := c.Do("PING")
			return err
		},
	}
}
