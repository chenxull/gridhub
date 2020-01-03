package config

import (
	"github.com/chenxull/goGridhub/gridhub/src/common"
	comcfg "github.com/chenxull/goGridhub/gridhub/src/common/config"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"os"
	"strings"
)

const (
	defaultKeyPath                     = "/etc/core/key"
	defaultRegistryTokenPrivateKeyPath = "/etc/core/private_key.pem"

	// SessionCookieName is the name of the cookie for session ID
	SessionCookieName = "sid"
)

var (
	cfgMgr      *comcfg.CfgManager
	keyProvider comcfg.KeyProvider
)

func initKeyProvider() {
	path := os.Getenv("KEY_PATH")
	if len(path) == 0 {
		path = defaultKeyPath
	}
	logger.Infof("key path: %s", path)

	keyProvider = comcfg.NewFileKeyProvider(path)
}

// CoreSecret returns a secret to mark harbor-core when communicate with
// other component
func CoreSecret() string {
	return os.Getenv("CORE_SECRET")
}

// JobserviceSecret returns a secret to mark Jobservice when communicate with
// other component
// TODO replace it with method of SecretStore
func JobserviceSecret() string {
	return os.Getenv("JOBSERVICE_SECRET")
}

// SecretKey returns the secret key to encrypt the password of target
func SecretKey() (string, error) {
	return keyProvider.Get(nil)
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	return strings.TrimSuffix(cfgMgr.Get(common.JobServiceURL).GetString(), "/")
}

// InternalCoreURL returns the local harbor core url
func InternalCoreURL() string {
	return strings.TrimSuffix(cfgMgr.Get(common.CoreURL).GetString(), "/")
}
