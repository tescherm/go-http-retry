package retry

import (
	"net"
	"net/http"
)

// RetryPolicyFunc is a retry policy function interface.
// Given the http request, a retry policy determine if a retry should be performed.
type RetryPolicyFunc func(*http.Request, *http.Response, error) bool

// DefaultRetryPolicy provides a sensible default request retry policy.
// Note that one can provide a custom retry policy by implementing RetryPolicyFunc
func DefaultRetryPolicy(req *http.Request, res *http.Response, err error) bool {
	retry := false

	// retry if there is a temporary network error.
	if neterr, ok := err.(net.Error); ok {
		if neterr.Temporary() {
			retry = true
		}
	}

	// retry if we get a 5xx series error.
	if res != nil {
		if res.StatusCode >= 500 && res.StatusCode < 600 {
			retry = true
		}
	}

	return retry
}
