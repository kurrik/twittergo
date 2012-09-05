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

package main

import (
	"bytes"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

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

func GetBody() (body io.ReadWriter, header string, err error) {
	var (
		mp     *multipart.Writer
		media  []byte
		writer io.Writer
	)
	body = bytes.NewBufferString("")
	mp = multipart.NewWriter(body)
	media, err = ioutil.ReadFile("examples/tweet_media/media.png")
	if err != nil {
		return
	}
	mp.WriteField("status", fmt.Sprintf("Hello %v!", time.Now()))
	writer, err = mp.CreateFormFile("media[]", "media.png")
	if err != nil {
		return
	}
	writer.Write(media)
	header = fmt.Sprintf("multipart/form-data;boundary=%v", mp.Boundary())
	mp.Close()
	return
}

func main() {
	var (
		err    error
		client *twittergo.Client
		req    *http.Request
		resp   *twittergo.APIResponse
		tweet  *twittergo.Tweet
	)
	client, err = LoadCredentials()
	if err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}

	body, header, err := GetBody()
	if err != nil {
		fmt.Printf("Problem loading body: %v\n", err)
		os.Exit(1)
	}

	endpoint := "https://upload.twitter.com/1/statuses/update_with_media.json"
	req, err = http.NewRequest("POST", endpoint, body)
	if err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", header)

	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}
	tweet = &twittergo.Tweet{}
	err = resp.Parse(tweet)
	if err != nil {
		fmt.Printf("Problem parsing response: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("ID:                         %v\n", tweet.Id())
	fmt.Printf("Tweet:                      %v\n", tweet.Text())
	fmt.Printf("User:                       %v\n", tweet.User().Name())
	if resp.HasRateLimit() {
		fmt.Printf("Rate limit:                 %v\n", resp.RateLimit())
		fmt.Printf("Rate limit remaining:       %v\n", resp.RateLimitRemaining())
		fmt.Printf("Rate limit reset:           %v\n", resp.RateLimitReset())
	} else {
		fmt.Printf("Could not parse rate limit from response.\n")
	}
	if resp.HasMediaRateLimit() {
		fmt.Printf("Media Rate limit:           %v\n", resp.MediaRateLimit())
		fmt.Printf("Media Rate limit remaining: %v\n", resp.MediaRateLimitRemaining())
		fmt.Printf("Media Rate limit reset:     %v\n", resp.MediaRateLimitReset())
	} else {
		fmt.Printf("Could not parse media rate limit from response.\n")
	}
}
