package sipuri

import (
	"log"
	"net/url"
	"testing"
)

var query url.Values

func init() {
	query = make(url.Values)
	query.Add("dog", "bark!")
	query.Add("dog", "woof@")
	query.Add("cat", "meow")
	query.Add("parrot", "(hellow)")
	query.Add("mouse", "ee  eeΔ")
}

func TestURLEncodeURLValues(t *testing.T) {
	got := encodeURLValues(query)
	expect := "cat=meow&dog=bark%21&dog=woof%40&mouse=ee%20%20ee%CE%94&parrot=%28hellow%29"
	if got != expect {
		log.Fatalf("encodeURLValues(%v) = %q want %q", query, got, expect)
	}
}

func BenchmarkURLValues_Encode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = query.Encode()
	}
}

func BenchmarkEncodeURLValues(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = encodeURLValues(query)
	}
}
