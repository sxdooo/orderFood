package timeutil

import "time"

var shanghai *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	shanghai = loc
}

func Now() time.Time {
	return time.Now().In(shanghai)
}

func Today() time.Time {
	n := Now()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, shanghai)
}

func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, shanghai)
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, shanghai)
}

func FormatDate(t time.Time) string {
	return t.In(shanghai).Format("2006-01-02")
}
