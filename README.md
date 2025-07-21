# SIP URI Parser

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/percivalalb/sipuri/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/percivalalb/sipuri/tree/main)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/percivalalb/sipuri)](https://pkg.go.dev/github.com/percivalalb/sipuri)

This module is a pure [Golang](https://go.dev/) implementation to parse URIs with the scheme `sip:` & `sips:`. It adheres to the spec in [RFC-3261 19.1.1](https://www.rfc-editor.org/rfc/rfc3261#section-19.1.1). It is meant to be small and efficent and require no libraries outside the [standard lib](https://pkg.go.dev/std).

Requires Go 1.18+

```console
go get github.com/percivalalb/sipuri
```

## Example

Parse input:

```golang
package main

import (
	"fmt"

	"github.com/percivalalb/sipuri"
)

func main() {
	// Parse the URI string. Returns an error for invalid format or malformed components
	sipURI, err := sipuri.Parse("sip:user:password@host:port;uri-parameters?headers")
	if err != nil {
		panic(err)
	}

	// Print the constituent components
	fmt.Println(sipURI.User()) // "user"
	fmt.Println(sipURI.Password()) // "password"
	fmt.Println(sipURI.Host()) // "host:port"
	fmt.Println(sipURI.Params().Get("uri-parameters"))  // ""
	fmt.Println(sipURI.Headers().GetAll("headers")) // []string{""}
	fmt.Println(sipURI.Headers().Keys()) // []string{"headers"}

	// Re-construct the URI string
	fmt.Println(sipURI.String()) // "sip:user:password@host:port;uri-parameters=?headers="
}
```

Construct URI:

```golang
package main

import (
	"fmt"

	"github.com/percivalalb/sipuri"
)

func main() {
	// Build the URI using the builder. Requires a user and host; optional parameters can be set as needed
	uri := sipuri.New(
		"user",
		"host:port",
		// sipuri.Secure(), // Change from sip: to sips:
		sipuri.WithPassword("password"),
		sipuri.WithParams(sipuri.Params{
			"uri-parameters": {""},
		}),
		sipuri.WithHeaders(sipuri.Headers{
			"headers": {""},
		}),
	)

	// Construct the URI string, encoding parts as required
	fmt.Println(sipURI.String()) // "sip:user:password@host:port;uri-parameters=?headers="
}
```

## Disclaimer

The module *should* parse common `sip:` & `sips:` URIs, thought the module has yet to be thoroughly tested by a variety of users. Please report any issues, thanks!
