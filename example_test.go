package retry_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tescherm/go-http-retry"
)

func ExampleDefaultClient() {
	client := retry.RetryingClient

	res, err := client.Get("http://www.google.com/robots.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(res.StatusCode)
	// Output: 200
}

func ExampleCustomClient() {
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
	// Output: 200
}

func ExampleCustomBackoff() {
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

	wait := backoff(1)
	fmt.Print(wait)
	// Output: 1s
}

func ExampleCustomRetryPolicy() {
	err := fmt.Errorf("some error")

	var retryPolicy retry.RetryPolicyFunc
	retryPolicy = func(req *http.Request, res *http.Response, err error) bool {
		retry := false

		if err.Error() == "some error" {
			return true
		}

		return retry
	}

	shouldRetry := retryPolicy(nil, nil, err)
	fmt.Print(shouldRetry)
	// Output: true
}
