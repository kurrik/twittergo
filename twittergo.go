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
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/kurrik/oauth1a"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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
}

func getEnvEitherCase(k string) string {
	if v := os.Getenv(strings.ToUpper(k)); v != "" {
		return v
	}
	return os.Getenv(strings.ToLower(k))
}

// Creates a new Twitter client with the supplied OAuth configuration.
// Supports the use of HTTP proxies through the $HTTP_PROXY env var.
// For example:
//     export HTTP_PROXY=http://localhost:8888
//
// When using a proxy, disable TLS certificate verification with the following:
//     export TLS_INSECURE=1
func NewClient(config *oauth1a.ClientConfig, user *oauth1a.UserConfig) *Client {
	var (
		host      = "api.twitter.com"
		base      = "https://" + host
		req, _    = http.NewRequest("GET", "https://api.twitter.com", nil)
		proxy, _  = http.ProxyFromEnvironment(req)
		transport *http.Transport
		tlsconfig *tls.Config
	)
	if proxy != nil {
		tlsconfig = &tls.Config{
			InsecureSkipVerify: getEnvEitherCase("TLS_INSECURE") != "",
		}
		if tlsconfig.InsecureSkipVerify {
			log.Printf("WARNING: SSL cert verification  disabled\n")
		}
		transport = &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: tlsconfig,
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

// Sets the app-only auth token to the specified string.
func (c *Client) SetAppToken(token string) {
	c.AppToken = &BearerToken{
		AccessToken: token,
	}
}

// Returns the current app-only auth token or an empty string.
// Must call FetchAppToken to populate this value before getting it.
// You may call SetAppToken with the value returned by this call in order
// to restore a previously-fetched bearer token to use.
func (c *Client) GetAppToken() string {
	if c.AppToken == nil {
		return ""
	}
	return c.AppToken.AccessToken
}

// Requests a new app auth bearer token and stores it.
func (c *Client) FetchAppToken() (err error) {
	var (
		req  *http.Request
		resp *http.Response
		rb   []byte
		rj   = map[string]interface{}{}
		url  = fmt.Sprintf("https://%v/oauth2/token", c.Host)
		ct   = "application/x-www-form-urlencoded;charset=UTF-8"
		body = "grant_type=client_credentials"
		ek   = oauth1a.Rfc3986Escape(c.OAuth.ClientConfig.ConsumerKey)
		es   = oauth1a.Rfc3986Escape(c.OAuth.ClientConfig.ConsumerSecret)
		cred = fmt.Sprintf("%v:%v", ek, es)
		ec   = base64.StdEncoding.EncodeToString([]byte(cred))
		h    = fmt.Sprintf("Basic %v", ec)
	)
	req, err = http.NewRequest("POST", url, bytes.NewBufferString(body))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", h)
	req.Header.Set("Content-Type", ct)
	if resp, err = c.HttpClient.Do(req); err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Got HTTP %v instead of 200", resp.StatusCode)
		return
	}
	if rb, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	if err = json.Unmarshal(rb, &rj); err != nil {
		return
	}
	var (
		token_type   = rj["token_type"].(string)
		access_token = rj["access_token"].(string)
	)
	if token_type != "bearer" {
		err = fmt.Errorf("Got invalid token type: %v", token_type)
	}
	c.SetAppToken(access_token)
	return nil
}

// Signs the request with app-only auth, fetching a bearer token if needed.
func (c *Client) Sign(req *http.Request) (err error) {
	if c.AppToken == nil {
		if err = c.FetchAppToken(); err != nil {
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
