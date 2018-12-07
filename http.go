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

// HTTPHandler implements a standard http.Handler interface for accessing
// a LookupDB, and provides in-memory caching for results.
type HTTPHandler struct {
	DB           *LookupDB
	OriginPolicy string
	MemCache     *cache.Cache
}

func NewHTTPHandler(db *LookupDB) HTTPHandler {
	return HTTPHandler{
		DB: db,
	}
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
