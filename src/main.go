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

func main() {
	client := twittergo.NewClient()
	fmt.Println("Getting public timeline")
	tweets, err := client.GetPublicTimeline()
	if err != nil {
		fmt.Println("Error:", err)
	}
	PrintTweets(tweets)
}
