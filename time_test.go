package loong

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	for _, test := range [][2]string{
		[2]string{
			"2008-01-02 15:04:05",
			"2008-01-02T15:04:05+08:00",
		},
		[2]string{
			"2008-01-2 15:04:05",
			"2008-01-02T15:04:05+08:00",
		},
		[2]string{
			"2007/01/2 15:04:05",
			"2007-01-02T15:04:05+08:00",
		},
		[2]string{
			"2007/01/2",
			"2007-01-02T00:00:00+08:00",
		},
	} {
		date, err := ToDatetime(test[0])
		if err != nil {
			t.Error(err)
			return
		}
		s := date.Format(time.RFC3339Nano)
		if s != test[1] {
			t.Error("want", test[1])
			t.Error(" got", s)
		}
	}
}
