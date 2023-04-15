package sipuri

import (
	"net/url"
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
func DecodeURLValues(input string, separator string) (url.Values, error) {
	pairs := strings.Split(input, separator)

	// len(pairs) is the maximum number of unique keys possible. This may
	// end up using more memory but in our use case duplicate keys are
	// unlikely making this a worthy optimisation.
	result := make(url.Values, len(pairs))

	for _, pair := range pairs {
		split := strings.SplitN(pair, "=", 2) //nolint:gomnd

		key, err := Unescape(split[0])
		if err != nil {
			return nil, err
		}

		value, err := Unescape(split[1])
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
func EncodeURLValues(input url.Values) string {
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
