package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"github.com/pmylund/go-cache"
)

const DefaultCacheExpiration = 5 * time.Minute
const DefaultCacheCleanup = 10 * time.Minute

// LookupHandler is essentially wrapper around a maxminddb.Reader
// which handles simple location lookups in an efficient way,
// as well as implements a standard http.Handler interface.
type LookupHandler struct {
	DB           *maxminddb.Reader
	OriginPolicy string
	MemCache     *cache.Cache
}

func NewLookupHandler(db *maxminddb.Reader) LookupHandler {
	return LookupHandler{
		DB: db,
	}
}

// https://dev.maxmind.com/geoip/geoip2/web-services/
// maybe make same as https://github.com/bluesmoon/node-geoip
type LookupResult struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code" json:"iso_code"`
	} `maxminddb:"country" json:"country"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude" json:"lat"`
		Longitude float64 `maxminddb:"longitude" json:"long"`
		// The approximate accuracy radius, in kilometers, around the
		// latitude and longitude for the geographical entity (country,
		// subdivision, city or postal code) associated with the IP address.
		// We have a 67% confidence that the location of the end-user falls
		// within the area defined by the accuracy radius and the latitude
		// and longitude coordinates.
		Accuracy int `maxminddb:"accuracy_radius" json:"accuracy"`
		// The time zone associated with location, as specified by the IANA
		// Time Zone Database, e.g., “America/New_York”.
		// Timezone string `maxminddb:"time_zone"`
	} `maxminddb:"location" json:"location"`
}

// Lookup a specified ip and returns the location
func (lh *LookupHandler) Lookup(ip net.IP) (*LookupResult, error) {
	var r LookupResult
	err := lh.DB.Lookup(ip, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ServeHTTP implements the http.Handler interface
func (lh *LookupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set headers
	if lh.OriginPolicy != "" {
		w.Header().Set("Access-Control-Allow-Origin", lh.OriginPolicy)
	}
	w.Header().Set("Content-Type", "application/json")
	// w.Header().Set("Last-Modified", serverStart)

	// attempt to parse IP from query
	ipText := r.URL.Query().Get("ip")
	if ipText == "" {
		ipText = strings.Trim(r.URL.Path, "/")
	}

	// check for cached result
	if lh.MemCache != nil {
		v, found := lh.MemCache.Get(ipText)
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

	loc, err := lh.Lookup(ip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: error text
		return
	}

	b, err := json.Marshal(loc)
	w.Write(b)
	if lh.MemCache != nil {
		lh.MemCache.Set(ipText, b, cache.DefaultExpiration)
	}
}
