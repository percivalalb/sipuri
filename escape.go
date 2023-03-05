package sipuri

import (
	"net/url"
	"sort"
	"strings"
)

// This file is a copy of the stdlib module url with some tweaks to encode a ' ' as "%20" rather than a '+'
// All credit goes there for unstanding of the RFC
type encoding int

const (
	encodeHost encoding = 1 + iota
	encodeUserPassword
	encodeQueryComponent
)

const upperhex = "0123456789ABCDEF"

func shouldEscape(c byte, mode encoding) bool {
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
		// Different sections of the URL allow a few of
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

func escape(s string, mode encoding) string {
	var hexCount int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c, mode) {
			hexCount++
		}
	}

	if hexCount == 0 {
		return s
	}

	required := len(s) + 2*hexCount
	t := make([]byte, required)

	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case shouldEscape(c, mode):
			t[j] = '%'
			t[j+1] = upperhex[c>>4]
			t[j+2] = upperhex[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

// EncodeURLValues has the same as url.Values.Encode() but encodes spaces as "%20" rather than a '+'.
func EncodeURLValues(v url.Values) string {
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		keyEscaped := escape(k, encodeQueryComponent)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(escape(v, encodeQueryComponent))
		}
	}
	return buf.String()
}
