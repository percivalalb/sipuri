// Package sipuri parses SIP or SIPS URI into their constitution components.
//
// A general SIP uri looks like:
//
//	sip:user:password@host:port;uri-parameters?headers
//
// From https://www.rfc-editor.org/rfc/rfc3261#section-19
package sipuri

import (
	"net"
	"strings"
)

// The two sip protocols.
const (
	SIPProtocol  = "sip:"
	SIPSProtocol = "sips:"
)

// Protocol represents the protocol/scheme used. SIP or SIPS.
type Protocol bool

// The two SIP protocols.
const (
	SIP  Protocol = false
	SIPS Protocol = true
)

// URI stores the components of that make up a SIP URI.
type URI struct {
	proto Protocol // default is SIP

	user    string
	pass    string
	host    string
	params  KeyValueStore
	headers KeyValueStore

	hadPass   bool
	hadParam  bool
	hadHeader bool
}

type uriOption func(u *URI)

// WithParams allows URI params to be set.
func WithParams(params KeyValueStore) uriOption {
	return func(u *URI) {
		u.params = params
	}
}

// WithHeaders allows URI headers to be set.
func WithHeaders(headers KeyValueStore) uriOption {
	return func(u *URI) {
		u.headers = headers
	}
}

// WithPassword allows the password portion of the user-info to be set.
//
// Use of a password is not advised and is inherently insecure. Use other
// methods to ensure communication.
func WithPassword(pass string) uriOption {
	return func(u *URI) {
		u.pass = pass
	}
}

// Secure upgrades the URI to the SIPS protocol.
func Secure() uriOption {
	return func(u *URI) {
		u.proto = SIPS
	}
}

// New constructs a SIP URI with the given options.
func New(user, host string, opts ...uriOption) URI {
	u := URI{
		user: user,
		host: host,
	}

	for _, opt := range opts {
		opt(&u)
	}

	return u
}

// Transport returns the Transport protocols that would be used to make a
// connection to the host.
func (sipURI URI) Transport() string {
	if transport := sipURI.Params().Get("transport"); transport != "" {
		return strings.ToUpper(transport)
	}

	// §19.1.2 "The default transport is scheme dependent. For sip:, it is UDP. For sips:, it is TCP."
	switch sipURI.proto {
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
	if sipURI.proto == SIPS {
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

// String rebuilds the string representation of the URI respecting the quirks of the input.
//
//nolint:cyclop
func (sipURI URI) String() string {
	var builder strings.Builder

	switch sipURI.proto {
	case SIPS:
		builder.WriteString(SIPSProtocol)
	case SIP:
		builder.WriteString(SIPProtocol)
	}

	if sipURI.user != "" {
		builder.WriteString(escape(sipURI.user, encodeUserPassword))

		if sipURI.hadPass || sipURI.pass != "" {
			builder.WriteRune(':')
		}

		if sipURI.pass != "" {
			builder.WriteString(escape(sipURI.pass, encodeUserPassword))
		}

		builder.WriteByte('@') // only present when user is non-empty
	}

	builder.WriteString(escape(sipURI.host, encodeHost))

	if sipURI.hadParam || !sipURI.Params().Empty() {
		builder.WriteByte(';')
	}

	if !sipURI.Params().Empty() {
		builder.WriteString(sipURI.Params().Encode())
	}

	if sipURI.hadHeader || !sipURI.Headers().Empty() {
		builder.WriteByte('?')
	}

	if !sipURI.Headers().Empty() {
		builder.WriteString(sipURI.Headers().Encode())
	}

	return builder.String()
}

// Secure returns if the URI has been upgrade to the SIPS scheme.
func (sipURI URI) Secure() Protocol {
	return sipURI.proto == SIPS
}

// Proto returns what scheme the SIP URI is.
func (sipURI URI) Proto() Protocol {
	return sipURI.proto
}

// User returns the decoded user portion of the URI.
func (sipURI URI) User() string {
	return sipURI.user
}

// Password returns the decoded password portion of the URI.
func (sipURI URI) Password() string {
	return sipURI.pass
}

// Host returns the decoded host portion of the URI.
//
// You may want to use SplitHostPort.
func (sipURI URI) Host() string {
	return sipURI.host
}

// SplitHostPort splits the port from the host portion into.
func (sipURI URI) SplitHostPort() (string, string, error) {
	ipv6 := len(sipURI.host) > 0 && sipURI.host[0] == '['
	colonCount := strings.Count(sipURI.host, ":")

	if (!ipv6 && colonCount > 0) || (ipv6 && (colonCount%2 == 1 || sipURI.host[len(sipURI.host)-1] != ']')) {
		return net.SplitHostPort(sipURI.host) //nolint:wrapcheck
	}

	return sipURI.host, "", nil
}

// Params returns the decoded params portion of the URI.
func (sipURI URI) Params() KeyValueStore {
	if sipURI.params == nil {
		return EmptyStore{}
	}

	return sipURI.params
}

// Headers returns the decoded headers portion of the URI.
func (sipURI URI) Headers() KeyValueStore {
	if sipURI.headers == nil {
		return EmptyStore{}
	}

	return sipURI.headers
}
