package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
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
	if ipText == "" {
		ipText = strings.Trim(r.URL.Path, "/")
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

	ip := net.ParseIP(ipText)
	if ip == nil {
		// returnError = "unable to decode ip"
		w.WriteHeader(http.StatusBadRequest)
		// TODO: error text
		return
	}

	loc, err := hh.DB.Lookup(ip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: error text
		return
	}

	b, err := json.Marshal(loc)
	w.Write(b)
	if hh.MemCache != nil {
		hh.MemCache.Set(ipText, b, cache.DefaultExpiration)
	}
}
