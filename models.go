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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	STATUS_INVALID  = 400
	STATUS_NOTFOUND = 404
	STATUS_LIMIT    = 429
	STATUS_GATEWAY  = 502
)

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	msg := "Error %v: %v"
	return fmt.Sprintf(msg, e.Code, e.Message)
}

type Errors struct {
	Errors []Error
}

func (e Errors) Error() string {
	msg := ""
	for _, err := range e.Errors {
		msg += err.Error() + ". "
	}
	return msg
}

type RateLimitError struct {
	Limit     uint32
	Remaining uint32
	Reset     time.Time
}

func (e RateLimitError) Error() string {
	msg := "Rate limit: %v, Remaining: %v, Reset: %v"
	return fmt.Sprintf(msg, e.Limit, e.Remaining, e.Reset)
}

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

// Parses a JSON encoded HTTP response into the supplied interface.
func (r APIResponse) Parse(out interface{}) (err error) {
	switch r.StatusCode {
	case STATUS_NOTFOUND:
		fallthrough
	case STATUS_GATEWAY:
		fallthrough
	case STATUS_INVALID:
		e := &Errors{}
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(e)
		err = *e
		return
	case STATUS_LIMIT:
		err = RateLimitError{
			Limit:     r.RateLimit(),
			Remaining: r.RateLimitRemaining(),
			Reset:     r.RateLimitReset(),
		}
		return
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(out)
	if err == io.EOF {
		err = nil
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

func (t Tweet) User() User {
	return User(t["user"].(map[string]interface{}))
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

func (t Tweet) JSON() (out []byte) {
	var err error
	if out, err = json.MarshalIndent(t, "", "  "); err != nil {
		panic(fmt.Sprintf("Problem converting to JSON: %v", err))
	}
	return
}

// It's a structured list of Tweets!
type SearchResults struct {
	Statuses []Tweet
}

// It's a less structured list of Tweets!
type Timeline []Tweet
