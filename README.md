# e
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/paudley/e)
![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)
![GitHub License](https://img.shields.io/github/license/paudley/colorout)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/paudley/colorout)


Golang error handling extensions that support structured errors.

Package e implements an alternative go error class that adds the following:

  * much nicer console error output
  * automatic stack tracings in an efficient manor
  * variable annotation for errors
  * error lambdas for running code to capture errors only IFF there is an error

## Documentation ##

Full `go doc` style documentation for the project can be viewed online without
installing this package by using the excellent GoDoc site here:
http://godoc.org/github.com/paudley/e

You can also view the documentation locally once the package is installed with
the `godoc` tool by running `godoc -http=":6060"` and pointing your browser to
http://localhost:6060/pkg/github.com/paudley/e

## Installation

```bash
$ go get -u github.com/paudley/e
```

## Quick Start

Add this import line to the file you're working in:

```Go
import "github.com/paudley/e"
```
