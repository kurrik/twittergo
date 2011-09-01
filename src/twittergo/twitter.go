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
	"reflect"
	"strings"
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

// Implements optional parameters which may be passed to an API method.
type Parameters struct {
	overrides          map[string]string
	ContributorDetails interface{} `contributor_details`
	Count              interface{}
	ExcludeReplies     interface{} `exclude_replies`
	Id                 interface{}
	IncludeEntities    interface{} `include_entities`
	IncludeRts         interface{} `include_rts`
	MaxId              interface{} `max_id`
	Page               interface{}
	ScreenName         interface{} `screen_name`
	SinceId            interface{} `since_id`
	TrimUser           interface{} `trim_user`
	UserId             interface{} `user_id`
}

// Convert parameters to a map to be used in a HTTP request.
func (p *Parameters) Map() map[string]string {
	params := map[string]string{}
	pType := reflect.TypeOf(*p)
	pValue := reflect.Indirect(reflect.ValueOf(p))
	for i := 0; i < pType.NumField(); i++ {
		field := pType.Field(i)
		value := reflect.Indirect(pValue.FieldByName(field.Name))
		if value.IsNil() != true && field.Name != "overrides" {
			key := string(field.Tag)
			if key == "" {
				key = strings.ToLower(field.Name)
			}
			params[key] = fmt.Sprintf("%v", value.Interface())
		}
	}
	for key, value := range p.overrides {
		params[key] = value
	}
	return params
}

// Sets an "override" parameter value, creating a new Parameters if needed.
func (p *Parameters) Set(key string, value string) *Parameters {
	if p == nil {
		p = new(Parameters)
		p.overrides = map[string]string{}
	}
	p.overrides[key] = value
	return p
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
func (c *Client) getJson(method string, path string, params *Parameters, auth bool, out interface{}) os.Error {
	var paramsMap map[string]string = nil
	if params != nil {
		paramsMap = params.Map()
	}
	request := NewRequest(method, c.BaseUrl+path+".json", paramsMap, nil)
	response, err := c.sendRequest(request, auth)
	if err != nil {
		return err
	}
	if err := c.parseJson(response, out); err != nil {
		return err
	}
	return nil
}

// Makes a request for a path which contains a feed of status updates.
func (c *Client) getStatuses(path string, params *Parameters, auth bool) ([]Status, os.Error) {
	var statuses []Status
	err := c.getJson("GET", path, params, auth, &statuses)
	return statuses, err
}

// Makes a request for a path which contains a feed of arbitrary data.
func (c *Client) getData(path string, params *Parameters, auth bool, out interface{}) os.Error {
	err := c.getJson("GET", path, params, auth, &out)
	return err
}

// Issues a request for an OAuth Request Token.
func (c *Client) GetRequestToken() os.Error {
	return c.OAuth.GetRequestToken(c.HttpClient)
}

// Returns an URL which the authorizing user should visit to grant access.
func (c *Client) GetAuthorizeUrl() (string, os.Error) {
	return c.OAuth.GetAuthorizeUrl()
}

// Issues a request for an OAuth Access Token.
func (c *Client) GetAccessToken(token string, verifier string) os.Error {
	return c.OAuth.GetAccessToken(token, verifier, c.HttpClient)
}

/*
 * The following methods should reflect https://dev.twitter.com/docs/api
 */

// Returns the authenticated user's home timeline.
func (c *Client) GetHomeTimeline(params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/retweeted_by_me", params, true)
}

// Returns the mentions of the current user.
func (c *Client) GetMentions(params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/mentions", params, true)
}

// Returns the global public timeline.
func (c *Client) GetPublicTimeline(params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/public_timeline", params, false)
}

// Returns retweets by the currently authenticated user.
func (c *Client) GetRetweetedByMe(params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/retweeted_by_me", params, true)
}

// Returns retweets to the currently authenticated user.
func (c *Client) GetRetweetedToMe(params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/retweeted_to_me", params, true)
}

// Returns retweets of the currently authenticated user's tweets.
func (c *Client) GetRetweetsOfMe(params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/retweets_of_me", params, true)
}

// Returns the timeline for a user, defaults to the current authenticated user.
func (c *Client) GetUserTimeline(auth bool, params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/user_timeline", params, auth)
}

// Returns the retweets to the specified user.
func (c *Client) GetRetweetedToUser(auth bool, params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/retweeted_to_user", params, auth)
}

// Returns the retweets from the specified user.
func (c *Client) GetRetweetedByUser(auth bool, params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/retweeted_by_user", params, auth)
}

// Returns the users who have retweeted the specified status.
func (c *Client) GetStatusRetweetedBy(id string, auth bool, params *Parameters) ([]User, os.Error) {
	var users []User
	err := c.getData("statuses/"+id+"/retweeted_by", params, auth, &users)
	return users, err
}

// Returns the IDs of users who have retweeted the specified status.
// Numerical IDs are not supported - just convert the string array.
func (c *Client) GetStatusRetweetedByIds(id string, params *Parameters) ([]string, os.Error) {
	var ids []string
	params = params.Set("stringify_ids", "true")
	err := c.getData("statuses/"+id+"/retweeted_by/ids", params, true, &ids)
	return ids, err
}

// Get the retweets of a given status.
func (c *Client) GetStatusRetweets(id string, params *Parameters) ([]Status, os.Error) {
	return c.getStatuses("statuses/retweets/"+id, params, true)
}

// Gets a single status
func (c *Client) GetStatus(id string, params *Parameters) (Status, os.Error) {
	var status Status
	err := c.getData("statuses/show/"+id, params, false, &status)
	return status, err
}
