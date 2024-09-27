package api

import "time"

type TimeProvider interface {
	Now() time.Time
}

type UTCTimeProvider struct{}

func (UTCTimeProvider) Now() time.Time {
	return time.Now().UTC()
}
