package loong

import (
	"time"

	"github.com/runner-mei/errors"
)

var (
	TimeFormats = []string{
		time.RFC3339,
		time.RFC3339Nano,
	}
	TimeLocation = time.Local
)

func ToDatetime(s string) (time.Time, error) {
	for _, format := range TimeFormats {
		t, err := time.ParseInLocation(format, s, TimeLocation)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("'" + s + "' isnot datetime")
}

type Result struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Error    *Error      `json:"error,omitempty"`
	Messages []string    `json:"messages,omitempty"`
}

func WrapErrorResult(c *Context, httpCode int, err error) interface{} {
	return &Result{Success: false, Messages: c.LogArray, Error: errors.ToError(err, httpCode)}
}

func WrapResult(c *Context, httpCode int, i interface{}) interface{} {
	return &Result{Success: true, Data: i}
}
