package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var extractDurationExpr = regexp.MustCompile(`(\d+)\s*(m|min|minutes?|мин|минут[\wа-я]{0,3})(\s|[^\wа-я]|$)`)

func ExtractDuration(text string) (time.Duration, error) {
	matches := extractDurationExpr.FindStringSubmatch(text)

	if len(matches) != 4 {
		return 0, fmt.Errorf("duration not found in text")
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid number: %v", err)
	}

	return time.Duration(value) * time.Minute, nil
}
