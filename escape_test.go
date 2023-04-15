package sipuri_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/percivalalb/sipuri"
)

func TestURLEncodeURLValues(t *testing.T) {
	t.Parallel()

	query := getTestURLValues()
	got := sipuri.EncodeURLValues(query)

	equalF(t, testQueryString, got, "encodeURLValues(%v) = %q want %q", query, got, testQueryString)
}

func TestUnescape(t *testing.T) {
	t.Parallel()

	expect := "cat=meow&dog=bark!&dog=woof@&mouse=ee  eeΔ&parrot=(hellow)"
	got, _ := sipuri.Unescape(testQueryString)

	equalF(t, expect, got, "Unescape(%q) = %q want %q", testQueryString, got, expect)
}

func TestUnescapeError(t *testing.T) {
	t.Parallel()

	_, err := sipuri.Unescape("bark%2y")

	if !errors.Is(err, sipuri.EscapeError("%2y")) {
		t.Fatalf("err %v", err)
	}

	_, err = sipuri.Unescape("bark%2")

	if !errors.Is(err, sipuri.EscapeError("%2")) {
		t.Fatalf("err %v", err)
	}

	_, err = sipuri.Unescape("bark%")

	if !errors.Is(err, sipuri.EscapeError("%")) {
		t.Fatalf("err %v", err)
	}
}

// func FuzzReverse(f *testing.F) {
// 	testcases := []string{"Hello, world", " ", "!12345"}
// 	for _, tc := range testcases {
// 		f.Add(tc) // Use f.Add to provide a seed corpus
// 	}
// 	f.Fuzz(func(t *testing.T, orig string) {
// 		query, err := url.ParseQuery(orig)
// 		if err != nil {
// 			return
// 		}

// 		doubleRev := sipuri.EncodeURLValues(query)
// 		if orig != doubleRev {
// 			t.Errorf("Before: %q, after: %q", orig, doubleRev)
// 		}
// 	})
// }

func BenchmarkURLValues_Encode(b *testing.B) {
	query := getTestURLValues()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = query.Encode()
	}
}

func BenchmarkEncodeURLValues(b *testing.B) {
	query := getTestURLValues()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = sipuri.EncodeURLValues(query)
	}
}

func BenchmarkURLUnescape(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = url.QueryUnescape(testQueryString)
	}
}

func BenchmarkUnescape(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = sipuri.Unescape(testQueryString)
	}
}

const testQueryString = "cat=meow&dog=bark%21&dog=woof%40&mouse=ee%20%20ee%CE%94&parrot=%28hellow%29"

func getTestURLValues() url.Values {
	query := make(url.Values)
	query.Add("dog", "bark!")
	query.Add("dog", "woof@")
	query.Add("cat", "meow")
	query.Add("parrot", "(hellow)")
	query.Add("mouse", "ee  eeΔ")

	return query
}
