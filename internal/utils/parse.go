package utils

import (
	"fmt"
	"strconv"
	"strings"
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
