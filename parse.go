package sipuri

import (
	"strings"

	"github.com/percivalalb/sipuri/v2/internal"
)

// Parse parses the given URI.
func Parse(uri string) (*URI, error) {
	if strings.HasPrefix(uri, SIPProtocol) {
		return parse(SIP, uri[len(SIPProtocol):], false)
	}

	if strings.HasPrefix(uri, SIPSProtocol) {
		return parse(SIPS, uri[len(SIPSProtocol):], false)
	}

	return nil, MalformedError{Cause: InvalidScheme}
}

// ParseLazy parses the given URI, lazily decoding the URI params & headers.
//
// Prefer [Parse] over this function. This version is intended where latency
// is paramount and params & headers are not inspected.
func ParseLazy(uri string) (*URI, error) {
	if strings.HasPrefix(uri, SIPProtocol) {
		return parse(SIP, uri[len(SIPProtocol):], true)
	}

	if strings.HasPrefix(uri, SIPSProtocol) {
		return parse(SIPS, uri[len(SIPSProtocol):], true)
	}

	return nil, MalformedError{Cause: InvalidScheme}
}

//nolint:cyclop,funlen
func parse(proto Protocol, uri string, lazy bool) (*URI, error) {
	sipURI := URI{proto: proto}

	// @ in the set of reserved chars of the user portion. Therefore the first
	userinfo, postfix, hasAt := strings.Cut(uri, "@") // @ must be encoded in the host and pass

	if hasAt {
		// §19.1.1 "If the @ sign is present in a SIP or SIPS URI, the user field MUST NOT be empty."
		if userinfo == "" {
			return nil, MalformedError{Cause: MissingUser}
		}
	} else {
		userinfo, postfix = postfix, userinfo // swap (makes userinfo empty)
	}

	// RFC requires : to be escaped in the userinfo. So split on :.
	sipURI.user, sipURI.pass, sipURI.hadPass = strings.Cut(userinfo, ":")

	user, err := internal.Unescape(sipURI.user)
	if err != nil {
		return nil, MalformedError{Cause: MalformedUser, Err: err}
	}

	sipURI.user = user

	// The uri must have been a single '@'
	if postfix == "" {
		return nil, MalformedError{Cause: MissingHost}
	}

	prefix, headers, hadHeader := strings.Cut(postfix, "?")
	host, params, hadParam := strings.Cut(prefix, ";")

	// §19.1.2 host mandatory in all contexts
	if host == "" {
		return nil, MalformedError{Cause: MissingHost}
	}

	sipURI.hadHeader = hadHeader
	sipURI.hadParam = hadParam

	// Typically the host should not contain any escaped characters but
	// it is possible in the spec.
	host, err = internal.Unescape(host)
	if err != nil {
		return nil, MalformedError{Cause: MalformedHost, Err: err}
	}

	sipURI.host = host

	// Check the host port is not malformed
	if _, _, err := sipURI.SplitHostPort(); err != nil {
		return nil, MalformedError{Cause: MalformedHost, Err: err}
	}

	switch {
	case params == "":
		sipURI.params = internal.EmptyStore{}
	case lazy:
		var temp internal.LazyStore
		if err := (&temp).Decode(params, ";"); err != nil {
			return nil, MalformedError{Cause: MalformedParams, Err: err}
		}

		sipURI.params = &temp
	default:
		var temp internal.KeyValuePairs
		if err := (&temp).Decode(params, ";"); err != nil {
			return nil, MalformedError{Cause: MalformedParams, Err: err}
		}

		sipURI.params = temp
	}

	switch {
	case headers == "":
		sipURI.headers = internal.EmptyStore{}
	case lazy:
		var temp internal.LazyStore
		if err := (&temp).Decode(headers, "&"); err != nil {
			return nil, MalformedError{Cause: MalformedHeaders, Err: err}
		}

		sipURI.headers = &temp
	default:
		var temp internal.KeyValuePairs
		if err := (&temp).Decode(headers, "&"); err != nil {
			return nil, MalformedError{Cause: MalformedHeaders, Err: err}
		}

		sipURI.headers = temp
	}

	return &sipURI, nil
}
