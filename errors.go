package sipuri

import (
	"fmt"
)

// ErrInvalidScheme is returned when a string that does not start sip: or sips: is given.
type ErrInvalidScheme struct {
}

// Error returns a string representation of the error.
func (err ErrInvalidScheme) Error() string {
	return "sip: scheme invalid"
}

// ErrMalformedURI encapsulates an error while processing a sip or sips URI.
type ErrMalformedURI struct {
	Desc string
	Err  error
}

// Error returns a string representation of the error.
func (err ErrMalformedURI) Error() string {
	if err.Desc == "" {
		return fmt.Sprintf("sip: malformed uri: %s", err.Err.Error())
	}

	return fmt.Sprintf("sip: malformed uri: %s: %s", err.Desc, err.Err.Error())
}

// Unwrap returns the underlying error.
func (err ErrMalformedURI) Unwrap() error {
	return err.Err
}
