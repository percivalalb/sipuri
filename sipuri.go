// Package to parse SIP or SIPS URI into the constitution components.
//
// A general SIP uri looks like:
//
//	sip:user:password@host:port;uri-parameters?headers
//
// From https://www.rfc-editor.org/rfc/rfc3261#section-19
package sipuri

import (
	"net"
	"net/url"
	"strings"
)

const (
	SIPProtocol  = "sip:"
	SIPSProtocol = "sips:"
)

// Protocol represents the protocol/scheme used. SIP or SIPS.
type Protocol bool

const (
	SIP  Protocol = false
	SIPS Protocol = true
)

// URI stores the components of that make up a SIP URI.
type URI struct {
	Proto Protocol // default is SIP

	User    string
	Pass    string
	Host    string
	Params  KeyValueStore
	Headers KeyValueStore

	hadPass   bool
	hadParam  bool
	hadHeader bool
}

// SplitHostPort splits the port from the host portion into.
func (sipURI URI) SplitHostPort() (string, string, error) {
	ipv6 := len(sipURI.Host) > 0 && sipURI.Host[0] == '['
	colonCount := strings.Count(sipURI.Host, ":")

	if (!ipv6 && colonCount > 0) || (ipv6 && (colonCount%2 == 1 || sipURI.Host[len(sipURI.Host)-1] != ']')) {
		return net.SplitHostPort(sipURI.Host) //nolint:wrapcheck
	}

	return sipURI.Host, "", nil
}

// Transport returns the Transport protocols that would be used to make a
// connection to the host.
func (sipURI URI) Transport() string {
	if transport := sipURI.Params.Get("transport"); transport != "" {
		return strings.ToUpper(transport)
	}

	// §19.1.2 "The default transport is scheme dependent. For sip:, it is UDP. For sips:, it is TCP."
	switch sipURI.Proto {
	case SIP:
		return "UDP"
	case SIPS:
		return "TCP"
	default:
		panic("unreachable")
	}
}

// Port returns the port split from the host portion returning the
// defaults based on transport protocol & scheme if not present.
//
// Returns an empty string in the case of the sip proto & unexpected transport.
func (sipURI URI) Port() string {
	_, port, _ := sipURI.SplitHostPort()

	if port != "" {
		return port
	}

	// §19.1.2 says "The default port value is transport and scheme dependent.
	// The default is 5060 for sip: using UDP, TCP, or SCTP. The default
	// is 5061 for sip: using TLS over TCP and sips: over TCP."
	if sipURI.Proto == SIPS {
		return "5061"
	}

	switch sipURI.Transport() {
	case "UDP", "TCP", "SCTP":
		return "5060"
	// "The default is 5061 for sip: using TLS over TCP"
	case "TLS":
		return "5061"
	}

	return ""
}

// Strings rebuilds the string representation of the URI respecting the quirks of the input.
//
//nolint:cyclop
func (sipURI URI) String() string {
	var builder strings.Builder

	switch sipURI.Proto {
	case SIPS:
		builder.WriteString(SIPSProtocol)
	case SIP:
		builder.WriteString(SIPProtocol)
	}

	if sipURI.User != "" {
		builder.WriteString(escape(sipURI.User, encodeUserPassword))

		if sipURI.hadPass || sipURI.Pass != "" {
			builder.WriteRune(':')
		}

		if sipURI.Pass != "" {
			builder.WriteString(escape(sipURI.Pass, encodeUserPassword))
		}

		builder.WriteByte('@') // only present when user is non-empty
	}

	builder.WriteString(escape(sipURI.Host, encodeHost))

	if sipURI.hadParam || !sipURI.Params.Empty() {
		builder.WriteByte(';')
	}

	if !sipURI.Params.Empty() {
		builder.WriteString(sipURI.Params.Encode())
	}

	if sipURI.hadHeader || !sipURI.Headers.Empty() {
		builder.WriteByte('?')
	}

	if !sipURI.Headers.Empty() {
		builder.WriteString(sipURI.Headers.Encode())
	}

	return builder.String()
}

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
	sipURI := URI{Proto: proto}

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
	sipURI.User, sipURI.Pass, sipURI.hadPass = strings.Cut(userinfo, ":")

	// Hack: to work around stdlib decoding "+" as whitespace.
	user, err := url.QueryUnescape(strings.ReplaceAll(sipURI.User, "+", "%2B"))
	if err != nil {
		return nil, MalformedURIError{Cause: MalformedUser, Err: err}
	}

	sipURI.User = user

	// Typically the host should not contain any escaped characters but
	// it is possible in the spec.
	host, err = url.QueryUnescape(host)
	if err != nil {
		return nil, MalformedURIError{Cause: MalformedHost, Err: err}
	}

	sipURI.Host = host

	if params == "" {
		sipURI.Params = EmptyStore{}
	} else if lazy {
		var temp LazyStore // &LazyStore{}
		if err := (&temp).Decode(params, ";"); err != nil {
			return nil, MalformedURIError{Cause: MalformedParams, Err: err}
		}
		sipURI.Params = &temp
	} else {
		var temp KeyValuePairs // &LazyStore{}
		if err := (&temp).Decode(params, ";"); err != nil {
			return nil, MalformedURIError{Cause: MalformedParams, Err: err}
		}
		sipURI.Params = temp
	}

	if headers == "" {
		sipURI.Headers = EmptyStore{}
	} else if lazy {
		var temp LazyStore
		if err := (&temp).Decode(headers, "&"); err != nil {
			return nil, MalformedURIError{Cause: MalformedParams, Err: err}
		}
		sipURI.Params = &temp
	} else {
		var temp KeyValuePairs // &LazyStore{}
		if err := (&temp).Decode(headers, "&"); err != nil {
			return nil, MalformedURIError{Cause: MalformedHeaders, Err: err}
		}
		sipURI.Headers = temp
	}

	// Check the host port is not malformed
	if _, _, err := sipURI.SplitHostPort(); err != nil {
		return nil, MalformedURIError{Cause: MalformedHost, Err: err}
	}

	return &sipURI, nil
}
