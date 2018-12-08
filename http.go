package geominder

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pmylund/go-cache"
)

// DefaultCacheExpiration is the default time duration until a cache expiry
const DefaultCacheExpiration = 5 * time.Minute

// DefaultCacheCleanup is the default time duration until cache cleanup
const DefaultCacheCleanup = 10 * time.Minute

// DefaultOriginPolicy is the default for `Access-Control-Allow-Origin` header
const DefaultOriginPolicy = "*"

// HTTPHandler implements a standard http.Handler interface for accessing
// a LookupDB, and provides in-memory caching for results.
type HTTPHandler struct {
	DB *LookupDB
	// Value for `Access-Control-Allow-Origin` header.
	//
	// Header will be omitted if set to zero value.
	OriginPolicy string
	MemCache     *cache.Cache
	// TODO: before v1.0, the memcache should potentially be privatized so that
	// API stability can be more easily preserved if it is switched out.
}

// NewHTTPHandler creates a HTTPHandler for requests againt the given LookupDB
//
// By default caching is enabled, and DefaultOriginPolicy is applied.
func NewHTTPHandler(db *LookupDB) *HTTPHandler {
	hh := HTTPHandler{
		DB:           db,
		OriginPolicy: DefaultOriginPolicy,
	}
	hh.EnableCache()
	return &hh
}

// EnableCache activates the memory cache for a HTTPHandler with default values
//
// If you wish to provide custom cache values, you'll need to manipulate the
// struct values directly for now.
//
// Returns pointer to the HTTPHandler to enable chaining in builder pattern.
func (hh *HTTPHandler) EnableCache() *HTTPHandler {
	hh.MemCache = cache.New(DefaultCacheExpiration, DefaultCacheCleanup)
	return hh
}

// DisableCache deactivates the memory cache for a HTTPHandler
//
// Returns pointer to the HTTPHandler to enable chaining in builder pattern.
func (hh *HTTPHandler) DisableCache() *HTTPHandler {
	hh.MemCache = nil
	return hh
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
	// w.Header().Set("Last-Modified", serverStart)

	// attempt to parse IP from query
	ipText := r.URL.Query().Get("ip")

	// nice error message when missing data
	if ipText == "" {
		w.WriteHeader(http.StatusBadRequest)
		const parseIPError = `{"error": "missing IP query parameter, try ?ip=foo"}`
		w.Write([]byte(parseIPError))
		return
	}

	// check for cached result
	if hh.MemCache != nil {
		v, found := hh.MemCache.Get(ipText)
		if found {
			cached := v.([]byte)
			w.Write(cached)
			return
		}
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

	// rerturn results as JSON + update in cache if cache enabled
	//
	// (yes, we're swallowing a potential marshall error here, but we already
	// know loc should not be nil since we checked for err on the previous case)
	b, _ := json.Marshal(loc)
	w.Write(b)
	if hh.MemCache != nil {
		hh.MemCache.Set(ipText, b, cache.DefaultExpiration)
	}
}
