package sipuri

import (
	"sort"
	"strings"
)

// This file is an alternative to the stdlib module url with some
// tweaks to encode a ' ' as "%20" rather than a '+'.
//
// This module takes inspiration from the stdlib url package.
// All credit goes to the Go Devs in unstanding the RFC there.

type encoding int

const (
	encodeHost encoding = 1 + iota
	encodeUserPassword
	encodeQueryComponent
)

// shouldEscape returns if the given character should be escaped in the
// given context.
//
// Based on stdlib url.shouldEscape implementation & derived and checked with
// the RFC https://www.rfc-editor.org/rfc/rfc3986#section-2
//
//nolint:cyclop
func (mode encoding) shouldEscape(char byte) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || '0' <= char && char <= '9' {
		return false
	}

	if mode == encodeHost {
		// §3.2.2 Host allows:
		switch char {
		case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '[', ']', '<', '>', '"':
			return false
		}
	}

	switch char {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@': // §2.2 Reserved characters (reserved)
		// Different sections of the URI allow a few of
		// the reserved characters to appear unescaped.
		switch mode { //nolint:exhaustive
		case encodeUserPassword: // §3.2.1
			// The RFC allows ';', ':', '&', '=', '+', '$', and ',' in
			// userinfo, so we must escape only '@', '/', and '?'.
			// The parsing of userinfo treats ':' as special so we must escape
			// that too.
			return char == '@' || char == '/' || char == '?' || char == ':'
		case encodeQueryComponent: // §3.4
			// The RFC reserves (so we must escape) everything.
			return true
		}
	}

	// Everything else must be escaped.
	return true
}

// escape encodes characters based on the context of the string
//
// Based on url.escape but tweaked and optimised.
func escape(input string, mode encoding) string {
	var hexCount int

	for i := 0; i < len(input); i++ {
		if mode.shouldEscape(input[i]) {
			hexCount++
		}
	}

	// short-circuit in case no escaping is required
	if hexCount == 0 {
		return input
	}

	required := len(input) + 2*hexCount //nolint:gomnd
	result := make([]byte, required)

	escapeInto(input, 0, result)

	return string(result)
}

// DecodeURLValues decodes the input into the url.Values type, spliting
// key-value pairs on the separator.
func DecodeURLValues(input string, separator string) (KeyValuePairs, error) {
	pairs := strings.Split(input, separator)

	// len(pairs) is the maximum number of unique keys possible. This may
	// end up using more memory but in our use case duplicate keys are
	// unlikely making this a worthy optimisation.
	result := make(KeyValuePairs, len(pairs))

	for _, pair := range pairs {
		key, value, _ := strings.Cut(pair, "=")

		key, err := Unescape(key)
		if err != nil {
			return nil, err
		}

		value, err = Unescape(value)
		if err != nil {
			return nil, err
		}

		result[key] = append(result[key], value)
	}

	return result, nil
}

// EncodeURLValues encodes all non-alpha numeric byte values;
// notibly it encodes spaces as "%20" rather than a '+'.
//
// Based on url.Values.Encode() but encodes spaces differently.
// It is also slightly more efficient at 10% faster, with around 35% less
// bytes written & over half the allocations per operation.
//
//nolint:cyclop
func EncodeURLValues(input map[string][]string) string {
	// short-circuit in the empty case
	keyCount := len(input)
	if keyCount == 0 {
		return ""
	}

	var charCount, hexCount, keyValuesCount int

	keys := make([]string, 0, keyCount)
	for key, vals := range input {
		keys = append(keys, key)
		vsCount := len(vals)

		for i := 0; i < len(key); i++ {
			if encodeQueryComponent.shouldEscape(key[i]) {
				hexCount += vsCount
			}
		}

		charCount += len(key) * vsCount
		keyValuesCount += vsCount

		for _, val := range vals {
			for i := 0; i < len(val); i++ {
				if encodeQueryComponent.shouldEscape(val[i]) {
					hexCount++
				}
			}

			charCount += len(val)
		}
	}

	required := charCount + // total characters in the keys
		2*hexCount + // additional characters due to the encoding %xx that's two more x's
		2*keyValuesCount - 1 // separating & and =
	result := make([]byte, required)

	sort.Strings(keys)

	pos := 0

	for _, key := range keys {
		for _, val := range input[key] {
			if pos > 0 {
				result[pos] = '&'
				pos++
			}

			pos = escapeInto(key, pos, result)
			result[pos] = '='
			pos = escapeInto(val, pos+1, result)
		}
	}

	return string(result)
}

const upperhex = "0123456789ABCDEF"

// escapeInto escapes all of "input", writing the "result" into target
// starting at index "offset".
func escapeInto(input string, offset int, target []byte) int {
	for pos := 0; pos < len(input); pos++ {
		switch c := input[pos]; {
		case encodeQueryComponent.shouldEscape(c):
			target[offset] = '%'
			target[offset+1] = upperhex[c>>4]
			target[offset+2] = upperhex[c&15]
			offset += 3
		default:
			target[offset] = c
			offset++
		}
	}

	return offset
}

// Unescape URL decodes the input.
func Unescape(input string) (string, error) {
	// Count how many escaped bytes there are and
	// guarantee that they are all of 2 characters
	// in length.
	var hexCount int

	for i := 0; i < len(input); i++ {
		if input[i] == '%' {
			hexCount++

			i += 2

			// not enought characters for
			if i >= len(input) {
				return "", EscapeError(input[i-2:])
			}
		}
	}

	// short-circuit in case no unescaping is required
	if hexCount == 0 {
		return input, nil
	}

	required := len(input) - 2*hexCount //nolint:gomnd
	result := make([]byte, required)

	_, err := unescapeInto(input, 0, result)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// UnescapeErrorChecker scans the input checking for malformed encoded entities.
//
// It is a stripped down version of Unescape without actually extracting the parts
// or decoding the string it returns an error if and only if the aforementioned does.
func UnescapeErrorChecker(input string) error {
	l := len(input)
	if input[l-1] == '%' {
		return EscapeError("%")
	} else if input[l-2] == '%' {
		return EscapeError(input[l-2:])
	}

	for pos := 0; pos < l; pos++ {
		if input[pos] == '%' {
			gByte := checkValidHexCharacter(input[pos+1])
			lByte := checkValidHexCharacter(input[pos+2])

			if (gByte|lByte)&hexCharErrorBit != 0 {
				return EscapeError(input[pos : pos+3])
			}

			pos += 2
		}
	}

	return nil
}

// 10000 = 16 in decimal.
const hexCharErrorBit byte = 1 << 4

func unescapeInto(input string, offset int, target []byte) (int, error) {
	for pos := 0; pos < len(input); pos++ {
		switch c := input[pos]; {
		case c == '%':
			gByte := checkValidHexCharacter(input[pos+1])
			lByte := checkValidHexCharacter(input[pos+2])

			if (gByte|lByte)&hexCharErrorBit != 0 {
				return 0, EscapeError(input[pos : pos+3])
			}

			target[offset] = gByte<<4 + lByte //nolint:gomnd
			offset++

			pos += 2
		default:
			target[offset] = c
			offset++
		}
	}

	return offset, nil
}

func checkValidHexCharacter(hex byte) byte {
	const alphabetStartIdx = 10

	// Relies on the follow ranges being sequantial in ASCII/UTF-8 encoding.
	switch {
	case 'A' <= hex && hex <= 'F':
		return hex - 'A' + alphabetStartIdx
	case 'a' <= hex && hex <= 'f':
		return hex - 'a' + alphabetStartIdx
	case '0' <= hex && hex <= '9':
		return hex - '0'
	}

	return hexCharErrorBit
}

// KeyValueStore provides access to a multi-valued map.
type KeyValueStore interface {
	// Get returns the first value for the given key. Empty string otherwise.
	Get(key string) string
	// Encode stringifies the multi-valued map, url encoding keys and values
	// joining with an ampersand.
	Encode() string
	// Len returns the number of distinct keys.
	Len() int
	// Empty returns if the store contains no keys.
	Empty() bool
}

// KeyValuePairs stores key to values similar to that of [url.Values]
// and implements [KeyValueStore].
type KeyValuePairs map[string][]string

// Decode populates the Store with the given data, returing any encoding errors
// encountered.
func (m *KeyValuePairs) Decode(input, separator string) error {
	var err error
	*m, err = DecodeURLValues(input, separator)

	return err
}

// Get returns the first value for the given key. Empty string otherwise.
func (m KeyValuePairs) Get(key string) string {
	if m == nil {
		return ""
	}

	vs := m[key]
	if len(vs) == 0 {
		return ""
	}

	return vs[0]
}

// Encode stringifies the multi-valued map, url encoding keys and values
// joining with an ampersand.
func (m KeyValuePairs) Encode() string {
	return EncodeURLValues(m)
}

// Len returns the number of distinct keys.
func (m KeyValuePairs) Len() int {
	return len(m)
}

// Empty returns if the store contains no keys.
func (m KeyValuePairs) Empty() bool {
	return len(m) == 0
}

// EmptyStore represents an always empty multi-valued map.
type EmptyStore struct{}

// Decode populates the Store with the given data, returing any encoding errors
// encountered.
func (EmptyStore) Decode(_, _ string) error {
	return nil
}

// Get returns the first value for the given key. Empty string otherwise.
func (EmptyStore) Get(_ string) string {
	return ""
}

// Encode stringifies the multi-valued map, url encoding keys and values
// joining with an ampersand.
func (EmptyStore) Encode() string {
	return ""
}

// Len returns the number of distinct keys.
func (EmptyStore) Len() int {
	return 0
}

// Empty returns if the store contains no keys.
func (EmptyStore) Empty() bool {
	return true
}

// LazyStore lazily loads a [KeyValuePairs] struct when inspected.
type LazyStore struct {
	KeyValuePairs
	input     string
	separator string
}

// Decode populates the Store with the given data. Always scans the input for encoding errors.
func (s *LazyStore) Decode(input, separator string) error {
	s.input = input
	s.separator = separator

	return UnescapeErrorChecker(input)
}

// Get returns the first value for the given key. Empty string otherwise.
func (s *LazyStore) Get(key string) string {
	s.load()

	return s.KeyValuePairs.Get(key)
}

// Encode stringifies the multi-valued map, url encoding keys and values
// joining with an ampersand.
func (s *LazyStore) Encode() string {
	s.load()

	return s.KeyValuePairs.Encode()
}

// Len returns the number of distinct keys.
func (s *LazyStore) Len() int {
	s.load()

	return s.KeyValuePairs.Len()
}

// Empty returns if the store contains no keys.
func (s *LazyStore) Empty() bool {
	if s.KeyValuePairs != nil {
		return s.KeyValuePairs.Empty()
	}

	return s.input == ""
}

func (s *LazyStore) load() {
	if s.KeyValuePairs != nil {
		return
	}

	// Any possible errors have already been checked in the Decode
	// call to [UnescapeErrorChecker].
	//nolint:errcheck,gosec
	(&s.KeyValuePairs).Decode(s.input, s.separator)

	s.input = ""
	s.separator = ""
}
