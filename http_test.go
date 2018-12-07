package geominder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmylund/go-cache"
)

const testIPv4Path1 = "/?ip=89.160.20.112"
const testIPv4Path2 = "/?ip=81.2.69.142"
const testIPv6Path1 = "/?ip=2001:218:85a3:0000:0000:8a2e:0370:7334"

const testIPv4Body1 = `{"country":{"iso_code":"SE"},"location":{"lat":58.4167,"long":15.6167,"accuracy":76}}`
const testIPv4Body2 = `{"country":{"iso_code":"GB"},"location":{"lat":51.5142,"long":-0.0931,"accuracy":10}}`
const testIPv6Body1 = `{"country":{"iso_code":"JP"},"location":{"lat":35.68536,"long":139.75309,"accuracy":100}}`

func TestHTTPLookup(t *testing.T) {
	var httpCases = []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
		expectedBody   string
	}{
		{
			name:           "happy1 IPv4",
			path:           testIPv4Path1,
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			expectedBody:   testIPv4Body1,
		},
		{
			name:           "happy2 IPv4",
			path:           testIPv4Path2,
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			expectedBody:   testIPv4Body2,
		},
		{
			name:           "happy IPv6",
			path:           testIPv6Path1,
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			expectedBody:   testIPv6Body1,
		},
		{
			// re-request the first valid path after other path requests, in
			// order to make certain about exercising cache validity
			name:           "happy1 IPv4 repeated",
			path:           testIPv4Path1,
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			expectedBody:   testIPv4Body1,
		},
		{
			name:           "request empty",
			path:           "/",
			expectedStatus: http.StatusBadRequest,
			expectedType:   "application/json",
			expectedBody:   `{"error": "missing IP query parameter, try ?ip=foo"}`,
		},
		{
			name:           "IP empty",
			path:           "/?ip=",
			expectedStatus: http.StatusBadRequest,
			expectedType:   "application/json",
			expectedBody:   `{"error": "could not parse IP address"}`,
		},
		{
			name:           "IP malformed",
			path:           "/?ip=192.168.a.b.c",
			expectedStatus: http.StatusBadRequest,
			expectedType:   "application/json",
			expectedBody:   `{"error": "could not parse invalid IP address"}`,
		},
		{
			name:           "IP not found",
			path:           "/?ip=127.0.0.1",
			expectedStatus: http.StatusInternalServerError,
			expectedType:   "application/json",
			expectedBody:   `{"error": "no match for 127.0.0.1 found in database"}`,
		},
	}

	db, err := NewLookupDB(*dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// test both with a cache and without
	rawHandler := HTTPHandler{DB: db, MemCache: nil}
	cachedHandler := HTTPHandler{DB: db, MemCache: cache.New(DefaultCacheExpiration, DefaultCacheCleanup)}
	var handlerCases = []struct {
		name    string
		handler HTTPHandler
	}{
		{name: "raw", handler: rawHandler},
		{name: "cached", handler: cachedHandler},
	}
	for _, hc := range handlerCases {
		t.Run(hc.name, func(t *testing.T) {

			for _, tc := range httpCases {
				t.Run(tc.name, func(t *testing.T) {
					req, err := http.NewRequest(http.MethodGet, tc.path, nil)
					if err != nil {
						t.Fatal(err)
					}

					rr := httptest.NewRecorder()
					hc.handler.ServeHTTP(rr, req)

					// check the status code is what we expect
					if status := rr.Code; status != tc.expectedStatus {
						t.Errorf("handler returned wrong status code: got %v want %v",
							status, tc.expectedStatus)
					}

					// check content type is what we expect
					if ct := rr.Header().Get("content-type"); ct != tc.expectedType {
						t.Errorf("handler returned wrong content-type: got %v want %v",
							ct, tc.expectedType)
					}

					// check the response body is valid json
					if bytes := rr.Body.Bytes(); !json.Valid(bytes) {
						t.Errorf("json response did not validate! %v", bytes)
					}

					// check the response body is what we expect
					if body := rr.Body.String(); body != tc.expectedBody {
						t.Errorf("handler returned unexpected body: got %v want %v",
							body, tc.expectedBody)
					}
				})
			}

		})
	}
}

// // Below is a leftover test struct I was using instead of httptest.ResponseRecorder,
// // thinking that it would reduce perf overhead in benchmarking, but it didnt seem to
// // make a big difference, so leaving out for now but preserving for future thoughts.
// type NullResponseWriter struct{}

// func (n NullResponseWriter) Header() http.Header {
// 	return http.Header{}
// }
// func (n NullResponseWriter) Write([]byte) (int, error) {
// 	return 0, nil
// }
// func (n NullResponseWriter) WriteHeader(statusCode int) {}

func BenchmarkHTTPRequest(b *testing.B) {
	db, err := NewLookupDB(*dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	req, _ := http.NewRequest("GET", "/?ip=89.160.20.112", nil)
	rr := httptest.NewRecorder() //NullResponseWriter{}
	handler := HTTPHandler{
		DB:       db,
		MemCache: nil,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkHTTPRequestWithCache(b *testing.B) {
	db, err := NewLookupDB(*dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	req, _ := http.NewRequest("GET", "/?ip=89.160.20.112", nil)
	rr := httptest.NewRecorder()
	handler := HTTPHandler{
		DB:       db,
		MemCache: cache.New(DefaultCacheExpiration, DefaultCacheCleanup),
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		handler.ServeHTTP(rr, req)
	}
}
