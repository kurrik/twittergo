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

// An implementation of Twitter's API in Go.
package twittergo

import (
	"json"
	"http"
	"os"
	"fmt"
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

// Implements a Twitter client.
type Client struct {
	BaseUrl    string
	OAuth      *OAuthService
	HttpClient *http.Client
}

// Creates a new Twitter client with the supplied OAuth configuration.
func NewClient(config *OAuthConfig) *Client {
	oauth := &OAuthService{
		RequestUrl:   "http://api.twitter.com/oauth/request_token",
		AuthorizeUrl: "https://api.twitter.com/oauth/authorize",
		AccessUrl:    "https://api.twitter.com/oauth/access_token",
		Config:       config,
		Signer:       new(HmacSha1Signer),
	}
	return &Client{
		BaseUrl:    "https://api.twitter.com/1/",
		OAuth:      oauth,
		HttpClient: new(http.Client),
	}
}

// Sends a HTTP request through this instance's HTTP client.  If auth is
// requested, then the request is signed with the configured OAuth parameters.
func (c *Client) sendRequest(request *Request, auth bool) (*http.Response, os.Error) {
	if auth {
		return c.OAuth.Send(request, c.HttpClient)
	}
	httpRequest, err := request.GetHttpRequest()
	response, err := c.HttpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		fmt.Println(httpRequest)
		fmt.Println(response)
		return nil, os.NewError("Endpoint response: " + response.Status)
	}
	return response, nil
}

// Parses a JSON encoded HTTP response into the supplied interface.
func (c *Client) parseJson(response *http.Response, out interface{}) os.Error {
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return err
	}
	return nil
}

// Makes a request for a specific path in the API, using supplied params and
// method, and signing the request if needed.  The result is decoded into
// the supplied interface.
func (c *Client) getJson(method string, path string, params map[string]string, auth bool, out interface{}) os.Error {
	request := NewRequest(method, c.BaseUrl+path+".json", params, nil)
	response, err := c.sendRequest(request, auth)
	if err != nil {
		return err
	}
	if err := c.parseJson(response, out); err != nil {
		return err
	}
	return nil
}

// Issues a request for an OAuth Request Token.
func (c *Client) GetRequestToken() os.Error {
	return c.OAuth.GetRequestToken(c.HttpClient)
}

// Returns an URL which the authorizing user should visit to grant access.
func (c *Client) GetAuthorizeUrl() (string, os.Error) {
	return c.OAuth.GetAuthorizeUrl()
}

// Issues a request for an OAuth Access Token
func (c *Client) GetAccessToken(token string, verifier string) os.Error {
	return c.OAuth.GetAccessToken(token, verifier, c.HttpClient)
}

// Returns the global public timeline.
func (c *Client) GetPublicTimeline() ([]Status, os.Error) {
	path := "statuses/public_timeline"
	params := map[string]string{
		"include_entities": "true",
		"count":            "1",
	}
	var statuses []Status
	err := c.getJson("GET", path, params, false, &statuses)
	return statuses, err
}

// Returns retweets by the currently authenticated user.
func (c *Client) GetRetweetedByMe() ([]Status, os.Error) {
	path := "statuses/retweeted_by_me"
	params := map[string]string{
	}
	var statuses []Status
	err := c.getJson("GET", path, params, true, &statuses)
	return statuses, err
}

func (c *Client) GetHomeTimeline() []Status {
	return nil
}

func (c *Client) GetMentions() []Status {
	return nil
}

