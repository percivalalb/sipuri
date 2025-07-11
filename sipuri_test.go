package sipuri_test

import (
	"fmt"
	"testing"

	"github.com/percivalalb/sipuri"
)

func TestNew(t *testing.T) {
	t.Parallel()

	type test struct {
		uri    string
		sipURI sipuri.URI
		msg    string
	}

	tests := []test{
		{"sip:user:password@host:port;uri-parameters=?headers=", sipuri.New(
			"user",
			"host:port",
			sipuri.WithPassword("password"),
			sipuri.WithParams(sipuri.Params{
				"uri-parameters": {""},
			}),
			sipuri.WithHeaders(sipuri.Headers{
				"headers": {""},
			}),
		), "template uri"},
		{"sips:user@host", sipuri.New("user", "host", sipuri.Secure()), "secure upgrade"},
		{"sip:user@host;key1=value1&key2=value2&key2=value3", sipuri.New(
			"user", "host",
			sipuri.WithParams(sipuri.Params{
				"key1": {"value1"},
				"key2": {"value2", "value3"},
				"key3": nil,
			}),
		), "secure"},
	}

	for _, test := range tests {
		equalF(t, test.uri, test.sipURI.String(), "stringify mismatch")
	}

	uri := sipuri.New(
		"user",
		"host:port",
		sipuri.WithPassword("password"),
		sipuri.WithParams(sipuri.Params{
			"uri-parameters": {""},
		}),
		sipuri.WithHeaders(sipuri.Headers{
			"headers": {""},
		}),
	)

	equalF(t, sipuri.SIP, uri.Proto(), "protocol mismatch")
	equalF(t, "user", uri.User(), "user mismatch")
	equalF(t, "password", uri.Password(), "password mismatch")
	equalF(t, "host:port", uri.Host(), "host mismatch")
	equalF(t, "", uri.Params().Get("uri-parameters"), "param mismatch")
	equalF(t, "", uri.Headers().Get("headers"), "header mismatch")
	equalF(t, []string{""}, uri.Params().GetAll("uri-parameters"), "param mismatch")
	equalF(t, []string{""}, uri.Headers().GetAll("headers"), "header mismatch")
	equalF(t, []string{"uri-parameters"}, uri.Params().Keys(), "all params present")
	equalF(t, []string{"headers"}, uri.Headers().Keys(), "all headers present")
	equalF(t, "", uri.Headers().Get("example"), "non-existent header present")
	equalF(t, []string(nil), uri.Headers().GetAll("example"), "non-existent header present")
	equalF(t, "sip:user:password@host:port;uri-parameters=?headers=", uri.String(), "stringify mismatch")
}

func ExampleNew() {
	sipURI := sipuri.New(
		"user",
		"host:port",
		sipuri.WithPassword("password"),
		sipuri.WithParams(sipuri.Params{
			"uri-parameters": {""},
		}),
		sipuri.WithHeaders(sipuri.Headers{
			"headers": {""},
		}),
	)

	// Re-construct the URI
	fmt.Println(sipURI.String())

	// Output:
	// sip:user:password@host:port;uri-parameters=?headers=
}
