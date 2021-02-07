package modinputs

import (
	"time"
)

// GetEpochNow returns an the current time as Epoch, expressed in seconds with a decimal part
func GetEpochNow() float64 {
	return float64(time.Now().UnixNano()) / 1000000000.0
}

// GetEpoch returns an Epoch timestamp with millisecond precision starting from time t
func GetEpoch(t time.Time) float64 {
	return float64(t.UnixNano()) / 1000000000.0
}
