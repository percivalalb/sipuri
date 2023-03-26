package sipuri_test

import (
	"net/url"
	"testing"

	"github.com/percivalalb/sipuri"
)

func TestURLEncodeURLValues(t *testing.T) {
	t.Parallel()

	query := getTestURLValues()

	got := sipuri.EncodeURLValues(query)
	expect := "cat=meow&dog=bark%21&dog=woof%40&mouse=ee%20%20ee%CE%94&parrot=%28hellow%29"

	if got != expect {
		t.Fatalf("encodeURLValues(%v) = %q want %q", query, got, expect)
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

func getTestURLValues() url.Values {
	query := make(url.Values)
	query.Add("dog", "bark!")
	query.Add("dog", "woof@")
	query.Add("cat", "meow")
	query.Add("parrot", "(hellow)")
	query.Add("mouse", "ee  eeΔ")

	return query
}
