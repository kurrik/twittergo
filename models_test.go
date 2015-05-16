// Copyright 2013 Arne Roomann-Kurrik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package twittergo

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"
)

type Body struct {
	*bytes.Buffer
}

func NewBody(body string) *Body {
	return &Body{
		Buffer: bytes.NewBufferString(body),
	}
}

func (b *Body) Close() error {
	return nil
}

func getResponse(code int, body string) *http.Response {
	var resp = &http.Response{}
	resp.Body = NewBody(body)
	resp.Header = http.Header(map[string][]string{})
	resp.StatusCode = code
	return resp
}

func TestRateLimitError(t *testing.T) {
	// Setup
	var body = `{"errors":[{"message":"Rate limit exceeded","code":88}]}`
	var resp = getResponse(429, body)
	resp.Status = "Too Many Requests"
	resp.Header = http.Header(map[string][]string{
		"Content-Length":         []string{"56"},
		"X-Rate-Limit-Limit":     []string{"180"},
		"X-Rate-Limit-Remaining": []string{"0"},
		"X-Rate-Limit-Reset":     []string{"1369331745"},
	})

	// Test
	var (
		api_resp *APIResponse
		tweet    *Tweet
		err      error
		rle      RateLimitError
		ok       bool
	)
	api_resp = (*APIResponse)(resp)
	tweet = &Tweet{}
	err = api_resp.Parse(tweet)
	if err == nil {
		t.Fatalf("Expected an error in Parse")
	}
	if rle, ok = err.(RateLimitError); !ok {
		t.Fatalf("Expected a RateLimitError error")
	}
	if !rle.Reset.Equal(time.Unix(1369331745, 0)) {
		t.Errorf("Reset not parsed correctly, got %v", rle.Reset)
	}
	if rle.Remaining != 0 {
		t.Errorf("Remaining not parsed correctly, got %v", rle.Remaining)
	}
	if rle.Limit != 180 {
		t.Errorf("Limit not parsed correctly, got %v", rle.Limit)
	}
}

func TestErrorsError(t *testing.T) {
	// Setup
	var err1 = `{"code":187,"message":"Status is a duplicate"}`
	var err2 = `{"message":"Rate limit exceeded","code":88}`
	var body = fmt.Sprintf(`{"errors":[%v,%v]}`, err1, err2)
	var resp = getResponse(403, body)
	resp.Status = "Forbidden"
	resp.Header.Set("Content-Length", fmt.Sprintf("%v", len(body)))

	// Test
	var (
		api_resp *APIResponse
		tweet    *Tweet
		err      error
		errs     Errors
		ok       bool
		code     int64
		msg      string
	)
	api_resp = (*APIResponse)(resp)
	tweet = &Tweet{}
	err = api_resp.Parse(tweet)
	if err == nil {
		t.Fatalf("Expected an error in Parse")
	}
	if errs, ok = err.(Errors); !ok {
		t.Fatalf("Expected a RateLimitError error")
	}
	if len(errs.Errors()) != 2 {
		t.Fatalf("Expected 2 errors to be parsed")
	}
	code = errs.Errors()[0].Code()
	if code != 187 {
		t.Errorf("Expected 187, got %v", code)
	}
	msg = errs.Errors()[0].Message()
	if msg != "Status is a duplicate" {
		t.Errorf("Got incorrect dupe status text: %v", msg)
	}
	code = errs.Errors()[1].Code()
	if code != 88 {
		t.Errorf("Expected 88, got %v", code)
	}
	msg = errs.Errors()[1].Message()
	if msg != "Rate limit exceeded" {
		t.Errorf("Got incorrect rle text: %v", msg)
	}
}

func TestNonJSONErrorWith500(t *testing.T) {
	// Setup
	var body = `<!DOCTYPE html><html><body>Foo</body></html>`
	var resp = getResponse(500, body)
	resp.Status = "Server Error"
	resp.Header.Set("Content-Length", fmt.Sprintf("%v", len(body)))

	// Test
	var (
		api_resp *APIResponse
		tweet    *Tweet
		err      error
		rerr     ResponseError
		ok       bool
		expected string
	)
	api_resp = (*APIResponse)(resp)
	tweet = &Tweet{}
	err = api_resp.Parse(tweet)
	if err == nil {
		t.Fatalf("Expected an error in Parse")
	}
	if _, ok = err.(RateLimitError); ok {
		t.Fatalf("Error should not parse to a RateLimitError error")
	}
	if _, ok = err.(Errors); ok {
		t.Fatalf("Error should not parse to a Errors error (ugh)")
	}
	expected = fmt.Sprintf("Unable to handle response (status code 500): `%v`", body)
	if err.Error() != expected {
		t.Fatalf("Rendered error string `%s`, expected `%s`", err.Error(), expected)
	}
	if rerr, ok = err.(ResponseError); !ok {
		t.Fatalf("Error should be castable to a ResponseError")
	}
	if rerr.Code != 500 {
		t.Errorf("ResponseError code should be 500, got %d", rerr.Code)
	}
	if rerr.Body != body {
		t.Errorf("ResponseError body should be `%s`, got `%s`", body, rerr.Body)
	}
}

func TestNonJSONErrorWith502(t *testing.T) {
	// Setup
	var body = `<!DOCTYPE html><html><body>Foo</body></html>`
	var resp = getResponse(502, body)
	resp.Status = "Bad Gateway"
	resp.Header.Set("Content-Length", fmt.Sprintf("%v", len(body)))

	// Test
	var (
		api_resp *APIResponse
		tweet    *Tweet
		err      error
		rerr     ResponseError
		ok       bool
		expected string
	)
	api_resp = (*APIResponse)(resp)
	tweet = &Tweet{}
	err = api_resp.Parse(tweet)
	if err == nil {
		t.Fatalf("Expected an error in Parse")
	}
	if _, ok = err.(RateLimitError); ok {
		t.Fatalf("Error should not parse to a RateLimitError error")
	}
	if _, ok = err.(Errors); ok {
		t.Fatalf("Error should not parse to a Errors error (ugh)")
	}
	expected = fmt.Sprintf("Unable to handle response (status code 502): `%v`", body)
	if err.Error() != expected {
		t.Fatalf("Rendered error string `%s`, expected `%s`", err.Error(), expected)
	}
	if rerr, ok = err.(ResponseError); !ok {
		t.Fatalf("Error should be castable to a ResponseError")
	}
	if rerr.Code != 502 {
		t.Errorf("ResponseError code should be 502, got %d", rerr.Code)
	}
	if rerr.Body != body {
		t.Errorf("ResponseError body should be `%s`, got `%s`", body, rerr.Body)
	}
}

func TestEmptyErrorWith403(t *testing.T) {
	// Setup
	var body = ``
	var resp = getResponse(403, body)
	resp.Status = "Forbidden"
	resp.Header.Set("Content-Length", "0")

	// Test
	var (
		api_resp *APIResponse
		tweet    *Tweet
		err      error
		rerr     ResponseError
		ok       bool
		expected string
	)
	api_resp = (*APIResponse)(resp)
	tweet = &Tweet{}
	err = api_resp.Parse(tweet)
	if err == nil {
		t.Fatalf("Expected an error in Parse")
	}
	expected = "Unable to handle response (status code 403): ``"
	if err.Error() != expected {
		t.Fatalf("Rendered error string `%s`, expected `%s`", err.Error(), expected)
	}
	if _, ok = err.(RateLimitError); ok {
		t.Fatalf("Error should not be castable to a RateLimitError error")
	}
	if rerr, ok = err.(ResponseError); !ok {
		t.Fatalf("Error should be castable to a ResponseError")
	}
	if rerr.Code != 403 {
		t.Errorf("ResponseError code should be 403, got %d", rerr.Code)
	}
	if rerr.Body != `` {
		t.Errorf("ResponseError body should be ``, got `%s`", rerr.Body)
	}
}
