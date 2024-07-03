package dateFormatter

import "time"

func FirstDayOfWeek() string {
	date := time.Now()
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
	}

	format := date.Format(time.DateOnly)

	return format
}

func LastDayOfWeek() string {
	date := time.Now()
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, 1)
	}

	format := date.Format(time.DateOnly)

	return format
}
