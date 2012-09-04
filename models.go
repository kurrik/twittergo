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
	H_LIMIT        = "X-Rate-Limit-Limit"
	H_LIMIT_REMAIN = "X-Rate-Limit-Remaining"
	H_LIMIT_RESET  = "X-Rate-Limit-Reset"
)

const (
	STATUS_LIMIT = 429
)

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

// Parses a JSON encoded HTTP response into the supplied interface.
func (r APIResponse) Parse(out interface{}) (err error) {
	switch r.StatusCode {
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
