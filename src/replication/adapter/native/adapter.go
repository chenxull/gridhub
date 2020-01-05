package native

import (
	"github.com/chenxull/goGridhub/gridhub/src/common/http/modifier"
	common_http_auth "github.com/chenxull/goGridhub/gridhub/src/common/http/modifier/auth"
	registry_pkg "github.com/chenxull/goGridhub/gridhub/src/common/utils/registry"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	adp "github.com/chenxull/goGridhub/gridhub/src/replication/adapter"
	"github.com/chenxull/goGridhub/gridhub/src/replication/model"
	"net/http"
	"sync"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeDockerRegistry, new(factory)); err != nil {
		logger.Errorf("failed to register factory for %s: %v", model.RegistryTypeDockerRegistry, err)
		return
	}
	logger.Infof("the factory for adapter %s registered", model.RegistryTypeDockerRegistry)
}

var _ adp.Adapter = &Adapter{}

type factory struct {
}

func (f *factory) Create(*model.Registry) (adp.Adapter, error) {
	panic("implement me")
}

func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

// Adapter implements an adapter for Docker registry. It can be used to all registries
// that implement the registry V2 API
type Adapter struct {
	sync.RWMutex
	*registry_pkg.Registry
	registry *model.Registry
	client   *http.Client
	clients  map[string]*registry_pkg.Repository // client for repositories
}

func NewAdapter(registry *model.Registry) (*Adapter, error) {
	var cred modifier.Modifier
	if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
		if registry.Credential.Type == model.CredentialTypeSecret {
			cred = common_http_auth.NewSecretAuthorizer(registry.Credential.AccessSecret)
		} else {
			cred = auth.NewBasicAuthCredential(
				registry.Credential.AccessKey,
				registry.Credential.AccessSecret)
		}
	}
}
