package driver

import (
	"errors"
	"github.com/chenxull/goGridhub/gridhub/src/common/http"
	"github.com/chenxull/goGridhub/gridhub/src/common/http/modifier"
)

// 配置信息存储在远程，需要使用client 发送请求来获取
type RESTDriver struct {
	configRESTURL string
	client        *http.Client //todo 修改http
}

// NewRESTDriver - Create RESTDriver
func NewRESTDriver(configRESTURL string, modifiers ...modifier.Modifier) *RESTDriver {
	return &RESTDriver{
		configRESTURL: configRESTURL,
		client:        http.NewClient(nil, modifiers...),
	}
}

// Load - load config data from REST server

func (h *RESTDriver) Load() (map[string]interface{}, error) {
	cfgMap := map[string]interface{}{}
	//log.Infof("get configuration from url: %+v", h.configRESTURL)
	err := h.client.Get(h.configRESTURL, &cfgMap)
	if err != nil {
		//log.Errorf("Failed on load rest config err:%v, url:%v", err, h.configRESTURL)
	}
	if len(cfgMap) < 1 {
		return cfgMap, errors.New("failed to load rest config")
	}
	return cfgMap, err
}

func (h *RESTDriver) Save(cfg map[string]interface{}) error {
	return h.client.Put(h.configRESTURL, cfgMap)
}
