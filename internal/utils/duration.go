package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var extractMinutesDurationExpr = regexp.MustCompile(`(\d+)\s*(m|min|minutes?|м|мин|минут[\wа-я]{0,3})(\s|[^\wа-я]|$)`)
var extractHoursDurationExpr = regexp.MustCompile(`(\d+)\s*(h|hours?|ч|час[\wа-я]{0,3})(\s|[^\wа-я]|$)`)

func ExtractDuration(text string) (duration time.Duration) {
	var matches []string

	matches = extractMinutesDurationExpr.FindStringSubmatch(text)
	if len(matches) == 4 {
		if value, err := strconv.Atoi(matches[1]); err == nil {
			duration += time.Duration(value) * time.Minute
		}
	}

	matches = extractHoursDurationExpr.FindStringSubmatch(text)
	if len(matches) == 4 {
		if value, err := strconv.Atoi(matches[1]); err == nil {
			duration += time.Duration(value) * time.Hour
		}
	}

	return
}

func FormatDurationToString(duration time.Duration) (result string) {
	durationMinutes := int(duration.Minutes())

	hours := durationMinutes / 60
	minutes := durationMinutes % 60

	if hours > 0 {
		result += fmt.Sprintf("%dh", hours)
	}
	if hours > 0 && minutes > 0 {
		result += " "
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm", minutes)
	}

	return
}
