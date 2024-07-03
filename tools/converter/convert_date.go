package converter

import (
	"time"
)

func Convert(date string) string {
	t, _ := time.Parse(time.RFC3339, date)
	date = t.Format(time.DateOnly)
	return date
}
