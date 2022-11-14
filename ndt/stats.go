package ndt

import "time"

type Stats struct {
	BytesReceived uint64
	StartTime time.Time
	ElapsedTime time.Duration
}