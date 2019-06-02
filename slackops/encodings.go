package slackops

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func encodeWeekAndYear(weekNo, year int) string {
	return fmt.Sprintf("%d-%d", weekNo, year)
}

func decodeWeekAndYear(input string) (int, int) {
	log.Println("[INFO] Incoming encoded input - ", input)
	s := strings.Split(input, "-")
	if len(s) < 2 {
		return 0, time.Now().Year()
	}
	year, err := strconv.Atoi(s[1])
	if err != nil || year < 2019 || year > 2050 {
		return 0, time.Now().Year()
	}
	week, err := strconv.Atoi(s[0])
	if err != nil || week < 0 || week > 52 {
		return 0, time.Now().Year()
	}

	return week, year
}

type scheduleActionInfo struct {
	SlotID     string
	ReviewerID string
	WeekNo     int
	Year       int
}

func encodeScheduleActionInfo(input scheduleActionInfo) string {
	return fmt.Sprintf("%s-%s-%d-%d", input.SlotID, input.ReviewerID, input.WeekNo, input.Year)
}

func decodeScheduleActionInfo(actionID string) (scheduleActionInfo, error) {
	s := strings.Split(actionID, "-")
	if len(s) < 4 {
		return scheduleActionInfo{}, errors.New("[ERROR] Encoding for actionID is not correct")
	}

	weekNo, err := strconv.Atoi(s[2])
	if err != nil {
		return scheduleActionInfo{}, err
	}

	year, err := strconv.Atoi(s[3])
	if err != nil {
		return scheduleActionInfo{}, err
	}

	return scheduleActionInfo{
		SlotID:     s[0],
		ReviewerID: s[1],
		WeekNo:     weekNo,
		Year:       year,
	}, nil
}
