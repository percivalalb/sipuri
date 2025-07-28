package internal_test

import (
	"sort"
	"testing"

	"github.com/percivalalb/sipuri/v2/internal"
)

func TestKeyValueStore(t *testing.T) {
	t.Parallel()

	type test struct {
		input        internal.KeyValueStore
		expectValues map[string][]string
		expectKeys   []string
		msg          string
	}

	tests := []test{
		{
			internal.KeyValuePairs{},
			map[string][]string{
				"unknown": nil,
			},
			nil,
			"empty key/value pairs",
		},
		{
			internal.KeyValuePairs{
				"key1": {""},
			},
			map[string][]string{
				"key1": {""},
			},
			[]string{"key1"},
			"key/value pairs single key",
		},
		{
			internal.KeyValuePairs{
				"animals":   {"cat", "dog", "parrot"},
				"health":    {"good", "good", "good"},
				"locations": {"england"},
			},
			map[string][]string{
				"animals":   {"cat", "dog", "parrot"},
				"health":    {"good", "good", "good"},
				"locations": {"england"},
			},
			[]string{"animals", "health", "locations"},
			"many key/value pairs",
		},
		{
			internal.EmptyStore{},
			map[string][]string{
				"unknown": nil,
			},
			nil,
			"empty store",
		},
		{
			createLazyStore(t, "key1=value1&key2=value2&key3=value2&key1=value1", "&"),
			map[string][]string{
				"key1":    {"value1", "value1"},
				"key2":    {"value2"},
				"key3":    {"value2"},
				"unknown": nil,
			},
			[]string{"key1", "key2", "key3"},
			"lazy store",
		},
	}

	for _, test := range tests {
		for k, v := range test.expectValues {
			var firstV string
			if len(v) > 0 {
				firstV = v[0]
			}

			equalF(t, firstV, test.input.Get(k), "%s: first value mismatch", test.msg)
			equalF(t, v, test.input.GetAll(k), "%s: values mismatch", test.msg)
		}

		// Sorts as keys are returned in an arbitrary order.
		keys := test.input.Keys()
		sort.Strings(keys)

		equalF(t, test.expectKeys, keys, "%s: keys mismatch", test.msg)

		equalF(t, len(test.expectKeys), test.input.Len(), "%s: len mismatch", test.msg)
	}
}

func createLazyStore(t *testing.T, input, separator string) *internal.LazyStore {
	t.Helper()

	s := &internal.LazyStore{}

	if err := s.Decode(input, separator); err != nil {
		t.Fatalf("err %v", err)
	}

	return s
}
