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
// limitations under the License.

package twittergo

import (
	"fmt"
	"github.com/kurrik/oauth1a"
	"net/http"
	"net/url"
	"strings"
)

// Implements a Twitter client.
type Client struct {
	Host       string
	OAuth      *oauth1a.Service
	User       *oauth1a.UserConfig
	HttpClient *http.Client
}

// Creates a new Twitter client with the supplied OAuth configuration.
// Supports the use of HTTP proxies through the $HTTP_PROXY env var.
// For example:
//     export HTTP_PROXY=http://localhost:8888
func NewClient(config *oauth1a.ClientConfig, user *oauth1a.UserConfig) *Client {
	var (
		host      = "api.twitter.com"
		base      = "https://" + host
		req, _    = http.NewRequest("GET", "https://api.twitter.com", nil)
		proxy, _  = http.ProxyFromEnvironment(req)
		transport *http.Transport
	)
	if proxy != nil {
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	} else {
		transport = &http.Transport{}
	}
	return &Client{
		Host: host,
		HttpClient: &http.Client{
			Transport: transport,
		},
		User: user,
		OAuth: &oauth1a.Service{
			RequestURL:   base + "/oauth/request_token",
			AuthorizeURL: base + "/oauth/authorize",
			AccessURL:    base + "/oauth/access_token",
			ClientConfig: config,
			Signer:       new(oauth1a.HmacSha1Signer),
		},
	}
}

// Changes the user authorization credentials for this client.
func (c *Client) SetUser(user *oauth1a.UserConfig) {
	c.User = user
}

// Sends a HTTP request through this instance's HTTP client.
func (c *Client) SendRequest(req *http.Request) (resp *APIResponse, err error) {
	u := req.URL.String()
	if !strings.HasPrefix(u, "http") {
		u = fmt.Sprintf("https://%v%v", c.Host, u)
		req.URL, err = url.Parse(u)
		if err != nil {
			return
		}
	}
	c.OAuth.Sign(req, c.User)
	var r *http.Response
	r, err = c.HttpClient.Do(req)
	resp = (*APIResponse)(r)
	return
}
