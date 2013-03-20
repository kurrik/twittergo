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

package main

import (
	"../../" // Use github.com/kurrik/twittergo for your code.
	"flag"
	"fmt"
	"github.com/kurrik/oauth1a"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const MINWAIT = time.Duration(10) * time.Second

func LoadCredentials() (client *twittergo.Client, err error) {
	credentials, err := ioutil.ReadFile("CREDENTIALS")
	if err != nil {
		return
	}
	lines := strings.Split(string(credentials), "\n")
	config := &oauth1a.ClientConfig{
		ConsumerKey:    lines[0],
		ConsumerSecret: lines[1],
	}
	user := oauth1a.NewAuthorizedConfig(lines[2], lines[3])
	client = twittergo.NewClient(config, user)
	return
}

type Args struct {
	Query string
	ResultType string
}

func parseArgs() *Args {
	a := &Args{}
	flag.StringVar(&a.Query, "query", "twitterapi", "Search query")
	flag.StringVar(&a.ResultType, "result_type", "", "Type of search results to receive")
	flag.Parse()
	return a
}

func main() {
	var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		results *twittergo.SearchResults
		args    *Args
		i       int
	)
	args = parseArgs()
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	query := url.Values{}
	query.Set("q", args.Query)
	if args.ResultType != "" {
		query.Set("result_type", args.ResultType)
	}
	i = 1
	for {
		url := fmt.Sprintf("/1.1/search/tweets.json?%v", query.Encode())
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			break
		}
		resp, err = client.SendRequest(req)
		if err != nil {
			fmt.Printf("Could not send request: %v\n", err)
			break
		}
		results = &twittergo.SearchResults{}
		if err = resp.Parse(results); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < MINWAIT {
					// Don't wait less than minwait.
					dur = MINWAIT
				}
				msg := "Rate limited. Reset at %v. Waiting for %v\n"
				fmt.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				fmt.Printf("Problem parsing response: %v\n", err)
				break
			}
		}
		fmt.Printf("\n")
		for _, tweet := range results.Statuses() {
			user := tweet.User()
			fmt.Printf("%v.) %v\n", i, tweet.Text())
			fmt.Printf("From %v (@%v) ", user.Name(), user.ScreenName())
			fmt.Printf("at %v\n\n", tweet.CreatedAt().Format(time.RFC1123))
			i += 1
		}
		if query, err = results.NextQuery(); err != nil {
			fmt.Printf("No next query: %v\n", err)
			break
		}
		if resp.HasRateLimit() {
			fmt.Printf("Rate limit:           %v\n", resp.RateLimit())
			fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
			fmt.Printf("Rate limit reset:     %v\n", resp.RateLimitReset())
		} else {
			fmt.Printf("Could not parse rate limit from response.\n")
		}
	}
}
