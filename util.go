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
		"2006-01-02T15:04:05",
		"2006-1-_2T15:04:05",
		"2006-1-_2 15:04:05.999999999Z07:00",
		"2006-1-_2 15:04:05Z07:00",
		"2006-1-_2 15:04:05",
		"2006-1-_2",
		"2006/1/_2 15:04:05Z07:00",
		"2006/1/_2 15:04:05",
		"2006/1/_2",
		"2006-01-02T15:04:05 07:00",
	}
	TimeLocation = time.Local
)


func splitDuration(s string, delim string) (string, int, error) {
	elems := strings.Split(s, delim)
	if len(elems) == 2 {
		i64, err := strconv.ParseInt(elems[0], 10, 64)
		if err != nil {
			return "", 0, err
		}
		return elems[len(elems)-1], int(i64), nil
	}

	return s, 0, nil
}


// Additional time.Duration constants
const (
	Day   = time.Hour * 24
	// Week  = Day * 7
	// Month = Day * 30
	Year  = Day * 365
)

func ParseDuration(s string) (time.Duration, error) {
	s, years, err := splitDuration(s, "y")
	if err != nil {
		return 0, err
	}
	s, days, err := splitDuration(s, "d")
	if err != nil {
		return 0, err
	}

	big := time.Duration(years)*Year +
		time.Duration(days)*Day
	if s == "" {
		return big, nil
	}

	little, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}

	return big + little, nil
}

func ToDatetime(s string) (time.Time, error) {
	for _, format := range TimeFormats {
		t, err := time.ParseInLocation(format, s, TimeLocation)
		if err == nil {
			return t, nil
		}
	}

	if strings.HasPrefix(s, "now()") {
		s = strings.TrimPrefix(s, "now()")
		s = strings.TrimSpace(s)

		if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
			hasMinus := true
			if strings.HasPrefix(s, "+") {
				hasMinus = false
				s = strings.TrimPrefix(s, "+")
			} else {
				s = strings.TrimPrefix(s, "-")
			}
			s = strings.TrimSpace(s)

			duration, err := ParseDuration(s)
			if err != nil {
				return time.Time{}, errors.New("'" + s + "' isnot duration")
			}
			if hasMinus {
				return time.Now().Add(-duration), nil
			}
			return time.Now().Add(duration), nil
		} else if s == "" {
			return time.Now(), nil
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
