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
	ScreenName string
	Count      string
}

func parseArgs() *Args {
	a := &Args{}
	flag.StringVar(&a.ScreenName, "screen_name", "twitterapi", "Screen name to look up")
	flag.StringVar(&a.Count, "count", "5", "Number of results / page")
	flag.Parse()
	return a
}

func handleRateLimit(err error) error {
	if rle, ok := err.(twittergo.RateLimitError); ok {
		dur := rle.Reset.Sub(time.Now()) + time.Second
		if dur < MINWAIT {
			// Don't wait less than minwait.
			dur = MINWAIT
		}
		msg := "Rate limited. Reset at %v. Waiting for %v\n"
		fmt.Printf(msg, rle.Reset, dur)
		time.Sleep(dur)
		return nil
	}
	return err
}

func fetchAndPrintCursoredList(client *twittergo.Client, path string, query url.Values) (err error) {
	var (
		req     *http.Request
		resp    *twittergo.APIResponse
		results twittergo.CursoredLists
		i       int64
	)
	i = 1
	for {
		url := fmt.Sprintf("%v?%v", path, query.Encode())
		req, err = http.NewRequest("GET", url, nil)
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		if err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			break
		}
		resp, err = client.SendRequest(req)
		if err != nil {
			fmt.Printf("Could not send request: %v\n", err)
			break
		}
		results = twittergo.CursoredLists{}
		if err = resp.Parse(&results); err != nil {
			if err = handleRateLimit(err); err != nil {
				fmt.Printf("Problem parsing response: %v\n", err)
				break
			} else {
				continue
			}
		}
		fmt.Printf("\n")
		for _, list := range results.Lists() {
			user := list.User()
			fmt.Printf("%v.) %v\n", i, list.Name())
			fmt.Printf("Owner: %v (@%v)\n", user.Name(), user.ScreenName())
			fmt.Printf("Members: %v\n", list.MemberCount())
			fmt.Printf("Subscribers: %v\n\n", list.SubscriberCount())
			i += 1
		}
		if results.NextCursorStr() == "0" {
			break
		}
		query.Set("cursor", results.NextCursorStr())
		if resp.HasRateLimit() {
			fmt.Printf("Rate limit:           %v\n", resp.RateLimit())
			fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
			fmt.Printf("Rate limit reset:     %v\n", resp.RateLimitReset())
		} else {
			fmt.Printf("Could not parse rate limit from response.\n")
		}
	}
	return
}

func main() {
	var (
		err    error
		args   *Args
		client *twittergo.Client
	)
	args = parseArgs()
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	query := url.Values{}
	query.Set("screen_name", args.ScreenName)
	query.Set("count", args.Count)
	query.Set("cursor", "-1")

	fmt.Printf("Printing the lists %v is subscribed to:\n", args.ScreenName)
	fmt.Printf("=========================================================\n")
	fetchAndPrintCursoredList(client, "/1.1/lists/subscriptions.json", query)
}
