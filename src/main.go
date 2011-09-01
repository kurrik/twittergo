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
// limitations under the License.package main

package main

import (
	"twittergo"
	"fmt"
	"reflect"
	"os"
	"path"
	"gob"
)

func PrintValue(value reflect.Value, indent string) {
	value = reflect.Indirect(value)
	if value.IsValid() {
		switch value.Kind() {
		case reflect.Array, reflect.Slice:
			PrintArray(value.Interface(), indent)
		case reflect.Struct:
			PrintStruct(value.Interface(), indent)
		case reflect.Bool:
			fmt.Println(indent, value.Bool())
		case reflect.Uint32, reflect.Uint64:
			fmt.Println(indent, value.Uint())
		case reflect.Int, reflect.Int32, reflect.Int64:
			fmt.Println(indent, value.Int())
		default:
			fmt.Println(indent, value.String(), value.Kind())
		}
	}
}

func PrintArray(item interface{}, indent string) {
	itemValue := reflect.ValueOf(item)
	for i := 0; i < itemValue.Len(); i++ {
		value := itemValue.Index(i)
		fmt.Println(indent, "[")
		PrintValue(value, indent+"    ")
		fmt.Println(indent, "]")
	}
}

func PrintStruct(item interface{}, indent string) {
	itemType := reflect.TypeOf(item)
	itemValue := reflect.ValueOf(item)
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		value := itemValue.FieldByName(field.Name)
		value = reflect.Indirect(value)
		if value.IsValid() {
			fmt.Println(indent, field.Name, field.Type)
			PrintValue(value, indent+"    ")
		}
	}
}

func PrintTweets(tweets []twittergo.Status) {
	for _, tweet := range tweets {
		fmt.Println("Tweet:")
		fmt.Println("------------------------------------")
		PrintStruct(tweet, "")
		fmt.Println("")
	}
}

// Prompts the user and returns their input as a string.
func PromptUser(prompt string) string {
	fmt.Print(prompt + ": ")
	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return ""
	}
	return input
}

// Returns the directory path of the directory this executable is stored in.
func GetExecutableDirectory() string {
	wd, _ := os.Getwd()
	dir, _ := path.Split(path.Clean(path.Join(wd, os.Args[0])))
	return dir
}

// Saves a serialized version of a twittergo.OAuthConfig to the specified path.
// Sets the file mask to 600.
func SaveConfig(path string, config *twittergo.OAuthConfig) os.Error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	encoder.Encode(config)
	return os.Chmod(path, 384) // Same as chmod 600
}

// Loads a serialized version of a twittergo.OAuthConfig from the specified
// path.
func LoadConfig(path string) (*twittergo.OAuthConfig, os.Error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var config twittergo.OAuthConfig
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	path := path.Join(GetExecutableDirectory(), "config.bin")
	config, err := LoadConfig(path)
	if err != nil {
		fmt.Println("Unable to read configuration at ", path)
		key := PromptUser("Enter your Twitter Consumer Key")
		secret := PromptUser("Enter your Twitter Consumer Secret")
		config = twittergo.NewOAuthConfig(key, secret, "oob")
		err = SaveConfig(path, config)
		if err != nil {
			fmt.Println("Could not save configuration file:", err)
			os.Exit(1)
		}
		fmt.Println("Saved configuration file to ", path)
	}
	client := twittergo.NewClient(config)
	if config.AccessTokenKey == "" {
		fmt.Println("No access token, starting OAuth flow.")
		if err := client.GetRequestToken(); err != nil {
			fmt.Println("Could not get a request token:", err)
			os.Exit(1)
		}
		url, err := client.GetAuthorizeUrl()
		if err != nil {
			fmt.Println("Could not get an authorize URL:", err)
			os.Exit(1)
		}
		fmt.Println("Please visit this URL in your browser:", url)
		pin := PromptUser("Please input the PIN displayed")
		if err := client.GetAccessToken(config.RequestTokenKey, pin); err != nil {
			fmt.Println("Could not get an access token:", err)
			os.Exit(1)
		}
		err = SaveConfig(path, config)
		if err != nil {
			fmt.Println("Could not save configuration file:", err)
			os.Exit(1)
		}
		fmt.Println("Saved configuration file to ", path)
	}

	fmt.Println("Getting public timeline")
	fmt.Println("-----------------------")
	tweets, err := client.GetPublicTimeline(nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	PrintTweets(tweets)

	fmt.Println("\nGetting your retweets")
	fmt.Println("-----------------------")
	tweets, err = client.GetRetweetedByMe(&twittergo.Parameters{Count: 1})
	if err != nil {
		fmt.Println("Error:", err)
	}
	PrintTweets(tweets)

	fmt.Println("\nGetting ChromiumDev's retweets")
	fmt.Println("-----------------------")
	tweets, err = client.GetRetweetedByUser(true, &twittergo.Parameters{
		Id:    "ChromiumDev",
		Count: 1,
	})
	if err != nil {
		fmt.Println("Error:", err)
	}
	PrintTweets(tweets)

	fmt.Println("\nGetting users who retweeted https://twitter.com/#!/kurrik/status/108988671176359937")
	fmt.Println("-----------------------")
	users, err := client.GetStatusRetweetedBy("108988671176359937", true, nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	PrintArray(users, "")
	ids, err := client.GetStatusRetweetedByIds("108988671176359937", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	PrintArray(ids, "")

	fmt.Println("\nGetting https://twitter.com/#!/kurrik/status/108988671176359937")
	fmt.Println("-----------------------")
	tweet, err := client.GetStatus("108988671176359937", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	PrintStruct(tweet, "")
}
