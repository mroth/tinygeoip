package tinygeoip

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// DefaultOriginPolicy is the default for `Access-Control-Allow-Origin` header.
const DefaultOriginPolicy = "*"

// HTTPHandler implements a standard http.Handler interface for accessing
// a LookupDB.
type HTTPHandler struct {
	// Handle to the LookupDB used for queries.
	DB *LookupDB
	// Value for `Access-Control-Allow-Origin` header.
	//
	// Header will be omitted if set to zero value.
	OriginPolicy string
}

// NewHTTPHandler creates a HTTPHandler for requests against the given LookupDB.
//
// By default caching is enabled, and DefaultOriginPolicy is applied.
func NewHTTPHandler(db *LookupDB) *HTTPHandler {
	return &HTTPHandler{
		DB:           db,
		OriginPolicy: DefaultOriginPolicy,
	}
}

// SetOriginPolicy sets value for `Access-Control-Allow-Origin` header.
//
// Returns pointer to the HTTPHandler to enable chaining in builder pattern.
func (hh *HTTPHandler) SetOriginPolicy(origins string) *HTTPHandler {
	hh.OriginPolicy = origins
	return hh
}

// ServeHTTP implements the standard http.Handler interface.
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
		w.Write(responseErrMissingIP)
		return
	}

	// attempt to parse the provided IP address
	ip := net.ParseIP(ipText)
	if ip == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseErrInvalidIP)
		return
	}

	// do a DB lookup on the IP address
	loc, err := hh.DB.Lookup(ip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fmt.Appendf(nil, `{"error": "%v"}`, err.Error()))
		return
	}

	// return results as JSON
	w.Header().Set("Last-Modified", serverStartTime)
	json.NewEncoder(w).Encode(loc)
}

var (
	responseErrMissingIP = []byte(`{"error": "missing IP query in path, try /192.168.1.1"}`)
	responseErrInvalidIP = []byte(`{"error": "could not parse invalid IP address"}`)
)

// for the last-modified time to hint to HTTP caching of results, we just use
// program launch time, as the values will never change outside of that. (we
// don't use the underlying database build time because the program itself may
// be modified.)
var (
	serverStartTime = time.Now().Format(http.TimeFormat)
)
