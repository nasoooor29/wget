package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseRateLimit(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	value = strings.TrimSpace(value)

	multiplier := int64(1)

	last := value[len(value)-1]
	switch last {
	case 'k', 'K':
		multiplier = 1024
		value = value[:len(value)-1]
	case 'm', 'M':
		multiplier = 1024 * 1024
		value = value[:len(value)-1]
	}

	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rate limit: %s", value)
	}

	return number * multiplier, nil
}
func FormatBytes(value int64) string {
	if value < 1024 {
		return fmt.Sprintf("%d B", value)
	}

	units := []string{"KiB", "MiB", "GiB", "TiB"}
	current := float64(value)
	for _, unit := range units {
		current /= 1024
		if current < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", current, unit)
		}
	}

	return fmt.Sprintf("%d B", value)
}

func FormatDuration(value time.Duration) string {
	if value < 0 {
		value = 0
	}

	totalSeconds := int64(value.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
