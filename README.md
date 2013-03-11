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
The examples in `/examples` should give a good idea of how to use this client.
You may run them by installing the library, and adding a file called
`CREDENTIALS` in the project root.  The format of this file should be:

    <Twitter consumer key>
    <Twitter consumer secret>
    <Twitter access token>
    <Twitter access token secret>

then:

    go run examples/<path to example>

The simplest example is probably `verify_credentials`.  This calls an
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
[examples/search_app_auth/main.go](https://github.com/kurrik/twittergo/blob/master/examples/search_app_auth/main.go)
for an example of this.

Debugging
---------
To see what requests are being issued by the library, set up an HTTP proxy
such as Charles Proxy and then set the following environment variable:

    export HTTP_PROXY=http://localhost:8888

Development
-----------
Clone the repo and then run:

    scripts/setup_devel.sh

This will symlink the checkout directory to your $GOPATH.  Any local
modifications will be picked up by programs depending on
`github.com/kurrik/twittergo`.  Note that you don't need to run this to run
the examples included with this project.


