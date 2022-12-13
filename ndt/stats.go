package ndt

import "time"

type Stats struct {
	TransferKind  TransferKind
	BytesReceived uint64
	StartTime     time.Time
	ElapsedTime   time.Duration
}
