package retry_test

import (
	"net"
	"net/http"
	"net/http/httptest"

	"fmt"
	"io/ioutil"

	"strings"
	"testing"
	"time"

	"github.com/tescherm/go-http-retry"
)

// http method -> request map
var methods = map[string]string{
	"GET":     "",
	"POST":    "ping",
	"PUT":     "ping",
	"PATCH":   "ping",
	"DELETE":  "",
	"HEAD":    "",
	"CONNECT": "",
	"OPTIONS": "",
}

// happy path test - we are able to make http requests
// without a retries kicking in
func TestClient_NoRetries(t *testing.T) {
	var handler http.HandlerFunc

	tries := 0

	handler = func(w http.ResponseWriter, r *http.Request) {
		checkRequestBody(r, t)

		res := "pong"
		fmt.Fprintln(w, res)
		tries++
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client := retry.RetryingClient

	for method, reqBody := range methods {
		resBody, code, err := makeRequest(method, server.URL, reqBody, client)
		if err != nil {
			t.Errorf("%s request failed: %s", method, err)
		}

		if code != http.StatusOK {
			t.Errorf("%s request did not return : %d", method, code)
		}

		if method != "HEAD" && resBody != "pong" {
			t.Errorf("%s request did not return pong: [%s]", method, resBody)
		}

		if method == "HEAD" && resBody != "" {
			t.Errorf("expected an empty response body for HEAD: [%s]", resBody)
		}

		if tries != 1 {
			t.Errorf("expected a single try, got %d", tries)
		}

		tries = 0
	}
}

// happy path test - a retry is not triggered despite a slow server
func TestClient_RequestTimeoutSuccess(t *testing.T) {
	var handler http.HandlerFunc

	tries := 0
	wait := 100
	handler = func(w http.ResponseWriter, r *http.Request) {
		checkRequestBody(r, t)

		res := "pong"

		// sleep is less than the connection timeout default
		time.Sleep(time.Duration(wait) * time.Millisecond)
		fmt.Fprintln(w, res)

		tries++
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client := retry.RetryingClient

	for method, reqBody := range methods {
		resBody, code, err := makeRequest(method, server.URL, reqBody, client)
		if err != nil {
			t.Errorf("%s request failed: %s", method, err)
		}

		if code != http.StatusOK {
			t.Errorf("%s request did not return : %d", method, code)
		}

		if method != "HEAD" && resBody != "pong" {
			t.Errorf("%s request did not return pong: [%s]", method, resBody)
		}

		if method == "HEAD" && resBody != "" {
			t.Errorf("expected an empty response body for HEAD: [%s]", resBody)
		}

		if tries != 1 {
			t.Errorf("expected a single try, got %d", tries)
		}

		tries = 0
	}
}

// we retry on request timeout errors
func TestClient_RequestTimeoutFail(t *testing.T) {
	var handler http.HandlerFunc

	tries := 0
	wait := 100
	handler = func(w http.ResponseWriter, r *http.Request) {
		checkRequestBody(r, t)

		tries++

		res := "pong"

		// sleep is greater than the request timeout default
		time.Sleep(time.Duration(wait) * time.Millisecond)
		fmt.Fprintln(w, res)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client := retry.NewClient(&retry.ResilientTransport{
		Deadline: func() time.Time {
			return time.Now().Add(5 * time.Millisecond)
		},
		DialTimeout: 10 * time.Second,
		MaxTries:    3,
		ShouldRetry: retry.DefaultRetryPolicy,
		Backoff:     retry.ExponentialBackoff,
	})

	for method, reqBody := range methods {
		_, _, err := makeRequest(method, server.URL, reqBody, client)
		if err == nil {
			t.Errorf("%s request did not return an error", method)
		}

		switch e := err.(type) {
		case net.Error:
			if !e.Timeout() {
				t.Errorf("expected a timeout error %s", e)
			}
		default:
			t.Errorf("expected a network error %s", e)
		}

		if tries != 3 {
			t.Errorf("expected three retries, got %d", tries)
		}

		tries = 0
	}
}

// we do not retry on 4xx errors
func TestClient_No4xxRetry(t *testing.T) {
	var handler http.HandlerFunc

	tries := 0

	handler = func(w http.ResponseWriter, r *http.Request) {
		checkRequestBody(r, t)

		res := "pong"
		http.Error(w, res, http.StatusNotFound)
		tries++
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client := retry.RetryingClient

	for method, reqBody := range methods {
		resBody, code, err := makeRequest(method, server.URL, reqBody, client)
		if err != nil {
			t.Errorf("%s request failed: %s", method, err)
		}

		if code != http.StatusNotFound {
			t.Errorf("%s request did not return a 404: %d", method, code)
		}

		if method != "HEAD" && resBody != "pong" {
			t.Errorf("%s request did not return pong: [%s]", method, resBody)
		}

		if method == "HEAD" && resBody != "" {
			t.Errorf("expected an empty response body for HEAD: [%s]", resBody)
		}

		if tries != 1 {
			t.Errorf("expected a single try, got %d", tries)
		}

		tries = 0
	}
}

// fail at least once, then succeed
func TestClient_SomeRetries(t *testing.T) {
	var handler http.HandlerFunc

	failed := false
	tries := 0

	handler = func(w http.ResponseWriter, r *http.Request) {
		checkRequestBody(r, t)

		res := "pong"
		if !failed {
			http.Error(w, res, http.StatusInternalServerError)
			failed = true
		} else {
			fmt.Fprintln(w, res)
		}

		tries++
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client := retry.RetryingClient

	for method, reqBody := range methods {
		resBody, code, err := makeRequest(method, server.URL, reqBody, client)
		if err != nil {
			t.Errorf("%s request failed: %s", method, err)
		}

		if code != http.StatusOK {
			t.Errorf("%s request did not return a 200: %d", method, code)
		}

		if method != "HEAD" && resBody != "pong" {
			t.Errorf("%s request did not return pong: [%s]", method, resBody)
		}

		if method == "HEAD" && resBody != "" {
			t.Errorf("expected an empty response body for HEAD: [%s]", resBody)
		}

		if tries != 2 {
			t.Errorf("expected two tries, got %d", tries)
		}

		failed = false
		tries = 0
	}
}

// fail such that the number of retries is exceeded
func TestClient_RetriesExceeded(t *testing.T) {
	var handler http.HandlerFunc

	tries := 0

	handler = func(w http.ResponseWriter, r *http.Request) {
		checkRequestBody(r, t)

		res := "pong"
		http.Error(w, res, http.StatusInternalServerError)

		tries++
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client := retry.RetryingClient

	for method, reqBody := range methods {
		resBody, code, err := makeRequest(method, server.URL, reqBody, client)
		if err != nil {
			t.Errorf("%s request failed: %s", method, err)
		}

		if code != http.StatusInternalServerError {
			t.Errorf("%s request did not return a 500: %d", method, code)
		}

		if method != "HEAD" && resBody != "pong" {
			t.Errorf("%s request did not return pong: [%s]", method, resBody)
		}

		if method == "HEAD" && resBody != "" {
			t.Errorf("expected an empty response body for HEAD: [%s]", resBody)
		}

		if tries != 3 {
			t.Errorf("expected three tries, got %d", tries)
		}

		tries = 0
	}
}

func makeRequest(method, url, reqBody string, client *http.Client) (string, int, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(reqBody))
	if err != nil {
		return "", 0, err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}

	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", 0, err
	}

	return strings.TrimSpace(string(resBody)), res.StatusCode, nil
}

func checkRequestBody(r *http.Request, t *testing.T) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}

	bodyStr := string(reqBody)
	if bodyStr != methods[r.Method] {
		t.Fatalf("request body [%s] did not match expected for %s", bodyStr, r.Method)
	}
}
