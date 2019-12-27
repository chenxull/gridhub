package hook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

// Client for handing the hook events
type Client interface {
	SendEvent(evt *Event) error
}

// Client is used to post the related data to the interested parties.
type basicClient struct {
	client *http.Client
	ctx    context.Context
}

//NewClient return the ptr of the new hook client
func NewClient(ctx context.Context) Client {
	// Create transport
	transport := &http.Transport{
		MaxIdleConns:    20,
		IdleConnTimeout: 30 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Proxy:                 http.ProxyFromEnvironment,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
	return &basicClient{
		client: client,
		ctx:    ctx,
	}
}

// ReportStatus reports the status change info to the subscribed party.
// The status includes 'checkin' info with format 'check_in:<message>'
func (bc *basicClient) SendEvent(evt *Event) error {
	if evt == nil {
		return errors.New("nil event")
	}

	if err := evt.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(evt.Data)
	if err != nil {
		return nil
	}
	//New Post request
	req, err := http.NewRequest(http.MethodPost, evt.URL, strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	res, err := bc.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}() // close connection for reuse
	// Should be 200
	if res.StatusCode != http.StatusOK {
		if res.ContentLength > 0 {
			// read error content and return
			dt, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}
			return errors.New(string(dt))
		}

		return fmt.Errorf("failed to report status change via hook, expect '200' but got '%d'", res.StatusCode)
	}

	return nil
}
