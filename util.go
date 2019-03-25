package loong

import (
	"errors"
	"time"
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
