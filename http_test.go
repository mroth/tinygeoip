package geominder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmylund/go-cache"
)

func TestHTTPLookup(t *testing.T) {
	var httpCases = []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
		expectedBody   string
	}{
		{
			name:           "happy",
			path:           "/?ip=8.8.8.8", // TODO: replace with testdb version
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			expectedBody:   `{"country":{"iso_code":"US"},"location":{"lat":37.751,"long":-97.822,"accuracy":1000}}`,
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
			path:           "/?ip=127.0.0.1",     // TODO: what ip to check to generate???
			expectedStatus: http.StatusNoContent, // TODO: what is the best status here?
			expectedType:   "application/json",
			expectedBody:   `{"error": "TODO - LETS SEE WHAT WE GET BACK FROM MAXMIND"}`,
		},
	}

	db, err := NewLookupDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler := HTTPHandler{
		DB:       db,
		MemCache: nil,
	}

	for _, tc := range httpCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

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
				t.Errorf("json resopnse did not validate! %v", bytes)
			}

			// check the response body is what we expect
			if body := rr.Body.String(); body != tc.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					body, tc.expectedBody)
			}
		})
	}
}

// TODO: test caching validity
// idea: send req 1A,2B,3A.  verify 1==3, 1!=2, 2!=3

// type NullResponseWriter struct{}

// func (n NullResponseWriter) Header() http.Header {
// 	return http.Header{}
// }
// func (n NullResponseWriter) Write([]byte) (int, error) {
// 	return 0, nil
// }
// func (n NullResponseWriter) WriteHeader(statusCode int) {}

func BenchmarkHTTPRequest(b *testing.B) {
	db, err := NewLookupDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	req, _ := http.NewRequest("GET", "/?ip=8.8.8.8", nil)
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
	db, err := NewLookupDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	req, _ := http.NewRequest("GET", "/?ip=8.8.8.8", nil)
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
