# SIP URI Parser

[![CircleCI](https://circleci.com/gh/percivalalb/sipuri.svg?style=svg)](https://circleci.com/gh/percivalalb/sipuri)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/percivalalb/sipuri)](https://pkg.go.dev/github.com/percivalalb/sipuri)

This module is a pure [Golang](https://go.dev/) implementation to parse URIs with the scheme `sip:` & `sips:`. It tries to adhere to the spec in [RFC-3261 19.1.1](https://www.rfc-editor.org/rfc/rfc3261#section-19.1.1). It is meant to be small and efficent and require no libraries outside the [standard lib](https://pkg.go.dev/std).

Requires go 1.18+

```console
go get github.com/percivalalb/sipuri
```

## Example

```golang
package main

import (
	"fmt"

	"github.com/percivalalb/sipuri"
)

func main() {
	// Parse the URI. Errors on unexpected schemes or malformed URIs
  	sipURI, err := sipuri.Parse("sip:user:password@host:port;uri-parameters?headers")
	if err != nil {
		panic(err)
	}

	// Print the consistent components
   	fmt.Println(sipURI.User()) // user
   	fmt.Println(sipURI.Password()) // password
    	fmt.Println(sipURI.Host()) // host:port
    	fmt.Printf("%v\n", sipURI.Params())  // map[uri-parameters:[]]
    	fmt.Printf("%v\n", sipURI.Headers()) // map[headers:[]]

	// Re-construct the URI
	fmt.Println(sipURI.String()) // sip:user:password@host:port;uri-parameters=?headers=
}
```

## Disclaimer

The module *should* parse common `sip:` & `sips:` URIs, thought the module has yet to be thoroughly tested against outliers. Please report any issues, thanks!
