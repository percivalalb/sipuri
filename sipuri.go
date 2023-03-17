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
	Params  url.Values
	Headers url.Values

	hadPass   bool
	hadParam  bool
	hadHeader bool
}

// SplitHostPort splits the port from the host portion into.
func (sipURI URI) SplitHostPort() (string, string, error) {
	if strings.Contains(sipURI.Host, ":") {
		return net.SplitHostPort(sipURI.Host)
	}

	return sipURI.Host, "", nil
}

// Transport returns the Transport protocols
func (sipURI URI) Transport() string {
	if transport := sipURI.Params.Get("transport"); transport != "" {
		return strings.ToUpper(transport)
	}

	// §19.1.2 "The default transport is scheme dependent. For sip:, it is UDP. For sips:, it is TCP."
	switch sipURI.Proto {
	case SIP:
		return "UDP"
	default: //case SIPS:
		return "TCP"
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
	switch sipURI.Proto {
	case SIPS:
		return "5061"
		// TODO: Handle case for sip using TLS over TCP.
	}

	// case SIP:

	switch sipURI.Transport() {
	case "UDP":
	case "TCP":
	case "SCTP":
		return "5060"
	}

	return ""
}

// Strings rebuilds the string representation of the URI respecting the
func (sipURI URI) String() string {
	var sb strings.Builder

	switch sipURI.Proto {
	case SIPS:
		sb.WriteString(SIPSProtocol)
	default:
		sb.WriteString(SIPProtocol)
	}

	if sipURI.User != "" {
		// TODO: QueryEscapes all reserved chars.
		// The RFC allows ';', ':', '&', '=', '+', '$', and ',' in
		// userinfo, so we must escape only '@', '/', and '?'.

		sb.WriteString(escape(sipURI.User, encodeUserPassword))

		if sipURI.hadPass || sipURI.Pass != "" {
			sb.WriteRune(':')
		}

		if sipURI.Pass != "" {
			sb.WriteString(escape(sipURI.Pass, encodeUserPassword))
		}

		sb.WriteByte('@') // only present when user is non-empty
	}

	sb.WriteString(escape(sipURI.Host, encodeHost))

	if sipURI.hadParam || len(sipURI.Params) > 0 {
		sb.WriteByte(';')
	}

	if len(sipURI.Params) > 0 {
		sb.WriteString(encodeURLValues(sipURI.Params))
	}

	if sipURI.hadHeader || len(sipURI.Headers) > 0 {
		sb.WriteByte('?')
	}

	if len(sipURI.Headers) > 0 {
		sb.WriteString(encodeURLValues(sipURI.Headers))
	}

	return sb.String()
}

// Parse parses the given uri.
func Parse(uri string) (*URI, error) {
	if strings.HasPrefix(uri, SIPProtocol) {
		return parse(SIP, uri[len(SIPProtocol):])
	}

	if strings.HasPrefix(uri, SIPSProtocol) {
		return parse(SIPS, uri[len(SIPSProtocol):])
	}

	return nil, ErrInvalidScheme{}
}

func parse(proto Protocol, uri string) (*URI, error) {
	sipURI := URI{
		Proto: proto,
	}

	// @ in the set of reserved chars of the user portion. Therefore the first
	userinfo, postfix, hasAt := strings.Cut(uri, "@") // @ must be encoded in the host and pass

	if hasAt {
		// §19.1.1 "If the @ sign is present in a SIP or SIPS URI, the user field MUST NOT be empty."
		if userinfo == "" {
			return nil, ErrMalformedURI{}
		}
	} else {
		userinfo, postfix = postfix, userinfo // swap (makes userinfo empty)
	}

	// The uri must have been a single '@'
	if postfix == "" {
		return nil, ErrMalformedURI{}
	}

	prefix, headers, hadHeader := strings.Cut(postfix, "?")
	host, params, hadParam := strings.Cut(prefix, ";")

	// §19.1.2 host mandatory in all contexts
	if host == "" {
		return nil, ErrMalformedURI{}
	}

	sipURI.hadHeader = hadHeader
	sipURI.hadParam = hadParam

	// RFC requires : to be escaped in the userinfo. So split on :.
	sipURI.User, sipURI.Pass, sipURI.hadPass = strings.Cut(userinfo, ":")

	var err error
	// TODO: propertly unescape userinfo.
	sipURI.User = strings.ReplaceAll(sipURI.User, "+", "%2B")
	sipURI.User, err = url.QueryUnescape(sipURI.User) // encodes + since go libary decodes + as whitespace
	if err != nil {
		return nil, ErrMalformedURI{Err: err}
	}

	sipURI.Host = host

	if params != "" {
		var err error
		if sipURI.Params, err = url.ParseQuery(params); err != nil {
			return nil, ErrMalformedURI{Err: err}
		}
	}

	if headers != "" {
		var err error
		if sipURI.Headers, err = url.ParseQuery(headers); err != nil {
			return nil, ErrMalformedURI{Err: err}
		}
	}

	return &sipURI, nil
}
