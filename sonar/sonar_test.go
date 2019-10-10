// Copyright 2013 The go-sonar-sdk AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sonar

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative. It only makes a difference
	// when there's a non-empty base URL path. So, use that. See issue #752.
	apiHandler := http.NewServeMux()
	apiHandler.Handle("/"+basePath+"/", http.StripPrefix("/"+basePath, mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(os.Stderr, "FAIL: Client.BaseURL path prefix is not preserved in the request URL:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\t"+req.URL.String())
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tDid you accidentally use an absolute endpoint URL rather than relative?")
		fmt.Fprintln(os.Stderr, "\tSee https://github.com/google/go-github/issues/752 for information.")
		http.Error(w, "Client.BaseURL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// client is the GitHub client being tested and is
	// configured to use test server.
	client = NewClient(nil, "")
	url, _ := url.Parse(server.URL + "/" + basePath + "/")
	client.BaseURL = url

	return client, mux, server.URL, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testURLParseError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %+v", err)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil, "")

	if got, want := c.BaseURL.String(), defaultBaseURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
	if got, want := c.UserAgent, userAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}

	c2 := NewClient(nil, "")
	if c.client == c2.client {
		t.Error("NewClient returned same http.Clients, but they should differ")
	}
}

func TestNewRequest(t *testing.T) {
	c := NewClient(nil, "")

	inURL, outURL := "foo", defaultBaseURL+"foo"
	inBody, outBody := &Project{Organization: String("teste")}, `{"organization":"teste"}`+"\n"
	req, _ := c.NewRequest("GET", inURL, inBody)

	// test that relative URL was expanded
	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("NewRequest(%q) URL is %v, want %v", inURL, got, want)
	}

	// test that body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	if got, want := string(body), outBody; got != want {
		t.Errorf("NewRequest(%q) Body is %v, want %v", inBody, got, want)
	}

	// test that default user-agent is attached to the request
	if got, want := req.Header.Get("User-Agent"), c.UserAgent; got != want {
		t.Errorf("NewRequest() User-Agent is %v, want %v", got, want)
	}
}

func TestNewRequest_invalidJSON(t *testing.T) {
	c := NewClient(nil, "")

	type T struct {
		A map[interface{}]interface{}
	}
	_, err := c.NewRequest("GET", ".", &T{})

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*json.UnsupportedTypeError); !ok {
		t.Errorf("Expected a JSON error; got %#v.", err)
	}
}

func TestNewRequest_badURL(t *testing.T) {
	c := NewClient(nil, "")
	_, err := c.NewRequest("GET", ":", nil)
	testURLParseError(t, err)
}

// ensure that no User-Agent header is set if the client's UserAgent is empty.
// This caused a problem with Google's internal http client.
func TestNewRequest_emptyUserAgent(t *testing.T) {
	c := NewClient(nil, "")
	c.UserAgent = ""
	req, err := c.NewRequest("GET", ".", nil)
	if err != nil {
		t.Fatalf("NewRequest returned unexpected error: %v", err)
	}
	if _, ok := req.Header["User-Agent"]; ok {
		t.Fatal("constructed request contains unexpected User-Agent header")
	}
}

// If a nil body is passed to github.NewRequest, make sure that nil is also
// passed to http.NewRequest. In most cases, passing an io.Reader that returns
// no content is fine, since there is no difference between an HTTP request
// body that is an empty string versus one that is not set at all. However in
// certain cases, intermediate systems may treat these differently resulting in
// subtle errors.
func TestNewRequest_emptyBody(t *testing.T) {
	c := NewClient(nil, "")
	req, err := c.NewRequest("GET", ".", nil)
	if err != nil {
		t.Fatalf("NewRequest returned unexpected error: %v", err)
	}
	if req.Body != nil {
		t.Fatalf("constructed request contains a non-nil Body")
	}
}

func TestNewRequest_errorForNoTrailingSlash(t *testing.T) {
	tests := []struct {
		rawurl    string
		wantError bool
	}{
		{rawurl: "https://example.com/api/v3", wantError: true},
		{rawurl: "https://example.com/api/v3/", wantError: false},
	}
	c := NewClient(nil, "")
	for _, test := range tests {
		u, err := url.Parse(test.rawurl)
		if err != nil {
			t.Fatalf("url.Parse returned unexpected error: %v.", err)
		}
		c.BaseURL = u
		if _, err := c.NewRequest(http.MethodGet, "test", nil); test.wantError && err == nil {
			t.Fatalf("Expected error to be returned.")
		} else if !test.wantError && err != nil {
			t.Fatalf("NewRequest returned unexpected error: %v.", err)
		}
	}
}

func TestDo(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest("GET", ".", nil)
	body := new(foo)
	_, err := client.Do(context.Background(), req, body)

	if err != nil {
		t.Fatalf("Do returned unexpected error: %v", err)
	}

	want := &foo{"a"}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("Response body = %v, want %v", body, want)
	}
}

func TestDo_httpError(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.NewRequest("GET", ".", nil)
	resp, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Fatal("Expected HTTP 400 error, got no error.")
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected HTTP 400 error, got %d status code.", resp.StatusCode)
	}
}

// Test handling of an error caused by the internal http client's Do()
// function. A redirect loop is pretty unlikely to occur within the GitHub
// API, but does allow us to exercise the right code path.
func TestDo_redirectLoop(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath, http.StatusFound)
	})

	req, _ := client.NewRequest("GET", ".", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*url.Error); !ok {
		t.Errorf("Expected a URL error; got %#v.", err)
	}
}

func TestDo_noContent(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	var body json.RawMessage

	req, _ := client.NewRequest("GET", ".", nil)
	_, err := client.Do(context.Background(), req, &body)
	if err != nil {
		t.Fatalf("Do returned unexpected error: %v", err)
	}
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"/?a=b", "/?a=b"},
		{"/?a=b&client_secret=secret", "/?a=b&client_secret=REDACTED"},
		{"/?a=b&client_id=id&client_secret=secret", "/?a=b&client_id=id&client_secret=REDACTED"},
	}

	for _, tt := range tests {
		inURL, _ := url.Parse(tt.in)
		want, _ := url.Parse(tt.want)

		if got := sanitizeURL(inURL); !reflect.DeepEqual(got, want) {
			t.Errorf("sanitizeURL(%v) returned %v, want %v", tt.in, got, want)
		}
	}
}

func TestCheckResponse(t *testing.T) {
	res := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body: ioutil.NopCloser(strings.NewReader(`{"message":"m",
			"errors": [{"resource": "r", "field": "f", "code": "c"}],
			"block": {"reason": "dmca", "created_at": "2016-03-17T15:39:46Z"}}`)),
	}
	err := CheckResponse(res).(*ErrorResponse)

	if err == nil {
		t.Errorf("Expected error response.")
	}

	want := &ErrorResponse{
		Response: res,
		Message:  "m",
		Errors:   []Error{{Resource: "r", Field: "f", Code: "c"}},
	}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("Error = %#v, want %#v", err, want)
	}
}

// ensure that we properly handle API errors that do not contain a response body
func TestCheckResponse_noBody(t *testing.T) {
	res := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}
	err := CheckResponse(res).(*ErrorResponse)

	if err == nil {
		t.Errorf("Expected error response.")
	}

	want := &ErrorResponse{
		Response: res,
	}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("Error = %#v, want %#v", err, want)
	}
}

func TestParseBooleanResponse_true(t *testing.T) {
	result, err := parseBoolResponse(nil)
	if err != nil {
		t.Errorf("parseBoolResponse returned error: %+v", err)
	}

	if want := true; result != want {
		t.Errorf("parseBoolResponse returned %+v, want: %+v", result, want)
	}
}

func TestParseBooleanResponse_false(t *testing.T) {
	v := &ErrorResponse{Response: &http.Response{StatusCode: http.StatusNotFound}}
	result, err := parseBoolResponse(v)
	if err != nil {
		t.Errorf("parseBoolResponse returned error: %+v", err)
	}

	if want := false; result != want {
		t.Errorf("parseBoolResponse returned %+v, want: %+v", result, want)
	}
}

func TestParseBooleanResponse_error(t *testing.T) {
	v := &ErrorResponse{Response: &http.Response{StatusCode: http.StatusBadRequest}}
	result, err := parseBoolResponse(v)

	if err == nil {
		t.Errorf("Expected error to be returned.")
	}

	if want := false; result != want {
		t.Errorf("parseBoolResponse returned %+v, want: %+v", result, want)
	}
}

func TestErrorResponse_Error(t *testing.T) {
	res := &http.Response{Request: &http.Request{}}
	err := ErrorResponse{Message: "m", Response: res}
	if err.Error() == "" {
		t.Errorf("Expected non-empty ErrorResponse.Error()")
	}
}

func TestError_Error(t *testing.T) {
	err := Error{}
	if err.Error() == "" {
		t.Errorf("Expected non-empty Error.Error()")
	}
}

func TestBasicAuthTransport(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	username, password, otp := "u", "p", "123456"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok {
			t.Errorf("request does not contain basic auth credentials")
		}
		if u != username {
			t.Errorf("request contained basic auth username %q, want %q", u, username)
		}
		if p != password {
			t.Errorf("request contained basic auth password %q, want %q", p, password)
		}
	})

	tp := &BasicAuthTransport{
		Username: username,
		Password: password,
		OTP:      otp,
	}
	basicAuthClient := NewClient(tp.Client(), "")
	basicAuthClient.BaseURL = client.BaseURL
	req, _ := basicAuthClient.NewRequest("GET", ".", nil)
	_, err := basicAuthClient.Do(context.Background(), req, nil)

	if err != nil {
		t.Fatalf("Do returned unexpected error: %v", err)
	}
}

func TestBasicAuthTransport_transport(t *testing.T) {
	// default transport
	tp := &BasicAuthTransport{}
	if tp.transport() != http.DefaultTransport {
		t.Errorf("Expected http.DefaultTransport to be used.")
	}

	// custom transport
	tp = &BasicAuthTransport{
		Transport: &http.Transport{},
	}
	if tp.transport() == http.DefaultTransport {
		t.Errorf("Expected custom transport to be used.")
	}
}
