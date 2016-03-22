package retry

import (
	"time"
)

// DeadlineFunc is a Deadline (read/write timeout) function interface
type DeadlineFunc func() time.Time

// DefaultDeadlineFunc provides a sensible default deadline.
// Note that one can use DeadlineFunc to implement their own deadline.
func DefaultDeadlineFunc() time.Time {
	return time.Now().Add(5 * time.Second)
}
