package core

import (
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/mgt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/worker"
)

// basicController implement the core interface and provides related job handle methods.
// basicController will coordinate the lower components to complete the process as a commander role.
type basicController struct {
	//Refer the backend worker
	backendWorker worker.Interface
	//Refer the job stats manager
	manager mgt.Manager
}
