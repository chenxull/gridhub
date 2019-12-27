package rds

import (
	"github.com/chenxull/goGridhub/gridhub/src/jobservice/common/utils"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"time"
)

// ErrNoElements is a pre defined error to describe the case that no elements got
// from the backend database.
var ErrNoElements = errors.New("no elements got from the backend")

func HmSet(conn redis.Conn, key string, fieldAndValues ...interface{}) error {
	if conn == nil {
		return errors.New("nil redis connection")
	}

	if utils.IsEmptyStr(key) {
		return errors.New("no key specified to do HMSET")
	}

	if len(fieldAndValues) == 0 {
		return errors.New("no properties specified to do HMSET")
	}

	args := make([]interface{}, 0, len(fieldAndValues)+2)

	args = append(args, key)
	args = append(args, fieldAndValues...)
	args = append(args, "update_time", time.Now().Unix()) // Add update timestamp
	_, err := conn.Do("HMSET", args...)
	return err
}

// HmGet gets values of multiple fields
// Values have same order with the provided fields
func HmGet(conn redis.Conn, key string, fields ...interface{}) ([]interface{}, error) {
	if conn == nil {
		return nil, errors.New("nil redis connection")
	}

	if utils.IsEmptyStr(key) {
		return nil, errors.New("no key specified to do HMGET")
	}

	if len(fields) == 0 {
		return nil, errors.New("no fields specified to do HMGET")
	}

	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, key)
	args = append(args, fields...)

	return redis.Values(conn.Do("HMGET", args...))
}

// JobScore represents the data item with score in the redis db.
type JobScore struct {
	JobBytes []byte
	Score    int64
}

// ZPopMin pops the element with lowest score in the zset
func ZPopMin(conn redis.Conn, key string) (interface{}, error) {
	err := conn.Send("MULTI")
	err = conn.Send("ZRANGE", key, 0, 0) // lowest one
	err = conn.Send("ZREMRANGEBYRANK", key, 0, 0)
	if err != nil {
		return nil, err
	}

	replies, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return nil, err
	}

	if len(replies) < 2 {
		return nil, errors.Errorf("zpopmin error: not enough results returned, expected %d but got %d", 2, len(replies))
	}

	zrangeReply := replies[0]
	if zrangeReply != nil {
		if elements, ok := zrangeReply.([]interface{}); ok {
			if len(elements) == 0 {
				return nil, ErrNoElements
			}

			return elements[0], nil
		}
	}

	return nil, errors.New("zpopmin error: bad result reply")
}
