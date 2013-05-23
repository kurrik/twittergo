twittergo
=========
This project implements a Go client library for the Twitter APIs.  This
library supports version 1.1 of Twitter's API and application-only auth.

The goal of this project is to provide a thin veneer over my `oauth1a` library
in order to simplify access to Twitter's APIs from Go.  Where possible, I've
tried to defer to native objects (use of http requests for example).
Additionally, responses could be parsed directly as JSON, but some wrapper
types have been defined in order to provide some convenience methods for
accessing data.

Installing
----------
Run

    go get github.com/kurrik/twittergo

Include in your source:

    import "github.com/kurrik/twittergo"

Godoc
-----
See http://godoc.org/github.com/kurrik/twittergo

Using
-----
I have moved all of the examples to the
https://github.com/kurrik/twittergo-examples project in order to make this
library easier to import.  Please reference that project for ways to address
specific use cases.

The simplest example in the `twittergo-examples` project
is probably `verify_credentials`.  This calls an
endpoint which will return the current user if the request is signed
correctly.

The example starts by loading credentials, which can be done in
many ways.  The example implements a `LoadCredentials` which looks for
the `CREDENTIALS` file mentioned above:

    var (
    	err    error
    	client *twittergo.Client
    	req    *http.Request
    	resp   *twittergo.APIResponse
    	user   *twittergo.User
    )
    client, err = LoadCredentials()
    if err != nil {
    	fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
    	os.Exit(1)
    }

Then, a standard `http` request is created to a `/1.1/` endpoint:

    req, err = http.NewRequest("GET", "/1.1/account/verify_credentials.json", nil)
    if err != nil {
    	fmt.Printf("Could not parse request: %v\n", err)
    	os.Exit(1)
    }

The client object handles sending the request:

    resp, err = client.SendRequest(req)
    if err != nil {
    	fmt.Printf("Could not send request: %v\n", err)
    	os.Exit(1)
    }

The response object has some convenience methods for checking rate limits, etc:

    if resp.HasRateLimit() {
    	fmt.Printf("Rate limit:           %v\n", resp.RateLimit())
    	fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
    	fmt.Printf("Rate limit reset:     %v\n", resp.RateLimitReset())
    } else {
    	fmt.Printf("Could not parse rate limit from response.\n")
    }

Finally, if the response format is known, the library provides some standard
objects which make parsing response data easier:

    user = &twittergo.User{}
    err = resp.Parse(user)
    if err != nil {
    	fmt.Printf("Problem parsing response: %v\n", err)
    	os.Exit(1)
    }
    fmt.Printf("ID:                   %v\n", user.Id())
    fmt.Printf("Name:                 %v\n", user.Name())

Error handling
--------------
Errors are returned by most methods as is Golang convention. However, these
errors may sometimes be cast into `twittergo.Errors`
or `twittergo.RateLimitError` structs which will provide additional information.

To check for rate limiting or other types of server errors, attempt to cast
any errors returned by the `APIResponse.Parse` method.

    resp, err = client.SendRequest(req)
    if err != nil {
    	fmt.Printf("Could not send request: %v\n", err)
    	os.Exit(1)
    }
    tweet = &twittergo.Tweet{}
    err = resp.Parse(tweet)
    if err != nil {
    	if rle, ok := err.(twittergo.RateLimitError); ok {
    		fmt.Printf("Rate limited, reset at %v\n", rle.Reset)
    	} else if errs, ok := err.(twittergo.Errors); ok {
    		for i, val := range errs.Errors() {
    			fmt.Printf("Error #%v - ", i + 1)
    			fmt.Printf("Code: %v ", val.Code())
    			fmt.Printf("Msg: %v\n", val.Message())
    		}
    	} else {
    		fmt.Printf("Problem parsing response: %v\n", err)
    	}
    	os.Exit(1)
    }

The previous snippet would print the following if a user attempted to Tweet
the same text twice in a row:

    Error #1 - Code: 187 Msg: Status is a duplicate

Rate limit errors are pretty easy to use.  They're a simple struct containing
what the limit for the request was, how many were remaining (should be 0)
and when the limiting resets:

    type RateLimitError struct {
    	Limit     uint32
    	Remaining uint32
    	Reset     time.Time
    }

The Errors type is a little more complicated, as it may return one or more
server side errors.  It is possible to cast one to a string using the standard
`Error` method, but if you need to handle individual errors, iterate over
the slice returned by `Errors` (plural) instead:

    for i, val := range errs.Errors() {
    	fmt.Printf("Error #%v - ", i + 1)
    	fmt.Printf("Code: %v ", val.Code())
    	fmt.Printf("Msg: %v\n", val.Message())
    }

Each of *those* errors has a `Code` and a `Message` method, which return
values and strings corresponding to those listed in the "Error codes" section
of this page: https://dev.twitter.com/docs/error-codes-responses

Application-only auth
---------------------
If no user credentials are set, then the library falls back to attempting
to authenticate with application-only auth, as described here:
https://dev.twitter.com/docs/auth/application-only-auth

If you want to obtain an access token for later use, create a client with
no user credentials.

    config := &oauth1a.ClientConfig{
    	ConsumerKey:    "consumer_key",
    	ConsumerSecret: "consumer_secret",
    }
    client = twittergo.NewClient(config, nil)
    if err := c.FetchAppToken(); err != nil {
    	// Handle error ...
    }
    token := c.GetAppToken()
    // ... Save token in data store

To restore a previously obtained token, just call SetAppToken():

    // Get token from data store ...
    c.SetAppToken(token)

Saving and restoring the token isn't necessary if you keep the client in
memory, though.  If you just create a client without any user credentials,
calls to `SendRequest` will automatically fetch and persist the app token
in memory.  See
[search_app_auth/main.go](https://github.com/kurrik/twittergo-examples/blob/master/search_app_auth/main.go)
for an example of this.

Google App Engine
-----------------
This library works with Google App Engine's Go runtime but requires slight
modification to fall back on the `urlfetch` package for http transport.

After creating a `Client`, replace its `HttpClient` with an instance of
`urlfetch.Client`:

    var (
        r      *http.Request
        config *oauth1a.ClientConfig
        user   *oauth1a.UserConfig
    )
    ...
    ctx = appengine.NewContext(r)
    c = twittergo.NewClient(config, user)
    c.HttpClient = urlfetch.Client(ctx)
    
For a comprehensive example, see
[user_timeline_appengine](https://github.com/kurrik/twittergo-examples/blob/master/user_timeline_appengine/src/app/app.go#L138)

Debugging 
---------
To see what requests are being issued by the library, set up an HTTP proxy
such as Charles Proxy and then set the following environment variable:

    export HTTP_PROXY=http://localhost:8888

