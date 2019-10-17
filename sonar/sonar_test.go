// Copyright 2013 The go-sonar-sdk AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sonar

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the Fanfou client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// mock base url
	mockBaseURL, _ := url.Parse(server.URL)

	// Fanfou client configured to use test server
	client = NewClient("", nil)
	client.BaseURL = mockBaseURL
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("", nil)

	want := defaultBaseURL
	if c.BaseURL.String() != want {
		t.Errorf("NewClient BaseURL = %v, want %v", c.BaseURL.String(), want)
	}
	want = userAgent
	if c.UserAgent != want {
		t.Errorf("NewClient UserAgent = %v, want %v", c.UserAgent, want)
	}
}

func TestNewRequest(t *testing.T) {
	c := NewClient("", nil)

	inURL, outURL := "projects/search", c.BaseURL.String()+"projects/search"
	req, _ := c.NewRequest("GET", inURL, "")

	// test that relative URL was expanded and access token appears in query string
	if req.URL.String() != outURL {
		t.Errorf("NewRequest(%v) URL = %v, want %v", inURL, req.URL, outURL)
	}

	// test that default user-agent is attached to the requet
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("NewRequest() User-Agent = %v, want %v", userAgent, c.UserAgent)
	}
}

func TestCheckResponse(t *testing.T) {
	mockRes := http.Response{
		StatusCode: http.StatusOK,
	}
	err := CheckResponse(&mockRes)
	if err != nil {
		t.Errorf("CheckResponse() while 200, result %v, want %v", err, nil)
	}

	mockRes = http.Response{
		StatusCode: http.StatusBadRequest,
		Request: &http.Request{
			Method: http.MethodPost,
			URL: &url.URL{
				Scheme: "https",
				Host:   "test.url.com",
				Path:   "/",
			},
		},
		Body: ioutil.NopCloser(bytes.NewBufferString(`{"error":"test_error", "request": "test_request"}`)),
	}

	err = CheckResponse(&mockRes)
	if err == nil {
		t.Errorf("CheckResponse() while 400, result %v, want err", err)
	}

	want := ""
	actual := err.Error()
	if reflect.TypeOf(actual) != reflect.TypeOf(want) {
		t.Errorf("CheckResponse() while 400, err.Error() type is %v, want %v", reflect.TypeOf(actual), reflect.TypeOf(want))
	}

	fanfouErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Errorf("CheckResponse() while 400, error is not ErrorResponse, want ErrorResponse")
	}

	want = "400"
	actual = fanfouErr.GetStatusCode()
	if want != actual {
		t.Errorf("CheckResponse() while 400, fanfouErr.GetStatusCode() is %v, want %v", actual, want)
	}

	want = "POST"
	actual = fanfouErr.GetRequestMethod()
	if want != actual {
		t.Errorf("CheckResponse() while 400, fanfouErr.GetRequestMethod() is %v, want %v", actual, want)
	}

	want = "https://test.url.com/"
	actual = fanfouErr.GetRequestURL()
	if want != actual {
		t.Errorf("CheckResponse() while 400, fanfouErr.GetRequestURL() is %v, want %v", actual, want)
	}

	want = "test_error"
	actual = fanfouErr.GetSonarError()
	if want != actual {
		t.Errorf("CheckResponse() while 400, fanfouErr.GetFanfouError() is %v, want %v", actual, want)
	}
}
