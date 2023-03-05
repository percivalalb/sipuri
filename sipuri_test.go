package sipuri_test

import (
	"testing"

	"github.com/percivalalb/sipuri"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()

	type test struct {
		uri    string
		sipURI sipuri.URI
		transp string
		msg    string
	}

	tests := []test{
		{"sip:user:password@host:port;uri-parameters=?headers=",
			sipuri.URI{
				User: "user",
				Pass: "password",
				Host: "host:port",
				Params: map[string][]string{
					"uri-parameters": {""},
				},
				Headers: map[string][]string{
					"headers": {""},
				},
			},
			"UDP",
			"template uri",
		},

		// From https://www.rfc-editor.org/rfc/rfc3261#section-19.1.1
		{"sip:alice@atlanta.com", sipuri.URI{
			User: "alice",
			Host: "atlanta.com",
		}, "UDP", ""},
		{"sip:alice:secretword@atlanta.com;transport=tcp", sipuri.URI{
			User: "alice",
			Pass: "secretword",
			Host: "atlanta.com",
			Params: map[string][]string{
				"transport": {"tcp"},
			},
		}, "TCP", ""},
		{"sips:alice@atlanta.com?priority=urgent&subject=project%20x", sipuri.URI{
			Proto: sipuri.SIPS,
			User:  "alice",
			Host:  "atlanta.com",
			Headers: map[string][]string{
				"subject":  {"project x"},
				"priority": {"urgent"},
			},
		}, "TCP", ""},
		{"sip:+1-212-555-1212:1234@gateway.com;user=phone", sipuri.URI{
			User: "+1-212-555-1212",
			Pass: "1234",
			Host: "gateway.com",
			Params: map[string][]string{
				"user": {"phone"},
			},
		}, "UDP", ""},
		{"sips:1212@gateway.com", sipuri.URI{
			Proto: sipuri.SIPS,
			User:  "1212",
			Host:  "gateway.com",
		}, "TCP", ""},
		{"sip:alice@192.0.2.4", sipuri.URI{
			User: "alice",
			Host: "192.0.2.4",
		}, "UDP", ""},
		{"sip:atlanta.com;method=REGISTER?to=alice%40atlanta.com", sipuri.URI{
			User: "",
			Pass: "",
			Host: "atlanta.com",
			Params: map[string][]string{
				"method": {"REGISTER"},
			},
			Headers: map[string][]string{
				"to": {"alice@atlanta.com"},
			},
		}, "UDP", ""},
		{"sip:alice;day=tuesday@atlanta.com", sipuri.URI{
			User: "alice;day=tuesday",
			Pass: "",
			Host: "atlanta.com",
		}, "UDP", ""},

		{"sip:j%40s0n@example.com", sipuri.URI{
			User: "j@s0n",
			Pass: "",
			Host: "example.com",
		}, "UDP", ""},

		// Examples present on O'Reilly sites.
		// https://www.oreilly.com/library/view/the-ims-ip/9780470019061/9780470019061_the_sip_uri.html
		{"sip:bob.smith@nokia.com", sipuri.URI{
			User: "bob.smith",
			Host: "nokia.com",
		}, "UDP", ""},

		{"sip:bob@nokia.com;transport=tcp", sipuri.URI{
			User: "bob",
			Host: "nokia.com",
			Params: map[string][]string{
				"transport": {"tcp"},
			},
		}, "TCP", ""},

		{"sip:+1-212-555-1234@gw.com;user=phone", sipuri.URI{
			User: "+1-212-555-1234",
			Host: "gw.com",
			Params: map[string][]string{
				"user": {"phone"},
			},
		}, "UDP", ""},

		{"sip:root@136.16.20.100:8001", sipuri.URI{
			User: "root",
			Host: "136.16.20.100:8001",
		}, "UDP", ""},

		{"sip:bob.smith@registrar.com;method=REGISTER", sipuri.URI{
			User: "bob.smith",
			Host: "registrar.com",
			Params: map[string][]string{
				"method": {"REGISTER"},
			},
		}, "UDP", ""},

		{"alb@t2hws4-netcraft.sip.twilio.com;transport=TCP", sipuri.URI{}, "", "ada"},
		{"sip:alb@", sipuri.URI{}, "", "asdas"},
		{"sip:@", sipuri.URI{}, "", "asdas"},
	}

	for _, test := range tests {
		sipURI, err := sipuri.Parse(test.uri)

		if test.sipURI.Host == "" {
			assert.Error(t, err, "expected error %s", test.msg)
			continue
		}

		assert.NoError(t, err)

		assert.Equalf(t, test.sipURI.Proto, sipURI.Proto, "protocol mismatch in %s", test.msg)
		assert.Equalf(t, test.sipURI.User, sipURI.User, "user mismatch in %s", test.msg)
		assert.Equalf(t, test.sipURI.Pass, sipURI.Pass, "password mismatch in %s", test.msg)
		assert.Equalf(t, test.sipURI.Host, sipURI.Host, "host mismatch in %s", test.msg)

		assert.Equalf(t, test.sipURI.Params.Encode(), sipURI.Params.Encode(), "param mismatch in %s", test.msg)
		assert.Equalf(t, test.sipURI.Headers.Encode(), sipURI.Headers.Encode(), "header mismatch in %s", test.msg)

		assert.Equalf(t, test.uri, sipURI.String(), "reconstructing string %s", test.msg)

		assert.Equalf(t, test.transp, sipURI.Transport(), "reconstructing string %s", test.msg)
	}
}
