package loong

import (
	"strconv"
	"strings"
	"time"

	"github.com/runner-mei/errors"
)

func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func ToBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "on" || s == "yes" || s == "enabled"
}

func ToInt64Array(array []string) ([]int64, error) {
	var int64Array []int64
	for _, s := range array {
		ss := strings.Split(s, ",")
		for _, v := range ss {
			if v == "" {
				continue
			}
			i64, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, err
			}
			int64Array = append(int64Array, i64)
		}
	}
	return int64Array, nil
}

func ToBoolArray(array []string) ([]bool, error) {
	var boolArray []bool
	for _, s := range array {
		ss := strings.Split(s, ",")
		for _, v := range ss {
			if v == "" {
				continue
			}

			switch v {
			case "TRUE", "True", "true", "YES", "Yes", "yes", "on", "enabled":
				boolArray = append(boolArray, true)
			case "FALSE", "False", "false", "NO", "No", "no", "off":
				boolArray = append(boolArray, false)
			default:
				return nil, errors.New("convert '" + v + "' to bool failure")
			}
		}
	}
	return boolArray, nil
}

var (
	TimeFormats = []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-1-_2 15:04:05.999999999Z07:00",
		"2006-1-_2 15:04:05Z07:00",
		"2006-1-_2 15:04:05",
		"2006-1-_2",
		"2006/1/_2 15:04:05Z07:00",
		"2006/1/_2 15:04:05",
		"2006/1/_2",
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

func ToDatetimes(array []string) ([]time.Time, error) {
	var timeArray []time.Time
	for _, s := range array {
		ss := strings.Split(s, ",")
		for _, v := range ss {
			if v == "" {
				continue
			}
			t, err := ToDatetime(v)
			if err != nil {
				return nil, err
			}
			timeArray = append(timeArray, t)
		}
	}
	return timeArray, nil
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
