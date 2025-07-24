package internal

import "strconv"

// URIEscapeError is returned when a '%' character in a URL string is not
// followed by a valid hexadecimal byte.
type URIEscapeError string

// Error returns the string representation of the error.
func (e URIEscapeError) Error() string {
	return "sip: invalid URL escape " + strconv.Quote(string(e))
}

// Is allows [EscapeError] to be compared by [errors.Is].
func (e URIEscapeError) Is(input error) bool {
	_, ok := input.(URIEscapeError)

	return ok
}
