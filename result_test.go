package tinygeoip

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFastJSON(t *testing.T) {
	for _, tc := range testCases {
		res := tc.expected
		expected, _ := json.Marshal(res)
		actual := res.FastJSON()
		if !bytes.Equal(expected, actual) {
			t.Errorf("JSON mismatch! want %s, got %s", expected, actual)
		}
	}
}

func TestFasterJSON(t *testing.T) {
	for _, tc := range testCases {
		res := tc.expected
		expected, _ := json.Marshal(res)
		actual := res.FasterJSON()
		if !bytes.Equal(expected, *actual) {
			t.Errorf("JSON mismatch! want %s, got %s", expected, *actual)
		}
		res.PoolReturn(actual)
	}
}

func BenchmarkDBResultJSON(b *testing.B) {
	res := testCases[2].expected

	b.Run("Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			json.Marshal(res)
		}
	})
	b.Run("FastJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			res.FastJSON()
		}
	})
	b.Run("FasterJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bs := res.FasterJSON()
			res.PoolReturn(bs)
		}
	})
}
