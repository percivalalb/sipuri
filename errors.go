package sipuri

import (
	"strings"
)

// InvalidSchemeError is returned when a string that does not start sip: or sips: is given.
type InvalidSchemeError struct{}

// Error returns a string representation of the error.
func (err InvalidSchemeError) Error() string {
	return "sip: scheme invalid"
}

// MalformedURIError encapsulates an error while processing a sip or sips URI.
type MalformedURIError struct {
	Desc string
	Err  error
}

// Error returns a string representation of the error.
func (err MalformedURIError) Error() string {
	var builder strings.Builder

	builder.WriteString("sip: malformed uri")

	if err.Desc != "" {
		builder.WriteString(": " + err.Desc)
	}

	if err.Err != nil {
		builder.WriteString(": " + err.Err.Error())
	}

	return builder.String()
}

// Is returns if the given error is also a MalformedURIError struct.
//
// Overridden since [MalformedURIError] is comparable (as all it's fields are
// comparable). We want errors.Is to match if it is just the same struct type.
func (err MalformedURIError) Is(input error) bool {
	_, ok := input.(MalformedURIError) //nolint:errorlint

	return ok
}

// Unwrap returns the underlying error.
func (err MalformedURIError) Unwrap() error {
	return err.Err
}
