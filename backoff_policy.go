package retry

import (
	"math"
	"time"
)

// BackoffPolicyFunc is backoff function interface.
// Examples of backoff functions include linear
// backoff or exponential backoff.
type BackoffPolicyFunc func(try int) time.Duration

// ExponentialBackoff implements an exponential backoff function
func ExponentialBackoff(try int) time.Duration {
	return 100 * time.Millisecond *
		time.Duration(math.Exp2(float64(try)))
}

// LinearBackoff implements a linear backoff function
func LinearBackoff(try int) time.Duration {
	return time.Duration(try*100) * time.Millisecond
}
