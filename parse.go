package sipuri

import "strings"

// Parse parses the given uri.
func Parse(uri string) (*URI, error) {
	if strings.HasPrefix(uri, SIPProtocol) {
		return parse(SIP, uri[len(SIPProtocol):], false)
	}

	if strings.HasPrefix(uri, SIPSProtocol) {
		return parse(SIPS, uri[len(SIPSProtocol):], false)
	}

	return nil, ErrInvalidScheme
}

// Parse parses the given uri.
func ParseLazy(uri string) (*URI, error) {
	if strings.HasPrefix(uri, SIPProtocol) {
		return parse(SIP, uri[len(SIPProtocol):], true)
	}

	if strings.HasPrefix(uri, SIPSProtocol) {
		return parse(SIPS, uri[len(SIPSProtocol):], true)
	}

	return nil, ErrInvalidScheme
}

//nolint:cyclop,funlen
func parse(proto Protocol, uri string, lazy bool) (*URI, error) {
	sipURI := URI{proto: proto}

	// @ in the set of reserved chars of the user portion. Therefore the first
	userinfo, postfix, hasAt := strings.Cut(uri, "@") // @ must be encoded in the host and pass

	if hasAt {
		// §19.1.1 "If the @ sign is present in a SIP or SIPS URI, the user field MUST NOT be empty."
		if userinfo == "" {
			return nil, MalformedURIError{Cause: MissingUser}
		}
	} else {
		userinfo, postfix = postfix, userinfo // swap (makes userinfo empty)
	}

	// The uri must have been a single '@'
	if postfix == "" {
		return nil, MalformedURIError{Cause: MissingHost}
	}

	prefix, headers, hadHeader := strings.Cut(postfix, "?")
	host, params, hadParam := strings.Cut(prefix, ";")

	// §19.1.2 host mandatory in all contexts
	if host == "" {
		return nil, MalformedURIError{Cause: MissingHost}
	}

	sipURI.hadHeader = hadHeader
	sipURI.hadParam = hadParam

	// RFC requires : to be escaped in the userinfo. So split on :.
	sipURI.user, sipURI.pass, sipURI.hadPass = strings.Cut(userinfo, ":")

	user, err := Unescape(sipURI.user)
	if err != nil {
		return nil, MalformedURIError{Cause: MalformedUser, Err: err}
	}

	sipURI.user = user

	// Typically the host should not contain any escaped characters but
	// it is possible in the spec.
	host, err = Unescape(host)
	if err != nil {
		return nil, MalformedURIError{Cause: MalformedHost, Err: err}
	}

	sipURI.host = host

	// Check the host port is not malformed
	if _, _, err := sipURI.SplitHostPort(); err != nil {
		return nil, MalformedURIError{Cause: MalformedHost, Err: err}
	}

	if params == "" {
		sipURI.params = EmptyStore{}
	} else if lazy {
		var temp LazyStore // &LazyStore{}
		if err := (&temp).Decode(params, ";"); err != nil {
			return nil, MalformedURIError{Cause: MalformedParams, Err: err}
		}

		sipURI.params = &temp
	} else {
		var temp KeyValuePairs // &LazyStore{}
		if err := (&temp).Decode(params, ";"); err != nil {
			return nil, MalformedURIError{Cause: MalformedParams, Err: err}
		}

		sipURI.params = temp
	}

	if headers == "" {
		sipURI.headers = EmptyStore{}
	} else if lazy {
		var temp LazyStore
		if err := (&temp).Decode(headers, "&"); err != nil {
			return nil, MalformedURIError{Cause: MalformedHeaders, Err: err}
		}

		sipURI.headers = &temp
	} else {
		var temp KeyValuePairs
		if err := (&temp).Decode(headers, "&"); err != nil {
			return nil, MalformedURIError{Cause: MalformedHeaders, Err: err}
		}

		sipURI.headers = temp
	}

	return &sipURI, nil
}
