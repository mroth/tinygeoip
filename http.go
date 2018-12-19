package geominder

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// DefaultOriginPolicy is the default for `Access-Control-Allow-Origin` header
const DefaultOriginPolicy = "*"

// HTTPHandler implements a standard http.Handler interface for accessing
// a LookupDB, and provides in-memory caching for results.
type HTTPHandler struct {
	// Handle to the LookupDB used for queries.
	DB *LookupDB
	// Value for `Access-Control-Allow-Origin` header.
	//
	// Header will be omitted if set to zero value.
	OriginPolicy string
}

// NewHTTPHandler creates a HTTPHandler for requests againt the given LookupDB
//
// By default caching is enabled, and DefaultOriginPolicy is applied.
func NewHTTPHandler(db *LookupDB) *HTTPHandler {
	return &HTTPHandler{
		DB:           db,
		OriginPolicy: DefaultOriginPolicy,
	}
}

// SetOriginPolicy sets value for `Access-Control-Allow-Origin` header
//
// Returns pointer to the HTTPHandler to enable chaining in builder pattern.
func (hh *HTTPHandler) SetOriginPolicy(origins string) *HTTPHandler {
	hh.OriginPolicy = origins
	return hh
}

// ServeHTTP implements the http.Handler interface
func (hh *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set headers
	if hh.OriginPolicy != "" {
		w.Header().Set("Access-Control-Allow-Origin", hh.OriginPolicy)
	}
	w.Header().Set("Content-Type", "application/json")

	// attempt to parse IP from query
	ipText := strings.TrimPrefix(r.URL.Path, "/")

	// nice error message when missing data
	if ipText == "" {
		w.WriteHeader(http.StatusBadRequest)
		const parseIPError = `{"error": "missing IP query parameter, try ?ip=foo"}`
		w.Write([]byte(parseIPError))
		return
	}

	// attempt to parse the provided IP address
	ip := net.ParseIP(ipText)
	if ip == nil {
		w.WriteHeader(http.StatusBadRequest)
		const parseIPError = `{"error": "could not parse invalid IP address"}`
		w.Write([]byte(parseIPError))
		return
	}

	// do a DB lookup on the IP address
	loc, err := hh.DB.Lookup(ip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err.Error())))
		return
	}

	// rerturn results as JSON
	//
	// (yes, we're swallowing a potential marshall error here, but we already
	// know loc should not be nil since we checked for err on the previous case)
	b, _ := json.Marshal(loc)
	w.Header().Set("Last-Modified", serverStartTime)
	w.Write(b)
}

// for the last-modified time to hint to HTTP caching of results, we just use
// program launch time, as the values will never change outside of that. (we
// don't use the underlying database build time because the program itself may
// be modified.)
var (
	serverStartTime = time.Now().Format(http.TimeFormat)
)
