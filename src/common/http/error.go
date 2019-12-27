package http

import (
	"encoding/json"
	"fmt"
)

// Error wrap HTTP status code and message as an error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error ...
func (e *Error) Error() string {
	return fmt.Sprintf("http error: code %d, message %s", e.Code, e.Message)
}

// String wraps the error msg to the well formatted error message
func (e *Error) String() string {
	data, err := json.Marshal(&e)
	if err != nil {
		return e.Message
	}
	return string(data)
}
