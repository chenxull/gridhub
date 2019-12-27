package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gocraft/work"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type NodeIDContextKey string

const (
	NodeID NodeIDContextKey = "node_id"
)

// MakeIdentifier creates uuid for job.
func MakeIdentifier() string {
	b := make([]byte, 12)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}
func IsEmptyStr(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

func ReadEnv(key string) string {
	return os.Getenv(key)
}

// FileExists check if the specified exists.
func FileExists(file string) bool {
	if !IsEmptyStr(file) {
		_, err := os.Stat(file)
		if err == nil {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}

		return true
	}

	return false
}

// DirExists check if the specified dir exists
func DirExists(path string) bool {
	if IsEmptyStr(path) {
		return false
	}

	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return f.IsDir()
}

// IsValidPort check if port is valid.
func IsValidPort(port uint) bool {
	return port != 0 && port < 65536
}

// IsValidURL validates if the url is well-formted
func IsValidURL(address string) bool {
	if IsEmptyStr(address) {
		return false
	}

	if _, err := url.Parse(address); err != nil {
		return false
	}

	return true
}

func TranslateRedisAddress(commaFormat string) (string, bool) {
	if IsEmptyStr(commaFormat) {
		return "", false
	}
	sections := strings.Split(commaFormat, ",")
	totalSections := len(sections)
	if totalSections == 0 {
		return "", false
	}

	urlParts := make([]string, 0)
	// section[0] should be host:port
	redisURL := fmt.Sprintf("redis://%s", sections[0])
	if _, err := url.Parse(redisURL); err != nil {
		return "", false
	}
	urlParts = append(urlParts, "redis://", sections[0])
	// Ignore weight
	// Check password
	if totalSections >= 3 && !IsEmptyStr(sections[2]) {
		urlParts = []string{urlParts[0], fmt.Sprintf("%s:%s@", "arbitrary_username", sections[2]), urlParts[1]}
	}

	if totalSections >= 4 && !IsEmptyStr(sections[3]) {
		if _, err := strconv.Atoi(sections[3]); err == nil {
			urlParts = append(urlParts, "/", sections[3])
		}
	}

	return strings.Join(urlParts, ""), true
}

func ResolveHostNameAndIp() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return fmt.Sprintf("%s:%s", host, ipnet.IP.String()), nil
			}
		}
	}
	return "", errors.New("failed to resolve local host&ip")

}

func GenerateNodeID() string {
	hIP, err := ResolveHostNameAndIp()
	if err != nil {
		return MakeIdentifier()
	}
	return hIP
}

// SerializeJob encodes work.Job to json data.
func SerializeJob(job *work.Job) ([]byte, error) {
	return json.Marshal(job)
}

// DeSerializeJob decodes bytes to ptr of work.Job.
func DeSerializeJob(jobBytes []byte) (*work.Job, error) {
	var j work.Job
	err := json.Unmarshal(jobBytes, &j)

	return &j, err
}
