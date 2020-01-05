package registry

import (
	"github.com/chenxull/goGridhub/gridhub/src/common/http/modifier"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/logger"
	"net/http"
)

type Transport struct {
	transport http.RoundTripper
	modifiers []modifier.Modifier
}

// NewTransport ...
func NewTransport(transport http.RoundTripper, modifiers ...modifier.Modifier) *Transport {
	return &Transport{
		transport: transport,
		modifiers: modifiers,
	}
}

// RoundTrip ...
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, modifier := range t.modifiers {
		if err := modifier.Modify(req); err != nil {
			return nil, err
		}
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	logger.Debugf("%d | %s %s", resp.StatusCode, req.Method, req.URL.String())

	return resp, err
}
