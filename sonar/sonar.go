// Copyright 2013 The sdk AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sonar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	hostURL        = "https://sonarcloud.io"
	userAgent      = "sonar"
	basePath       = "api"
	defaultBaseURL = hostURL + "/" + basePath + "/"
)

// A Client manages communication with the Sonar API.
type Client struct {
	client *http.Client // HTTP client used to communicate with the API.

	// Base URL for API requests. Defaults to the public Sonar API, but can be
	// set to a domain endpoint to use with Sonar Enterprise. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the Sonar API.
	UserAgent string

	// Services used for talking to different parts of the Sonar API.
	Projects *ProjectsService

	// Temporary Response
	Response *Response
}

// NewClient returns a new Sonar API client. If a nil httpClient is
// provided, a new http.Client will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the golang.org/x/oauth2 library).
func NewClient(sonarURL string, httpClient *http.Client) *Client {
	var baseURL *url.URL
	if sonarURL != "" {
		baseURL, _ = url.Parse(sonarURL)
	} else {
		baseURL, _ = url.Parse(defaultBaseURL)
	}

	c := &Client{BaseURL: baseURL, UserAgent: userAgent}

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	c.client = httpClient

	c.Projects = &ProjectsService{client: c}
	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, uri string, body string) (*http.Request, error) {
	rel, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequest(method, u.String(), bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		req.Header.Add("Content-Type", "application/json")
	}

	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// Response is a Sonar API response. This wraps the standard http.Response
// returned from Sonar and provides convenient access to things like
// pagination links.
type Response struct {
	Response   *http.Response // HTTP response
	BodyStrPtr *string
	Data       interface{}
	Meta       *ResponseMeta
}

// ResponseMeta represents information about the response. If all goes well,
// only a Code key with value 200 will present. However, sometimes things
// go wrong, and in that case ErrorType and ErrorMessage are present.
type ResponseMeta struct {
	Request string `json:"request,omitempty" xml:"request"`
	Error   string `json:"error,omitempty" xml:"error"`
}

// Do sends an API request and returns the API response. The API response is
// decoded and stored in the value pointed to by v, or returned as an error if
// an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("error closing body: %+v", err)
		}
	}()

	err = CheckResponse(resp)
	if err != nil {
		return nil, err
	}

	response := new(Response)

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	tempStr := string(bodyBytes)
	response.BodyStrPtr = &tempStr

	if v != nil {
		response.Data = v
		err = json.Unmarshal(bodyBytes, response.Data)
		c.Response = response
	}

	return response, err
}

// ErrorResponse represents a Response which contains an error
type ErrorResponse Response

// Error implements the error interface
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %s",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Meta.Error)
}

// GetStatusCode gets the status code of the error response
func (r *ErrorResponse) GetStatusCode() string {
	return fmt.Sprintf("%d", r.Response.StatusCode)
}

// GetRequestMethod gets the request method of the error response
func (r *ErrorResponse) GetRequestMethod() string {
	return fmt.Sprintf("%s", r.Response.Request.Method)
}

// GetRequestURL gets the request url of the error response
func (r *ErrorResponse) GetRequestURL() string {
	return fmt.Sprintf("%s", r.Response.Request.URL)
}

// GetSonarError gets the error message returned by Fanfou API
// if presented in the response
func (r *ErrorResponse) GetSonarError() string {
	return fmt.Sprintf("%s", r.Meta.Error)
}

// CheckResponse checks the API response for error, and returns it
// if present. A response is considered an error if it has non StatusOK
// code.
func CheckResponse(res *http.Response) error {
	if res.StatusCode == http.StatusOK {
		return nil
	}

	r := new(ErrorResponse)
	r.Response = res
	// default error message
	r.Meta = &ResponseMeta{
		Error:   "api request error",
		Request: res.Request.URL.String(),
	}

	if res.StatusCode >= http.StatusInternalServerError {
		r.Meta.Error = http.StatusText(res.StatusCode)
		return r
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		r.Meta.Error = err.Error()
	}

	if err := json.Unmarshal(data, &r.Meta); err != nil {
		r.Meta.Error = err.Error()
	}

	return r
}

// BasicAuthTransport is an http.RoundTripper that authenticates all requests
// using HTTP Basic Authentication with the provided username and password. It
// additionally supports users who have two-factor authentication enabled on
// their Sonar account.
type BasicAuthTransport struct {
	Username  string // Sonar username
	Password  string // Sonar password
	Transport http.RoundTripper
}

// Client returns an *http.Client that makes requests that are authenticated
// using HTTP Basic Authentication.
func (t *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *BasicAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// RoundTrip implements the RoundTripper interface.
func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// To set extra headers, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	//
	// Since we are going to modify only req.Header here, we only need a deep copy
	// of req.Header.
	req2 := new(http.Request)
	*req2 = *req
	req2.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		req2.Header[k] = append([]string(nil), s...)
	}

	req2.SetBasicAuth(t.Username, t.Password)

	return t.transport().RoundTrip(req2)
}
