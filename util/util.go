package util

import (
	"time"
)

func ParseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}
