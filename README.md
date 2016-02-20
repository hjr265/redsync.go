# Redsync.go

**This package is being replaced with https://gopkg.in/redsync.v1. I will continue to maintain this package for a while so that its users do not feel abandoned. But I request everyone to gradually switch to the other package which will be maintained more actively.**

[![Build Status](https://drone.io/github.com/hjr265/redsync.go/status.png)](https://drone.io/github.com/hjr265/redsync.go/latest)

Redsync.go provides a Redis-based distributed mutual exclusion lock implementation for Go as described in [this](http://antirez.com/news/77) blog post. A reference library (by [antirez](https://github.com/antirez)) for Ruby is available at [github.com/antirez/redlock-rb](https://github.com/antirez/redlock-rb).

## Installation

Install Redsync.go using the go get command:

    $ go get github.com/hjr265/redsync.go/redsync

The only dependencies are the Go distribution and `github.com/garyburd/redigo/redis`.

## Documentation

- [Reference](http://godoc.org/github.com/hjr265/redsync.go/redsync)

## Contributing

Contributions are welcome.

## License

Redsync.go is available under the [BSD (3-Clause) License](http://opensource.org/licenses/BSD-3-Clause).

## Disclaimer

This code implements an algorithm which is currently a proposal, it was not formally analyzed. Make sure to understand how it works before using it in production environments.
