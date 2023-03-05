# SIP URI Parser

This module is a pure [Golang](https://go.dev/) implementation to parse URIs with the scheme `sip:` & `sips:`. It tries to adhere to the spec in [RFC-3261 19.1.1](https://www.rfc-editor.org/rfc/rfc3261#section-19.1.1).

Requires go 1.16+ (tested up to 1.20)

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
    sipURI, _ := sipuri.Parse("sip:user:password@host:port;uri-parameters?headers")

    fmt.Println(sipURI.User) // user
    fmt.Println(sipURI.Pass) // password
    fmt.Println(sipURI.Host) // host:port
	fmt.Printf("%v\n", sipURI.Params)  // map[uri-parameters:[]]
	fmt.Printf("%v\n", sipURI.Headers) // map[headers:[]]

    fmt.Println(sipURI.String()) // sip:user:password@host:port;uri-parameters=?headers=
}
```

## Disclaimer

The module should correctly parses common `sip:` & `sips:` URIs, the module has yet to be thoroughly tested against outliers. Please report any issues, thanks!