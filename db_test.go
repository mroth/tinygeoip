package geominder

import (
	"flag"
	"net"
	"reflect"
	"sync"
	"testing"
)

// location of the test data database
const testdbPath = "testdata/GeoIP2-City-Test.mmdb"

// allow the database used to be overridden at test time
var dbPath = flag.String("db", testdbPath, "path of GeoIP database to use during tests")

// These test cases are hard coded based on values present in the
// GeoIP2-City-Test test database.  If you override the DB and use a more up to
// date production dataset, they will likely fail due to changes in the data.
//
// We intentionally use a test database here for size reasons and to try to
// keep the tests "evergreen" for unit testing our DB parsing rather than the
// data itself.
var testCases = []struct {
	ip       net.IP
	expected LookupResult
}{
	{net.ParseIP("89.160.20.112"), LookupResult{
		Country: country{
			ISOCode: "SE",
		}, Location: location{
			Latitude:  58.4167,
			Longitude: 15.6167,
			Accuracy:  76,
		},
	}},
	{net.ParseIP("81.2.69.142"), LookupResult{
		Country: country{
			ISOCode: "GB",
		}, Location: location{
			Latitude:  51.5142,
			Longitude: -0.0931,
			Accuracy:  10,
		},
	}},
	{net.ParseIP("2001:218:85a3:0000:0000:8a2e:0370:7334"), LookupResult{
		Country: country{
			ISOCode: "JP",
		}, Location: location{
			Latitude:  35.68536,
			Longitude: 139.75309,
			Accuracy:  100,
		},
	}},

	// Fun fact? Easter egg? Amusing to me only? But these test IPs are the
	// locations I was working on writing this code....
	// SO SAD, cant use my real ones anymore, figure out how to bring them back for nostalgia before deleting from code
	// // my home fios connection at some point
	// {net.ParseIP("71.246.111.168"), LookupResult{
	// 	Country: country{
	// 		ISOCode: "US",
	// 	}, Location: location{
	// 		Latitude:  40.7095,
	// 		Longitude: -73.9563,
	// 		Accuracy:  5,
	// 	},
	// }},
	// // COALMINE coffee in Seoul
	// {net.ParseIP("175.211.82.153"), LookupResult{
	// 	Country: country{
	// 		ISOCode: "KR",
	// 	}, Location: location{
	// 		Latitude:  37.5985,
	// 		Longitude: 126.9783,
	// 		Accuracy:  10,
	// 	},
	// }},
	// // 히피 도끼 (고인물), 서울시 [커피 + 코딩]
	// // Hippytokki, Goinmool, Seoul (Coffee + Code meetup)
	// {net.ParseIP("121.131.15.99"), LookupResult{
	// 	Country: country{
	// 		ISOCode: "KR",
	// 	}, Location: location{
	// 		Latitude:  37.5333,
	// 		Longitude: 126.95,
	// 		Accuracy:  10,
	// 	},
	// }},
	// {net.ParseIP("8.8.8.8"), LookupResult{
	// 	Country: country{
	// 		ISOCode: "US",
	// 	}, Location: location{
	// 		Latitude:  37.751,
	// 		Longitude: -97.822,
	// 		Accuracy:  1000,
	// 	},
	// }},
	// {net.ParseIP("1.1.1.1"), LookupResult{
	// 	Country: country{
	// 		ISOCode: "AU",
	// 	}, Location: location{
	// 		Latitude:  -33.494,
	// 		Longitude: 143.2104,
	// 		Accuracy:  1000,
	// 	},
	// }},
}

func TestDBLookup(t *testing.T) {
	db, err := NewLookupDB(*dbPath)
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

func TestDBFastLookup(t *testing.T) {
	db, err := NewLookupDB(*dbPath)
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
		err := db.FastLookup(tc.ip, res)
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
	db, err := NewLookupDB(*dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Lookup(testCases[0].ip)
	}
}

func BenchmarkDBFastLookup(b *testing.B) {
	db, err := NewLookupDB(*dbPath)
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
		db.FastLookup(testCases[0].ip, res)
		pool.Put(res)
	}
}
