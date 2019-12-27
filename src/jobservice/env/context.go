package env

import (
	"context"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/job"
	"sync"
)

// 用来保存一些可共享变量 和系统控制通道
type Context struct {
	// The system context with cancel capability.
	SystemContext context.Context

	WG *sync.WaitGroup

	// Report errors to bootstrap component
	ErrorChan chan error

	// The base job context reference
	// It will be the parent context of job execution context
	JobContext job.Context
}
