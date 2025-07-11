package internal

import "sync"

// KeyValueStore provides access to a multi-valued map.
type KeyValueStore interface {
	// Get returns the first value for the given key. Empty string otherwise.
	Get(key string) string
	// GetAll returns all the values for the given key. Nil slice otherwise.
	GetAll(key string) []string
	// Keys returns all the keys in no particular order. Nil slice if there are none.
	Keys() []string
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

// GetAll returns all the values for the given key. Nil slice otherwise.
func (m KeyValuePairs) GetAll(key string) []string {
	if m == nil {
		return nil
	}

	vs := m[key]
	if len(vs) == 0 {
		return nil
	}

	c := make([]string, len(vs))
	copy(c, vs)

	return c
}

// Keys returns all the keys in no particular order. Nil slice if there are none.
func (m KeyValuePairs) Keys() []string {
	l := len(m)
	if l == 0 {
		return nil
	}

	keys := make([]string, 0, l)

	for k := range m {
		keys = append(keys, k)
	}

	return keys
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

// GetAll returns all the values for the given key. Nil slice otherwise.
func (EmptyStore) GetAll(_ string) []string {
	return nil
}

// Keys returns all the keys in no particular order. Nil slice if there are none.
func (EmptyStore) Keys() []string {
	return nil
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

	once      sync.Once // protects the above
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

// GetAll returns all the values for the given key. Nil slice otherwise.
func (s *LazyStore) GetAll(key string) []string {
	s.load()

	return s.KeyValuePairs.GetAll(key)
}

// Keys returns all the keys in no particular order. Nil slice if there are none.
func (s *LazyStore) Keys() []string {
	s.load()

	return s.KeyValuePairs.Keys()
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
	s.once.Do(func() {
		// Any possible errors have already been checked in the Decode
		// call to [UnescapeErrorChecker].
		//nolint:errcheck,gosec
		(&s.KeyValuePairs).Decode(s.input, s.separator)

		s.input = ""
		s.separator = ""
	})
}
