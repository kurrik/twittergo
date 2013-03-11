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
	"../../" // Use github.com/kurrik/twittergo for your code.
	"fmt"
	"github.com/kurrik/oauth1a"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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
	client = twittergo.NewClient(config, nil)
	return
}

func main() {
	var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		results *map[string]interface{}
	)
	client, err = LoadCredentials()
	if err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	url := fmt.Sprintf("/1.1/application/rate_limit_status.json")
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}
	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Response: %v\n", resp)
	//fmt.Printf("%v\n", resp.ReadBody())

	results = &map[string]interface{}{}
	if err = resp.Parse(results); err != nil {
		fmt.Printf("Could not parse results: %v\n", err)
		os.Exit(1)
	}
	PrintMap(*results)
}

func PrintMap(obj interface{}) {
	printMap(obj, "", "  ")
}

func printMap(i interface{}, prefix string, step string) {
	var (
		m map[string]interface{}
		ok bool
	)
	if m, ok = i.(map[string]interface{}); ok {
		for key, val := range m {
			if _, ok = val.(map[string]interface{}); ok {
				fmt.Printf("%v%v:\n", prefix, key)
				printMap(val, fmt.Sprintf("%v%v", prefix, step), step)
			} else {
				fmt.Printf("%v%v: %v\n", prefix, key, val)
			}
		}
	} else {
		fmt.Printf("%v\n", i)
	}
}

