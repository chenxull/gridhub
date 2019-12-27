package metadata

import (
	"sync"
)

var metaDataOnce sync.Once
var metaDataInstance *CfgMetaData

// Get Instance ,make it singleton
func Instance() *CfgMetaData {
	metaDataOnce.Do(func() {
		metaDataInstance = newCfgMetaData()
		metaDataInstance.init()
	})
	return metaDataInstance

}

func newCfgMetaData() *CfgMetaData {
	return &CfgMetaData{metaMap: make(map[string]Item)}
}

// CfgMetaData 中的metaMap包含了整个项目的配置信息
type CfgMetaData struct {
	metaMap map[string]Item
}

//init
func (c *CfgMetaData) init() {
	//ConfigList 是这个项目配置信息的聚集地
	c.initFromArray(ConfigList)
}

func (c *CfgMetaData) initFromArray(items []Item) {
	c.metaMap = make(map[string]Item)
	for _, item := range items {
		c.metaMap[item.Name] = item
	}
}

func (c *CfgMetaData) GetByName(name string) (*Item, bool) {
	if item, ok := c.metaMap[name]; ok {
		return &item, true
	}
	return nil, false
}

func (c *CfgMetaData) GetAll() []Item {
	metaDataList := make([]Item, 0)
	for _, value := range c.metaMap {
		metaDataList = append(metaDataList, value)
	}
	return metaDataList
}
