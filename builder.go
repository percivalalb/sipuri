package sipuri

import "github.com/percivalalb/sipuri/internal"

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

type uriOption func(u *URI)

// Params is a throwaway type used by [WithParams] in the [New]
// builder to hide internal types.
type Params = map[string][]string

// WithParams allows URI params to be set.
//
// It takes a map of string keys paired to a slice of string values.
// The underlying type is map[string][]string. [net/url.Values] can also
// be used as a drop in. Example:
//
//	sipuri.Params{
//		"param1": {"value1", "value2"},
//	}
func WithParams(params Params) uriOption {
	return func(u *URI) {
		u.params = internal.KeyValuePairs(params)
		u.hadParam = true
	}
}

// Headers is a throwaway type used by [WithHeaders] in the [New]
// builder to hide internal types.
type Headers = map[string][]string

// WithHeaders allows URI headers to be set.
//
// It takes a map of string keys paired to a slice of string values.
// The underlying type is map[string][]string. [net/url.Values] can also
// be used as a drop in. Example:
//
//	sipuri.Headers{
//		"header1": {"value1", "value2"},
//	}
func WithHeaders(headers Headers) uriOption {
	return func(u *URI) {
		u.headers = internal.KeyValuePairs(headers)
		u.hadHeader = true
	}
}

// WithPassword allows the password portion of the user-info to be set.
//
// Use of a password is not advised and is inherently insecure. Use other
// methods to ensure communication.
func WithPassword(pass string) uriOption {
	return func(u *URI) {
		u.pass = pass
		u.hadPass = true
	}
}

// Secure upgrades the URI to the SIPS protocol.
func Secure() uriOption {
	return func(u *URI) {
		u.proto = SIPS
	}
}
