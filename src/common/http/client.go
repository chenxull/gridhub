package http

import (
	"crypto/tls"
	"github.com/chenxull/goGridhub/gridhub/src/common/http/modifier"
	"net/http"
)

// Client is a util for common HTTP operations, such Get, Head, Post, Put and Delete.

type Client struct {
	modifier []modifier.Modifier
	client   *http.Client
}

var defaultHTTPTransport, secureHTTPTransport, insecureHTTPTransport *http.Transport

func init() {
	defaultHTTPTransport = &http.Transport{}
	secureHTTPTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	insecureHTTPTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

// GetHTTPTransport returns HttpTransport based on insecure configuration
func GetHTTPTransport(insecure ...bool) *http.Transport {
	if len(insecure) == 0 {
		return defaultHTTPTransport
	}
	if insecure[0] {
		return insecureHTTPTransport
	}
	return secureHTTPTransport
}

// NewClient creates an instance of Client.
// Modifiers modify the request before sending it.
func NewClient(c *http.Client, modifiers ...modifier.Modifier) *Client {
	client := &Client{
		client: c,
	}
	if client.client == nil {
		client.client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}
	}
	if len(modifiers) > 0 {
		client.modifier = modifiers
	}
	return client
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	for _, modifier := range c.modifier {
		if err := modifier.Modify(req); err != nil {
			return nil, err
		}
	}
}
