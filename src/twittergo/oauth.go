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
	"strings"
	"fmt"
	"sort"
	"time"
	"crypto/sha1"
	"crypto/hmac"
	"os"
	"encoding/base64"
	"io/ioutil"
	"bytes"
)

// Container for client-specific configuration related to the OAuth process.
// This struct is intended to be serialized and stored for future use.
type OAuthConfig struct {
	ConsumerSecret string
	ConsumerKey    string
	Callback       string
}

// Creates an OAuth config with minimal data suitable for starting an OAuth
// flow.
func NewOAuthConfig(key string, secret string, callback string) *OAuthConfig {
	return &OAuthConfig{
		ConsumerSecret: secret,
		ConsumerKey:    key,
		Callback:       callback,
	}
}

// Container for user-specific keys and secrets related to the OAuth process.
// This struct is intended to be serialized and stored for future use.
// Request and Access tokens are each stored separately, so that the current
// position in the auth flow may be inferred.
type OAuthUserConfig struct {
	RequestTokenSecret string
	RequestTokenKey    string
	AccessTokenSecret  string
	AccessTokenKey     string
	Verifier           string
	AccessValues       http.Values
}

// Returns a token and secret corresponding to where in the OAuth flow this
// config is currently in.  The priority is Access token, Request token, empty
// string.
func (c *OAuthUserConfig) GetToken() (string, string) {
	if c.AccessTokenKey != "" {
		return c.AccessTokenKey, c.AccessTokenSecret
	}
	if c.RequestTokenKey != "" {
		return c.RequestTokenKey, c.RequestTokenSecret
	}
	return "", ""
}

// Represents an API which offers OAuth access.
type OAuthService struct {
	RequestUrl   string
	AuthorizeUrl string
	AccessUrl    string
	Config       *OAuthConfig
	Signer       Signer
}

// Sign and send a Request using the current configuration.
func (o *OAuthService) Send(request *Request, user *OAuthUserConfig, client *http.Client) (*http.Response, os.Error) {
	o.Signer.Sign(request, o.Config, user)
	httpRequest, err := request.GetHttpRequest()
	if err != nil {
		return nil, err
	}
	response, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, os.NewError("Endpoint response: " + response.Status)
	}
	return response, nil
}

// Issue a request to exchange the current request token for an access token.
func (o *OAuthService) GetAccessToken(token string, verifier string, user *OAuthUserConfig, client *http.Client) os.Error {
	if user.RequestTokenKey == "" || user.RequestTokenSecret == "" {
		return os.NewError("No configured request token")
	}
	if user.RequestTokenKey != token {
		return os.NewError("Returned token did not match request token")
	}
	user.Verifier = verifier
	params := map[string]string{
		"oauth_verifier": verifier,
	}
	request := NewRequest("POST", o.AccessUrl, params, nil)
	response, err := o.Send(request, user, client)
	if err != nil {
		return err
	}
	err = o.parseAccessToken(response, user)
	return err
}

// Given the returned response from the access token request, pull out the
// access token and token secret.  Store a copy of any other values returned,
// too, since Twitter returns handy information such as the username.
func (o *OAuthService) parseAccessToken(response *http.Response, user *OAuthUserConfig) os.Error {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	params, err := http.ParseQuery(string(body))
	tokenKey := params.Get("oauth_token")
	tokenSecret := params.Get("oauth_token_secret")
	if tokenKey == "" || tokenSecret == "" {
		return os.NewError("No token or secret found")
	}
	user.AccessTokenKey = tokenKey
	user.AccessTokenSecret = tokenSecret
	user.AccessValues = params
	return nil
}

// Obtain a URL which will allow the current user to authorize access to their
// OAuth-protected data.
func (o *OAuthService) GetAuthorizeUrl(user *OAuthUserConfig) (string, os.Error) {
	if user.RequestTokenKey == "" || user.RequestTokenSecret == "" {
		return "", os.NewError("No configured request token")
	}
	token := http.URLEscape(user.RequestTokenKey)
	return o.AuthorizeUrl + "?oauth_token=" + token, nil
}

// Issue a request to obtain a Request token.
func (o *OAuthService) GetRequestToken(user *OAuthUserConfig, client *http.Client) os.Error {
	params := map[string]string{
		"oauth_callback": o.Config.Callback,
	}
	request := NewRequest("POST", o.RequestUrl, params, nil)
	response, err := o.Send(request, user, client)
	if err != nil {
		return err
	}
	err = o.parseRequestToken(response, user)
	return err
}

// Given the returned response from a Request token request, parse out the
// appropriate request token and secret fields.
func (o *OAuthService) parseRequestToken(response *http.Response, user *OAuthUserConfig) os.Error {
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	params, err := http.ParseQuery(string(body))
	tokenKey := params.Get("oauth_token")
	tokenSecret := params.Get("oauth_token_secret")
	if tokenKey == "" || tokenSecret == "" {
		return os.NewError("No token or secret found")
	}
	user.RequestTokenKey = tokenKey
	user.RequestTokenSecret = tokenSecret
	if params.Get("oauth_callback_confirmed") == "false" {
		return os.NewError("OAuth callback not confirmed")
	}
	return nil
}

// Interface for any OAuth signing implementations.
type Signer interface {
	Sign(request *Request, config *OAuthConfig, user *OAuthUserConfig)
}

// A Signer which implements the HMAC-SHA1 signing algorithm.
type HmacSha1Signer struct{}

// Given an unsigned request, add the appropriate OAuth Authorization header
// using the HMAC-SHA1 algorithm.
func (s *HmacSha1Signer) Sign(request *Request, config *OAuthConfig, user *OAuthUserConfig) {
	oauthParams := map[string]string{
		"oauth_consumer_key":     config.ConsumerKey,
		"oauth_nonce":            s.generateNonce(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        fmt.Sprintf("%v", time.Seconds()),
		"oauth_version":          "1.0",
	}
	tokenKey, tokenSecret := user.GetToken()
	if tokenKey != "" {
		oauthParams["oauth_token"] = tokenKey
	}
	signingParams := map[string]string{}
	for key, value := range oauthParams {
		signingParams[key] = value
	}
	for key, value := range request.Parameters {
		signingParams[key] = value
	}
	signatureParts := []string{
		request.Method,
		http.URLEscape(request.Url),
		s.encodeParameters(signingParams)}
	signatureBase := strings.Join(signatureParts, "&")
	signingKey := config.ConsumerSecret + "&" + tokenSecret
	signer := hmac.NewSHA1([]byte(signingKey))
	signer.Write([]byte(signatureBase))
	oauthSignature := base64.StdEncoding.EncodeToString(signer.Sum())
	oauthParams["oauth_signature"] = oauthSignature
	headerParts := make([]string, len(oauthParams))
	var i = 0
	for key, value := range oauthParams {
		headerParts[i] = Rfc3986Escape(key) + "=\"" + Rfc3986Escape(value) + "\""
		i += 1
	}

	oauthHeader := "OAuth " + strings.Join(headerParts, ", ")
	request.Headers["Authorization"] = oauthHeader
}

// Sort a set of request parameters alphabetically, and encode according to the
// OAuth 1.0a specification.
func (HmacSha1Signer) encodeParameters(params map[string]string) string {
	keys := make([]string, len(params))
	encodedParts := make([]string, len(params))
	i := 0
	for key, _ := range params {
		keys[i] = key
		i += 1
	}
	sort.Strings(keys)
	for i, key := range keys {
		value := params[key]
		encoded := Rfc3986Escape(key) + "=" + Rfc3986Escape(value)
		encodedParts[i] = encoded
	}
	return http.URLEscape(strings.Join(encodedParts, "&"))
}

// Generate a unique nonce value.  Should not be called more than once per
// nanosecond
// TODO: Come up with a better generation method.
func (HmacSha1Signer) generateNonce() string {
	ns := time.Nanoseconds()
	token := "OAuth Client Lib" + string(ns)
	h := sha1.New()
	h.Write([]byte(token))
	return fmt.Sprintf("%x", h.Sum())
}

// Characters which should not be escaped according to RFC 3986.
const UNESCAPE_CHARS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-._~"

// Escapes a string more in line with Rfc3986 than http.URLEscape.
// URLEscape was converting spaces to "+" instead of "%20", which was messing up
// the signing of requests.
func Rfc3986Escape(input string) string {
	var output bytes.Buffer
	// Convert string to bytes because iterating over a unicode string
	// in go parses runes, not bytes.
	for _, c := range []byte(input) {
		if strings.IndexAny(string(c), UNESCAPE_CHARS) == -1 {
			encoded := fmt.Sprintf("%%%X", c)
			output.Write([]uint8(encoded))
		} else {
			output.WriteByte(uint8(c))
		}
	}
	return string(output.Bytes())
}
