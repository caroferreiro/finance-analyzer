package timeale

import (
	"fmt"
	"time"
)

func NowStrWithTimeZone() string {
	tz := "America/Buenos_Aires"
	loc, err := time.LoadLocation(tz)
	if err != nil {
		panic(fmt.Sprintf("error loading timezone location: %v", err))
	}
	now := time.Now().In(loc)
	nowStr := now.Format("2006-01-02 15:04:05")
	nowStrWithTZ := fmt.Sprintf("%s %s", nowStr, tz)
	return nowStrWithTZ
}

func ToLocalDateString(date time.Time) string {
	loc, err := time.LoadLocation("America/Buenos_Aires")
	if err != nil {
		panic(fmt.Sprintf("error loading timezone location: %v", err))
	}
	return date.In(loc).Format("2006-01-02")
}

func ToLocalDateTimeString(date time.Time) string {
	loc, err := time.LoadLocation("America/Buenos_Aires")
	if err != nil {
		panic(fmt.Sprintf("error loading timezone location: %v", err))
	}
	return date.In(loc).Format("2006-01-02 15:04:05")
}

func DaysBetweenDates(date1, date2 time.Time) int {
	return int(date2.Sub(date1).Hours() / 24)
}

func MiddleOfNextMonth(t time.Time) time.Time {
	month := t.Month()
	year := t.Year()
	tmpTime := time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	nextMonth := tmpTime.AddDate(0, 1, 0)
	return nextMonth.AddDate(0, 0, 14)
}
