# go-http-retry

## Overview

`go-http-retry` offers an http client wrapper that provides retry handling with a backoff and retry policy, as well as connection and read/write timeouts.

## Usage

`go-http-retry` has a default client that offers sensible defaults:

```go
import (
	"fmt"
	"log"
	"time"

	"github.com/tescherm/go-http-retry"
)

// retry.RetryingClient uses sensible defaults
client := retry.RetryingClient

res, err := client.Get("http://www.google.com/robots.txt")
if err != nil {
	log.Fatal(err)
}
fmt.Print(res.StatusCode)
```

however it is possible to provide custom timeout, backoff, or retry values:

```go
import (
	"fmt"
	"log"
	"time"

	"github.com/tescherm/go-http-retry"
)

// uses a default retry policy with exponential backoff
client := retry.NewClient(&retry.ResilientTransport{
	Deadline: func() time.Time {
		return time.Now().Add(5 * time.Second)
	},
	DialTimeout: 10 * time.Second,
	MaxTries:    3,
	ShouldRetry: retry.DefaultRetryPolicy,
	Backoff:     retry.ExponentialBackoff,
})

res, err := client.Get("http://www.google.com/robots.txt")
if err != nil {
	log.Fatal(err)
}
fmt.Print(res.StatusCode)
```

See the [examples](./example_test.go) for more usage.

## Backoff Policy

`go-http-retry` offers a few backoff policies out of the box:

| Policy               | Description                 |
|----------------------|-----------------------------|
| LinearBackoff        | retry every `try * 100` milliseconds   |
| ExponentialBackoff   | retry every `(try * 100)^try` seconds |

A custom backoff policy can be also be provided:

```go
import (
	"time"
)

var backoff retry.BackoffPolicyFunc
	backoff = func(try int) time.Duration {
		if try < 3 {
			return time.Duration(1) * time.Second
		}

		if try < 5 {
			return time.Duration(2) * time.Second
		}

		return time.Duration(3) * time.Second
	}
```

## Retry Policy

By default `go-http-retry` retries on temporary network errors (as determined by [AddrError.Temporary](https://golang.org/pkg/net/#AddrError.Temporary)), and 5xx errors. If necessary it is possible to provide a custom retry policy: 

```go
err := fmt.Errorf("some error")

var retryPolicy retry.RetryPolicyFunc
retryPolicy = func(req *http.Request, res *http.Response, err error) bool {
	retry := false

	if err.Error() == "some error" {
		return true
	}

	return retry
}
```

# License

MIT
