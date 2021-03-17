package loong

import (
	"time"

	"github.com/runner-mei/errors"
)

var (
	TimeFormats = []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-_1-_2 15:04:05",
		"2006/_1/_2 15:04:05",
		"2006-_1-_2",
		"2006/_1/_2",
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

func ResultWrapMiddleware(okResult func(c *Context, code int, i interface{}) interface{},
	errorResult func(ctx *Context, code int, err error) interface{}) MiddlewareFunc {
	return MiddlewareFunc(func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			ctx.WrapOkResult = okResult
			ctx.WrapErrorResult = errorResult
			return next(ctx)
		}
	})
}
