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
	"fmt"
	"io"
	"json"
	"http"
	"os"
)

type BoundingBox struct {
	Type        string
	Coordinates [][][]float32
}

type Entities struct {
	Hashtags     []Hashtag
	Media        []Media
	Urls         []Url
	UserMentions []UserMention `json:"user_mentions"`
}

type Geo struct {
	Type        string
	Coordinates []float32
}

type Hashtag struct {
	Indices []uint32
	Text    string
}

type Media struct {
	DisplayUrl    string
	ExpandedUrl   string
	Id            string
	IdStr         string `json:"id_str"`
	Indices       []uint32
	MediaUrl      string `json:"media_url"`
	MediaUrlHttps string `json:"media_url_https"`
	Sizes         map[string]MediaSize
	Type          string
	Url           string
}

type MediaSize struct {
	Height uint32 `json:"h"`
	Resize string
	Width  uint32 `json:"w"`
}

type Place struct {
	Attributres map[string]string
	BoundingBox *BoundingBox `json:"bounding_box"`
	Country     string
	CountryCode string `json:"country_code"`
	FullName    string `json:"full_name"`
	Id          string
	Name        string
	PlaceType   string `json:place_type"`
	Url         string
}

type Status struct {
	Annotations          *string
	CreatedAt            string `json:"created_at"`
	Contributors         *string
	Coordinates          *Geo
	Entities             *Entities
	Favorited            bool
	Geo                  *Geo
	Id                   uint64
	IdStr                *string `json:"id_str"`
	InReplyToScreenName  *string `json:"in_reply_to_screen_name"`
	InReplyToStatusId    *uint64 `json:"in_reply_to_status_id"`
	InReplyToStatusIdStr *string `json:"in_reply_to_status_id_str"`
	InReplyToUserId      *uint64 `json:"in_reply_to_user_id"`
	InReplyToUserIdStr   *string `json:"in_reply_to_user_id_str"`
	Place                *Place
	RetweetCount         uint32 `json:"retweet_count"`
	Retweeted            bool
	Source               string
	Text                 string
	Truncated            bool
	User                 *User
}

type Url struct {
	DisplayUrl  *string `json:"display_url"`
	ExpandedUrl *string `json:"expanded_url"`
	Indices     []uint32
	Url         *string
}

type User struct {
	ContributorsEnabled       *bool  `json:"contributors_enabled"`
	CreatedAt                 string `json:"created_at"`
	Description               *string
	FavoritesCount            uint32 `json:"favorites_count"`
	FollowersCount            uint32 `json:"followers_count"`
	FriendsCount              uint32 `json:"friends_count"`
	GeoEnabled                bool   `json:"geo_enabled"`
	Id                        uint64
	Lang                      *string
	Location                  *string
	Name                      *string
	Notifications             *bool
	ProfileBackgroundColor    *string `json:"profile_background_color"`
	ProfileBackgroundImageUrl *string `json:"profile_background_image_url;"`
	ProfileBackgroundTile     *bool   `json:"profile_background_tile"`
	ProfileImageUrl           *string `json:"profile_image_url"`
	ProfileLinkColor          *string `json:"profile_link_color"`
	ProfileSidebarBorderColor *string `json:"profile_sidebar_border_color"`
	ProfileSidebarFillColor   *string `json:"profile_sidebar_fill_color"`
	ProfileTextColor          *string `json:"profile_text_color"`
	ProfileUseBackgroundImage bool    `json:"profile_use_background_image"`
	Protected                 bool
	Status                    *Status
	StatusesCount             uint32  `json:"statuses_count"`
	TimeZone                  *string `json:"time_zone"`
	Url                       *string
	UtcOffset                 int32
	Verified                  bool
}

type UserMention struct {
	Id         uint64
	IdStr      string `json:"id_str"`
	Indices    []uint32
	Name       string
	ScreenName string `json:"screen_name"`
}

type LoggingReader struct {
	reader io.ReadCloser
}

func NewLoggingReader(reader io.ReadCloser) *LoggingReader {
	return &LoggingReader{reader: reader}
}

func (lr *LoggingReader) Read(p []byte) (int, os.Error) {
	n, err := lr.reader.Read(p)
	fmt.Println(string(p))
	return n, err
}

func (lr *LoggingReader) Close() os.Error {
	return lr.reader.Close()
}

type Client struct {
	apiBase string
}

func NewClient() *Client {
	return &Client{apiBase: "https://api.twitter.com/1/"}
}

func (c *Client) sendRequest(method string, path string, params map[string]string) (io.ReadCloser, os.Error) {
	url := c.apiBase + path + "?" + UrlEncode(params)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	reader := NewLoggingReader(response.Body)
	return reader, nil
}

func (c *Client) parseStatusList(reader io.ReadCloser) ([]Status, os.Error) {
	defer reader.Close()
	var statuses []Status
	if err := json.NewDecoder(reader).Decode(&statuses); err != nil {
		return nil, err
	}
	return statuses, nil
}

func (c *Client) GetHomeTimeline() []Status {
	return nil
}

func (c *Client) GetMentions() []Status {
	return nil
}

func (c *Client) GetPublicTimeline() ([]Status, os.Error) {
	params := map[string]string{
		"include_entities": "true",
		"count":            "1",
	}
	body, _ := c.sendRequest("GET", "statuses/public_timeline.json", params)
	return c.parseStatusList(body)
}

func (c *Client) GetRetweetedByMe() []Status {
	return nil
}
