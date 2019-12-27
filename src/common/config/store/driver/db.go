package driver

// 数据库驱动，用来讲配置信息存储在数据库
type Database struct {
}

func (d *Database) Load() (map[string]interface{}, error) {
	panic("implement me")
}

func (d *Database) Save(cfg map[string]interface{}) error {
	panic("implement me")
}
