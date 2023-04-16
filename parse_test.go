package sipuri_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/percivalalb/sipuri"
)

var parseFuncs = [2](func(string) (*sipuri.URI, error)){
	sipuri.Parse,
	sipuri.ParseLazy,
}

func TestParse(t *testing.T) {
	t.Parallel()

	type test struct {
		uri    string
		sipURI sipuri.URI
		transp string
		msg    string
	}

	tests := []test{
		{"sip:user:password@host:port;uri-parameters=?headers=", sipuri.New(
			"user",
			"host:port",
			sipuri.WithPassword("password"),
			sipuri.WithParams(sipuri.KeyValuePairs{
				"uri-parameters": {""},
			}),
			sipuri.WithHeaders(sipuri.KeyValuePairs{
				"headers": {""},
			}),
		), "UDP", "template uri"},

		// From https://www.rfc-editor.org/rfc/rfc3261#section-19.1.1
		{"sip:alice@atlanta.com", sipuri.New(
			"alice",
			"atlanta.com",
		), "UDP", ""},
		{"sip:alice:secretword@atlanta.com;transport=tcp", sipuri.New(
			"alice",
			"atlanta.com",
			sipuri.WithPassword("secretword"),
			sipuri.WithParams(sipuri.KeyValuePairs{
				"transport": {"tcp"},
			}),
		), "TCP", "RFC example #1"},
		{"sips:alice@atlanta.com?priority=urgent&subject=project%20x", sipuri.New(
			"alice",
			"atlanta.com",
			sipuri.Secure(),
			sipuri.WithHeaders(sipuri.KeyValuePairs{
				"subject":  {"project x"},
				"priority": {"urgent"},
			}),
		), "TCP", "RFC example #2"},
		{"sip:+1-212-555-1212:1234@gateway.com;user=phone", sipuri.New(
			"+1-212-555-1212",
			"gateway.com",
			sipuri.WithPassword("1234"),
			sipuri.WithParams(sipuri.KeyValuePairs{
				"user": {"phone"},
			}),
		), "UDP", "RFC example #3"},
		{"sips:1212@gateway.com", sipuri.New(
			"1212",
			"gateway.com",
			sipuri.Secure(),
		), "TCP", "RFC example #4"},
		{"sip:alice@192.0.2.4", sipuri.New(
			"alice",
			"192.0.2.4",
		), "UDP", "RFC example #5"},
		{"sip:atlanta.com;method=REGISTER?to=alice%40atlanta.com", sipuri.New(
			"",
			"atlanta.com",
			sipuri.WithParams(sipuri.KeyValuePairs{
				"method": {"REGISTER"},
			}),
			sipuri.WithHeaders(sipuri.KeyValuePairs{
				"to": {"alice@atlanta.com"},
			}),
		), "UDP", "RFC example #6"},
		{"sip:alice;day=tuesday@atlanta.com", sipuri.New(
			"alice;day=tuesday",
			"atlanta.com",
		), "UDP", "RFC example #7"},

		{"sip:j%40s0n@example.com", sipuri.New(
			"j@s0n",
			"example.com",
		), "UDP", "RFC example #8"},

		// Examples present on O'Reilly sites.
		// https://www.oreilly.com/library/view/the-ims-ip/9780470019061/9780470019061_the_sip_uri.html
		{"sip:bob.smith@nokia.com", sipuri.New(
			"bob.smith",
			"nokia.com",
		), "UDP", "O'Reilly example #1"},
		{"sip:bob@nokia.com;transport=tcp", sipuri.New(
			"bob",
			"nokia.com",
			sipuri.WithParams(sipuri.KeyValuePairs{
				"transport": {"tcp"},
			}),
		), "TCP", "O'Reilly example #2"},
		{"sip:+1-212-555-1234@gw.com;user=phone", sipuri.New(
			"+1-212-555-1234",
			"gw.com",
			sipuri.WithParams(sipuri.KeyValuePairs{
				"user": {"phone"},
			}),
		), "UDP", "O'Reilly example #3"},
		{"sip:root@136.16.20.100:8001", sipuri.New(
			"root",
			"136.16.20.100:8001",
		), "UDP", "O'Reilly example #4"},
		{"sip:bob.smith@registrar.com;method=REGISTER", sipuri.New(
			"bob.smith",
			"registrar.com",
			sipuri.WithParams(sipuri.KeyValuePairs{
				"method": {"REGISTER"},
			}),
		), "UDP", "O'Reilly example #5"},

		{"sip:[::]", sipuri.New(
			"", "[::]",
		), "UDP", "IPv6 local address"},
		{"sip:[::]:1111", sipuri.New(
			"", "[::]:1111",
		), "UDP", "IPv6 local address"},
	}

	for _, test := range tests {
		for _, parse := range parseFuncs {
			sipURI, err := parse(test.uri)
			if err != nil {
				t.Fatalf(`failed to parse SIP URI %q, %v error`, test.uri, err)
			}

			equalF(t, test.sipURI.Proto(), sipURI.Proto(), "protocol mismatch in %s", test.msg)
			equalF(t, test.sipURI.User(), sipURI.User(), "user mismatch in %s", test.msg)
			equalF(t, test.sipURI.Password(), sipURI.Password(), "password mismatch in %s", test.msg)
			equalF(t, test.sipURI.Host(), sipURI.Host(), "host mismatch in %s", test.msg)

			equalF(t, test.sipURI.Params().Encode(), sipURI.Params().Encode(), "param mismatch in %s", test.msg)
			equalF(t, test.sipURI.Headers().Encode(), sipURI.Headers().Encode(), "header mismatch in %s", test.msg)

			equalF(t, test.uri, sipURI.String(), "reconstructing string %s", test.msg)

			equalF(t, test.transp, sipURI.Transport(), "determining transport protocol %s", test.msg)
		}
	}
}

func TestParseError(t *testing.T) {
	t.Parallel()

	type test struct {
		uri string
		err error
		msg string
	}

	tests := []test{
		{
			"user@example.sip.twilio.com;transport=TCP",
			sipuri.ErrInvalidScheme,
			"no scheme present",
		},
		{
			"sip:user@",
			sipuri.MalformedURIError{Cause: sipuri.MissingHost},
			"no host present",
		},
		{
			"sip:@",
			sipuri.MalformedURIError{Cause: sipuri.MissingUser},
			"lonely at symbol",
		},
		{
			"sip:@;",
			sipuri.MalformedURIError{Cause: sipuri.MissingUser},
			"lonely at symbol",
		},
		{
			"sip:user@;",
			sipuri.MalformedURIError{Cause: sipuri.MissingHost},
			"lonely at symbol",
		},
		{
			"sip:@example.sip.twilio.com",
			sipuri.MalformedURIError{Cause: sipuri.MissingUser},
			"no user present",
		},
		{
			"sip:%xx@example.sip.twilio.com",
			sipuri.MalformedURIError{Cause: sipuri.MalformedUser},
			"malformed url encoded users",
		},
		{
			"sip:user@%xxexample.sip.twilio.com",
			sipuri.MalformedURIError{Cause: sipuri.MalformedHost},
			"malformed url encoded host",
		},
		{
			"sip:%xxexample.sip.twilio.com",
			sipuri.MalformedURIError{Cause: sipuri.MalformedHost},
			"malformed url encoded host",
		},
		{
			"sip:user@example.sip.twilio.com;%xx",
			sipuri.MalformedURIError{Cause: sipuri.MalformedParams},
			"malformed url encoded params",
		},
		{
			"sip:user@example.sip.twilio.com?%xx",
			sipuri.MalformedURIError{Cause: sipuri.MalformedHeaders},
			"malformed url encoded headers",
		},
		{
			"sip:[::1",
			sipuri.MalformedURIError{Cause: sipuri.MalformedHost},
			"malformed ipv6 host",
		},
	}

	for _, test := range tests {
		for _, parse := range parseFuncs {
			nul, err := parse(test.uri)

			if !errors.Is(err, test.err) {
				t.Fatalf(`expected error %q but got %q in %s`, test.err, err, test.msg)
			}

			equalF(t, (*sipuri.URI)(nil), nul, "nil recieved %s", test.msg)
		}
	}
}

func ExampleParse() {
	sipURI, err := sipuri.Parse("sip:user:password@host:port;uri-parameters?headers")
	if err != nil {
		panic(err)
	}

	// Print the consistent components
	fmt.Println(sipURI.User())
	fmt.Println(sipURI.Password())
	fmt.Println(sipURI.Host())
	fmt.Printf("%v\n", sipURI.Params())
	fmt.Printf("%v\n", sipURI.Headers())

	// Re-construct the URI
	fmt.Println(sipURI.String())

	// Output:
	// user
	// password
	// host:port
	// map[uri-parameters:[]]
	// map[headers:[]]
	// sip:user:password@host:port;uri-parameters=?headers=
}

func equalF(t *testing.T, e interface{}, g interface{}, m string, a ...interface{}) {
	t.Helper()

	if !reflect.DeepEqual(e, g) {
		t.Fatalf(`%q != %q, %s`, e, g, fmt.Sprintf(m, a...))
	}
}
