package config

import (
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/common"
	"github.com/chenxull/goGridhub/gridhub/src/common/config/metadata"
	"github.com/chenxull/goGridhub/gridhub/src/common/config/models"
	"github.com/chenxull/goGridhub/gridhub/src/common/config/store"
	"github.com/chenxull/goGridhub/gridhub/src/common/config/store/driver"
	"github.com/chenxull/goGridhub/gridhub/src/common/http/modifier/auth"
	"github.com/chenxull/goGridhub/gridhub/src/common/utils"
	"log"
	"os"
	"sync"
)

/*
	配置信息整体的管理器，最高层次的抽象，在下一层还有存储驱动，存储对象实体。
 	同时也是对外暴露的窗口，其他组件在获取项目的配置信息时，都是通过此处暴露的方法来获取配置信息的
*/

type CfgManager struct {
	store *store.ConfigStore
}

// NewDBCfgManager - create DB config manager
func NewDBCfgManager() *CfgManager {
	manager := &CfgManager{store: store.NewConfigStore(&driver.Database{})}
	// load default value
	manager.loadDefault()
	manager.loadSystemConfigFromEnv()
	return manager
}

// NewRESTCfgManager - create REST config manager
func NewRESTCfgManager(configURL, secret string) *CfgManager {
	secAuth := auth.NewSecretAuthorizer(secret)
	manager := &CfgManager{
		store: store.NewConfigStore(driver.NewRESTDriver(configURL, secAuth))}
	return manager
}

// InMemoryDriver driver for unit testing
type InMemoryDriver struct {
	sync.Mutex
	cfgMap map[string]interface{}
}

func (d *InMemoryDriver) Load() (map[string]interface{}, error) {
	d.Lock()
	defer d.Unlock()
	res := make(map[string]interface{})
	for k, v := range d.cfgMap {
		res[k] = v
	}
	return res, nil
}

func (d *InMemoryDriver) Save(cfg map[string]interface{}) error {
	d.Lock()
	defer d.Unlock()
	for k, v := range cfg {
		d.cfgMap[k] = v
	}
	return nil
}

// 载入数据 从 metadata 中将数据存储到具体的存储驱动中, 主要是处理一些默认配置信息
func (c *CfgManager) loadDefault() {
	// Init Default value
	itemArray := metadata.Instance().GetAll()
	for _, item := range itemArray {
		if len(item.DefaultValue) > 0 {
			cfgValue, err := metadata.NewCfgValue(item.Name, item.DefaultValue)
			if err != nil {
				//todo log.Errorf("loadDefault failed, config item, key: %v,  err: %v", item.Name, err)
				continue
			}
			c.store.Set(item.Name, *cfgValue)
		}
	}
}

// 从环境变量中获取配置信息
func (c *CfgManager) loadSystemConfigFromEnv() {
	itemArray := metadata.Instance().GetAll()
	for _, item := range itemArray {
		if item.Scope == metadata.SystemScope && len(item.EnvKey) > 0 {
			if envValue, ok := os.LookupEnv(item.EnvKey); ok {
				configValue, err := metadata.NewCfgValue(item.Name, envValue)
				if err != nil {
					//todo log.Errorf("loadSystemConfigFromEnv failed, config item, key: %v,  err: %v", item.Name, err)
					continue
				}
				c.store.Set(item.Name, *configValue)
			}
		}
	}
}

//GetAll get all settings
func (c *CfgManager) GetAll() map[string]interface{} {
	resultMap := map[string]interface{}{}
	// 先从存储后端读取数据到本地内存中
	if err := c.store.Load(); err != nil {
		log.Fatalf("GetAll failed, error %v", err)
		return resultMap
	}
	// 从刚刚载入到内存中的配置信息获取所有数据
	metaDataList := metadata.Instance().GetAll()
	for _, item := range metaDataList {
		cfgValue, err := c.store.GetAnyType(item.Name)
		if err != nil {
			if err != metadata.ErrValueNotSet {
				log.Fatalf("Failed to get value of key %v, error %v", item.Name, err)
			}
			continue
		}
		resultMap[item.Name] = cfgValue
	}
	return resultMap
}

// 获取Scope 为user 的配置信息
func (c *CfgManager) GetUserCfgs() map[string]interface{} {
	resultMap := map[string]interface{}{}
	// 先从存储后端读取数据到本地内存中
	if err := c.store.Load(); err != nil {
		log.Fatalf("GetAll failed, error %v", err)
		return resultMap
	}
	metaDataList := metadata.Instance().GetAll()
	for _, item := range metaDataList {
		// 判断是否为用户配置
		if item.Scope == metadata.UserScope {
			cfgValue, err := c.store.GetAnyType(item.Name)
			if err != nil {
				if err == metadata.ErrValueNotSet {
					if _, ok := item.ItemType.(*metadata.StringType); ok {
						cfgValue = ""
					}
					if _, ok := item.ItemType.(*metadata.NonEmptyStringType); ok {
						cfgValue = ""
					}
				} else {
					log.Printf("Failed to get value of key %v, error %v", item.Name, err)
					continue
				}
			}
			resultMap[item.Name] = cfgValue
		}
	}
	return resultMap
}

// Load load configuration from storage, like database or redis
func (c *CfgManager) Load() error {
	return c.store.Load()
}

// Save - Save all current configuration to storage
func (c *CfgManager) Save() error {
	return c.store.Save()
}

// Get ...
func (c *CfgManager) Get(key string) *metadata.ConfigureValue {
	configValue, err := c.store.Get(key)
	if err != nil {
		log.Printf("failed to get key %v, error: %v", key, err)
		configValue = &metadata.ConfigureValue{}
	}
	return configValue
}

// Set ...
func (c *CfgManager) Set(key string, value interface{}) {
	configValue, err := metadata.NewCfgValue(key, utils.GetStrValueOfAnyType(value))
	if err != nil {
		log.Fatalf("error when setting key: %v,  error %v", key, err)
		return
	}
	c.store.Set(key, *configValue)
}

// UpdateConfig - Update config store with a specified configuration and also save updated configure.
func (c *CfgManager) UpdateConfig(cfgs map[string]interface{}) error {
	return c.store.Update(cfgs)
}

// 返回数据库的配置信息
func (c *CfgManager) GetDatabaseCfg() *models.Database {
	return &models.Database{
		Type: c.Get(common.DatabaseType).GetString(),
		PostGreSQL: &models.PostGreSQL{
			Host:         c.Get(common.PostGreSQLHOST).GetString(),
			Port:         c.Get(common.PostGreSQLPort).GetInt(),
			Username:     c.Get(common.PostGreSQLUsername).GetString(),
			Password:     c.Get(common.PostGreSQLPassword).GetString(),
			Database:     c.Get(common.PostGreSQLDatabase).GetString(),
			SSLMode:      c.Get(common.PostGreSQLSSLMode).GetString(),
			MaxIdleConns: c.Get(common.PostGreSQLMaxIdleConns).GetInt(),
			MaxOpenConns: c.Get(common.PostGreSQLMaxOpenConns).GetInt(),
		},
	}
}

// ValidateCfg validate config by metadata. return the first error if exist.
func (c *CfgManager) ValidateCfg(cfgs map[string]interface{}) error {
	for key, value := range cfgs {
		strVal := utils.GetStrValueOfAnyType(value)
		_, err := metadata.NewCfgValue(key, strVal)
		if err != nil {
			return fmt.Errorf("%v, item name: %v", err, key)
		}
	}
	return nil
}

// DumpTrace dump all configurations
func (c *CfgManager) DumpTrace() {
	cfgs := c.GetAll()
	for k, v := range cfgs {
		log.Printf(k, ":=", v)
	}
}
