package sipuri

import (
	"errors"
	"strings"

	"github.com/percivalalb/sipuri/internal"
)

// MalformCause indicates what part of the URI failed to be parsed.
type MalformCause uint8

// Below are the possible reasons a URI could be malformed. The causes are ordered
// in the order they are checked.
const (
	Unspecified MalformCause = iota
	InvalidScheme
	MissingUser
	MalformedUser
	MissingHost
	MalformedHost
	MalformedParams
	MalformedHeaders
)

// String returns a description of the cause.
func (c MalformCause) String() string {
	switch c {
	case Unspecified:
		return "unspecified"
	case InvalidScheme:
		return "invalid scheme"
	case MissingUser:
		return "missing user"
	case MissingHost:
		return "missing host"
	case MalformedUser:
		return "malformed user"
	case MalformedHost:
		return "malformed host"
	case MalformedParams:
		return "malformed params"
	case MalformedHeaders:
		return "malformed headers"
	default:
		panic("unreachable")
	}
}

// MalformedError encapsulates an error while parsing a sip or sips URI.
type MalformedError struct {
	Cause MalformCause
	Err   error
}

// Error returns a string representation of the error.
func (err MalformedError) Error() string {
	var builder strings.Builder

	builder.WriteString("sip: malformed URI")

	if err.Cause != Unspecified {
		builder.WriteString(": " + err.Cause.String())
	}

	if err.Err != nil {
		builder.WriteString(": " + err.Err.Error())
	}

	return builder.String()
}

// Is returns if the given error is also a [MalformedError] struct of the same cause.
//
// If the input does not have a cause specified then it matches any
// [MalformedError] struct.
func (err MalformedError) Is(input error) bool {
	var inputMal MalformedError
	if errors.As(input, &inputMal) {
		return inputMal.Cause == Unspecified || inputMal.Cause == err.Cause
	}

	return false
}

// Unwrap returns the underlying error.
func (err MalformedError) Unwrap() error {
	return err.Err
}

// EscapeError is returned when a '%' character in a URL string is not
// followed by a valid hexadecimal byte.
type EscapeError = internal.URIEscapeError
