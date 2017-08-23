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
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	H_LIMIT              = "X-Rate-Limit-Limit"
	H_LIMIT_REMAIN       = "X-Rate-Limit-Remaining"
	H_LIMIT_RESET        = "X-Rate-Limit-Reset"
	H_MEDIA_LIMIT        = "X-MediaRateLimit-Limit"
	H_MEDIA_LIMIT_REMAIN = "X-MediaRateLimit-Remaining"
	H_MEDIA_LIMIT_RESET  = "X-MediaRateLimit-Reset"
)

const (
	STATUS_OK           = 200
	STATUS_CREATED      = 201
	STATUS_ACCEPTED     = 202
	STATUS_NO_CONTENT   = 204
	STATUS_INVALID      = 400
	STATUS_UNAUTHORIZED = 401
	STATUS_FORBIDDEN    = 403
	STATUS_NOTFOUND     = 404
	STATUS_LIMIT        = 429
	STATUS_GATEWAY      = 502
)

// Error returned if there was an issue parsing the response body.
type ResponseError struct {
	Body string
	Code int
}

func NewResponseError(code int, body string) ResponseError {
	return ResponseError{Code: code, Body: body}
}

func (e ResponseError) Error() string {
	return fmt.Sprintf(
		"Unable to handle response (status code %d): `%v`",
		e.Code,
		e.Body)
}

type Error map[string]interface{}

func (e Error) Code() int64 {
	return int64(e["code"].(float64))
}

func (e Error) Message() string {
	return e["message"].(string)
}

func (e Error) Error() string {
	msg := "Error %v: %v"
	return fmt.Sprintf(msg, e.Code(), e.Message())
}

type Errors map[string]interface{}

func (e Errors) Error() string {
	var (
		msg string = ""
		err Error
		ok  bool
	)
	if e["errors"] == nil {
		return msg
	}
	for _, val := range e["errors"].([]interface{}) {
		if err, ok = val.(map[string]interface{}); ok {
			msg += err.Error() + ". "
		}
	}
	return msg
}

func (e Errors) String() string {
	return e.Error()
}

func (e Errors) Errors() []Error {
	var errs = e["errors"].([]interface{})
	var out = make([]Error, len(errs))
	for i, val := range errs {
		out[i] = Error(val.(map[string]interface{}))
	}
	return out
}

// RateLimitResponse is implemented by both RateLimitError and APIResponse.
type RateLimitResponse interface {
	// HasRateLimit returns false if the ratelimiting information is
	// optional and missing.
	HasRateLimit() bool
	// RateLimit returns the requests per time period capacity of the
	// limit.
	RateLimit() uint32
	// RateLimitRemaining returns how many requests are still available
	// in the current time period.
	RateLimitRemaining() uint32
	// RateLimitReset returns when the rate limit will reset.
	RateLimitReset() time.Time
}

// RateLimitError is returned from SendRequest when a rate limit is encountered.
type RateLimitError struct {
	Limit     uint32
	Remaining uint32
	Reset     time.Time
}

func (e RateLimitError) Error() string {
	msg := "Rate limit: %v, Remaining: %v, Reset: %v"
	return fmt.Sprintf(msg, e.Limit, e.Remaining, e.Reset)
}

func (e RateLimitError) HasRateLimit() bool {
	return true
}

func (e RateLimitError) RateLimit() uint32 {
	return e.Limit
}

func (e RateLimitError) RateLimitRemaining() uint32 {
	return e.Remaining
}

func (e RateLimitError) RateLimitReset() time.Time {
	return e.Reset
}

// APIResponse provides methods for retrieving information from the HTTP
// headers in a Twitter API response.
type APIResponse http.Response

func (r APIResponse) HasRateLimit() bool {
	return r.Header.Get(H_LIMIT) != ""
}

func (r APIResponse) RateLimit() uint32 {
	h := r.Header.Get(H_LIMIT)
	i, _ := strconv.ParseUint(h, 10, 32)
	return uint32(i)
}

func (r APIResponse) RateLimitRemaining() uint32 {
	h := r.Header.Get(H_LIMIT_REMAIN)
	i, _ := strconv.ParseUint(h, 10, 32)
	return uint32(i)
}

func (r APIResponse) RateLimitReset() time.Time {
	h := r.Header.Get(H_LIMIT_RESET)
	i, _ := strconv.ParseUint(h, 10, 32)
	t := time.Unix(int64(i), 0)
	return t
}

func (r APIResponse) HasMediaRateLimit() bool {
	return r.Header.Get(H_MEDIA_LIMIT) != ""
}

func (r APIResponse) MediaRateLimit() uint32 {
	h := r.Header.Get(H_MEDIA_LIMIT)
	i, _ := strconv.ParseUint(h, 10, 32)
	return uint32(i)
}

func (r APIResponse) MediaRateLimitRemaining() uint32 {
	h := r.Header.Get(H_MEDIA_LIMIT_REMAIN)
	i, _ := strconv.ParseUint(h, 10, 32)
	return uint32(i)
}

func (r APIResponse) MediaRateLimitReset() time.Time {
	h := r.Header.Get(H_MEDIA_LIMIT_RESET)
	i, _ := strconv.ParseUint(h, 10, 32)
	t := time.Unix(int64(i), 0)
	return t
}

func (r APIResponse) readBody() (b []byte, err error) {
	var (
		header string
		reader io.Reader
	)
	defer r.Body.Close()
	header = strings.ToLower(r.Header.Get("Content-Encoding"))
	if header == "" || strings.Index(header, "gzip") == -1 {
		reader = r.Body
	} else {
		if reader, err = gzip.NewReader(r.Body); err != nil {
			return
		}
	}
	b, err = ioutil.ReadAll(reader)
	return
}

// ReadBody returns the body of the response as a string.
// Only one of ReadBody and Parse may be called on a given APIResponse.
func (r APIResponse) ReadBody() string {
	var (
		b   []byte
		err error
	)
	if b, err = r.readBody(); err != nil {
		return ""
	}
	return string(b)
}

// Parse unmarshals a JSON encoded HTTP response into the supplied interface,
// with handling for the various kinds of errors the Twitter API can return.
//
// The returned error may be of the type Errors, RateLimitError,
// ResponseError, or an error returned from io.Reader.Read().
func (r APIResponse) Parse(out interface{}) (err error) {
	var b []byte
	switch r.StatusCode {
	case STATUS_UNAUTHORIZED:
		fallthrough
	case STATUS_NOTFOUND:
		fallthrough
	case STATUS_GATEWAY:
		fallthrough
	case STATUS_FORBIDDEN:
		fallthrough
	case STATUS_INVALID:
		e := &Errors{}
		if b, err = r.readBody(); err != nil {
			return
		}
		if err = json.Unmarshal(b, e); err != nil {
			err = NewResponseError(r.StatusCode, string(b))
		} else {
			err = *e
		}
		return
	case STATUS_LIMIT:
		err = RateLimitError{
			Limit:     r.RateLimit(),
			Remaining: r.RateLimitRemaining(),
			Reset:     r.RateLimitReset(),
		}
		// consume the request body even if we don't need it
		r.readBody()
		return
	case STATUS_NO_CONTENT:
		return
	case STATUS_CREATED:
		fallthrough
	case STATUS_ACCEPTED:
		fallthrough
	case STATUS_OK:
		if b, err = r.readBody(); err != nil {
			return
		}
		err = json.Unmarshal(b, out)
		if err == io.EOF {
			err = nil
		}
	default:
		if b, err = r.readBody(); err != nil {
			return
		}
		err = NewResponseError(r.StatusCode, string(b))
	}
	return
}

// It's a user!
type User map[string]interface{}

func (u User) Id() uint64 {
	id, _ := strconv.ParseUint(u["id_str"].(string), 10, 64)
	return id
}

func (u User) IdStr() string {
	return u["id_str"].(string)
}

func (u User) Name() string {
	return u["name"].(string)
}

func (u User) ScreenName() string {
	return u["screen_name"].(string)
}

// It's a Tweet! (Adorably referred to by the API as a "status").
type Tweet map[string]interface{}

func (t Tweet) Id() (id uint64) {
	var (
		err error
		src = t["id_str"].(string)
	)
	if id, err = strconv.ParseUint(src, 10, 64); err != nil {
		panic(fmt.Sprintf("Could not parse ID: %v", err))
	}
	return
}

func (t Tweet) IdStr() string {
	return t["id_str"].(string)
}

func (t Tweet) Text() string {
	return t["text"].(string)
}

func (t Tweet) FullText() string {
	return t["full_text"].(string)
}

func (t Tweet) User() User {
	return User(t["user"].(map[string]interface{}))
}

func (t Tweet) Language() string {
	return t["lang"].(string)
}

func (t Tweet) CreatedAt() (out time.Time) {
	var (
		err error
		src = t["created_at"].(string)
	)
	if out, err = time.Parse(time.RubyDate, src); err != nil {
		panic(fmt.Sprintf("Could not parse time: %v", err))
	}
	return
}

// It's a less structured list of Tweets!
type Timeline []Tweet

// It's a structured list of Tweets!
type SearchResults map[string]interface{}

func (sr SearchResults) Statuses() []Tweet {
	var a []interface{} = sr["statuses"].([]interface{})
	b := make([]Tweet, len(a))
	for i, v := range a {
		b[i] = v.(map[string]interface{})
	}
	return b
}

func (sr SearchResults) SearchMetadata() map[string]interface{} {
	a := sr["search_metadata"].(map[string]interface{})
	return a
}

func (sr SearchResults) NextQuery() (val url.Values, err error) {
	var (
		sm   map[string]interface{}
		n    interface{}
		next string
		ok   bool
	)
	sm = sr.SearchMetadata()
	if n, ok = sm["next_results"]; !ok {
		err = fmt.Errorf("Could not get next_results from search")
		return
	}
	if next, ok = n.(string); !ok {
		err = fmt.Errorf("Could not convert next_results to str: %v", n)
		return
	}
	if next[0] == '?' {
		next = next[1:]
	}
	val, err = url.ParseQuery(next)
	return
}

// A List!
type List map[string]interface{}

func (l List) User() User {
	return User(l["user"].(map[string]interface{}))
}

func (l List) Id() (id uint64) {
	var (
		err error
		src = l["id_str"].(string)
	)
	if id, err = strconv.ParseUint(src, 10, 64); err != nil {
		panic(fmt.Sprintf("Could not parse ID: %v", err))
	}
	return
}

func (l List) IdStr() string {
	return l["id_str"].(string)
}

func (l List) Mode() string {
	return l["mode"].(string)
}

func (l List) Name() string {
	return l["name"].(string)
}

func (l List) Slug() string {
	return l["slug"].(string)
}

func (l List) SubscriberCount() int64 {
	return l["subscriber_count"].(int64)
}

func (l List) MemberCount() int64 {
	return l["member_count"].(int64)
}

// It's a less structured list of Lists!
type Lists []List

// It's a cursored list of Lists!

type CursoredLists map[string]interface{}

func (cl CursoredLists) NextCursorStr() string {
	return cl["next_cursor_str"].(string)
}

func (cl CursoredLists) PreviousCursorStr() string {
	return cl["previous_cursor_str"].(string)
}

func (cl CursoredLists) Lists() Lists {
	var a []interface{} = cl["lists"].([]interface{})
	b := make([]List, len(a))
	for i, v := range a {
		b[i] = v.(map[string]interface{})
	}
	return b
}

// Nested response structure for video uploads.
type VideoUpload map[string]interface{}

func (v VideoUpload) Type() string {
	return v["video_type"].(string)
}

// Response for media upload requests.
type MediaResponse map[string]interface{}

func (r MediaResponse) MediaId() int64 {
	return r["media_id"].(int64)
}

func (r MediaResponse) Size() int64 {
	return r["size"].(int64)
}

func (r MediaResponse) ExpiresAfterSecs() int32 {
	return r["expires_after_secs"].(int32)
}

func (r MediaResponse) Video() VideoUpload {
	return VideoUpload(r["video"].(map[string]interface{}))
}
