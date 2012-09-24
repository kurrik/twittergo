// Copyright 2012 Arne Roomann-Kurrik
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
	"testing"
	"encoding/json"
)

var RAW = []byte(`{"created_at":"Thu Sep 20 20:08:32 +0000 2012","id":248876316206714880,"id_str":"248876316206714880","text":"@mynetx you can send me an email - jasoncosta@twitter.com.","source":"web","truncated":false,"in_reply_to_status_id":248875708464640000,"in_reply_to_status_id_str":"248875708464640000","in_reply_to_user_id":14648265,"in_reply_to_user_id_str":"14648265","in_reply_to_screen_name":"mynetx","user":{"id":14927800,"id_str":"14927800","name":"Jason Costa","screen_name":"jasoncosta","location":"","description":"Platform at Twitter","url":"http:\/\/t.co\/YCA3ZKY","entities":{"url":{"urls":[{"url":"http:\/\/t.co\/YCA3ZKY","expanded_url":"http:\/\/www.jason-costa.blogspot.com\/","display_url":"jason-costa.blogspot.com","indices":[0,19]}]},"description":{"urls":[]}},"protected":false,"followers_count":9128,"friends_count":166,"listed_count":150,"created_at":"Wed May 28 00:20:15 +0000 2008","favourites_count":911,"utc_offset":-28800,"time_zone":"Pacific Time (US & Canada)","geo_enabled":true,"verified":false,"statuses_count":5567,"lang":"en","contributors_enabled":false,"is_translator":true,"profile_background_color":"709397","profile_background_image_url":"http:\/\/a0.twimg.com\/images\/themes\/theme6\/bg.gif","profile_background_image_url_https":"https:\/\/si0.twimg.com\/images\/themes\/theme6\/bg.gif","profile_background_tile":false,"profile_image_url":"http:\/\/a0.twimg.com\/profile_images\/1751674923\/new_york_beard_normal.jpg","profile_image_url_https":"https:\/\/si0.twimg.com\/profile_images\/1751674923\/new_york_beard_normal.jpg","profile_link_color":"FF3300","profile_sidebar_border_color":"86A4A6","profile_sidebar_fill_color":"A0C5C7","profile_text_color":"333333","profile_use_background_image":true,"default_profile":false,"default_profile_image":false,"following":null,"follow_request_sent":false,"notifications":null},"geo":null,"coordinates":null,"place":{"id":"3c1f852ebe886470","url":"https:\/\/api.twitter.com\/1.1\/geo\/id\/3c1f852ebe886470.json","place_type":"city","name":"Northwest Marin","full_name":"Northwest Marin, CA","country_code":"US","country":"United States","bounding_box":{"type":"Polygon","coordinates":[[[-123.134523,37.886947],[-122.563580,37.886947],[-122.563580,38.321227],[-123.134523,38.321227]]]},"attributes":{}},"contributors":null,"retweet_count":0,"entities":{"hashtags":[],"urls":[],"user_mentions":[{"screen_name":"mynetx","name":"J.M.","id":14648265,"id_str":"14648265","indices":[0,7]}]},"favorited":false,"retweeted":false}`)

func BenchmarkParseTweet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &Tweet{}
		json.Unmarshal(RAW, t)
	}
}

type SlowTweet Tweet

func (t *SlowTweet) UnmarshalJSON(b []byte) (err error) {
	out := (*map[string]interface{})(t)
	if err = json.Unmarshal(b, out); err == nil {
		t = (*SlowTweet)(out)
		c := make([]byte, len(b))
		copy(c, b)
		(*t)["json"] = c
	}
	return
}

func BenchmarkSlowTweet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &SlowTweet{}
		json.Unmarshal(RAW, t)
	}
}

func BenchmarkCustomJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := &Tweet{}
		Unmarshal(RAW, t)
	}
}

func TestUnmarshal(t *testing.T) {
	tweet := Tweet{}
	if err := Unmarshal(RAW, &tweet); err != nil {
		t.Errorf(err.Error())
	}
	t.Logf("tweet: %v", tweet)
	const id int64 = 248876316206714880
	if tweet["id"] != id {
		t.Errorf("%v != %v", tweet["id"], id)
	}
}
