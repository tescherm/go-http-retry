package retry

import (
	"net"
	"net/http"
	"time"
)

// NewClient creates a new http.Client given
// a set of ResilientTransport options
func NewClient(rt *ResilientTransport) *http.Client {
	dial := func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, rt.DialTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(rt.Deadline())
		return conn, nil
	}

	rt.transport = &http.Transport{
		Dial:              dial,
		DisableKeepAlives: true,
		Proxy:             http.ProxyFromEnvironment,
	}

	return &http.Client{
		Transport: rt,
	}
}

var retryingTransport = &ResilientTransport{
	Deadline:    DefaultDeadlineFunc,
	DialTimeout: 10 * time.Second,
	MaxTries:    3,
	ShouldRetry: DefaultRetryPolicy,
	Backoff:     ExponentialBackoff,
}

// RetryingClient is a retry client with sensible defaults
var RetryingClient = NewClient(retryingTransport)
