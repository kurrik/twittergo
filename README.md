twittergo
=========
This project implements a Go client library for the Twitter APIs.

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

Development
-----------
Clone the repo and then run:

    scripts/setup_devel.sh

This will symlink the checkout directory to your $GOPATH.  Any local
modifications will be picked up by programs depending on twittergo.


