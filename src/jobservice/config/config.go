package config

import (
	"errors"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
)

const (
	jobServiceProtocol                   = "JOB_SERVICE_PROTOCOL"
	jobServicePort                       = "JOB_SERVICE_PORT"
	jobServiceHTTPCert                   = "JOB_SERVICE_HTTPS_CERT"
	jobServiceHTTPKey                    = "JOB_SERVICE_HTTPS_KEY"
	jobServiceWorkerPoolBackend          = "JOB_SERVICE_POOL_BACKEND"
	jobServiceWorkers                    = "JOB_SERVICE_POOL_WORKERS"
	jobServiceRedisURL                   = "JOB_SERVICE_POOL_REDIS_URL"
	jobServiceRedisNamespace             = "JOB_SERVICE_POOL_REDIS_NAMESPACE"
	jobServiceRedisIdleConnTimeoutSecond = "JOB_SERVICE_POOL_REDIS_CONN_IDLE_TIMEOUT_SECOND"
	jobServiceAuthSecret                 = "JOBSERVICE_SECRET"
	coreURL                              = "CORE_URL"

	// JobServiceProtocolHTTPS points to the 'https' protocol
	JobServiceProtocolHTTPS = "https"
	// JobServiceProtocolHTTP points to the 'http' protocol
	JobServiceProtocolHTTP = "http"

	// JobServicePoolBackendRedis represents redis backend
	JobServicePoolBackendRedis = "redis"

	// secret of UI
	uiAuthSecret = "CORE_SECRET"

	// redis protocol schema
	redisSchema = "redis://"
)

//DefaultConfig is the default configuration reference
var DefaultConfig = &Configuration{}

type Configuration struct {
	Protocol    string       `yaml:"protocol"`
	Port        uint         `yaml:"port"`
	HTTPSConfig *HTTPSConfig `yaml:"https_config,omitempty"`
	PoolConfig  *PoolConfig  `yaml:"worker_pool,omitempty"`
	// Job logger configurations
	JobLoggerConfigs []*LoggerConfig `yaml:"job_loggers,omitempty"`

	// Logger configurations
	LoggerConfigs []*LoggerConfig `yaml:"loggers,omitempty"`
}

type HTTPSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type RedisPoolConfig struct {
	RedisURL  string `yaml:"redis_url"`
	Namespace string `yaml:"namespace"`
	// IdleTimeoutSecond closes connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeoutSecond int64 `yaml:"idle_timeout_second"`
}

type PoolConfig struct {
	// Worker concurrency
	WorkerCount  uint             `yaml:"workers"`
	Backend      string           `yaml:"backend"`
	RedisPoolCfg *RedisPoolConfig `yaml:"redis_pool,omitempty"`
}

// CustomizedSettings keeps the customized settings of logger
type CustomizedSettings map[string]interface{}

// LogSweeperConfig keeps settings of log sweeper
type LogSweeperConfig struct {
	Duration int                `yaml:"duration"`
	Settings CustomizedSettings `yaml:"settings"`
}

type LoggerConfig struct {
	Name     string             `yaml:"name"`
	Level    string             `yaml:"level"`
	Settings CustomizedSettings `yaml:"settings"`
	Sweeper  *LogSweeperConfig  `yaml:"sweeper"`
}

func (c *Configuration) Load(yamlFilePath string, detectEnv bool) error {
	if !utils.IsEmptyStr(yamlFilePath) {
		data, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return err
		}
		if err = yaml.Unmarshal(data, c); err != nil {
			return err
		}
	}
	if detectEnv {
		c.loadEnvs()
	}
	if c.PoolConfig != nil && c.PoolConfig.RedisPoolCfg != nil {
		redisAddress := c.PoolConfig.RedisPoolCfg.RedisURL
		if !utils.IsEmptyStr(redisAddress) {
			if _, err := url.Parse(redisAddress); err != nil {
				if redisURL, ok := utils.TranslateRedisAddress(redisAddress); ok {
					c.PoolConfig.RedisPoolCfg.RedisURL = redisURL
				}
			} else {
				if !strings.HasPrefix(redisAddress, redisSchema) {
					c.PoolConfig.RedisPoolCfg.RedisURL = fmt.Sprintf("%s%s", redisSchema, redisAddress)
				}
			}
		}
	}
	return c.validate()

}

func (c *Configuration) loadEnvs() {
	prot := utils.ReadEnv(jobServiceProtocol)
	if !utils.IsEmptyStr(prot) {
		c.Protocol = prot
	}

	p := utils.ReadEnv(jobServicePort)
	if !utils.IsEmptyStr(p) {
		if po, err := strconv.Atoi(p); err == nil {
			c.Port = uint(po)
		}
	}

	// Only when protocol is https
	if c.Protocol == JobServiceProtocolHTTPS {
		cert := utils.ReadEnv(jobServiceHTTPCert)
		if !utils.IsEmptyStr(cert) {
			if c.HTTPSConfig != nil {
				c.HTTPSConfig.Cert = cert
			} else {
				c.HTTPSConfig = &HTTPSConfig{
					Cert: cert,
				}
			}
		}

		certKey := utils.ReadEnv(jobServiceHTTPKey)
		if !utils.IsEmptyStr(certKey) {
			if c.HTTPSConfig != nil {
				c.HTTPSConfig.Key = certKey
			} else {
				c.HTTPSConfig = &HTTPSConfig{
					Key: certKey,
				}
			}
		}
	}

	backend := utils.ReadEnv(jobServiceWorkerPoolBackend)
	if !utils.IsEmptyStr(backend) {
		if c.PoolConfig == nil {
			c.PoolConfig = &PoolConfig{}
		}
		c.PoolConfig.Backend = backend
	}

	workers := utils.ReadEnv(jobServiceWorkers)
	if !utils.IsEmptyStr(workers) {
		if count, err := strconv.Atoi(workers); err == nil {
			if c.PoolConfig == nil {
				c.PoolConfig = &PoolConfig{}
			}
			c.PoolConfig.WorkerCount = uint(count)
		}
	}

	if c.PoolConfig != nil && c.PoolConfig.Backend == JobServicePoolBackendRedis {
		redisURL := utils.ReadEnv(jobServiceRedisURL)
		if !utils.IsEmptyStr(redisURL) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			c.PoolConfig.RedisPoolCfg.RedisURL = redisURL
		}

		rn := utils.ReadEnv(jobServiceRedisNamespace)
		if !utils.IsEmptyStr(rn) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			c.PoolConfig.RedisPoolCfg.Namespace = rn
		}

		it := utils.ReadEnv(jobServiceRedisIdleConnTimeoutSecond)
		if !utils.IsEmptyStr(it) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			v, err := strconv.Atoi(it)
			if err != nil {
				//log.Warningf("Invalid idle timeout second: %s, will use 0 instead", it)
			} else {
				c.PoolConfig.RedisPoolCfg.IdleTimeoutSecond = int64(v)
			}
		}
	}
}

// GetAuthSecret get the auth secret from the env
func GetAuthSecret() string {
	return utils.ReadEnv(jobServiceAuthSecret)
}

// GetCoreURL get the core url from the env
func GetCoreURL() string {
	return utils.ReadEnv(coreURL)
}

// GetUIAuthSecret get the auth secret of UI side
func GetUIAuthSecret() string {
	return utils.ReadEnv(uiAuthSecret)
}

func (c *Configuration) validate() error {
	if c.Protocol != JobServiceProtocolHTTPS &&
		c.Protocol != JobServiceProtocolHTTP {
		return fmt.Errorf("protocol should be %s or %s, but current setting is %s",
			JobServiceProtocolHTTP,
			JobServiceProtocolHTTPS,
			c.Protocol)
	}
	if !utils.IsValidPort(c.Port) {
		return fmt.Errorf("port number should be a none zero integer and less or equal 65535, but current is %d", c.Port)
	}

	if c.Protocol == JobServiceProtocolHTTPS {
		if c.HTTPSConfig == nil {
			return fmt.Errorf("certificate must be configured if serve with protocol %s", c.Protocol)
		}

		if utils.IsEmptyStr(c.HTTPSConfig.Cert) ||
			!utils.FileExists(c.HTTPSConfig.Cert) ||
			utils.IsEmptyStr(c.HTTPSConfig.Key) ||
			!utils.FileExists(c.HTTPSConfig.Key) {
			return fmt.Errorf("certificate for protocol %s is not correctly configured", c.Protocol)
		}
	}

	if c.PoolConfig == nil {
		return errors.New("no worker worker is configured")
	}

	if c.PoolConfig.Backend != JobServicePoolBackendRedis {
		return fmt.Errorf("worker worker backend %s does not support", c.PoolConfig.Backend)
	}

	// When backend is redis
	if c.PoolConfig.Backend == JobServicePoolBackendRedis {
		if c.PoolConfig.RedisPoolCfg == nil {
			return fmt.Errorf("redis worker must be configured when backend is set to '%s'", c.PoolConfig.Backend)
		}
		if utils.IsEmptyStr(c.PoolConfig.RedisPoolCfg.RedisURL) {
			return errors.New("URL of redis worker is empty")
		}

		if !strings.HasPrefix(c.PoolConfig.RedisPoolCfg.RedisURL, redisSchema) {
			return errors.New("invalid redis URL")
		}

		if _, err := url.Parse(c.PoolConfig.RedisPoolCfg.RedisURL); err != nil {
			return fmt.Errorf("invalid redis URL: %s", err.Error())
		}

		if utils.IsEmptyStr(c.PoolConfig.RedisPoolCfg.Namespace) {
			return errors.New("namespace of redis worker is required")
		}
	}

	// Job service loggers
	if len(c.LoggerConfigs) == 0 {
		return errors.New("missing logger config of job service")
	}

	// Job loggers
	if len(c.JobLoggerConfigs) == 0 {
		return errors.New("missing logger config of job")
	}

	return nil // valid
}
