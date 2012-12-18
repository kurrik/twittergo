twittergo
=========
This project implements a Go client library for the Twitter APIs.

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


