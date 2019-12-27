package driver

import "net/http"

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

func (R RESTDriver) Load() (map[string]interface{}, error) {
	panic("implement me")
}

func (R RESTDriver) Save(cfg map[string]interface{}) error {
	panic("implement me")
}
