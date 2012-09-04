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
	"github.com/kurrik/oauth1a"
	"net/http"
	"net/url"
	"strings"
	"encoding/json"
	"fmt"
)

// Implements a Twitter client.
type Client struct {
	Host       string
	OAuth      *oauth1a.Service
	User       *oauth1a.UserConfig
	HttpClient *http.Client
}

// Creates a new Twitter client with the supplied OAuth configuration.
func NewClient(config *oauth1a.ClientConfig, user *oauth1a.UserConfig) *Client {
	var (
		host = "api.twitter.com"
		base = "https://" + host
	)
	return &Client{
		Host:       host,
		HttpClient: new(http.Client),
		User:       user,
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
func (c *Client) SendRequest(req *http.Request) (resp *http.Response, err error) {
	u := req.URL.String()
	if !strings.HasPrefix(u, "http") {
		u = fmt.Sprintf("https://%v%v", c.Host, u)
		req.URL, err = url.Parse(u)
		if err != nil {
			return
		}
	}
	c.OAuth.Sign(req, c.User)
	resp, err = c.HttpClient.Do(req)
	return
}

// Parses a JSON encoded HTTP response into the supplied interface.
func ParseJson(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
