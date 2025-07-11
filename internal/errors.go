package internal

import "strconv"

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
