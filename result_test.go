package tinygeoip

import (
	"encoding/json"
	"io"
	"testing"
)

func BenchmarkLookupResult_EncodeJSON(b *testing.B) {
	tc := testCases[2].expected

	encoder := json.NewEncoder(io.Discard)
	for b.Loop() {
		encoder.Encode(tc)
	}
}
