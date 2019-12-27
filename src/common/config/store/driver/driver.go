package driver

// 管理配置信息的驱动

type Driver interface {
	// Load - load config item from config driver
	Load() (map[string]interface{}, error)
	// Save - save config item into config driver
	Save(cfg map[string]interface{}) error
}
