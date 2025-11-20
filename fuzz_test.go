//nolint:testpackage // Needs access to private type uriOption
package sipuri

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func FuzzNew(f *testing.F) {
	f.Add(false, "user", "", "example.com", "", "")
	f.Add(true, "name", "password123", "sipserver.com", `{"key":["value"]}`, "")
	f.Add(false, "", "", "sipserver.com:1234", "", `{"key":["value"]}`)

	f.Fuzz(func(t *testing.T, secure bool, user, password, host, headersJSON, parametersJSON string) {
		func() {
			uriStr := "<not-computed>"

			// Catch any unexpected (or expected in places below) panics and report failure.
			defer func() {
				err := recover()
				if err != nil {
					t.Errorf("panic building SIP URL %q %s", uriStr, err)
				}
			}()

			var headers map[string][]string
			if err := json.Unmarshal([]byte(headersJSON), &headers); headersJSON != "" && err != nil {
				t.Skip()
			}

			var parameters map[string][]string
			if err := json.Unmarshal([]byte(parametersJSON), &parameters); parametersJSON != "" && err != nil {
				t.Skip()
			}

			var opts []uriOption

			if secure {
				opts = append(opts, Secure())
			}

			if password != "" {
				opts = append(opts, WithPassword(password))
			}

			if headers != nil {
				opts = append(opts, WithHeaders(headers))
			}

			if parameters != nil {
				opts = append(opts, WithParams(parameters))
			}

			uri := New(user, host, opts...)

			// This is the main part we are checking does not panic.
			uriStr = uri.String()

			// Check other methods return sensible values and also do not panic.

			if atCount := strings.Count(uriStr, "@"); atCount > 1 {
				panic(fmt.Sprintf("at symbol count not 0 or 1 (actual: %d)", atCount))
			}

			if colonCount := strings.Count(uriStr, ":"); colonCount == 0 {
				panic(fmt.Sprintf("colon symbol (actual: %d)", colonCount))
			}

			port := uri.Port()
			if port == "" {
				panic("empty port SIP URL")
			}

			transport := uri.Transport()
			if transport == "" {
				panic("empty transport protocol")
			}
		}()
	})
}

func FuzzParse(f *testing.F) {
	f.Add(false, "sip:user:password@host:port;uri-parameters?headers")
	f.Add(true, "sip:user@example.invalid")

	f.Fuzz(func(t *testing.T, lazy bool, input string) {
		func() {
			// Catch any unexpected (or expected in places below) panics and report failure.
			defer func() {
				err := recover()
				if err != nil {
					t.Errorf("panic parsing SIP URL %s, %s", input, err)
				}
			}()

			parseFunc := Parse
			if lazy {
				parseFunc = ParseLazy
			}

			// This is the main part we are checking does not panic.
			uri, err := parseFunc(input)

			if uri == nil && err == nil {
				panic("nil SIP uri struct")
			}
		}()
	})
}
