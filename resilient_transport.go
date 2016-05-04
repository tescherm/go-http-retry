package retry

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

type OnRetryFunc func(req *http.Request, res *http.Response, err error, try int)

// ResilientTransport is a http.Transport wrapper that provides
// retry handling with a backoff and retry policy, as well as
// connection and read/write timeouts
type ResilientTransport struct {
	Deadline    DeadlineFunc
	DialTimeout time.Duration
	MaxTries    int
	ShouldRetry RetryPolicyFunc
	Backoff     BackoffPolicyFunc
	OnRetry     OnRetryFunc

	transport *http.Transport
}

// RoundTrip implements the http.RoundTripper interface.
func (t *ResilientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.tries(req)
}

// retry the request up to t.MaxTries with a backoff policy, if specified
func (t *ResilientTransport) tries(req *http.Request) (*http.Response, error) {
	try := 1

	var orig []byte
	var err error

	// if the request has a body, maintain
	// the original request body to use for each retry request
	if req.Body != nil {
		orig, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		defer req.Body.Close()
	}

	for {
		// the request body is consumed on each retry request
		// assuming that the request has a body, let's provide
		// a new reader for each retry
		if req.Body != nil {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(orig))
		}

		res, err := t.transport.RoundTrip(req)

		// either we have exceeded the max number of retry attempts,
		// or the request was successful
		if try == t.MaxTries || !t.ShouldRetry(req, res, err) {
			return res, err
		}

		// invoke on retry callback
		onRetry := t.OnRetry
		if onRetry != nil {
			onRetry(req, res, err, try)
		}

		if res != nil {
			// close the response reader. Note that we only do
			// this if this is a retry request
			res.Body.Close()
		}
		if t.Backoff != nil {
			time.Sleep(t.Backoff(try))
		}

		try++
	}
}
