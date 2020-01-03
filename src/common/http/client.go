package http

import (
	"bytes"
	"crypto/tls"
	"errors"

	"encoding/json"
	"github.com/chenxull/goGridhub/gridhub/src/common/http/modifier"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
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
	return c.client.Do(req)
}

//Get...
func (c *Client) Get(url string, v ...interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	data, err := c.do(req)
	if err != nil {
		return err
	}

	if len(v) == 0 {
		return nil
	}

	return json.Unmarshal(data, v[0])

}

// Head ...
func (c *Client) Head(url string) error {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

//Post
func (c *Client) Post(url string, v ...interface{}) error {
	var reader io.Reader
	if len(v) > 0 {
		if r, ok := v[0].(io.Reader); ok {
			reader = r
		} else {
			data, err := json.Marshal(v[0])
			if err != nil {
				return err
			}
			reader = bytes.NewReader(data)
		}

	}
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.do(req)
	return err
}

// Put ...
func (c *Client) Put(url string, v ...interface{}) error {
	var reader io.Reader
	if len(v) > 0 {
		data := []byte{}
		data, err := json.Marshal(v[0])
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(http.MethodPut, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.do(req)
	return err
}

// Delete ...
func (c *Client) Delete(url string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}
func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}
	return data, nil
}

// GetAndIteratePagination iterates the pagination header and returns all resources
// The parameter "v" must be a pointer to a slice
func (c *Client) GetAndIteratePagination(endpoint string, v interface{}) error {
	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v should be a pointer to a slice")
	}
	elemType := rv.Elem().Type()
	if elemType.Kind() != reflect.Slice {
		return errors.New("v should be a pointer to a slice")
	}

	resources := reflect.Indirect(reflect.New(elemType))
	for len(endpoint) > 0 {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return &Error{
				Code:    resp.StatusCode,
				Message: string(data),
			}
		}

		res := reflect.New(elemType)
		if err = json.Unmarshal(data, res.Interface()); err != nil {
			return err
		}
		resources = reflect.AppendSlice(resources, reflect.Indirect(res))

		endpoint = ""
		link := resp.Header.Get("Link")
		for _, str := range strings.Split(link, ",") {
			if strings.HasSuffix(str, `rel="next"`) &&
				strings.Index(str, "<") >= 0 &&
				strings.Index(str, ">") >= 0 {
				endpoint = url.Scheme + "://" + url.Host + str[strings.Index(str, "<")+1:strings.Index(str, ">")]
				break
			}
		}
	}
	rv.Elem().Set(resources)
	return nil
}
