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
	{net.ParseIP("2001:220::1337"), LookupResult{
		Country: country{
			ISOCode: "KR",
		}, Location: location{
			Latitude:  37,
			Longitude: 127.5,
			Accuracy:  100,
		},
	}},
}

// use a ipv6 ip for benchmarks since node geoip-lite thinks it's harder
var benchIP = testCases[2].ip

func TestDBLookup(t *testing.T) {
	db := newTestDB(t)
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
	db := newTestDB(t)
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
	db := newTestDB(b)
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Lookup(benchIP)
	}
}

func BenchmarkDBFastLookup(b *testing.B) {
	db := newTestDB(b)
	defer db.Close()

	pool := &sync.Pool{
		New: func() interface{} {
			return new(LookupResult)
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := pool.Get().(*LookupResult)
		db.FastLookup(benchIP, res)
		pool.Put(res)
	}
}

// newTestDB calls NewLookupDB with the default test db (which can be overriden
// in flags), or causes the originating test/benchmark to fail if it errors.
//
// literally the only reason this exists it to save us the err != nil check 3
// lines of boilerplate visual noise on every single test initialization.
//
// worth it? IMHO heck yes!
func newTestDB(tb testing.TB) *LookupDB {
	db, err := NewLookupDB(*dbPath)
	if err != nil {
		tb.Fatal(err)
	}
	return db
}
