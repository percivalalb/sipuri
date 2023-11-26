package sipuri

import (
	"errors"
	"strconv"
	"strings"
)

// ErrInvalidScheme is returned when a string that does not start sip: or sips: is given.
var ErrInvalidScheme = errors.New("sip: scheme invalid")

// MalformCause indicates what part of the URI failed to be parsed.
type MalformCause uint8

// The possible reasons a URI could be malformed. The cause which relates to the
// earliest part of the URI is returned.
const (
	Unspecified MalformCause = iota
	MissingUser
	MissingHost
	MalformedUser
	MalformedHost
	MalformedParams
	MalformedHeaders
)

// String returns a description of the cause.
func (c MalformCause) String() string {
	switch c {
	case Unspecified:
		return "unspecified"
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

// MalformedURIError encapsulates an error while processing a sip or sips URI.
type MalformedURIError struct {
	Cause MalformCause
	Err   error
}

// Error returns a string representation of the error.
func (err MalformedURIError) Error() string {
	var builder strings.Builder

	builder.WriteString("sip: malformed uri")

	if err.Cause != Unspecified {
		builder.WriteString(": " + err.Cause.String())
	}

	if err.Err != nil {
		builder.WriteString(": " + err.Err.Error())
	}

	return builder.String()
}

// Is returns if the given error is also a [MalformedURIError] struct of the same cause.
//
// If the input does not have a cause specified then it matches any
// [MalformedURIError] struct.
func (err MalformedURIError) Is(input error) bool {
	var inputMal MalformedURIError
	if errors.As(input, &inputMal) {
		return inputMal.Cause == Unspecified || inputMal.Cause == err.Cause
	}

	return false
}

// Unwrap returns the underlying error.
func (err MalformedURIError) Unwrap() error {
	return err.Err
}

// EscapeError is returned when a byte-pair has been incorrectly URL encoded.
type EscapeError string

// Error returns the string representation of the error.
func (e EscapeError) Error() string {
	return "sip: invalid URL escape " + strconv.Quote(string(e))
}

// Is allows [EscapeError] to be compared by [errors.Is].
func (e EscapeError) Is(input error) bool {
	_, ok := input.(EscapeError)

	return ok
}
