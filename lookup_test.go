package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oschwald/maxminddb-golang"
	"github.com/pmylund/go-cache"
)

const dbPath = "GeoLite2-City.mmdb"

func TestLookupByParam(t *testing.T) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()
	q.Add("ip", "8.8.8.8")
	req.URL.RawQuery = q.Encode()

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := LookupHandler{DB: db}

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"country":{"iso_code":"US"},"location":{"lat":37.751,"long":-97.822,"accuracy":1000}}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

// type NullResponseWriter struct{}

// func (n NullResponseWriter) Header() http.Header {
// 	return http.Header{}
// }
// func (n NullResponseWriter) Write([]byte) (int, error) {
// 	return 0, nil
// }
// func (n NullResponseWriter) WriteHeader(statusCode int) {}

func BenchmarkLookup(b *testing.B) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	ip := net.ParseIP("71.246.111.168")
	if ip == nil {
		b.Fatal("failure parsing benchmark ip?!")
	}

	handler := LookupHandler{
		DB:       db,
		MemCache: nil,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = handler.Lookup(ip)
	}
}

func BenchmarkRequest(b *testing.B) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	req, _ := http.NewRequest("GET", "/8.8.8.8", nil)
	rr := httptest.NewRecorder() //NullResponseWriter{}
	handler := LookupHandler{
		DB:       db,
		MemCache: nil,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkRequestWithCache(b *testing.B) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	req, _ := http.NewRequest("GET", "/8.8.8.8", nil)
	rr := httptest.NewRecorder()
	handler := LookupHandler{
		DB:       db,
		MemCache: cache.New(DefaultCacheExpiration, DefaultCacheCleanup),
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		handler.ServeHTTP(rr, req)
	}
}
