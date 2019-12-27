package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/common"
	comcfg "github.com/chenxull/goGridhub/gridhub/src/common/config"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/config"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job/impl"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/runtime"
)

func main() {
	configPath := flag.String("c", "", "Specify the yaml config file path")
	flag.Parse()

	if configPath == nil || utils.IsEmptyStr(*configPath) {
		flag.Usage()
		panic("no config file is specified")
	}

	// load configuration
	if err := config.DefaultConfig.Load(*configPath, true); err != nil {
		panic(fmt.Sprintf("load configurations error: %s\n", err))
	}
	//Append node ID
	vCtx := context.WithValue(context.Background(), utils.NodeID, utils.GenerateNodeID())
	//Create the root context
	ctx, cancel := context.WithCancel(vCtx)
	defer cancel()

	//todo Initialize logger
	//if err := logger.Init(ctx); err != nil {
	//	panic(err)
	//}

	runtime.JobService.SetJobContextInitializer(func(ctx context.Context) (job.Context, error) {
		secret := config.GetAuthSecret()
		if utils.IsEmptyStr(secret) {
			return nil, errors.New("empty auth secret")
		}
		coreURL := config.GetCoreURL()
		configURL := coreURL + common.CoreConfigPath
		cfgMgr := comcfg.NewRESTCfgManager(configURL, secret)
		jobCtx := impl.NewContext(ctx, cfgMgr)

		if err := jobCtx.Init(); err != nil {
			return nil, err
		}

		return jobCtx, nil
	})

	// start
	if err := runtime.JobService.LoadAndRun(ctx, cancel); err != nil {
		logger.Fatal(err)
	}
}
