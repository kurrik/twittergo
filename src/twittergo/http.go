// Copyright 2011 Arne Roomann-Kurrik
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
// limitations under the License.package twittergo

package twittergo

import (
	"http"
	"os"
	"bytes"
	"io"
)

// Abstracts a simple HTTP request.  Only one set of parameters are supported.
type Request struct {
	Method     string
	Url        string
	Parameters map[string]string
	Headers    map[string]string
}

// Creates a new request, initializing the parameters and headers to empty
// maps if passed as nil.
func NewRequest(method string, url string, params map[string]string, headers map[string]string) *Request {
	if params == nil {
		params = map[string]string{}
	}
	if headers == nil {
		headers = map[string]string{}
	}
	return &Request{method, url, params, headers}
}

// Converts a Request object to a http.Request object, suitable for sending
// to a http.Client.
// The Parameters in the request object are automatically form URL-encoded
// and added to the body if the request is anything other than a GET.
func (r *Request) GetHttpRequest() (*http.Request, os.Error) {
	url := r.Url
	var body io.Reader
	if len(r.Parameters) > 0 {
		params := UrlEncode(r.Parameters)
		if r.Method == "GET" {
			url += "?" + params
		} else {
			body = bytes.NewBuffer([]byte(params))
			r.Headers["Content-Type"] = "application/x-www-form-urlencoded"
		}
	}
	header := http.Header{}
	for key, value := range r.Headers {
		header.Set(key, value)
	}
	request, err := http.NewRequest(r.Method, url, body)
	if err != nil {
		return nil, err
	}
	request.Header = header
	return request, nil
}
