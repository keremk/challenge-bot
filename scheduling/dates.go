package scheduling

import (
	"fmt"
	"time"
)

func FirstDayOfWeek(day time.Time) time.Time {
	var firstDay time.Time
	weekDay := int(day.Weekday())
	if weekDay == 0 {
		// Sunday -> Add one more day
		firstDay = day.AddDate(0, 0, 1)
	} else if weekDay == 1 {
		firstDay = day
	} else {
		firstDay = day.AddDate(0, 0, -(weekDay - 1))
	}

	return firstDay
}

func WeekDescription(week time.Time) string {
	_, weekNo := week.ISOWeek()

	beginWeekMonth := week.Month().String()
	beginWeekDay := week.Day()
	endWeek := week.AddDate(0, 0, 4)
	endWeekMonth := endWeek.Month().String()
	endWeekDay := endWeek.Day()
	return fmt.Sprintf("Week %d : %s %d - %s %d", weekNo, beginWeekMonth, beginWeekDay, endWeekMonth, endWeekDay)
}
