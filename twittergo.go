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

// Implements a Twitter client library in Go.
package twittergo

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/kurrik/oauth1a"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Implements a Twitter client.
type Client struct {
	Host       string
	OAuth      *oauth1a.Service
	User       *oauth1a.UserConfig
	AppToken   *BearerToken
	HttpClient *http.Client
}

type BearerToken struct {
	AccessToken string
	Expires     *time.Time
}

func (t *BearerToken) Expired() bool {
	if t.Expires == nil {
		return false
	}
	return time.Now().After(*t.Expires)
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
		User:     user,
		AppToken: nil,
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

// Requests a new app auth bearer token and stores it.
func (c *Client) GetAppToken() (err error) {
	var (
		req  *http.Request
		rb   []byte
		rj   map[string]interface{}
		url  = fmt.Sprintf("https://%v/oauth/2/token", c.Host)
		ct   = "application/x-www-form-urlencoded;charset=UTF-8"
		body = "grant_type=client_credentials"
		ek   = oauth1a.Rfc3986Escape(c.OAuth.ClientConfig.ConsumerKey)
		es   = oauth1a.Rfc3986Escape(c.OAuth.ClientConfig.ConsumerSecret)
		cred = fmt.Sprintf("%v:%v", ek, es)
		ec   = base64.StdEncoding.EncodeToString([]byte(cred))
		h    = fmt.Sprintf("Basic %v", ec)
	)
	req = http.NewRequest("POST", url, bytes.NewBufferString(body))
	req.Header.Set("Authorization", h)
	req.Header.Set("Content-Type", ct)
	if r, err = c.HttpClient.Do(req); err != nil {
		return
	}
	if r.StatusCode != 200 {
		err = fmt.Errorf("Got HTTP %v for OAuth2 request", r.StatusCode)
		return
	}
	if rb, err = ioutil.ReadAll(r.Body); err != nil {
		return
	}
	if err = json.Unmarshal(rb, rj); err != nil {
		return
	}
	var (
		token_type   = string(rj["token_type"])
		access_token = string(rj["access_token"])
		expires_in   = int(rj["expires_in"])
	)
	if token_type != "bearer" {
		err = fmt.Errorf("Got invalid token type: %v", token_type)
	}
	c.AppToken = &BearerToken{
		AccessToken: access_token,
		Expires:     time.Now().Add(time.Second * expires_in),
	}
	return nil
}

// Signs the request with app-only auth, fetching a bearer token if needed.
func (c *Client) Sign(req *http.Request) (err error) {
	if c.AppToken == nil {
		if err = c.GetAppToken(); err != nil {
			return
		}
	}
	var (
		h = fmt.Sprintf("Bearer %v", c.AppToken.AccessToken)
	)
	req.Header.Set("Authorization", h)
	return
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
	if c.User != nil {
		c.OAuth.Sign(req, c.User)
	} else {
		if err = c.Sign(req); err != nil {
			return
		}
	}
	var r *http.Response
	r, err = c.HttpClient.Do(req)
	resp = (*APIResponse)(r)
	return
}
