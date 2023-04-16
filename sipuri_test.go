package sipuri_test

import (
	"testing"

	"github.com/percivalalb/sipuri"
)

func TestNew(t *testing.T) {
	t.Parallel()

	uri := sipuri.New(
		"user",
		"host:port",
		sipuri.WithPassword("password"),
		sipuri.WithParams(sipuri.KeyValuePairs{
			"uri-parameters": {""},
		}),
		sipuri.WithHeaders(sipuri.KeyValuePairs{
			"headers": {""},
		}),
	)

	equalF(t, sipuri.SIP, uri.Proto(), "protocol mismatch")
	equalF(t, "user", uri.User(), "user mismatch")
	equalF(t, "password", uri.Password(), "password mismatch")
	equalF(t, "host:port", uri.Host(), "host mismatch")
}
