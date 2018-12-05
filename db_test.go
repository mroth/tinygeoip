package main

import (
	"net"
	"reflect"
	"sync"
	"testing"
)

const dbPath = "data/GeoLite2-City.mmdb" // TODO: replace with testdb
// TODO: for benchmark purposes, once testdb is implemented, see if we
// can also require the full db when -long or something is passed.

// Fun fact? Easter egg? Amusing to me only? But these test IPs are the
// locations I was working on writing this code.
var testCases = []struct {
	ip       net.IP
	expected LookupResult
}{
	// my home fios connection at some point
	{net.ParseIP("71.246.111.168"), LookupResult{
		Country: country{
			ISOCode: "US",
		}, Location: location{
			Latitude:  40.7095,
			Longitude: -73.9563,
			Accuracy:  5,
		},
	}},
	// COALMINE coffee in Seoul
	{net.ParseIP("175.211.82.153"), LookupResult{
		Country: country{
			ISOCode: "KR",
		}, Location: location{
			Latitude:  37.5985,
			Longitude: 126.9783,
			Accuracy:  10,
		},
	}},
	// 히피 도끼 (고인물), 서울시 [커피 + 코딩]
	// Hippytokki, Goinmool, Seoul (Coffee + Code meetup)
	{net.ParseIP("121.131.15.99"), LookupResult{
		Country: country{
			ISOCode: "KR",
		}, Location: location{
			Latitude:  37.5333,
			Longitude: 126.95,
			Accuracy:  10,
		},
	}},
	{net.ParseIP("8.8.8.8"), LookupResult{
		Country: country{
			ISOCode: "US",
		}, Location: location{
			Latitude:  37.751,
			Longitude: -97.822,
			Accuracy:  1000,
		},
	}},
	{net.ParseIP("1.1.1.1"), LookupResult{
		Country: country{
			ISOCode: "AU",
		}, Location: location{
			Latitude:  -33.494,
			Longitude: 143.2104,
			Accuracy:  1000,
		},
	}},
}

func TestDBLookup(t *testing.T) {
	db, err := NewLookupDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, tc := range testCases {
		actual, err := db.Lookup(tc.ip)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(&tc.expected, actual) {
			t.Errorf("testing: %v, want: %+v, got: %+v", tc.ip, tc.expected, actual)
		}
	}
}

func TestDBLookupFast(t *testing.T) {
	db, err := NewLookupDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	pool := &sync.Pool{
		New: func() interface{} {
			return new(LookupResult)
		},
	}

	for _, tc := range testCases {
		res := pool.Get().(*LookupResult)
		err := db.LookupFast(tc.ip, res)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(&tc.expected, res) {
			t.Errorf("testing: %v, want: %+v, got: %+v", tc.ip, tc.expected, res)
		}
		pool.Put(res)
	}
}

func BenchmarkDBLookup(b *testing.B) {
	db, err := NewLookupDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Lookup(testCases[0].ip)
	}
}

func BenchmarkDBLookupFast(b *testing.B) {
	db, err := NewLookupDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	pool := &sync.Pool{
		New: func() interface{} {
			return new(LookupResult)
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := pool.Get().(*LookupResult)
		db.LookupFast(testCases[0].ip, res)
		pool.Put(res)
	}
}
