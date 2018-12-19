package geominder

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache"
)

// DefaultCacheExpiration is the default time duration until a cache expiry
const DefaultCacheExpiration = 24 * time.Hour

// DefaultMaxCacheSize is the default max memory used for caching responses (in MB)
const DefaultMaxCacheSize = 512

// DefaultOriginPolicy is the default for `Access-Control-Allow-Origin` header
const DefaultOriginPolicy = "*"

// contentType defines the content type used for all responses.
//
// This is effectively used as a constant. We define it as a package variable in
// memory here because it allows a hacky but effective performance optimization
// -- as the normal API for setting HTTP response headers via Set() in Go will
// create a memory allocation for each header set on the request. However, if we
// store this as a slice to begin with, we can set in the header map directly
// and avoid the memory allocation.
var contentType = []string{"application/json"}

// HTTPHandler implements a standard http.Handler interface for accessing
// a LookupDB, and provides in-memory caching for results.
type HTTPHandler struct {
	// Handle to the LookupDB used for queries.
	DB *LookupDB
	// Values for `Access-Control-Allow-Origin` header.
	//
	// Header will be omitted if set to zero value.
	OriginPolicy []string
	// Backing cache used for in-memory caching of responses.
	//
	// TODO: before v1.0, the memcache should potentially be privatized so that
	// API stability can be more easily preserved if it is switched out.
	MemCache *bigcache.BigCache
}

// NewHTTPHandler creates a HTTPHandler for requests againt the given LookupDB
//
// By default caching is enabled, and DefaultOriginPolicy is applied.
func NewHTTPHandler(db *LookupDB) *HTTPHandler {
	hh := HTTPHandler{
		DB:           db,
		OriginPolicy: []string{DefaultOriginPolicy},
	}
	hh.EnableCache()
	return &hh
}

// EnableCache activates the memory cache for a HTTPHandler with default values.
//
// Returns pointer to the HTTPHandler to enable chaining in builder pattern.
func (hh *HTTPHandler) EnableCache() *HTTPHandler {
	return hh._enableCache(DefaultMaxCacheSize)
}

// EnableCacheOfSize activates the memory cache for a HTTPHandler with max size.
//
// Returns pointer to the HTTPHandler to enable chaining in builder pattern.
func (hh *HTTPHandler) EnableCacheOfSize(maxCacheSize uint) *HTTPHandler {
	return hh._enableCache(maxCacheSize)
}

func (hh *HTTPHandler) _enableCache(maxCacheSize uint) *HTTPHandler {
	config := bigcache.DefaultConfig(DefaultCacheExpiration)
	config.HardMaxCacheSize = int(maxCacheSize)
	// the swallowed error here is only generated when passing an invalid config
	// to NewBigCache, e.g. number of shards is not a power of two, so should be
	// "unreachable!"
	hh.MemCache, _ = bigcache.NewBigCache(config)
	return hh
}

// DisableCache deactivates the memory cache for a HTTPHandler
//
// Returns pointer to the HTTPHandler to enable chaining in builder pattern.
func (hh *HTTPHandler) DisableCache() *HTTPHandler {
	if hh.MemCache != nil {
		cacheHandle := hh.MemCache
		defer cacheHandle.Close()
	}
	hh.MemCache = nil
	return hh
}

// SetOriginPolicy sets a single value for `Access-Control-Allow-Origin` header
//
// Returns pointer to the HTTPHandler to enable chaining in builder pattern.
func (hh *HTTPHandler) SetOriginPolicy(origin string) *HTTPHandler {
	if origin != "" {
		hh.OriginPolicy = []string{origin}
	} else {
		hh.OriginPolicy = []string{}
	}
	return hh
}

// ServeHTTP implements the http.Handler interface
func (hh *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set headers
	header := w.Header()
	header["Content-Type"] = contentType
	if len(hh.OriginPolicy) != 0 {
		header["Access-Control-Allow-Origin"] = hh.OriginPolicy
	}
	// w.Header().Set("Last-Modified", serverStart)

	// attempt to parse IP from query
	ipText := strings.TrimPrefix(r.URL.Path, "/")

	// nice error message when missing data
	if ipText == "" {
		w.WriteHeader(http.StatusBadRequest)
		const parseIPError = `{"error": "missing IP query parameter, try ?ip=foo"}`
		w.Write([]byte(parseIPError))
		return
	}

	// check for cached result
	if hh.MemCache != nil {
		cached, err := hh.MemCache.Get(ipText) // EntryNotFoundError on cache miss
		if err == nil {
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
	loc := locPool.Get().(*LookupResult)
	err := hh.DB.FastLookup(ip, loc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err.Error())))
		return
	}

	// rerturn results as JSON + update in cache if cache enabled
	b := loc.FasterJSON()
	w.Write(*b)
	if hh.MemCache != nil {
		hh.MemCache.Set(ipText, *b)
	} else {
		// only return to pool if we didnt cache
		loc.PoolReturn(b)
	}

	locPool.Put(loc)
}

// pool for loc lookup responses
var locPool = &sync.Pool{
	New: func() interface{} {
		return new(LookupResult)
	},
}
