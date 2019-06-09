package scheduling

import (
	"errors"
	"fmt"
	"time"
)

func FirstDayOfWeek(date time.Time) time.Time {
	var firstDay time.Time
	weekDay := int(date.Weekday())
	if weekDay == 0 {
		// Sunday -> Add one more day
		firstDay = date.AddDate(0, 0, 1)
	} else if weekDay == 1 {
		firstDay = date
	} else {
		firstDay = date.AddDate(0, 0, -(weekDay - 1))
	}

	return firstDay
}

func WeekDescription(date time.Time) string {
	_, weekNo := date.ISOWeek()

	beginWeekMonth := date.Month().String()
	beginWeekDay := date.Day()
	endWeek := date.AddDate(0, 0, 4)
	endWeekMonth := endWeek.Month().String()
	endWeekDay := endWeek.Day()
	return fmt.Sprintf("Week %d : %s %d - %s %d", weekNo, beginWeekMonth, beginWeekDay, endWeekMonth, endWeekDay)
}

func WeekDescriptionFromWeekNo(weekNo, year int) (string, error) {
	if weekNo == 0 {
		return "General", nil
	}
	date, err := TimeFromWeekNo(weekNo, year)
	if err != nil {
		return "", err
	}

	date = FirstDayOfWeek(date)
	return WeekDescription(date), nil
}

func TimeFromWeekNo(weekNo, year int) (time.Time, error) {
	if weekNo <= 0 || weekNo > 52 {
		return time.Time{}, errors.New("Invalid week")
	}
	if year < 0 {
		return time.Time{}, errors.New("Invalid year")
	}
	day := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	newDay := day.AddDate(0, 0, 7*(weekNo-1))
	return newDay, nil
}
