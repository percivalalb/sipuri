package sipuri

import (
	"net/url"
	"sort"
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
func (mode encoding) shouldEscape(c byte) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
		return false
	}

	if mode == encodeHost {
		// §3.2.2 Host allows:
		switch c {
		case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '[', ']', '<', '>', '"':
			return false
		}
	}

	switch c {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@': // §2.2 Reserved characters (reserved)
		// Different sections of the URI allow a few of
		// the reserved characters to appear unescaped.
		switch mode {
		case encodeUserPassword: // §3.2.1
			// The RFC allows ';', ':', '&', '=', '+', '$', and ',' in
			// userinfo, so we must escape only '@', '/', and '?'.
			// The parsing of userinfo treats ':' as special so we must escape
			// that too.
			return c == '@' || c == '/' || c == '?' || c == ':'
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
func escape(s string, mode encoding) string {
	var hexCount int
	for i := 0; i < len(s); i++ {
		if mode.shouldEscape(s[i]) {
			hexCount++
		}
	}

	// short-circuit in case no escaping is required
	if hexCount == 0 {
		return s
	}

	required := len(s) + 2*hexCount
	t := make([]byte, required)

	escapeInto(s, 0, t)

	return string(t)
}

// encodeURLValues encodes all non-alpha numeric byte values;
// notibly it encodes spaces as "%20" rather than a '+'.
//
// Based on url.Values.Encode() but encodes spaces differently.
// It is also slightly more efficent at 10% faster, with around 35% less
// bytes written & over half the allocations per operation.
func encodeURLValues(input url.Values) string {
	// short-circuit in the empty case
	keyCount := len(input)
	if keyCount == 0 {
		return ""
	}

	var charCount, hexCount, keyValuesCount int

	keys := make([]string, 0, keyCount)
	for k, vs := range input {
		keys = append(keys, k)
		vsCount := len(vs)

		for i := 0; i < len(k); i++ {
			if encodeQueryComponent.shouldEscape(k[i]) {
				hexCount += vsCount
			}
		}

		charCount += len(k) * vsCount
		keyValuesCount += vsCount

		for _, v := range vs {
			for i := 0; i < len(v); i++ {
				if encodeQueryComponent.shouldEscape(v[i]) {
					hexCount++
				}
			}

			charCount += len(v)
		}
	}

	required := charCount + // total characters in the keys
		2*hexCount + // additional characters due to the encoding %xx that's two more x's
		2*keyValuesCount - 1 // seperating & and =
	t := make([]byte, required)

	sort.Strings(keys)

	j := 0
	for _, k := range keys {
		for _, v := range input[k] {
			if j > 0 {
				t[j] = '&'
				j++
			}

			j = escapeInto(k, j, t)
			t[j] = '='
			j = escapeInto(v, j+1, t)
		}
	}

	return string(t)
}

const upperhex = "0123456789ABCDEF"

// escapeInto escapes all of v, writing the result into t starting at index 0.
func escapeInto(v string, o int, t []byte) int {
	for i := 0; i < len(v); i++ {
		switch c := v[i]; {
		case encodeQueryComponent.shouldEscape(c):
			t[o] = '%'
			t[o+1] = upperhex[c>>4]
			t[o+2] = upperhex[c&15]
			o += 3
		default:
			t[o] = v[i]
			o++
		}
	}

	return o
}
