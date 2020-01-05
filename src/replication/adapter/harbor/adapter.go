package harbor

import (
	common_http "github.com/chenxull/goGridhub/gridhub/src/common/http"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	adp "github.com/chenxull/goGridhub/gridhub/src/replication/adapter"
	"github.com/chenxull/goGridhub/gridhub/src/replication/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeHarbor, new(factory)); err != nil {
		logger.Errorf("failed to register factory for %s: %v", model.RegistryTypeHarbor, err)
		return
	}
	logger.Infof("the factory for adapter %s registered", model.RegistryTypeHarbor)

}

type factory struct {
}

func (f *factory) Create(*model.Registry) (adp.Adapter, error) {
	panic("implement me")
}

func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

type adapter struct {
	*native.Adapter
	registry *model.Registry
	url      string
	client   *common_http.Client
}

func newAdapter(registry *model.Registry) (*adpter, error) {

}
