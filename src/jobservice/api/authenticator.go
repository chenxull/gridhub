package api

import (
	"errors"
	"fmt"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/config"
	"net/http"
	"strings"
)

const (
	secretPrefix = "Harbor-Secret"
	authHeader   = "Authorization"
)

// Authenticator defined behaviors of doing auth checking.
type Authenticator interface {
	//Auth incoming request
	DoAuth(req *http.Request) error
}

type SecretAuthenticator struct {
}

func (sa *SecretAuthenticator) DoAuth(req *http.Request) error {
	if req == nil {
		return errors.New("nil request")
	}

	h := strings.TrimSpace(req.Header.Get(authHeader))
	if utils.IsEmptyStr(h) {
		return fmt.Errorf("header '%s' missing", authHeader)
	}

	if !strings.HasPrefix(h, secretPrefix) {
		return fmt.Errorf("'%s' should start with '%s'", authHeader, secretPrefix)
	}

	// 从请求中获取加密信息字段，后面的验证需要用到
	secret := strings.TrimSpace(strings.TrimPrefix(h, secretPrefix))
	// incase both two are empty
	if utils.IsEmptyStr(secret) {
		return errors.New("empty secret is not allowed")
	}
	expectedSecret := config.GetAuthSecret()
	if expectedSecret != secret {
		return errors.New("unauthorized")
	}
	return nil
}
