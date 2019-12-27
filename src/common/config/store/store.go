package store

import (
	"errors"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/common/config/metadata"
	"github.com/chenxull/goGridhub/gridhub/src/common/config/store/driver"
	"github.com/chenxull/goGridhub/gridhub/src/common/utils"
	"sync"
)

// 用来存储配置信息, 下游对接的是各种驱动。专注于存储信息
type ConfigStore struct {
	cfgDriver driver.Driver
	cfgValues sync.Map
}

// NewConfigStore create config store
func NewConfigStore(cfgDriver driver.Driver) *ConfigStore {
	// 初始化只要是给配置信息分配存储驱动
	return &ConfigStore{cfgDriver: cfgDriver}
}

//Get
func (c *ConfigStore) Get(key string) (*metadata.ConfigureValue, error) {
	if value, ok := c.cfgValues.Load(key); ok {
		// 配置信息必须要以 ConfigureValue 类型存储在后端存储中
		if result, ok := value.(metadata.ConfigureValue); ok {
			return &result, nil
		}
		return nil, errors.New("data in config store is not a ConfigureValue type")
	}
	return nil, metadata.ErrValueNotSet
}

// GetAnyType get interface{} type for config items
func (c *ConfigStore) GetAnyType(key string) (interface{}, error) {
	if value, ok := c.cfgValues.Load(key); ok {
		if result, ok := value.(metadata.ConfigureValue); ok {
			return result.GetAnyType()
		}
		return nil, errors.New("data in config store is not a ConfigureValue type")
	}
	return nil, metadata.ErrValueNotSet
}

// Set - Set configure value in store, not saved to config driver
func (c *ConfigStore) Set(key string, value metadata.ConfigureValue) error {
	c.cfgValues.Store(key, value)
	return nil
}

//Load data from driver,
func (c *ConfigStore) Load() error {
	if c.cfgDriver == nil {
		return errors.New("failed to load store, cfgDriver is nil")
	}
	// 载入具体的后端驱动
	cfgs, err := c.cfgDriver.Load()
	if err != nil {
		return err
	}
	for key, value := range cfgs {
		cfgValue := metadata.ConfigureValue{}
		strValue := fmt.Sprintf("%v", value)
		err = cfgValue.Set(key, strValue)
		if err != nil {
			//log.Errorf("error when loading data item, key %v, value %v, error %v", key, value, err)
			continue
		}
		c.cfgValues.Store(key, cfgValue)
	}
	return nil
}

// Save - Save all data in current store
func (c *ConfigStore) Save() error {
	cfgMap := map[string]interface{}{}
	c.cfgValues.Range(func(key, value interface{}) bool {
		keyStr := fmt.Sprintf("%v", key)
		if configValue, ok := value.(metadata.ConfigureValue); ok {
			valueStr := configValue.Value
			if _, ok := metadata.Instance().GetByName(keyStr); ok {
				cfgMap[keyStr] = valueStr
			} else {
				//log.Errorf("failed to get metadata for key %v", keyStr)
			}
		}
		return true
	})
	if c.cfgDriver == nil {
		return errors.New("failed to save store, cfgDriver is nil")
	}

	return c.cfgDriver.Save(cfgMap)
}

// Update - Only update specified settings in cfgMap in store and driver
func (c *ConfigStore) Update(cfgMap map[string]interface{}) error {
	// Update to store
	for key, value := range cfgMap {
		configValue, err := metadata.NewCfgValue(key, utils.GetStrValueOfAnyType(value))
		if err != nil {
			//todo log.Warningf("error %v, skip to update configure item, key:%v ", err, key)
			delete(cfgMap, key)
			continue
		}
		c.Set(key, *configValue)
	}
	// Update to driver
	return c.cfgDriver.Save(cfgMap)
}
