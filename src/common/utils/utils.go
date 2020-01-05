package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

//ParseEndpoint parse endpoint to a URL
func ParseEndpoint(endpoint string) (*url.URL, error) {
	endpoint = strings.Trim(endpoint, " ")
	endpoint = strings.TrimRight(endpoint, "/")

	if len(endpoint) == 0 {
		return nil, fmt.Errorf("empty URL")
	}
	i := strings.Index(endpoint, "://")
	if i >= 0 {
		scheme := endpoint[:i]
		if scheme != "http" && scheme != "https" {
			return nil, fmt.Errorf("invalid scheme: %s", scheme)
		}
	} else {
		endpoint = "http://" + endpoint
	}
	return url.ParseRequestURI(endpoint)
}

// ParseRepository splits a repository into two parts: project and rest
func ParseRepository(repository string) (project, rest string) {
	repository = strings.TrimLeft(repository, "/")
	repository = strings.TrimRight(repository, "/")
	if !strings.ContainsRune(repository, '/') {
		rest = repository
		return
	}
	index := strings.Index(repository, "/")
	project = repository[0:index]
	rest = repository[index+1:]
	return
}

// GetStrValueOfAnyType return string format of any value, for map, need to convert to json
func GetStrValueOfAnyType(value interface{}) string {
	var strVal string
	if _, ok := value.(map[string]interface{}); ok {
		b, err := json.Marshal(value)
		if err != nil {
			log.Fatalf("can not marshal json object, error %v", err)
			return ""
		}
		strVal = string(b)
	} else {
		switch val := value.(type) {
		case float64:
			strVal = strconv.FormatFloat(val, 'f', -1, 64)
		case float32:
			strVal = strconv.FormatFloat(float64(val), 'f', -1, 32)
		default:
			strVal = fmt.Sprintf("%v", value)
		}
	}
	return strVal
}
