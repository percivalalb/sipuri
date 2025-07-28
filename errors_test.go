package sipuri_test

import (
	"errors"
	"testing"

	"github.com/percivalalb/sipuri/v2"
)

func TestMalformedError(t *testing.T) {
	t.Parallel()

	unspecifiedErr := sipuri.MalformedError{} // zero value
	missingUserErr := sipuri.MalformedError{Cause: sipuri.MissingUser}
	missingHostErr := sipuri.MalformedError{Cause: sipuri.MissingHost}

	if !errors.Is(missingUserErr, unspecifiedErr) {
		t.Fatalf("unspecified cause matches any malform error")
	}

	if errors.Is(unspecifiedErr, missingUserErr) {
		t.Fatalf("specific cause does not match an unspecified cause")
	}

	if errors.Is(missingHostErr, missingUserErr) {
		t.Fatalf("specific cause matches does not match another cause")
	}

	var malformError sipuri.MalformedError
	if !errors.As(missingUserErr, &malformError) {
		t.Fatalf("unspecified cause malform error")
	}

	equalF(t, missingUserErr.Cause, malformError.Cause, "error.As for same cause")

	equalF(t, "sip: malformed URI", unspecifiedErr.Error(), "unspecified cause string representation")
	equalF(t, "sip: malformed URI: missing user", missingUserErr.Error(), "missing user cause string representation")
	equalF(t, "sip: malformed URI: missing host", missingHostErr.Error(), "missing host cause string representation")
}

func TestMalformCause(t *testing.T) {
	t.Parallel()

	tests := []sipuri.MalformCause{
		sipuri.Unspecified, sipuri.InvalidScheme, sipuri.MissingUser, sipuri.MissingHost,
		sipuri.MalformedUser, sipuri.MalformedParams, sipuri.MalformedHeaders,
	}

	for _, test := range tests {
		if test.String() == "" {
			t.Fatalf("malform cause %d returned no string description", test)
		}
	}
}
